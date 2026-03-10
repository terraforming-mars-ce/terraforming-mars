import { v4 as uuidv4 } from "uuid";
import { getWebSocketUrl } from "../config";
import {
  CardPaymentDto,
  ConfirmDemoSetupRequest,
  ErrorPayload,
  FullStatePayload,
  GameUpdatedPayload,
  LogUpdatePayload,
  MessageType,
  MessageTypeError,
  MessageTypeFullState,
  MessageTypeGameUpdated,
  MessageTypeLogUpdate,
  MessageTypePlayerConnect,
  MessageTypePlayerConnected,
  MessageTypePlayerDisconnected,
  MessageTypePlayerKicked,
  // New message types
  MessageTypeActionSellPatents,
  MessageTypeActionLaunchAsteroid,
  MessageTypeActionBuildPowerPlant,
  MessageTypeActionBuildAquifer,
  MessageTypeActionPlantGreenery,
  MessageTypeActionBuildCity,
  MessageTypeActionStartGame,
  MessageTypeActionSkipAction,
  MessageTypeActionPlayCard,
  MessageTypeActionCardAction,
  MessageTypeActionSelectStartingChoices,
  MessageTypeActionConfirmSellPatents,
  MessageTypeActionConfirmProductionCards,
  MessageTypeActionCardDrawConfirmed,
  MessageTypeActionTileSelected,
  MessageTypeActionConvertPlantsToGreenery,
  MessageTypeActionConvertHeatToTemperature,
  MessageTypeActionConfirmDemoSetup,
  MessageTypeActionClaimMilestone,
  MessageTypeActionFundAward,
  MessageTypeAddBot,
  MessageTypeKickPlayer,
  MessageTypeEndGame,
  MessageTypeGameEnded,
  MessageTypeConvertToBot,
  MessageTypeActionBehaviorChoiceConfirmed,
  MessageTypeActionConfirmStealTarget,
  MessageTypeActionCardDiscardConfirmed,
  MessageTypeActionConfirmInitAdvance,
  MessageTypeRequestLogs,
  MessageTypeSetPlayerColor,
  MessageTypeSpectatorConnect,
  MessageTypeSpectatorConnected,
  MessageTypeChatMessage,
  MessageTypeChatUpdate,
  MessageTypeKickSpectator,
  MessageTypeSpectatorKicked,
  // Payload types
  ChatUpdatePayload,
  PlayerConnectedPayload,
  PlayerDisconnectedPayload,
  WebSocketMessage,
} from "../types/generated/api-types.ts";

type EventCallback = (data: any) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private readonly url: string;
  private listeners: { [event: string]: EventCallback[] } = {};
  private isConnected = false;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private currentGameId: string | null = null;
  private currentPlayerId: string | null = null;
  private pendingConnection: Promise<void> | null = null;
  private shouldReconnect = true;

  constructor(url?: string) {
    this.url = url || getWebSocketUrl();
  }

  connect(): Promise<void> {
    // If already connected, resolve immediately
    if (this.isConnected && this.ws && this.ws.readyState === WebSocket.OPEN) {
      return Promise.resolve();
    }

    // If already connecting, return the existing pending promise
    if (this.pendingConnection) {
      return this.pendingConnection;
    }

    // Create new connection promise
    this.pendingConnection = new Promise((resolve, reject) => {
      try {
        // Close existing connection if it exists
        if (this.ws) {
          this.ws.close();
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.isConnected = true;
          this.pendingConnection = null;
          this.reconnectAttempts = 0;
          this.emit("connect");
          resolve();
        };

        this.ws.onmessage = (event) => {
          let message: any;
          try {
            message = JSON.parse(event.data);
          } catch (error) {
            console.error("Failed to parse WebSocket message:", error);
            return;
          }

          try {
            this.handleMessage(message);
          } catch (error) {
            console.error("Error handling WebSocket message:", error);
          }
        };

        this.ws.onclose = (event) => {
          this.isConnected = false;
          this.emit("disconnect");

          if (this.shouldReconnect && event.code !== 1000) {
            this.attemptReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
          this.pendingConnection = null;
          this.emit("error", error);
          if (!this.isConnected) {
            reject(error);
          }
        };
      } catch (error) {
        this.pendingConnection = null;
        reject(error);
      }
    });

    return this.pendingConnection;
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case MessageTypeGameUpdated: {
        const gamePayload = message.payload as GameUpdatedPayload;
        // Handle both direct game data and nested structure
        const gameData = gamePayload.game || gamePayload;
        this.emit("game-updated", gameData);
        break;
      }
      case MessageTypePlayerConnected: {
        const connectedPayload = message.payload as PlayerConnectedPayload;
        // This is a confirmation that player joined successfully
        // The full game state will arrive via game-updated from broadcaster
        this.emit("player-connected", connectedPayload);
        break;
      }
      case MessageTypePlayerDisconnected: {
        const disconnectedPayload = message.payload as PlayerDisconnectedPayload;
        this.emit("player-disconnected", disconnectedPayload);
        break;
      }
      case MessageTypeError: {
        const errorPayload = message.payload as ErrorPayload;
        this.emit("error", errorPayload);
        break;
      }
      case MessageTypeFullState: {
        const statePayload = message.payload as FullStatePayload;
        this.currentPlayerId = statePayload.playerId;
        this.emit("full-state", statePayload);
        break;
      }
      case MessageTypeLogUpdate: {
        const logPayload = message.payload as LogUpdatePayload;
        this.emit("log-update", logPayload.logs);
        break;
      }
      case MessageTypePlayerKicked: {
        this.emit("player-kicked", message.payload);
        break;
      }
      case MessageTypeGameEnded: {
        this.emit("game-ended", message.payload);
        break;
      }
      case MessageTypeSpectatorConnected: {
        this.emit("spectator-connected", message.payload);
        break;
      }
      case MessageTypeChatUpdate: {
        const chatPayload = message.payload as ChatUpdatePayload;
        this.emit("chat-update", chatPayload.chatMessage);
        break;
      }
      case MessageTypeSpectatorKicked: {
        this.emit("spectator-kicked", message.payload);
        break;
      }
      default:
        console.warn("Unknown message type:", message.type);
    }
  }

  send(type: MessageType, payload: unknown, gameId?: string): string {
    const reqId = uuidv4();

    if (!this.isConnected || !this.ws) {
      throw new Error("WebSocket is not connected");
    }

    const message: WebSocketMessage = {
      type,
      payload,
      gameId: gameId || this.currentGameId || undefined,
    };

    this.ws.send(JSON.stringify(message));

    return reqId;
  }

  playerConnect(playerName: string, gameId: string, playerId?: string): void {
    const payload: any = { playerName, gameId };
    if (playerId) {
      payload.playerId = playerId;
    }

    this.send(MessageTypePlayerConnect, payload, gameId);
    this.currentGameId = gameId;
  }

  sellPatents(): string {
    return this.send(MessageTypeActionSellPatents, {});
  }

  launchAsteroid(): string {
    return this.send(MessageTypeActionLaunchAsteroid, {});
  }

  buildPowerPlant(): string {
    return this.send(MessageTypeActionBuildPowerPlant, {});
  }

  buildAquifer(): string {
    return this.send(MessageTypeActionBuildAquifer, {});
  }

  plantGreenery(): string {
    return this.send(MessageTypeActionPlantGreenery, {});
  }

  buildCity(): string {
    return this.send(MessageTypeActionBuildCity, {});
  }

  convertPlantsToGreenery(): string {
    return this.send(MessageTypeActionConvertPlantsToGreenery, {
      type: "convert-plants-to-greenery",
    });
  }

  convertHeatToTemperature(): string {
    return this.send(MessageTypeActionConvertHeatToTemperature, {
      type: "convert-heat-to-temperature",
    });
  }

  startGame(): string {
    return this.send(MessageTypeActionStartGame, {});
  }

  skipAction(): string {
    return this.send(MessageTypeActionSkipAction, {});
  }

  playCard(
    cardId: string,
    payment: CardPaymentDto,
    choiceIndex?: number,
    cardStorageTargets?: string[],
    targetPlayerId?: string,
    selectedAmount?: number,
  ): string {
    return this.send(MessageTypeActionPlayCard, {
      type: "play-card",
      cardId,
      payment,
      ...(choiceIndex !== undefined && { choiceIndex }),
      ...(cardStorageTargets !== undefined && { cardStorageTargets }),
      ...(targetPlayerId !== undefined && { targetPlayerId }),
      ...(selectedAmount !== undefined && { selectedAmount }),
    });
  }

  playCardAction(
    cardId: string,
    behaviorIndex: number,
    choiceIndex?: number,
    cardStorageTargets?: string[],
    targetPlayerId?: string,
    sourceCardForInput?: string,
    selectedAmount?: number,
    payment?: CardPaymentDto,
    reuseSourceCardId?: string,
  ): string {
    return this.send(MessageTypeActionCardAction, {
      type: "card-action",
      cardId,
      behaviorIndex,
      ...(choiceIndex !== undefined && { choiceIndex }),
      ...(cardStorageTargets !== undefined && { cardStorageTargets }),
      ...(targetPlayerId !== undefined && { targetPlayerId }),
      ...(sourceCardForInput !== undefined && { sourceCardForInput }),
      ...(selectedAmount !== undefined && { selectedAmount }),
      ...(payment !== undefined && { payment }),
      ...(reuseSourceCardId !== undefined && { reuseSourceCardId }),
    });
  }

  selectStartingChoices(corporationId: string, preludeIds: string[], cardIds: string[]): string {
    return this.send(MessageTypeActionSelectStartingChoices, {
      corporationId,
      preludeIds,
      cardIds,
    });
  }

  confirmInitAdvance(): string {
    return this.send(MessageTypeActionConfirmInitAdvance, {});
  }

  selectCards(cardIds: string[]): string {
    return this.send(MessageTypeActionConfirmSellPatents, {
      selectedCardIds: cardIds,
    });
  }

  confirmProductionCards(cardIds: string[]): string {
    return this.send(MessageTypeActionConfirmProductionCards, { cardIds });
  }

  confirmCardDraw(cardsToTake: string[], cardsToBuy: string[]): string {
    return this.send(MessageTypeActionCardDrawConfirmed, {
      cardsToTake,
      cardsToBuy,
    });
  }

  selectTile(coordinate: { q: number; r: number; s: number }): string {
    const hex = `${coordinate.q},${coordinate.r},${coordinate.s}`;
    return this.send(MessageTypeActionTileSelected, { hex });
  }

  confirmDemoSetup(request: ConfirmDemoSetupRequest): string {
    return this.send(MessageTypeActionConfirmDemoSetup, request);
  }

  claimMilestone(milestoneType: string): string {
    return this.send(MessageTypeActionClaimMilestone, { milestoneType });
  }

  fundAward(awardType: string): string {
    return this.send(MessageTypeActionFundAward, { awardType });
  }

  playerTakeover(targetPlayerId: string, gameId: string): void {
    this.send("player-takeover" as MessageType, { targetPlayerId, gameId }, gameId);
    this.currentGameId = gameId;
  }

  confirmCardDiscard(cardsToDiscard: string[]): string {
    return this.send(MessageTypeActionCardDiscardConfirmed, { cardsToDiscard });
  }

  confirmBehaviorChoice(choiceIndex: number, cardStorageTargets?: string[]): string {
    return this.send(MessageTypeActionBehaviorChoiceConfirmed, {
      choiceIndex,
      ...(cardStorageTargets !== undefined && { cardStorageTargets }),
    });
  }

  confirmStealTarget(targetPlayerId: string): string {
    return this.send(MessageTypeActionConfirmStealTarget, { targetPlayerId });
  }

  addBot(botName?: string, difficulty?: string, speed?: string): string {
    return this.send(MessageTypeAddBot, {
      botName: botName || "",
      difficulty: difficulty || "normal",
      speed: speed || "normal",
    });
  }

  kickPlayer(targetPlayerId: string): string {
    return this.send(MessageTypeKickPlayer, { targetPlayerId });
  }

  endGame(): string {
    return this.send(MessageTypeEndGame, {});
  }

  convertToBot(targetPlayerId: string): string {
    return this.send(MessageTypeConvertToBot, { targetPlayerId });
  }

  requestLogs(): void {
    this.send(MessageTypeRequestLogs, {});
  }

  setPlayerColor(color: string, targetPlayerId?: string): void {
    this.send(MessageTypeSetPlayerColor, { color, targetPlayerId });
  }

  spectatorConnect(spectatorName: string, gameId: string): void {
    this.send(MessageTypeSpectatorConnect, { spectatorName, gameId }, gameId);
    this.currentGameId = gameId;
  }

  sendChatMessage(message: string): string {
    return this.send(MessageTypeChatMessage, { message });
  }

  kickSpectator(targetSpectatorId: string): string {
    return this.send(MessageTypeKickSpectator, { targetSpectatorId });
  }

  on(event: string, callback: EventCallback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  off(event: string, callback: EventCallback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter((cb) => cb !== callback);
    }
  }

  private emit(event: string, data?: unknown) {
    if (this.listeners[event]) {
      this.listeners[event].forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event listener for ${event}:`, error);
        }
      });
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;

      setTimeout(() => {
        this.connect().catch((error) => {
          console.error("Reconnection failed:", error);
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error("Max reconnection attempts reached");
      this.emit("max-reconnects-reached");
    }
  }

  disconnect() {
    this.shouldReconnect = false;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    this.currentGameId = null;
    this.currentPlayerId = null;
  }

  get connected() {
    return this.isConnected;
  }

  get playerId() {
    return this.currentPlayerId;
  }

  get gameId() {
    return this.currentGameId;
  }
}

// Singleton instance for application-wide use
export const webSocketService = new WebSocketService();
