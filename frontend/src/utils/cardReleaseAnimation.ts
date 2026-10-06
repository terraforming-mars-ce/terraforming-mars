/** The "card played" release: the card shrinks slightly and fades out. */
export const CARD_RELEASE_KEYFRAMES: Keyframe[] = [
  { transform: "scale(1)", opacity: 1 },
  { transform: "scale(0.85)", opacity: 0 },
];

export const CARD_RELEASE_DURATION_MS = 350;

export const CARD_RELEASE_TIMING: KeyframeAnimationOptions = {
  duration: CARD_RELEASE_DURATION_MS,
  easing: "ease-out",
  fill: "forwards",
};

/** Runs the release animation on an element, the same one used when a card is played. */
export function playCardRelease(element: HTMLElement): Animation {
  return element.animate(CARD_RELEASE_KEYFRAMES, CARD_RELEASE_TIMING);
}
