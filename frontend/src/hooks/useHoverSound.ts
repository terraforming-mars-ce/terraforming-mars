import { useSoundEffects } from "./useSoundEffects.ts";

export function useHoverSound(disabled?: boolean) {
  const { playButtonHoverSound, playButtonClickSound } = useSoundEffects();

  return {
    onMouseEnter: disabled ? undefined : () => void playButtonHoverSound(),
    onClick: disabled ? undefined : () => void playButtonClickSound(),
  };
}
