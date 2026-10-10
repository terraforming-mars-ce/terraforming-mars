import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React from "react";
import { APP_VERSION } from "@/config.ts";
import { gatewayServer } from "@/utils/gateway.ts";
import { availableUpdate, useVersionStore } from "@/stores/versionStore.ts";
import { displayVersion } from "@/utils/version.ts";
import { UpdateIcon } from "./menuIcons.tsx";

interface MenuPopoverItemProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  variant?: "default" | "danger";
  onMouseEnter?: () => void;
}

export const MenuPopoverItem: React.FC<MenuPopoverItemProps> = ({
  icon,
  label,
  onClick,
  variant = "default",
  onMouseEnter,
}) => {
  const textColor = variant === "danger" ? "text-red-400" : "text-white";
  return (
    <GameButton
      emphasis="quiet"
      onClick={onClick}
      onMouseEnter={onMouseEnter}
      className={`w-full flex items-center justify-start gap-3 px-4 py-3 compact:min-h-12 ${textColor} text-sm hover:bg-white/10 transition-colors text-left`}
    >
      <span className="inline-flex w-5 shrink-0 items-center justify-center">{icon}</span>
      <span>{label}</span>
    </GameButton>
  );
};

export const MenuPopoverDivider: React.FC = () => <div className="border-t border-[#333]" />;

export const MenuPopoverVersion: React.FC = () => {
  const update = useVersionStore(availableUpdate);
  return (
    <>
      {update && (
        <MenuPopoverItem
          icon={<UpdateIcon />}
          label={`Update to ${displayVersion(update)}`}
          onClick={() => window.location.reload()}
        />
      )}
      <div className="px-4 py-2 text-white/25 text-xs text-center select-none">
        {gatewayServer ? `${gatewayServer.name} · ${APP_VERSION}` : APP_VERSION}
      </div>
    </>
  );
};
