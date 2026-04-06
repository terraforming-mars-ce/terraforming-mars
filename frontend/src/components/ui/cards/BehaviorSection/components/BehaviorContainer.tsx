import React, { useRef, useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { FormattedDescription } from "../../../display/FormattedDescription.tsx";
import { ClassifiedBehavior } from "../types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface BehaviorContainerProps {
  classifiedBehavior: ClassifiedBehavior;
  index: number;
  description?: string;
  isHovered?: boolean;
  onHover?: (index: number | null) => void;
  noContainer?: boolean;
  children: React.ReactNode;
}

const DescriptionPortal: React.FC<{
  description: string;
  anchorRef: React.RefObject<HTMLDivElement | null>;
}> = ({ description, anchorRef }) => {
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);

  useEffect(() => {
    if (anchorRef.current) {
      const rect = anchorRef.current.getBoundingClientRect();
      setPos({ x: rect.left + rect.width / 2, y: rect.bottom });
    }
  }, [anchorRef]);

  if (!pos) return null;

  return createPortal(
    <div
      className="fixed w-[184px] max-md:w-[148px] -translate-x-1/2 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
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
  children,
}) => {
  const { type: rawType } = classifiedBehavior;
  const type = noContainer ? ("auto-no-background" as const) : rawType;
  const containerRef = useRef<HTMLDivElement>(null);

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
        className={`relative flex items-center justify-center my-px p-[3px] min-h-8 max-md:p-px max-md:my-px ${isHovered ? "z-10" : ""}`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {children}
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
      "storage-payment-substitute":
        "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
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
      type === "storage-payment-substitute" ||
      type === "value-modifier" ||
      type === "defense"
        ? "w-fit"
        : "w-[calc(100%-20px)]";

    return (
      <div
        ref={containerRef}
        key={index}
        className={`relative rounded-[3px] px-2 py-1 min-h-8 my-px border border-white/10 backdrop-blur-[2px] flex items-center ${widthClass} ${typeStyles[type] || ""} max-md:px-1.5 max-md:py-[3px] max-md:min-h-7 max-md:my-px ${isHovered ? "z-10" : ""}`}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        <div className="flex items-center gap-1.5 flex-nowrap w-full justify-center max-md:gap-1">
          {children}
        </div>
        {isHovered && description && (
          <DescriptionPortal description={description} anchorRef={containerRef} />
        )}
      </div>
    );
  }
};

export default BehaviorContainer;
