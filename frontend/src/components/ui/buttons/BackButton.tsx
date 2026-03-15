import React from "react";
import GameButton from "./GameButton.tsx";

interface BackButtonProps {
  onClick: (e: React.MouseEvent) => void;
  className?: string;
  size?: "sm" | "md" | "lg";
  children?: React.ReactNode;
}

const BackButton: React.FC<BackButtonProps> = ({
  onClick,
  className = "",
  size = "sm",
  children = "Back",
}) => {
  return (
    <GameButton
      buttonType="secondary"
      size={size}
      onClick={onClick}
      className={`flex items-center gap-2 ${className}`}
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <polyline points="15 18 9 12 15 6" />
      </svg>
      {children}
    </GameButton>
  );
};

export default BackButton;
