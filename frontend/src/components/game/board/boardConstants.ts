export const SPHERE_RADIUS = 2.02;
export const CHROME_Z_BASE = 0.0156;

export const VENUS_RADIUS = 2.02;
export const VENUS_POSITION = [-45, 25, -148] as const;

export function easeOutCubic(t: number): number {
  return 1 - (1 - t) * (1 - t) * (1 - t);
}
