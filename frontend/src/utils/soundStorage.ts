/**
 * Sound settings persistence using localStorage
 */

const STORAGE_KEY = "openmars.sound";

export interface SoundSettings {
  enabled: boolean;
  musicEnabled: boolean;
  volume: number;
  musicVolume: number;
}

const DEFAULT_SETTINGS: SoundSettings = {
  enabled: true,
  musicEnabled: true,
  volume: 0.5,
  musicVolume: 0.5,
};

/**
 * Get sound settings from localStorage, falling back to defaults
 */
export function getSoundSettings(): SoundSettings {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      return DEFAULT_SETTINGS;
    }

    const parsed = JSON.parse(stored) as Partial<SoundSettings>;

    return {
      enabled: typeof parsed.enabled === "boolean" ? parsed.enabled : DEFAULT_SETTINGS.enabled,
      musicEnabled:
        typeof parsed.musicEnabled === "boolean"
          ? parsed.musicEnabled
          : DEFAULT_SETTINGS.musicEnabled,
      volume: typeof parsed.volume === "number" ? parsed.volume : DEFAULT_SETTINGS.volume,
      musicVolume:
        typeof parsed.musicVolume === "number" ? parsed.musicVolume : DEFAULT_SETTINGS.musicVolume,
    };
  } catch {
    return DEFAULT_SETTINGS;
  }
}

/**
 * Save sound settings to localStorage
 */
export function saveSoundSettings(settings: SoundSettings): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch {
    console.warn("Failed to save sound settings to localStorage");
  }
}
