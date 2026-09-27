import React, { forwardRef } from "react";
import { Link, type LinkProps } from "react-router-dom";
import { useHoverSound } from "@/hooks/useHoverSound.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

export const ANGLE_INDENT = 20;
export const BUTTON_SPACING = 6;
export type EdgeStyle = "slope-left" | "slope-right" | "flat";
type Emphasis = "primary" | "secondary" | "quiet";
type Tone = "info" | "success" | "warn" | "error";

interface Appearance {
  emphasis?: Emphasis;
  tone?: Tone;
  size?: "xs" | "sm" | "md" | "lg";
  shape?: "cut" | "toolbar";
  selected?: boolean;
  loading?: boolean;
  accent?: string;
  width?: number;
  height?: number;
  leftEdge?: EdgeStyle;
  rightEdge?: EdgeStyle;
}

type ButtonProps = Appearance &
  Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "title"> & {
    as?: "button";
  };
type NavigationProps = Appearance &
  Omit<LinkProps, "title"> & {
    as: "link";
    disabled?: boolean;
  };
export type GameButtonProps = ButtonProps | NavigationProps;

const tones: Record<Tone, string> = {
  info: "#142c58",
  success: "#4caa72",
  warn: "#b78a36",
  error: "#b94f56",
};

export function buttonGeometry(left: EdgeStyle, right: EdgeStyle): string {
  const tl = left === "slope-left" ? "var(--button-angle)" : "0px";
  const bl = left === "slope-right" ? "var(--button-angle)" : "0px";
  const tr = right === "slope-left" ? "calc(100% - var(--button-angle))" : "100%";
  const br = right === "slope-right" ? "calc(100% - var(--button-angle))" : "100%";
  return `polygon(${tl} 0, ${tr} 0, ${br} 100%, ${bl} 100%)`;
}

const GameButton = forwardRef<HTMLButtonElement | HTMLAnchorElement, GameButtonProps>(
  (props, ref) => {
    const {
      emphasis = "primary",
      tone = "info",
      size = "md",
      shape = "cut",
      selected = false,
      loading = false,
      accent,
      width,
      height,
      leftEdge = "flat",
      rightEdge = "flat",
      className = "",
      style,
      children,
      disabled,
      ...rest
    } = props;
    const inactive = disabled || loading;
    const sound = useHoverSound(inactive);
    const surfaceStyle = {
      "--control-accent": accent ?? tones[tone],
      ...(shape === "toolbar" ? { "--control-shape": buttonGeometry(leftEdge, rightEdge) } : {}),
      width,
      height,
      ...style,
    } as React.CSSProperties;
    const shared = {
      className: `game-button ${className}`,
      style: surfaceStyle,
      "data-emphasis": emphasis,
      "data-size": size,
      "data-shape": shape,
      "data-selected": selected || undefined,
      "aria-busy": loading || undefined,
    };
    const content = (
      <>
        {emphasis !== "quiet" && <span className="game-button-frame" aria-hidden="true" />}
        <span className="game-button-label" style={{ zIndex: Z_INDEX.GAME_BOARD_BASE }}>
          {children}
        </span>
      </>
    );
    if (rest.as === "link") {
      const { as: _, onClick, onMouseEnter, ...link } = rest;
      return (
        <Link
          {...link}
          {...shared}
          ref={ref as React.Ref<HTMLAnchorElement>}
          aria-disabled={inactive || undefined}
          tabIndex={inactive ? -1 : link.tabIndex}
          onClick={(event) => {
            if (inactive) {
              event.preventDefault();
              return;
            }
            sound.onClick?.();
            onClick?.(event);
          }}
          onMouseEnter={(event) => {
            sound.onMouseEnter?.();
            onMouseEnter?.(event);
          }}
        >
          {content}
        </Link>
      );
    }
    const { as: _, onClick, onMouseEnter, type = "button", ...button } = rest;
    return (
      <button
        {...button}
        {...shared}
        ref={ref as React.Ref<HTMLButtonElement>}
        type={type}
        disabled={inactive}
        onClick={(event) => {
          if (!inactive) {
            sound.onClick?.();
            onClick?.(event);
          }
        }}
        onMouseEnter={(event) => {
          sound.onMouseEnter?.();
          onMouseEnter?.(event);
        }}
      >
        {content}
      </button>
    );
  },
);

export default GameButton as {
  (props: ButtonProps & React.RefAttributes<HTMLButtonElement>): React.ReactElement;
  (props: NavigationProps & React.RefAttributes<HTMLAnchorElement>): React.ReactElement;
};
