import { useCallback } from "react";
import { audioService } from "../services/audioService.ts";

/**
 * Hook for playing sound effects in components
 * Thin wrapper around audioService methods
 */
export function useSoundEffects() {
  const playSound = useCallback((soundKey: string) => {
    return audioService.playSound(soundKey);
  }, []);

  const playProductionSound = useCallback(() => {
    return audioService.playProductionSound();
  }, []);

  const playTemperatureSound = useCallback(() => {
    return audioService.playTemperatureSound();
  }, []);

  const playWaterPlacementSound = useCallback(() => {
    return audioService.playWaterPlacementSound();
  }, []);

  const playOxygenSound = useCallback(() => {
    return audioService.playOxygenSound();
  }, []);

  const playCardHoverSound = useCallback(() => {
    return audioService.playCardHoverSound();
  }, []);

  return {
    playSound,
    playProductionSound,
    playTemperatureSound,
    playWaterPlacementSound,
    playOxygenSound,
    playCardHoverSound,
  };
}
