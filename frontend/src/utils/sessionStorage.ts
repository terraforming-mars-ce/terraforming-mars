/**
 * Utility functions for managing game session data in localStorage
 */

const STORAGE_KEY = "terraforming-mars-game";

export interface StoredGameData {
  gameId: string;
  playerId: string;
  playerName: string;
  isSpectator?: boolean;
  joinedAt?: string;
  timestamp?: number;
}

/**
 * Clears the saved game session from localStorage
 */
export function clearGameSession(): void {
  localStorage.removeItem(STORAGE_KEY);
}

/**
 * Retrieves the saved game session from localStorage
 */
export function getGameSession(): StoredGameData | null {
  const storedData = localStorage.getItem(STORAGE_KEY);
  if (!storedData) {
    return null;
  }

  try {
    return JSON.parse(storedData) as StoredGameData;
  } catch (error) {
    console.error("Failed to parse stored game data:", error);
    clearGameSession(); // Clear corrupted data
    return null;
  }
}

/**
 * Saves game session data to localStorage
 */
export function saveGameSession(data: StoredGameData): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
}
