import { getSoundSettings } from "../utils/soundStorage.ts";

interface AudioFileEntry {
  key: string;
  path: string;
  volumeMultiplier: number;
}

class AudioService {
  private audioCache: Map<string, HTMLAudioElement> = new Map();
  private ambientAudio: HTMLAudioElement | null = null;
  private isEnabled: boolean = true;
  private isMusicEnabled: boolean = true;
  private volume: number = 0.5;
  private musicVolume: number = 0.5;
  private volumeMultipliers: Map<string, number> = new Map();
  private ambientVolumeMultiplier: number = 0.3;
  private fadeOutInterval: ReturnType<typeof setInterval> | null = null;
  private ambientTracks: string[] = ["/sounds/main-ambient.mp3", "/sounds/choral-chambers.mp3"];
  private currentTrackIndex: number = 0;

  constructor() {
    const settings = getSoundSettings();
    this.isEnabled = settings.enabled;
    this.isMusicEnabled = settings.musicEnabled;
    this.volume = settings.volume;
    this.musicVolume = settings.musicVolume;

    // Empty functions (not null) intercept media keys without triggering playback
    if (navigator.mediaSession) {
      const noop = () => {};
      navigator.mediaSession.setActionHandler("play", noop);
      navigator.mediaSession.setActionHandler("pause", noop);
      navigator.mediaSession.setActionHandler("stop", noop);
      navigator.mediaSession.setActionHandler("seekbackward", noop);
      navigator.mediaSession.setActionHandler("seekforward", noop);
      navigator.mediaSession.setActionHandler("previoustrack", noop);
      navigator.mediaSession.setActionHandler("nexttrack", noop);
    }

    const unlockAudio = () => {
      const ctx = new AudioContext();
      void ctx.resume().then(() => ctx.close());
      document.removeEventListener("click", unlockAudio);
      document.removeEventListener("touchstart", unlockAudio);
      document.removeEventListener("keydown", unlockAudio);
    };
    document.addEventListener("click", unlockAudio);
    document.addEventListener("touchstart", unlockAudio);
    document.addEventListener("keydown", unlockAudio);

    this.preloadAudioFiles();
  }

  private preloadAudioFiles() {
    const audioFiles: AudioFileEntry[] = [
      { key: "production", path: "/sounds/production.mp3", volumeMultiplier: 1.0 },
      {
        key: "temperature-increase",
        path: "/sounds/temperature-increase.mp3",
        volumeMultiplier: 1.0,
      },
      { key: "water-placement", path: "/sounds/water-placement.mp3", volumeMultiplier: 1.0 },
      { key: "oxygen-increase", path: "/sounds/oxygen-increase.mp3", volumeMultiplier: 1.0 },
      { key: "button-hover", path: "/sounds/button-hover.mp3", volumeMultiplier: 0.4 },
      { key: "button-click", path: "/sounds/button-click.mp3", volumeMultiplier: 0.4 },
      { key: "card-hover", path: "/sounds/card-hover.mp3", volumeMultiplier: 0.2 },
      { key: "construction", path: "/sounds/construction.mp3", volumeMultiplier: 1.0 },
      { key: "asteroid-impact", path: "/sounds/asteroid-impact.mp3", volumeMultiplier: 1.0 },
      { key: "your-turn", path: "/sounds/your-turn.mp3", volumeMultiplier: 1.0 },
    ];

    audioFiles.forEach(({ key, path, volumeMultiplier }) => {
      try {
        const audio = new Audio(path);
        audio.preload = "auto";
        audio.volume = this.volume * volumeMultiplier;

        audio.addEventListener("error", (e) => {
          console.warn(`Failed to preload audio: ${key}`, e);
        });

        this.audioCache.set(key, audio);
        this.volumeMultipliers.set(key, volumeMultiplier);
      } catch (error) {
        console.warn(`Error creating audio element for ${key}:`, error);
      }
    });
  }

  public async playSound(soundKey: string): Promise<void> {
    if (!this.isEnabled) {
      return;
    }

    const audio = this.audioCache.get(soundKey);
    if (!audio) {
      console.warn(`Sound not found: ${soundKey}`);
      return;
    }

    try {
      const audioClone = audio.cloneNode() as HTMLAudioElement;
      const multiplier = this.volumeMultipliers.get(soundKey) ?? 1.0;
      audioClone.volume = this.volume * multiplier;

      await audioClone.play();
    } catch (error) {
      console.warn(`Failed to play sound ${soundKey}:`, error);
    }
  }

  public async playProductionSound(): Promise<void> {
    return this.playSound("production");
  }

  public async playTemperatureSound(): Promise<void> {
    return this.playSound("temperature-increase");
  }

  public async playWaterPlacementSound(): Promise<void> {
    return this.playSound("water-placement");
  }

  public async playOxygenSound(): Promise<void> {
    return this.playSound("oxygen-increase");
  }

  public async playButtonHoverSound(): Promise<void> {
    return this.playSound("button-hover");
  }

  public async playButtonClickSound(): Promise<void> {
    return this.playSound("button-click");
  }

  public async playCardHoverSound(): Promise<void> {
    return this.playSound("card-hover");
  }

  public async playConstructionSound(): Promise<void> {
    return this.playSound("construction");
  }

  public async playAsteroidImpactSound(): Promise<void> {
    return this.playSound("asteroid-impact");
  }

  public async playYourTurnSound(): Promise<void> {
    return this.playSound("your-turn");
  }

  private createAmbientAudio(): HTMLAudioElement {
    const audio = new Audio(this.ambientTracks[this.currentTrackIndex]);
    audio.loop = false;
    audio.addEventListener("ended", () => {
      this.currentTrackIndex = (this.currentTrackIndex + 1) % this.ambientTracks.length;
      this.ambientAudio = this.createAmbientAudio();
      this.ambientAudio.volume = this.musicVolume * this.ambientVolumeMultiplier;
      if (this.isMusicEnabled) {
        void this.ambientAudio.play().catch(() => {});
      }
    });
    return audio;
  }

  public playAmbient(): void {
    if (this.fadeOutInterval !== null) {
      clearInterval(this.fadeOutInterval);
      this.fadeOutInterval = null;
    }

    if (!this.ambientAudio) {
      this.ambientAudio = this.createAmbientAudio();
    }
    this.ambientAudio.volume = this.musicVolume * this.ambientVolumeMultiplier;

    if (this.isMusicEnabled) {
      void this.ambientAudio.play().catch(() => {});
    }
  }

  private fadeOut(audio: HTMLAudioElement, duration: number = 300): void {
    const steps = 15;
    const interval = duration / steps;
    const volumeStep = audio.volume / steps;

    this.fadeOutInterval = setInterval(() => {
      audio.volume = Math.max(0, audio.volume - volumeStep);
      if (audio.volume <= 0.01) {
        if (this.fadeOutInterval !== null) {
          clearInterval(this.fadeOutInterval);
          this.fadeOutInterval = null;
        }
        audio.pause();
        audio.currentTime = 0;
        audio.volume = this.musicVolume * this.ambientVolumeMultiplier;
      }
    }, interval);
  }

  public stopAmbient(): void {
    if (this.ambientAudio) {
      this.fadeOut(this.ambientAudio);
    }
  }

  public stopAmbientWithDuration(duration: number): void {
    if (this.ambientAudio) {
      this.fadeOut(this.ambientAudio, duration);
    }
  }

  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
    if (this.ambientAudio) {
      if (enabled) {
        void this.ambientAudio.play().catch(() => {});
      } else {
        this.fadeOut(this.ambientAudio);
      }
    }
  }

  public setVolume(volume: number): void {
    this.volume = Math.max(0, Math.min(1, volume));

    this.audioCache.forEach((audio, key) => {
      const multiplier = this.volumeMultipliers.get(key) ?? 1.0;
      audio.volume = this.volume * multiplier;
    });
  }

  public setMusicVolume(volume: number): void {
    this.musicVolume = Math.max(0, Math.min(1, volume));

    if (this.ambientAudio) {
      this.ambientAudio.volume = this.musicVolume * this.ambientVolumeMultiplier;
    }
  }

  public setMusicEnabled(enabled: boolean): void {
    this.isMusicEnabled = enabled;
    if (this.ambientAudio) {
      if (enabled) {
        void this.ambientAudio.play().catch(() => {});
      } else {
        this.fadeOut(this.ambientAudio);
      }
    }
  }

  public getSettings() {
    return {
      enabled: this.isEnabled,
      musicEnabled: this.isMusicEnabled,
      volume: this.volume,
      musicVolume: this.musicVolume,
    };
  }
}

export const audioService = new AudioService();
export default audioService;
