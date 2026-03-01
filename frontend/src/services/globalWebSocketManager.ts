import { webSocketService } from "./webSocketService.ts";
import { WebSocketConnection } from "../types/webSocketTypes.ts";
import type {
  CardPaymentDto,
  ConfirmDemoSetupRequest,
  GameDto,
  PlayerDisconnectedPayload,
  FullStatePayload,
  StateDiffDto,
} from "../types/generated/api-types.ts";

class GlobalWebSocketManager implements WebSocketConnection {
  private isInitialized = false;
  private initializationPromise: Promise<void> | null = null;
  private currentPlayerId: string | null = null;
  private eventCallbacks: { [event: string]: ((data: any) => void)[] } = {};
  private isIntentionalDisconnect = false;
  private handlersSetUp = false;

  async initialize() {
    if (this.isInitialized) {
      return;
    }

    if (this.initializationPromise) {
      return this.initializationPromise;
    }

    this.isIntentionalDisconnect = false;
    this.initializationPromise = this._doInitialize();

    try {
      await this.initializationPromise;
    } finally {
      this.initializationPromise = null;
    }
  }

  private async _doInitialize() {
    try {
      await webSocketService.connect();
      this.setupGlobalEventHandlers();
      this.isInitialized = true;
    } catch (error) {
      console.error("Failed to initialize global WebSocket connection:", error);
      throw error;
    }
  }

  async ensureConnected() {
    if (!this.isInitialized) {
      await this.initialize();
    }

    if (!webSocketService.connected) {
      return new Promise<void>((resolve, reject) => {
        const checkConnection = () => {
          if (webSocketService.connected) {
            resolve();
          } else {
            setTimeout(checkConnection, 100);
          }
        };

        checkConnection();

        setTimeout(() => {
          reject(new Error("WebSocket connection timeout"));
        }, 10000);
      });
    }
  }

  private setupGlobalEventHandlers() {
    if (this.handlersSetUp) {
      return;
    }
    this.handlersSetUp = true;

    webSocketService.on("game-updated", (updatedGame: GameDto) => {
      this.emit("game-updated", updatedGame);
    });

    webSocketService.on("full-state", (statePayload: FullStatePayload) => {
      this.emit("full-state", statePayload);
    });

    webSocketService.on("player-disconnected", (payload: PlayerDisconnectedPayload) => {
      this.emit("player-disconnected", payload);
    });

    webSocketService.on("player-kicked", (payload: any) => {
      this.emit("player-kicked", payload);
    });

    webSocketService.on("log-update", (logs: StateDiffDto[]) => {
      this.emit("log-update", logs);
    });

    webSocketService.on("available-cards", (payload: any) => {
      this.emit("available-cards", payload);
    });

    webSocketService.on("error", (error: any) => {
      console.error("WebSocket error:", error);
      this.emit("error", error);
    });

    webSocketService.on("disconnect", () => {
      this.emit("disconnect");
    });

    webSocketService.on("connect", () => {
      this.emit("connect");
    });

    webSocketService.on("max-reconnects-reached", () => {
      this.emit("max-reconnects-reached");
    });
  }

  setCurrentPlayerId(playerId: string) {
    this.currentPlayerId = playerId;
  }

  getCurrentPlayerId(): string | null {
    return this.currentPlayerId;
  }

  on(event: string, callback: (data: any) => void) {
    if (!this.eventCallbacks[event]) {
      this.eventCallbacks[event] = [];
    }
    this.eventCallbacks[event].push(callback);
  }

  off(event: string, callback: (data: any) => void) {
    if (this.eventCallbacks[event]) {
      this.eventCallbacks[event] = this.eventCallbacks[event].filter((cb) => cb !== callback);
    }
  }

  private emit(event: string, data?: any) {
    if (this.eventCallbacks[event]) {
      this.eventCallbacks[event].forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in WebSocket event callback for ${event}:`, error);
        }
      });
    }
  }

  async playerConnect(playerName: string, gameId: string, playerId?: string) {
    await this.ensureConnected();
    return webSocketService.playerConnect(playerName, gameId, playerId);
  }

  async sellPatents(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.sellPatents();
  }

  async launchAsteroid(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.launchAsteroid();
  }

  async buildPowerPlant(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildPowerPlant();
  }

  async buildAquifer(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildAquifer();
  }

  async plantGreenery(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.plantGreenery();
  }

  async buildCity(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildCity();
  }

  async convertPlantsToGreenery(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.convertPlantsToGreenery();
  }

  async convertHeatToTemperature(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.convertHeatToTemperature();
  }

  async startGame(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.startGame();
  }

  async skipAction(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.skipAction();
  }

  async playCard(
    cardId: string,
    payment: CardPaymentDto,
    choiceIndex?: number,
    cardStorageTargets?: string[],
    targetPlayerId?: string,
    selectedAmount?: number,
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.playCard(
      cardId,
      payment,
      choiceIndex,
      cardStorageTargets,
      targetPlayerId,
      selectedAmount,
    );
  }

  async playCardAction(
    cardId: string,
    behaviorIndex: number,
    choiceIndex?: number,
    cardStorageTargets?: string[],
    targetPlayerId?: string,
    sourceCardForInput?: string,
    selectedAmount?: number,
    payment?: CardPaymentDto,
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.playCardAction(
      cardId,
      behaviorIndex,
      choiceIndex,
      cardStorageTargets,
      targetPlayerId,
      sourceCardForInput,
      selectedAmount,
      payment,
    );
  }

  async selectStartingChoices(
    corporationId: string,
    preludeIds: string[],
    cardIds: string[],
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectStartingChoices(corporationId, preludeIds, cardIds);
  }

  async selectCards(cardIds: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectCards(cardIds);
  }

  async confirmProductionCards(cardIds: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmProductionCards(cardIds);
  }

  async confirmCardDraw(cardsToTake: string[], cardsToBuy: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmCardDraw(cardsToTake, cardsToBuy);
  }

  async selectTile(coordinate: { q: number; r: number; s: number }): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectTile(coordinate);
  }

  async confirmDemoSetup(request: ConfirmDemoSetupRequest): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmDemoSetup(request);
  }

  async sendAdminCommand(adminRequest: any): Promise<string> {
    await this.ensureConnected();
    const { MessageTypeAdminCommand } = await import("../types/generated/api-types.ts");
    return webSocketService.send(MessageTypeAdminCommand, adminRequest);
  }

  async playerTakeover(targetPlayerId: string, gameId: string): Promise<void> {
    await this.ensureConnected();
    return webSocketService.playerTakeover(targetPlayerId, gameId);
  }

  async confirmCardDiscard(cardsToDiscard: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmCardDiscard(cardsToDiscard);
  }

  async confirmBehaviorChoice(choiceIndex: number, cardStorageTargets?: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmBehaviorChoice(choiceIndex, cardStorageTargets);
  }

  async kickPlayer(targetPlayerId: string): Promise<string> {
    await this.ensureConnected();
    return webSocketService.kickPlayer(targetPlayerId);
  }

  async requestLogs(): Promise<void> {
    await this.ensureConnected();
    webSocketService.requestLogs();
  }

  get connected() {
    return webSocketService.connected;
  }

  get playerId() {
    return webSocketService.playerId;
  }

  get gameId() {
    return webSocketService.gameId;
  }

  disconnect() {
    this.isIntentionalDisconnect = true;
    webSocketService.disconnect();
    this.isInitialized = false;
    this.currentPlayerId = null;
  }

  isGracefulDisconnect(): boolean {
    return this.isIntentionalDisconnect;
  }
}

// Singleton instance - initialized once globally
export const globalWebSocketManager = new GlobalWebSocketManager();
