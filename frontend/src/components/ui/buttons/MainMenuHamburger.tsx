import React, { useState, useRef, useEffect, useCallback } from "react";
import SoundToggleButton from "./SoundToggleButton.tsx";
import GameButton from "./GameButton.tsx";
import { GamePopover } from "../GamePopover";
import { MenuPopoverItem, MenuPopoverDivider, MenuPopoverVersion } from "../MenuPopoverItem.tsx";
import {
  CopyIcon,
  FullscreenIcon,
  ExitFullscreenIcon,
  BugIcon,
  LeaveIcon,
  EndGameIcon,
} from "../menuIcons.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface MainMenuHamburgerProps {
  gameId?: string;
  onLeaveGame?: () => void;
  onEndGame?: () => void;
}

const MainMenuHamburger: React.FC<MainMenuHamburgerProps> = ({
  gameId,
  onLeaveGame,
  onEndGame,
}) => {
  const [menuOpen, setMenuOpen] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(!!document.fullscreenElement);
  const buttonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  const handleToggleFullscreen = useCallback(() => {
    if (isFullscreen) {
      void document.exitFullscreen();
    } else {
      void document.documentElement.requestFullscreen();
    }
    setMenuOpen(false);
  }, [isFullscreen]);

  const handleReportBug = useCallback(() => {
    setMenuOpen(false);
    window.dispatchEvent(new CustomEvent("toggle-bug-report-window"));
  }, []);

  const handleCopyGameLink = useCallback(async () => {
    if (gameId) {
      const url = `${window.location.origin}/game/${gameId}`;
      await navigator.clipboard.writeText(url);
      setMenuOpen(false);
    }
  }, [gameId]);

  return (
    <div
      className="fixed top-[30px] right-[30px]"
      data-overlay-layer
      style={{ zIndex: Z_INDEX.POPOVER }}
    >
      <GameButton
        ref={buttonRef}
        buttonType="secondary"
        size="sm"
        onClick={() => setMenuOpen(!menuOpen)}
        className="p-2.5"
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
        >
          <line x1="3" y1="6" x2="21" y2="6" />
          <line x1="3" y1="12" x2="21" y2="12" />
          <line x1="3" y1="18" x2="21" y2="18" />
        </svg>
      </GameButton>

      <GamePopover
        isVisible={menuOpen}
        onClose={() => setMenuOpen(false)}
        position={{ type: "anchor", anchorRef: buttonRef, placement: "below" }}
        theme="menu"
        width={200}
        maxHeight="auto"
        animation="slideDown"
        excludeRef={buttonRef}
      >
        {gameId && (
          <>
            <MenuPopoverItem
              icon={<CopyIcon />}
              label="Copy game link"
              onClick={() => void handleCopyGameLink()}
            />
            <MenuPopoverDivider />
          </>
        )}
        <SoundToggleButton />
        <MenuPopoverDivider />
        <MenuPopoverItem
          icon={isFullscreen ? <ExitFullscreenIcon /> : <FullscreenIcon />}
          label={isFullscreen ? "Exit Fullscreen" : "Fullscreen"}
          onClick={handleToggleFullscreen}
        />
        <MenuPopoverDivider />
        <MenuPopoverItem icon={<BugIcon />} label="Report Bug" onClick={handleReportBug} />
        {onLeaveGame && (
          <>
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<LeaveIcon />}
              label="Leave game"
              variant="danger"
              onClick={() => {
                setMenuOpen(false);
                onLeaveGame();
              }}
            />
          </>
        )}
        {onEndGame && (
          <>
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<EndGameIcon />}
              label="End game"
              variant="danger"
              onClick={() => {
                setMenuOpen(false);
                onEndGame();
              }}
            />
          </>
        )}
        <MenuPopoverDivider />
        <MenuPopoverVersion />
      </GamePopover>
    </div>
  );
};

export default MainMenuHamburger;
