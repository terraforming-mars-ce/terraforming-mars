import React from "react";
import { getIconPath } from "@/utils/iconStore.ts";

interface CardIconProps {
  amount: number;
  badgeType: "peek" | "take" | "buy" | "discard" | "none";
  isAffordable?: boolean;
  isAttack?: boolean;
  totalCardTypes?: number;
}

const cardActionBadges = {
  peek: {
    label: "Peek at card",
    glyph: (
      <>
        <path d="M1.5 8s2.5-4 6.5-4 6.5 4 6.5 4-2.5 4-6.5 4-6.5-4-6.5-4Z" />
        <circle cx="8" cy="8" r="1.5" />
      </>
    ),
  },
  take: {
    label: "Take card",
    glyph: <path d="M8 2v11m-4-4 4 4 4-4" />,
  },
  buy: {
    label: "Buy card",
    glyph: (
      <>
        <path d="M11.5 4.5C10 2.5 4 2.5 4 5.5s8 2 8 5-6 3-8 1" />
        <path d="M8 1.5v13" />
      </>
    ),
  },
  discard: {
    label: "Discard card",
    glyph: <path d="m4 4 8 8m0-8-8 8" />,
  },
};

const CardIcon: React.FC<CardIconProps> = ({
  amount,
  badgeType,
  isAffordable = true,
  isAttack = false,
  totalCardTypes = 1,
}) => {
  const cardIcon = getIconPath("card-draw");

  if (!cardIcon) {
    return null;
  }

  const badge = badgeType === "none" ? null : cardActionBadges[badgeType];
  const badgeOpacity = isAffordable ? "" : "opacity-40";

  const glowFilter = isAttack
    ? "drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))"
    : "drop-shadow(0_1px_3px_rgba(0,0,0,0.6))_drop-shadow(0_0_1px_rgba(255,248,220,0.6))_drop-shadow(0_0_2px_rgba(255,248,220,0.4))";

  const attackAnimation = isAttack ? " animate-[attackPulse_2s_ease-in-out_infinite]" : "";

  const iconClass = isAffordable
    ? `w-[26px] h-[26px] object-contain [filter:${glowFilter}]${attackAnimation} max-md:w-[22px] max-md:h-[22px]`
    : `w-[26px] h-[26px] object-contain opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]`;

  const renderSingleIcon = () => (
    <div className="relative inline-block" role="img" aria-label={badge?.label ?? "Card"}>
      <img src={cardIcon} alt="" className={iconClass} />
      {badge && (
        <svg
          className={`absolute -bottom-[2px] -right-[3px] w-4 h-4 text-[#fff8e7] pointer-events-none max-md:w-3.5 max-md:h-3.5 ${badgeOpacity}`}
          viewBox="0 0 16 16"
          fill="none"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
          focusable="false"
        >
          <g stroke="#090b10" strokeWidth="4.5">
            {badge.glyph}
          </g>
          <g stroke="currentColor" strokeWidth="1.75">
            {badge.glyph}
          </g>
        </svg>
      )}
    </div>
  );

  // Amount 1: Show single icon, no number
  if (amount === 1) {
    return <div className="flex items-center gap-0.5 relative">{renderSingleIcon()}</div>;
  }

  // Amount 2-3 with single card type: Show individual icons side by side, no number
  if (amount <= 3 && totalCardTypes === 1) {
    return (
      <div className="flex items-center gap-0.5 relative">
        {Array.from({ length: amount }, (_, i) => (
          <React.Fragment key={i}>{renderSingleIcon()}</React.Fragment>
        ))}
      </div>
    );
  }

  // Multiple card types or amount > 2: Show number + single icon
  return (
    <div className="flex items-center gap-0.5 relative">
      <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] mr-px">
        {amount}
      </span>
      {renderSingleIcon()}
    </div>
  );
};

export default CardIcon;
