/**
 * Runtime configuration that can be set via environment variables at container startup.
 * This allows the same Docker image to be deployed to different environments
 * without rebuilding.
 */

interface RuntimeConfig {
  apiUrl: string;
}

declare global {
  interface Window {
    __RUNTIME_CONFIG__?: RuntimeConfig;
  }
}

const DEFAULT_API_URL = "/api/v1";

/**
 * Resolves a server path against the origin this bundle was loaded from. A
 * gateway boots the bundle from another origin than the page, so server URLs
 * must never resolve against `window.location`.
 */
export function serverUrl(path: string): string {
  return new URL(path, import.meta.url).href;
}

/**
 * Get the runtime configuration.
 * Priority:
 * 1. window.__RUNTIME_CONFIG__ (set by runtime-config.js at container startup)
 * 2. `/api/v1` on the server that served this bundle (the Vite dev server proxies it to the backend)
 */
function getConfig(): RuntimeConfig {
  const runtimeConfig = window.__RUNTIME_CONFIG__;

  return {
    apiUrl: serverUrl(runtimeConfig?.apiUrl || DEFAULT_API_URL),
  };
}

export const config = getConfig();

export const APP_VERSION: string = import.meta.env.VITE_APP_VERSION || "localbuild";

/**
 * Derives the WebSocket URL from the API URL's host.
 */
export function getWebSocketUrl(): string {
  const url = new URL(config.apiUrl);
  const wsProtocol = url.protocol === "https:" ? "wss:" : "ws:";
  return `${wsProtocol}//${url.host}/ws`;
}
