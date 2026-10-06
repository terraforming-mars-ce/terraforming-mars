import React, {
  createContext,
  useContext,
  useLayoutEffect,
  useRef,
  useEffect,
  useState,
} from "react";
import { createPortal } from "react-dom";
import { FormattedDescription } from "../../../display/FormattedDescription.tsx";
import { ClassifiedBehavior } from "../types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

const BehaviorLayoutContext = createContext({
  compact: false,
  cardId: undefined as string | undefined,
});
export const useBehaviorLayout = () => useContext(BehaviorLayoutContext);

interface BehaviorContainerProps {
  classifiedBehavior: ClassifiedBehavior;
  index: number;
  description?: string;
  isHovered?: boolean;
  onHover?: (index: number | null) => void;
  noContainer?: boolean;
  cardId?: string;
  children: React.ReactNode;
}

const DescriptionPortal: React.FC<{
  description: string;
  anchorRef: React.RefObject<HTMLDivElement | null>;
}> = ({ description, anchorRef }) => {
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);

  useLayoutEffect(() => {
    const anchor = anchorRef.current;
    if (!anchor) {
      return;
    }
    const updatePosition = () => {
      const rect = anchor.getBoundingClientRect();
      let bottom = rect.bottom;
      for (let parent = anchor.parentElement; parent; parent = parent.parentElement) {
        if (getComputedStyle(parent).overflowY !== "visible") {
          bottom = Math.min(bottom, parent.getBoundingClientRect().bottom);
        }
      }
      setPos({ x: rect.left + rect.width / 2, y: bottom });
    };
    updatePosition();
    const observer = new ResizeObserver(updatePosition);
    observer.observe(anchor);
    const section = anchor.closest(".behavior-section");
    if (section) {
      observer.observe(section);
    }
    window.addEventListener("resize", updatePosition);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", updatePosition);
    };
  }, [anchorRef]);

  if (!pos) return null;

  return createPortal(
    <div
      role="tooltip"
      className="font-sans fixed w-[184px] max-md:w-[148px] -translate-x-1/2 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: pos.x, top: pos.y, zIndex: Z_INDEX.LOADING_OVERLAY }}
    >
      <div
        className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
        style={{
          clipPath:
            "polygon(0 0, calc(100% - 14px) 0, 100% 14px, 100% 100%, 14px 100%, 0 calc(100% - 14px))",
        }}
      >
        <FormattedDescription text={description} />
        <svg
          className="absolute top-0 right-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
        <svg
          className="absolute bottom-0 left-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
      </div>
    </div>,
    document.body,
  );
};

const BehaviorContainer: React.FC<BehaviorContainerProps> = ({
  classifiedBehavior,
  index,
  description,
  isHovered = false,
  onHover,
  noContainer = false,
  cardId,
  children,
}) => {
  const { type: rawType } = classifiedBehavior;
  const type = noContainer ? ("auto-no-background" as const) : rawType;
  const containerRef = useRef<HTMLDivElement>(null);
  const [compact, setCompact] = useState(false);

  useLayoutEffect(() => {
    const element = containerRef.current;
    const parent = element?.parentElement;
    if (!element || !parent) {
      return;
    }
    let previousWidth = parent.clientWidth;
    const measure = () => {
      const width = parent.clientWidth;
      if (width !== previousWidth) {
        previousWidth = width;
        if (compact) {
          setCompact(false);
          return;
        }
      }
      const bounds = element.getBoundingClientRect();
      const overflow = Array.from(element.querySelectorAll("img, [data-behavior-atom]")).some(
        (child) => {
          const rect = child.getBoundingClientRect();
          return rect.left < bounds.left - 1 || rect.right > bounds.right + 1;
        },
      );
      const overflowingRow = Array.from(element.querySelectorAll(".behavior-flow")).some(
        (row) => row.scrollWidth > row.clientWidth + 1,
      );
      if (element.scrollWidth > element.clientWidth + 1 || overflow || overflowingRow) {
        setCompact(true);
      }
    };
    const observer = new ResizeObserver(measure);
    observer.observe(parent);
    observer.observe(element);
    measure();
    return () => observer.disconnect();
  }, [children, compact]);

  const content = (
    <BehaviorLayoutContext.Provider value={{ compact, cardId }}>
      {children}
    </BehaviorLayoutContext.Provider>
  );

  const handleMouseEnter = () => onHover?.(index);
  const handleMouseLeave = () => onHover?.(null);

  useEffect(() => {
    if (!isHovered) return;
    const dismiss = () => onHover?.(null);
    window.addEventListener("scroll", dismiss, true);
    return () => window.removeEventListener("scroll", dismiss, true);
  }, [isHovered, onHover]);

  if (type === "auto-no-background") {
    return (
      <div
        ref={containerRef}
        key={index}
        className="behavior-fitting relative flex items-center justify-center my-px p-[3px] min-h-8 max-w-full min-w-0 font-orbitron max-md:p-px max-md:my-px"
        data-compact={compact}
        onMouseEnter={onHover ? handleMouseEnter : undefined}
        onMouseLeave={onHover ? handleMouseLeave : undefined}
      >
        {content}
        {isHovered && description && (
          <DescriptionPortal description={description} anchorRef={containerRef} />
        )}
      </div>
    );
  } else {
    const typeStyles = {
      "manual-action":
        "bg-[linear-gradient(135deg,rgba(33,150,243,0.35)_0%,rgba(25,118,210,0.3)_100%)] border-[rgba(33,150,243,0.5)] shadow-[0_2px_4px_rgba(33,150,243,0.3)]",
      "triggered-effect": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      discount: "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "payment-substitute": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "value-modifier": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      defense: "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "immediate-production":
        "bg-[linear-gradient(135deg,rgba(139,89,42,0.35)_0%,rgba(101,67,33,0.3)_100%)] border-[rgba(139,89,42,0.5)] shadow-[0_2px_4px_rgba(139,89,42,0.25)]",
      "immediate-effect": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
    };

    const widthClass =
      type === "manual-action" ||
      type === "triggered-effect" ||
      type === "discount" ||
      type === "payment-substitute" ||
      type === "value-modifier" ||
      type === "defense"
        ? "w-fit"
        : "w-[calc(100%-20px)]";

    return (
      <div
        ref={containerRef}
        key={index}
        className={`behavior-fitting relative px-2 py-1 min-h-8 my-px border border-white/10 backdrop-blur-[2px] flex items-center max-w-full min-w-0 font-orbitron ${widthClass} ${typeStyles[type] || ""} max-md:px-1.5 max-md:py-[3px] max-md:min-h-7 max-md:my-px`}
        data-compact={compact}
        onMouseEnter={onHover ? handleMouseEnter : undefined}
        onMouseLeave={onHover ? handleMouseLeave : undefined}
      >
        <div className="flex items-center gap-1.5 min-w-0 w-full justify-center max-md:gap-1">
          {content}
        </div>
        {isHovered && description && (
          <DescriptionPortal description={description} anchorRef={containerRef} />
        )}
      </div>
    );
  }
};

export default BehaviorContainer;
