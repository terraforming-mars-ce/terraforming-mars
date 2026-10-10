package harness

import (
	"testing"

	"openmars/internal/delivery/dto"
)

// Admin runs an admin command, which requires a development-mode game, and fails the
// test if the server rejects it.
func (c *Client) Admin(t testing.TB, command dto.AdminCommandType, payload map[string]any) {
	t.Helper()
	c.Send(t, dto.MessageTypeAdminCommand, map[string]any{"commandType": command, "payload": payload})
	for _, msg := range c.Sync(t) {
		if msg.Type == dto.MessageTypeError {
			t.Fatalf("%s: admin %s rejected: %s", c.Name, command, msg.Payload)
		}
	}
}

// DevMode turns on development mode, which allows admin commands.
func DevMode(s *dto.GameSetupDto) {
	s.DevelopmentMode = true
}
