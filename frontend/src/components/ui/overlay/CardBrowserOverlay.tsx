import React, { useRef } from "react";
import CardBrowser from "../CardBrowser.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface CardBrowserOverlayProps {
  isVisible: boolean;
  onClose: () => void;
}

const CardBrowserOverlay: React.FC<CardBrowserOverlayProps> = ({ isVisible, onClose }) => {
  const scrollRef = useRef<HTMLDivElement>(null);

  if (!isVisible) {
    return null;
  }

  return (
    <div
      ref={scrollRef}
      className="fixed inset-0 overflow-y-auto bg-black"
      style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
    >
      <CardBrowser onBack={onClose} backLabel="Back to Game" scrollContainerRef={scrollRef} />
    </div>
  );
};

export default CardBrowserOverlay;
