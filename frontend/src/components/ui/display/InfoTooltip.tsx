import { type FC, type ReactNode, useId, useRef, useState } from "react";
import DecorBoxTooltip from "./DecorBoxTooltip.tsx";

interface InfoTooltipProps {
  children: ReactNode;
  size?: "small" | "medium";
}

const InfoTooltip: FC<InfoTooltipProps> = ({ children, size = "medium" }) => {
  const id = useId();
  const anchor = useRef<HTMLSpanElement>(null);
  const [position, setPosition] = useState<{ x: number; y: number } | null>(null);
  const show = () => {
    const rect = anchor.current?.getBoundingClientRect();
    if (rect) {
      setPosition({ x: rect.left + rect.width / 2, y: rect.top - 6 });
    }
  };

  return (
    <span className="inline-flex align-middle">
      <span
        ref={anchor}
        tabIndex={0}
        aria-label="More information"
        aria-describedby={position ? id : undefined}
        className={`inline-flex shrink-0 items-center justify-center cursor-default text-white/50 hover:text-white focus-visible:text-white focus-visible:outline focus-visible:outline-1 focus-visible:outline-white/60 ${size === "small" ? "size-4" : "size-5"}`}
        onMouseEnter={show}
        onMouseLeave={() => {
          if (document.activeElement !== anchor.current) {
            setPosition(null);
          }
        }}
        onFocus={show}
        onBlur={() => setPosition(null)}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            event.stopPropagation();
            setPosition(null);
          }
        }}
      >
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <path d="M1.5 1.5h10l3 3v10h-10l-3-3z" stroke="currentColor" />
          <path d="M8 7v4M8 4.5v.5" stroke="currentColor" strokeWidth="1.5" />
        </svg>
      </span>
      <DecorBoxTooltip
        id={id}
        position={position}
        placement="above"
        maxWidth={size === "small" ? 260 : 280}
        cornerSize={8}
      >
        <span className="font-sans text-xs font-normal leading-relaxed">{children}</span>
      </DecorBoxTooltip>
    </span>
  );
};

export default InfoTooltip;
