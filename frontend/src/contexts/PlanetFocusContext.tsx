import {
  createContext,
  useContext,
  useRef,
  useState,
  useCallback,
  ReactNode,
  MutableRefObject,
} from "react";

export type PlanetTarget =
  | "mars"
  | "venus"
  | "jupiter"
  | "earth"
  | "mercury"
  | "saturn"
  | "neptune"
  | "uranus"
  | "ceres"
  | "orbital-station"
  | "solar-system";

interface PlanetFocusContextType {
  activePlanet: PlanetTarget;
  setActivePlanet: (planet: PlanetTarget) => void;
  previousPlanetRef: MutableRefObject<PlanetTarget>;
}

const PlanetFocusContext = createContext<PlanetFocusContextType | null>(null);

export function PlanetFocusProvider({ children }: { children: ReactNode }) {
  const [activePlanet, setActivePlanetRaw] = useState<PlanetTarget>("mars");
  const previousPlanetRef = useRef<PlanetTarget>("mars");
  const activePlanetRef = useRef<PlanetTarget>("mars");

  const setActivePlanet = useCallback((planet: PlanetTarget) => {
    if (planet === activePlanetRef.current) {
      return;
    }
    previousPlanetRef.current = activePlanetRef.current;
    activePlanetRef.current = planet;
    setActivePlanetRaw(planet);
  }, []);

  return (
    <PlanetFocusContext.Provider
      value={{
        activePlanet,
        setActivePlanet,
        previousPlanetRef,
      }}
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
