import { useEffect, useRef } from "react";
import { APP_VERSION } from "@/config.ts";
import { apiService } from "@/services/apiService.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { useVersionStore } from "@/stores/versionStore.ts";
import { serverScopedKey } from "@/utils/gateway.ts";
import { isLocalBuild } from "@/utils/version.ts";

const CHECK_INTERVAL_MS = 5 * 60 * 1000;
const LAST_SEEN_KEY = "openmars.lastSeenVersion";

function readLastSeen(): string | null {
  try {
    return localStorage.getItem(serverScopedKey(LAST_SEEN_KEY));
  } catch {
    return null;
  }
}

function writeLastSeen(version: string) {
  try {
    localStorage.setItem(serverScopedKey(LAST_SEEN_KEY), version);
  } catch {
    console.warn("Failed to save last seen version to localStorage");
  }
}

/**
 * Keeps the server's version current so a stale client can offer an update, and
 * opens the changelog once after the client itself was updated. A redeploy
 * restarts the server, so reconnects and returning to the tab are checked too.
 */
export function useVersionCheck(inMenuRoute: boolean) {
  const whatsNewHandled = useRef(false);

  useEffect(() => {
    if (isLocalBuild(APP_VERSION)) {
      return;
    }

    const check = () => {
      if (document.hidden) {
        return;
      }
      apiService.getMeta().then(
        (meta) => useVersionStore.getState().setServerVersion(meta.version),
        () => undefined,
      );
    };
    const handleVisibility = () => {
      if (!document.hidden) {
        check();
      }
    };

    check();
    const interval = window.setInterval(check, CHECK_INTERVAL_MS);
    document.addEventListener("visibilitychange", handleVisibility);
    globalWebSocketManager.on("connect", check);
    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", handleVisibility);
      globalWebSocketManager.off("connect", check);
    };
  }, []);

  useEffect(() => {
    if (!inMenuRoute || whatsNewHandled.current || isLocalBuild(APP_VERSION)) {
      return;
    }
    whatsNewHandled.current = true;

    const lastSeen = readLastSeen();
    writeLastSeen(APP_VERSION);
    if (lastSeen === null || lastSeen === APP_VERSION) {
      return;
    }
    apiService.getChangelog().then(
      (changelog) => {
        if (changelog.entries.some((entry) => entry.version === APP_VERSION)) {
          useVersionStore.getState().openChangelog(APP_VERSION);
        }
      },
      () => undefined,
    );
  }, [inMenuRoute]);
}
