import { cp, readFile, rm, writeFile } from "node:fs/promises";
import { basename, join } from "node:path";

const root = join(import.meta.dir, "..");
const repo = join(root, "..");
export const outDir = join(root, "dist");

/**
 * Builds the gateway into dist/: the hashed loader, the shell page that loads
 * it, and the app icons and manifest shared with the game frontend.
 * servers.js is written at startup from TM_SERVERS, not here.
 */
export async function build(): Promise<void> {
  await rm(outDir, { recursive: true, force: true });
  const result = await Bun.build({
    entrypoints: [join(root, "src/loader.ts")],
    outdir: join(outDir, "assets"),
    naming: "[name]-[hash].[ext]",
    minify: true,
    sourcemap: "linked",
  });
  if (!result.success) {
    throw new AggregateError(result.logs, "loader build failed");
  }
  const loader = result.outputs.find((output) => output.kind === "entry-point");
  if (!loader) {
    throw new Error("loader build produced no entry point");
  }
  const shell = await readFile(join(root, "src/index.html"), "utf8");
  await writeFile(
    join(outDir, "index.html"),
    shell.replace("%LOADER%", `assets/${basename(loader.path)}`),
  );
  await cp(join(repo, "assets/original/app"), outDir, { recursive: true });
  await cp(join(repo, "frontend/public/manifest.json"), join(outDir, "manifest.json"));
}

if (import.meta.main) {
  await build();
  console.log(`Gateway built to ${outDir}`);
}
