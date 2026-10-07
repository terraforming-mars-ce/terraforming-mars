package bot

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// HappeningKind classifies something that happened at the table.
type HappeningKind string

const (
	HappeningTile         HappeningKind = "tile"
	HappeningMilestone    HappeningKind = "milestone"
	HappeningAward        HappeningKind = "award"
	HappeningColony       HappeningKind = "colony"
	HappeningResourceLoss HappeningKind = "resource-loss"
	HappeningCard         HappeningKind = "card"
	HappeningChat         HappeningKind = "chat"
)

// Happening is one table event a bot may react to.
type Happening struct {
	Kind    HappeningKind
	ActorID string
	// Target is the hex, milestone, award, colony, resource or card name involved.
	Target string
	// Detail is the tile type for tiles and the message for chat.
	Detail string
	Amount int
	// Big marks events worth an occasional reaction even when they do not touch the plan.
	Big bool
	// FromBot marks chat written by another bot, which never bypasses the cooldown,
	// so two bots cannot answer each other forever.
	FromBot bool
	At      time.Time
}

// Assessment is the relevance filter's verdict on a batch of happenings.
type Assessment struct {
	Lines []string
	// PlanHit means an opponent took something the plan wanted, so the plan is stale.
	PlanHit bool
	// Personal means the bot was spoken to or hurt directly.
	Personal bool
	// Directed means a chat message was addressed to the bot, which bypasses the cooldown.
	Directed bool
	Big      bool
	// Hostile lists what opponents did against the bot in this batch, for its grudges.
	Hostile []Hostility
}

// Hostility is one act by an opponent against the bot.
type Hostility struct {
	ActorID string
	// What describes the act from the bot's side, e.g. "destroyed 6 of your plants".
	What string
}

// Relevant reports whether anything in the batch is worth a reaction before chance is applied.
func (a Assessment) Relevant() bool {
	return a.PlanHit || a.Personal || a.Big
}

// AssessHappenings decides which happenings matter to a bot, using its plan.
// soleOpponent is true when the bot has exactly one opponent, so all chat is addressed to it.
func AssessHappenings(plan Plan, botID, botName string, soleOpponent bool, happenings []Happening, nameOf func(playerID string) string) Assessment {
	var a Assessment
	for _, h := range happenings {
		if h.ActorID == botID {
			continue
		}
		actor := nameOf(h.ActorID)
		switch h.Kind {
		case HappeningTile:
			if slices.Contains(plan.WantedHexes, h.Target) {
				a.PlanHit = true
				a.Lines = append(a.Lines, fmt.Sprintf("%s placed a %s on %s, a hex you wanted.", actor, h.Detail, h.Target))
				a.Hostile = append(a.Hostile, Hostility{ActorID: h.ActorID, What: fmt.Sprintf("took hex %s you wanted", h.Target)})
			}
		case HappeningMilestone:
			a.addRace(plan.TargetMilestones, h, fmt.Sprintf("%s claimed the %s milestone", actor, h.Target), "claimed the "+h.Target+" milestone you were going for")
		case HappeningAward:
			a.addRace(plan.TargetAwards, h, fmt.Sprintf("%s funded the %s award", actor, h.Target), "funded the "+h.Target+" award you were going for")
		case HappeningColony:
			a.addRace(plan.TargetColonies, h, fmt.Sprintf("%s built a colony on %s", actor, h.Target), "built on the "+h.Target+" colony you wanted")
		case HappeningResourceLoss:
			a.Personal = true
			a.Lines = append(a.Lines, fmt.Sprintf("%s took away %d of your %s.", actor, h.Amount, h.Target))
			a.Hostile = append(a.Hostile, Hostility{ActorID: h.ActorID, What: fmt.Sprintf("took away %d of your %s", h.Amount, h.Target)})
		case HappeningCard:
			if h.Big {
				a.Big = true
				a.Lines = append(a.Lines, fmt.Sprintf("%s played %s.", actor, h.Target))
			}
		case HappeningChat:
			if soleOpponent || mentions(h.Detail, botName) {
				a.Personal = true
				a.Directed = a.Directed || !h.FromBot
				a.Lines = append(a.Lines, fmt.Sprintf("%s said to you: %q", actor, h.Detail))
			}
		}
	}
	return a
}

func (a *Assessment) addRace(targets []string, h Happening, line, hostile string) {
	if slices.Contains(targets, h.Target) {
		a.PlanHit = true
		a.Lines = append(a.Lines, line+", which you were going for.")
		a.Hostile = append(a.Hostile, Hostility{ActorID: h.ActorID, What: hostile})
		return
	}
	a.Big = true
	a.Lines = append(a.Lines, line+".")
}

func mentions(text, name string) bool {
	text = strings.ToLower(text)
	name = strings.ToLower(name)
	if name == "" {
		return false
	}
	if strings.Contains(text, name) {
		return true
	}
	first, _, _ := strings.Cut(name, " ")
	return len(first) >= 3 && strings.Contains(text, first)
}

// ReactionGate throttles reactions: one per cooldown, unless the bot was spoken to directly.
type ReactionGate struct {
	Cooldown   time.Duration
	BigChance  float64
	lastAction time.Time
}

// Allow decides whether to react now. roll is a uniform number in [0,1).
func (g *ReactionGate) Allow(a Assessment, now time.Time, roll float64) bool {
	if !a.Relevant() {
		return false
	}
	if !a.Directed && now.Sub(g.lastAction) < g.Cooldown {
		return false
	}
	if !a.PlanHit && !a.Personal && roll >= g.BigChance {
		return false
	}
	g.lastAction = now
	return true
}
