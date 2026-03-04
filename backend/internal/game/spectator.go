package game

// MaxSpectators is the maximum number of spectators allowed per game.
const MaxSpectators = 4

// PlayerColors is the palette of 10 visually distinct colors available to players.
var PlayerColors = []string{
	"#e53935", "#1e88e5", "#43a047", "#ffb300", "#8e24aa",
	"#00acc1", "#f4511e", "#3949ab", "#c0ca33", "#d81b60",
}

// SpectatorColors is the palette of colors assigned to spectators.
var SpectatorColors = []string{"#9b9b9b", "#7eb8da", "#c4a6e8", "#e8c4a6"}

// Spectator represents a lightweight, ephemeral observer of a game.
type Spectator struct {
	id    string
	name  string
	color string
}

// NewSpectator creates a new spectator with the given ID, name, and color.
func NewSpectator(id, name, color string) *Spectator {
	return &Spectator{
		id:    id,
		name:  name,
		color: color,
	}
}

// ID returns the spectator's unique identifier.
func (s *Spectator) ID() string {
	return s.id
}

// Name returns the spectator's display name.
func (s *Spectator) Name() string {
	return s.name
}

// Color returns the spectator's assigned color.
func (s *Spectator) Color() string {
	return s.color
}
