import type { MetaResponse, ServerEntry } from "./servers.ts";

/**
 * - online: answers as itself and can be booted
 * - down: no usable answer (network error, timeout, 5xx); shown as offline
 * - invalid: answers, but not as this server (no /meta, bad JSON, wrong alias); hidden
 */
export type ProbeResult =
  | { status: "online"; meta: MetaResponse; latencyMs: number }
  | { status: "down"; reason: string }
  | { status: "invalid"; reason: string };

const PROBE_TIMEOUT_MS = 4000;

/**
 * Asks a server who it is. A server that is down has no CORS headers on its
 * error page either, so the browser reports a network error: also down.
 */
export async function probeServer(
  server: ServerEntry,
  fetchMeta: typeof fetch = fetch,
  timeoutMs = PROBE_TIMEOUT_MS,
): Promise<ProbeResult> {
  const started = performance.now();
  let response: Response;
  try {
    response = await fetchMeta(new URL("/api/v1/meta", server.url), {
      cache: "no-store",
      signal: AbortSignal.timeout(timeoutMs),
    });
  } catch {
    return { status: "down", reason: "unreachable" };
  }
  if (response.status >= 500) {
    return { status: "down", reason: `responded ${response.status}` };
  }
  if (!response.ok) {
    return { status: "invalid", reason: `responded ${response.status}` };
  }
  let meta: MetaResponse;
  try {
    meta = (await response.json()) as MetaResponse;
  } catch {
    return { status: "invalid", reason: "not a meta response" };
  }
  if (meta.alias !== server.alias) {
    return { status: "invalid", reason: `identifies as "${meta.alias}"` };
  }
  return { status: "online", meta, latencyMs: Math.round(performance.now() - started) };
}
