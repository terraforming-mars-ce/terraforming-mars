import React, { useEffect, useState } from "react";
import GameButton from "./buttons/GameButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useNotifications } from "@/contexts/NotificationContext.tsx";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { useAppPhaseStore } from "@/stores/appPhaseStore.ts";
import { useInstallAppStore } from "@/stores/installAppStore.ts";
import { useInstallOffer } from "@/utils/installApp.ts";

const DISMISS_KEY = "tm.installBar.dismissedUntil";
const DISMISS_MS = 30 * 24 * 60 * 60 * 1000;
const SHOW_DELAY_MS = 2000;

function readDismissed(): boolean {
  try {
    return Number(localStorage.getItem(DISMISS_KEY) ?? 0) > Date.now();
  } catch {
    return false;
  }
}

function writeDismissed() {
  try {
    localStorage.setItem(DISMISS_KEY, String(Date.now() + DISMISS_MS));
  } catch {
    console.warn("Failed to save install bar dismissal to localStorage");
  }
}

function useInstalledNotice() {
  const { showNotification } = useNotifications();
  useEffect(
    () =>
      useInstallAppStore.subscribe((state, previous) => {
        if (state.installed && !previous.installed) {
          showNotification({ message: "Installed. Open it from your home screen.", type: "info" });
        }
      }),
    [showNotification],
  );
}

interface InstallAppBarProps {
  active: boolean;
}

const InstallAppBar: React.FC<InstallAppBarProps> = ({ active }) => {
  useInstalledNotice();
  const { isCompact } = useLayoutMode();
  const reducedMotion = useReducedMotion();
  const onLanding = useAppPhaseStore((s) => s.phase.kind === "menu" && s.phase.route === "landing");
  const { platform, install } = useInstallOffer();
  const [dismissed, setDismissed] = useState(readDismissed);
  const [delayElapsed, setDelayElapsed] = useState(false);

  const eligible = active && isCompact && onLanding && platform !== null && !dismissed;

  useEffect(() => {
    if (!eligible) {
      setDelayElapsed(false);
      return;
    }
    const timer = setTimeout(() => setDelayElapsed(true), SHOW_DELAY_MS);
    return () => clearTimeout(timer);
  }, [eligible]);

  if (!eligible || !delayElapsed) {
    return null;
  }

  const handleNotNow = () => {
    writeDismissed();
    setDismissed(true);
  };

  const handleInstall = async () => {
    const outcome = await install();
    if (outcome === "dismissed") {
      setDismissed(true);
    }
  };

  return (
    <div
      role="region"
      aria-label="Install app"
      className={`game-panel fixed inset-x-0 bottom-0 text-white ${reducedMotion ? "" : "animate-[sheetSlideUp_250ms_ease-out]"}`}
      style={
        {
          "--panel-cut": "10px",
          zIndex: Z_INDEX.INSTALL_APP_BAR,
          paddingBottom: "var(--safe-bottom)",
          paddingLeft: "calc(12px + var(--safe-left))",
          paddingRight: "calc(8px + var(--safe-right))",
        } as React.CSSProperties
      }
    >
      <div className="flex items-center gap-2 h-11">
        <span className="flex-1 min-w-0 truncate text-left text-sm text-white/90">
          Play full screen as an app
        </span>
        <GameButton emphasis="quiet" size="sm" height={40} onClick={handleNotNow}>
          Not now
        </GameButton>
        <GameButton size="sm" height={36} onClick={() => void handleInstall()}>
          {platform === "prompt" ? "Install" : "How?"}
        </GameButton>
      </div>
    </div>
  );
};

export default InstallAppBar;
