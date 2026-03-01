import { createContext, useContext, useState, ReactNode } from "react";

type PlanetTarget = "mars" | "venus";

interface PlanetFocusContextType {
  activePlanet: PlanetTarget;
  setActivePlanet: (planet: PlanetTarget) => void;
  isTransitioning: boolean;
  setIsTransitioning: (transitioning: boolean) => void;
}

const PlanetFocusContext = createContext<PlanetFocusContextType | null>(null);

export function PlanetFocusProvider({ children }: { children: ReactNode }) {
  const [activePlanet, setActivePlanet] = useState<PlanetTarget>("mars");
  const [isTransitioning, setIsTransitioning] = useState(false);

  return (
    <PlanetFocusContext.Provider
      value={{ activePlanet, setActivePlanet, isTransitioning, setIsTransitioning }}
    >
      {children}
    </PlanetFocusContext.Provider>
  );
}

export function usePlanetFocus() {
  const context = useContext(PlanetFocusContext);
  if (!context) {
    throw new Error("usePlanetFocus must be used within a PlanetFocusProvider");
  }
  return context;
}
