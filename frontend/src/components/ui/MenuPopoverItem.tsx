import React from "react";
import { APP_VERSION } from "@/config.ts";

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
    <button
      onClick={onClick}
      onMouseEnter={onMouseEnter}
      className={`w-full flex items-center gap-3 px-4 py-3 ${textColor} text-sm hover:bg-white/10 transition-colors text-left`}
    >
      {icon}
      {label}
    </button>
  );
};

export const MenuPopoverDivider: React.FC = () => <div className="border-t border-[#333]" />;

export const MenuPopoverVersion: React.FC = () => (
  <div className="px-4 py-2 text-white/25 text-xs text-center select-none">{APP_VERSION}</div>
);
