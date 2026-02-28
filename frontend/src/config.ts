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
 * Get the runtime configuration.
 * Priority:
 * 1. window.__RUNTIME_CONFIG__ (set by runtime-config.js at container startup)
 * 2. import.meta.env.VITE_API_URL (build-time env var, for development)
 * 3. Default fallback
 */
function getConfig(): RuntimeConfig {
  const runtimeConfig = window.__RUNTIME_CONFIG__;

  return {
    apiUrl: runtimeConfig?.apiUrl || import.meta.env.VITE_API_URL || DEFAULT_API_URL,
  };
}

export const config = getConfig();

export const APP_VERSION: string = import.meta.env.VITE_APP_VERSION || "localbuild";

/**
 * Derives the WebSocket URL from the API URL.
 * - If apiUrl is a relative path (e.g., "/api/v1"), uses current host with appropriate protocol
 * - If apiUrl is absolute, derives the WS URL from it
 */
export function getWebSocketUrl(): string {
  const { apiUrl } = config;

  // Handle relative URL (e.g., "/api/v1")
  if (apiUrl.startsWith("/")) {
    if (typeof window !== "undefined") {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return `${protocol}//${window.location.host}/ws`;
    }
    return "ws://localhost:3001/ws";
  }

  // Handle absolute URL - derive WS URL from it
  try {
    const url = new URL(apiUrl);
    const wsProtocol = url.protocol === "https:" ? "wss:" : "ws:";
    return `${wsProtocol}//${url.host}/ws`;
  } catch {
    // Fallback for invalid URLs
    return "ws://localhost:3001/ws";
  }
}
