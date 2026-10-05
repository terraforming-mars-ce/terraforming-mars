import type { PaymentDto } from "./generated/api-types.ts";

// Common interface for WebSocket connections used throughout the app
export interface WebSocketConnection {
  connected: boolean;
  playerId: string | null;
  gameId: string | null;

  // Connection
  playerConnect(playerName: string, gameId: string, playerId?: string): Promise<any>;

  // Standard project actions
  standardProject(projectId: string): Promise<string>;

  // Resource conversion actions
  convertPlantsToGreenery(payment?: PaymentDto): Promise<string>;
  convertHeatToTemperature(payment?: PaymentDto): Promise<string>;

  // Game management actions
  startGame(): Promise<string>;
  skipAction(): Promise<string>;

  // Card actions
  playCard(
    cardId: string,
    payment: PaymentDto | undefined,
    choiceIndex?: number,
    cardStorageTargets?: string[],
  ): Promise<string>;
  selectStartingChoices(
    corporationId: string,
    preludeIds: string[],
    cardIds: string[],
  ): Promise<string>;
  selectCards(cardIds: string[]): Promise<string>;
  confirmProductionCards(cardIds: string[]): Promise<string>;

  // Tile selection actions
  selectTile(coordinate: { q: number; r: number; s: number }): Promise<string>;
}
