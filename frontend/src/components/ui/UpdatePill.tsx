import React from "react";
import GameButton from "./buttons/GameButton.tsx";
import CloseButton from "./buttons/CloseButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { availableUpdate, showsUpdateNotice, useVersionStore } from "@/stores/versionStore.ts";
import { displayVersion } from "@/utils/version.ts";

interface UpdatePillProps {
  inMenuRoute: boolean;
}

function pillTop(inMenuRoute: boolean, isCompact: boolean): string {
  if (inMenuRoute) {
    return "calc(var(--menu-edge) + var(--safe-top))";
  }
  if (isCompact) {
    return "calc(var(--hud-top-h) + var(--safe-top) + 6px)";
  }
  return "68px";
}

const UpdatePill: React.FC<UpdatePillProps> = ({ inMenuRoute }) => {
  const { isCompact } = useLayoutMode();
  const update = useVersionStore(availableUpdate);
  const visible = useVersionStore(showsUpdateNotice);
  const openChangelog = useVersionStore((s) => s.openChangelog);
  const dismissUpdate = useVersionStore((s) => s.dismissUpdate);

  if (!visible || update === null) {
    return null;
  }

  return (
    <div
      role="status"
      className="game-panel fixed left-1/2 -translate-x-1/2 flex items-center gap-1 pl-4 pr-1 max-w-[calc(100vw-32px)] text-white"
      style={
        {
          "--panel-cut": "10px",
          top: pillTop(inMenuRoute, isCompact),
          zIndex: Z_INDEX.UPDATE_PILL,
        } as React.CSSProperties
      }
    >
      <span className="min-w-0 truncate font-orbitron text-sm">
        {displayVersion(update)} available
      </span>
      <GameButton emphasis="quiet" size="sm" height={44} onClick={() => openChangelog(update)}>
        What's new
      </GameButton>
      <GameButton size="sm" height={36} onClick={() => window.location.reload()}>
        Update
      </GameButton>
      <CloseButton label="Hide update notice" onClick={() => dismissUpdate(update)} />
    </div>
  );
};

export default UpdatePill;
