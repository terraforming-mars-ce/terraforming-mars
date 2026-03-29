import React, { useCallback, useEffect, useState } from "react";
import { GamePopover } from "./GamePopover";
import { MenuPopoverItem, MenuPopoverDivider, MenuPopoverVersion } from "./MenuPopoverItem.tsx";
import {
  CopyIcon,
  FullscreenIcon,
  ExitFullscreenIcon,
  PerformanceIcon,
  FeedbackIcon,
  LeaveIcon,
  EndGameIcon,
} from "./menuIcons.tsx";
import SoundToggleButton from "./buttons/SoundToggleButton.tsx";
import { useHoverSound } from "@/hooks/useHoverSound.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface GameHamburgerMenuProps {
  isOpen: boolean;
  onClose: () => void;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
  gameId?: string;
  isHost: boolean;
  onLeaveGame?: () => void;
  onEndGame?: () => void;
}

const GameHamburgerMenu: React.FC<GameHamburgerMenuProps> = ({
  isOpen,
  onClose,
  anchorRef,
  gameId,
  isHost,
  onLeaveGame,
  onEndGame,
}) => {
  const menuItemHover = useHoverSound();
  const [isFullscreen, setIsFullscreen] = useState(!!document.fullscreenElement);

  useEffect(() => {
    const handler = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", handler);
    return () => document.removeEventListener("fullscreenchange", handler);
  }, []);

  const handleToggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) {
      void document.exitFullscreen();
    } else {
      void document.documentElement.requestFullscreen();
    }
    onClose();
  }, [onClose]);

  const handleCopyGameLink = useCallback(async () => {
    if (!gameId) {
      return;
    }
    const url = `${window.location.origin}/join/${gameId}`;
    await navigator.clipboard.writeText(url);
    onClose();
  }, [gameId, onClose]);

  const handleLeaveGame = useCallback(() => {
    menuItemHover.onClick?.();
    onClose();
    onLeaveGame?.();
  }, [onLeaveGame, onClose, menuItemHover]);

  const handleEndGame = useCallback(() => {
    menuItemHover.onClick?.();
    onClose();
    onEndGame?.();
  }, [onEndGame, onClose, menuItemHover]);

  return (
    <GamePopover
      isVisible={isOpen}
      onClose={onClose}
      position={{
        type: "anchor",
        anchorRef: anchorRef,
        placement: "below",
      }}
      theme="menu"
      width={200}
      maxHeight="auto"
      animation="slideDown"
      excludeRef={anchorRef}
      zIndex={Z_INDEX.TOP_MENU_ALWAYS_ON_TOP + 1}
      overlayLayer
    >
      <div className="py-1">
        <MenuPopoverItem
          icon={<CopyIcon />}
          label="Copy game link"
          onClick={() => {
            menuItemHover.onClick?.();
            void handleCopyGameLink();
          }}
          onMouseEnter={menuItemHover.onMouseEnter}
        />
        <MenuPopoverDivider />
        <SoundToggleButton />
        <MenuPopoverDivider />
        <MenuPopoverItem
          icon={isFullscreen ? <ExitFullscreenIcon /> : <FullscreenIcon />}
          label={isFullscreen ? "Exit Fullscreen" : "Fullscreen"}
          onClick={handleToggleFullscreen}
        />
        <MenuPopoverDivider />
        <MenuPopoverItem
          icon={<PerformanceIcon />}
          label="Performance"
          onClick={() => {
            menuItemHover.onClick?.();
            onClose();
            window.dispatchEvent(new CustomEvent("toggle-performance-window"));
          }}
          onMouseEnter={menuItemHover.onMouseEnter}
        />
        <MenuPopoverDivider />
        <MenuPopoverItem
          icon={<FeedbackIcon />}
          label="Feedback"
          onClick={() => {
            menuItemHover.onClick?.();
            onClose();
            window.dispatchEvent(new CustomEvent("toggle-feedback-window"));
          }}
          onMouseEnter={menuItemHover.onMouseEnter}
        />
        <MenuPopoverDivider />
        <MenuPopoverItem
          icon={<LeaveIcon />}
          label="Leave game"
          variant="danger"
          onClick={handleLeaveGame}
          onMouseEnter={menuItemHover.onMouseEnter}
        />
        {isHost && onEndGame && (
          <>
            <MenuPopoverDivider />
            <MenuPopoverItem
              icon={<EndGameIcon />}
              label="End game"
              variant="danger"
              onClick={handleEndGame}
              onMouseEnter={menuItemHover.onMouseEnter}
            />
          </>
        )}
        <MenuPopoverDivider />
        <MenuPopoverVersion />
      </div>
    </GamePopover>
  );
};

export default GameHamburgerMenu;
