import {
  GameDto,
  GamePhase,
  GameStatusActive,
  GamePhaseAction,
  GamePhaseFinalPhase,
} from "@/types/generated/api-types.ts";

export const isPlayerActionPhase = (phase?: GamePhase): boolean => {
  return phase === GamePhaseAction || phase === GamePhaseFinalPhase;
};

/**
 * Checks if a player has actions available
 * Handles both normal actions (positive numbers) and unlimited actions (-1)
 */
export const hasActionsAvailable = (availableActions?: number): boolean => {
  if (availableActions === undefined || availableActions === null) {
    return false;
  }
  // Unlimited actions (-1) or any positive number means actions are available
  return availableActions === -1 || availableActions > 0;
};

/**
 * Checks if a player can perform actions based on game state and conditions.
 * Returns false if the player has a pending tile selection (must complete tile placement first).
 */
export const canPerformActions = (gameState?: GameDto): boolean => {
  if (!gameState?.currentPlayer) {
    return false;
  }

  const isGameActive = gameState.status === GameStatusActive;
  const isActionPhase = isPlayerActionPhase(gameState.currentPhase);
  const isCurrentPlayerTurn = gameState.currentTurn === gameState.viewingPlayerId;
  const hasActions = hasActionsAvailable(gameState.currentPlayer.availableActions);
  const hasPendingTilePlacement = !!gameState.currentPlayer.pendingTileSelection;

  return (
    isGameActive && isActionPhase && isCurrentPlayerTurn && hasActions && !hasPendingTilePlacement
  );
};

/**
 * Checks if a player has unlimited actions (-1)
 */
export const hasUnlimitedActions = (availableActions?: number): boolean => {
  return availableActions === -1;
};

/**
 * Gets the display text for available actions
 * Returns "∞" for -1, number for positive values, "0" for zero or undefined
 */
export const getActionsDisplayText = (availableActions?: number): string => {
  if (availableActions === -1) {
    return "∞";
  }
  return (availableActions || 0).toString();
};

/**
 * Checks if the action button should be enabled for play actions
 */
export const canPlayAction = (gameState?: GameDto): boolean => {
  return canPerformActions(gameState);
};
