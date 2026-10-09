/**
 * Resolves a server path against the origin this bundle was loaded from. A
 * gateway boots the bundle from another origin than the page, so server URLs
 * must never resolve against `window.location`.
 */
export function serverUrl(path: string): string {
  return new URL(path, import.meta.url).href;
}

/** The API lives on the server that served this bundle; the Vite dev server proxies it. */
export const config = {
  apiUrl: serverUrl("/api/v1"),
};

export const APP_VERSION: string = import.meta.env.VITE_APP_VERSION || "localbuild";

/**
 * Derives the WebSocket URL from the API URL's host.
 */
export function getWebSocketUrl(): string {
  const url = new URL(config.apiUrl);
  const wsProtocol = url.protocol === "https:" ? "wss:" : "ws:";
  return `${wsProtocol}//${url.host}/ws`;
}
