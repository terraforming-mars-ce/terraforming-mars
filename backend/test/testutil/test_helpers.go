package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// TestContext provides a reusable test context
func TestContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "test", true)
}

// TestLogger creates a test logger (no-op or minimal output)
func TestLogger() *zap.Logger {
	return logger.Get()
}

// MockBroadcaster records broadcast calls for test assertions.
type MockBroadcaster struct {
	BroadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	GameID    string
	PlayerIDs []string
	Timestamp time.Time
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		BroadcastCalls: make([]BroadcastCall, 0),
	}
}

func (m *MockBroadcaster) CallCount() int {
	return len(m.BroadcastCalls)
}

func (m *MockBroadcaster) Reset() {
	m.BroadcastCalls = make([]BroadcastCall, 0)
}

// CreateTestCardRegistry creates a card registry with test cards
func CreateTestCardRegistry() cards.CardRegistry {
	testCards := []gamecards.Card{
		// Corporations (need at least 8 for 4 players getting 2 each)
		{
			ID:   "corp-tharsis-republic",
			Name: "Tharsis Republic",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-mining-guild",
			Name: "Mining Guild",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-inventrix",
			Name: "Inventrix",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-credicor",
			Name: "Credicor",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-ecoline",
			Name: "Ecoline",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Outputs: []shared.ResourceCondition{
						{
							ResourceType: shared.ResourceDiscount,
							Amount:       1,
							Target:       "self-player",
							Selectors: []shared.Selector{
								{Resources: []string{"plant"}, StandardProjects: []shared.StandardProject{shared.StandardProjectConvertPlantsToGreenery}},
							},
						},
					},
				},
			},
		},
		{
			ID:   "corp-helion",
			Name: "Helion",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-interplanetary-cinematics",
			Name: "Interplanetary Cinematics",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-phobolog",
			Name: "Phobolog",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-thorgate",
			Name: "ThorGate",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
			Tags: []shared.CardTag{shared.TagPower},
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Outputs: []shared.ResourceCondition{
						{
							ResourceType: shared.ResourceDiscount,
							Amount:       3,
							Target:       "self-player",
							Selectors: []shared.Selector{
								{Tags: []shared.CardTag{shared.TagPower}},
								{StandardProjects: []shared.StandardProject{shared.StandardProjectPowerPlant}},
							},
						},
					},
				},
			},
		},
		{
			ID:   "corp-saturn-systems",
			Name: "Saturn Systems",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-arklight",
			Name: "Arklight",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
			Tags: []shared.CardTag{shared.TagAnimal},
			ResourceStorage: &gamecards.ResourceStorage{
				Type:     shared.ResourceAnimal,
				Starting: 0,
			},
			VPConditions: []gamecards.VictoryPointCondition{
				{
					Amount:    1,
					Condition: gamecards.VPConditionPer,
					Per: &gamecards.PerCondition{
						Type:   shared.ResourceAnimal,
						Amount: 2,
						Target: ptrTargetType(gamecards.TargetSelfCard),
					},
				},
			},
		},
		{
			ID:   "corp-teractor",
			Name: "Teractor",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
			Tags: []shared.CardTag{shared.TagEarth},
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Outputs: []shared.ResourceCondition{
						{
							ResourceType: shared.ResourceDiscount,
							Amount:       3,
							Target:       "self-player",
							Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagEarth}}},
						},
					},
				},
			},
		},
		// Project cards (need at least 40 for typical games with 4 players getting 10 cards each)
		{
			ID:   "card-power-plant",
			Name: "Power Plant",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-asteroid",
			Name: "Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 14,
			Tags: []shared.CardTag{shared.TagSpace, shared.TagEvent},
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: "auto"}},
					Outputs: []shared.ResourceCondition{
						{ResourceType: shared.ResourcePlant, Amount: 3, Target: "any-player"},
					},
				},
				{
					Triggers: []shared.Trigger{{Type: "auto"}},
					Outputs: []shared.ResourceCondition{
						{ResourceType: shared.ResourceTitanium, Amount: 2, Target: "self-player"},
						{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "none"},
					},
				},
			},
		},
		{
			ID:   "card-water-import",
			Name: "Water Import",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-ai-central",
			Name: "AI Central",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 21,
		},
		{
			ID:   "card-aquifer-pumping",
			Name: "Aquifer Pumping",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 18,
		},
		{
			ID:   "card-asteroid-mining",
			Name: "Asteroid Mining",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 30,
		},
		{
			ID:   "card-asteroid-mining-consortium",
			Name: "Asteroid Mining Consortium",
			Type: gamecards.CardTypeAutomated,
			Pack: "corporate-era",
			Cost: 13,
			Tags: []shared.CardTag{shared.TagJovian},
			Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
				{Type: "production", Min: func() *int { v := 1; return &v }(), Resource: func() *shared.ResourceType { v := shared.ResourceTitaniumProduction; return &v }()},
			}},
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: "auto"}},
					Outputs: []shared.ResourceCondition{
						{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
					},
				},
				{
					Triggers: []shared.Trigger{{Type: "auto"}},
					Outputs: []shared.ResourceCondition{
						{ResourceType: shared.ResourceTitaniumProduction, Amount: -1, Target: "any-player"},
					},
				},
			},
		},
		{
			ID:   "card-biomass-combustors",
			Name: "Biomass Combustors",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-building-industries",
			Name: "Building Industries",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-capital",
			Name: "Capital",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 26,
		},
		{
			ID:   "card-carbonate-processing",
			Name: "Carbonate Processing",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-cartel",
			Name: "Cartel",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-colonizer-training-camp",
			Name: "Colonizer Training Camp",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-comet",
			Name: "Comet",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 21,
		},
		{
			ID:   "card-research",
			Name: "Research",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 11,
		},
		{
			ID:   "card-development-center",
			Name: "Development Center",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 11,
		},
		{
			ID:   "card-dust-seals",
			Name: "Dust Seals",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 2,
		},
		{
			ID:          "card-earth-catapult",
			Name:        "Earth Catapult",
			Type:        gamecards.CardTypeActive,
			Pack:        "base",
			Cost:        23,
			Tags:        []shared.CardTag{shared.TagEarth},
			Description: "Effect: When you play a card, you pay 2 M€ less for it.",
		},
		{
			ID:          "card-earth-office",
			Name:        "Earth Office",
			Type:        gamecards.CardTypeActive,
			Pack:        "base",
			Cost:        1,
			Tags:        []shared.CardTag{shared.TagEarth},
			Description: "Effect: When you play an Earth tag, you pay 3 M€ less for it.",
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Outputs: []shared.ResourceCondition{
						{
							ResourceType: shared.ResourceDiscount,
							Amount:       3,
							Target:       "self-player",
							Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagEarth}}},
						},
					},
				},
			},
		},
		{
			ID:          "card-sponsors",
			Name:        "Sponsors",
			Type:        gamecards.CardTypeAutomated,
			Pack:        "base",
			Cost:        6,
			Tags:        []shared.CardTag{shared.TagEarth},
			Description: "Increase your M€ production 2 steps.",
		},
		{
			ID:   "card-energy-tapping",
			Name: "Energy Tapping",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 3,
		},
		{
			ID:   "card-eos-chasma-national-park",
			Name: "Eos Chasma National Park",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 16,
		},
		{
			ID:   "card-extreme-cold-fungus",
			Name: "Extreme-Cold Fungus",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 13,
		},
		{
			ID:   "card-birds",
			Name: "Birds",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 10,
			Tags: []shared.CardTag{shared.TagAnimal},
			ResourceStorage: &gamecards.ResourceStorage{
				Type:     shared.ResourceAnimal,
				Starting: 0,
			},
			VPConditions: []gamecards.VictoryPointCondition{
				{
					Amount:    1,
					Condition: gamecards.VPConditionPer,
					Per: &gamecards.PerCondition{
						Type:   shared.ResourceAnimal,
						Amount: 1,
						Target: ptrTargetType(gamecards.TargetSelfCard),
					},
				},
			},
		},
		{
			ID:   "card-fish",
			Name: "Fish",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 9,
		},
		{
			ID:   "card-food-factory",
			Name: "Food Factory",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-fuel-factory",
			Name: "Fuel Factory",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 7,
		},
		{
			ID:   "card-fusion-power",
			Name: "Fusion Power",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 14,
		},
		{
			ID:   "card-ganymede-colony",
			Name: "Ganymede Colony",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 20,
			Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
			VPConditions: []gamecards.VictoryPointCondition{
				{
					Amount:    1,
					Condition: gamecards.VPConditionPer,
					Per: &gamecards.PerCondition{
						Type:   shared.ResourceType(shared.TagJovian),
						Amount: 1,
						Tag:    ptrCardTag(shared.TagJovian),
					},
				},
			},
		},
		{
			ID:   "card-gene-repair",
			Name: "Gene Repair",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-giant-ice-asteroid",
			Name: "Giant Ice Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 36,
		},
		{
			ID:   "card-giant-space-mirror",
			Name: "Giant Space Mirror",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 17,
		},
		{
			ID:   "card-greenhouse",
			Name: "Greenhouse",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-heather",
			Name: "Heather",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-heat-trappers",
			Name: "Heat Trappers",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-herbivores",
			Name: "Herbivores",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-hiring-grant",
			Name: "Hiring Grant",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-ice-asteroid",
			Name: "Ice Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 23,
		},
		{
			ID:   "card-io-mining-industries",
			Name: "Io Mining Industries",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 41,
			Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		},
		{
			ID:   "card-immigration-shuttles",
			Name: "Immigration Shuttles",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 31,
			Tags: []shared.CardTag{shared.TagSpace},
			VPConditions: []gamecards.VictoryPointCondition{
				{
					Amount:    1,
					Condition: gamecards.VPConditionPer,
					Per: &gamecards.PerCondition{
						Type:   shared.ResourceCityTile,
						Amount: 3,
					},
				},
			},
		},
		{
			ID:   "card-imported-ghg",
			Name: "Imported GHG",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 7,
		},
		{
			ID:   "card-imported-hydrogen",
			Name: "Imported Hydrogen",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 16,
		},
		{
			ID:   "card-imported-nitrogen",
			Name: "Imported Nitrogen",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 23,
		},
		{
			ID:   "card-industrial-center",
			Name: "Industrial Center",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-insects",
			Name: "Insects",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 9,
		},
		{
			ID:   "card-investment-loan",
			Name: "Investment Loan",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 3,
		},
		{
			ID:   "card-kelp-farming",
			Name: "Kelp Farming",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 17,
		},
		{
			ID:          "card-hired-raiders",
			Name:        "Hired Raiders",
			Type:        gamecards.CardTypeEvent,
			Pack:        "corporate-era",
			Cost:        1,
			Description: "Steal up to 2 steel, or 3 M€ from any player.",
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Choices: []shared.Choice{
						{
							Outputs: []shared.ResourceCondition{
								{ResourceType: shared.ResourceSteel, Amount: 2, Target: "steal-any-player"},
							},
						},
						{
							Outputs: []shared.ResourceCondition{
								{ResourceType: shared.ResourceCredit, Amount: 3, Target: "steal-any-player"},
							},
						},
					},
				},
			},
		},
		// Preludes
		{
			ID:   "prelude-allied-banks",
			Name: "Allied Banks",
			Type: gamecards.CardTypePrelude,
			Pack: "prelude",
		},
		// Cards with discount effects for testing requirement modifiers
		{
			ID:          "card-space-station",
			Name:        "Space Station",
			Type:        gamecards.CardTypeActive,
			Pack:        "corporate-era",
			Cost:        10,
			Tags:        []shared.CardTag{shared.TagSpace},
			Description: "Effect: When you play a space card, you pay 2 M€ less for it.",
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: "auto"}},
					Outputs: []shared.ResourceCondition{
						{
							ResourceType: shared.ResourceDiscount,
							Amount:       2,
							Selectors:    []shared.Selector{{Tags: []shared.CardTag{shared.TagSpace}}},
						},
					},
				},
			},
		},
		{
			ID:          "card-space-mirrors",
			Name:        "Space Mirrors",
			Type:        gamecards.CardTypeActive,
			Pack:        "base",
			Cost:        3,
			Tags:        []shared.CardTag{shared.TagSpace, shared.TagPower},
			Description: "Action: Spend 7 M€ to increase your energy production 1 step.",
		},
		{
			ID:          "card-arctic-algae",
			Name:        "Arctic Algae",
			Type:        gamecards.CardTypeActive,
			Pack:        "base",
			Cost:        12,
			Tags:        []shared.CardTag{shared.TagPlant},
			Description: "Test card without space tag for discount testing.",
		},
		// Test card with only power tag for ThorGate testing
		{
			ID:          "card-deep-well-heating",
			Name:        "Deep Well Heating",
			Type:        gamecards.CardTypeAutomated,
			Pack:        "base",
			Cost:        12,
			Tags:        []shared.CardTag{shared.TagPower, shared.TagBuilding},
			Description: "Increase temperature 1 step. Increase your energy production 1 step.",
		},
		// Card with choice for testing choice-based cards
		{
			ID:          "card-artificial-photosynthesis",
			Name:        "Artificial Photosynthesis",
			Type:        gamecards.CardTypeAutomated,
			Pack:        "base",
			Cost:        12,
			Tags:        []shared.CardTag{shared.TagScience},
			Description: "Increase your plant production 1 step or your energy production 2 steps.",
			Behaviors: []shared.CardBehavior{
				{
					Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
					Choices: []shared.Choice{
						{
							Outputs: []shared.ResourceCondition{
								{
									ResourceType: shared.ResourcePlantProduction,
									Amount:       1,
									Target:       "self-player",
								},
							},
						},
						{
							Outputs: []shared.ResourceCondition{
								{
									ResourceType: shared.ResourceEnergyProduction,
									Amount:       2,
									Target:       "self-player",
								},
							},
						},
					},
				},
			},
		},
	}
	testCards = append(testCards, gamecards.Card{
		ID:          "card-insulation",
		Name:        "Insulation",
		Type:        gamecards.CardTypeAutomated,
		Pack:        "base",
		Cost:        2,
		Description: "Decrease your heat production any number of steps and increase your M€ production the same number of steps.",
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: -1, Target: "self-player", VariableAmount: true},
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player", VariableAmount: true},
				},
			},
		},
	})

	// Indentured Workers: next card costs 8 M€ less (temporary discount)
	testCards = append(testCards, gamecards.Card{
		ID:          "card-indentured-workers",
		Name:        "Indentured Workers",
		Type:        gamecards.CardTypeEvent,
		Pack:        "corporate-era",
		Cost:        0,
		Description: "The next card you play this generation costs 8 M€ less.",
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceDiscount, Amount: 8, Target: "self-player", Temporary: shared.TemporaryNextCard},
				},
			},
		},
	})

	// Special Design: next card has +/- 2 global parameter lenience (temporary)
	testCards = append(testCards, gamecards.Card{
		ID:          "card-special-design",
		Name:        "Special Design",
		Type:        gamecards.CardTypeEvent,
		Pack:        "base",
		Cost:        4,
		Tags:        []shared.CardTag{shared.TagScience},
		Description: "The next card you play this generation is +2 or -2 in global requirements, your choice.",
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceGlobalParameterLenience, Amount: 2, Target: "self-player", Temporary: shared.TemporaryNextCard},
				},
			},
		},
	})

	// Test card with temperature requirement for testing lenience
	minTemp := -24
	testCards = append(testCards, gamecards.Card{
		ID:          "card-temp-req-test",
		Name:        "Temperature Requirement Test",
		Type:        gamecards.CardTypeAutomated,
		Pack:        "base",
		Cost:        5,
		Description: "Requires temperature >= -24°C.",
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: &minTemp},
		}},
	})

	// Test card with max oxygen requirement for testing lenience
	maxOxygen := 5
	testCards = append(testCards, gamecards.Card{
		ID:          "card-oxygen-max-req-test",
		Name:        "Max Oxygen Requirement Test",
		Type:        gamecards.CardTypeAutomated,
		Pack:        "base",
		Cost:        5,
		Description: "Requires oxygen <= 5%.",
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementOxygen, Max: &maxOxygen},
		}},
	})

	// Venus-tagged card for testing storage payment substitutes (Dirigibles)
	testCards = append(testCards, gamecards.Card{
		ID:          "card-venus-test",
		Name:        "Venus Test Card",
		Type:        gamecards.CardTypeAutomated,
		Pack:        "base",
		Cost:        7,
		Tags:        []shared.CardTag{shared.TagVenus},
		Description: "A Venus-tagged card for testing.",
	})

	return cards.NewInMemoryCardRegistry(testCards)
}

// CreateTestCardRegistryWithAdditionalCards creates a card registry with standard test cards plus additional cards
func CreateTestCardRegistryWithAdditionalCards(additionalCards []gamecards.Card) cards.CardRegistry {
	baseRegistry := CreateTestCardRegistry()
	allCards := baseRegistry.GetAll()
	allCards = append(allCards, additionalCards...)
	return cards.NewInMemoryCardRegistry(allCards)
}

// CreateTestGameWithPlayers creates a game with specified number of players
func CreateTestGameWithPlayers(t *testing.T, numPlayers int, broadcaster *MockBroadcaster) (*game.Game, game.GameRepository) {
	t.Helper()

	repo := game.NewInMemoryGameRepository()
	cardRegistry := CreateTestCardRegistry()

	// Create game
	settings := game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame("test-game-id", "", settings)
	allCards := cardRegistry.GetAll()

	// Separate cards by type
	projectCards := make([]string, 0)
	corpCards := make([]string, 0)
	preludeCards := make([]string, 0)

	for _, card := range allCards {
		switch card.Type {
		case gamecards.CardTypeCorporation:
			corpCards = append(corpCards, card.ID)
		case gamecards.CardTypePrelude:
			preludeCards = append(preludeCards, card.ID)
		default:
			projectCards = append(projectCards, card.ID)
		}
	}

	// Create and set deck
	gameDeck := deck.NewDeck(testGame.ID(), projectCards, corpCards, preludeCards)
	testGame.SetDeck(gameDeck)
	testGame.SetVPCardLookup(cards.NewVPCardLookupAdapter(cardRegistry))

	err := repo.Create(context.Background(), testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Add players
	ctx := context.Background()
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("player-%d", i+1)
		playerName := "Player " + string(rune('A'+i))

		// Create player
		newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), playerID, playerName)

		// Add to game
		err := testGame.AddPlayer(ctx, newPlayer)
		if err != nil {
			t.Fatalf("Failed to add player %d: %v", i, err)
		}
	}

	// Set first player as host (like JoinGameAction does)
	if numPlayers > 0 {
		if err := testGame.SetHostPlayerID(ctx, "player-1"); err != nil {
			t.Fatalf("Failed to set host player: %v", err)
		}
	}

	return testGame, repo
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", message)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected == actual {
		t.Fatalf("%s: expected not equal to %v", message, expected)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true, got false", message)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false, got true", message)
	}
}

// IntPtr returns a pointer to the given int value.
func IntPtr(v int) *int { return &v }

// StrPtr returns a pointer to the given string value.
func StrPtr(v string) *string { return &v }

// TagPtr returns a pointer to the given CardTag value.
func TagPtr(v shared.CardTag) *shared.CardTag { return &v }

// ResourceTypePtr returns a pointer to the given ResourceType value.
func ResourceTypePtr(v shared.ResourceType) *shared.ResourceType { return &v }

func ptrTargetType(t gamecards.TargetType) *gamecards.TargetType {
	return &t
}

func ptrCardTag(t shared.CardTag) *shared.CardTag {
	return &t
}
