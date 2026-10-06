package bot

import (
	"encoding/json"
	"fmt"
	rand "math/rand/v2"
	"os"
)

// Persona is a bot personality: its voice in chat and the names that carry it.
type Persona struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	Voice string   `json:"voice"`
	Names []string `json:"names"`
}

// PersonaCatalog holds all personas and maps each bot name to its persona.
type PersonaCatalog struct {
	personas []Persona
	byID     map[string]Persona
}

// LoadPersonaCatalog reads the persona catalog from a JSON file.
func LoadPersonaCatalog(path string) (*PersonaCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read persona catalog: %w", err)
	}
	var file struct {
		Personas []Persona `json:"personas"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse persona catalog: %w", err)
	}
	return NewPersonaCatalog(file.Personas)
}

// NewPersonaCatalog builds a catalog from personas, rejecting duplicates and empty entries.
func NewPersonaCatalog(personas []Persona) (*PersonaCatalog, error) {
	if len(personas) == 0 {
		return nil, fmt.Errorf("persona catalog is empty")
	}
	byID := make(map[string]Persona, len(personas))
	names := map[string]bool{}
	for _, p := range personas {
		if p.ID == "" || p.Voice == "" || len(p.Names) == 0 {
			return nil, fmt.Errorf("persona %q is incomplete", p.ID)
		}
		if _, dup := byID[p.ID]; dup {
			return nil, fmt.Errorf("duplicate persona %q", p.ID)
		}
		for _, n := range p.Names {
			if names[n] {
				return nil, fmt.Errorf("bot name %q used by two personas", n)
			}
			names[n] = true
		}
		byID[p.ID] = p
	}
	return &PersonaCatalog{personas: personas, byID: byID}, nil
}

// Get returns the persona with the given ID, falling back to the first persona.
func (c *PersonaCatalog) Get(id string) Persona {
	if p, ok := c.byID[id]; ok {
		return p
	}
	return c.personas[0]
}

// botIdentityRNGStream keeps bot identity selection independent of the deck and setup RNG streams.
const botIdentityRNGStream uint64 = 0xB07

// AssignIdentity picks a free bot name and its persona, deterministically from the game seed.
// takenNames are the names of players already in the game.
func (c *PersonaCatalog) AssignIdentity(seed uint64, takenNames []string) (name, personaID string) {
	taken := make(map[string]bool, len(takenNames))
	for _, n := range takenNames {
		taken[n] = true
	}
	type candidate struct{ name, persona string }
	var all []candidate
	for _, p := range c.personas {
		for _, n := range p.Names {
			all = append(all, candidate{n, p.ID})
		}
	}
	rng := rand.New(rand.NewPCG(seed, botIdentityRNGStream+uint64(len(takenNames))))
	for _, i := range rng.Perm(len(all)) {
		if !taken[all[i].name] {
			return all[i].name, all[i].persona
		}
	}
	persona := c.personas[rng.IntN(len(c.personas))]
	for i := 1; ; i++ {
		n := fmt.Sprintf("%s %d", persona.Names[0], i)
		if !taken[n] {
			return n, persona.ID
		}
	}
}

// AssignPersona picks a persona for an existing player turned into a bot.
func (c *PersonaCatalog) AssignPersona(hash uint64) string {
	return c.personas[hash%uint64(len(c.personas))].ID
}
