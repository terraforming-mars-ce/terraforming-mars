import { GamePhaseAction, GameStatusActive, type GameDto } from "@/types/generated/api-types.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import type { PlayerInfo } from "@/utils/colonyUtils.ts";

export function listGamePlayers(gameState: GameDto | undefined): PlayerInfo[] {
  if (!gameState) {
    return [];
  }
  const players: PlayerInfo[] = [];
  if (gameState.currentPlayer?.id) {
    players.push({
      id: gameState.currentPlayer.id,
      name: gameState.currentPlayer.name,
      color: gameState.currentPlayer.color,
    });
  }
  for (const p of gameState.otherPlayers) {
    players.push({ id: p.id, name: p.name, color: p.color });
  }
  return players;
}

export function canActOnProjects(gameState: GameDto | undefined): boolean {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  return isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);
}
