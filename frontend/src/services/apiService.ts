import {
  BugReportDto,
  BugReportResponse,
  CreateGameRequest,
  CreateGameResponse,
  CreateDemoLobbyRequest,
  CreateDemoLobbyResponse,
  GameDto,
  GameSettingsDto,
  GetGameResponse,
  ListGamesResponse,
  ListCardsResponse,
} from "../types/generated/api-types.ts";
import { config } from "../config";

export class ApiService {
  private baseUrl: string;

  constructor(baseUrl: string = config.apiUrl) {
    this.baseUrl = baseUrl;
  }

  async createGame(settings: GameSettingsDto): Promise<GameDto> {
    try {
      const request: CreateGameRequest = {
        maxPlayers: settings.maxPlayers,
        developmentMode: settings.developmentMode,
        cardPacks: settings.cardPacks,
      };

      const response = await fetch(`${this.baseUrl}/games`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const gameResponse: CreateGameResponse = await response.json();
      return gameResponse.game;
    } catch (error) {
      console.error("Failed to create game:", error);
      throw error;
    }
  }

  async createDemoLobby(settings: CreateDemoLobbyRequest): Promise<CreateDemoLobbyResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/games/demo/lobby`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(settings),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error("Failed to create demo lobby:", error);
      throw error;
    }
  }

  async getGame(gameId: string, playerId?: string): Promise<GameDto | null> {
    try {
      const url = new URL(`${this.baseUrl}/games/${gameId}`, window.location.origin);
      if (playerId) {
        url.searchParams.set("playerId", playerId);
      }

      const response = await fetch(url.toString());

      if (response.status === 404) {
        return null;
      }

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const gameResponse: GetGameResponse = await response.json();
      return gameResponse.game;
    } catch (error) {
      console.error("Failed to get game:", error);
      throw error;
    }
  }

  async listGames(status?: string): Promise<GameDto[]> {
    try {
      const url = new URL(`${this.baseUrl}/games`, window.location.origin);
      if (status) {
        url.searchParams.set("status", status);
      }

      const response = await fetch(url.toString());

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const data: ListGamesResponse = await response.json();
      return data.games || [];
    } catch (error) {
      console.error("Failed to list games:", error);
      throw error;
    }
  }

  async listCards(offset: number = 0, limit: number = 50): Promise<ListCardsResponse> {
    try {
      const url = new URL(`${this.baseUrl}/cards`, window.location.origin);
      url.searchParams.set("offset", offset.toString());
      url.searchParams.set("limit", limit.toString());

      const response = await fetch(url.toString());

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const data: ListCardsResponse = await response.json();
      return data;
    } catch (error) {
      console.error("Failed to list cards:", error);
      throw error;
    }
  }

  async getBugReportStatus(): Promise<{ available: boolean; reason?: string }> {
    try {
      const response = await fetch(`${this.baseUrl}/bugs/status`);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error("Failed to get bug report status:", error);
      throw error;
    }
  }

  async submitBugReport(request: {
    description: string;
    author?: string;
    includeScreenshot: boolean;
    screenshot?: string;
    gameState?: GameDto;
  }): Promise<BugReportDto> {
    try {
      const response = await fetch(`${this.baseUrl}/bugs`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.message || errorData.error || `HTTP error! status: ${response.status}`,
        );
      }

      const data: BugReportResponse = await response.json();
      return data.report;
    } catch (error) {
      console.error("Failed to submit bug report:", error);
      throw error;
    }
  }

  async getBugReport(id: string): Promise<BugReportDto> {
    try {
      const response = await fetch(`${this.baseUrl}/bugs/${id}`);

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const data: BugReportResponse = await response.json();
      return data.report;
    } catch (error) {
      console.error("Failed to get bug report:", error);
      throw error;
    }
  }

}

// Singleton instance
export const apiService = new ApiService();
