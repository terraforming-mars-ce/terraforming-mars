import * as THREE from "three";

export const NUCLEAR_PLAYBACK_RATE = 0.8;
export function nuclearElapsed(realSeconds: number) {
  return realSeconds * NUCLEAR_PLAYBACK_RATE;
}

export const NUCLEAR_FLIGHT_SECONDS = 0.8;
export const NUCLEAR_DURATION_SECONDS = 7;
export const NUCLEAR_WAVE_SECONDS = 2.8;
export const NUCLEAR_WAVE_DUST_SECONDS = nuclearElapsed(1);
export const NUCLEAR_WAVE_RADIUS = 0.98;
export const NUCLEAR_SEGMENTS = 128;

export function createNuclearWaveGeometry() {
  // Radial subdivisions keep the projected shockwave above the curved planet.
  return new THREE.RingGeometry(0, NUCLEAR_WAVE_RADIUS + 0.08, 128, 40).translate(0, 0, 0.006);
}

const smooth = (a: number, b: number, x: number) => {
  const t = THREE.MathUtils.clamp((x - a) / (b - a), 0, 1);
  return t * t * (3 - 2 * t);
};

export function nuclearRandom(seed: number) {
  let state = seed | 0;
  return () => {
    state |= 0;
    state = (state + 0x6d2b79f5) | 0;
    let n = Math.imul(state ^ (state >>> 15), 1 | state);
    n = (n + Math.imul(n ^ (n >>> 7), 61 | n)) ^ n;
    return ((n ^ (n >>> 14)) >>> 0) / 4294967296;
  };
}

export interface NuclearCraterProfile {
  center: THREE.Vector2;
  floorOffset: THREE.Vector2;
  depth: number;
  rotation: number;
  aspect: number;
  radii: Float64Array;
  rimHeights: Float64Array;
  breaches: THREE.Vector3[];
  fans: number[];
  gouges: THREE.Vector4[];
}
const profiles = new Map<number, NuclearCraterProfile>();
const angleDistance = (a: number, b: number) => Math.atan2(Math.sin(a - b), Math.cos(a - b));

export function nuclearProfile(seed: number): NuclearCraterProfile {
  const cached = profiles.get(seed);
  if (cached) {
    return cached;
  }
  const random = nuclearRandom(seed);
  const rotation = random() * Math.PI * 2;
  const aspect = 1.08 + random() * 0.12;
  const center = new THREE.Vector2((random() - 0.5) * 0.009, (random() - 0.5) * 0.009);
  const floorOffset = new THREE.Vector2((random() - 0.5) * 0.026, (random() - 0.5) * 0.026);
  const depth = 0.014 + random() * 0.004;
  const count = 5 + Math.floor(random() * 3);
  const lobes = Array.from({ length: count }, (_, i) => ({
    angle: rotation + ((i + random() * 0.65) * Math.PI * 2) / count,
    width: 0.36 + random() * 0.3,
    amount: (i % 2 === 0 ? -1 : 1) * (0.005 + random() * 0.007),
  }));
  const fans = Array.from(
    { length: 3 },
    (_, i) => rotation + ((i + random() * 0.6) * Math.PI * 2) / 3,
  );
  const radii = Float64Array.from({ length: NUCLEAR_SEGMENTS }, (_, i) => {
    const angle = (i * Math.PI * 2) / NUCLEAR_SEGMENTS;
    const c = Math.cos(angle - rotation),
      sn = Math.sin(angle - rotation);
    let radius = 0.11 / Math.sqrt(c * c + aspect * aspect * sn * sn);
    for (const lobe of lobes) {
      radius += lobe.amount * Math.exp(-Math.pow(angleDistance(angle, lobe.angle) / lobe.width, 2));
    }
    return THREE.MathUtils.clamp(radius, 0.079, 0.121);
  });
  const gouges = Array.from({ length: 5 }, (_, i) => {
    const angle = fans[i % 3] + (random() - 0.5) * 0.45;
    const start = 0.052 + random() * 0.024;
    const end = start + 0.025 + random() * 0.025;
    return new THREE.Vector4(
      center.x + Math.cos(angle) * start,
      center.y + Math.sin(angle) * start,
      center.x + Math.cos(angle + (random() - 0.5) * 0.2) * end,
      center.y + Math.sin(angle + (random() - 0.5) * 0.2) * end,
    );
  });
  const fractureCount = 20 + Math.floor(random() * 11);
  const fractures = Array.from({ length: fractureCount }, (_, i) => ({
    angle: rotation + ((i + random() * 0.8) * Math.PI * 2) / fractureCount,
    width: 0.035 + random() * 0.075,
    height: (random() - 0.35) * 0.0024,
  }));
  const breaches = Array.from(
    { length: 3 + Math.floor(random() * 3) },
    (_, i) =>
      new THREE.Vector3(
        fans[i % 3] + (random() - 0.5) * 0.8,
        0.1 + random() * 0.13,
        0.4 + random() * 0.25,
      ),
  );
  const rimHeights = Float64Array.from({ length: NUCLEAR_SEGMENTS }, (_, i) => {
    const angle = (i * Math.PI * 2) / NUCLEAR_SEGMENTS;
    let height = 0.0075 + Math.sin(angle * 3 + rotation) * 0.001;
    for (const fracture of fractures) {
      height +=
        fracture.height *
        Math.max(0, 1 - Math.abs(angleDistance(angle, fracture.angle)) / fracture.width);
    }
    for (const breach of breaches) {
      height *= 1 - breach.z * Math.exp(-Math.pow(angleDistance(angle, breach.x) / breach.y, 2));
    }
    return THREE.MathUtils.clamp(height, 0.0015, 0.0095) * 1.65;
  });
  const profile = {
    center,
    floorOffset,
    depth,
    rotation,
    aspect,
    radii,
    rimHeights,
    breaches,
    fans,
    gouges,
  };
  if (profiles.size >= 128) {
    profiles.delete(profiles.keys().next().value!);
  }
  profiles.set(seed, profile);
  return profile;
}
export function nuclearRadius(angle: number, seed: number) {
  return angularSample(nuclearProfile(seed).radii, angle);
}
function angularSample(profile: Float64Array, angle: number) {
  const cycle = Math.PI * 2;
  const sample = ((((angle % cycle) + cycle) % cycle) / cycle) * NUCLEAR_SEGMENTS;
  const index = Math.floor(sample);
  return THREE.MathUtils.lerp(
    profile[index],
    profile[(index + 1) % NUCLEAR_SEGMENTS],
    sample - index,
  );
}
export function nuclearMoundRadius(angle: number, seed: number) {
  const profile = nuclearProfile(seed);
  return Math.min(
    nuclearRadius(angle, seed) + 0.023 + 0.004 * Math.sin(angle * 5 + profile.rotation),
    0.134 - profile.center.length(),
  );
}
export function nuclearFan(angle: number, seed: number) {
  let strength = 0;
  for (const fan of nuclearProfile(seed).fans) {
    strength = Math.max(strength, Math.exp(-Math.pow(angleDistance(angle, fan) / 0.25, 2)));
  }
  return strength;
}
export function nuclearSurfaceHeight(x: number, y: number, seed: number) {
  const profile = nuclearProfile(seed);
  const dx = x - profile.center.x,
    dy = y - profile.center.y;
  const angle = Math.atan2(dy, dx);
  const radius = nuclearRadius(angle, seed);
  const r = Math.hypot(dx, dy) / radius;
  const foot = nuclearMoundRadius(angle, seed) / radius;
  const bowl = -profile.depth * (1 - smooth(0.46, 1, r));
  const floorNoise =
    (Math.sin(x * 120 + profile.rotation) * Math.sin(y * 93 - profile.rotation) +
      Math.sin(x * 61 - y * 87)) *
    0.0006;
  const hollow =
    -0.0012 *
    Math.exp(-((dx - profile.floorOffset.x) ** 2 + (dy - profile.floorOffset.y) ** 2) / 0.00025);
  const innerRise = THREE.MathUtils.clamp((r - 0.62) / 0.38, 0, 1);
  const outerFall = 1 - THREE.MathUtils.clamp((r - 1) / (foot - 1), 0, 1);
  const rim = angularSample(profile.rimHeights, angle) * innerRise * outerFall * outerFall;
  return 0.0008 + bowl + (floorNoise + hollow) * (1 - smooth(0.45, 0.9, r)) + rim;
}

export interface NuclearGroundSample {
  height: number;
  coverage: number;
  borderDistance: number;
}
export type NuclearGroundSampler = (x: number, y: number) => NuclearGroundSample;
export const bareNuclearGround: NuclearGroundSampler = () => ({
  height: 0.0003,
  coverage: 1,
  borderDistance: 1,
});

export function nuclearGroundHeight(
  x: number,
  y: number,
  seed: number,
  ground: NuclearGroundSampler,
) {
  const profile = nuclearProfile(seed);
  const dx = x - profile.center.x,
    dy = y - profile.center.y;
  const angle = Math.atan2(dy, dx);
  const radius = nuclearRadius(angle, seed);
  const r = Math.hypot(dx, dy) / radius;
  const foot = nuclearMoundRadius(angle, seed) / radius;
  return (
    nuclearSurfaceHeight(x, y, seed) +
    (ground(profile.center.x, profile.center.y).height - 0.0003) * (1 - smooth(0.7, foot, r)) +
    (ground(x, y).height + 0.0006 - 0.0008) * smooth(0.7, foot, r)
  );
}

export interface NuclearProjection {
  origin: THREE.Vector3;
  rotation: THREE.Quaternion;
}
function disc(
  seed: number,
  rings: number,
  ground: NuclearGroundSampler,
  projection: NuclearProjection,
  kind: "mask" | "bowl" | "scorch",
) {
  const positions: number[] = [],
    uv: number[] = [],
    coverage: number[] = [],
    apron: number[] = [],
    planetDirections: number[] = [],
    indices: number[] = [];
  const profile = nuclearProfile(seed);
  const direction = new THREE.Vector3();
  const mask = kind === "mask",
    scorch = kind === "scorch";
  for (let ring = 0; ring <= rings; ring++) {
    const t = ring / rings;
    for (let segment = 0; segment <= NUCLEAR_SEGMENTS; segment++) {
      const angle = (segment / NUCLEAR_SEGMENTS) * Math.PI * 2;
      const crest = nuclearRadius(angle, seed);
      const foot = nuclearMoundRadius(angle, seed);
      let radius = crest * t;
      if (kind === "bowl") {
        if (t <= 0.25) {
          radius = crest * t * 1.8;
        } else if (t <= 0.5) {
          radius = crest * THREE.MathUtils.lerp(0.45, 0.83, (t - 0.25) * 4);
        } else if (t <= 0.75) {
          radius = crest * THREE.MathUtils.lerp(0.83, 1, (t - 0.5) * 4);
        } else {
          radius = THREE.MathUtils.lerp(crest, foot, (t - 0.75) * 4);
        }
      }
      if (scorch) {
        const debrisEdge = THREE.MathUtils.clamp(
          0.158 +
            nuclearFan(angle, seed) * 0.036 +
            Math.sin(angle * 3 + profile.rotation) * 0.007 +
            Math.sin(angle * 7 - profile.rotation) * 0.004,
          0.137,
          0.193,
        );
        radius = THREE.MathUtils.lerp(foot, debrisEdge, t);
      }
      const x = profile.center.x + Math.cos(angle) * radius;
      const y = profile.center.y + Math.sin(angle) * radius;
      positions.push(x, y, mask ? 0.003 : nuclearGroundHeight(x, y, seed, ground));
      uv.push(segment / NUCLEAR_SEGMENTS, radius / crest);
      coverage.push(scorch ? ground(x, y).coverage : 1);
      apron.push(scorch ? t : 0);
      direction
        .set(x, y, 0)
        .applyQuaternion(projection.rotation)
        .add(projection.origin)
        .normalize();
      planetDirections.push(direction.x, direction.y, direction.z);
      if (ring < rings && segment < NUCLEAR_SEGMENTS) {
        const a = ring * (NUCLEAR_SEGMENTS + 1) + segment;
        const b = a + NUCLEAR_SEGMENTS + 1;
        indices.push(a, b, b + 1, a, b + 1, a + 1);
      }
    }
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uv, 2));
  geometry.setAttribute("nuclearCoverage", new THREE.Float32BufferAttribute(coverage, 1));
  geometry.setAttribute("nuclearApron", new THREE.Float32BufferAttribute(apron, 1));
  geometry.setAttribute(
    "nuclearPlanetDirection",
    new THREE.Float32BufferAttribute(planetDirections, 3),
  );
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  geometry.computeBoundingSphere();
  return geometry;
}
export function createNuclearGeometry(
  seed: number,
  ground: NuclearGroundSampler,
  projection: NuclearProjection,
) {
  return {
    mask: disc(seed, 1, ground, projection, "mask"),
    bowl: disc(seed, 40, ground, projection, "bowl"),
    scorch: disc(seed, 20, ground, projection, "scorch"),
  };
}

export function nuclearPhase(elapsed: number) {
  if (elapsed < NUCLEAR_FLIGHT_SECONDS) {
    return "flight";
  }
  if (elapsed < NUCLEAR_DURATION_SECONDS) {
    return "blast";
  }
  return "complete";
}

export function nuclearShake(age: number) {
  if (age < 0 || age >= 0.35) {
    return [0, 0] as const;
  }
  const amplitude = 1.5 * Math.pow(1 - age / 0.35, 2);
  return [Math.sin(age * 90) * amplitude * 0.8, Math.sin(age * 113) * amplitude * 0.6] as const;
}
