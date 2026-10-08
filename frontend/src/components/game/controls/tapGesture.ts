export const TAP_SLOP_PX = { mouse: 4, pen: 8, touch: 12 } as const;

export function slopFor(pointerType: string): number {
  if (pointerType === "touch") {
    return TAP_SLOP_PX.touch;
  }
  if (pointerType === "pen") {
    return TAP_SLOP_PX.pen;
  }
  return TAP_SLOP_PX.mouse;
}
