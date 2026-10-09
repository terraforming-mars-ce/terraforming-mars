import { probeServer, type ProbeResult } from "./probe.ts";
import type { MetaResponse, ServerEntry } from "./servers.ts";

export interface PickerOptions {
  servers: ServerEntry[];
  notice?: string;
  currentAlias: string | null;
  onPick: (server: ServerEntry, meta: MetaResponse) => void;
}

/**
 * Lists every server with its live status. Servers that do not answer, or
 * answer as someone else, stay listed but cannot be picked.
 */
export function showPicker({ servers, notice, currentAlias, onPick }: PickerOptions): void {
  const root = document.getElementById("gateway");
  if (!root) {
    return;
  }
  root.hidden = false;
  root.replaceChildren();

  const heading = element("h1", "Terraforming Mars");
  const subtitle = element("p", "Choose a server", "subtitle");
  root.append(heading, subtitle);
  if (notice) {
    root.append(element("p", notice, "notice"));
  }
  if (servers.length === 0) {
    root.append(element("p", "No servers are configured.", "notice"));
    return;
  }

  const list = element("ul", "", "servers");
  root.append(list);
  for (const server of servers) {
    const row = document.createElement("li");
    const button = document.createElement("button");
    button.type = "button";
    button.disabled = true;
    const name = element("span", server.alias, "name");
    const detail = element("span", "Checking…", "detail");
    button.append(name, detail);
    if (server.alias === currentAlias) {
      button.append(element("span", "Last used", "badge"));
    }
    row.append(button);
    list.append(row);

    void probeServer(server).then((result) => renderStatus(result, server, button, name, detail));
  }

  function renderStatus(
    result: ProbeResult,
    server: ServerEntry,
    button: HTMLButtonElement,
    name: HTMLElement,
    detail: HTMLElement,
  ) {
    if (!result.ok) {
      detail.textContent = `Offline (${result.reason})`;
      button.classList.add("offline");
      return;
    }
    name.textContent = result.meta.name;
    detail.textContent = `${result.meta.version} · ${result.latencyMs} ms`;
    button.disabled = false;
    button.addEventListener("click", () => onPick(server, result.meta));
  }
}

export function hidePicker(): void {
  const root = document.getElementById("gateway");
  if (root) {
    root.hidden = true;
    root.replaceChildren();
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
