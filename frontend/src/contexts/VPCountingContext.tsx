import { createContext, useContext, type ReactNode } from "react";

export interface VPPhase {
  id: string;
  label: string;
  color: string;
}

export interface VPCountingState {
  isActive: boolean;
  currentPhaseIndex: number;
  currentPhaseId: string | null;
  activePlayerId: string | null;
  playerAccumulatedVP: Record<string, number>;
  highlightedTiles: Set<string>;
  secondaryHighlightedTiles: Set<string>;
  isFinishing: boolean;
  isComplete: boolean;
}

export interface VPCountingControls {
  start: () => void;
  skip: () => void;
  goToPhase: (index: number) => void;
  phases: VPPhase[];
}

interface VPCountingContextValue {
  state: VPCountingState;
  controls: VPCountingControls;
}

const defaultState: VPCountingState = {
  isActive: false,
  currentPhaseIndex: -1,
  currentPhaseId: null,
  activePlayerId: null,
  playerAccumulatedVP: {},
  highlightedTiles: new Set(),
  secondaryHighlightedTiles: new Set(),
  isFinishing: false,
  isComplete: false,
};

const defaultControls: VPCountingControls = {
  start: () => {},
  skip: () => {},
  goToPhase: () => {},
  phases: [],
};

const VPCountingContext = createContext<VPCountingContextValue>({
  state: defaultState,
  controls: defaultControls,
});

export function VPCountingProvider({
  state,
  controls,
  children,
}: {
  state: VPCountingState;
  controls: VPCountingControls;
  children: ReactNode;
}) {
  return (
    <VPCountingContext.Provider value={{ state, controls }}>{children}</VPCountingContext.Provider>
  );
}

export function useVPCounting(): VPCountingContextValue {
  return useContext(VPCountingContext);
}
