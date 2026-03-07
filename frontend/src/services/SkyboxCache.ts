import * as THREE from "three";
import { EXRLoader } from "three/examples/jsm/loaders/EXRLoader.js";

export interface SkyboxLoadingState {
  isLoading: boolean;
  isLoaded: boolean;
  error: Error | null;
  texture: THREE.Texture | null;
}

class SkyboxCacheService {
  private static instance: SkyboxCacheService;
  private loadingState: SkyboxLoadingState = {
    isLoading: false,
    isLoaded: false,
    error: null,
    texture: null,
  };
  private loadPromise: Promise<THREE.Texture> | null = null;
  private currentPath: string | null = null;
  private textureCache: Map<string, THREE.Texture> = new Map();
  private listeners: Set<(state: SkyboxLoadingState) => void> = new Set();

  static getInstance(): SkyboxCacheService {
    if (!SkyboxCacheService.instance) {
      SkyboxCacheService.instance = new SkyboxCacheService();
    }
    return SkyboxCacheService.instance;
  }

  subscribe(listener: (state: SkyboxLoadingState) => void): () => void {
    this.listeners.add(listener);
    listener({ ...this.loadingState });

    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners() {
    this.listeners.forEach((listener) => {
      listener({ ...this.loadingState });
    });
  }

  async loadSkybox(path?: string): Promise<THREE.Texture> {
    const targetPath = path ?? "/assets/backgrounds/starmap_2020_8k.exr";

    const cached = this.textureCache.get(targetPath);
    if (cached) {
      this.currentPath = targetPath;
      this.loadingState = { isLoading: false, isLoaded: true, error: null, texture: cached };
      this.notifyListeners();
      return cached;
    }

    if (this.loadPromise && this.currentPath === targetPath) {
      return this.loadPromise;
    }

    this.currentPath = targetPath;
    this.loadingState = { isLoading: true, isLoaded: false, error: null, texture: null };
    this.notifyListeners();

    this.loadPromise = new Promise<THREE.Texture>((resolve, reject) => {
      const loader = new EXRLoader();

      loader.load(
        targetPath,
        (texture) => {
          try {
            texture.mapping = THREE.EquirectangularReflectionMapping;
            texture.colorSpace = THREE.LinearSRGBColorSpace;

            this.textureCache.set(targetPath, texture);
            this.loadingState = { isLoading: false, isLoaded: true, error: null, texture };
            this.notifyListeners();
            this.loadPromise = null;
            resolve(texture);
          } catch (error) {
            const err =
              error instanceof Error ? error : new Error("Failed to configure skybox texture");
            this.loadingState = { isLoading: false, isLoaded: false, error: err, texture: null };
            this.notifyListeners();
            this.loadPromise = null;
            reject(err);
          }
        },
        (_progress) => {},
        (error) => {
          const err = error instanceof Error ? error : new Error("Failed to load EXR skybox");
          this.loadingState = { isLoading: false, isLoaded: false, error: err, texture: null };
          this.notifyListeners();
          this.loadPromise = null;
          reject(err);
        },
      );
    });

    return this.loadPromise;
  }

  getState(): SkyboxLoadingState {
    return { ...this.loadingState };
  }

  preload(path?: string): Promise<THREE.Texture> {
    return this.loadSkybox(path);
  }

  isReady(): boolean {
    return this.loadingState.isLoaded && this.loadingState.texture !== null;
  }
}

// Export singleton instance
export const skyboxCache = SkyboxCacheService.getInstance();
