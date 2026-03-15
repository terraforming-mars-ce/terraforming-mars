import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { GamePopover } from "../GamePopover";
import GameButton from "../buttons/GameButton.tsx";

interface SpectatePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameId: string;
  anchorRef: React.RefObject<HTMLElement | null>;
}

const SpectatePopover: React.FC<SpectatePopoverProps> = ({
  isVisible,
  onClose,
  gameId,
  anchorRef,
}) => {
  const navigate = useNavigate();
  const [name, setName] = useState("");

  const handleSpectate = () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    navigate(`/game/${gameId}`, { state: { spectatorName: trimmed } });
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSpectate();
    }
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "below" }}
      theme="menu"
      header={{ title: "Spectate game", showCloseButton: true }}
      width={360}
      maxHeight="none"
      animation="slideDown"
    >
      <div className="flex flex-row gap-3 items-center px-4 py-4 min-w-0 overflow-hidden">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Enter your name"
          spellCheck={false}
          autoComplete="off"
          autoCorrect="off"
          maxLength={50}
          autoFocus
          className="flex-1 min-w-0 bg-black/50 border border-white/20 rounded-lg py-3 px-4 text-white text-base outline-none placeholder:text-white/50 focus:border-white/60 focus:shadow-[0_0_20px_rgba(255,255,255,0.1)] transition-all duration-200 disabled:opacity-60"
        />
        <GameButton size="md" onClick={handleSpectate} disabled={!name.trim()} className="shrink-0">
          Spectate
        </GameButton>
      </div>
    </GamePopover>
  );
};

export default SpectatePopover;
