import * as THREE from "three";
import type { HexCoordinate } from "../../../utils/hex-grid-2d";
import { BOARD_SCALE } from "./landscapeGeometry";

export interface HighlightTile {
  coordinate: HexCoordinate;
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
}
export interface HighlightState {
  hovered: boolean;
  inspected: boolean;
  available: boolean;
  selected: boolean;
  vpIntensity: number;
  vpColor: readonly number[];
}
export interface HighlightRecord {
  tile: HighlightTile;
  state: HighlightState;
  hover: number;
  vp: number;
  visibility: number;
  slot: number;
}
const UP = new THREE.Vector3(0, 0, 1);
const TEXELS = 5;
const keyOf = (tile: HighlightTile) =>
  `${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`;

export class TileHighlightBoard {
  readonly records = new Map<string, HighlightRecord>();
  readonly texture = new THREE.DataTexture(
    new Float32Array(4),
    1,
    1,
    THREE.RGBAFormat,
    THREE.FloatType,
  );
  readonly uniforms = {
    uTileHighlightData: { value: this.texture },
    uTileHighlightGrid: { value: new THREE.Vector4(0, 0, 1, 1) },
    uTileHighlightProjection: { value: new THREE.Vector2() },
    uTileHighlightBodyInverse: { value: new THREE.Matrix4() },
    uTileHighlightCameraWorld: { value: new THREE.Matrix4() },
    uTileHighlightProjectionInverse: { value: new THREE.Matrix4() },
    uTileHighlightTime: { value: 0 },
    uTileHighlightActive: { value: 0 },
  };
  root: THREE.Object3D | null = null;
  private layoutDirty = true;
  private readonly basis = new THREE.Quaternion();
  private readonly axis = new THREE.Vector3();
  constructor(
    readonly radius: number,
    readonly ySign: number,
    readonly offset: { q: number; r: number },
  ) {
    this.texture.minFilter = this.texture.magFilter = THREE.NearestFilter;
    this.texture.generateMipmaps = false;
    this.uniforms.uTileHighlightProjection.value.set(radius, ySign);
  }
  register(tile: HighlightTile, state: HighlightState) {
    const record: HighlightRecord = { tile, state, hover: 0, vp: 0, visibility: 1, slot: 0 };
    this.records.set(keyOf(tile), record);
    this.layoutDirty = true;
    return record;
  }
  remove(record: HighlightRecord) {
    if (this.records.get(keyOf(record.tile)) === record) {
      this.records.delete(keyOf(record.tile));
      this.layoutDirty = true;
    }
  }
  updateTile(record: HighlightRecord, tile: HighlightTile) {
    if (
      !record.tile.spherePosition.equals(tile.spherePosition) ||
      !record.tile.normal.equals(tile.normal)
    ) {
      this.layoutDirty = true;
    }
    record.tile = tile;
  }
  private layout() {
    let minQ = Infinity,
      minR = Infinity,
      maxQ = -Infinity,
      maxR = -Infinity;
    for (const { tile } of this.records.values()) {
      minQ = Math.min(minQ, tile.coordinate.q - this.offset.q);
      maxQ = Math.max(maxQ, tile.coordinate.q - this.offset.q);
      minR = Math.min(minR, tile.coordinate.r - this.offset.r);
      maxR = Math.max(maxR, tile.coordinate.r - this.offset.r);
    }
    if (!this.records.size) {
      minQ = minR = maxQ = maxR = 0;
    }
    const width = maxQ - minQ + 1,
      height = maxR - minR + 1;
    const data = new Float32Array(width * height * TEXELS * 4);
    this.uniforms.uTileHighlightGrid.value.set(minQ, minR, width, height);
    for (const record of this.records.values()) {
      const { tile } = record;
      const x = tile.coordinate.q - this.offset.q - minQ;
      const y = tile.coordinate.r - this.offset.r - minR;
      const slot = (y * width * TEXELS + x * TEXELS) * 4;
      record.slot = slot;
      data[slot + 7] = 1;
      tile.normal.toArray(data, slot + 8);
      data[slot + 11] = tile.spherePosition.length() + 0.01;
      this.basis.setFromUnitVectors(UP, tile.normal);
      this.axis
        .set(1, 0, 0)
        .applyQuaternion(this.basis)
        .toArray(data, slot + 12);
      this.axis
        .set(0, 1, 0)
        .applyQuaternion(this.basis)
        .toArray(data, slot + 16);
    }
    this.texture.dispose();
    this.texture.image = { data, width: width * TEXELS, height };
    this.layoutDirty = false;
  }
  tick(delta: number, time: number, camera: THREE.Camera, hoverAllowed = true) {
    let dirty = this.layoutDirty;
    if (dirty) {
      this.layout();
    }
    let active = false;
    const data = this.texture.image.data as Float32Array;
    for (const record of this.records.values()) {
      if (!hoverAllowed) {
        record.state.hovered = false;
      }
      const hover = record.state.hovered || record.state.inspected ? 0.3 : 0;
      record.hover = fade(record.hover, hover, delta, 9.75);
      record.vp = fade(
        record.vp,
        record.state.vpIntensity,
        delta,
        record.state.vpIntensity > record.vp ? 5 : 2.45,
      );
      const slot = record.slot;
      dirty = write(data, slot, record.hover) || dirty;
      const placement =
        Number(record.state.available) + Number(record.state.available && record.state.selected);
      dirty = write(data, slot + 1, placement) || dirty;
      dirty = write(data, slot + 2, record.vp) || dirty;
      dirty = write(data, slot + 3, record.visibility) || dirty;
      for (let i = 0; i < 3; i++) {
        dirty = write(data, slot + 4 + i, record.state.vpColor[i]) || dirty;
      }
      active ||=
        record.visibility > 0.002 && (record.hover > 0 || record.state.available || record.vp > 0);
    }
    if (dirty) {
      this.texture.needsUpdate = true;
    }
    this.uniforms.uTileHighlightActive.value = Number(active && this.root !== null);
    this.uniforms.uTileHighlightTime.value = time;
    this.uniforms.uTileHighlightCameraWorld.value = camera.matrixWorld;
    this.uniforms.uTileHighlightProjectionInverse.value = camera.projectionMatrixInverse;
    if (this.root) {
      this.uniforms.uTileHighlightBodyInverse.value.copy(this.root.matrixWorld).invert();
    }
  }
  dispose() {
    this.texture.dispose();
    this.records.clear();
    this.root = null;
  }
}
function write(data: Float32Array, index: number, value: number) {
  const next = Math.fround(value);
  if (data[index] === next) {
    return false;
  }
  data[index] = next;
  return true;
}
function fade(current: number, target: number, delta: number, rate: number) {
  const next = THREE.MathUtils.lerp(current, target, 1 - Math.exp(-rate * Math.min(delta, 0.1)));
  return Math.abs(next - target) < 0.002 ? target : next;
}

export function highlightCoordinate(direction: THREE.Vector3, radius: number, ySign: number) {
  const d = Math.hypot(direction.x, direction.y);
  const scale = d > 1e-10 ? (Math.acos(THREE.MathUtils.clamp(direction.z, -1, 1)) * radius) / d : 0;
  const r = (direction.y * scale * ySign) / (0.45 * BOARD_SCALE);
  const q = (direction.x * scale) / (0.3 * Math.sqrt(3) * BOARD_SCALE) - r / 2;
  let rq = Math.round(q),
    rr = Math.round(r);
  const s = -q - r,
    rs = Math.round(s);
  const dq = Math.abs(rq - q),
    dr = Math.abs(rr - r),
    ds = Math.abs(rs - s);
  if (dq > dr && dq > ds) {
    rq = -rr - rs;
  } else if (dr > ds) {
    rr = -rq - rs;
  }
  return { q: rq + 0, r: rr + 0 };
}
