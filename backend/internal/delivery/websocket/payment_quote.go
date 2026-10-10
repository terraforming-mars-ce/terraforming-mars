package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"openmars/internal/action"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
)

type paymentQuoteHandler struct{ broadcaster *Broadcaster }

func (h *paymentQuoteHandler) HandleMessage(ctx context.Context, c *core.Connection, message dto.WebSocketMessage) {
	var request struct {
		RequestID string               `json:"requestId"`
		Intent    action.PaymentIntent `json:"intent"`
	}
	response := map[string]any{}
	err := func() error {
		raw, err := json.Marshal(message.Payload)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(raw, &request); err != nil {
			return err
		}
		id := c.Identity()
		if !id.IsPlayer() {
			return fmt.Errorf("not connected to a game")
		}
		b := h.broadcaster
		g, err := b.gameRepo.Get(ctx, id.GameID)
		if err != nil {
			return err
		}
		p, err := g.GetPlayer(id.PlayerID)
		if err != nil {
			return err
		}
		quote, err := action.QuoteActionPayment(g, p, b.cardRegistry, b.standardProjectRegistry, b.milestoneRegistry, b.awardRegistry, request.Intent)
		if err != nil {
			return err
		}
		response["quote"] = dto.ToPaymentQuoteDto(quote)
		return nil
	}()
	response["requestId"] = request.RequestID
	if err != nil {
		response["error"] = err.Error()
	}
	c.Send(dto.WebSocketMessage{Type: dto.MessageTypePaymentQuote, GameID: c.GameID(), Payload: response})
}
