import React, { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import { Z_INDEX } from "@/constants/zIndex";

type GameFlowType = "immediate" | "interactive" | "interactive-mandatory";

interface GameFlowPopoverContext {
  requestClose: () => void;
  onDragStart: (e: React.PointerEvent) => void;
  onDragMove: (e: React.PointerEvent) => void;
  onDragEnd: (e: React.PointerEvent) => void;
}

const GameFlowCtx = createContext<GameFlowPopoverContext | null>(null);

function useGameFlowContext() {
  const ctx = useContext(GameFlowCtx);
  if (!ctx) {
    throw new Error("GameFlow* components must be used within GameFlowPopover");
  }
  return ctx;
}

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
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const isDraggingRef = useRef(false);
  const dragStartRef = useRef({ x: 0, y: 0 });

  const isDismissible = type === "interactive" || (type === "immediate" && !!onClose);

  useEffect(() => {
    if (isVisible) {
      setDragOffset({ x: 0, y: 0 });
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
      if (popoverRef.current && popoverRef.current.contains(event.target as Node)) {
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

  const onDragStart = useCallback(
    (e: React.PointerEvent) => {
      isDraggingRef.current = true;
      dragStartRef.current = {
        x: e.clientX - dragOffset.x,
        y: e.clientY - dragOffset.y,
      };
      (e.target as HTMLElement).setPointerCapture(e.pointerId);
    },
    [dragOffset],
  );

  const onDragMove = useCallback((e: React.PointerEvent) => {
    if (!isDraggingRef.current) {
      return;
    }
    setDragOffset({
      x: e.clientX - dragStartRef.current.x,
      y: e.clientY - dragStartRef.current.y,
    });
  }, []);

  const onDragEnd = useCallback(() => {
    isDraggingRef.current = false;
  }, []);

  if (!isVisible) {
    return null;
  }

  const animationClass = isClosing ? "animate-fadeOut" : "animate-popIn";

  const content = renderSiblings ? (
    <div ref={popoverRef} className={`relative ${animationClass}`}>
      <div
        className={`
          min-w-[240px] w-fit max-w-[90vw] max-h-[500px]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          ${className}
        `}
      >
        {children}
      </div>
      {renderSiblings}
    </div>
  ) : (
    <div
      ref={popoverRef}
      className={`
        min-w-[240px] w-fit max-w-[90vw] max-h-[500px]
        bg-space-black-darker/95
        border-2 border-space-blue-500
        rounded-xl
        shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
        backdrop-blur-space
        flex flex-col overflow-hidden isolate
        pointer-events-auto
        ${animationClass}
        ${className}
      `}
    >
      {children}
    </div>
  );

  return (
    <GameFlowCtx.Provider value={{ requestClose, onDragStart, onDragMove, onDragEnd }}>
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
      <div
        className={`
          fixed top-0 left-0 right-0 bottom-0
          flex items-center justify-center
          pointer-events-none overflow-hidden
          ${outerClassName}
        `}
        style={{
          zIndex: type === "immediate" ? Z_INDEX.IMMEDIATE_POPOVER : Z_INDEX.SELECTION_POPOVER + 1,
        }}
      >
        <div
          className="pointer-events-auto"
          style={{
            transform: `translate(${dragOffset.x}px, ${dragOffset.y}px)`,
          }}
        >
          {content}
        </div>
      </div>

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.9) translateY(-20px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
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
            transform: translateX(-20px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
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
    </GameFlowCtx.Provider>
  );
}

interface GameFlowTitleProps {
  children: React.ReactNode;
  className?: string;
}

export function GameFlowTitle({ children, className = "" }: GameFlowTitleProps) {
  const { onDragStart, onDragMove, onDragEnd } = useGameFlowContext();

  return (
    <div
      className={`py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60 select-none ${className}`}
      onPointerDown={onDragStart}
      onPointerMove={onDragMove}
      onPointerUp={onDragEnd}
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
      className={`flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-space-blue-500/50 scrollbar-track-space-blue-900/30 ${className}`}
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
      className={`px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center ${className}`}
    >
      {children}
    </div>
  );
}
