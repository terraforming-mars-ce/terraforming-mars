import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import { cityWarmupLayouts, type CityPlot } from "./cityLayout";
import type { NuclearTransitions } from "./nuclearTransitions";
import { NUCLEAR_FLIGHT_SECONDS, NUCLEAR_PLAYBACK_RATE } from "./nuclearGeometry";
import { createGeometry } from "./cityGeometry";
import { projectBoardPoint } from "./landscapeGeometry";
import { coldStartTrace } from "../../../services/performanceStore";

const NOOP = () => {};
const PREPARATION_FRAMES = 3;

type Entry = {
  plot: CityPlot;
  birth: number;
  collapseStart?: number;
  parts: { key: string; geometry: number }[];
};
/** Keep unchanged cities in GPU buffers; editing one tile must not rebuild the board. */
export class CityBatchStore {
  readonly group = Object.assign(new THREE.Group(), { userData: { perfGroup: "cities" } });
  readonly meshes = new Map<string, THREE.BatchedMesh>();
  private entries = new Map<string, Entry>();
  private preparation: ReturnType<typeof cityWarmupLayouts> | null = null;
  private capacities = new Map<number, number>();
  private warmupParts: { mesh: THREE.BatchedMesh; geometry: number; frames: number }[] = [];
  ready = false;
  builds = 0;
  constructor(private materials: Map<string, THREE.MeshStandardMaterial>) {}
  beginPreparation() {
    this.dispose();
    this.preparation = cityWarmupLayouts();
    this.group.visible = true;
  }
  prepareNext() {
    if (this.ready) {
      return true;
    }
    if (this.preparation) {
      const end = coldStartTrace.span("city:prepare-fixture");
      const next = this.preparation.next();
      if (!next.done) {
        const worldPosition = projectBoardPoint(0, 0);
        const geometries = createGeometry(next.value, {
          layout: next.value,
          coordinate: { q: 0, r: 0, s: 0 },
          worldPosition,
          normal: worldPosition.clone().normalize(),
        });
        for (const [index, geometry] of geometries.entries()) {
          if (geometry) {
            this.capacities.set(
              index,
              Math.max(this.capacities.get(index) ?? 0, geometry.getAttribute("position").count),
            );
            geometry.dispose();
          }
        }
        end();
        return false;
      }
      this.preparation = null;
      const geometry = createCityWarmupGeometry();
      const matrix = new THREE.Matrix4().makeScale(0.00001, 0.00001, 0.00001);
      for (const key of this.materials.keys()) {
        const index = Number(key.split(":")[1]);
        const capacity = Math.ceil((2 * (this.capacities.get(index) ?? 36)) / 1024) * 1024;
        const mesh = this.createMesh(key, capacity);
        mesh.perObjectFrustumCulled = false;
        const geometryId = mesh.addGeometry(geometry);
        mesh.setMatrixAt(mesh.addInstance(geometryId), matrix);
        const part = { mesh, geometry: geometryId, frames: 0 };
        let lastFrame = -1;
        mesh.onAfterRender = (renderer) => {
          if (renderer.info.render.frame !== lastFrame) {
            lastFrame = renderer.info.render.frame;
            part.frames++;
          }
        };
        this.warmupParts.push(part);
      }
      geometry.dispose();
      end();
      return false;
    }
    if (
      this.warmupParts.length === 0 ||
      this.warmupParts.some((p) => p.frames < PREPARATION_FRAMES)
    ) {
      return false;
    }
    for (const { mesh, geometry } of this.warmupParts) {
      mesh.onAfterRender = NOOP;
      mesh.deleteGeometry(geometry);
      mesh.optimize();
      mesh.perObjectFrustumCulled = true;
    }
    this.warmupParts.length = 0;
    this.capacities.clear();
    this.ready = true;
    coldStartTrace.mark("city:prepared", { batches: this.meshes.size });
    return true;
  }
  private createMesh(key: string, capacity: number) {
    const mesh = new THREE.BatchedMesh(128, capacity, 0, this.materials.get(key)!);
    mesh.renderOrder = 15;
    mesh.frustumCulled = false;
    mesh.raycast = NOOP;
    this.meshes.set(key, mesh);
    this.group.add(mesh);
    return mesh;
  }
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
      const endGeometry = coldStartTrace.span("city:geometry", { key });
      const geometries = createGeometry(plot.layout, plot);
      endGeometry();
      const endBatch = coldStartTrace.span("city:batch-update", { key });
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
        geometry.setAttribute(
          "cityCollapseStart",
          new THREE.Float32BufferAttribute(new Float32Array(count).fill(1e9), 1),
        );
        let mesh = this.meshes.get(materialKey);
        if (!mesh) {
          mesh = this.createMesh(materialKey, Math.max(1024, count * 2));
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
      endBatch();
    }
  }
  setCollapses(transitions: NuclearTransitions) {
    for (const [key, entry] of this.entries) {
      const transition = transitions.get(key);
      const start = transition
        ? transition.start + NUCLEAR_FLIGHT_SECONDS / NUCLEAR_PLAYBACK_RATE
        : 1e9;
      if (entry.collapseStart === start) {
        continue;
      }
      entry.collapseStart = start;
      for (const part of entry.parts) {
        const mesh = this.meshes.get(part.key)!;
        const range = mesh.getGeometryRangeAt(part.geometry)!;
        const attribute = mesh.geometry.getAttribute("cityCollapseStart") as THREE.BufferAttribute;
        (attribute.array as Float32Array).fill(
          start,
          range.vertexStart,
          range.vertexStart + range.vertexCount,
        );
        attribute.addUpdateRange(range.vertexStart, range.vertexCount);
        attribute.needsUpdate = true;
      }
    }
  }
  private remove(entry: Entry) {
    for (const part of entry.parts) {
      this.meshes.get(part.key)!.deleteGeometry(part.geometry);
    }
  }
  dispose() {
    for (const mesh of this.meshes.values()) {
      mesh.onAfterRender = NOOP;
      mesh.dispose();
    }
    this.group.clear();
    this.meshes.clear();
    this.entries.clear();
    this.preparation = null;
    this.capacities.clear();
    this.warmupParts.length = 0;
    this.ready = false;
  }
}

export function createCityWarmupGeometry() {
  const box = new THREE.BoxGeometry(1, 1, 1);
  const geometry = box.toNonIndexed();
  box.dispose();
  geometry.clearGroups();
  const count = geometry.getAttribute("position").count;
  geometry.setAttribute(
    "cityFrostExposure",
    new THREE.Float32BufferAttribute(new Float32Array(count).fill(1), 1),
  );
  geometry.setAttribute(
    "cityBirth",
    new THREE.Float32BufferAttribute(new Float32Array(count).fill(-1000), 1),
  );
  geometry.setAttribute(
    "color",
    new THREE.Float32BufferAttribute(new Float32Array(count * 3).fill(1), 3),
  );
  geometry.setAttribute(
    "cityCollapseStart",
    new THREE.Float32BufferAttribute(new Float32Array(count).fill(1e9), 1),
  );
  geometry.setAttribute(
    "cityCollapsePivot",
    new THREE.Float32BufferAttribute(new Float32Array(count * 3), 3),
  );
  geometry.setAttribute(
    "cityCollapseMotion",
    new THREE.Float32BufferAttribute(new Float32Array(count * 4), 4),
  );
  return geometry;
}
