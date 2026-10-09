import { bootServer } from "./boot.ts";
import { hidePicker, showPicker } from "./picker.ts";
import { probeServer } from "./probe.ts";
import { selectServer } from "./select.ts";
import type { MetaResponse, ServerEntry } from "./servers.ts";

// Loader keys never collide with the game's own storage keys on this origin
const TAB_KEY = "tm.gateway.tab";
const LAST_KEY = "tm.gateway.last";

const servers = window.__TM_SERVERS__ ?? [];

function read(storage: Storage, key: string): string | null {
  try {
    return storage.getItem(key);
  } catch {
    return null;
  }
}

function write(storage: Storage, key: string, value: string): void {
  try {
    storage.setItem(key, value);
  } catch {
    // Private mode or blocked storage: the server is just not remembered
  }
}

function pick(notice?: string): void {
  showPicker({
    servers,
    notice,
    currentAlias: read(sessionStorage, TAB_KEY) ?? read(localStorage, LAST_KEY),
    onPick: (server, meta) => void start(server, meta),
  });
}

/** Remembers the server, drops the gateway's own query params, then hands the page to the server's bundle. */
async function start(server: ServerEntry, meta: MetaResponse): Promise<void> {
  write(sessionStorage, TAB_KEY, server.alias);
  write(localStorage, LAST_KEY, server.alias);

  const url = new URL(window.location.href);
  url.searchParams.delete("s");
  url.searchParams.delete("pick");
  history.replaceState(history.state, "", url);

  hidePicker();
  try {
    await bootServer(server, meta);
  } catch (error) {
    console.error("Failed to boot server", server.alias, error);
    pick(`${meta.name} could not be started. Pick another server.`);
  }
}

async function main(): Promise<void> {
  const selection = selectServer({
    params: new URLSearchParams(window.location.search),
    tabAlias: read(sessionStorage, TAB_KEY),
    lastAlias: read(localStorage, LAST_KEY),
    servers,
  });
  if (selection.kind === "pick") {
    pick(selection.notice);
    return;
  }
  const result = await probeServer(selection.server);
  if (!result.ok) {
    pick(`${selection.server.alias} is not available (${result.reason}). Pick another server.`);
    return;
  }
  await start(selection.server, result.meta);
}

void main();
