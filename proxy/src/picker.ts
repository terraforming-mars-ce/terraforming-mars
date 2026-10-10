import { probeServer, type ProbeResult } from "./probe.ts";
import type { MetaResponse, ServerEntry } from "./servers.ts";

const RECHECK_MS = 5000;

export interface PickerOptions {
  servers: ServerEntry[];
  notice?: string;
  currentAlias: string | null;
  onPick: (server: ServerEntry, meta: MetaResponse) => void;
}

let recheck: ReturnType<typeof setInterval> | undefined;

/**
 * Lists every server in order with its live status, re-checked every few
 * seconds: online servers can be picked, down ones show as offline, and ones
 * that answer but not as a gateway server are hidden.
 */
export function showPicker({ servers, notice, currentAlias, onPick }: PickerOptions): void {
  const root = document.getElementById("gateway");
  if (!root) {
    return;
  }
  hidePicker();
  root.hidden = false;

  const shell = element("div", "", "shell");
  const intro = element("header", "", "intro");
  const title = element("h1", "", "title");
  title.append("OPEN ", document.createElement("br"), "MARS");
  intro.append(title, element("p", "Server selector", "subtitle"));

  const choices = element("section", "", "choices");
  const message = element("p", notice ?? "", "notice");
  message.hidden = !notice;
  const list = element("ul", "", "servers");
  choices.append(message, list);
  shell.append(intro, choices);
  root.append(shell);

  const rows = servers.map((server) => {
    const row = new ServerRow(server, server.alias === currentAlias, onPick);
    list.append(row.item);
    return row;
  });

  const probeAll = async () => {
    const results = await Promise.all(servers.map((server) => probeServer(server)));
    results.forEach((result, index) => rows[index].update(result));
    const anyOnline = results.some((result) => result.status === "online");
    if (!anyOnline) {
      message.textContent = "No server is reachable right now.";
      message.hidden = false;
    } else if (!notice) {
      message.hidden = true;
    }
  };
  void probeAll();
  recheck = setInterval(() => void probeAll(), RECHECK_MS);
}

export function hidePicker(): void {
  clearInterval(recheck);
  recheck = undefined;
  const root = document.getElementById("gateway");
  if (root) {
    root.hidden = true;
    root.replaceChildren();
  }
}

/** One server's button, updated in place by each probe. */
class ServerRow {
  readonly item = document.createElement("li");
  private readonly button = document.createElement("button");
  private readonly name: HTMLElement;
  private readonly version = element("span", "", "version");
  private readonly status = element("span", "Checking…", "status");
  private meta: MetaResponse | null = null;

  constructor(
    private readonly server: ServerEntry,
    current: boolean,
    onPick: (server: ServerEntry, meta: MetaResponse) => void,
  ) {
    this.name = element("span", server.alias.toUpperCase(), "name");
    this.button.type = "button";
    this.button.className = "server";
    this.button.dataset.status = "checking";
    this.button.disabled = true;
    this.button.append(element("span", "", "dot"), this.name, this.version);
    if (current) {
      this.button.append(element("span", "Current", "current"));
    }
    this.button.append(this.status);
    this.button.addEventListener("click", () => {
      if (this.meta) {
        onPick(this.server, this.meta);
      }
    });
    this.item.append(this.button);
  }

  update(result: ProbeResult) {
    this.item.hidden = result.status === "invalid";
    this.button.dataset.status = result.status;
    this.button.disabled = result.status !== "online";
    if (result.status !== "online") {
      this.meta = null;
      this.version.textContent = "";
      this.status.textContent = "Offline";
      this.button.title = "";
      return;
    }
    this.meta = result.meta;
    this.name.textContent = result.meta.name;
    this.version.textContent = result.meta.version;
    this.status.textContent = `${result.latencyMs} ms`;
    this.button.title = `${result.meta.name} · ${result.meta.version}`;
  }
}

function element(tag: string, text: string, className?: string): HTMLElement {
  const node = document.createElement(tag);
  node.textContent = text;
  if (className) {
    node.className = className;
  }
  return node;
}
