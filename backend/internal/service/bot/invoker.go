package bot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"terraforming-mars-backend/internal/delivery/dto"

	"go.uber.org/zap"
)

// Invoker executes the Claude CLI as a subprocess to decide turns.
type Invoker struct {
	HistoryPath string
	StatePath   string
	CommandPath string
	Model       string
	APIKey      string
	Difficulty  string
	logger      *zap.Logger
}

// Stream event types from claude --output-format stream-json
type streamEvent struct {
	Type       string         `json:"type"`
	Subtype    string         `json:"subtype,omitempty"`
	Message    *streamMessage `json:"message,omitempty"`
	Result     string         `json:"result,omitempty"`
	Cost       float64        `json:"total_cost_usd,omitempty"`
	NumTurns   int            `json:"num_turns,omitempty"`
	DurationMs int            `json:"duration_ms,omitempty"`

	ToolUseResult *toolUseResult `json:"tool_use_result,omitempty"`
}

type streamMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type contentBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	ID      string          `json:"id,omitempty"`
	Name    string          `json:"name,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`

	ToolUseID string `json:"tool_use_id,omitempty"`
}

type toolUseResult struct {
	Type string       `json:"type"`
	File *toolUseFile `json:"file,omitempty"`
}

type toolUseFile struct {
	FilePath string `json:"filePath,omitempty"`
	NumLines int    `json:"numLines,omitempty"`
}

// NewInvoker creates a new Claude CLI invoker.
func NewInvoker(historyPath, statePath, commandPath, model, apiKey, difficulty string, logger *zap.Logger) *Invoker {
	return &Invoker{
		HistoryPath: historyPath,
		StatePath:   statePath,
		CommandPath: commandPath,
		Model:       model,
		APIKey:      apiKey,
		Difficulty:  difficulty,
		logger:      logger,
	}
}

// PlayTurn invokes Claude CLI to decide and execute actions for the current turn.
func (inv *Invoker) PlayTurn(ctx context.Context, gameDto *dto.GameDto, myPlayerID string) error {
	systemPrompt := buildSystemPrompt(inv.CommandPath, inv.Difficulty, inv.logger)
	turnPrompt := buildTurnPrompt(gameDto, inv.StatePath, inv.CommandPath, inv.HistoryPath)

	inv.logger.Info("🤖 Invoking Claude CLI",
		zap.String("model", inv.Model),
		zap.String("player_id", myPlayerID))

	cmd := exec.CommandContext(ctx, "claude",
		"--print",
		"--output-format", "stream-json",
		"--verbose",
		"--model", inv.Model,
		"--dangerously-skip-permissions",
		"--tools", "Read,Bash",
		"--append-system-prompt", systemPrompt,
		turnPrompt,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// Set API key in environment
	if inv.APIKey != "" {
		cmd.Env = append(cmd.Environ(), "CLAUDE_CODE_OAUTH_TOKEN="+inv.APIKey)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start claude CLI: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event streamEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		inv.displayEvent(&event)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("claude CLI: %w", err)
	}

	inv.logger.Info("✅ Claude CLI exited successfully")
	return nil
}

func (inv *Invoker) displayEvent(event *streamEvent) {
	switch event.Type {
	case "assistant":
		if event.Message == nil {
			return
		}
		var blocks []contentBlock
		if err := json.Unmarshal(event.Message.Content, &blocks); err != nil {
			return
		}
		for _, block := range blocks {
			switch block.Type {
			case "text":
				if block.Text != "" {
					inv.logger.Info("🤖 Claude", zap.String("text", block.Text))
				}
			case "tool_use":
				inv.displayToolUse(block)
			}
		}

	case "user":
		if event.ToolUseResult != nil && event.ToolUseResult.File != nil {
			f := event.ToolUseResult.File
			inv.logger.Debug("📄 Tool result", zap.String("file", f.FilePath), zap.Int("lines", f.NumLines))
		}

	case "result":
		if event.Subtype == "success" {
			inv.logger.Info("🤖 Claude done",
				zap.Int("turns", event.NumTurns),
				zap.Int("duration_ms", event.DurationMs),
				zap.Float64("cost_usd", event.Cost))
		} else {
			inv.logger.Error("🤖 Claude error", zap.String("result", event.Result))
		}
	}
}

func (inv *Invoker) displayToolUse(block contentBlock) {
	switch block.Name {
	case "Read":
		var input struct {
			FilePath string `json:"file_path"`
		}
		json.Unmarshal(block.Input, &input)
		inv.logger.Debug("📖 Read", zap.String("file", input.FilePath))

	case "Bash":
		var input struct {
			Command string `json:"command"`
		}
		json.Unmarshal(block.Input, &input)
		cmd := input.Command
		if len(cmd) > 200 {
			cmd = cmd[:200] + "..."
		}
		if strings.Contains(cmd, ">> ") {
			inv.logger.Info("📤 Command", zap.String("cmd", cmd))
		} else {
			inv.logger.Debug("🔧 Bash", zap.String("cmd", cmd))
		}

	default:
		inv.logger.Debug("🔧 Tool", zap.String("name", block.Name))
	}
}

func loadStrategyGuide(difficulty string, logger *zap.Logger) string {
	if difficulty == "" {
		difficulty = "normal"
	}

	wd, err := os.Getwd()
	if err != nil {
		logger.Warn("Failed to get working directory for strategy file", zap.Error(err))
		return ""
	}

	strategyPath := filepath.Join(wd, "assets", "bot", difficulty+".md")
	data, err := os.ReadFile(strategyPath)
	if err != nil {
		logger.Warn("Failed to load strategy file", zap.String("path", strategyPath), zap.Error(err))
		return ""
	}

	return string(data)
}

func buildSystemPrompt(commandPath string, difficulty string, logger *zap.Logger) string {
	strategyGuide := loadStrategyGuide(difficulty, logger)
	strategySection := ""
	if strategyGuide != "" {
		strategySection = fmt.Sprintf("\n\n=====================\nSTRATEGY GUIDE\n=====================\n\n%s", strategyGuide)
	}

	return fmt.Sprintf(`You are an expert Terraforming Mars player controlling a bot via WebSocket commands.
Your job is to analyze the game state and decide what actions to take for your turn.

CRITICAL RULES:
1. Read the game state file FIRST to understand the current situation.
2. Check "Actions remaining: N" in YOUR STATUS section. This is how many main actions you can take.
3. FAST EXIT: If Actions remaining ≤ 0 AND there is NO pending action, immediately send skip-action and STOP. Do not analyze cards, do not deliberate. Just skip.
4. Send AT MOST N action commands. Do NOT send more than the remaining actions allow.
5. Pending follow-ups (tile placement, card selection, behavior choices, discard) do NOT consume actions - resolve them after the action that triggered them.
6. After EACH command, read the tail of the history file to verify it was accepted before sending the next.
7. After you have used all remaining actions (or passed), STOP. Do not send any more commands.
8. If "Actions remaining: 0", you likely only have a pending follow-up to resolve. Resolve it and STOP.
9. skip-action counts as an action and ends your participation for this generation.

HOW TO SEND COMMANDS:
Use Bash to append each command as a JSON line:
  echo '{"type": "action.game-management.skip-action", "payload": {}}' >> %s

After sending a command, read the tail of the history file to verify it was accepted (look for "action-success" or error messages).

=====================
PHASE-SPECIFIC GUIDE
=====================

--- STARTING SELECTION PHASE ---
You must pick a corporation and starting cards in a single command.
The state file shows your available corporations and available starting cards.
Each starting card costs 3M€ to buy (deducted from your corporation's starting credits).
Send: {"type": "action.card.select-starting-choices", "payload": {"corporationId": "CORP_ID", "preludeIds": [], "cardIds": ["c1", "c2"]}}
  - corporationId: pick one from the available corporations
  - preludeIds: pick preludes if available (usually empty in base game)
  - cardIds: array of card IDs you want to BUY (can be empty [] to buy none)

How to choose:
  - Read each corporation's description carefully. Note starting credits, production, and special effects.
  - Calculate budget: corporation starting credits minus (3 x number of cards bought).
  - Pick cards that synergize with your corporation. E.g., Tharsis Republic (city bonuses) pairs well with city-related cards.
  - Prefer cards with good economy (credit production), terraform rating bumps, or strong VP potential.
  - Don't buy cards you can't afford to play for many generations. Avoid expensive cards (20+ cost) early unless they have amazing value.
  - It's OK to buy 0-3 cards. Don't overbuy and start broke.

--- FORCED FIRST ACTION ---
Some corporations require a specific first action (e.g., Tharsis Republic must place a city).
The state file shows "FORCED FIRST ACTION" with the required action type.
For city placement: a pending tile selection will appear with available hexes.
Send: {"type": "action.tile-selection.tile-selected", "payload": {"hex": "q,r,s"}}
  - Pick a hex from the available hexes list in the state file.
  - For cities: prefer hexes with placement bonuses (steel, credits) and room for adjacent greeneries later.
  - Avoid edges of the board if possible - more adjacency options in the center.

--- ACTION PHASE (main gameplay) ---
You get 2 actions per turn. Each action can be one of:
  1. Play a card from hand (if PLAYABLE - check the state file for availability)
  2. Use a card action from a played card (if AVAILABLE)
  3. Standard project (sell patents, build power plant, asteroid, aquifer, greenery, city)
  4. Resource conversion (8 plants -> greenery, 8 heat -> temperature)
  5. Claim a milestone (if CLAIMABLE)
  6. Fund an award (if FUNDABLE)
  7. Skip/pass (ends your participation this generation)

Priority order for each action:
  1. Resolve any pending actions first (tile placement, card selection, etc.)
  2. Claim a milestone if you qualify and can afford it (8M€ first, 8M€ second, 8M€ third - only 3 total)
  3. Play strong cards that boost production or give TR
  4. Use available card actions
  5. Convert resources if beneficial (plants -> greenery for TR + VP, heat -> temperature for TR)
  6. Standard projects as fallback
  7. Pass if nothing good to do (saves money for next generation)

When passing: you skip the rest of this generation. All players must pass to end the generation.

--- PRODUCTION PHASE ---
At the end of each generation, you draw 4 cards and choose which to buy at 3M€ each.
The state file shows "PRODUCTION PHASE - Select cards to buy" with available cards.
Send: {"type": "action.card.confirm-production-cards", "payload": {"cardIds": ["id1"]}}
  - cardIds: array of card IDs you want to buy (can be empty [])
  - Only buy cards you'll realistically play. Don't waste credits on cards you can't use.

--- TILE PLACEMENT ---
When a pending tile selection appears, you must place a tile.
Send: {"type": "action.tile-selection.tile-selected", "payload": {"hex": "q,r,s"}}
  - Use the exact coordinates from the "Available hexes" list (format: "q,r,s")
  - Cities: must NOT be adjacent to other cities. Place near future greenery spots.
  - Greeneries: must be adjacent to one of your tiles if possible. Each greenery = 1 VP.
  - Oceans: placed on ocean spaces only. Give 2 TR when placed.

=====================
COMMAND REFERENCE
=====================

== PLAY A CARD FROM HAND ==
{"type": "action.card.play-card", "payload": {"cardId": "CARD_ID", "payment": {"credits": N, "steel": N, "titanium": N}}}
Steel pays for building tags (1 steel = 2M€). Titanium pays for space tags (1 titanium = 3M€).
Optional payload fields: choiceIndex (int), targetPlayerId (string), selectedAmount (int)

== USE A CARD ACTION (from played cards) ==
{"type": "action.card.card-action", "payload": {"cardId": "CARD_ID", "behaviorIndex": N}}
Optional: choiceIndex, targetPlayerId, sourceCardForInput, selectedAmount

== STANDARD PROJECTS ==
{"type": "action.standard-project.sell-patents", "payload": {}}
  (After sending, you must confirm with: {"type": "action.standard-project.confirm-sell-patents", "payload": {"selectedCardIds": ["id1", "id2"]}})
{"type": "action.standard-project.launch-asteroid", "payload": {}}
{"type": "action.standard-project.build-power-plant", "payload": {}}
{"type": "action.standard-project.build-aquifer", "payload": {}}
{"type": "action.standard-project.plant-greenery", "payload": {}}
{"type": "action.standard-project.build-city", "payload": {}}

== RESOURCE CONVERSIONS ==
{"type": "action.resource-conversion.convert-plants-to-greenery", "payload": {}}
{"type": "action.resource-conversion.convert-heat-to-temperature", "payload": {}}

== SKIP/PASS ==
{"type": "action.game-management.skip-action", "payload": {}}

== TILE PLACEMENT ==
{"type": "action.tile-selection.tile-selected", "payload": {"hex": "q,r,s"}}

== STARTING SELECTION ==
{"type": "action.card.select-starting-choices", "payload": {"corporationId": "CORP_ID", "preludeIds": [], "cardIds": ["c1", "c2"]}}

== PRODUCTION PHASE CARD SELECTION ==
{"type": "action.card.confirm-production-cards", "payload": {"cardIds": ["id1"]}}

== CARD DRAW CONFIRMATION ==
{"type": "action.card.card-draw-confirmed", "payload": {"cardsToTake": ["id1"], "cardsToBuy": ["id2"]}}

== CARD DISCARD ==
{"type": "action.card.card-discard-confirmed", "payload": {"cardsToDiscard": ["id1"]}}

== BEHAVIOR CHOICE ==
{"type": "action.card.behavior-choice-confirmed", "payload": {"choiceIndex": 0}}

== CARD SELECTION ==
{"type": "action.card.select-cards", "payload": {"cardIds": ["id1"]}}

== MILESTONES ==
{"type": "action.milestone.claim-milestone", "payload": {"milestoneType": "TYPE"}}

== AWARDS ==
{"type": "action.award.fund-award", "payload": {"awardType": "TYPE"}}

== CHAT MESSAGE ==
{"type": "chat.send-message", "payload": {"message": "Your message here"}}
Chat is a big part of the game experience. Use it to engage with other players:
- React to opponent moves that affect you (stealing resources, blocking tiles, claiming milestones you wanted)
- Respond to chat messages from other players — don't leave people hanging
- Trash talk, banter, celebrate your own big plays
- Comment on the game state when something dramatic happens
Keep messages to 1 sentence, max 2 chat messages per turn. Do NOT narrate your actions — react and engage instead.%s`, commandPath, strategySection)
}

func buildTurnPrompt(game *dto.GameDto, statePath, commandPath, historyPath string) string {
	pendingAction := GetPendingActionType(game)
	pendingNote := ""
	if pendingAction != "" {
		pendingNote = fmt.Sprintf("\nIMPORTANT: You have a PENDING ACTION (%s) that MUST be resolved first! This does NOT consume an action.", pendingAction)
	}

	phaseNote := ""
	actionsNote := ""
	if game != nil {
		switch game.CurrentPhase {
		case dto.GamePhaseStartingSelection:
			phaseNote = "\nYou are in the STARTING SELECTION phase. Pick your corporation, preludes (if any), and starting cards. Send exactly 1 command."
		case dto.GamePhaseProductionAndCardDraw:
			phaseNote = "\nYou are in the PRODUCTION PHASE. Select which drawn cards to buy. Send exactly 1 command."
		case dto.GamePhaseAction:
			remaining := game.CurrentPlayer.AvailableActions
			actionsNote = fmt.Sprintf("\nYou have %d action(s) remaining this turn. Send at most %d action command(s), then STOP.", remaining, remaining)
		}
	}

	return fmt.Sprintf(`It's your turn to play Terraforming Mars!

GAME STATE FILE: %s
Read this file to see the current game state, your hand, resources, and available actions.

COMMAND FILE: %s
Append your commands here, one JSON object per line, using: echo '...' >> %s

HISTORY FILE: %s
After each command, read the tail of this file to check if it succeeded before sending the next.
%s%s%s
Begin by reading the game state file, then decide on your action(s).

CHAT: After reading the game state, check the RECENT GAME LOG and RECENT CHAT sections.
- If there are unanswered chat messages from other players, ALWAYS reply. Don't ignore people.
- If an opponent did something that affected you (stole resources, blocked your tile spot, claimed a milestone you were going for), react to it.
- If something big happened in the game log, comment on it.
- Send chat BEFORE your game actions. Stay in character per your personality style.`,
		statePath, commandPath, commandPath, historyPath, pendingNote, phaseNote, actionsNote)
}
