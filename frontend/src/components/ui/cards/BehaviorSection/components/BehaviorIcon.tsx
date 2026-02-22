import React from "react";
import { getIconPath, isTagIcon } from "@/utils/iconStore.ts";

interface TileScaleInfo {
  scale: 1 | 1.25 | 1.5 | 2;
  tileType: string | null;
}

interface BehaviorIconProps {
  resourceType: string;
  isProduction?: boolean;
  isAttack?: boolean;
  context?: "standalone" | "action" | "production" | "default";
  isAffordable?: boolean;
  tileScaleInfo: TileScaleInfo;
}

const BehaviorIcon: React.FC<BehaviorIconProps> = ({
  resourceType,
  isProduction: _isProduction = false,
  isAttack = false,
  context = "default",
  isAffordable = true,
  tileScaleInfo,
}) => {
  const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
  const icon = getIconPath(resourceType);

  if (!icon) return null;

  const isScaledTile = tileScaleInfo.scale > 1 && cleanType === tileScaleInfo.tileType;

  let iconClass: string;
  if (isScaledTile) {
    if (tileScaleInfo.scale === 2) {
      iconClass =
        "w-[52px] h-[52px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[44px] max-md:h-[44px]";
    } else if (tileScaleInfo.scale === 1.5) {
      iconClass =
        "w-[39px] h-[39px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[33px] max-md:h-[33px]";
    } else if (tileScaleInfo.scale === 1.25) {
      iconClass =
        "w-[33px] h-[33px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[28px] max-md:h-[28px]";
    } else {
      iconClass =
        "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]";
    }
  } else {
    iconClass =
      "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]";
  }

  const isTag = isTagIcon(cleanType);
  const isPlacement =
    cleanType === "city-placement" ||
    cleanType === "greenery-placement" ||
    cleanType === "ocean-placement" ||
    cleanType === "volcano-placement" ||
    cleanType === "land-claim";
  const isTR = cleanType === "tr";
  const isCard =
    cleanType === "card-draw" ||
    cleanType === "card-take" ||
    cleanType === "card-peek" ||
    cleanType === "card";

  const isStandaloneTile =
    cleanType === "city-tile" ||
    cleanType === "greenery-tile" ||
    cleanType === "ocean-tile" ||
    cleanType === "volcano-tile";
  const isStandaloneCard = cleanType === "card-draw" || cleanType === "card";
  const shouldUseStandaloneSize =
    context === "standalone" && (isStandaloneTile || isStandaloneCard);

  if (!isScaledTile) {
    if (isAttack) {
      iconClass =
        "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite] max-md:w-[22px] max-md:h-[22px]";
    } else if (shouldUseStandaloneSize) {
      const cardGlow = isStandaloneCard
        ? "_drop-shadow(0_0_1px_rgba(255,248,220,0.6))_drop-shadow(0_0_2px_rgba(255,248,220,0.4))"
        : "";
      iconClass = `w-9 h-9 object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.7))${cardGlow}] max-md:w-8 max-md:h-8`;
    } else if (isPlacement) {
      iconClass =
        "w-[30px] h-[30px] object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-md:w-[26px] max-md:h-[26px]";
    } else if (isTR) {
      iconClass =
        "w-8 h-8 object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-md:w-7 max-md:h-7";
    } else if (isCard) {
      iconClass =
        "w-[30px] h-[30px] object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))_drop-shadow(0_0_1px_rgba(255,248,220,0.6))_drop-shadow(0_0_2px_rgba(255,248,220,0.4))] max-md:w-[26px] max-md:h-[26px]";
    } else if (isTag) {
      iconClass =
        "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]";
    }
  }

  const finalIconClass = !isAffordable
    ? `${iconClass} opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`
    : iconClass;

  return <img src={icon} alt={cleanType} className={finalIconClass} />;
};

export default BehaviorIcon;
