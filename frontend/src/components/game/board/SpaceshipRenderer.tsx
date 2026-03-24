import { useEffect, useMemo, useRef, useState } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils.js";
import { useModels } from "../../../hooks/useModels";
import { useTextures } from "../../../hooks/useTextures";
import { usePrimitiveInstances } from "./PrimitiveManager";
import { SPHERE_RADIUS, VENUS_POSITION, VENUS_RADIUS, easeOutCubic } from "./boardConstants";
import { GameDto, ColonyTileDto } from "../../../types/generated/api-types";
import { addSphereProjectionWithSoftEdges } from "./GreeneryRenderer";
import ExhaustEffect from "./effects/ExhaustEffect";
import EngineTrail from "./effects/EngineTrail";
import HoverSmoke from "./effects/HoverSmoke";
import EngineGlow from "./effects/EngineGlow";

let spaceshipCache: {
  geometry: THREE.BufferGeometry;
  material: THREE.Material;
} | null = null;

function extractSpaceshipPrimitive(scene: THREE.Group) {
  if (spaceshipCache) {
    return spaceshipCache;
  }

  const box = new THREE.Box3().setFromObject(scene);
  const size = box.getSize(new THREE.Vector3());
  const center = box.getCenter(new THREE.Vector3());
  const targetSize = 0.096;
  const scaleF = targetSize / Math.max(size.x, size.y, size.z);

  const geometries: THREE.BufferGeometry[] = [];

  scene.traverse((child) => {
    if (child instanceof THREE.Mesh) {
      child.updateWorldMatrix(true, false);
      const geo = child.geometry.clone();
      geo.applyMatrix4(child.matrixWorld);
      geometries.push(geo);
    }
  });

  if (geometries.length === 0) {
    return null;
  }

  const merged = mergeGeometries(geometries);
  if (!merged) {
    return null;
  }

  merged.translate(-center.x, -center.y, -center.z);
  merged.applyMatrix4(new THREE.Matrix4().makeScale(scaleF, scaleF, scaleF));
  merged.applyMatrix4(new THREE.Matrix4().makeRotationY(Math.PI));
  merged.computeVertexNormals();

  const material = new THREE.MeshStandardMaterial({
    color: new THREE.Color(0.6, 0.6, 0.62),
    metalness: 0.7,
    roughness: 0.35,
  });

  spaceshipCache = { geometry: merged, material };
  return spaceshipCache;
}

const SPACING = 0.17;
const PAD_GEO_RADIUS = 0.264;
const PAD_SOLID_RADIUS = 0.096;
const WARNING_LINE_RADIUS = 0.052;
const WARNING_LINE_WIDTH = 0.007;

const PHASE_LIFTOFF_MS = 1200;
const PHASE_ARC_MS = 12000;
const PHASE_CRUISE_MS = 6000;
const PHASE_LANDING_MS = 3000;
const TOTAL_FLIGHT_MS = PHASE_LIFTOFF_MS + PHASE_ARC_MS + PHASE_CRUISE_MS + PHASE_LANDING_MS;

const HOVER_HEIGHT = 0.15;

const VENUS_CENTER = new THREE.Vector3(VENUS_POSITION[0], VENUS_POSITION[1], VENUS_POSITION[2]);
const VENUS_LANDING_NORMAL = new THREE.Vector3()
  .subVectors(new THREE.Vector3(0, 0, 0), VENUS_CENTER)
  .normalize();
const VENUS_LANDING_POS = VENUS_CENTER.clone().addScaledVector(VENUS_LANDING_NORMAL, VENUS_RADIUS);

const FAR_AWAY_POS = new THREE.Vector3(50, 80, -200);
const FAR_AWAY_NORMAL = FAR_AWAY_POS.clone().normalize();

function easeInCubic(t: number): number {
  return t * t * t;
}

function easeInOutQuad(t: number): number {
  return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
}

function easeInQuad(t: number): number {
  return t * t;
}

function smoothstep(t: number): number {
  return t * t * (3 - 2 * t);
}

function cubicBezier(
  P0: THREE.Vector3,
  P1: THREE.Vector3,
  P2: THREE.Vector3,
  P3: THREE.Vector3,
  t: number,
): THREE.Vector3 {
  const u = 1 - t;
  return new THREE.Vector3()
    .addScaledVector(P0, u * u * u)
    .addScaledVector(P1, 3 * u * u * t)
    .addScaledVector(P2, 3 * u * t * t)
    .addScaledVector(P3, t * t * t);
}

function cubicBezierTangent(
  P0: THREE.Vector3,
  P1: THREE.Vector3,
  P2: THREE.Vector3,
  P3: THREE.Vector3,
  t: number,
): THREE.Vector3 {
  const u = 1 - t;
  const tangent = new THREE.Vector3();
  tangent.addScaledVector(new THREE.Vector3().subVectors(P1, P0), 3 * u * u);
  tangent.addScaledVector(new THREE.Vector3().subVectors(P2, P1), 6 * u * t);
  tangent.addScaledVector(new THREE.Vector3().subVectors(P3, P2), 3 * t * t);
  return tangent;
}

function pointOnSphere(
  theta: number,
  phi: number,
): { position: THREE.Vector3; normal: THREE.Vector3 } {
  const pos = new THREE.Vector3(
    SPHERE_RADIUS * Math.sin(phi) * Math.cos(theta),
    SPHERE_RADIUS * Math.sin(phi) * Math.sin(theta),
    SPHERE_RADIUS * Math.cos(phi),
  );
  return { position: pos, normal: pos.clone().normalize() };
}

const _tmpLookMatrix = new THREE.Matrix4();

function orientationFromDirection(forward: THREE.Vector3, up: THREE.Vector3): THREE.Quaternion {
  const z = forward.clone().normalize().negate();
  const x = new THREE.Vector3().crossVectors(up, z).normalize();
  const y = new THREE.Vector3().crossVectors(z, x).normalize();
  _tmpLookMatrix.makeBasis(x, y, z);
  return new THREE.Quaternion().setFromRotationMatrix(_tmpLookMatrix);
}

function getDestination(location: string): {
  landingPos: THREE.Vector3;
  landingNormal: THREE.Vector3;
} {
  if (location === "venus") {
    return { landingPos: VENUS_LANDING_POS.clone(), landingNormal: VENUS_LANDING_NORMAL.clone() };
  }
  return { landingPos: FAR_AWAY_POS.clone(), landingNormal: FAR_AWAY_NORMAL.clone() };
}

interface FlightPath {
  startPos: THREE.Vector3;
  startNormal: THREE.Vector3;
  liftoffEnd: THREE.Vector3;
  parkedQuat: THREE.Quaternion;
  arcP0: THREE.Vector3;
  arcP1: THREE.Vector3;
  arcP2: THREE.Vector3;
  arcP3: THREE.Vector3;
  cruiseStart: THREE.Vector3;
  cruiseEnd: THREE.Vector3;
  destApproach: THREE.Vector3;
  destLandingPos: THREE.Vector3;
  destLandingNormal: THREE.Vector3;
  landedQuat: THREE.Quaternion;
}

function computeFlightPath(parkedMatrix: THREE.Matrix4, location: string): FlightPath {
  const startPos = new THREE.Vector3().setFromMatrixPosition(parkedMatrix);
  const startNormal = startPos.clone().normalize();
  const parkedQuat = new THREE.Quaternion().setFromRotationMatrix(parkedMatrix);

  const liftoffEnd = startPos.clone().addScaledVector(startNormal, HOVER_HEIGHT);

  const { landingPos, landingNormal } = getDestination(location);
  const destCenter = landingPos.clone().addScaledVector(landingNormal, -VENUS_RADIUS);
  const destApproach = landingPos.clone().addScaledVector(landingNormal, 1.0);

  const arcP0 = liftoffEnd.clone();
  const parkedForward = new THREE.Vector3(0, 0, -1).applyQuaternion(parkedQuat).normalize();

  const arcP1 = liftoffEnd
    .clone()
    .addScaledVector(parkedForward, 17.0)
    .addScaledVector(startNormal, 8.0);

  const toVenus = new THREE.Vector3().subVectors(destCenter, liftoffEnd).normalize();
  const midpoint = new THREE.Vector3().lerpVectors(liftoffEnd, destApproach, 0.5);
  const midNormal = midpoint.clone().normalize();
  const arcP2 = midpoint.clone().addScaledVector(midNormal, 6.0).addScaledVector(toVenus, 10.0);

  const arcP3 = new THREE.Vector3().lerpVectors(liftoffEnd, destApproach, 0.6);

  const cruiseStart = arcP3.clone();
  const cruiseEnd = destApproach.clone();

  const landedQuat = orientationFromDirection(
    landingNormal.clone().negate(),
    new THREE.Vector3(0, 1, 0),
  );

  return {
    startPos,
    startNormal,
    liftoffEnd,
    parkedQuat,
    arcP0,
    arcP1,
    arcP2,
    arcP3,
    cruiseStart,
    cruiseEnd,
    destApproach,
    destLandingPos: landingPos,
    destLandingNormal: landingNormal,
    landedQuat,
  };
}

interface ShipState {
  playerId: string;
  status: "parked" | "flying" | "landed";
  startTime: number;
  parkedMatrix: THREE.Matrix4;
  flightPath: FlightPath | null;
  landingExhaustSpawned: boolean;
  currentQuat: THREE.Quaternion;
  currentPosition: THREE.Vector3;
  currentForward: THREE.Vector3;
  hoverActive: boolean;
  hoverSmokeDone: boolean;
}

interface ExhaustConfig {
  id: number;
  position: THREE.Vector3;
  normal: THREE.Vector3;
  particleCount?: number;
  particleScale?: number;
  duration?: number;
}

let nextExhaustId = 0;

const ENTRANCE_DELAY_MS = 1200;
const ENTRANCE_DURATION_MS = 400;

interface SpaceshipRendererProps {
  gameState?: GameDto;
  animateEntrance?: boolean;
  startHidden?: boolean;
  sphereCenter?: THREE.Vector3;
  groupInverseMatrix?: THREE.Matrix4;
}

export default function SpaceshipRenderer({
  gameState,
  animateEntrance = false,
  startHidden = false,
  sphereCenter,
  groupInverseMatrix,
}: SpaceshipRendererProps) {
  const { spaceshipScene } = useModels();
  const {
    concrete: concreteTexture,
    noiseMid: noiseTexture,
    noiseHigh: noiseHighTexture,
  } = useTextures();
  const [exhaustEffects, setExhaustEffects] = useState<ExhaustConfig[]>([]);
  const [entranceScale, setEntranceScale] = useState(animateEntrance || startHidden ? 0 : 1);
  const entranceStartRef = useRef<number | null>(null);
  const entranceDoneRef = useRef(!animateEntrance && !startHidden);

  useEffect(() => {
    if (animateEntrance && entranceDoneRef.current) {
      setEntranceScale(0);
      entranceStartRef.current = null;
      entranceDoneRef.current = false;
    }
  }, [animateEntrance]);

  const primitive = useMemo(() => {
    return extractSpaceshipPrimitive(spaceshipScene);
  }, [spaceshipScene]);

  const padGeometry = useMemo(() => {
    return new THREE.CircleGeometry(PAD_GEO_RADIUS, 48);
  }, []);

  const padConcreteTexture = useMemo(() => {
    if (!concreteTexture) {
      return null;
    }
    const tex = concreteTexture.clone();
    tex.repeat.set(2, 2);
    tex.needsUpdate = true;
    return tex;
  }, [concreteTexture]);

  const padMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      map: padConcreteTexture,
      color: new THREE.Color(0.3, 0.28, 0.26),
      roughness: 0.95,
      metalness: 0.05,
      side: THREE.DoubleSide,
      transparent: true,
      alphaTest: 0.01,
      depthWrite: false,
    });
    addSphereProjectionWithSoftEdges(
      mat,
      0.004,
      noiseTexture,
      noiseHighTexture,
      PAD_SOLID_RADIUS,
      sphereCenter,
      groupInverseMatrix,
    );
    return mat;
  }, [padConcreteTexture, noiseTexture, noiseHighTexture]);

  const warningLineGeometry = useMemo(() => {
    return new THREE.RingGeometry(
      WARNING_LINE_RADIUS - 0.001,
      WARNING_LINE_RADIUS + WARNING_LINE_WIDTH,
      56,
      1,
    );
  }, []);

  const players = useMemo(() => {
    if (!gameState) {
      return [];
    }
    const list: { id: string; color: string }[] = [];
    if (gameState.currentPlayer) {
      list.push({
        id: gameState.currentPlayer.id,
        color: gameState.currentPlayer.color || "#888888",
      });
    }
    if (gameState.otherPlayers) {
      for (const p of gameState.otherPlayers) {
        list.push({ id: p.id, color: p.color || "#888888" });
      }
    }
    return list;
  }, [gameState?.currentPlayer?.id, gameState?.otherPlayers?.length]);

  const warningLineMaterials = useMemo(() => {
    return players.map((p) => {
      const c = new THREE.Color(p.color);
      return new THREE.ShaderMaterial({
        uniforms: {
          uConcreteTexture: { value: concreteTexture ?? null },
          uSphereRadius: { value: SPHERE_RADIUS },
          uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
          uPlayerColor: { value: new THREE.Vector3(c.r, c.g, c.b) },
        },
        vertexShader: `
          uniform float uSphereRadius;
          uniform vec3 uSphereCenter;
          varying vec2 vPos;
          varying vec2 vUv;
          void main() {
            vPos = position.xy;
            vUv = uv;
            vec4 worldPos = modelMatrix * vec4(position, 1.0);
            vec3 sphereDir = normalize(worldPos.xyz - uSphereCenter);
            vec3 projected = uSphereCenter + sphereDir * (uSphereRadius + 0.007);
            gl_Position = projectionMatrix * viewMatrix * vec4(projected, 1.0);
          }
        `,
        fragmentShader: `
          uniform sampler2D uConcreteTexture;
          uniform vec3 uPlayerColor;
          varying vec2 vPos;
          varying vec2 vUv;
          void main() {
            vec3 concrete = texture2D(uConcreteTexture, vUv * 8.0).rgb;
            float angle = atan(vPos.y, vPos.x);
            float stripe = step(0.0, sin(angle * 24.0));
            vec3 black = vec3(0.08, 0.07, 0.06);
            vec3 paint = mix(black, uPlayerColor, stripe);
            vec3 color = concrete * 0.3 + paint * 0.6;
            gl_FragColor = vec4(color, 1.0);
          }
        `,
        side: THREE.DoubleSide,
      });
    });
  }, [concreteTexture, players]);

  const shipCount = players.length;

  const { setTransforms, setColors } = usePrimitiveInstances(
    "spaceship",
    primitive?.geometry ?? null,
    primitive?.material ?? null,
    20,
    true,
  );

  const { parkedMatrices, padPositions } = useMemo(() => {
    if (shipCount === 0) {
      return { parkedMatrices: [], padPositions: [] };
    }
    const basePhi = 0.82;
    const baseTheta = -0.55;
    const lineAngle = 0.15;
    const shipMats: THREE.Matrix4[] = [];
    const pads: { position: THREE.Vector3; quaternion: THREE.Quaternion }[] = [];
    const offset = ((shipCount - 1) * SPACING) / 2;
    const CIRCLE_NORMAL = new THREE.Vector3(0, 0, 1);

    // Shared forward direction: from parking area toward tile board center
    const { position: centerPos } = pointOnSphere(baseTheta, basePhi);
    const boardCenter = pointOnSphere(0, Math.PI / 2).position;
    const toBoard = new THREE.Vector3().subVectors(boardCenter, centerPos).normalize();

    for (let i = 0; i < shipCount; i++) {
      const t = i * SPACING - offset;
      const theta = baseTheta + t * Math.cos(lineAngle);
      const phi = basePhi + t * Math.sin(lineAngle);
      const { position, normal } = pointOnSphere(theta, phi);

      // Pad sits on sphere surface — align circle's +Z normal to sphere surface normal
      const padQuat = new THREE.Quaternion().setFromUnitVectors(CIRCLE_NORMAL, normal);
      pads.push({ position: position.clone(), quaternion: padQuat });

      // Ship sits close to surface
      position.addScaledVector(normal, 0.015);

      // Project board direction onto this ship's tangent plane for consistent facing
      const projected = toBoard.clone().addScaledVector(normal, -toBoard.dot(normal)).normalize();
      const finalQuat = orientationFromDirection(projected.negate(), normal);

      shipMats.push(new THREE.Matrix4().compose(position, finalQuat, new THREE.Vector3(1, 1, 1)));
    }
    return { parkedMatrices: shipMats, padPositions: pads };
  }, [shipCount]);

  const shipsRef = useRef<ShipState[]>([]);

  useEffect(() => {
    if (parkedMatrices.length === 0 || players.length === 0) {
      return;
    }
    shipsRef.current = parkedMatrices.map((mat, i) => ({
      playerId: players[i].id,
      status: "parked" as const,
      startTime: 0,
      parkedMatrix: mat.clone(),
      flightPath: null,
      landingExhaustSpawned: false,
      currentQuat: new THREE.Quaternion().setFromRotationMatrix(mat),
      currentPosition: new THREE.Vector3().setFromMatrixPosition(mat),
      currentForward: new THREE.Vector3(0, 0, -1)
        .applyQuaternion(new THREE.Quaternion().setFromRotationMatrix(mat))
        .normalize(),
      hoverActive: false,
      hoverSmokeDone: false,
    }));
    setTransforms(parkedMatrices);
    const colors = players.map(() => new THREE.Color(0.35, 0.35, 0.37));
    setColors(colors);
  }, [parkedMatrices, players, setTransforms, setColors]);

  const prevTradeStateRef = useRef<Map<string, boolean>>(new Map());
  const initializedRef = useRef(false);

  useEffect(() => {
    if (!gameState?.colonyTiles || shipsRef.current.length === 0) {
      return;
    }

    const colonies = gameState.colonyTiles;
    const prevState = prevTradeStateRef.current;
    const isFirstLoad = !initializedRef.current;

    if (isFirstLoad) {
      initializedRef.current = true;
      for (const colony of colonies) {
        if (colony.tradedThisGen && colony.traderId) {
          const ship = shipsRef.current.find(
            (s) => s.playerId === colony.traderId && s.status === "parked",
          );
          if (ship) {
            const fp = computeFlightPath(ship.parkedMatrix, colony.location || "venus");
            ship.status = "landed";
            ship.parkedMatrix = new THREE.Matrix4().compose(
              fp.destLandingPos,
              fp.landedQuat,
              new THREE.Vector3(1, 1, 1),
            );
            ship.currentPosition.copy(fp.destLandingPos);
            ship.currentQuat.copy(fp.landedQuat);
          }
        }
      }
      const matrices = shipsRef.current.map((s) => s.parkedMatrix);
      setTransforms(matrices);
    } else {
      const anyPrevTraded = Array.from(prevState.values()).some((v) => v);
      const allNowUntraded = colonies.every((c: ColonyTileDto) => !c.tradedThisGen);

      if (anyPrevTraded && allNowUntraded) {
        for (let i = 0; i < shipsRef.current.length; i++) {
          const ship = shipsRef.current[i];
          if (ship.status !== "parked") {
            ship.status = "parked";
            ship.flightPath = null;
            ship.landingExhaustSpawned = false;
            ship.hoverSmokeDone = false;
          }
          if (parkedMatrices[i]) {
            ship.parkedMatrix = parkedMatrices[i].clone();
            ship.currentPosition.setFromMatrixPosition(parkedMatrices[i]);
            ship.currentQuat.setFromRotationMatrix(parkedMatrices[i]);
          }
        }
        if (parkedMatrices.length > 0) {
          setTransforms(parkedMatrices);
        }
      }

      for (const colony of colonies) {
        const wasTraded = prevState.get(colony.id) ?? false;
        if (!wasTraded && colony.tradedThisGen && colony.traderId) {
          const ship = shipsRef.current.find(
            (s) => s.playerId === colony.traderId && s.status === "parked",
          );
          if (ship) {
            ship.status = "flying";
            ship.startTime = -1;
            ship.flightPath = computeFlightPath(ship.parkedMatrix, colony.location || "venus");
            ship.landingExhaustSpawned = false;
            ship.hoverSmokeDone = false;
          }
        }
      }
    }

    const newState = new Map<string, boolean>();
    for (const colony of colonies) {
      newState.set(colony.id, colony.tradedThisGen);
    }
    prevTradeStateRef.current = newState;
  }, [gameState?.colonyTiles, parkedMatrices, setTransforms]);

  useFrame((state) => {
    if (animateEntrance && !entranceDoneRef.current) {
      if (entranceStartRef.current === null) {
        entranceStartRef.current = state.clock.elapsedTime;
      }
      const elapsed = (state.clock.elapsedTime - entranceStartRef.current) * 1000;
      if (elapsed >= ENTRANCE_DELAY_MS) {
        const t = Math.min((elapsed - ENTRANCE_DELAY_MS) / ENTRANCE_DURATION_MS, 1);
        setEntranceScale(easeOutCubic(t));
        if (t >= 1) {
          entranceDoneRef.current = true;
        }
      }
    }

    const ships = shipsRef.current;
    if (ships.length === 0) {
      return;
    }

    let needsUpdate = false;
    const matrices: THREE.Matrix4[] = [];

    for (let i = 0; i < ships.length; i++) {
      const ship = ships[i];

      if (ship.status === "parked") {
        matrices.push(ship.parkedMatrix);
        continue;
      }

      if (ship.status === "landed") {
        matrices.push(ship.parkedMatrix);
        continue;
      }

      needsUpdate = true;
      const fp = ship.flightPath!;

      if (ship.startTime < 0) {
        ship.startTime = state.clock.elapsedTime * 1000;
      }

      const elapsed = state.clock.elapsedTime * 1000 - ship.startTime;
      let position: THREE.Vector3;
      let quat: THREE.Quaternion;

      const arcStart = PHASE_LIFTOFF_MS;
      const cruiseStart = arcStart + PHASE_ARC_MS;
      const landStart = cruiseStart + PHASE_CRUISE_MS;

      if (elapsed < PHASE_LIFTOFF_MS) {
        const t = easeInQuad(elapsed / PHASE_LIFTOFF_MS);
        position = fp.startPos.clone().addScaledVector(fp.startNormal, HOVER_HEIGHT * t);
        quat = fp.parkedQuat.clone();
      } else if (elapsed < cruiseStart) {
        const rawT = (elapsed - arcStart) / PHASE_ARC_MS;
        const t = Math.pow(rawT, 1.5);
        position = cubicBezier(fp.arcP0, fp.arcP1, fp.arcP2, fp.arcP3, t);
        const tangent = cubicBezierTangent(fp.arcP0, fp.arcP1, fp.arcP2, fp.arcP3, t).normalize();
        const outward = position.clone().normalize();
        const accelDir = new THREE.Vector3().crossVectors(tangent, outward).normalize();
        const bankUp = outward.clone().addScaledVector(accelDir, 0.6).normalize();
        const tangentQuat = orientationFromDirection(tangent, bankUp);
        const orientBlend = smoothstep(Math.min(rawT / 0.4, 1.0));
        quat = fp.parkedQuat.clone().slerp(tangentQuat, orientBlend);
      } else if (elapsed < landStart) {
        const t = easeInOutQuad((elapsed - cruiseStart) / PHASE_CRUISE_MS);
        position = new THREE.Vector3().lerpVectors(fp.cruiseStart, fp.cruiseEnd, t);
        const dir = new THREE.Vector3().subVectors(fp.cruiseEnd, fp.cruiseStart).normalize();
        quat = orientationFromDirection(dir, new THREE.Vector3(0, 1, 0));

        const rawT = (elapsed - cruiseStart) / PHASE_CRUISE_MS;
        if (rawT > 0.95 && !ship.landingExhaustSpawned) {
          ship.landingExhaustSpawned = true;
          setExhaustEffects((prev) => [
            ...prev,
            {
              id: nextExhaustId++,
              position: fp.destLandingPos.clone(),
              normal: fp.destLandingNormal.clone(),
            },
          ]);
        }
      } else if (elapsed < TOTAL_FLIGHT_MS) {
        const t = easeInCubic((elapsed - landStart) / PHASE_LANDING_MS);
        position = new THREE.Vector3().lerpVectors(fp.destApproach, fp.destLandingPos, t);
        const cruiseDir = new THREE.Vector3().subVectors(fp.cruiseEnd, fp.cruiseStart).normalize();
        const cruiseQuat = orientationFromDirection(cruiseDir, new THREE.Vector3(0, 1, 0));
        quat = cruiseQuat.clone().slerp(fp.landedQuat, t);
      } else {
        position = fp.destLandingPos.clone();
        quat = fp.landedQuat.clone();
        ship.status = "landed";
        ship.parkedMatrix = new THREE.Matrix4().compose(position, quat, new THREE.Vector3(1, 1, 1));
      }

      ship.currentPosition.copy(position);
      ship.currentQuat.copy(quat);
      ship.currentForward.set(0, 0, -1).applyQuaternion(quat).normalize();
      ship.hoverActive = elapsed < PHASE_LIFTOFF_MS * 0.5;
      if (elapsed > PHASE_LIFTOFF_MS) {
        ship.hoverSmokeDone = true;
      }
      matrices.push(new THREE.Matrix4().compose(position, quat, new THREE.Vector3(1, 1, 1)));
    }

    if (needsUpdate || entranceScale < 1) {
      if (entranceScale < 1) {
        const s = entranceScale;
        const scaleVec = new THREE.Vector3(s, s, s);
        for (let j = 0; j < matrices.length; j++) {
          const pos = new THREE.Vector3().setFromMatrixPosition(matrices[j]);
          const rot = new THREE.Quaternion().setFromRotationMatrix(matrices[j]);
          matrices[j] = new THREE.Matrix4().compose(pos, rot, scaleVec);
        }
      }
      setTransforms(matrices);
    }
  });

  const removeExhaust = (id: number) => {
    setExhaustEffects((prev) => prev.filter((e) => e.id !== id));
  };

  return (
    <>
      {padPositions.map((pad, i) => (
        <group key={`pad-${i}`} scale={entranceScale}>
          <mesh
            geometry={padGeometry}
            material={padMaterial}
            position={pad.position}
            quaternion={pad.quaternion}
            renderOrder={5}
          />
          {warningLineMaterials[i] && (
            <mesh
              geometry={warningLineGeometry}
              material={warningLineMaterials[i]}
              position={pad.position}
              quaternion={pad.quaternion}
              renderOrder={6}
            />
          )}
        </group>
      ))}
      {exhaustEffects.map((effect) => (
        <ExhaustEffect
          key={effect.id}
          position={effect.position}
          normal={effect.normal}
          particleCount={effect.particleCount}
          particleScale={effect.particleScale}
          duration={effect.duration}
          onComplete={() => removeExhaust(effect.id)}
        />
      ))}
      {shipsRef.current.map((ship, i) => {
        if (ship.status !== "flying") {
          return null;
        }
        const surfacePos = ship.flightPath?.startPos;
        const surfaceNormal = ship.flightPath?.startNormal;
        return (
          <group key={`effects-${i}`}>
            {!ship.hoverSmokeDone && surfacePos && surfaceNormal && (
              <HoverSmoke position={surfacePos} normal={surfaceNormal} active={ship.hoverActive} />
            )}
            <EngineTrail
              getShipPosition={() => ship.currentPosition.clone()}
              getShipForward={() => ship.currentForward.clone()}
              active={ship.status === "flying"}
              startDelay={PHASE_LIFTOFF_MS}
              trailDuration={1500}
            />
            <EngineGlow
              getShipPosition={() => ship.currentPosition.clone()}
              getShipQuat={() => ship.currentQuat.clone()}
              active={!ship.hoverActive && ship.status === "flying"}
            />
          </group>
        );
      })}
    </>
  );
}
