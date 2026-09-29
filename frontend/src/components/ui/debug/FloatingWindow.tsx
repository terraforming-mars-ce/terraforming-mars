import type { CSSProperties, MouseEventHandler, ReactNode } from "react";
import CloseButton from "../buttons/CloseButton.tsx";

interface Props {
  title: string;
  subtitle?: string;
  onClose: () => void;
  onMouseDown: MouseEventHandler<HTMLDivElement>;
  style: CSSProperties;
  children: ReactNode;
  className?: string;
}

export default function FloatingWindow({
  title,
  subtitle,
  onClose,
  onMouseDown,
  style,
  children,
  className = "",
}: Props) {
  return (
    <section
      role="dialog"
      aria-label={title}
      data-overlay-layer
      className={`game-panel text-white text-left flex flex-col overflow-hidden ${className}`}
      onMouseDown={onMouseDown}
      style={{
        ...style,
        position: "fixed",
        maxWidth: "calc(100vw - 32px)",
        left: `clamp(16px, ${style.left}px, max(16px, calc(100vw - ${style.width}px - 16px)))`,
        top: `clamp(16px, ${style.top}px, calc(100dvh - 80px))`,
      }}
    >
      <header className="flex items-center justify-between gap-4 pl-5 pr-2 py-2 border-b border-white/15 select-none shrink-0">
        <div>
          <h3 className="font-orbitron text-sm m-0">{title}</h3>
          {subtitle && <p className="text-xs text-white/60 mt-1">{subtitle}</p>}
        </div>
        <CloseButton onClick={onClose} label={`Close ${title}`} />
      </header>
      <div
        className="flex flex-col min-h-0 overflow-auto p-4"
        onMouseDown={(event) => event.stopPropagation()}
      >
        {children}
      </div>
    </section>
  );
}
