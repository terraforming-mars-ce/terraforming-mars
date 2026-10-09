import { watch } from "node:fs";
import { join } from "node:path";
import { build, outDir } from "./build.ts";

const root = join(import.meta.dir, "..");
const port = Number(process.env.PORT ?? 4000);
const servers = process.env.TM_SERVERS ?? "local=http://localhost:4173";

// servers.js comes from the same entrypoint script the container runs
async function writeServers(): Promise<void> {
  const entrypoint = Bun.spawn(["sh", join(root, "docker-entrypoint.sh"), "true"], {
    env: { ...process.env, TM_SERVERS: servers, SERVERS_JS: join(outDir, "servers.js") },
    stdout: "inherit",
    stderr: "inherit",
  });
  if ((await entrypoint.exited) !== 0) {
    throw new Error("docker-entrypoint.sh rejected TM_SERVERS");
  }
}

async function rebuild(): Promise<void> {
  try {
    await build();
    await writeServers();
    console.log("Gateway rebuilt");
  } catch (error) {
    console.error(error);
  }
}

await rebuild();
let pending: ReturnType<typeof setTimeout> | undefined;
watch(join(root, "src"), { recursive: true }, () => {
  clearTimeout(pending);
  pending = setTimeout(() => void rebuild(), 100);
});

// Same routing as nginx.conf: files when they exist, the shell for every other path
Bun.serve({
  port,
  async fetch(request) {
    const { pathname } = new URL(request.url);
    const file = Bun.file(join(outDir, pathname));
    if (pathname !== "/" && !pathname.includes("..") && (await file.exists())) {
      return new Response(file, { headers: { "Cache-Control": "no-store" } });
    }
    if (pathname.startsWith("/assets/")) {
      return new Response("Not found", { status: 404 });
    }
    return new Response(Bun.file(join(outDir, "index.html")), {
      headers: { "Cache-Control": "no-store" },
    });
  },
});
console.log(`Gateway on http://localhost:${port} for ${servers}`);
