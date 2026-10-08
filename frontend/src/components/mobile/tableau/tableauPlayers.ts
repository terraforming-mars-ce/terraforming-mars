import type { GameDto, OtherPlayerDto, PlayerDto } from "@/types/generated/api-types.ts";

export type TableauPlayer = PlayerDto | OtherPlayerDto;

export function playersInTurnOrder(gameState: GameDto): TableauPlayer[] {
  const byId = new Map<string, TableauPlayer>();
  if (gameState.currentPlayer?.id) {
    byId.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  for (const other of gameState.otherPlayers ?? []) {
    byId.set(other.id, other);
  }
  const order = gameState.turnOrder?.length ? gameState.turnOrder : (gameState.playerOrder ?? []);
  const ordered = order
    .map((id) => byId.get(id))
    .filter((player): player is TableauPlayer => player !== undefined);
  if (ordered.length > 0) {
    return ordered;
  }
  return [...byId.values()];
}

export function playerColor(player: TableauPlayer, colorMap: Map<string, string>): string {
  return colorMap.get(player.id) || player.color || "#6496ff";
}
