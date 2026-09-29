# Board - 3D Mars Tile Rendering

React Three Fiber components that render the hexagonal game board on a 3D Mars sphere. Tiles are projected from 2D hex coordinates onto the sphere surface using azimuthal projection.

## Component Hierarchy

```
MarsSphere          -> Textured sphere + rotation context
+-- TileGrid        -> Generates projected hex positions, detects new tiles
|   +-- Tile        -> Base hex tile (chrome border, hover, highlights, VP text)
|   +-- CityRenderer     -> Procedural city geometry with rise emergence
|   |   +-- VolcanoTile   -> Procedural volcano (custom GLSL shaders + smoke)
|   +-- LandscapeRenderer -> Shared ground fields, recessed lakes, and shoreline materials
|   +-- GreeneryRenderer  -> InstancedMesh vegetation (trees, bushes, clover, rocks)
+-- GpuWarmup       -> Invisible warmup meshes to prevent first-render stalls
```

## Tile Types

Each hex can be one of: `empty`, `ocean`, `greenery`, `city`, `special`, `volcano`.

- **Tile** (`Tile.tsx`) — Base component for every hex. Renders the chrome border, hover/available glow, owner color, VP text. Delegates to child components based on type.
- **LandscapeRenderer** — Owns `LandscapeSurface`: instanced lake stencil masks, opaque basin terrain, transparent outer ground, and water. `landscapeFields.ts` builds continuous height/material fields in a worker; adjacent ocean junctions are filled while island holes remain. Water is below the Mars surface and sandy banks rise to surrounding ground. Lake and Mohole masks use independent stencil bits. All passes share texture layers and a 600 ms transition clock.
- **CityRenderer** — Renders seeded procedural cities on Mars and celestial bodies; surface projection uses the body radius.
- **VolcanoTile** — Procedural volcano cone with crater, lava flows, and smoke particles. Uses custom vertex/fragment shaders (`volcano.vert.glsl`, `volcano.frag.glsl`). Height field is generated on both GPU (shaders) and CPU (flow map computation in `volcanoFlowMap.ts`). Smoke effect in `effects/VolcanoSmoke.tsx`.
- **GreeneryRenderer** — Handles ALL vegetation (trees, bushes, clover, rocks) for both greenery tiles AND volcano tiles using InstancedMesh.

## Vegetation Management

**CRITICAL**: All trees and bushes MUST be placed through `GreeneryRenderer`. Never create standalone tree/bush meshes in individual tile components.

- GreeneryRenderer receives the displayed and previous landscape states and the terrain transition timestamp from TileGrid.
- The landscape worker deterministically scatters plants in world-space cells, with city/volcano exclusions, beach clearances, and reduced tree density near exposed hex borders.
- Uses InstancedMesh per variant with stable slots; plant heights interpolate with the displayed terrain.
- Vegetation models come from `useModels()` through `createVariantsFromScene()`.

## Coordinate System

Cube coordinates `(q, r, s)` where `q + r + s = 0`. TileGrid converts these to 2D pixel positions, then projects onto the sphere via azimuthal projection (`projectToSphere`). The projection maps the flat hex grid onto the front hemisphere.

## Key Constants

- `SPHERE_RADIUS = 2.02` — Mars sphere radius, shared by the board and landscape projection
- `CHROME_Z_BASE = 0.0156` — Base z-offset for tile chrome to prevent z-fighting
- `HEX_SIZE = 0.3` — Hex cell size for coordinate conversion
- `HEX_RADIUS = 0.166` — Hex tile radius in world units
- Projection scale `0.4` — Controls hex spacing on sphere

## Volcano Shader System

The volcano uses a custom procedural heightfield:

- **Vertex shader** (`volcano.vert.glsl`): Generates cone + crater + rim + gullies from seed-based parameters. Uses domain warp and shape field to break circular silhouette. Normals via finite differences.
- **Fragment shader** (`volcano.frag.glsl`): 3-layer rock material, flow-map-based lava, crater magma, grass at base. Debug modes (uDebugMode: 0=normal, 1=height, 2=slope, 3=gully, 4=flow).
- **Flow map** (`volcanoFlowMap.ts`): CPU-side height sampling on 64x64 grid with steepest-descent flow accumulation. Uploaded as DataTexture for valley-seeking lava.
- **Material factory** (`shaders/index.ts`): `createVolcanoMaterial(grassTexture, flowTexture, seed)` creates the ShaderMaterial with all uniforms.

**Important**: The vertex shader and flow map TypeScript MUST stay in sync. Both implement the same `getHeight()` function. Any change to height generation in the shader must be mirrored in `volcanoFlowMap.ts`.

## Shader Uniform Updates — useMemo + useRef Pitfall

**CRITICAL**: Never use `useRef` to track materials created in `useMemo`, then update uniforms via the ref in `useFrame`. React Strict Mode runs `useMemo` callbacks **twice** in development, discarding the second return value but keeping side effects. A `materialRef.current = mat` inside `useMemo` ends up pointing to the discarded material, so all uniform updates in `useFrame` silently go to a ghost material the mesh doesn't render.

```tsx
// BAD — ref points to wrong material in Strict Mode
const materialRef = useRef(null);
const material = useMemo(() => {
  const mat = createMyMaterial();
  materialRef.current = mat; // overwrites with discarded 2nd run
  return mat;
}, [deps]);
useFrame(() => {
  materialRef.current.uniforms.uFoo.value = x;
}); // ghost material

// GOOD — use the useMemo return value directly
const material = useMemo(() => createMyMaterial(), [deps]);
useFrame(() => {
  material.uniforms.uFoo.value = x;
}); // always the real material
```

## Asset Loading

All 3D models and textures are loaded via centralized hooks in `hooks/`:

- `useModels()` — vegetation, rocks, and celestial models; cities use procedural geometry
- `useTextures()` — terrain textures, resource icons, effects textures

No direct `useGLTF`, `useTexture`, or `useLoader(TextureLoader)` calls in board components.

- **GpuWarmup** renders tiny invisible instances of every material/shader to eliminate first-use GPU compilation hitches.

## Shaders (`shaders/`)

All GLSL shaders live in `.glsl` files imported via Vite `?raw`. The `shaders/index.ts` barrel exports all shaders plus a `splitSnippet()` utility for `onBeforeCompile` snippets (header/body separated by `//#pragma body`).

- **Complete shaders**: ocean-renderer (vert+frag, GLSL 300 es), sphere projection (shared vert), ocean border, hover/available/endgame glow, tile border (vert+frag), volcano (vert+frag)
- **onBeforeCompile snippets**: tile surface projection, greenery ground (vert+frag with hex SDF soft edges)
- Z-offsets use `uZOffset` uniform (not baked into shader strings)

## Effects (`effects/`)

- **DustEffect** — Smoke particle cloud using billboard planes with a smoke texture. Used by CityRenderer for city placement.
- **VolcanoSmoke** — Sprite-based smoke particles emitted from crater. Uses `renderOrder = 100` to render above tile geometry.

Landscape placement transitions live in `LandscapeSurface.tick`: changed layers upload before activation, lake cutouts cover the previous/current basin union during interpolation, and removed slots are released after the transition.

## External Exports

Other parts of the codebase import from this directory:

- `TileHighlightMode` type from `Tile` — used by view, layout, and display components for VP scoring highlights
- `variantCache`, `createVariantsFromScene`, tree/bush/clover name arrays from `GreeneryRenderer` — used by GpuWarmup
- `addSphereProjectionWithSoftEdges` from `GreeneryRenderer` — used by VolcanoTile for ground mesh
