import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { WsConnection } from "./connection.js";
import { GameState } from "./state.js";
import { summarizeGameState } from "./summarizer.js";
import type { GameDto, ErrorPayload, FullStatePayload, PlayerConnectedPayload } from "./types.js";
import {
  MessageTypePlayerConnect,
  MessageTypeActionStartGame,
  MessageTypeActionSkipAction,
  MessageTypeActionPlayCard,
  MessageTypeActionCardAction,
  MessageTypeActionSellPatents,
  MessageTypeActionBuildPowerPlant,
  MessageTypeActionLaunchAsteroid,
  MessageTypeActionBuildAquifer,
  MessageTypeActionPlantGreenery,
  MessageTypeActionBuildCity,
  MessageTypeActionConvertPlantsToGreenery,
  MessageTypeActionConvertHeatToTemperature,
  MessageTypeActionSelectStartingChoices,
  MessageTypeActionConfirmSellPatents,
  MessageTypeActionConfirmProductionCards,
  MessageTypeActionCardDrawConfirmed,
  MessageTypeActionCardDiscardConfirmed,
  MessageTypeActionBehaviorChoiceConfirmed,
  MessageTypeActionTileSelected,
  MessageTypeActionClaimMilestone,
  MessageTypeActionFundAward,
} from "./message-types.js";

const DEFAULT_SERVER_URL = "ws://localhost:3001/ws";

export function registerTools(
  server: McpServer,
  conn: WsConnection,
  state: GameState,
) {
  // --- connect_to_game ---
  server.tool(
    "connect_to_game",
    "Connect to a Terraforming Mars game. Joins the game as a player via WebSocket.",
    {
      gameId: z.string().describe("Game ID to join"),
      playerName: z.string().describe("Display name for this player"),
      playerId: z
        .string()
        .optional()
        .describe("Existing player ID for reconnection"),
      serverUrl: z
        .string()
        .optional()
        .describe(
          `WebSocket server URL (default: ${DEFAULT_SERVER_URL})`,
        ),
    },
    async ({ gameId, playerName, playerId, serverUrl }) => {
      try {
        const url = serverUrl || DEFAULT_SERVER_URL;

        if (!conn.connected) {
          await conn.connect(url);
        }

        const gamePromise = new Promise<GameDto>((resolve, reject) => {
          const timeout = setTimeout(() => {
            reject(new Error("Timeout waiting for game state after connect"));
          }, 10000);

          let resolved = false;
          let receivedGame: GameDto | null = null;
          let receivedPlayerId: string | null = null;

          const origOnGameUpdated = conn.onGameUpdated;
          const origOnPlayerConnected = conn.onPlayerConnected;
          const origOnFullState = conn.onFullState;
          const origOnError = conn.onError;

          const cleanup = () => {
            conn.onGameUpdated = origOnGameUpdated;
            conn.onPlayerConnected = origOnPlayerConnected;
            conn.onFullState = origOnFullState;
            conn.onError = origOnError;
            clearTimeout(timeout);
          };

          const tryResolve = () => {
            if (resolved || !receivedGame) return;
            state.myGameId = gameId;
            if (receivedPlayerId) {
              state.myPlayerId = receivedPlayerId;
            }
            state.update(receivedGame);
            resolved = true;
            cleanup();
            resolve(receivedGame);
          };

          conn.onGameUpdated = (game: GameDto) => {
            receivedGame = game;
            if (game.currentPlayer?.id) {
              receivedPlayerId = game.currentPlayer.id;
            }
            tryResolve();
            origOnGameUpdated?.(game);
          };

          conn.onPlayerConnected = (payload: PlayerConnectedPayload) => {
            const pid = payload.playerId || (payload as any).playerID;
            if (pid) receivedPlayerId = pid;
            if (payload.game) {
              receivedGame = payload.game;
            }
            tryResolve();
            origOnPlayerConnected?.(payload);
          };

          conn.onFullState = (payload: FullStatePayload) => {
            const pid = payload.playerId || (payload as any).playerID;
            if (pid) receivedPlayerId = pid;
            if (payload.game) {
              receivedGame = payload.game;
            }
            tryResolve();
            origOnFullState?.(payload);
          };

          conn.onError = (payload: ErrorPayload) => {
            cleanup();
            reject(new Error(payload.message || (payload as any).error || "Connection error"));
            origOnError?.(payload);
          };
        });

        conn.gameId = gameId;
        conn.playerConnect(playerName, gameId, playerId);

        await gamePromise;

        return {
          content: [
            {
              type: "text" as const,
              text:
                `Connected to game ${gameId} as "${playerName}" (player ID: ${state.myPlayerId})\n\n` +
                summarizeGameState(state),
            },
          ],
        };
      } catch (err) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Failed to connect: ${err instanceof Error ? err.message : String(err)}`,
            },
          ],
          isError: true,
        };
      }
    },
  );

  // --- get_game_state ---
  server.tool(
    "get_game_state",
    "Get the current game state as a text summary.",
    {
      verbose: z
        .boolean()
        .optional()
        .describe("Include full card descriptions and complete board"),
    },
    async ({ verbose }) => {
      return {
        content: [
          {
            type: "text" as const,
            text: summarizeGameState(state, verbose ?? false),
          },
        ],
      };
    },
  );

  // --- play_card ---
  server.tool(
    "play_card",
    "Play a card from your hand.",
    {
      cardId: z.string().describe("Card ID to play"),
      credits: z.number().describe("Credits to spend"),
      steel: z.number().optional().describe("Steel resources to use"),
      titanium: z.number().optional().describe("Titanium resources to use"),
      heat: z
        .number()
        .optional()
        .describe("Heat to use as M€ (Helion corporation ability)"),
      choiceIndex: z
        .number()
        .optional()
        .describe("Choice index for cards with multiple options"),
      cardStorageTargets: z
        .array(z.string())
        .optional()
        .describe(
          "Target card IDs for resource storage (positional, one per any-card output)",
        ),
      targetPlayerId: z
        .string()
        .optional()
        .describe("Target player ID for cards that target opponents"),
      selectedAmount: z
        .number()
        .optional()
        .describe("Selected amount for variable-amount effects"),
    },
    async ({
      cardId,
      credits,
      steel,
      titanium,
      heat,
      choiceIndex,
      cardStorageTargets,
      targetPlayerId,
      selectedAmount,
    }) => {
      const payment: Record<string, unknown> = {
        credits,
        steel: steel ?? 0,
        titanium: titanium ?? 0,
      };
      if (heat !== undefined && heat > 0) {
        payment.substitutes = { heat };
      }
      return sendAction(conn, state, MessageTypeActionPlayCard, {
        type: "play-card",
        cardId,
        payment,
        ...(choiceIndex !== undefined && { choiceIndex }),
        ...(cardStorageTargets !== undefined && { cardStorageTargets }),
        ...(targetPlayerId !== undefined && { targetPlayerId }),
        ...(selectedAmount !== undefined && { selectedAmount }),
      });
    },
  );

  // --- use_card_action ---
  server.tool(
    "use_card_action",
    "Activate a played card's action.",
    {
      cardId: z.string().describe("Card ID of the played card"),
      behaviorIndex: z.number().describe("Index of the behavior to activate"),
      choiceIndex: z
        .number()
        .optional()
        .describe("Choice index for actions with multiple options"),
      cardStorageTargets: z
        .array(z.string())
        .optional()
        .describe("Target card IDs for resource storage"),
      targetPlayerId: z
        .string()
        .optional()
        .describe("Target player ID"),
      sourceCardForInput: z
        .string()
        .optional()
        .describe("Source card ID for resource input"),
      selectedAmount: z
        .number()
        .optional()
        .describe("Selected amount for variable-amount effects"),
      credits: z
        .number()
        .optional()
        .describe("Credits to pay (for actions with costs)"),
      steel: z.number().optional().describe("Steel to pay"),
      titanium: z.number().optional().describe("Titanium to pay"),
    },
    async ({
      cardId,
      behaviorIndex,
      choiceIndex,
      cardStorageTargets,
      targetPlayerId,
      sourceCardForInput,
      selectedAmount,
      credits,
      steel,
      titanium,
    }) => {
      const hasPayment =
        credits !== undefined ||
        steel !== undefined ||
        titanium !== undefined;

      return sendAction(conn, state, MessageTypeActionCardAction, {
        type: "card-action",
        cardId,
        behaviorIndex,
        ...(choiceIndex !== undefined && { choiceIndex }),
        ...(cardStorageTargets !== undefined && { cardStorageTargets }),
        ...(targetPlayerId !== undefined && { targetPlayerId }),
        ...(sourceCardForInput !== undefined && { sourceCardForInput }),
        ...(selectedAmount !== undefined && { selectedAmount }),
        ...(hasPayment && {
          payment: {
            credits: credits ?? 0,
            steel: steel ?? 0,
            titanium: titanium ?? 0,
          },
        }),
      });
    },
  );

  // --- standard_project ---
  server.tool(
    "standard_project",
    "Execute a standard project (sell-patents, power-plant, asteroid, aquifer, greenery, city).",
    {
      project: z
        .enum([
          "sell-patents",
          "power-plant",
          "asteroid",
          "aquifer",
          "greenery",
          "city",
        ])
        .describe("Standard project type"),
    },
    async ({ project }) => {
      const messageMap: Record<string, string> = {
        "sell-patents": MessageTypeActionSellPatents,
        "power-plant": MessageTypeActionBuildPowerPlant,
        asteroid: MessageTypeActionLaunchAsteroid,
        aquifer: MessageTypeActionBuildAquifer,
        greenery: MessageTypeActionPlantGreenery,
        city: MessageTypeActionBuildCity,
      };

      return sendAction(conn, state, messageMap[project], {});
    },
  );

  // --- convert_resources ---
  server.tool(
    "convert_resources",
    "Convert resources: plants to greenery or heat to temperature.",
    {
      conversion: z
        .enum(["plants-to-greenery", "heat-to-temperature"])
        .describe("Conversion type"),
    },
    async ({ conversion }) => {
      const messageType =
        conversion === "plants-to-greenery"
          ? MessageTypeActionConvertPlantsToGreenery
          : MessageTypeActionConvertHeatToTemperature;

      return sendAction(conn, state, messageType, {
        type:
          conversion === "plants-to-greenery"
            ? "convert-plants-to-greenery"
            : "convert-heat-to-temperature",
      });
    },
  );

  // --- skip_action ---
  server.tool(
    "skip_action",
    "Pass/skip your turn action.",
    {},
    async () => {
      return sendAction(conn, state, MessageTypeActionSkipAction, {});
    },
  );

  // --- select_tile ---
  server.tool(
    "select_tile",
    "Place a tile on a hex coordinate.",
    {
      q: z.number().describe("Cube coordinate q"),
      r: z.number().describe("Cube coordinate r"),
      s: z.number().describe("Cube coordinate s"),
    },
    async ({ q, r, s }) => {
      return sendAction(conn, state, MessageTypeActionTileSelected, {
        hex: `${q},${r},${s}`,
      });
    },
  );

  // --- select_starting_choices ---
  server.tool(
    "select_starting_choices",
    "Select corporation, preludes, and starting cards during game setup.",
    {
      corporationId: z.string().describe("Corporation card ID to select"),
      preludeIds: z
        .array(z.string())
        .optional()
        .describe("Prelude card IDs to select"),
      cardIds: z
        .array(z.string())
        .describe("Starting card IDs to buy (3M€ each)"),
    },
    async ({ corporationId, preludeIds, cardIds }) => {
      return sendAction(
        conn,
        state,
        MessageTypeActionSelectStartingChoices,
        {
          corporationId,
          preludeIds: preludeIds ?? [],
          cardIds,
        },
      );
    },
  );

  // --- confirm_cards ---
  server.tool(
    "confirm_cards",
    "Confirm card draw/discard/production/sell/behavior-choice selections.",
    {
      action: z
        .enum(["select", "production", "draw", "discard", "behavior-choice"])
        .describe("Type of card confirmation"),
      cardIds: z
        .array(z.string())
        .optional()
        .describe("Card IDs for select/production actions"),
      cardsToTake: z
        .array(z.string())
        .optional()
        .describe("Card IDs to take for free (draw action)"),
      cardsToBuy: z
        .array(z.string())
        .optional()
        .describe("Card IDs to buy (draw action)"),
      cardsToDiscard: z
        .array(z.string())
        .optional()
        .describe("Card IDs to discard"),
      choiceIndex: z
        .number()
        .optional()
        .describe("Choice index for behavior-choice action"),
      cardStorageTargets: z
        .array(z.string())
        .optional()
        .describe("Target card IDs for resource storage (behavior-choice)"),
    },
    async ({
      action,
      cardIds,
      cardsToTake,
      cardsToBuy,
      cardsToDiscard,
      choiceIndex,
      cardStorageTargets,
    }) => {
      switch (action) {
        case "select":
          return sendAction(
            conn,
            state,
            MessageTypeActionConfirmSellPatents,
            { selectedCardIds: cardIds ?? [] },
          );
        case "production":
          return sendAction(
            conn,
            state,
            MessageTypeActionConfirmProductionCards,
            { cardIds: cardIds ?? [] },
          );
        case "draw":
          return sendAction(
            conn,
            state,
            MessageTypeActionCardDrawConfirmed,
            {
              cardsToTake: cardsToTake ?? [],
              cardsToBuy: cardsToBuy ?? [],
            },
          );
        case "discard":
          return sendAction(
            conn,
            state,
            MessageTypeActionCardDiscardConfirmed,
            { cardsToDiscard: cardsToDiscard ?? [] },
          );
        case "behavior-choice":
          return sendAction(
            conn,
            state,
            MessageTypeActionBehaviorChoiceConfirmed,
            {
              choiceIndex: choiceIndex ?? 0,
              ...(cardStorageTargets !== undefined && { cardStorageTargets }),
            },
          );
      }
    },
  );

  // --- claim_milestone ---
  server.tool(
    "claim_milestone",
    "Claim a milestone.",
    {
      milestoneType: z.string().describe("Milestone type to claim"),
    },
    async ({ milestoneType }) => {
      return sendAction(conn, state, MessageTypeActionClaimMilestone, {
        milestoneType,
      });
    },
  );

  // --- fund_award ---
  server.tool(
    "fund_award",
    "Fund an award.",
    {
      awardType: z.string().describe("Award type to fund"),
    },
    async ({ awardType }) => {
      return sendAction(conn, state, MessageTypeActionFundAward, {
        awardType,
      });
    },
  );

  // --- start_game ---
  server.tool(
    "start_game",
    "Start the game (host only).",
    {},
    async () => {
      return sendAction(conn, state, MessageTypeActionStartGame, {});
    },
  );

  // --- wait_for_turn ---
  server.tool(
    "wait_for_turn",
    "Block until it's your turn. Returns the game state when your turn begins.",
    {
      timeoutSeconds: z
        .number()
        .optional()
        .describe("Max seconds to wait (default: 120)"),
    },
    async ({ timeoutSeconds }) => {
      try {
        const timeoutMs = (timeoutSeconds ?? 120) * 1000;
        await state.waitForMyTurn(timeoutMs);
        return {
          content: [
            {
              type: "text" as const,
              text: "It's your turn!\n\n" + summarizeGameState(state),
            },
          ],
        };
      } catch (err) {
        return {
          content: [
            {
              type: "text" as const,
              text: `Wait failed: ${err instanceof Error ? err.message : String(err)}`,
            },
          ],
          isError: true,
        };
      }
    },
  );
}

async function sendAction(
  conn: WsConnection,
  state: GameState,
  messageType: string,
  payload: unknown,
): Promise<{ content: Array<{ type: "text"; text: string }>; isError?: boolean }> {
  try {
    if (!conn.connected) {
      return {
        content: [
          {
            type: "text" as const,
            text: "Not connected. Use connect_to_game first.",
          },
        ],
        isError: true,
      };
    }

    const game = await conn.sendAndWaitForUpdate(
      messageType,
      payload,
      state.myGameId ?? undefined,
    );
    state.update(game);

    return {
      content: [
        {
          type: "text" as const,
          text: summarizeGameState(state),
        },
      ],
    };
  } catch (err) {
    return {
      content: [
        {
          type: "text" as const,
          text: `Action failed: ${err instanceof Error ? err.message : String(err)}`,
        },
      ],
      isError: true,
    };
  }
}
