import { afterEach, expect, test } from "bun:test";
import { Document, NodeIO } from "@gltf-transform/core";
import { ALL_EXTENSIONS } from "@gltf-transform/extensions";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import sharp from "sharp";
import { buildAssets, safePath } from "./pipeline.ts";

const roots: string[] = [];
async function fixture() {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "tm-assets-test-"));
  roots.push(root);
  await fs.mkdir(path.join(root, "assets/original"), { recursive: true });
  const image = await sharp({
    create: {
      width: 80,
      height: 40,
      channels: 4,
      background: { r: 220, g: 20, b: 30, alpha: 0.5 },
    },
  })
    .png()
    .toBuffer();
  await fs.writeFile(path.join(root, "assets/original/test.png"), image);
  await fs.writeFile(
    path.join(root, "assets/catalog.json"),
    JSON.stringify({
      version: 1,
      assets: [{ id: "icons/test", source: "test.png", profile: "icon" }],
    }),
  );
  await fs.writeFile(
    path.join(root, "assets/profiles.json"),
    JSON.stringify({ icon: { sizes: [32, 64, 128, 256], lossless: true } }),
  );
  return root;
}
async function state(root: string) {
  return JSON.parse(
    await fs.readFile(path.join(root, "frontend/.cache/assets/state.json"), "utf8"),
  );
}
afterEach(async () => {
  for (const root of roots.splice(0)) {
    await fs.rm(root, { recursive: true, force: true });
  }
});

test("exports retain transparency and aspect ratio without upscaling; unchanged runs are no-ops", async () => {
  const root = await fixture();
  expect((await buildAssets(root)).built).toBe(1);
  const first = await state(root);
  const variants = first.entries["icons/test"].asset.variants;
  expect(variants.map((v: { width: number; height: number }) => [v.width, v.height])).toEqual([
    [32, 16],
    [64, 32],
    [80, 40],
  ]);
  const registry = path.join(root, "frontend/src/assets/generated/registry.ts");
  const mtime = (await fs.stat(registry)).mtimeMs;
  const pixels = await sharp(path.join(root, "frontend/public", variants[0].url))
    .raw()
    .toBuffer();
  expect(pixels[3]).toBe(128);
  expect((await buildAssets(root)).built).toBe(0);
  expect((await fs.stat(registry)).mtimeMs).toBe(mtime);
  expect(await state(root)).toEqual(first);
  await buildAssets(root, true);
});

test("missing and corrupted outputs regenerate; changes replace hashes and prune only owned outputs", async () => {
  const root = await fixture();
  await buildAssets(root);
  const previous = await state(root);
  const output = Object.keys(previous.entries["icons/test"].outputs)[0];
  const full = path.join(root, "frontend/public", output);
  await fs.unlink(full);
  expect((await buildAssets(root)).built).toBe(1);
  await fs.writeFile(full, "corrupt");
  expect((await buildAssets(root)).built).toBe(1);
  const sentinel = path.join(root, "frontend/public/keep.txt");
  await fs.writeFile(sentinel, "unowned");
  await fs.writeFile(
    path.join(root, "assets/profiles.json"),
    JSON.stringify({ icon: { sizes: [20], lossless: true } }),
  );
  expect((await buildAssets(root)).built).toBe(1);
  expect(
    await fs.access(full).then(
      () => true,
      () => false,
    ),
  ).toBe(false);
  expect(await fs.readFile(sentinel, "utf8")).toBe("unowned");
  await fs.writeFile(
    path.join(root, "assets/original/test.png"),
    await sharp({ create: { width: 80, height: 40, channels: 4, background: "blue" } })
      .png()
      .toBuffer(),
  );
  expect((await buildAssets(root)).built).toBe(1);
});

test("failed builds preserve the published registry and previously owned files", async () => {
  const root = await fixture();
  await buildAssets(root);
  const registry = path.join(root, "frontend/src/assets/generated/registry.ts");
  const before = await fs.readFile(registry);
  const previous = await state(root);
  await fs.writeFile(
    path.join(root, "assets/original/test.png"),
    "version https://git-lfs.github.com/spec/v1\noid sha256:123\n",
  );
  await expect(buildAssets(root)).rejects.toThrow("Unresolved Git LFS");
  expect(await fs.readFile(registry)).toEqual(before);
  expect(await state(root)).toEqual(previous);
  for (const file of Object.keys(previous.entries["icons/test"].outputs)) {
    expect(await fs.stat(path.join(root, "frontend/public", file))).toBeTruthy();
  }
  await fs.unlink(path.join(root, "assets/original/test.png"));
  await expect(buildAssets(root)).rejects.toThrow();
});

test("rejects duplicate IDs, output collisions, unsafe paths and invalid profiles", async () => {
  const root = await fixture();
  const file = path.join(root, "assets/catalog.json");
  const initial = JSON.parse(await fs.readFile(file, "utf8"));
  await fs.writeFile(
    file,
    JSON.stringify({ version: 1, assets: [...initial.assets, ...initial.assets] }),
  );
  await expect(buildAssets(root)).rejects.toThrow("duplicate asset ID");
  await fs.writeFile(
    file,
    JSON.stringify({ version: 1, assets: [{ ...initial.assets[0], source: "../secret" }] }),
  );
  await expect(buildAssets(root)).rejects.toThrow("Unsafe asset path");
  await fs.writeFile(
    file,
    JSON.stringify({
      version: 1,
      assets: [
        { ...initial.assets[0], publicName: "test.png" },
        { ...initial.assets[0], id: "icons/second", publicName: "test.png" },
      ],
    }),
  );
  await expect(buildAssets(root)).rejects.toThrow("duplicate public output");
  await fs.writeFile(file, JSON.stringify(initial));
  await fs.writeFile(
    path.join(root, "assets/profiles.json"),
    JSON.stringify({ icon: { sizes: [-1] } }),
  );
  await expect(buildAssets(root)).rejects.toThrow("Invalid profile");
  expect(() => safePath(root, "../outside")).toThrow();
});

test("copies original bytes and serializes overlapping builds", async () => {
  const root = await fixture();
  const profiles = path.join(root, "assets/profiles.json");
  await fs.writeFile(profiles, JSON.stringify({ icon: { copy: true } }));
  const results = await Promise.all([buildAssets(root), buildAssets(root)]);
  expect(results.map((r) => r.built).sort()).toEqual([0, 1]);
  const output = Object.keys((await state(root)).entries["icons/test"].outputs)[0];
  expect(await fs.readFile(path.join(root, "frontend/public", output))).toEqual(
    await fs.readFile(path.join(root, "assets/original/test.png")),
  );
});

test("bakes the charcoal matte into alpha and crops logo margins", async () => {
  const root = await fixture();
  const pixels = Buffer.from([
    16, 20, 25, 255, 255, 255, 255, 255, 220, 30, 10, 255, 16, 20, 25, 255,
  ]);
  await fs.writeFile(
    path.join(root, "assets/original/test.png"),
    await sharp(pixels, { raw: { width: 4, height: 1, channels: 4 } })
      .png()
      .toBuffer(),
  );
  await fs.writeFile(
    path.join(root, "assets/catalog.json"),
    JSON.stringify({
      version: 1,
      assets: [
        {
          id: "icons/test",
          source: "test.png",
          profile: "icon",
          prepare: { crop: [0, 0, 0.75, 1], charcoalMatte: true },
        },
      ],
    }),
  );
  await buildAssets(root);
  const asset = (await state(root)).entries["icons/test"].asset;
  const { data, info } = await sharp(path.join(root, "frontend/public", asset.variants[0].url))
    .raw()
    .toBuffer({ resolveWithObject: true });
  expect(info.width).toBe(3);
  expect(data[3]).toBe(0);
  expect(data[7]).toBe(255);
  expect(data[11]).toBe(255);
});

test("preserves case-sensitive card IDs in the registry and output paths", async () => {
  const root = await fixture();
  await fs.writeFile(
    path.join(root, "assets/catalog.json"),
    JSON.stringify({
      version: 1,
      assets: [{ id: "cards/C01", source: "test.png", profile: "icon" }],
    }),
  );
  await buildAssets(root);
  const entry = (await state(root)).entries["cards/C01"];
  expect(entry.asset.variants[0].url).toContain("/C01.");
  expect(
    await fs.readFile(path.join(root, "frontend/src/assets/generated/registry.ts"), "utf8"),
  ).toContain('"cards/C01"');
});

test("model profiles shrink GLB textures and keep the scene graph and geometry", async () => {
  const root = await fixture();
  const document = new Document();
  const buffer = document.createBuffer();
  const positions = new Float32Array([0, 0, 0, 1.5, 0, 0, 0, 2.25, 0]);
  const position = document.createAccessor().setType("VEC3").setArray(positions).setBuffer(buffer);
  const uv = document
    .createAccessor()
    .setType("VEC2")
    .setArray(new Float32Array([0, 0, 1, 0, 0, 1]))
    .setBuffer(buffer);
  const pixels = Buffer.alloc(256 * 256 * 3);
  for (let i = 0; i < pixels.length; i++) {
    pixels[i] = (i * 7919) % 251;
  }
  const image = await sharp(pixels, { raw: { width: 256, height: 256, channels: 3 } })
    .png()
    .toBuffer();
  const texture = document.createTexture("color").setImage(image).setMimeType("image/png");
  const material = document.createMaterial("leaf").setBaseColorTexture(texture);
  const primitive = document
    .createPrimitive()
    .setAttribute("POSITION", position)
    .setAttribute("TEXCOORD_0", uv)
    .setMaterial(material);
  const mesh = document.createMesh("Tree").addPrimitive(primitive);
  const node = document.createNode("Tree-01-1").setMesh(mesh).setTranslation([1, 2, 3]);
  document.createScene().addChild(node).addChild(document.createNode("Marker"));
  const io = new NodeIO().registerExtensions(ALL_EXTENSIONS);
  const source = Buffer.from(await io.writeBinary(document));
  await fs.writeFile(path.join(root, "assets/original/test.glb"), source);
  await fs.writeFile(
    path.join(root, "assets/catalog.json"),
    JSON.stringify({
      version: 1,
      assets: [{ id: "models/test", source: "test.glb", profile: "model" }],
    }),
  );
  await fs.writeFile(
    path.join(root, "assets/profiles.json"),
    JSON.stringify({ model: { gltf: { maxTextureSize: 64 }, quality: 90 } }),
  );
  await buildAssets(root);
  const asset = (await state(root)).entries["models/test"].asset;
  expect(asset.variants[0].url).toMatch(/^\/assets\/models\/test\.[0-9a-f]{16}\.glb$/);
  const output = await fs.readFile(path.join(root, "frontend/public", asset.variants[0].url));
  expect(output.length).toBeLessThan(source.length);
  const optimized = (await io.readBinary(new Uint8Array(output))).getRoot();
  expect(optimized.listNodes().map((n) => n.getName())).toEqual(["Tree-01-1", "Marker"]);
  expect(optimized.listNodes()[0].getTranslation()).toEqual([1, 2, 3]);
  const [optimizedTexture] = optimized.listTextures();
  expect(optimizedTexture.getMimeType()).toBe("image/webp");
  expect(optimizedTexture.getSize()).toEqual([64, 64]);
  const optimizedPosition = optimized.listMeshes()[0].listPrimitives()[0].getAttribute("POSITION")!;
  expect(optimizedPosition.getComponentType()).toBe(5126);
  expect(Array.from(optimizedPosition.getArray()!)).toEqual(Array.from(positions));
  await fs.writeFile(
    path.join(root, "assets/catalog.json"),
    JSON.stringify({
      version: 1,
      assets: [{ id: "models/test", source: "test.png", profile: "model" }],
    }),
  );
  await expect(buildAssets(root)).rejects.toThrow("Invalid model profile");
});
