import React, { useState } from "react";
import GameButton from "./GameButton.tsx";

interface CopyLinkButtonProps {
  textToCopy: string;
  defaultText: string;
  copiedText?: string;
  className?: string;
  icon?: React.ReactNode;
  onCopySuccess?: () => void;
  onCopyError?: (error: Error) => void;
}

const CopyLinkButton: React.FC<CopyLinkButtonProps> = ({
  textToCopy,
  defaultText,
  copiedText = "Copied!",
  className = "",
  icon,
  onCopySuccess,
  onCopyError,
}) => {
  const [isCopied, setIsCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(textToCopy);
      setIsCopied(true);
      onCopySuccess?.();

      // Fade back to default text after 1 second
      setTimeout(() => {
        setIsCopied(false);
      }, 1000);
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      onCopyError?.(error as Error);
    }
  };

  return (
    <GameButton
      buttonType="secondary"
      size="md"
      onClick={handleCopy}
      disabled={isCopied}
      className={`min-w-[120px] ${className}`}
    >
      <span
        className={`inline-flex items-center gap-2 transition-opacity duration-300 ${isCopied ? "opacity-70" : "opacity-100"}`}
      >
        {isCopied ? copiedText : defaultText}
        {icon && !isCopied && icon}
      </span>
    </GameButton>
  );
};

export default CopyLinkButton;
