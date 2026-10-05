import { useCardPlayFlowStore } from "@/stores/cardPlayFlowStore";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex";

type GameFlowType = "immediate" | "interactive" | "interactive-mandatory";

interface GameFlowPopoverProps {
  isVisible: boolean;
  onClose?: () => void;
  type?: GameFlowType;
  className?: string;
  outerClassName?: string;
  renderSiblings?: React.ReactNode;
  handleEscapeKey?: boolean;
  children: React.ReactNode;
}

export function GameFlowPopover({
  isVisible,
  onClose,
  type = "interactive",
  className = "",
  outerClassName = "",
  renderSiblings,
  handleEscapeKey = true,
  children,
}: GameFlowPopoverProps) {
  const host = useCardPlayFlowStore((state) =>
    state.playSession?.phase === "choosing" ? state.playPromptHost : null,
  );
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  const isDismissible = type === "interactive" || (type === "immediate" && !!onClose);

  useEffect(() => {
    if (isVisible) {
      setIsClosing(false);
    }
  }, [isVisible]);

  const requestClose = useCallback(() => {
    if (!onClose) {
      return;
    }
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onClose();
    }, 200);
  }, [onClose]);

  useEffect(() => {
    const preventScroll = (event: WheelEvent | TouchEvent) => {
      if (
        (event.target instanceof Element && event.target.closest("[data-card-play-stage]")) ||
        (popoverRef.current && popoverRef.current.contains(event.target as Node))
      ) {
        return;
      }
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      document.body.style.overflow = "hidden";
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      document.body.style.overflow = "";
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible]);

  useEffect(() => {
    if (!isVisible || !isDismissible) {
      return;
    }

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        requestClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if ((event.target as HTMLElement).closest?.("[data-overlay-layer]")) {
        return;
      }
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        requestClose();
      }
    };

    if (type !== "immediate") {
      if (handleEscapeKey) {
        document.addEventListener("keydown", handleEscape);
      }
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, isDismissible, type, requestClose, handleEscapeKey]);

  if (!isVisible) {
    return null;
  }

  const animationClass = isClosing ? "animate-fadeOut" : "animate-popIn";

  const content = (
    <div
      ref={popoverRef}
      className={`relative pointer-events-auto ${animationClass}`}
      style={host ? { maxWidth: "100%" } : undefined}
    >
      <div
        className={`min-w-[240px] w-fit max-w-[90vw] max-h-[500px] game-panel game-panel-clipped game-window flex flex-col overflow-hidden pointer-events-auto ${className} ${host ? "max-w-full! min-w-0!" : ""}`}
      >
        {children}
      </div>
      {renderSiblings}
    </div>
  );

  const rendered = (
    <>
      {type === "immediate" && (
        <div
          className="fixed inset-0 transition-opacity duration-[245ms] ease-in-out"
          style={{
            background: "rgba(0, 0, 0, 0.4)",
            backdropFilter: "blur(6px)",
            zIndex: Z_INDEX.IMMEDIATE_BACKDROP,
            opacity: isClosing ? 0 : 1,
            pointerEvents: isClosing ? "none" : "auto",
          }}
          onClick={onClose ? requestClose : undefined}
        />
      )}
      {host ? (
        createPortal(content, host)
      ) : (
        <div
          className={`
          fixed top-0 left-0 right-0 bottom-0
          flex items-center justify-center
          pointer-events-none overflow-hidden
          ${outerClassName}
        `}
          style={{
            zIndex:
              type === "immediate" ? Z_INDEX.IMMEDIATE_POPOVER : Z_INDEX.SELECTION_POPOVER + 1,
          }}
        >
          <div className="pointer-events-auto">{content}</div>
        </div>
      )}

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        @keyframes fadeOut {
          from {
            opacity: 1;
          }
          to {
            opacity: 0;
          }
        }

        @keyframes choiceSlideIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }

        .animate-choiceSlideIn {
          animation: choiceSlideIn 0.3s ease-out both;
        }

        @media (max-width: 768px) {
          .min-w-\\[240px\\] {
            min-width: 180px;
          }
          .max-w-\\[90vw\\] {
            max-width: 95vw;
          }
        }
      `}</style>
    </>
  );

  if (type === "immediate") {
    return createPortal(rendered, document.body);
  }

  return rendered;
}

interface GameFlowTitleProps {
  children: React.ReactNode;
  className?: string;
}

export function GameFlowTitle({ children, className = "" }: GameFlowTitleProps) {
  return (
    <div
      className={`shrink-0 py-[15px] px-5 bg-black/40 border-b border-b-white/15 select-none ${className}`}
    >
      {children}
    </div>
  );
}

interface GameFlowBodyProps {
  children: React.ReactNode;
  className?: string;
}

export function GameFlowBody({ children, className = "" }: GameFlowBodyProps) {
  return (
    <div
      className={`min-h-0 flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-white/20 scrollbar-track-white/5 ${className}`}
    >
      {children}
    </div>
  );
}

interface GameFlowFooterProps {
  children: React.ReactNode;
  className?: string;
}

export function GameFlowFooter({ children, className = "" }: GameFlowFooterProps) {
  return (
    <div
      className={`shrink-0 px-4 py-3 bg-black/40 border-t border-white/15 flex justify-center ${className}`}
    >
      {children}
    </div>
  );
}
