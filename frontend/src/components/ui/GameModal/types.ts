import React from "react";

export interface ModalTheme {
  accent: string;
}

export const MODAL_THEMES = {
  actions: { accent: "#ff6464" },
  effects: { accent: "#ff96ff" },
  victoryPoints: { accent: "#ffd700" },
  cardsPlayed: { accent: "#9664ff" },
  production: { accent: "#4a90d9" },
  corporation: { accent: "#6496ff" },
  default: { accent: "#6496ff" },
} as const;

export type ModalThemeName = keyof typeof MODAL_THEMES;
export type ModalSize = "small" | "medium" | "large" | "full";
export type ModalAnimation = "slideIn" | "fadeIn" | "none";

export interface GameModalProps {
  isVisible: boolean;
  onClose: () => void;
  theme: ModalThemeName | ModalTheme;
  size?: ModalSize;
  animation?: ModalAnimation;
  zIndex?: number;
  closeOnBackdrop?: boolean;
  closeOnEscape?: boolean;
  lockScroll?: boolean;
  preventClose?: boolean;
  onPreventedClose?: () => void;
  glow?: boolean;
  outerContent?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}

export interface GameModalHeaderProps {
  title: string;
  subtitle?: string;
  stats?: React.ReactNode;
  controls?: React.ReactNode;
  showCloseButton?: boolean;
  onClose?: () => void;
}

export interface GameModalContentProps {
  children: React.ReactNode;
  padding?: "none" | "normal" | "large";
  className?: string;
}

export interface GameModalFooterProps {
  children: React.ReactNode;
  className?: string;
}

export interface GameModalEmptyProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}
