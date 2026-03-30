package game

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	internalgame "terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectDemoChoicesAction handles a player selecting cards during the demo lobby phase
type SelectDemoChoicesAction struct {
	gameRepo     internalgame.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewSelectDemoChoicesAction creates a new select demo choices action
func NewSelectDemoChoicesAction(
	gameRepo internalgame.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectDemoChoicesAction {
	return &SelectDemoChoicesAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute validates and stores a player's demo lobby card selections
func (a *SelectDemoChoicesAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	request *dto.SelectDemoChoicesRequest,
) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_demo_choices"),
	)
	log.Debug("Player selecting demo choices")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != shared.GameStatusLobby {
		return fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	if !g.Settings().DemoGame {
		return fmt.Errorf("game is not a demo game")
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("player not found: %s", playerID)
	}

	if request.CorporationID == "" {
		return fmt.Errorf("corporation ID is required")
	}
	corpCard, err := a.cardRegistry.GetByID(request.CorporationID)
	if err != nil {
		return fmt.Errorf("corporation not found: %s", request.CorporationID)
	}
	if corpCard.Type != "corporation" {
		return fmt.Errorf("card %s is not a corporation", request.CorporationID)
	}

	settings := g.Settings()
	if settings.HasPrelude() {
		if len(request.PreludeIDs) != 2 {
			return fmt.Errorf("must select exactly 2 prelude cards, got %d", len(request.PreludeIDs))
		}
		for _, id := range request.PreludeIDs {
			card, err := a.cardRegistry.GetByID(id)
			if err != nil {
				return fmt.Errorf("prelude card not found: %s", id)
			}
			if card.Type != "prelude" {
				return fmt.Errorf("card %s is not a prelude", id)
			}
		}
	} else if len(request.PreludeIDs) > 0 {
		return fmt.Errorf("prelude cards not enabled for this game")
	}

	for _, id := range request.CardIDs {
		card, err := a.cardRegistry.GetByID(id)
		if err != nil {
			return fmt.Errorf("card not found: %s", id)
		}
		if card.Type == "corporation" || card.Type == "prelude" {
			return fmt.Errorf("card %s is a %s, not a project card", id, card.Type)
		}
	}

	p.SetPendingDemoChoices(&shared.PendingDemoChoices{
		CorporationID: request.CorporationID,
		PreludeIDs:    request.PreludeIDs,
		CardIDs:       request.CardIDs,
		Resources: shared.Resources{
			Credits:  request.Resources.Credits,
			Steel:    request.Resources.Steel,
			Titanium: request.Resources.Titanium,
			Plants:   request.Resources.Plants,
			Energy:   request.Resources.Energy,
			Heat:     request.Resources.Heat,
		},
		Production: shared.Production{
			Credits:  request.Production.Credits,
			Steel:    request.Production.Steel,
			Titanium: request.Production.Titanium,
			Plants:   request.Production.Plants,
			Energy:   request.Production.Energy,
			Heat:     request.Production.Heat,
		},
		TerraformRating: request.TerraformRating,
	})

	isHost := g.HostPlayerID() == playerID
	if isHost && request.GlobalParameters != nil {
		settings.Temperature = &request.GlobalParameters.Temperature
		settings.Oxygen = &request.GlobalParameters.Oxygen
		settings.Oceans = &request.GlobalParameters.Oceans
	}
	if isHost && request.Generation != nil {
		settings.Generation = request.Generation
	}
	if isHost && (request.GlobalParameters != nil || request.Generation != nil) {
		g.UpdateSettings(ctx, settings)
	}

	log.Info("Demo choices selected",
		zap.String("corporation", corpCard.Name),
		zap.Int("prelude_count", len(request.PreludeIDs)),
		zap.Int("card_count", len(request.CardIDs)))

	return nil
}
