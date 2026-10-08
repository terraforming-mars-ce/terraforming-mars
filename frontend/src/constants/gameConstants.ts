/**
 * Maximum allowed length for a player name. Single source of truth for all
 * player-name inputs.
 */
export const MAX_PLAYER_NAME_LENGTH = 45;

/**
 * Human-readable display names for game phases.
 */
export const PHASE_DISPLAY_NAMES: Record<string, string> = {
  waiting_for_game_start: "Lobby",
  starting_selection: "Selection",
  init_apply_corp: "Pre-game",
  init_apply_prelude: "Pre-game",
  action: "In Game",
  production_and_card_draw: "Production",
  complete: "Complete",
};

export function getPhaseDisplayName(phase: string): string {
  return PHASE_DISPLAY_NAMES[phase] ?? phase;
}

/**
 * Corporation showcase pacing after the controller's Next: the 350ms card-play animation
 * plus a 300ms gap before the next card plays or the next player is revealed.
 */
export const SHOWCASE_PLAY_STEP_MS = 650;

/** How long the board stays in view after a card's last tile placement before the showcase returns. */
export const SHOWCASE_TILE_HOLD_MS = 2000;
export const SHOWCASE_FADE_MS = 400;
