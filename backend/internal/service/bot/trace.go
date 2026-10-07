package bot

import (
	"strconv"
	"sync"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
)

const (
	maxTracedCalls     = 15
	maxTracedReactions = 30
	maxStepText        = 4000
)

// trace records what one bot did for the admin inspector: its model calls with their
// prompts, reasoning and tool calls, its plan, and its reaction decisions. It is bounded.
type trace struct {
	mu        sync.Mutex
	playerID  string
	nextID    int
	plan      dto.BotPlanDto
	planAt    string
	calls     []*dto.BotCallDto
	reactions []dto.BotReactionDto
	grudges   []dto.BotGrudgeDto
}

func newTrace(playerID string) *trace {
	return &trace{playerID: playerID, calls: []*dto.BotCallDto{}, reactions: []dto.BotReactionDto{}, grudges: []dto.BotGrudgeDto{}}
}

func traceTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func (t *trace) startCall(role, model, prompt string) (string, dto.BotTraceEventDto) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.nextID++
	call := &dto.BotCallDto{
		ID:        strconv.Itoa(t.nextID),
		Role:      role,
		Model:     model,
		StartedAt: traceTime(time.Now()),
		Running:   true,
		Prompt:    prompt,
		Steps:     []dto.BotCallStepDto{},
	}
	t.calls = append(t.calls, call)
	if len(t.calls) > maxTracedCalls {
		t.calls = t.calls[len(t.calls)-maxTracedCalls:]
	}
	copied := *call
	return call.ID, dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "call-start", Call: &copied}
}

func (t *trace) addStep(callID string, step dto.BotCallStepDto) (dto.BotTraceEventDto, bool) {
	step.At = traceTime(time.Now())
	step.Text = truncate(step.Text, maxStepText)
	step.Result = truncate(step.Result, maxStepText)
	step.Input = truncate(step.Input, maxStepText)
	t.mu.Lock()
	defer t.mu.Unlock()
	call := t.find(callID)
	if call == nil {
		return dto.BotTraceEventDto{}, false
	}
	call.Steps = append(call.Steps, step)
	return dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "call-step", CallID: callID, Step: &step}, true
}

func (t *trace) endCall(callID string, result Result, err error) (dto.BotTraceEventDto, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	call := t.find(callID)
	if call == nil {
		return dto.BotTraceEventDto{}, false
	}
	call.Running = false
	call.EndedAt = traceTime(time.Now())
	call.CostUSD = result.CostUSD
	call.DurationMs = int(result.Duration.Milliseconds())
	if err != nil {
		call.Error = truncate(err.Error(), 500)
	}
	summary := *call
	summary.Prompt = ""
	summary.Steps = nil
	return dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "call-end", Call: &summary}, true
}

func (t *trace) setPlan(plan Plan) dto.BotTraceEventDto {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.plan = planDto(plan)
	t.planAt = traceTime(time.Now())
	copied := t.plan
	return dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "plan", Plan: &copied, PlanUpdatedAt: t.planAt}
}

func (t *trace) addReaction(r dto.BotReactionDto) dto.BotTraceEventDto {
	r.At = traceTime(time.Now())
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reactions = append(t.reactions, r)
	if len(t.reactions) > maxTracedReactions {
		t.reactions = t.reactions[len(t.reactions)-maxTracedReactions:]
	}
	return dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "reaction", Reaction: &r}
}

func (t *trace) setGrudges(grudges []dto.BotGrudgeDto) dto.BotTraceEventDto {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.grudges = grudges
	return dto.BotTraceEventDto{PlayerID: t.playerID, Kind: "grudges", Grudges: append([]dto.BotGrudgeDto{}, grudges...)}
}

func (t *trace) snapshot() dto.BotTraceDto {
	t.mu.Lock()
	defer t.mu.Unlock()
	calls := make([]dto.BotCallDto, 0, len(t.calls))
	for _, c := range t.calls {
		copied := *c
		copied.Steps = append([]dto.BotCallStepDto{}, c.Steps...)
		calls = append(calls, copied)
	}
	return dto.BotTraceDto{
		PlayerID:      t.playerID,
		Plan:          t.plan,
		PlanUpdatedAt: t.planAt,
		Calls:         calls,
		Reactions:     append([]dto.BotReactionDto{}, t.reactions...),
		Grudges:       append([]dto.BotGrudgeDto{}, t.grudges...),
	}
}

func (t *trace) find(callID string) *dto.BotCallDto {
	for _, c := range t.calls {
		if c.ID == callID {
			return c
		}
	}
	return nil
}

func planDto(p Plan) dto.BotPlanDto {
	moves := make([]dto.BotPlannedMoveDto, 0, len(p.NextMoves))
	for _, m := range p.NextMoves {
		moves = append(moves, dto.BotPlannedMoveDto{Move: m.Move, Why: m.Why})
	}
	orEmpty := func(v []string) []string {
		if v == nil {
			return []string{}
		}
		return v
	}
	return dto.BotPlanDto{
		Summary:          p.Summary,
		WantedHexes:      orEmpty(p.WantedHexes),
		TargetMilestones: orEmpty(p.TargetMilestones),
		TargetAwards:     orEmpty(p.TargetAwards),
		TargetColonies:   orEmpty(p.TargetColonies),
		Rivals:           orEmpty(p.Rivals),
		NextMoves:        moves,
		Mood:             p.Mood,
		Notes:            p.Notes,
	}
}
