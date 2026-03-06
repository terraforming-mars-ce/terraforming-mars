import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import { skyboxCache, SkyboxLoadingState } from "../services/SkyboxCache.ts";

interface SpaceBackgroundContextType {
  isLoading: boolean;
  isLoaded: boolean;
  error: Error | null;
  preloadSkybox: () => Promise<void>;
}

const SpaceBackgroundContext = createContext<SpaceBackgroundContextType | undefined>(undefined);

/**
 * SpaceBackgroundProvider - Manages 3D space background state globally
 * Ensures skybox stays loaded in memory across route changes
 */
export function SpaceBackgroundProvider({ children }: { children: React.ReactNode }) {
  const [loadingState, setLoadingState] = useState<SkyboxLoadingState>({
    isLoading: false,
    isLoaded: false,
    error: null,
    texture: null,
  });

  useEffect(() => {
    const unsubscribe = skyboxCache.subscribe((state) => {
      setLoadingState(state);
    });

    if (skyboxCache.isReady()) {
      setLoadingState(skyboxCache.getState());
    }

    return unsubscribe;
  }, []);

  const preloadSkybox = useCallback(async () => {
    try {
      await skyboxCache.preload();
    } catch (error) {
      console.error("Failed to preload skybox:", error);
    }
  }, []);

  const contextValue: SpaceBackgroundContextType = {
    isLoading: loadingState.isLoading,
    isLoaded: loadingState.isLoaded,
    error: loadingState.error,
    preloadSkybox,
  };

  return (
    <SpaceBackgroundContext.Provider value={contextValue}>
      {children}
    </SpaceBackgroundContext.Provider>
  );
}

/**
 * Hook to access space background state
 */
export function useSpaceBackground() {
  const context = useContext(SpaceBackgroundContext);
  if (context === undefined) {
    throw new Error("useSpaceBackground must be used within a SpaceBackgroundProvider");
  }
  return context;
}
