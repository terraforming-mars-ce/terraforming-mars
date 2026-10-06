import * as THREE from "three";
import { TileHighlightBoard } from "./tileHighlights";
import vertex from "./shaders/tile-highlight.vert.glsl?raw";
import fragment from "./shaders/tile-highlight.frag.glsl?raw";

interface Binding {
  board: TileHighlightBoard;
  compile: THREE.Material["onBeforeCompile"];
  key: THREE.Material["customProgramCacheKey"];
  restore: () => void;
}
const bindings = new WeakMap<THREE.Material, Binding>();

export function tileHighlightSource(material: THREE.Material) {
  const binding = bindings.get(material);
  return {
    compile: binding?.compile ?? material.onBeforeCompile,
    key: binding?.key ?? material.customProgramCacheKey,
  };
}
export function receivesTileHighlight(material: THREE.Material) {
  if (
    !material.colorWrite ||
    material.userData.tileHighlights === false ||
    material.blending === THREE.AdditiveBlending
  ) {
    return false;
  }
  return (
    material.userData.tileHighlights === true ||
    material instanceof THREE.MeshStandardMaterial ||
    material instanceof THREE.MeshLambertMaterial ||
    (material instanceof THREE.ShaderMaterial &&
      !(material instanceof THREE.RawShaderMaterial) &&
      material.depthWrite)
  );
}
export function addTileHighlight(material: THREE.Material, board: TileHighlightBoard) {
  const existing = bindings.get(material);
  if (existing) {
    return existing.restore;
  }
  const compile = material.onBeforeCompile,
    key = material.customProgramCacheKey;
  const baseKey = key.call(material);
  const [header, body] = vertex.split("//#pragma body\n");
  const wrapped: THREE.Material["onBeforeCompile"] = (shader, renderer) => {
    compile.call(material, shader, renderer);
    Object.assign(shader.uniforms, board.uniforms);
    shader.vertexShader =
      header +
      "\n" +
      shader.vertexShader.replace(/void\s+main\s*\(\s*\)/, "void tileHighlightSourceVertex()") +
      "\n" +
      body;
    const output = shader.fragmentShader.match(/out\s+vec4\s+(\w+)\s*;/)?.[1] ?? "gl_FragColor";
    shader.fragmentShader =
      fragment +
      "\n" +
      shader.fragmentShader.replace(/void\s+main\s*\(\s*\)/, "void tileHighlightSourceFragment()") +
      `\nvoid main() { tileHighlightSourceFragment(); ${output}.rgb = applyTileHighlight(${output}.rgb); }`;
  };
  const restore = () => {
    if (material.onBeforeCompile === wrapped) {
      material.onBeforeCompile = compile;
      material.customProgramCacheKey = key;
      material.needsUpdate = true;
    }
    material.removeEventListener("dispose", restore);
    bindings.delete(material);
  };
  bindings.set(material, { board, compile, key, restore });
  material.onBeforeCompile = wrapped;
  material.customProgramCacheKey = () => baseKey + vertex + fragment;
  material.addEventListener("dispose", restore);
  material.needsUpdate = true;
  return restore;
}

export class TileHighlightSystem {
  readonly boards = new Set<TileHighlightBoard>();
  private readonly roots = new Map<THREE.Object3D, TileHighlightBoard>();
  private readonly materials = new Map<
    THREE.Material,
    { board: TileHighlightBoard; restore: () => void }
  >();
  private readonly clones = new Map<TileHighlightBoard, Map<THREE.Material, THREE.Material>>();
  private readonly assignments = new Map<
    THREE.Mesh,
    {
      original: THREE.Material | THREE.Material[];
      assigned: THREE.Material | THREE.Material[];
      board: TileHighlightBoard;
    }
  >();
  private readonly warmup = new TileHighlightBoard(2.02, -1, { q: 0, r: 0 });
  tick(delta: number, time: number, camera: THREE.Camera, hoverAllowed: boolean) {
    this.roots.clear();
    for (const board of this.boards) {
      board.tick(delta, time, camera, hoverAllowed);
      if (board.root) {
        this.roots.set(board.root, board);
      }
    }
    this.warmup.tick(delta, time, camera, false);
  }
  private material(source: THREE.Material, board: TileHighlightBoard) {
    if (!receivesTileHighlight(source)) {
      return source;
    }
    let material = source;
    const binding = bindings.get(source);
    if (binding && binding.board !== board) {
      let cache = this.clones.get(board);
      if (!cache) {
        cache = new Map();
        this.clones.set(board, cache);
      }
      let clone = cache.get(source);
      if (!clone) {
        clone = source.clone();
        const original = tileHighlightSource(source);
        clone.onBeforeCompile = original.compile;
        clone.customProgramCacheKey = original.key;
        // The copied compiler already includes the source's atmosphere wrapper.
        clone.userData.planetHaze = false;
        cache.set(source, clone);
      }
      material = clone;
    }
    if (!this.materials.has(material)) {
      const unpatch = addTileHighlight(material, board);
      const release = () => {
        material.removeEventListener("dispose", release);
        this.materials.delete(material);
        unpatch();
      };
      material.addEventListener("dispose", release);
      this.materials.set(material, { board, restore: release });
    }
    return material;
  }
  readonly visit = (object: THREE.Object3D) => {
    if (!(object instanceof THREE.Mesh) || !object.material) {
      return;
    }
    let parent: THREE.Object3D | null = object;
    let board: TileHighlightBoard | undefined;
    while (parent) {
      if (parent.userData.tileHighlights === false) {
        return;
      }
      board = this.roots.get(parent);
      if (board) {
        break;
      }
      if (parent.userData.perfGroup === "gpu warmup") {
        board = this.warmup;
        break;
      }
      parent = parent.parent;
    }
    if (!board) {
      return;
    }
    const original = object.material;
    if (Array.isArray(original)) {
      let assigned: THREE.Material[] | undefined;
      for (let i = 0; i < original.length; i++) {
        const material = this.material(original[i], board);
        if (material !== original[i]) {
          assigned ??= original.slice();
          assigned[i] = material;
        }
      }
      if (assigned) {
        object.material = assigned;
        this.assignments.set(object, { original, assigned, board });
      }
    } else {
      const assigned = this.material(original, board);
      if (assigned !== original) {
        object.material = assigned;
        this.assignments.set(object, { original, assigned, board });
      }
    }
  };
  release(board: TileHighlightBoard) {
    this.boards.delete(board);
    for (const [mesh, assignment] of this.assignments) {
      if (assignment.board === board) {
        if (mesh.material === assignment.assigned) {
          mesh.material = assignment.original;
        }
        this.assignments.delete(mesh);
      }
    }
    for (const [material, binding] of this.materials) {
      if (binding.board === board) {
        binding.restore();
        // Flush cached programs/uniforms before a board is rebound after hot reload.
        material.dispose();
      }
    }
    for (const material of this.clones.get(board)?.values() ?? []) {
      material.dispose();
    }
    this.clones.delete(board);
    board.dispose();
  }
  dispose() {
    for (const board of this.boards) {
      this.release(board);
    }
    this.release(this.warmup);
    this.roots.clear();
  }
}
