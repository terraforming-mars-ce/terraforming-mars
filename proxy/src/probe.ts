import type { MetaResponse, ServerEntry } from "./servers.ts";

export type ProbeResult =
  | { ok: true; meta: MetaResponse; latencyMs: number }
  | { ok: false; reason: string };

const PROBE_TIMEOUT_MS = 4000;

/**
 * Asks a server who it is. A server that answers with another alias than the
 * gateway lists is treated as down, so a misconfigured list never boots the
 * wrong game.
 */
export async function probeServer(
  server: ServerEntry,
  fetchMeta: typeof fetch = fetch,
  timeoutMs = PROBE_TIMEOUT_MS,
): Promise<ProbeResult> {
  const started = performance.now();
  try {
    const response = await fetchMeta(new URL("/api/v1/meta", server.url), {
      cache: "no-store",
      signal: AbortSignal.timeout(timeoutMs),
    });
    if (!response.ok) {
      return { ok: false, reason: `responded ${response.status}` };
    }
    const meta = (await response.json()) as MetaResponse;
    if (meta.alias !== server.alias) {
      return { ok: false, reason: `identifies as "${meta.alias}"` };
    }
    return { ok: true, meta, latencyMs: Math.round(performance.now() - started) };
  } catch {
    return { ok: false, reason: "unreachable" };
  }
}
