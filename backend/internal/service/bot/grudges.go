package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"openmars/internal/delivery/dto"
)

const maxGrudgeRecent = 4

// grudge counts what one opponent has done against the bot this game.
type grudge struct {
	playerID string
	name     string
	hits     int
	recent   []string
	lastAt   time.Time
}

// grudgeLedger is the bot's exact memory of who has targeted it, kept in Go so it does not
// depend on the model remembering. Rivals, retaliation and chat build on it.
type grudgeLedger map[string]*grudge

func (l grudgeLedger) record(h Hostility, name string, at time.Time) {
	g := l[h.ActorID]
	if g == nil {
		g = &grudge{playerID: h.ActorID}
		l[h.ActorID] = g
	}
	g.name = name
	g.hits++
	g.lastAt = at
	g.recent = append(g.recent, h.What)
	if len(g.recent) > maxGrudgeRecent {
		g.recent = g.recent[len(g.recent)-maxGrudgeRecent:]
	}
}

func (l grudgeLedger) sorted() []*grudge {
	out := make([]*grudge, 0, len(l))
	for _, g := range l {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].hits != out[j].hits {
			return out[i].hits > out[j].hits
		}
		return out[i].lastAt.After(out[j].lastAt)
	})
	return out
}

// describe renders the ledger for prompts.
func (l grudgeLedger) describe() string {
	if len(l) == 0 {
		return "Nobody has targeted you yet."
	}
	lines := make([]string, 0, len(l))
	for _, g := range l.sorted() {
		times := "once"
		if g.hits > 1 {
			times = fmt.Sprintf("%d times", g.hits)
		}
		lines = append(lines, fmt.Sprintf("%s has acted against you %s. Recently: %s.", g.name, times, strings.Join(g.recent, "; ")))
	}
	return strings.Join(lines, "\n")
}

func (l grudgeLedger) dtos() []dto.BotGrudgeDto {
	out := make([]dto.BotGrudgeDto, 0, len(l))
	for _, g := range l.sorted() {
		out = append(out, dto.BotGrudgeDto{PlayerID: g.playerID, Name: g.name, Hits: g.hits, Recent: append([]string{}, g.recent...)})
	}
	return out
}
