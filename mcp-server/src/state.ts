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
    if (player.pendingCardDrawSelection) return true;
    if (player.pendingCardDiscardSelection) return true;
    if (player.pendingBehaviorChoiceSelection) return true;

    if (player.forcedFirstAction && !player.forcedFirstAction.completed) {
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
    if (p.pendingCardDrawSelection) return "card-draw-selection";
    if (p.pendingCardDiscardSelection) return "card-discard-selection";
    if (p.pendingBehaviorChoiceSelection) return "behavior-choice-selection";
    if (p.forcedFirstAction && !p.forcedFirstAction.completed)
      return "forced-first-action";
    return null;
  }
}
