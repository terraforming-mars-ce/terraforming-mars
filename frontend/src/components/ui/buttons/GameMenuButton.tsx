import React, { forwardRef } from "react";
import { Link } from "react-router-dom";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

const variantStyles = {
  primary: [
    "bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl",
    "font-orbitron font-semibold tracking-wide text-white",
    "transition-all duration-300 backdrop-blur-space cursor-pointer",
    "hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg",
    "disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:shadow-none",
  ].join(" "),
  secondary: [
    "bg-space-black-darker/80 border border-white/20 rounded-lg",
    "text-white text-sm",
    "transition-colors backdrop-blur-space cursor-pointer",
    "hover:bg-white/20",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ].join(" "),
  action: [
    "bg-space-blue-600 border border-space-blue-500 rounded-lg",
    "font-orbitron text-white text-sm font-medium",
    "transition-colors cursor-pointer",
    "hover:bg-space-blue-500",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ].join(" "),
  text: [
    "bg-transparent border-none rounded-lg",
    "text-white/70 text-sm",
    "transition-colors cursor-pointer",
    "hover:text-white",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ].join(" "),
  toolbar: [
    "bg-black border-2 border-white/20 rounded-xl",
    "font-orbitron font-bold text-white text-sm",
    "transition-all duration-200 cursor-pointer",
    "hover:bg-white/10",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ].join(" "),
  error: [
    "bg-red-900/90 border-2 border-red-700 rounded-xl",
    "font-orbitron font-semibold text-white",
    "transition-all duration-300 cursor-pointer",
    "hover:border-red-500 hover:shadow-[0_0_15px_rgba(220,38,38,0.4)]",
    "disabled:opacity-50 disabled:cursor-not-allowed",
  ].join(" "),
};

const sizeStyles = {
  sm: "py-2 px-4 text-sm",
  md: "py-3 px-5 text-sm",
  lg: "py-4 px-8 text-lg",
};

interface GameMenuButtonProps {
  variant?: "primary" | "secondary" | "action" | "text" | "toolbar" | "error";
  size?: "sm" | "md" | "lg";
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

const GameMenuButton = forwardRef<HTMLButtonElement, GameMenuButtonProps>(
  (
    {
      variant = "primary",
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
    const classes = `${variantStyles[variant]} ${sizeStyles[size]} ${className}`.trim();

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

export default GameMenuButton;
