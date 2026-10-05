import type { GameDto } from "./types.js";
import {
  GamePhaseAction,
  GamePhaseStartingSelection,
  GamePhaseProductionAndCardDraw,
  PlayerStatusActive,
  PlayerStatusSelectingStartingCards,
  PlayerStatusSelectingProductionCards,
} from "./types.js";

export class GameState {
  game: GameDto | null = null;
  myPlayerId: string | null = null;
  myGameId: string | null = null;

  private turnWaiters: Array<{
    resolve: () => void;
    reject: (err: Error) => void;
    timer: ReturnType<typeof setTimeout>;
  }> = [];

  update(game: GameDto) {
    this.game = game;

    if (this.isMyTurn()) {
      for (const waiter of this.turnWaiters) {
        clearTimeout(waiter.timer);
        waiter.resolve();
      }
      this.turnWaiters = [];
    }
  }

  isMyTurn(): boolean {
    if (!this.game || !this.myPlayerId) return false;
    const player = this.game.currentPlayer;
    if (!player) return false;

    if (
      this.game.currentPhase === GamePhaseAction &&
      this.game.currentTurn === this.myPlayerId
    ) {
      return true;
    }

    if (player.status === PlayerStatusActive) return true;
    if (player.status === PlayerStatusSelectingStartingCards) return true;
    if (player.status === PlayerStatusSelectingProductionCards) return true;

    if (player.pendingTileSelection) return true;
    if (player.pendingCardSelection) return true;
    if (player.pendingCardDrawSelection || player.pendingColonySelection || player.pendingColonyResourceSelection || player.pendingAwardFundSelection) return true;
    if (player.pendingBehaviorResolutions?.length) return true;
    if (player.pendingResourceRemovalSelection || player.pendingEffectSelection) return true;
    if (player.pendingCardReveal) return true;

    if (player.forcedFirstAction?.state === "resolving") {
      return true;
    }

    if (this.game.currentPhase === GamePhaseStartingSelection) return true;
    if (
      this.game.currentPhase === GamePhaseProductionAndCardDraw &&
      player.productionPhase &&
      !player.productionPhase.selectionComplete
    ) {
      return true;
    }

    return false;
  }

  waitForMyTurn(timeoutMs = 120000): Promise<void> {
    if (this.isMyTurn()) {
      return Promise.resolve();
    }

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.turnWaiters = this.turnWaiters.filter((w) => w.timer !== timer);
        reject(new Error("Timeout waiting for turn"));
      }, timeoutMs);

      this.turnWaiters.push({ resolve, reject, timer });
    });
  }

  getPendingActionType(): string | null {
    if (!this.game) return null;
    const p = this.game.currentPlayer;
    if (!p) return null;

    if (p.pendingTileSelection) return "tile-selection";
    if (p.pendingCardSelection) return "card-selection";
    if (p.pendingColonySelection) return "colony-selection";
    if (p.pendingColonyResourceSelection) return "colony-resource-selection";
    if (p.pendingAwardFundSelection) return "award-fund-selection";
    if (p.pendingCardDrawSelection) return "card-draw-selection";
    if (p.pendingBehaviorResolutions?.length) return "behavior-resolution";
    if (p.pendingResourceRemovalSelection) return "resource-removal";
    if (p.pendingEffectSelection) return "effect-selection";
    if (p.pendingCardReveal) return "card-reveal";
    if (p.forcedFirstAction?.state === "resolving")
      return "forced-first-action";
    return null;
  }
}
