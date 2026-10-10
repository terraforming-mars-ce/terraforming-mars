package bot_test

import (
	"testing"
	"time"

	"openmars/internal/service/bot"
	"openmars/test/testutil"
)

func nameOf(id string) string { return map[string]string{"h1": "Alice", "h2": "Bob"}[id] }

func TestAssessHappenings(t *testing.T) {
	plan := bot.Plan{
		Summary:          "Cities by the water",
		WantedHexes:      []string{"1,-1,0"},
		TargetMilestones: []string{"mayor"},
		TargetColonies:   []string{"luna"},
	}
	tests := []struct {
		name       string
		happening  bot.Happening
		sole       bool
		planHit    bool
		personal   bool
		directed   bool
		big        bool
		noReaction bool
	}{
		{name: "tile on a wanted hex", happening: bot.Happening{Kind: bot.HappeningTile, ActorID: "h1", Target: "1,-1,0", Detail: "city"}, planHit: true},
		{name: "tile elsewhere", happening: bot.Happening{Kind: bot.HappeningTile, ActorID: "h1", Target: "3,-3,0", Detail: "city"}, noReaction: true},
		{name: "own tile", happening: bot.Happening{Kind: bot.HappeningTile, ActorID: "bot", Target: "1,-1,0"}, noReaction: true},
		{name: "target milestone taken", happening: bot.Happening{Kind: bot.HappeningMilestone, ActorID: "h1", Target: "mayor"}, planHit: true},
		{name: "other milestone", happening: bot.Happening{Kind: bot.HappeningMilestone, ActorID: "h1", Target: "gardener"}, big: true},
		{name: "target colony built", happening: bot.Happening{Kind: bot.HappeningColony, ActorID: "h2", Target: "luna"}, planHit: true},
		{name: "resources taken", happening: bot.Happening{Kind: bot.HappeningResourceLoss, ActorID: "h1", Target: "plant", Amount: 3}, personal: true},
		{name: "big card", happening: bot.Happening{Kind: bot.HappeningCard, ActorID: "h1", Target: "Giant Ice Asteroid", Big: true}, big: true},
		{name: "small card", happening: bot.Happening{Kind: bot.HappeningCard, ActorID: "h1", Target: "Lichen"}, noReaction: true},
		{name: "chat mentioning the bot", happening: bot.Happening{Kind: bot.HappeningChat, ActorID: "h1", Detail: "nice move hal"}, personal: true, directed: true},
		{name: "chat to someone else", happening: bot.Happening{Kind: bot.HappeningChat, ActorID: "h1", Detail: "bob you sneak"}, noReaction: true},
		{name: "any chat with a sole opponent", happening: bot.Happening{Kind: bot.HappeningChat, ActorID: "h1", Detail: "gg"}, sole: true, personal: true, directed: true},
		{name: "another bot mentioning the bot", happening: bot.Happening{Kind: bot.HappeningChat, ActorID: "h2", Detail: "hal, nice try", FromBot: true}, personal: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := bot.AssessHappenings(plan, "bot", "HAL 9000", tt.sole, []bot.Happening{tt.happening}, nameOf)
			testutil.AssertEqual(t, tt.planHit, a.PlanHit, "plan hit")
			testutil.AssertEqual(t, tt.personal, a.Personal, "personal")
			testutil.AssertEqual(t, tt.directed, a.Directed, "directed")
			testutil.AssertEqual(t, tt.big, a.Big, "big")
			testutil.AssertEqual(t, !tt.noReaction, a.Relevant(), "relevant")
		})
	}
}

func TestReactionGate(t *testing.T) {
	now := time.Now()
	gate := bot.ReactionGate{Cooldown: 30 * time.Second, BigChance: 0.25}
	planHit := bot.Assessment{PlanHit: true}
	big := bot.Assessment{Big: true}
	directed := bot.Assessment{Personal: true, Directed: true}

	testutil.AssertTrue(t, gate.Allow(planHit, now, 0.9), "a plan hit reacts regardless of the roll")
	testutil.AssertFalse(t, gate.Allow(planHit, now.Add(10*time.Second), 0), "cooldown blocks the next reaction")
	testutil.AssertTrue(t, gate.Allow(directed, now.Add(11*time.Second), 0.9), "being spoken to bypasses the cooldown")
	testutil.AssertFalse(t, gate.Allow(big, now.Add(2*time.Minute), 0.9), "big moments react only on a lucky roll")
	testutil.AssertTrue(t, gate.Allow(big, now.Add(2*time.Minute), 0.1), "big moments react on a lucky roll")
	testutil.AssertFalse(t, gate.Allow(bot.Assessment{}, now.Add(time.Hour), 0), "nothing relevant, no reaction")
}

func TestParseReaction(t *testing.T) {
	tests := []struct {
		reply string
		want  bot.Reaction
	}{
		{`{"kind":"emote","emote":"angry"}`, bot.Reaction{Kind: "emote", Emote: "angry"}},
		{"Sure! {\"kind\":\"chat\",\"text\":\"  Hey,   that was MY spot. \"}", bot.Reaction{Kind: "chat", Text: "Hey, that was MY spot."}},
		{`{"kind":"emote","emote":"<script>"}`, bot.Reaction{Kind: "pass"}},
		{`{"kind":"chat","text":""}`, bot.Reaction{Kind: "pass"}},
		{`{"kind":"run","text":"rm"}`, bot.Reaction{Kind: "pass"}},
		{`no json at all`, bot.Reaction{Kind: "pass"}},
	}
	for _, tt := range tests {
		testutil.AssertEqual(t, tt.want, bot.ParseReaction(tt.reply), "reply: "+tt.reply)
	}
}

func TestAssessHappenings_RecordsHostileActs(t *testing.T) {
	plan := bot.Plan{WantedHexes: []string{"1,-1,0"}, TargetMilestones: []string{"mayor"}}
	a := bot.AssessHappenings(plan, "bot", "Wall-E", false, []bot.Happening{
		{Kind: bot.HappeningResourceLoss, ActorID: "h1", Target: "plants", Amount: 6},
		{Kind: bot.HappeningTile, ActorID: "h1", Target: "1,-1,0", Detail: "city"},
		{Kind: bot.HappeningMilestone, ActorID: "h2", Target: "mayor"},
		{Kind: bot.HappeningMilestone, ActorID: "h2", Target: "gardener"},
	}, nameOf)
	testutil.AssertEqual(t, 3, len(a.Hostile), "loss, wanted hex and targeted milestone are hostile; an untargeted milestone is not")
	testutil.AssertEqual(t, "took away 6 of your plants", a.Hostile[0].What, "loss is described from the bot's side")
	testutil.AssertEqual(t, "h2", a.Hostile[2].ActorID, "the actor is kept")
}
