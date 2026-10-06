import {
  CardTypePrelude,
  GamePhaseInitApplyPrelude,
  type CardDto,
  type GameDto,
  type OtherPlayerDto,
  type PlayerDto,
} from "@/types/generated/api-types.ts";

export type ShowcasePlayer = PlayerDto | OtherPlayerDto;

/** Active players in turn order, used by the corporation showcase. */
export function showcasePlayers(game: GameDto): ShowcasePlayer[] {
  const byId = new Map<string, ShowcasePlayer>();
  if (game.currentPlayer?.id) {
    byId.set(game.currentPlayer.id, game.currentPlayer);
  }
  for (const p of game.otherPlayers ?? []) {
    byId.set(p.id, p);
  }
  return (game.turnOrder ?? [])
    .map((id) => byId.get(id))
    .filter((p): p is ShowcasePlayer => p !== undefined && !p.isExited);
}

/**
 * The player who paces the showcase: the host while connected, otherwise the first
 * connected human so setup never stalls. Mirrors the server's confirm permission.
 */
export function showcaseControllerId(game: GameDto): string | null {
  const players = showcasePlayers(game);
  const host = players.find((p) => p.id === game.hostPlayerId);
  if (host && host.isConnected) {
    return host.id;
  }
  const fallback = players.find((p) => p.isConnected && p.playerType !== "bot");
  return fallback?.id ?? null;
}

/**
 * True when the showcase waits for the controller's Next: a player's corp or preludes are
 * revealed but not yet played, or the final roster is up. After Next the cards play and
 * the showcase moves on to the next reveal by itself.
 */
export function isShowcaseStepEnd(game: GameDto): boolean {
  const stage = game.initPhase?.stage;
  return stage === "reveal" || stage === "roster";
}

/** Preludes the player has played, in play order. */
export function shownPreludes(player: ShowcasePlayer): CardDto[] {
  return (player.playedCards ?? []).filter((card) => card.type === CardTypePrelude);
}

/** Whether the player's corporation has been presented in the showcase yet. */
export function corpShown(game: GameDto, playerIndex: number): boolean {
  const initPhase = game.initPhase;
  if (!initPhase) {
    return true;
  }
  if (initPhase.stage === "roster" || game.currentPhase === GamePhaseInitApplyPrelude) {
    return true;
  }
  const currentIndex = showcasePlayers(game).findIndex((p) => p.id === initPhase.currentPlayerId);
  return playerIndex <= currentIndex;
}
