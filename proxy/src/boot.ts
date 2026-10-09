import type { MetaResponse, ServerEntry } from "./servers.ts";

/**
 * Runs the server's own frontend on this page. The server's index.html names
 * its current bundle; its script and style tags are recreated here with
 * absolute URLs, so every file loads straight from the server.
 *
 * In the original document module scripts are deferred: classic scripts such
 * as runtime-config.js run first wherever they sit. Injected modules run as
 * soon as they load, so classic scripts are injected and awaited first.
 */
export async function bootServer(
  server: ServerEntry,
  meta: MetaResponse,
  servers: ServerEntry[],
): Promise<void> {
  const response = await fetch(new URL("/index.html", server.url), { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`index.html responded ${response.status}`);
  }
  const source = new DOMParser().parseFromString(await response.text(), "text/html");
  const resolve = (path: string | null) => new URL(path ?? "", server.url).href;

  window.__TM_GATEWAY__ = { alias: meta.alias, name: meta.name, servers };

  const scripts = [...source.querySelectorAll("script")];
  for (const original of scripts.filter((script) => script.type !== "module")) {
    // oxlint-disable-next-line no-await-in-loop -- classic scripts run in document order
    await runClassicScript(original, resolve);
  }

  for (const original of source.querySelectorAll<HTMLLinkElement>(
    "link[rel='stylesheet'], link[rel='modulepreload']",
  )) {
    const link = document.createElement("link");
    link.rel = original.rel;
    link.href = resolve(original.getAttribute("href"));
    if (original.rel === "modulepreload") {
      // Must match the module request, which always uses CORS
      link.crossOrigin = "anonymous";
    }
    document.head.append(link);
  }

  for (const original of scripts.filter((script) => script.type === "module")) {
    const script = document.createElement("script");
    script.type = "module";
    const src = original.getAttribute("src");
    if (src === null) {
      script.textContent = rebaseImports(original.textContent ?? "", server.url);
    } else {
      script.src = resolve(src);
      script.crossOrigin = "anonymous";
    }
    document.body.append(script);
  }
}

function runClassicScript(
  original: HTMLScriptElement,
  resolve: (path: string | null) => string,
): Promise<void> {
  const script = document.createElement("script");
  const src = original.getAttribute("src");
  if (src === null) {
    script.textContent = original.textContent;
    document.body.append(script);
    return Promise.resolve();
  }
  script.src = resolve(src);
  return new Promise((done, fail) => {
    script.addEventListener("load", () => done());
    script.addEventListener("error", () => fail(new Error(`failed to load ${script.src}`)));
    document.body.append(script);
  });
}

/**
 * An inline module resolves "/x" imports against this page, not the server.
 * Vite's dev server inlines one (the React refresh preamble importing
 * "/@react-refresh"); built bundles have none.
 */
export function rebaseImports(code: string, base: string): string {
  const origin = new URL(base).origin;
  return code.replace(
    /(\bfrom\s*|\bimport\s*\(?\s*)(["'])\/(?!\/)/g,
    (_match, keyword: string, quote: string) => `${keyword}${quote}${origin}/`,
  );
}
