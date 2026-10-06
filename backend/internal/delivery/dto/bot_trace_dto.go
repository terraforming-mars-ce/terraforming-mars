package dto

// BotInspectPayload subscribes the sending connection to one bot's trace. An empty
// playerId unsubscribes. Host only, development mode only.
type BotInspectPayload struct {
	PlayerID string `json:"playerId"`
}

// BotPlannedMoveDto is one candidate move in a bot's plan.
type BotPlannedMoveDto struct {
	Move string `json:"move"`
	Why  string `json:"why"`
}

// BotPlanDto is a bot's current plan and memory.
type BotPlanDto struct {
	Summary          string              `json:"summary"`
	WantedHexes      []string            `json:"wantedHexes"`
	TargetMilestones []string            `json:"targetMilestones"`
	TargetAwards     []string            `json:"targetAwards"`
	TargetColonies   []string            `json:"targetColonies"`
	Rivals           []string            `json:"rivals"`
	NextMoves        []BotPlannedMoveDto `json:"nextMoves"`
	Mood             string              `json:"mood"`
	Notes            string              `json:"notes"`
}

// BotCallStepDto is one step of a model call: a piece of reasoning text or a tool call.
type BotCallStepDto struct {
	At      string `json:"at"`
	Kind    string `json:"kind" tstype:"'text' | 'tool'"`
	Text    string `json:"text,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Input   string `json:"input,omitempty"`
	Result  string `json:"result,omitempty"`
	IsError bool   `json:"isError"`
}

// BotCallDto is one model call made by a bot.
type BotCallDto struct {
	ID         string           `json:"id"`
	Role       string           `json:"role" tstype:"'executor' | 'planner' | 'reactor' | 'recap'"`
	Model      string           `json:"model"`
	StartedAt  string           `json:"startedAt"`
	EndedAt    string           `json:"endedAt,omitempty"`
	Running    bool             `json:"running"`
	Prompt     string           `json:"prompt"`
	Steps      []BotCallStepDto `json:"steps"`
	CostUSD    float64          `json:"costUsd"`
	DurationMs int              `json:"durationMs"`
	Error      string           `json:"error,omitempty"`
}

// BotReactionDto records how a bot judged a batch of table events.
type BotReactionDto struct {
	At       string   `json:"at"`
	Lines    []string `json:"lines"`
	PlanHit  bool     `json:"planHit"`
	Personal bool     `json:"personal"`
	Directed bool     `json:"directed"`
	Big      bool     `json:"big"`
	Decision string   `json:"decision" tstype:"'reacted' | 'throttled' | 'busy' | 'own-turn'"`
	Output   string   `json:"output,omitempty"`
}

// BotTraceDto is everything recorded about one bot.
type BotTraceDto struct {
	PlayerID      string           `json:"playerId"`
	Plan          BotPlanDto       `json:"plan"`
	PlanUpdatedAt string           `json:"planUpdatedAt,omitempty"`
	Calls         []BotCallDto     `json:"calls"`
	Reactions     []BotReactionDto `json:"reactions"`
}

// BotTraceEventDto is one change to a bot's trace. call-start carries the new call;
// call-step carries callId and step; call-end carries the finished call without prompt
// and steps; plan carries plan and planUpdatedAt; reaction carries reaction.
type BotTraceEventDto struct {
	PlayerID      string          `json:"playerId"`
	Kind          string          `json:"kind" tstype:"'call-start' | 'call-step' | 'call-end' | 'plan' | 'reaction'"`
	Call          *BotCallDto     `json:"call,omitempty"`
	CallID        string          `json:"callId,omitempty"`
	Step          *BotCallStepDto `json:"step,omitempty"`
	Plan          *BotPlanDto     `json:"plan,omitempty"`
	PlanUpdatedAt string          `json:"planUpdatedAt,omitempty"`
	Reaction      *BotReactionDto `json:"reaction,omitempty"`
}
