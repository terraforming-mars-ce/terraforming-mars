import type { MetaResponse } from "../../frontend/src/types/generated/api-types.ts";

export type { MetaResponse };

/** A game server the gateway can boot, from the TM_SERVERS list in servers.js. */
export interface ServerEntry {
  alias: string;
  url: string;
}

declare global {
  interface Window {
    __TM_SERVERS__?: ServerEntry[];
    __TM_GATEWAY__?: { alias: string; name: string };
  }
}

export function findServer(servers: ServerEntry[], alias: string | null): ServerEntry | null {
  return servers.find((server) => server.alias === alias) ?? null;
}
