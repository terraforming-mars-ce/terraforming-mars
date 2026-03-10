package colony

import (
	"context"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// PendingResource represents a card-targeted resource that needs player selection
type PendingResource struct {
	PlayerID     string
	ResourceType string
	Amount       int
}

// applyOutput applies a colony output to a player's resources or production.
// Returns a PendingResource if the output is a card-targeted resource (microbe/animal/floater)
// that requires the player to choose which card to place it on.
func applyOutput(ctx context.Context, g *game.Game, p *player.Player, outputType string, amount int, log *zap.Logger) *PendingResource {
	rt := shared.ResourceType(outputType)
	switch rt {
	case shared.ResourceCredit, shared.ResourceSteel, shared.ResourceTitanium,
		shared.ResourcePlant, shared.ResourceEnergy, shared.ResourceHeat:
		p.Resources().Add(map[shared.ResourceType]int{rt: amount})
	case shared.ResourceCreditProduction, shared.ResourceSteelProduction,
		shared.ResourceTitaniumProduction, shared.ResourcePlantProduction,
		shared.ResourceEnergyProduction, shared.ResourceHeatProduction:
		p.Resources().AddProduction(map[shared.ResourceType]int{rt: amount})
	case shared.ResourceCardDraw:
		deck := g.Deck()
		if deck != nil {
			cardIDs, err := deck.DrawProjectCards(ctx, amount)
			if err != nil {
				log.Warn("Failed to draw cards for colony output", zap.Error(err))
				return nil
			}
			for _, cardID := range cardIDs {
				p.Hand().AddCard(cardID)
			}
			log.Debug("Drew cards for colony output",
				zap.String("player_id", p.ID()),
				zap.Int("count", len(cardIDs)))
		}
	case shared.ResourceMicrobe, shared.ResourceAnimal, shared.ResourceFloater:
		return &PendingResource{
			PlayerID:     p.ID(),
			ResourceType: outputType,
			Amount:       amount,
		}
	}
	return nil
}
