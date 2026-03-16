import { useRef, useEffect, useCallback } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";

interface Registration {
  transforms: THREE.Matrix4[];
  colors: THREE.Color[] | null;
}

interface Batch {
  geometry: THREE.BufferGeometry;
  material: THREE.Material;
  renderOrder: number;
  registrations: Map<string, Registration>;
  mesh: THREE.InstancedMesh | null;
  needsRebuild: boolean;
  needsUpdate: boolean;
}

const batches = new Map<string, Batch>();
let nextId = 0;

export interface PrimitiveHandle {
  setTransforms: (transforms: THREE.Matrix4[]) => void;
  setColors: (colors: THREE.Color[]) => void;
}

export function usePrimitiveInstances(
  batchKey: string,
  geometry: THREE.BufferGeometry | null,
  material: THREE.Material | null,
  renderOrder: number = 10,
): PrimitiveHandle {
  const regIdRef = useRef(`pm-${nextId++}`);
  const batchKeyRef = useRef(batchKey);

  useEffect(() => {
    if (!geometry || !material) {
      return;
    }

    let batch = batches.get(batchKey);
    if (!batch) {
      batch = {
        geometry,
        material,
        renderOrder,
        registrations: new Map(),
        mesh: null,
        needsRebuild: true,
        needsUpdate: false,
      };
      batches.set(batchKey, batch);
    }

    batch.registrations.set(regIdRef.current, { transforms: [], colors: null });
    batch.needsRebuild = true;
    batchKeyRef.current = batchKey;

    return () => {
      const b = batches.get(batchKeyRef.current);
      if (!b) {
        return;
      }
      b.registrations.delete(regIdRef.current);
      if (b.registrations.size === 0) {
        if (b.mesh) {
          b.mesh.removeFromParent();
          b.mesh.dispose();
        }
        batches.delete(batchKeyRef.current);
      } else {
        b.needsRebuild = true;
      }
    };
  }, [batchKey, geometry, material, renderOrder]);

  const setTransforms = useCallback((transforms: THREE.Matrix4[]) => {
    const batch = batches.get(batchKeyRef.current);
    if (!batch) {
      return;
    }
    const reg = batch.registrations.get(regIdRef.current);
    if (!reg) {
      return;
    }

    const countChanged = reg.transforms.length !== transforms.length;
    reg.transforms = transforms;

    if (countChanged) {
      batch.needsRebuild = true;
    } else {
      batch.needsUpdate = true;
    }
  }, []);

  const setColors = useCallback((colors: THREE.Color[]) => {
    const batch = batches.get(batchKeyRef.current);
    if (!batch) {
      return;
    }
    const reg = batch.registrations.get(regIdRef.current);
    if (!reg) {
      return;
    }

    reg.colors = colors;
    batch.needsUpdate = true;
  }, []);

  return { setTransforms, setColors };
}

function applyColors(mesh: THREE.InstancedMesh, registrations: Map<string, Registration>) {
  let hasColors = false;
  for (const [, reg] of registrations) {
    if (reg.colors && reg.colors.length > 0) {
      hasColors = true;
      break;
    }
  }
  if (!hasColors) {
    return;
  }

  let idx = 0;
  for (const [, reg] of registrations) {
    for (let i = 0; i < reg.transforms.length; i++) {
      if (reg.colors && i < reg.colors.length) {
        mesh.setColorAt(idx, reg.colors[i]);
      }
      idx++;
    }
  }
  if (mesh.instanceColor) {
    mesh.instanceColor.needsUpdate = true;
  }
}

export default function PrimitiveRenderer() {
  const containerRef = useRef<THREE.Group>(null);

  useFrame(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    for (const [, batch] of batches) {
      if (batch.needsRebuild) {
        if (batch.mesh) {
          container.remove(batch.mesh);
          batch.mesh.dispose();
        }

        let total = 0;
        for (const [, reg] of batch.registrations) {
          total += reg.transforms.length;
        }

        if (total === 0) {
          batch.mesh = null;
          batch.needsRebuild = false;
          continue;
        }

        const mesh = new THREE.InstancedMesh(batch.geometry, batch.material, total);
        mesh.renderOrder = batch.renderOrder;
        mesh.frustumCulled = false;

        let idx = 0;
        for (const [, reg] of batch.registrations) {
          for (const matrix of reg.transforms) {
            mesh.setMatrixAt(idx++, matrix);
          }
        }
        mesh.instanceMatrix.needsUpdate = true;

        applyColors(mesh, batch.registrations);

        container.add(mesh);
        batch.mesh = mesh;
        batch.needsRebuild = false;
        batch.needsUpdate = false;
      } else if (batch.needsUpdate && batch.mesh) {
        let idx = 0;
        for (const [, reg] of batch.registrations) {
          for (const matrix of reg.transforms) {
            batch.mesh.setMatrixAt(idx++, matrix);
          }
        }
        batch.mesh.instanceMatrix.needsUpdate = true;

        applyColors(batch.mesh, batch.registrations);

        batch.needsUpdate = false;
      }
    }
  });

  return <group ref={containerRef} />;
}
