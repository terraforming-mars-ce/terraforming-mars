import { createContext, useContext } from "react";
import type * as THREE from "three";
import type { AtmosphereProfile } from "../components/game/board/solarSystemConfig";

// Temporary A/B switch: reload with ?atmosphere=off to remove atmosphere CPU/GPU work.
export const ATMOSPHERE_ENABLED =
  !import.meta.env.DEV || new URLSearchParams(window.location.search).get("atmosphere") !== "off";

export interface AtmosphereBody {
  object: THREE.Object3D;
  radius: number;
  profile: AtmosphereProfile;
  halo: THREE.ShaderMaterial;
  followsMesh: boolean;
}

export const AtmospheresContext = createContext<Set<AtmosphereBody> | null>(null);

export function useAtmospheres() {
  const bodies = useContext(AtmospheresContext);
  if (!bodies) {
    throw new Error("PlanetAtmosphere must be inside AtmosphereRenderer");
  }
  return bodies;
}
