import { createContext, useContext } from "react";
import type * as THREE from "three";

export interface NuclearEffect {
  object: THREE.Group;
  seed: number;
  age: number;
}

export const NuclearEffectsContext = createContext<Set<NuclearEffect> | null>(null);

export function useNuclearEffects() {
  const effects = useContext(NuclearEffectsContext);
  if (!effects) {
    throw new Error("Nuclear effects require the scene renderer");
  }
  return effects;
}
