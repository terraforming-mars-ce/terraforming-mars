import {
  useId,
  useLayoutEffect,
  useState,
  type CSSProperties,
  type ReactNode,
  type RefObject,
} from "react";
import DecorBoxTooltip from "./DecorBoxTooltip.tsx";
import { useTapReveal } from "@/hooks/useTapReveal.ts";

const ABOVE_GAP_PX = 6;

interface RevealTriggerProps {
  content: ReactNode;
  children?: ReactNode;
  as?: "span" | "div" | "button";
  className?: string;
  style?: CSSProperties;
  placement?: "above" | "below";
  gap?: number;
  maxWidth?: number | string;
  cornerSize?: number;
  anchorRef?: RefObject<HTMLElement | null>;
  enabled?: boolean;
  tabIndex?: number;
  "aria-label"?: string;
}

interface TooltipAnchor {
  position: { x: number; y: number };
  flipY: number;
}

const RevealTrigger = ({
  content,
  children,
  as: Element = "span",
  className,
  style,
  placement = "above",
  gap,
  maxWidth,
  cornerSize,
  anchorRef,
  enabled = true,
  tabIndex,
  "aria-label": ariaLabel,
}: RevealTriggerProps) => {
  const id = useId();
  const hasContent =
    content !== null && content !== undefined && content !== false && content !== "";
  const { open, triggerRef, triggerProps } = useTapReveal<HTMLElement>({
    enabled: enabled && hasContent,
  });
  const [anchor, setAnchor] = useState<TooltipAnchor | null>(null);

  useLayoutEffect(() => {
    const element = anchorRef?.current ?? triggerRef.current;
    if (!open || !element) {
      setAnchor(null);
      return;
    }
    const rect = element.getBoundingClientRect();
    const x = rect.left + rect.width / 2;
    if (placement === "above") {
      const offset = gap ?? ABOVE_GAP_PX;
      setAnchor({ position: { x, y: rect.top - offset }, flipY: rect.bottom + offset });
    } else {
      const offset = gap ?? 0;
      setAnchor({ position: { x, y: rect.bottom + offset }, flipY: rect.top - offset });
    }
  }, [open, placement, gap, anchorRef, triggerRef]);

  return (
    <Element
      {...triggerProps}
      ref={triggerProps.ref as RefObject<HTMLSpanElement & HTMLDivElement & HTMLButtonElement>}
      type={Element === "button" ? "button" : undefined}
      className={className}
      style={style}
      tabIndex={tabIndex}
      aria-label={ariaLabel}
      aria-describedby={anchor ? id : undefined}
    >
      {children}
      <DecorBoxTooltip
        id={id}
        position={anchor?.position ?? null}
        flipY={anchor?.flipY}
        placement={placement}
        maxWidth={maxWidth}
        cornerSize={cornerSize}
        description={typeof content === "string" ? content : undefined}
      >
        {typeof content === "string" ? undefined : content}
      </DecorBoxTooltip>
    </Element>
  );
};

export default RevealTrigger;
