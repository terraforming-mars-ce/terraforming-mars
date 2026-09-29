import { acceptableRoad } from "../../src/components/game/board/landscapeRoutes";
import * as THREE from "three";
import { CityBatchStore, mountCityWarmup } from "../../src/components/game/board/cityBatch";
import { createGeometry } from "../../src/components/game/board/cityGeometry";
import {
  boardCenter,
  createLandscapeGround,
} from "../../src/components/game/board/landscapeGeometry";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import {
  generateCityLayout,
  visibleCityRoads,
  createCityShowcase,
} from "../../src/components/game/board/cityLayout";
import {
  buildLandscape,
  connectOceanGround,
  type LandscapeTile,
  isReserved,
} from "../../src/components/game/board/landscapeNetwork";
import {
  nearestOnRoad,
  projectBoardPoint,
  hashSeed,
} from "../../src/components/game/board/landscapeGeometry";
import { HexGrid2D } from "../../src/utils/hex-grid-2d";
import type { CityStyleRequestDto } from "../../src/types/generated/api-types";
import type { LandscapeRequest, LandscapeResult } from "../../src/components/game/board/useLandscape";
const cards = JSON.parse(
  readFileSync(
    new URL("../../../backend/assets/terraforming_mars_cards.json", import.meta.url),
    "utf8",
  ),
) as { id: string; name: string; style?: { tile: CityStyleRequestDto } }[];
const requests = cards.filter((c) => c.style?.tile);
assert.equal(requests.length, 7);
for (const card of requests) {
  for (let seed = 1; seed <= 20; seed++) {
    const layout = generateCityLayout(seed, card.style!.tile, card.name);
    assert.deepEqual(
      layout,
      generateCityLayout(seed, structuredClone(card.style!.tile), card.name),
    );
    assert.ok(layout.buildings.length > 0, card.name);
    assert.ok(layout.roads.length > 0);
    if (card.id === "120") {
      assert.ok(layout.entrances.length >= 2 && layout.entrances.length <= 3);
    }
    for (const b of layout.buildings) {
      assert.ok(Number.isFinite(b.x + b.y + b.height + b.rotation));
      if (b.entranceRoad) {
        const dx = b.entranceRoad.x - b.x,
          dy = b.entranceRoad.y - b.y;
        assert.ok(
          (dx * Math.sin(b.rotation) - dy * Math.cos(b.rotation)) / Math.hypot(dx, dy) > 0.999,
        );
      }
    }
    for (const plant of layout.plants) {
      const road = nearestOnRoad(plant, layout.roads);
      assert.ok(road.distance > road.halfWidth);
    }
  }
}
for (const card of requests) {
  const layout = generateCityLayout(9183, card.style!.tile, card.name);
  const coordinate = { q: 4, r: 0, s: -4 };
  const center = boardCenter(coordinate),
    worldPosition = projectBoardPoint(center.x, center.y);
  const meshes = createGeometry(layout, {
    coordinate,
    worldPosition,
    normal: worldPosition.clone().normalize(),
    layout,
  });
  assert.ok(meshes.some(Boolean));
  for (const mesh of meshes) {
    if (mesh) {
      for (const attr of Object.values(mesh.attributes)) {
        assert.ok(Array.from(attr.array).every(Number.isFinite), `${card.name}: invalid mesh`);
      }
      mesh.dispose();
    }
  }
}
const coord = (q: number, r = 0) => ({ q, r, s: -q - r });
const tile = (q: number, kind: LandscapeTile["kind"], r = 0): LandscapeTile => ({
  coordinate: coord(q, r),
  kind,
  seed: hashSeed(`${q},${r}`),
  layout: kind === "city" ? generateCityLayout(hashSeed(`${q},${r}`)) : undefined,
});
const cityA = tile(0, "city"),
  greenA = tile(1, "greenery"),
  greenB = tile(2, "greenery"),
  cityB = tile(3, "city");
const before = JSON.stringify(cityA.layout);
const chain = buildLandscape([cityA, greenA, greenB, cityB]);
for (const r of chain.roads) {
  for (const p of r.points) {
    for (const forest of [greenA, greenB]) {
      const center = boardCenter(forest.coordinate),
        x = p.x - center.x,
        y = p.y - center.y;
      assert.ok(
        Math.max(
          Math.abs(x),
          Math.abs(0.5 * x + 0.866025404 * y),
          Math.abs(-0.5 * x + 0.866025404 * y),
        ) >= 0.146,
      );
    }
  }
}
assert.equal(JSON.stringify(cityA.layout), before);
assert.deepEqual(chain.connections, buildLandscape([cityB, greenB, cityA, greenA]).connections);
assert.equal(buildLandscape([greenA, greenB]).connections.length, 0);
const alone = buildLandscape([greenA]);
const original = new Map(chain.plants.filter((p) => p.tileKey === "1,0,-1").map((p) => [p.id, p]));
for (const plant of alone.plants) {
  if (!isReserved(chain, plant.x, plant.y, plant.kind === "tree" ? 0.009 : 0.003)) {
    assert.deepEqual(
      original.get(plant.id),
      plant,
      "Neighbor changed a plant outside new roadworks",
    );
  }
}
for (const c of chain.connections) {
  for (let i = 1; i < c.routes.length; i++) {
    assert.deepEqual(c.routes[i - 1].points.at(-1), c.routes[i].points[0]);
  }
}
const full: LandscapeTile[] = [];
for (let q = -4; q <= 4; q++) {
  for (let r = -4; r <= 4; r++) {
    if (Math.abs(q + r) <= 4) {
      full.push(tile(q, (q - r) % 3 === 0 ? "city" : "greenery", r));
    }
  }
}
const start = performance.now();
const plan = buildLandscape(full);
console.log(
  JSON.stringify({
    checkedStyles: requests.length,
    tiles: full.length,
    connections: plan.connections.length,
    plants: plan.plants.length,
    planningMs: Math.round(performance.now() - start),
  }),
);

for (const neighbor of HexGrid2D.getNeighbors(coord(1))) {
  for (const card of requests) {
    const adjacent: LandscapeTile = {
      coordinate: neighbor,
      kind: "city",
      seed: 913,
      layout: generateCityLayout(913, card.style!.tile),
    };
    const grown = buildLandscape([greenA, adjacent]);
    const plants = new Map(
      grown.plants.filter((p) => p.tileKey === "1,0,-1").map((p) => [p.id, p]),
    );
    for (const p of alone.plants) {
      if (!isReserved(grown, p.x, p.y, p.kind === "tree" ? 0.009 : 0.003)) {
        assert.deepEqual(
          plants.get(p.id),
          p,
          `Growth near ${card.name} changed vegetation outside the road`,
        );
      }
    }
  }
}
assert.deepEqual(
  buildLandscape([greenA]).plants,
  alone.plants,
  "Removal changed the surviving forest",
);

// Roads may cross empty land; blocked tiles either cause a smooth detour or no road.
const spaces = HexGrid2D.generateGrid().map((t) => ({ coordinate: t.coordinate, blocked: false }));
function facingCity(q: number, r = 0, direction = 0): LandscapeTile {
  const city = tile(q, "city", r),
    original = generateCityLayout(71, { plan: "radial", entranceMin: 2, entranceMax: 2 });
  const rotation = direction - original.entrances[0];
  const rotate = (p: { x: number; y: number }) => ({
    x: p.x * Math.cos(rotation) - p.y * Math.sin(rotation),
    y: p.x * Math.sin(rotation) + p.y * Math.cos(rotation),
  });
  city.layout = {
    ...original,
    entrances: original.entrances.map((a) => a + rotation),
    roads: original.roads.map((r) => ({ ...r, points: r.points.map(rotate) })),
  };
  return city;
}
const directA = facingCity(0),
  directB = facingCity(2);
const emptyCrossing = buildLandscape([directA, directB], spaces);
assert.equal(emptyCrossing.connections.length, 1);
const blocked = spaces.map((s) => ({
  ...s,
  blocked: s.coordinate.q === 1 && s.coordinate.r === 0,
}));
const diverted = buildLandscape([directA, directB], blocked);
// A large detour is intentionally rejected by the bend budget.
for (const connection of diverted.connections) {
  const points = connection.routes.flatMap((r) => r.points);
  const obstacle = boardCenter(coord(1));
  for (const p of points) {
    const x = p.x - obstacle.x,
      y = p.y - obstacle.y;
    assert.ok(
      Math.max(
        Math.abs(x),
        Math.abs(0.5 * x + 0.866025404 * y),
        Math.abs(-0.5 * x + 0.866025404 * y),
      ) >= 0.17,
      "Road enters an obstacle",
    );
  }
}
const sealed = spaces.map((s) => ({ ...s, blocked: s.coordinate.q === 1 }));
assert.equal(buildLandscape([directA, directB], sealed).connections.length, 0);
for (const network of [chain, emptyCrossing, diverted]) {
  for (const connection of network.connections) {
    const points = connection.routes.flatMap((r, i) => (i ? r.points.slice(1) : r.points));
    for (let i = 1; i < points.length - 1; i++) {
      const a = points[i - 1],
        b = points[i],
        c = points[i + 1];
      const ax = b.x - a.x,
        ay = b.y - a.y,
        bx = c.x - b.x,
        by = c.y - b.y;
      const lengths = Math.hypot(ax, ay) * Math.hypot(bx, by);
      if (lengths > 1e-12) {
        assert.ok(
          (ax * bx + ay * by) / lengths > Math.cos(Math.PI / 6),
          "Spline contains a sharp corner",
        );
      }
    }
  }
}
for (let seed = 0; seed < 50; seed++) {
  const layout = generateCityLayout(seed);
  assert.equal(layout.style.plan, "grid");
  assert.ok(layout.buildings.length >= 30 && layout.buildings.length <= 55);
  assert.ok(new Set(layout.buildings.map((b) => b.tint)).size >= 5);
  assert.ok(
    layout.buildings.some((b) => Math.max(b.width, b.depth) / Math.min(b.width, b.depth) > 1.5),
  );
  for (const b of layout.buildings) {
    for (const sx of [-1, 1]) {
      for (const sy of [-1, 1]) {
        const x =
          b.x +
          ((sx * b.width) / 2) * Math.cos(b.rotation) -
          ((sy * b.depth) / 2) * Math.sin(b.rotation);
        const y =
          b.y +
          ((sx * b.width) / 2) * Math.sin(b.rotation) +
          ((sy * b.depth) / 2) * Math.cos(b.rotation);
        assert.ok(
          Math.max(
            Math.abs(x),
            Math.abs(x * 0.5 + y * 0.866025404),
            Math.abs(-x * 0.5 + y * 0.866025404),
          ) < 0.142,
          "City footprint crosses its hex border",
        );
      }
    }
  }
  assert.equal(layout.buildings.filter((b) => b.height > 0.045).length, 4);
  assert.ok(layout.buildings.filter((b) => b.habitat).length >= 2);
  assert.ok(visibleCityRoads(layout, []).every((r) => r.exit === undefined));
}
const isolated = buildLandscape([directA]);
assert.deepEqual(isolated.activeExits.get("0,0,0"), []);
for (const [key, active] of emptyCrossing.activeExits) {
  assert.equal(active.length, 1, key);
}
const badExit = buildLandscape(
  [facingCity(0, 0, Math.PI / 2), facingCity(2, 0, Math.PI / 2)],
  spaces,
);
assert.equal(badExit.connections.length, 0, "A hairpin must not be used to force a connection");
const junction = buildLandscape(
  [facingCity(0), facingCity(2), facingCity(1, -1, -Math.PI / 2)],
  spaces,
);
assert.equal(junction.connections.length, 2);
assert.ok(
  junction.connections.some((c) =>
    [c.routes[0].points[0], c.routes.at(-1)!.points.at(-1)!].some((p) =>
      junction.connections.some(
        (other) => other !== c && nearestOnRoad(p, other.routes).distance < 1e-7,
      ),
    ),
  ),
  "Nearby links should share a junction",
);
assert.equal(
  junction.activeExits.get("2,0,-2")!.length,
  1,
  "The trunk should use only one city exit",
);
const forestAvoidance = buildLandscape([directA, tile(1, "greenery"), directB], spaces);
for (const r of forestAvoidance.roads) {
  for (const p of r.points) {
    const c = boardCenter(coord(1));
    assert.ok(Math.hypot(p.x - c.x, p.y - c.y) > 0.146);
  }
}
for (const network of [emptyCrossing, diverted, junction, plan]) {
  for (const c of network.connections) {
    assert.ok(acceptableRoad(c.routes.flatMap((r, i) => (i ? r.points.slice(1) : r.points))));
  }
}
const showcase = createCityShowcase().map((layout, i) => {
  const coordinate = coord(i % 3, Math.floor(i / 3));
  const center = boardCenter(coordinate),
    worldPosition = projectBoardPoint(center.x, center.y);
  return { layout, coordinate, worldPosition, normal: worldPosition.clone().normalize() };
});
const materials = new Map<string, THREE.MeshStandardMaterial>();
for (const mode of ["normal", "bright"]) {
  for (let i = 0; i < 19; i++) {
    materials.set(`${mode}:${i}`, new THREE.MeshStandardMaterial());
  }
}
const store = new CityBatchStore(materials);
store.update(showcase, new Map(), new Map());
assert.equal(store.builds, showcase.length);
assert.ok(store.meshes.size <= 22, "Cities must share material batches");
const unchangedStart = performance.now();
store.update([...showcase], new Map(), new Map());
const unchangedMs = performance.now() - unchangedStart;
assert.equal(store.builds, showcase.length, "Unchanged cities must retain GPU geometry");
store.update(
  showcase,
  new Map([[HexGrid2D.coordinateToKey(showcase[0].coordinate), [0]]]),
  new Map(),
);
assert.equal(store.builds, showcase.length + 1, "Changing an exit should rebuild only its city");
let triangles = 0;
for (const plot of showcase) {
  for (const mesh of createGeometry(plot.layout, plot, [])) {
    if (!mesh) {
      continue;
    }
    triangles += mesh.getAttribute("position").count / 3;
    mesh.dispose();
  }
}
const batchCount = store.meshes.size;
store.update([], new Map(), new Map());
store.dispose();
// Exercise the same render preparation that threw after Strict Mode effect replay.
const warmupGroup = new THREE.Group();
const warmupMaterial = new THREE.MeshStandardMaterial();
const camera = new THREE.PerspectiveCamera();
const scene = new THREE.Scene();
function prepareWarmupRender() {
  for (const object of warmupGroup.children) {
    const mesh = object as THREE.BatchedMesh;
    assert.doesNotThrow(() =>
      mesh.onBeforeRender(
        {} as THREE.WebGLRenderer,
        scene,
        camera,
        mesh.geometry,
        warmupMaterial,
        {} as THREE.Group,
      ),
    );
  }
}
const firstCleanup = mountCityWarmup(warmupGroup, [warmupMaterial]);
const firstMesh = warmupGroup.children[0];
prepareWarmupRender();
firstCleanup();
assert.equal(warmupGroup.children.length, 0, "Disposed warmup meshes must leave the scene");
const secondCleanup = mountCityWarmup(warmupGroup, [warmupMaterial]);
assert.notEqual(
  warmupGroup.children[0],
  firstMesh,
  "Effect replay must create fresh batching textures",
);
prepareWarmupRender();
secondCleanup();
warmupMaterial.dispose();
store.update(showcase, new Map(), new Map());
assert.ok(store.group.children.length > 0, "City batches must survive effect replay too");
store.dispose();
materials.forEach((material) => material.dispose());
assert.ok(
  showcase
    .find((p) => p.layout.style.landmark === "hall")!
    .layout.buildings.find((b) => b.landmark)!.height >= 0.077,
);
console.log(
  JSON.stringify({
    cityCount: showcase.length,
    cityBatches: batchCount,
    cityTriangles: triangles,
    unchangedUpdateMs: unchangedMs,
  }),
);
console.log("Landscape invariants passed.");

for (const kind of ["city", "greenery"] as const) {
  const landTile = tile(0, kind);
  const landPlan = buildLandscape([landTile]);
  for (const coordinate of HexGrid2D.getNeighbors(coord(0))) {
    const coast = connectOceanGround(landPlan, [{ coordinate }]);
    assert.equal(coast.tiles.length, 7, "Moisture must cover every side of an ocean");
    assert.deepEqual(coast.links, landPlan.greenLinks, "Oceans must not create city-to-water grass strips");
    assert.deepEqual(coast.water, [boardCenter(coordinate)]);
    assert.deepEqual(landPlan, buildLandscape([landTile]), "Coasts must not relocate vegetation or roads");
  }
}
const isolatedCoast = connectOceanGround(buildLandscape([]), [{ coordinate: coord(0) }]);
assert.equal(isolatedCoast.tiles.length, 7, "Water must support vegetation without neighboring greenery");
const joinedCoast = connectOceanGround(buildLandscape([]), [{ coordinate: coord(0) }, { coordinate: coord(1) }]);
assert.equal(joinedCoast.waterLinks.length, 1, "Moisture must follow the shoreline between merged oceans");
const coastGeometry = createLandscapeGround(buildLandscape([]), joinedCoast);
const coastalPositions = coastGeometry.getAttribute("boardPosition");
const shoreDistances = coastGeometry.getAttribute("shoreDistance");
for (let i = 0; i < coastalPositions.count; i += 7) {
  const p = { x: coastalPositions.getX(i), y: coastalPositions.getY(i) };
  const expected = nearestOnRoad(p, [{ points: [boardCenter(coord(0)), boardCenter(coord(1))], width: 0, kind: "path" }]).distance - 0.17 * Math.PI * 0.2;
  if (expected < 0.15) {
    assert.ok(Math.abs(shoreDistances.getX(i) - expected) < 1e-6);
  }
}
coastGeometry.dispose();
for (const network of [junction, emptyCrossing, plan]) {
  const graph = new Map<string, Set<string>>();
  for (const edge of network.connections) {
    const pending = [edge.a],
      visited = new Set<string>();
    while (pending.length) {
      const key = pending.pop()!;
      assert.notEqual(key, edge.b, "A new road must not close a redundant loop");
      if (visited.has(key)) {
        continue;
      }
      visited.add(key);
      pending.push(...(graph.get(key) ?? []));
    }
    if (!graph.has(edge.a)) {
      graph.set(edge.a, new Set());
    }
    if (!graph.has(edge.b)) {
      graph.set(edge.b, new Set());
    }
    graph.get(edge.a)!.add(edge.b);
    graph.get(edge.b)!.add(edge.a);
  }
}
console.log("Coastal continuity and road topology passed.");

const groundPlan = connectOceanGround(plan, [{ coordinate: coord(5) }]);
const groundStart = performance.now();
const groundMesh = createLandscapeGround(plan, groundPlan);
const groundBuildMs = performance.now() - groundStart;
const oldGroundVertices = groundPlan.tiles.length * 6 * 12 * 12 * 3;
const vertices = groundMesh.getAttribute("position").count;
assert.ok(
  vertices < oldGroundVertices * 0.22,
  "Ground vertices should be shared within each patch",
);
assert.equal(
  groundMesh.index!.count,
  oldGroundVertices,
  "Baking must retain the same terrain triangles",
);
for (const attr of Object.values(groundMesh.attributes)) {
  assert.ok(Array.from(attr.array).every(Number.isFinite));
}
const fields = groundMesh.getAttribute("landscapeField");
const positions2d = groundMesh.getAttribute("boardPosition");
for (let i = 0; i < fields.count; i += 7) {
  const p = { x: positions2d.getX(i), y: positions2d.getY(i) };
  let expected = -1;
  for (const circle of plan.circles) {
    expected = Math.max(expected, circle.radius - Math.hypot(p.x - circle.x, p.y - circle.y));
  }
  for (const link of groundPlan.links) {
    expected = Math.max(expected, link.width / 2 - nearestOnRoad(p, [link]).distance);
  }
  assert.ok(
    expected < -0.06 || Math.abs(fields.getX(i) - expected) < 1e-6,
    "Baked coverage differs from the terrain shader's shape union",
  );
}
console.log(
  JSON.stringify({
    groundTiles: groundPlan.tiles.length,
    oldGroundVertices,
    groundVertices: vertices,
    groundBuildMs,
  }),
);
groundMesh.dispose();

const cachedGroundStart = performance.now();
const cachedGround = createLandscapeGround(plan, groundPlan);
console.log(JSON.stringify({ cachedGroundBuildMs: performance.now() - cachedGroundStart }));
assert.deepEqual(
  cachedGround.getAttribute("landscapeField").array,
  groundMesh.getAttribute("landscapeField").array,
);
cachedGround.dispose();

// Picking must remain correct when planets move, rotate, or use nonuniform scale.
const { sphereRaycast } = await import("../../src/utils/sphereRaycast");
const pickMesh = new THREE.Mesh(new THREE.SphereGeometry(2.02, 128, 64));
pickMesh.position.set(17, -4, 31);
pickMesh.rotation.set(0.3, -0.7, 0.2);
pickMesh.scale.set(1.2, 0.8, 1.1);
pickMesh.updateMatrixWorld();
const localPoint = new THREE.Vector3(0, 0, 2.02);
const expectedPoint = localPoint.clone().applyMatrix4(pickMesh.matrixWorld);
const origin = new THREE.Vector3(0, 0, 10).applyMatrix4(pickMesh.matrixWorld);
const pickRay = new THREE.Raycaster(origin, expectedPoint.clone().sub(origin).normalize());
const hits: THREE.Intersection[] = [];
sphereRaycast.call(pickMesh, pickRay, hits);
assert.equal(hits.length, 1);
assert.ok(hits[0].point.distanceTo(expectedPoint) < 1e-9);
assert.equal(hits[0].object, pickMesh);
pickRay.far = hits[0].distance - 0.01;
hits.length = 0;
sphereRaycast.call(pickMesh, pickRay, hits);
assert.equal(hits.length, 0);
pickRay.far = Infinity;
pickRay.near = origin.distanceTo(expectedPoint) + 0.01;
sphereRaycast.call(pickMesh, pickRay, hits);
assert.equal(hits.length, 0);
pickRay.near = 0;
pickRay.ray.direction.negate();
sphereRaycast.call(pickMesh, pickRay, hits);
assert.equal(hits.length, 0);
pickRay.ray.direction.negate();
const trianglePickStart = performance.now();
for (let i = 0; i < 500; i++) {
  hits.length = 0;
  pickMesh.raycast(pickRay, hits);
}
const trianglePickMs = performance.now() - trianglePickStart;
const spherePickStart = performance.now();
for (let i = 0; i < 500; i++) {
  hits.length = 0;
  sphereRaycast.call(pickMesh, pickRay, hits);
}
console.log(JSON.stringify({ trianglePickMs, spherePickMs: performance.now() - spherePickStart }));
pickMesh.geometry.dispose();
pickMesh.material.dispose();

// Transferring a result must not detach cached patches used by later placements.
const worker = new Worker(
  new URL("../../src/components/game/board/landscape.worker.ts", import.meta.url),
);
const workerJobs = new Map<number, (result: LandscapeResult) => void>();
worker.onmessage = ({ data }: MessageEvent<LandscapeResult>) => {
  workerJobs.get(data.id)?.(data);
  workerJobs.delete(data.id);
};
const requestLandscape = (request: LandscapeRequest) =>
  new Promise<LandscapeResult>((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error("Landscape worker timed out")), 10000);
    workerJobs.set(request.id, (result) => {
      clearTimeout(timeout);
      resolve(result);
    });
    worker.postMessage(request);
  });
try {
  const request: LandscapeRequest = {
    id: 1,
    tiles: [{ coordinate: { q: 0, r: 0, s: 0 }, kind: "greenery", seed: 432 }],
    spaces: [],
    oceans: [],
  };
  const first = await requestLandscape(request);
  const cached = await requestLandscape({ ...request, id: 2 });
  assert.deepEqual(cached.plan, first.plan);
  assert.deepEqual(cached.ground, first.ground);
  assert.ok(first.plan.reserveCells instanceof Map);
  assert.ok(first.ground.attributes.position.array instanceof Float32Array);
  assert.ok(first.ground.index!.length > 0);
  const coastal = await requestLandscape({
    ...request,
    id: 3,
    oceans: [{ coordinate: { q: 1, r: 0, s: -1 } }],
  });
  assert.equal(coastal.planKey, first.planKey);
  assert.ok(coastal.ground.index!.length > first.ground.index!.length);
  const [empty, replacement] = await Promise.all([
    requestLandscape({ ...request, id: 4, tiles: [] }),
    requestLandscape({ ...request, id: 5 }),
  ]);
  assert.equal(empty.plan.tiles.length, 0);
  assert.equal(replacement.plan.tiles.length, 1);
  assert.deepEqual(replacement.ground, first.ground);
  console.log("Worker transfers, cache reuse, and rapid placements passed.");
} finally {
  worker.terminate();
}
