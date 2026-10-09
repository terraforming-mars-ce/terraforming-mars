import { bootServer } from "./boot.ts";
import { hidePicker, showPicker } from "./picker.ts";
import { probeServer } from "./probe.ts";
import { defaultOrder, selectServer } from "./select.ts";
import type { MetaResponse, ServerEntry } from "./servers.ts";

// Loader keys never collide with the game's own storage keys on this origin
const TAB_KEY = "tm.gateway.tab";
const CHOICE_KEY = "tm.gateway.choice";

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
    currentAlias: read(sessionStorage, TAB_KEY) ?? read(localStorage, CHOICE_KEY),
    onPick: (server, meta) => void start(server, meta, true),
  });
}

/**
 * Pins the server to this tab, remembers it for the browser when the player
 * chose it, drops the gateway's own query params, then hands the page to the
 * server's bundle.
 */
async function start(server: ServerEntry, meta: MetaResponse, remember: boolean): Promise<void> {
  write(sessionStorage, TAB_KEY, server.alias);
  if (remember) {
    write(localStorage, CHOICE_KEY, server.alias);
  }

  const url = new URL(window.location.href);
  url.searchParams.delete("s");
  url.searchParams.delete("pick");
  history.replaceState(history.state, "", url);

  hidePicker();
  try {
    await bootServer(server, meta, servers);
  } catch (error) {
    console.error("Failed to boot server", server.alias, error);
    pick(`${meta.name} could not be started. Pick another server.`);
  }
}

/** Boots the first reachable server, starting with the default. */
async function startDefault(first: ServerEntry): Promise<void> {
  for (const server of defaultOrder(servers, first)) {
    // oxlint-disable-next-line no-await-in-loop -- try servers one at a time, in order
    const result = await probeServer(server);
    if (result.status === "online") {
      await start(server, result.meta, false);
      return;
    }
  }
  pick("No server is reachable right now.");
}

async function main(): Promise<void> {
  const selection = selectServer({
    params: new URLSearchParams(window.location.search),
    tabAlias: read(sessionStorage, TAB_KEY),
    choiceAlias: read(localStorage, CHOICE_KEY),
    servers,
  });
  if (selection.kind === "pick") {
    pick(selection.notice);
    return;
  }
  if (selection.source === "default") {
    await startDefault(selection.server);
    return;
  }
  const result = await probeServer(selection.server);
  if (result.status !== "online") {
    const name = selection.server.alias.toUpperCase();
    pick(
      result.status === "down"
        ? `${name} is not reachable right now. Pick another server.`
        : `${name} is not available. Pick another server.`,
    );
    return;
  }
  await start(selection.server, result.meta, selection.source === "link");
}

void main();
