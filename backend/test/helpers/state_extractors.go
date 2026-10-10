package helpers

import (
	"testing"

	"openmars/internal/delivery/dto"
)

// ExtractPlayerIDs extracts player IDs from game state map
func ExtractPlayerIDs(t *testing.T, gameState map[string]interface{}) []string {
	if gameState == nil {
		return []string{}
	}

	var playerIDs []string

	// Check for currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok {
		if id, ok := currentPlayer["id"].(string); ok && id != "" {
			playerIDs = append(playerIDs, id)
		}
	}

	// Check for otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		for _, player := range otherPlayers {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if id, ok := playerMap["id"].(string); ok && id != "" {
					playerIDs = append(playerIDs, id)
				}
			}
		}
	}

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		for _, player := range players {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if id, ok := playerMap["id"].(string); ok && id != "" {
					playerIDs = append(playerIDs, id)
				}
			}
		}
	}

	// Remove duplicates (shouldn't happen, but safety check)
	uniqueIDs := make([]string, 0, len(playerIDs))
	seen := make(map[string]bool)
	for _, id := range playerIDs {
		if !seen[id] {
			uniqueIDs = append(uniqueIDs, id)
			seen[id] = true
		}
	}

	return uniqueIDs
}

// ExtractPlayerNames extracts player names from game state map
func ExtractPlayerNames(gameState map[string]interface{}) []string {
	if gameState == nil {
		return []string{}
	}

	var playerNames []string

	// Check for currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok {
		if name, ok := currentPlayer["name"].(string); ok && name != "" {
			playerNames = append(playerNames, name)
		}
	}

	// Check for otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		for _, player := range otherPlayers {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if name, ok := playerMap["name"].(string); ok && name != "" {
					playerNames = append(playerNames, name)
				}
			}
		}
	}

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		for _, player := range players {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if name, ok := playerMap["name"].(string); ok && name != "" {
					playerNames = append(playerNames, name)
				}
			}
		}
	}

	return playerNames
}

// CountPlayersInGameState counts the number of players in a game state
func CountPlayersInGameState(t *testing.T, gameState map[string]interface{}) int {
	if gameState == nil {
		t.Logf("Game state is nil, cannot count players")
		return 0
	}

	playerCount := 0

	// Check for currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok && currentPlayer["id"] != nil {
		playerCount++
	}

	// Check for otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		playerCount += len(otherPlayers)
	}

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		playerCount = len(players)
	}

	return playerCount
}

// ExtractGameStatus extracts the game status from a game state or WebSocket message
func ExtractGameStatus(data map[string]interface{}) (string, bool) {
	if status, ok := data["status"].(string); ok {
		return status, true
	}
	return "", false
}

// ExtractGamePhase extracts the game phase from a game state
func ExtractGamePhase(gameState map[string]interface{}) (string, bool) {
	if phase, ok := gameState["phase"].(string); ok {
		return phase, true
	}
	return "", false
}

// ExtractGameID extracts the game ID from a game state or message payload
func ExtractGameID(data map[string]interface{}) (string, bool) {
	if gameID, ok := data["gameId"].(string); ok && gameID != "" {
		return gameID, true
	}
	if id, ok := data["id"].(string); ok && id != "" {
		return id, true
	}
	return "", false
}

// ExtractGlobalParameters extracts global parameters from game state
func ExtractGlobalParameters(gameState map[string]interface{}) map[string]interface{} {
	if globalParams, ok := gameState["globalParameters"].(map[string]interface{}); ok {
		return globalParams
	}
	return nil
}

// ExtractResourcesForPlayer extracts resources for a specific player from game state
func ExtractResourcesForPlayer(gameState map[string]interface{}, playerID string) map[string]interface{} {
	// Check currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok {
		if id, ok := currentPlayer["id"].(string); ok && id == playerID {
			if resources, ok := currentPlayer["resources"].(map[string]interface{}); ok {
				return resources
			}
		}
	}

	// Check otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		for _, player := range otherPlayers {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if id, ok := playerMap["id"].(string); ok && id == playerID {
					if resources, ok := playerMap["resources"].(map[string]interface{}); ok {
						return resources
					}
				}
			}
		}
	}

	return nil
}

// ExtractGameFromMessage extracts game state from a WebSocket message payload
func ExtractGameFromMessage(msg dto.WebSocketMessage) (map[string]interface{}, bool) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, false
	}

	gameData, ok := payload["game"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	return gameData, true
}

// FindDuplicatePlayerNames finds duplicate names in a list of player names
func FindDuplicatePlayerNames(playerNames []string) []string {
	nameCount := make(map[string]int)
	var duplicates []string

	// Count occurrences of each name
	for _, name := range playerNames {
		nameCount[name]++
	}

	// Find names that appear more than once
	for name, count := range nameCount {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}

	return duplicates
}
