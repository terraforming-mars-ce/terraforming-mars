import React, { forwardRef } from "react";
import { Link } from "react-router-dom";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

type ButtonType = "textonly" | "primary" | "secondary";
type ButtonVariant = "info" | "success" | "warn" | "error";
type ButtonSize = "xs" | "sm" | "md" | "lg";

const variantColors: Record<
  ButtonVariant,
  { bg: string; bgHover: string; border: string; borderHover: string; text: string }
> = {
  info: {
    bg: "bg-space-blue-600",
    bgHover: "hover:bg-space-blue-500",
    border: "border-space-blue-500",
    borderHover: "hover:border-space-blue-400",
    text: "text-space-blue-400",
  },
  success: {
    bg: "bg-green-600",
    bgHover: "hover:bg-green-500",
    border: "border-green-500",
    borderHover: "hover:border-green-400",
    text: "text-green-400",
  },
  warn: {
    bg: "bg-yellow-700",
    bgHover: "hover:bg-yellow-600",
    border: "border-yellow-600",
    borderHover: "hover:border-yellow-500",
    text: "text-yellow-400",
  },
  error: {
    bg: "bg-red-700",
    bgHover: "hover:bg-red-600",
    border: "border-red-600",
    borderHover: "hover:border-red-500",
    text: "text-red-400",
  },
};

function getTypeStyles(buttonType: ButtonType, variant: ButtonVariant): string {
  const c = variantColors[variant];

  switch (buttonType) {
    case "textonly":
      return [
        "bg-transparent border-none rounded-lg",
        "text-white/70",
        "transition-all duration-200 cursor-pointer",
        `hover:${c.text.replace("text-", "text-")} hover:brightness-125`,
        "disabled:opacity-50 disabled:cursor-default disabled:hover:brightness-100",
      ].join(" ");
    case "primary":
      return [
        `${c.bg} border-2 border-transparent rounded-lg`,
        "font-orbitron font-semibold text-white",
        "transition-all duration-200 cursor-pointer",
        "hover:opacity-85",
        "disabled:opacity-50 disabled:cursor-default disabled:hover:opacity-50",
      ].join(" ");
    case "secondary":
      return [
        `bg-space-black-darker/90 border-2 ${c.border} rounded-lg`,
        "font-orbitron font-semibold text-white",
        "transition-all duration-200 backdrop-blur-space cursor-pointer",
        `${c.borderHover} hover:shadow-[0_0_12px_rgba(255,255,255,0.15)]`,
        "disabled:opacity-50 disabled:cursor-default disabled:hover:shadow-none",
      ].join(" ");
  }
}

const sizeStyles: Record<ButtonSize, string> = {
  xs: "py-0.5 px-2 text-xs",
  sm: "py-1.5 px-3 text-sm",
  md: "py-2 px-4 text-sm",
  lg: "py-3 px-6 text-lg",
};

interface GameButtonProps {
  buttonType?: ButtonType;
  variant?: ButtonVariant;
  size?: ButtonSize;
  children: React.ReactNode;
  className?: string;
  disabled?: boolean;
  onClick?: (e: React.MouseEvent) => void;
  type?: "button" | "submit";
  title?: string;
  as?: "button" | "link";
  to?: string;
  linkOnClick?: (e: React.MouseEvent<HTMLAnchorElement>) => void;
  style?: React.CSSProperties;
  onMouseEnter?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  onMouseLeave?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  "aria-label"?: string;
}

const GameButton = forwardRef<HTMLButtonElement, GameButtonProps>(
  (
    {
      buttonType = "primary",
      variant = "info",
      size = "md",
      children,
      className = "",
      disabled,
      onClick,
      type = "button",
      title,
      as = "button",
      to,
      linkOnClick,
      style,
      onMouseEnter,
      onMouseLeave,
      "aria-label": ariaLabel,
    },
    ref,
  ) => {
    const { playButtonHoverSound, playButtonClickSound } = useSoundEffects();
    const classes = `${getTypeStyles(buttonType, variant)} ${sizeStyles[size]} ${className}`.trim();

    const handleMouseEnter = (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => {
      if (!disabled) {
        void playButtonHoverSound();
      }
      onMouseEnter?.(e as React.MouseEvent<HTMLButtonElement>);
    };

    const handleClick = (e: React.MouseEvent) => {
      if (!disabled) {
        void playButtonClickSound();
      }
      onClick?.(e);
    };

    const handleLinkClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
      void playButtonClickSound();
      linkOnClick?.(e);
    };

    if (as === "link" && to) {
      return (
        <Link
          to={to}
          onClick={handleLinkClick}
          className={`${classes} no-underline inline-block`}
          onMouseEnter={handleMouseEnter}
        >
          {children}
        </Link>
      );
    }

    return (
      <button
        ref={ref}
        type={type}
        onClick={handleClick}
        disabled={disabled}
        className={classes}
        title={title}
        style={style}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={onMouseLeave}
        aria-label={ariaLabel}
      >
        {children}
      </button>
    );
  },
);

export default GameButton;
