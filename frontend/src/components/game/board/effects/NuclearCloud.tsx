import { useRef, useMemo, useLayoutEffect, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils.js";
import { useNuclearEffects, type NuclearEffect } from "../../../../contexts/NuclearEffectsContext";
import { audioService } from "../../../../services/audioService";
import {
  nuclearPhase,
  nuclearElapsed,
  NUCLEAR_FLIGHT_SECONDS,
  NUCLEAR_WAVE_SECONDS,
  NUCLEAR_WAVE_DUST_SECONDS,
  createNuclearWaveGeometry,
} from "../nuclearGeometry";
import { createNuclearWaveMaterial } from "../shaders";

interface NuclearCloudProps {
  startTime?: number;
  seed: number;
  sphereCenter: THREE.Vector3;
  onImpact: () => void;
  onComplete: () => void;
}

export default function NuclearCloud({
  seed,
  startTime,
  sphereCenter,
  onImpact,
  onComplete,
}: NuclearCloudProps) {
  const effects = useNuclearEffects();
  const frameloop = useThree((state) => state.frameloop);
  const group = useRef<THREE.Group>(null);
  const missile = useRef<THREE.Mesh>(null);
  const wave = useRef<THREE.Mesh>(null);
  const start = useRef<number | null>(null);
  const impacted = useRef(false);
  const completed = useRef(false);
  const audio = useRef<HTMLAudioElement | null>(null);
  const effect = useMemo<NuclearEffect>(
    () => ({ object: null as unknown as THREE.Group, seed, age: -NUCLEAR_FLIGHT_SECONDS }),
    [seed],
  );
  const scratch = useMemo(
    () => ({
      start: new THREE.Vector3(-0.48, 0.18, 0.85),
      end: new THREE.Vector3(0, 0, -0.008),
      direction: new THREE.Vector3(),
      axis: new THREE.Vector3(0, 1, 0),
    }),
    [],
  );
  const missileGeometry = useMemo(() => {
    const cylinder = new THREE.CylinderGeometry(0.0035, 0.004, 0.043, 8);
    const nose = new THREE.ConeGeometry(0.0035, 0.012, 8);
    nose.translate(0, 0.0275, 0);
    const geometry = mergeGeometries([cylinder, nose]);
    cylinder.dispose();
    nose.dispose();
    return geometry;
  }, []);
  const waveGeometry = useMemo(createNuclearWaveGeometry, []);
  const waveMaterial = useMemo(() => createNuclearWaveMaterial(sphereCenter), [sphereCenter]);
  useEffect(
    () => () => {
      missileGeometry.dispose();
      waveGeometry.dispose();
      waveMaterial.dispose();
    },
    [missileGeometry, waveGeometry, waveMaterial],
  );
  useLayoutEffect(() => {
    effect.object = group.current!;
    effects.add(effect);
    return () => {
      effects.delete(effect);
    };
  }, [effects, effect]);
  useEffect(() => () => audioService.stopNuclearBlastSound(audio.current), []);
  useEffect(() => {
    const syncPlayback = () => {
      const playback = audio.current;
      if (!playback || playback.ended) {
        return;
      }
      if (frameloop === "never" || document.hidden) {
        playback.pause();
      } else {
        void playback.play().catch(() => {});
      }
    };
    syncPlayback();
    document.addEventListener("visibilitychange", syncPlayback);
    return () => document.removeEventListener("visibilitychange", syncPlayback);
  }, [frameloop]);
  useFrame((state) => {
    start.current ??= state.clock.elapsedTime;
    const elapsed = nuclearElapsed(state.clock.elapsedTime - (startTime ?? start.current));
    const phase = nuclearPhase(elapsed);
    effect.age = elapsed - NUCLEAR_FLIGHT_SECONDS;
    if (missile.current) {
      missile.current.visible = phase === "flight";
      const t = Math.min(elapsed / NUCLEAR_FLIGHT_SECONDS, 1);
      missile.current.position.lerpVectors(scratch.start, scratch.end, Math.pow(t, 1.5));
      scratch.direction.copy(scratch.end).sub(scratch.start).normalize();
      missile.current.quaternion.setFromUnitVectors(scratch.axis, scratch.direction);
    }
    if (wave.current) {
      wave.current.visible =
        effect.age >= 0 && effect.age < NUCLEAR_WAVE_SECONDS + NUCLEAR_WAVE_DUST_SECONDS;
      waveMaterial.uniforms.uAge.value = effect.age;
    }
    if (phase !== "flight" && !impacted.current) {
      impacted.current = true;
      onImpact();
      audio.current = audioService.playNuclearBlastSound();
    }
    if (phase === "complete" && !completed.current) {
      completed.current = true;
      effects.delete(effect);
      onComplete();
    }
  });
  return (
    <group ref={group} userData={{ tileHighlights: false }}>
      <mesh ref={missile} geometry={missileGeometry} raycast={() => {}}>
        <meshStandardMaterial color="#59633d" roughness={0.85} />
      </mesh>
      <mesh
        ref={wave}
        geometry={waveGeometry}
        material={waveMaterial}
        visible={false}
        renderOrder={90}
        raycast={() => {}}
        frustumCulled={false}
      ></mesh>
    </group>
  );
}
