import WebSocket from "ws";
import type {
  WebSocketMessage,
  MessageType,
  GameDto,
  ErrorPayload,
  GameUpdatedPayload,
  PlayerConnectedPayload,
  FullStatePayload,
  LogUpdatePayload,
} from "./types.js";
import {
  MessageTypeGameUpdated,
  MessageTypePlayerConnected,
  MessageTypePlayerDisconnected,
  MessageTypeError,
  MessageTypeFullState,
  MessageTypeLogUpdate,
  MessageTypePlayerKicked,
  MessageTypePlayerConnect,
} from "./message-types.js";

export type GameUpdateCallback = (game: GameDto) => void;
export type PlayerConnectedCallback = (payload: PlayerConnectedPayload) => void;
export type FullStateCallback = (payload: FullStatePayload) => void;
export type ErrorCallback = (payload: ErrorPayload) => void;
export type LogCallback = (logs: any[]) => void;
export type DisconnectCallback = () => void;

export class WsConnection {
  private ws: WebSocket | null = null;
  private url: string = "";
  private isConnected = false;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private currentGameId: string | null = null;
  private shouldReconnect = true;

  private pendingUpdate: {
    resolve: (game: GameDto) => void;
    reject: (err: Error) => void;
    timer: ReturnType<typeof setTimeout>;
  } | null = null;

  onGameUpdated: GameUpdateCallback | null = null;
  onPlayerConnected: PlayerConnectedCallback | null = null;
  onFullState: FullStateCallback | null = null;
  onError: ErrorCallback | null = null;
  onLog: LogCallback | null = null;
  onDisconnect: DisconnectCallback | null = null;

  connect(url: string): Promise<void> {
    this.url = url;

    if (this.isConnected && this.ws?.readyState === WebSocket.OPEN) {
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      if (this.ws) {
        this.ws.close();
      }

      this.ws = new WebSocket(url);

      this.ws.on("open", () => {
        this.isConnected = true;
        this.reconnectAttempts = 0;
        resolve();
      });

      this.ws.on("message", (data) => {
        let message: WebSocketMessage;
        try {
          message = JSON.parse(data.toString());
        } catch {
          return;
        }
        this.handleMessage(message);
      });

      this.ws.on("close", (code) => {
        this.isConnected = false;
        this.onDisconnect?.();
        if (this.shouldReconnect && code !== 1000) {
          this.attemptReconnect();
        }
      });

      this.ws.on("error", (err) => {
        if (!this.isConnected) {
          reject(err);
        }
      });
    });
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case MessageTypeGameUpdated: {
        const payload = message.payload as GameUpdatedPayload;
        const game = payload.game || (payload as unknown as GameDto);
        this.onGameUpdated?.(game);
        if (this.pendingUpdate) {
          clearTimeout(this.pendingUpdate.timer);
          this.pendingUpdate.resolve(game);
          this.pendingUpdate = null;
        }
        break;
      }
      case MessageTypePlayerConnected: {
        const payload = message.payload as PlayerConnectedPayload;
        this.onPlayerConnected?.(payload);
        if (this.pendingUpdate && payload.game) {
          clearTimeout(this.pendingUpdate.timer);
          this.pendingUpdate.resolve(payload.game);
          this.pendingUpdate = null;
        }
        break;
      }
      case MessageTypeFullState: {
        const payload = message.payload as FullStatePayload;
        this.onFullState?.(payload);
        if (this.pendingUpdate && payload.game) {
          clearTimeout(this.pendingUpdate.timer);
          this.pendingUpdate.resolve(payload.game);
          this.pendingUpdate = null;
        }
        break;
      }
      case MessageTypeError: {
        const payload = message.payload as ErrorPayload;
        this.onError?.(payload);
        if (this.pendingUpdate) {
          clearTimeout(this.pendingUpdate.timer);
          this.pendingUpdate.reject(
            new Error(payload.message || (payload as any).error || "Server error"),
          );
          this.pendingUpdate = null;
        }
        break;
      }
      case MessageTypeLogUpdate: {
        const payload = message.payload as LogUpdatePayload;
        this.onLog?.(payload.logs);
        break;
      }
      case MessageTypePlayerDisconnected:
      case MessageTypePlayerKicked:
        break;
    }
  }

  send(type: MessageType, payload: unknown, gameId?: string): void {
    if (!this.isConnected || !this.ws) {
      throw new Error("WebSocket is not connected");
    }

    const message: WebSocketMessage = {
      type,
      payload,
      gameId: gameId || this.currentGameId || undefined,
    };

    this.ws.send(JSON.stringify(message));
  }

  sendAndWaitForUpdate(
    type: MessageType,
    payload: unknown,
    gameId?: string,
    timeoutMs = 10000,
  ): Promise<GameDto> {
    return new Promise((resolve, reject) => {
      if (this.pendingUpdate) {
        clearTimeout(this.pendingUpdate.timer);
        this.pendingUpdate.reject(new Error("Superseded by new request"));
      }

      const timer = setTimeout(() => {
        this.pendingUpdate = null;
        reject(new Error(`Timeout waiting for game update after ${type}`));
      }, timeoutMs);

      this.pendingUpdate = { resolve, reject, timer };

      try {
        this.send(type, payload, gameId);
      } catch (err) {
        clearTimeout(timer);
        this.pendingUpdate = null;
        reject(err);
      }
    });
  }

  playerConnect(
    playerName: string,
    gameId: string,
    playerId?: string,
  ): void {
    const payload: Record<string, string> = { playerName, gameId };
    if (playerId) {
      payload.playerId = playerId;
    }
    this.send(MessageTypePlayerConnect, payload, gameId);
    this.currentGameId = gameId;
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        this.connect(this.url).catch(() => {});
      }, this.reconnectDelay * this.reconnectAttempts);
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
  }

  get connected() {
    return this.isConnected;
  }

  get gameId() {
    return this.currentGameId;
  }

  set gameId(id: string | null) {
    this.currentGameId = id;
  }
}
