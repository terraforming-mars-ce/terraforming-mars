const PLAYER_COLORS = [
  "#b91c2b", // Red
  "#232dc7", // Blue
  "#3abe3a", // Green
  "#ffa502", // Orange
  "#a55eea", // Purple
  "#26d0ce", // Cyan
];

export function getPlayerColor(index: number): string {
  return PLAYER_COLORS[index % PLAYER_COLORS.length];
}
