import React from "react";

export interface PopoverTheme {
  accent: string; // hex color, e.g., "#ff6464"
}

export const POPOVER_THEMES = {
  actions: { accent: "#ff6464" }, // Red
  effects: { accent: "#ff96ff" }, // Magenta
  storages: { accent: "#6496ff" }, // Blue
  tags: { accent: "#64ff96" }, // Green
  standardProjects: { accent: "#4a90e2" }, // Blue
  awards: { accent: "#f39c12" }, // Orange
  milestones: { accent: "#ff6b35" }, // Orange-red
  log: { accent: "#64c8ff" }, // Cyan/light blue
  victoryPoints: { accent: "#ffc864" }, // Gold
  colonies: { accent: "#6b7280" }, // Neutral gray
  menu: { accent: "#9ca3af" }, // Neutral gray
} as const;

export type PopoverThemeName = keyof typeof POPOVER_THEMES;

export type PopoverPosition =
  | {
      type: "anchor";
      anchorRef: React.RefObject<HTMLElement | null>;
      placement: "above" | "below";
    }
  | {
      type: "fixed";
      top?: number;
      left?: number;
      right?: number;
      bottom?: number;
    };

export interface PopoverHeader {
  title: string;
  badge?: React.ReactNode;
  showCloseButton?: boolean;
  centerContent?: React.ReactNode;
  rightContent?: React.ReactNode;
}

export interface GamePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  position: PopoverPosition;
  theme: PopoverThemeName | PopoverTheme;
  header?: PopoverHeader;
  arrow?: {
    enabled: boolean;
    position?: "left" | "center" | "right";
    offset?: number;
  };
  width?: number | string;
  maxHeight?: number | string;
  zIndex?: number;
  animation?: "slideUp" | "slideDown";
  children: React.ReactNode;
  className?: string;
  excludeRef?: React.RefObject<HTMLElement | null>;
  contentRef?: React.RefObject<HTMLDivElement | null>;
  overlayLayer?: boolean;
}

export type PopoverItemState = "available" | "disabled" | "claimed";

export interface PopoverItemError {
  message: string;
  count?: number;
}

export interface PopoverItemWarning {
  message: string;
}

export interface GamePopoverItemProps {
  state: PopoverItemState;
  onClick?: () => void;
  error?: PopoverItemError;
  warning?: PopoverItemWarning;
  statusBadge?: string; // "played", "claimed", "funded"
  hoverEffect?: "translate-x" | "glow" | "background" | "none";
  animationDelay?: number;
  children: React.ReactNode;
  className?: string;
}

export interface GamePopoverEmptyProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}
