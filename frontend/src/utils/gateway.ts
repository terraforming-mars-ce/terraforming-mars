/**
 * A gateway fronts several independent servers on one domain. Its loader boots
 * this bundle from the chosen server and describes that server here. On a
 * direct visit there is no gateway and the page origin is the server.
 */
interface GatewayServer {
  alias: string;
  name: string;
}

declare global {
  interface Window {
    __TM_GATEWAY__?: GatewayServer;
  }
}

export const gatewayServer: GatewayServer | null = window.__TM_GATEWAY__ ?? null;

/**
 * Keys for data that only means something on one server (game sessions). All
 * servers behind a gateway share the gateway origin's storage.
 */
export function serverScopedKey(key: string): string {
  return gatewayServer ? `${gatewayServer.alias}:${key}` : key;
}

/**
 * A link to a page on this origin. Under a gateway it names the server, so
 * whoever opens it lands on the same server.
 */
export function serverLink(path: string): string {
  const url = new URL(path, window.location.origin);
  if (gatewayServer) {
    url.searchParams.set("s", gatewayServer.alias);
  }
  return url.href;
}

/** Leaves the bundle for the gateway's server picker. */
export function changeServer(): void {
  window.location.assign("/?pick");
}
