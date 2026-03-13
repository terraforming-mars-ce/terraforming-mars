/**
 * Game VP values - centralized constants for victory point calculations.
 */
export const VP_VALUES = {
  MILESTONE: 5,
  AWARD_FIRST: 5,
  AWARD_SECOND: 2,
} as const;

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
