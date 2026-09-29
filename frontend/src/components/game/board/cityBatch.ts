import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import type { CityPlot } from "./cityLayout";
import { createGeometry } from "./cityGeometry";

type Entry = {
  plot: CityPlot;
  birth: number;
  parts: { key: string; geometry: number }[];
};
/** Keep unchanged cities in GPU buffers; editing one tile must not rebuild the board. */
export class CityBatchStore {
  readonly group = new THREE.Group();
  readonly meshes = new Map<string, THREE.BatchedMesh>();
  private entries = new Map<string, Entry>();
  builds = 0;
  constructor(private materials: Map<string, THREE.MeshStandardMaterial>) {}
  update(plots: CityPlot[], births: Map<string, number>) {
    const keys = new Set(plots.map((p) => HexGrid2D.coordinateToKey(p.coordinate)));
    for (const [key, entry] of this.entries) {
      if (!keys.has(key)) {
        this.remove(entry);
        this.entries.delete(key);
      }
    }
    for (const plot of plots) {
      const key = HexGrid2D.coordinateToKey(plot.coordinate),
        birth = births.get(key) ?? -1000;
      const previous = this.entries.get(key);
      if (previous?.plot === plot && previous.birth === birth) {
        continue;
      }
      if (previous) {
        this.remove(previous);
      }
      const entry: Entry = { plot, birth, parts: [] };
      const matrix = new THREE.Matrix4().compose(
        plot.worldPosition,
        new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 0, 1), plot.normal),
        new THREE.Vector3(1, 1, 1),
      );
      const geometries = createGeometry(plot.layout, plot);
      this.builds++;
      for (const [index, geometry] of geometries.entries()) {
        if (!geometry) {
          continue;
        }
        const mode = index < 3 ? plot.layout.style.lighting : "normal",
          materialKey = `${mode}:${index}`;
        const count = geometry.getAttribute("position").count;
        geometry.setAttribute(
          "cityBirth",
          new THREE.Float32BufferAttribute(new Float32Array(count).fill(birth), 1),
        );
        let mesh = this.meshes.get(materialKey);
        if (!mesh) {
          mesh = new THREE.BatchedMesh(
            128,
            Math.max(1024, count * 2),
            0,
            this.materials.get(materialKey)!,
          );
          mesh.renderOrder = 15;
          mesh.frustumCulled = false;
          mesh.raycast = () => {};
          this.meshes.set(materialKey, mesh);
          this.group.add(mesh);
        }
        if (mesh.unusedVertexCount < count) {
          mesh.optimize();
          if (mesh.unusedVertexCount < count) {
            const capacity = mesh.geometry.getAttribute("position").count;
            mesh.setGeometrySize(Math.max(capacity * 2, capacity + count), 0);
          }
        }
        const geometryId = mesh.addGeometry(geometry);
        const instance = mesh.addInstance(geometryId);
        mesh.setMatrixAt(instance, matrix);
        entry.parts.push({ key: materialKey, geometry: geometryId });
        geometry.dispose();
      }
      this.entries.set(key, entry);
    }
  }
  private remove(entry: Entry) {
    for (const part of entry.parts) {
      this.meshes.get(part.key)!.deleteGeometry(part.geometry);
    }
  }
  dispose() {
    for (const mesh of this.meshes.values()) {
      mesh.dispose();
    }
    this.group.clear();
    this.meshes.clear();
    this.entries.clear();
  }
}

export function mountCityWarmup(group: THREE.Group, materials: THREE.MeshStandardMaterial[]) {
  const geometry = new THREE.BoxGeometry(1, 1, 1);
  geometry.setAttribute(
    "cityBirth",
    new THREE.Float32BufferAttribute(new Float32Array(24).fill(-1000), 1),
  );
  geometry.setAttribute("color", new THREE.Float32BufferAttribute(new Float32Array(72).fill(1), 3));
  const meshes = materials.map((material) => {
    const mesh = new THREE.BatchedMesh(1, 24, 36, material);
    mesh.addInstance(mesh.addGeometry(geometry));
    mesh.frustumCulled = false;
    mesh.perObjectFrustumCulled = false;
    group.add(mesh);
    return mesh;
  });
  geometry.dispose();
  return () => {
    for (const mesh of meshes) {
      group.remove(mesh);
      mesh.dispose();
    }
  };
}
