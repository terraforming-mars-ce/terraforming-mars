import { findServer, type ServerEntry } from "./servers.ts";

/**
 * Where a boot decision came from. Only a shared link or an explicit pick is
 * remembered; the default follows whatever server the gateway lists first.
 */
export type BootSource = "link" | "tab" | "choice" | "default";

export type Selection =
  | { kind: "boot"; server: ServerEntry; source: BootSource }
  | { kind: "pick"; notice?: string };

export interface SelectionInput {
  params: URLSearchParams;
  /** The server this tab used last; survives reloads */
  tabAlias: string | null;
  /** The server this browser explicitly picked or was linked to */
  choiceAlias: string | null;
  servers: ServerEntry[];
}

/**
 * Decides which server to boot. `?pick` always asks; a shared link (`?s=`)
 * wins over this tab's server, which wins over the remembered choice. With
 * none of those the first listed server is the default.
 */
export function selectServer({
  params,
  tabAlias,
  choiceAlias,
  servers,
}: SelectionInput): Selection {
  if (params.has("pick")) {
    return { kind: "pick" };
  }
  const linked = params.get("s");
  if (linked !== null) {
    const server = findServer(servers, linked);
    return server
      ? { kind: "boot", server, source: "link" }
      : { kind: "pick", notice: `There is no server called "${linked}".` };
  }
  const tab = findServer(servers, tabAlias);
  if (tab) {
    return { kind: "boot", server: tab, source: "tab" };
  }
  const choice = findServer(servers, choiceAlias);
  if (choice) {
    return { kind: "boot", server: choice, source: "choice" };
  }
  if (servers.length > 0) {
    return { kind: "boot", server: servers[0], source: "default" };
  }
  return { kind: "pick" };
}

/** The default server first, then the rest in listed order: the fallback order when the default is down. */
export function defaultOrder(servers: ServerEntry[], first: ServerEntry): ServerEntry[] {
  return [first, ...servers.filter((server) => server !== first)];
}
