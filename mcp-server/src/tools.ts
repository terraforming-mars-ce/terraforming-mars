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
  MessageTypeActionStandardProject,
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

const paymentSchema = z.object({ allocations: z.array(z.object({
 source: z.object({ target: z.enum(["self-player", "self-card"]), resource: z.string(), cardId: z.string().optional() }),
 targetResource: z.string(), amount: z.number().int().nonnegative(),
})) }).describe("Source allocations from quote_payment; quantities are resources spent");

const DEFAULT_SERVER_URL = "ws://localhost:3001/ws";

export function registerTools(
  server: McpServer,
  conn: WsConnection,
  state: GameState,
) {
  server.tool("quote_payment", "Get current costs and eligible payment sources without spending resources.", {
    action: z.enum(["play-card", "card-action", "standard-project", "convert-heat", "convert-plants", "select-starting-choices", "confirm-production-cards", "confirm-card-draw", "claim-milestone", "fund-award", "build-colony", "colony-trade"]),
    cardId: z.string().optional(), behaviorIndex: z.number().int().optional(), choiceIndex: z.number().int().optional(), selectedAmount: z.number().int().nonnegative().optional(), cardStorageSources: z.array(z.string()).optional(),
    projectId: z.string().optional(), corporationId: z.string().optional(), cardIds: z.array(z.string()).optional(), cardsToBuy: z.array(z.string()).optional(), randomBuy: z.boolean().optional(), paymentType: z.string().optional(), milestoneType: z.string().optional(), awardType: z.string().optional(),
  }, async (intent) => {
    try { return { content: [{ type: "text" as const, text: JSON.stringify(await conn.quotePayment(intent)) }] }; }
    catch (error) { return { content: [{ type: "text" as const, text: String(error) }], isError: true }; }
  });

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
      payment: paymentSchema,
      choiceIndex: z
        .number()
        .optional()
        .describe("Choice index for cards with multiple options"),
      cardStorageSources: z.array(z.string()).optional().describe("Owned card IDs paying any-card storage inputs, in input order"),
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
      payment,
      choiceIndex,
      cardStorageTargets,
      cardStorageSources,
      targetPlayerId,
      selectedAmount,
    }) => {
      return sendAction(conn, state, MessageTypeActionPlayCard, {
        type: "play-card",
        cardId,
        payment,
        ...(choiceIndex !== undefined && { choiceIndex }),
        ...(cardStorageTargets !== undefined && { cardStorageTargets }),
        ...(cardStorageSources !== undefined && { cardStorageSources }),
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
      reuseSourceCardId: z.string().optional().describe("Card providing action reuse, such as Viron"),
      cardId: z.string().describe("Card ID of the played card"),
      behaviorIndex: z.number().describe("Index of the behavior to activate"),
      choiceIndex: z
        .number()
        .optional()
        .describe("Choice index for actions with multiple options"),
      cardStorageSources: z.array(z.string()).optional().describe("Owned card IDs paying any-card storage inputs, in input order"),
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
      payment: paymentSchema,
    },
    async ({
      reuseSourceCardId,
      cardId,
      behaviorIndex,
      choiceIndex,
      cardStorageTargets,
      cardStorageSources,
      targetPlayerId,
      sourceCardForInput,
      selectedAmount,
      payment,
    }) => {
      return sendAction(conn, state, MessageTypeActionCardAction, {
        type: "card-action",
        ...(reuseSourceCardId !== undefined && { reuseSourceCardId }),
        cardId,
        behaviorIndex,
        ...(choiceIndex !== undefined && { choiceIndex }),
        ...(cardStorageTargets !== undefined && { cardStorageTargets }),
        ...(cardStorageSources !== undefined && { cardStorageSources }),
        ...(targetPlayerId !== undefined && { targetPlayerId }),
        ...(sourceCardForInput !== undefined && { sourceCardForInput }),
        ...(selectedAmount !== undefined && { selectedAmount }),
        payment,
      });
    },
  );

  // --- standard_project ---
  server.tool(
    "standard_project",
    "Execute a standard project (sell-patents, power-plant, asteroid, aquifer, greenery, city).",
    {
      payment: paymentSchema,
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
    async ({ payment, project }) => {
      return sendAction(conn, state, MessageTypeActionStandardProject, { projectId: project, payment });
    },
  );

  // --- convert_resources ---
  server.tool(
    "convert_resources",
    "Convert resources: plants to greenery or heat to temperature.",
    {
      payment: paymentSchema,
      conversion: z
        .enum(["plants-to-greenery", "heat-to-temperature"])
        .describe("Conversion type"),
    },
    async ({ payment, conversion }) => {
      const messageType =
        conversion === "plants-to-greenery"
          ? MessageTypeActionConvertPlantsToGreenery
          : MessageTypeActionConvertHeatToTemperature;

      return sendAction(conn, state, messageType, {
        payment,
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
      payment: paymentSchema,
      corporationId: z.string().describe("Corporation card ID to select"),
      preludeIds: z
        .array(z.string())
        .optional()
        .describe("Prelude card IDs to select"),
      cardIds: z
        .array(z.string())
        .describe("Starting card IDs to buy (3M€ each)"),
    },
    async ({ payment, corporationId, preludeIds, cardIds }) => {
      return sendAction(
        conn,
        state,
        MessageTypeActionSelectStartingChoices,
        {
          corporationId,
          payment,
          preludeIds: preludeIds ?? [],
          cardIds,
        },
      );
    },
  );

  server.tool("acknowledge_cards_received", "Close a cards-received receipt. Cards have already been added to your hand.", { receiptId: z.string() }, async ({ receiptId }) => sendAction(conn, state, "action.acknowledge-card-receipt", { receiptId }));
  server.tool("confirm_colony_selection", "Confirm a pending build-colony or add-colony-tile choice using an offered colony ID.", { colonyId: z.string() }, async ({ colonyId }) => sendAction(conn, state, "action.confirm-colony-placement", { colonyId }));
  server.tool("confirm_colony_resource", "Place a pending colony resource on an owned card, or use an empty cardId to skip.", { cardId: z.string() }, async ({ cardId }) => sendAction(conn, state, "action.confirm-colony-resource", { cardId }));
  server.tool("confirm_free_award", "Confirm Vitor's pending free award funding choice.", { awardType: z.string() }, async ({ awardType }) => sendAction(conn, state, "action.confirm-award-fund", { awardType }));
  // --- confirm_cards ---
  server.tool(
    "confirm_resource_removal",
    "Remove a chosen amount from an eligible player, or skip with empty targetPlayerId and amount 0.",
    { selectionId: z.string(), targetPlayerId: z.string(), amount: z.number().int().min(0) },
    async ({ selectionId, targetPlayerId, amount }) => sendAction(conn, state, "action.card.confirm-resource-removal", { selectionId, targetPlayerId, amount }),
  );

  server.tool(
    "confirm_cards",
    "Confirm card draw/discard/production/sell/behavior-choice selections.",
    {
      payment: paymentSchema,
      resolutionId: z.string().optional().describe("Required for discard and behavior-choice; use the pending decision ID"),
      action: z
        .enum(["select", "production", "draw", "discard", "behavior-choice", "effect", "reveal"])
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
    async ({ payment,
      resolutionId,
      action,
      cardIds,
      cardsToTake,
      cardsToBuy,
      cardsToDiscard,
      choiceIndex,
      cardStorageTargets,
    }) => {
      if ((action === "discard" || action === "behavior-choice") && !resolutionId) {
        throw new Error("resolutionId is required");
      }
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
            { cardIds: cardIds ?? [], payment },
          );
        case "draw":
          return sendAction(
            conn,
            state,
            MessageTypeActionCardDrawConfirmed,
            {
              cardsToTake: cardsToTake ?? [],
              cardsToBuy: cardsToBuy ?? [],
              payment,
            },
          );
        case "discard":
          return sendAction(
            conn,
            state,
            MessageTypeActionCardDiscardConfirmed,
            { resolutionId, cardsToDiscard: cardsToDiscard ?? [] },
          );
        case "reveal":
          return sendAction(conn, state, "action.confirm-card-reveal", {});
        case "effect":
          return sendAction(conn, state, "action.confirm-effect-selection", { optionIndex: choiceIndex ?? 0 });
        case "behavior-choice":
          return sendAction(
            conn,
            state,
            MessageTypeActionBehaviorChoiceConfirmed,
            {
              resolutionId,
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
      payment: paymentSchema,
      milestoneType: z.string().describe("Milestone type to claim"),
    },
    async ({ payment, milestoneType }) => {
      return sendAction(conn, state, MessageTypeActionClaimMilestone, {
        milestoneType,
        payment,
      });
    },
  );

  // --- fund_award ---
  server.tool(
    "fund_award",
    "Fund an award.",
    {
      payment: paymentSchema,
      awardType: z.string().describe("Award type to fund"),
    },
    async ({ payment, awardType }) => {
      return sendAction(conn, state, MessageTypeActionFundAward, {
        awardType,
        payment,
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
