import type { MetaResponse } from "@/types/generated/api-types.ts";
import type { GatewayServerEntry } from "./gateway.ts";

/**
 * - online: answers as itself and can be switched to
 * - down: no usable answer (network error, timeout, 5xx); shown as offline
 * - invalid: answers, but not as this server (no /meta, bad JSON, wrong alias); hidden
 *
 * Same rules as the gateway's picker (proxy/src/probe.ts). The two deploy
 * separately and cannot share code, so keep them in step.
 */
export type ServerProbe =
  | { status: "online"; meta: MetaResponse; latencyMs: number }
  | { status: "down" }
  | { status: "invalid" };

const PROBE_TIMEOUT_MS = 4000;

export async function probeServer(server: GatewayServerEntry): Promise<ServerProbe> {
  const started = performance.now();
  let response: Response;
  try {
    response = await fetch(new URL("/api/v1/meta", server.url), {
      cache: "no-store",
      signal: AbortSignal.timeout(PROBE_TIMEOUT_MS),
    });
  } catch {
    return { status: "down" };
  }
  if (response.status >= 500) {
    return { status: "down" };
  }
  if (!response.ok) {
    return { status: "invalid" };
  }
  try {
    const meta = (await response.json()) as MetaResponse;
    if (meta.alias !== server.alias) {
      return { status: "invalid" };
    }
    return { status: "online", meta, latencyMs: Math.round(performance.now() - started) };
  } catch {
    return { status: "invalid" };
  }
}
