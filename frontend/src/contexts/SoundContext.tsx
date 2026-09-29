import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import { audioService } from "../services/audioService.ts";
import { getSoundSettings, saveSoundSettings, SoundSettings } from "../utils/soundStorage.ts";

interface SoundContextType {
  enabled: boolean;
  musicEnabled: boolean;
  volume: number;
  musicVolume: number;
  toggleMute: () => void;
  toggleMusicMute: () => void;
  setVolume: (volume: number) => void;
  setMusicVolume: (volume: number) => void;
}

const SoundContext = createContext<SoundContextType | undefined>(undefined);

export function SoundProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<SoundSettings>(() => getSoundSettings());

  useEffect(() => {
    audioService.setEnabled(settings.enabled);
    audioService.setMusicEnabled(settings.musicEnabled);
    audioService.setVolume(settings.volume);
    audioService.setMusicVolume(settings.musicVolume);
    saveSoundSettings(settings);
  }, [settings]);

  const toggleMute = useCallback(() => {
    setSettings((prev) => ({
      ...prev,
      enabled: !prev.enabled,
    }));
  }, []);

  const toggleMusicMute = useCallback(() => {
    setSettings((prev) => ({
      ...prev,
      musicEnabled: !prev.musicEnabled,
    }));
  }, []);

  const setVolume = useCallback((volume: number) => {
    setSettings((prev) => ({
      ...prev,
      enabled: true,
      volume: Math.max(0, Math.min(1, volume)),
    }));
  }, []);

  const setMusicVolume = useCallback((volume: number) => {
    setSettings((prev) => ({
      ...prev,
      musicEnabled: true,
      musicVolume: Math.max(0, Math.min(1, volume)),
    }));
  }, []);

  const contextValue: SoundContextType = {
    enabled: settings.enabled,
    musicEnabled: settings.musicEnabled,
    volume: settings.volume,
    musicVolume: settings.musicVolume,
    toggleMute,
    toggleMusicMute,
    setVolume,
    setMusicVolume,
  };

  return <SoundContext.Provider value={contextValue}>{children}</SoundContext.Provider>;
}

export function useSound() {
  const context = useContext(SoundContext);
  if (context === undefined) {
    throw new Error("useSound must be used within a SoundProvider");
  }
  return context;
}
