import { findServer, type ServerEntry } from "./servers.ts";

export type Selection = { kind: "boot"; server: ServerEntry } | { kind: "pick"; notice?: string };

export interface SelectionInput {
  params: URLSearchParams;
  /** The server this tab used last; survives reloads */
  tabAlias: string | null;
  /** The server this browser used last; picks the server for a fresh visit */
  lastAlias: string | null;
  servers: ServerEntry[];
}

/**
 * Decides which server to boot. A shared link (`?s=`) wins over what this tab
 * or browser used before; `?pick` always asks.
 */
export function selectServer({ params, tabAlias, lastAlias, servers }: SelectionInput): Selection {
  if (params.has("pick")) {
    return { kind: "pick" };
  }
  const linked = params.get("s");
  if (linked !== null) {
    const server = findServer(servers, linked);
    return server
      ? { kind: "boot", server }
      : { kind: "pick", notice: `There is no server called "${linked}".` };
  }
  const remembered = findServer(servers, tabAlias) ?? findServer(servers, lastAlias);
  if (remembered) {
    return { kind: "boot", server: remembered };
  }
  if (servers.length === 1) {
    return { kind: "boot", server: servers[0] };
  }
  return { kind: "pick" };
}
