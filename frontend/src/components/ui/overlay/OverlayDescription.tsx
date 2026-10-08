import { useLayoutEffect, useRef, useState, type ReactNode } from "react";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { OVERLAY_DESCRIPTION_CLASS } from "./overlayStyles.ts";

interface OverlayDescriptionProps {
  children: ReactNode;
  className?: string;
}

export default function OverlayDescription({ children, className = "" }: OverlayDescriptionProps) {
  const { isCompact } = useLayoutMode();
  const textRef = useRef<HTMLParagraphElement>(null);
  const [truncated, setTruncated] = useState(false);
  const [expanded, setExpanded] = useState(false);

  useLayoutEffect(() => {
    const element = textRef.current;
    if (!isCompact || !element) {
      return;
    }
    const update = () => {
      setTruncated(element.scrollWidth > element.clientWidth + 1);
    };
    update();
    const observer = new ResizeObserver(update);
    observer.observe(element);
    return () => observer.disconnect();
  }, [isCompact, children]);

  if (!isCompact) {
    return <p className={`${OVERLAY_DESCRIPTION_CLASS} ${className}`}>{children}</p>;
  }

  const expandable = truncated || expanded;
  const toggle = () => setExpanded((value) => !value);

  return (
    <div
      className={`min-w-0 -my-[14px] py-[14px] ${expandable ? "cursor-pointer" : ""}`}
      role={expandable ? "button" : undefined}
      tabIndex={expandable ? 0 : undefined}
      aria-expanded={expandable ? expanded : undefined}
      onClick={expandable ? toggle : undefined}
      onKeyDown={(event) => {
        if (expandable && (event.key === "Enter" || event.key === " ")) {
          event.preventDefault();
          toggle();
        }
      }}
    >
      <p
        ref={textRef}
        className={`${expanded ? OVERLAY_DESCRIPTION_CLASS.replace("compact:truncate", "") : OVERLAY_DESCRIPTION_CLASS} ${truncated ? "underline decoration-dotted decoration-white/40 underline-offset-2" : ""} ${className}`}
      >
        {children}
      </p>
    </div>
  );
}
