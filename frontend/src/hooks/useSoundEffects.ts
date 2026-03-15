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

  const playVenusSound = useCallback(() => {
    return audioService.playVenusSound();
  }, []);

  const playButtonHoverSound = useCallback(() => {
    return audioService.playButtonHoverSound();
  }, []);

  const playButtonClickSound = useCallback(() => {
    return audioService.playButtonClickSound();
  }, []);

  const playCardHoverSound = useCallback(() => {
    return audioService.playCardHoverSound();
  }, []);

  const playConstructionSound = useCallback(() => {
    return audioService.playConstructionSound();
  }, []);

  const playAsteroidImpactSound = useCallback(() => {
    return audioService.playAsteroidImpactSound();
  }, []);

  const playYourTurnSound = useCallback(() => {
    return audioService.playYourTurnSound();
  }, []);

  const playAwardFundedSound = useCallback(() => {
    return audioService.playAwardFundedSound();
  }, []);

  const playGameStartSound = useCallback(() => {
    return audioService.playGameStartSound();
  }, []);

  return {
    playSound,
    playProductionSound,
    playTemperatureSound,
    playWaterPlacementSound,
    playOxygenSound,
    playVenusSound,
    playButtonHoverSound,
    playButtonClickSound,
    playCardHoverSound,
    playConstructionSound,
    playAsteroidImpactSound,
    playYourTurnSound,
    playAwardFundedSound,
    playGameStartSound,
  };
}
