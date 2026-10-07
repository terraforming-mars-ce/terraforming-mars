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
  surface?: "console";
  selected?: boolean;
  loading?: boolean;
  clickSound?: boolean;
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

function buttonEdgeLight(edge: EdgeStyle, side: "left" | "right"): string {
  let start = "0px";
  let end = "0px";
  if (edge === "slope-left") {
    start = "var(--button-angle) * 0.65";
    end = "var(--button-angle) * 0.35";
  } else if (edge === "slope-right") {
    start = "var(--button-angle) * 0.35";
    end = "var(--button-angle) * 0.65";
  }
  const x = (inset: string, offset: number) => {
    if (side === "left") {
      return `calc(${inset} + ${offset}px)`;
    }
    return `calc(100% - (${inset}) - ${offset}px)`;
  };
  return `polygon(${x(start, 1)} 35%, ${x(start, 2)} 35%, ${x(end, 2)} 65%, ${x(end, 1)} 65%)`;
}

function buttonCornerLight(corner: "top-right" | "bottom-left"): string {
  const point = (fraction: number, inset: number) => {
    const x = `var(--control-cut) * ${fraction}`;
    const y = `var(--control-cut) * ${1 - fraction}`;
    if (corner === "top-right") {
      return `calc(100% - (${x}) - ${inset}px) calc(${y} + 1px)`;
    }
    return `calc(${x} + ${inset}px) calc(100% - (${y}) - 1px)`;
  };
  return `polygon(${point(0.75, 1)}, ${point(0.75, 2)}, ${point(0.25, 2)}, ${point(0.25, 1)})`;
}

const GameButton = forwardRef<HTMLButtonElement | HTMLAnchorElement, GameButtonProps>(
  (props, ref) => {
    const {
      emphasis = "primary",
      tone = "info",
      size = "md",
      shape = "cut",
      surface,
      selected = false,
      loading = false,
      clickSound = true,
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
    const defaultAccent = surface === "console" ? "var(--hud-accent)" : tones[tone];
    const surfaceStyle = {
      "--control-accent": accent ?? defaultAccent,
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
      "data-surface": surface,
      "data-selected": selected || undefined,
      "aria-busy": loading || undefined,
    };
    const content = (
      <>
        {(emphasis !== "quiet" || surface === "console") && (
          <span className="game-button-frame" aria-hidden="true" />
        )}
        {surface === "console" && (
          <span className="hud-control-lights" aria-hidden="true">
            <span className="hud-control-trace" />
            <span className="hud-control-availability" />
          </span>
        )}
        {surface === "console" && (
          <>
            <span
              className="hud-control-edge"
              style={{
                clipPath:
                  shape === "toolbar"
                    ? buttonEdgeLight(leftEdge, "left")
                    : buttonCornerLight("bottom-left"),
              }}
              aria-hidden="true"
            />
            <span
              className="hud-control-edge"
              style={{
                clipPath:
                  shape === "toolbar"
                    ? buttonEdgeLight(rightEdge, "right")
                    : buttonCornerLight("top-right"),
              }}
              aria-hidden="true"
            />
          </>
        )}
        <span className="game-button-label" style={{ zIndex: Z_INDEX.GAME_BOARD_BASE }}>
          {children}
        </span>
        {loading && (
          <span
            className="game-button-spinner"
            style={{ zIndex: Z_INDEX.GAME_BOARD_BASE }}
            aria-hidden="true"
          />
        )}
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
            if (clickSound) {
              sound.onClick?.();
            }
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
            if (clickSound) {
              sound.onClick?.();
            }
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
