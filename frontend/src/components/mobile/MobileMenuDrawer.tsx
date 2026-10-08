import { forwardRef, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useBackDismiss } from "@/hooks/useBackDismiss.ts";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import CloseButton from "../ui/buttons/CloseButton.tsx";
import GameButton from "../ui/buttons/GameButton.tsx";
import { HamburgerIcon } from "../ui/menuIcons.tsx";

interface MobileMenuButtonProps {
  isOpen: boolean;
  onClick: () => void;
  className?: string;
}

export const MobileMenuButton = forwardRef<HTMLButtonElement, MobileMenuButtonProps>(
  ({ isOpen, onClick, className = "" }, ref) => (
    <GameButton
      ref={ref}
      emphasis="quiet"
      aria-label="Menu"
      aria-expanded={isOpen}
      onClick={onClick}
      className={`!p-0 w-11 h-11 shrink-0 pointer-events-auto ${className}`}
    >
      <HamburgerIcon />
    </GameButton>
  ),
);

interface MobileMenuDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  children: ReactNode;
}

export default function MobileMenuDrawer({ isOpen, onClose, children }: MobileMenuDrawerProps) {
  if (!isOpen) {
    return null;
  }
  return createPortal(<DrawerPanel onClose={onClose}>{children}</DrawerPanel>, document.body);
}

function DrawerPanel({ onClose, children }: { onClose: () => void; children: ReactNode }) {
  const reducedMotion = useReducedMotion();
  useBackDismiss(onClose);

  return (
    <div
      data-overlay-layer
      className="fixed inset-0"
      style={{ zIndex: Z_INDEX.MOBILE_MENU_DRAWER }}
    >
      <div
        aria-hidden="true"
        className={`absolute inset-0 bg-black/60 ${reducedMotion ? "" : "animate-[fadeIn_180ms_ease-out]"}`}
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Menu"
        className={`absolute inset-y-0 left-0 w-[min(300px,85vw)] flex flex-col text-white bg-[rgba(3,3,4,0.97)] border-r border-white/15 ${reducedMotion ? "" : "animate-[drawerSlideIn_200ms_ease-out]"}`}
        style={{
          paddingTop: "var(--safe-top)",
          paddingBottom: "var(--safe-bottom)",
          paddingLeft: "var(--safe-left)",
        }}
      >
        <div className="flex items-center justify-between h-12 shrink-0 pl-4 pr-1 border-b border-white/15">
          <h2 className="m-0 font-orbitron text-sm font-bold tracking-wider uppercase text-shadow-glow">
            Menu
          </h2>
          <CloseButton onClick={onClose} label="Close menu" />
        </div>
        <div className="flex-1 min-h-0 overflow-y-auto overscroll-contain">{children}</div>
      </div>
    </div>
  );
}
