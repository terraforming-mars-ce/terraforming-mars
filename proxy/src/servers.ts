import type { MetaResponse } from "../../frontend/src/types/generated/api-types.ts";

export type { MetaResponse };

/** A game server the gateway can boot, from the OPENMARS_SERVERS list in servers.js. */
export interface ServerEntry {
  alias: string;
  url: string;
}

declare global {
  interface Window {
    __OPENMARS_SERVERS__?: ServerEntry[];
    /** Read by the game: the booted server and every server the gateway offers */
    __OPENMARS_GATEWAY__?: { alias: string; name: string; servers: ServerEntry[] };
  }
}

export function findServer(servers: ServerEntry[], alias: string | null): ServerEntry | null {
  return servers.find((server) => server.alias === alias) ?? null;
}
