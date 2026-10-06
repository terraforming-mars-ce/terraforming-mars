package bot

import (
	"fmt"
	"strings"
)

// PlannedMove is one candidate move the planner proposes for the bot's next turn.
type PlannedMove struct {
	Move string `json:"move" jsonschema:"the move, e.g. 'play Mining Rights' or 'city on 2,-3,1'"`
	Why  string `json:"why" jsonschema:"short reason, written as an in-character intent line other players may see"`
}

// Plan is the bot's memory between turns. The relevance filter for reactions reads
// its structured fields; the executor follows NextMoves; Notes carry grudges and promises.
type Plan struct {
	Summary          string        `json:"summary" jsonschema:"one sentence: the overall plan right now"`
	WantedHexes      []string      `json:"wantedHexes,omitempty" jsonschema:"hexes (q,r,s) you intend to take soon"`
	TargetMilestones []string      `json:"targetMilestones,omitempty" jsonschema:"milestone types you are racing for"`
	TargetAwards     []string      `json:"targetAwards,omitempty" jsonschema:"award types you intend to fund or win"`
	TargetColonies   []string      `json:"targetColonies,omitempty" jsonschema:"colony IDs you want to build on or trade with"`
	Rivals           []string      `json:"rivals,omitempty" jsonschema:"names of players you are watching or competing with most"`
	NextMoves        []PlannedMove `json:"nextMoves,omitempty" jsonschema:"ranked candidate moves for your next turn"`
	Mood             string        `json:"mood,omitempty" jsonschema:"one word for your current mood"`
	Notes            string        `json:"notes,omitempty" jsonschema:"memory to keep: grudges, promises made in chat, what opponents are doing"`
}

// IsEmpty reports whether no plan has been recorded yet.
func (p Plan) IsEmpty() bool {
	return p.Summary == "" && len(p.NextMoves) == 0
}

// Intent returns the line shown to other players when the bot starts its turn.
func (p Plan) Intent() string {
	if len(p.NextMoves) > 0 && p.NextMoves[0].Why != "" {
		return p.NextMoves[0].Why
	}
	return p.Summary
}

// Describe renders the plan for a prompt.
func (p Plan) Describe() string {
	if p.IsEmpty() {
		return "No plan recorded yet."
	}
	lines := []string{"Summary: " + p.Summary}
	add := func(label string, values []string) {
		if len(values) > 0 {
			lines = append(lines, label+": "+strings.Join(values, ", "))
		}
	}
	add("Wanted hexes", p.WantedHexes)
	add("Target milestones", p.TargetMilestones)
	add("Target awards", p.TargetAwards)
	add("Target colonies", p.TargetColonies)
	add("Rivals", p.Rivals)
	for i, m := range p.NextMoves {
		lines = append(lines, fmt.Sprintf("Move %d: %s (%s)", i+1, m.Move, m.Why))
	}
	if p.Mood != "" {
		lines = append(lines, "Mood: "+p.Mood)
	}
	if p.Notes != "" {
		lines = append(lines, "Notes: "+p.Notes)
	}
	return strings.Join(lines, "\n")
}
