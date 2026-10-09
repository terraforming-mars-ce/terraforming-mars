import { CardDto } from "../types/generated/api-types.ts";
import { config } from "../config.ts";

let cardCache: Map<string, CardDto> | null = null;
let corporationCache: CardDto[] | null = null;

/**
 * Fetches all cards from the backend and caches them
 */
export async function fetchAllCards(): Promise<Map<string, CardDto>> {
  if (cardCache !== null) {
    return cardCache;
  }

  try {
    const response = await fetch(`${config.apiUrl}/cards?limit=500`);
    if (!response.ok) {
      throw new Error(`Failed to fetch cards: ${response.statusText}`);
    }

    const data = await response.json();
    const cards = data.cards || [];

    cardCache = new Map();
    cards.forEach((card: CardDto) => {
      cardCache!.set(card.id, card);
    });

    return cardCache;
  } catch (error) {
    console.error("Failed to fetch cards:", error);
    return new Map();
  }
}

/**
 * Gets a card by its ID from the cache
 */
export async function getCardById(cardId: string): Promise<CardDto | null> {
  const cards = await fetchAllCards();
  return cards.get(cardId) || null;
}

/**
 * Fetches all corporations from the backend and caches them
 */
export async function fetchCorporations(): Promise<CardDto[]> {
  if (corporationCache !== null) {
    return corporationCache;
  }

  try {
    const response = await fetch(`${config.apiUrl}/corporations`);
    if (!response.ok) {
      throw new Error(`Failed to fetch corporations: ${response.statusText}`);
    }

    const corporations: CardDto[] = await response.json();
    corporationCache = corporations || [];

    return corporationCache;
  } catch (error) {
    console.error("Failed to fetch corporations:", error);
    return [];
  }
}
