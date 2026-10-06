import * as THREE from "three";
import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";
import { SPHERE_RADIUS } from "./boardConstants";
import { boardCenter, hashSeed, projectBoardPoint } from "./landscapeGeometry";
import { landscapeHeightAt, smooth, WATER_LEVEL } from "./landscapeFields";
import type { NuclearTransitions } from "./nuclearTransitions";
import { NUCLEAR_FLIGHT_SECONDS } from "./nuclearGeometry";
import type { LandscapeState } from "./landscapeTypes";
import {
  nuclearElapsed,
  nuclearGroundHeight,
  nuclearProfile,
  nuclearRadius,
  nuclearRandom,
  type NuclearGroundSampler,
} from "./nuclearGeometry";

export const NUCLEAR_DEBRIS_SECONDS = nuclearElapsed(1.5);
const UP = new THREE.Vector3(0, 0, 1);
const HEX_APOTHEM = (0.166 * Math.sqrt(3)) / 2;
export interface NuclearBoardTile {
  coordinate: HexCoordinate;
  kind: string;
}
export interface NuclearSite {
  key: string;
  seed: number;
  origin: THREE.Vector3;
  rotation: THREE.Quaternion;
  ground: NuclearGroundSampler;
}
export function nuclearDirection(site: NuclearSite, x: number, y: number, target: THREE.Vector3) {
  return target.set(x, y, 0).applyQuaternion(site.rotation).add(site.origin).normalize();
}
export function nuclearGroundNormal(
  site: NuclearSite,
  x: number,
  y: number,
  target: THREE.Vector3,
) {
  const position = (dx: number, dy: number) =>
    nuclearDirection(site, dx, dy, new THREE.Vector3()).multiplyScalar(
      SPHERE_RADIUS + nuclearGroundHeight(dx, dy, site.seed, site.ground),
    );
  const tangentX = position(x + 0.0005, y).sub(position(x - 0.0005, y));
  const tangentY = position(x, y + 0.0005).sub(position(x, y - 0.0005));
  return target.crossVectors(tangentX, tangentY).normalize();
}

export function createNuclearSites(
  tiles: NuclearBoardTile[],
  gameId: string,
  landscape: LandscapeState,
) {
  const frames = tiles.map((tile) => {
    const center = boardCenter(tile.coordinate);
    const normal = projectBoardPoint(center.x, center.y).normalize();
    const rotation = new THREE.Quaternion().setFromUnitVectors(UP, normal);
    return {
      key: HexGrid2D.coordinateToKey(tile.coordinate),
      kind: tile.kind,
      normal,
      rotation,
      inverse: rotation.clone().invert(),
      origin: normal.clone().multiplyScalar(SPHERE_RADIUS + 0.01),
    };
  });
  const sites = new Map<string, NuclearSite>();
  for (const frame of frames) {
    if (frame.kind !== "nuclear-zone") {
      continue;
    }
    const nearby = frames.filter((other) => other.origin.distanceTo(frame.origin) < 0.6);
    const direction = new THREE.Vector3(),
      local = new THREE.Vector3();
    const site: NuclearSite = {
      key: frame.key,
      seed: hashSeed(`${gameId}:${frame.key}`),
      origin: frame.origin,
      rotation: frame.rotation,
      ground: (x, y) => {
        nuclearDirection(site, x, y, direction);
        let clearance = -Infinity;
        let owner = frame;
        for (const other of nearby) {
          local
            .copy(direction)
            .multiplyScalar((SPHERE_RADIUS + 0.01) / direction.dot(other.normal))
            .sub(other.origin)
            .applyQuaternion(other.inverse);
          const edge =
            HEX_APOTHEM -
            Math.max(
              Math.abs(local.x),
              Math.abs(local.x * 0.5 + (local.y * Math.sqrt(3)) / 2),
              Math.abs(local.x * -0.5 + (local.y * Math.sqrt(3)) / 2),
            );
          if (edge > clearance) {
            clearance = edge;
            owner = other;
          }
        }
        const allowed = owner === frame || owner.kind === "empty" || owner.kind === "greenery";
        const d = Math.hypot(direction.x, direction.y);
        const scale =
          d > 1e-10
            ? (Math.acos(THREE.MathUtils.clamp(direction.z, -1, 1)) * SPHERE_RADIUS) / d
            : 0;
        const height = landscapeHeightAt(
          landscape.patches,
          direction.x * scale,
          direction.y * scale,
          landscape.relief,
        );
        return {
          height,
          coverage: allowed && height > WATER_LEVEL ? smooth(-0.006, 0.012, clearance) : 0,
          borderDistance: Math.abs(clearance),
        };
      },
    };
    sites.set(site.key, site);
  }
  return sites;
}

export interface NuclearFragment {
  x: number;
  y: number;
  size: number;
  zone: "rim" | "spill";
  start: THREE.Vector3;
  end: THREE.Vector3;
  up: THREE.Vector3;
  rotation: THREE.Quaternion;
  launchRotation: THREE.Quaternion;
  scale: THREE.Vector3;
  color: THREE.Color;
  delay: number;
  duration: number;
  arc: number;
}
export function createNuclearFragments(site: NuclearSite) {
  const random = nuclearRandom(site.seed ^ 0x5482);
  const profile = nuclearProfile(site.seed);
  const fragments: NuclearFragment[] = [];
  for (let i = 0; i < 144; i++) {
    const spill = i >= 120;
    const zone: NuclearFragment["zone"] = spill ? "spill" : "rim";
    const fan = profile.fans[i % 3];
    let accepted: { x: number; y: number; size: number } | undefined;
    for (let attempt = 0; attempt < (spill ? 1 : 12); attempt++) {
      let angle = random() * Math.PI * 2;
      if (spill) {
        angle = fan + (random() - 0.5) * 0.55;
      } else if (random() < 0.3) {
        const breach = profile.breaches[Math.floor(random() * profile.breaches.length)];
        angle = breach.x + (random() - 0.5) * (breach.y * 5);
      }
      let radius = nuclearRadius(angle, site.seed) * (0.99 + random() * 0.17);
      if (spill) {
        radius = 0.17 + random() * 0.025;
      }
      const roll = random();
      let size = 0.0015 + random() * 0.002;
      if (roll > 0.97) {
        size = 0.006 + random() * 0.003;
      } else if (roll > 0.75) {
        size = 0.0035 + random() * 0.0025;
      }
      if (spill) {
        size = Math.min(size, 0.004);
      }
      const x = profile.center.x + Math.cos(angle) * radius;
      const y = profile.center.y + Math.sin(angle) * radius;
      const ground = site.ground(x, y);
      if (ground.coverage < 0.99 || ground.borderDistance < 0.006 + size * 0.5) {
        continue;
      }
      accepted = { x, y, size };
      break;
    }
    if (!accepted) {
      continue;
    }
    const { x, y, size } = accepted;
    const up = nuclearDirection(site, x, y, new THREE.Vector3());
    const height = nuclearGroundHeight(x, y, site.seed, site.ground);
    const end = up.clone().multiplyScalar(SPHERE_RADIUS + height - size * (0.15 + random() * 0.2));
    const start = nuclearDirection(site, x * 0.3, y * 0.3, new THREE.Vector3()).multiplyScalar(
      SPHERE_RADIUS + 0.004,
    );
    const surfaceNormal = nuclearGroundNormal(site, x, y, new THREE.Vector3());
    const rotation = new THREE.Quaternion()
      .setFromUnitVectors(UP, surfaceNormal)
      .multiply(
        new THREE.Quaternion().setFromEuler(
          new THREE.Euler((random() - 0.5) * 0.28, (random() - 0.5) * 0.28, random() * Math.PI * 2),
        ),
      );
    const launchRotation = rotation
      .clone()
      .multiply(
        new THREE.Quaternion().setFromEuler(
          new THREE.Euler(random() * 3, random() * 3, random() * 3),
        ),
      );
    const shade = 0.65 + random() * 0.4;
    const delay = nuclearElapsed(random() * 0.12);
    fragments.push({
      x,
      y,
      size,
      zone,
      up,
      start,
      end,
      rotation,
      launchRotation,
      scale: new THREE.Vector3(1, 0.65 + random() * 0.45, 0.6 + random() * 0.45).multiplyScalar(
        size / 0.04,
      ),
      color: new THREE.Color(0.75 * shade, 0.52 * shade, 0.35 * shade),
      delay,
      duration: nuclearElapsed(1 + random() * 0.38),
      arc: 0.025 + random() * 0.04,
    });
  }
  return fragments;
}
export function debrisScratch() {
  return { position: new THREE.Vector3(), rotation: new THREE.Quaternion() };
}
export function nuclearFragmentMatrix(
  fragment: NuclearFragment,
  age: number,
  target: THREE.Matrix4,
  scratch: ReturnType<typeof debrisScratch>,
) {
  if (age < fragment.delay) {
    return target.makeScale(0, 0, 0);
  }
  const t = THREE.MathUtils.clamp((age - fragment.delay) / fragment.duration, 0, 1);
  scratch.position
    .lerpVectors(fragment.start, fragment.end, t)
    .addScaledVector(fragment.up, fragment.arc * 4 * t * (1 - t));
  scratch.rotation.slerpQuaternions(fragment.launchRotation, fragment.rotation, t);
  return target.compose(scratch.position, scratch.rotation, fragment.scale);
}

export class NuclearDebrisBuffer {
  private readonly entries;
  private readonly scratch = { ...debrisScratch(), matrix: new THREE.Matrix4() };
  private pending = false;
  readonly count: number;
  constructor(sites: ReadonlyMap<string, NuclearSite>) {
    this.entries = [...sites.values()].flatMap((site) =>
      createNuclearFragments(site).map((fragment) => ({ fragment, key: site.key, done: false })),
    );
    this.count = this.entries.length;
  }
  write(
    target: THREE.InstancedMesh,
    transitions: NuclearTransitions,
    now: number,
    initialize = false,
  ) {
    if (!initialize && !this.pending) {
      return;
    }
    if (initialize) {
      target.count = this.count;
    }
    let dirty = false;
    this.pending = false;
    for (let i = 0; i < this.entries.length; i++) {
      const entry = this.entries[i];
      if (entry.done && !initialize) {
        continue;
      }
      const transition = transitions.get(entry.key);
      const age = transition
        ? nuclearElapsed(now - transition.start) - NUCLEAR_FLIGHT_SECONDS
        : Infinity;
      entry.done = age >= NUCLEAR_DEBRIS_SECONDS;
      this.pending ||= !entry.done;
      if (!initialize && age < 0) {
        continue;
      }
      nuclearFragmentMatrix(entry.fragment, age, this.scratch.matrix, this.scratch);
      target.setMatrixAt(i, this.scratch.matrix);
      target.instanceMatrix.addUpdateRange(i * 16, 16);
      if (initialize) {
        target.setColorAt(i, entry.fragment.color);
      }
      dirty = true;
    }
    if (dirty) {
      target.instanceMatrix.needsUpdate = true;
    }
    if (initialize && target.instanceColor) {
      target.instanceColor.needsUpdate = true;
    }
  }
}
