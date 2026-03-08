import React from "react";

type CornerPosition = "bottom-right" | "top-left";

interface DecorBoxProps {
  corner?: CornerPosition;
  children: React.ReactNode;
  className?: string;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

const DecorBox: React.FC<DecorBoxProps> = ({
  corner = "bottom-right",
  children,
  className,
  onMouseEnter,
  onMouseLeave,
}) => {
  const isTopLeft = corner === "top-left";

  const clipPath = isTopLeft
    ? "polygon(8px 0, 100% 0, 100% 100%, 0 100%, 0 8px)"
    : "polygon(0 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%)";

  const borderClass = isTopLeft
    ? "border border-[rgba(60,60,70,0.7)] !border-b-0 !border-r-0"
    : "border border-[rgba(60,60,70,0.7)] !border-t-0";

  const outerClass = isTopLeft ? "relative w-fit" : "relative -mt-[2px] w-fit";

  return (
    <div className={outerClass}>
      <div
        className={`inline-flex items-center gap-1 px-1.5 py-px bg-[rgba(5,5,10,0.95)] ${borderClass} text-white font-orbitron ${className ?? ""}`}
        style={{ clipPath }}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        {children}
      </div>
      {isTopLeft ? (
        <svg className="absolute top-0 left-0 w-2 h-2 pointer-events-none" viewBox="0 0 8 8">
          <line x1="0" y1="8" x2="8" y2="0" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
      ) : (
        <svg className="absolute bottom-0 right-0 w-2 h-2 pointer-events-none" viewBox="0 0 8 8">
          <line x1="8" y1="0" x2="0" y2="8" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
      )}
    </div>
  );
};

export default DecorBox;
