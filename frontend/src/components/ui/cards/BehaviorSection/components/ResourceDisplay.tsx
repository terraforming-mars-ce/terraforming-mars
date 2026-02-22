import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import { getIconPath, getTagIconPath } from "@/utils/iconStore.ts";
import BehaviorIcon from "./BehaviorIcon.tsx";

interface IconDisplayInfo {
  resourceType: string;
  amount: number;
  displayMode: "individual" | "number";
  iconCount: number;
}

interface TileScaleInfo {
  scale: 1 | 1.25 | 1.5 | 2;
  tileType: string | null;
}

interface ResourceDisplayProps {
  displayInfo: IconDisplayInfo;
  isInput?: boolean;
  resource?: any;
  isGroupedWithOtherNegatives?: boolean;
  context?: "standalone" | "action" | "production" | "default";
  isAffordable?: boolean;
  tileScaleInfo: TileScaleInfo;
}

const ResourceDisplay: React.FC<ResourceDisplayProps> = ({
  displayInfo,
  isInput = false,
  resource,
  isGroupedWithOtherNegatives = false,
  context = "default",
  isAffordable = true,
  tileScaleInfo,
}) => {
  const { resourceType, amount, displayMode } = displayInfo;

  const isCredits = resourceType === "credit" || resourceType === "credit-production";
  const isDiscount = resourceType === "discount";
  const isProduction = resourceType?.includes("-production");
  const hasPer = resource?.per;
  const isAttack = resource?.target === "any-player";

  // Handle production with per condition (e.g., 1 plant production per plant tag)
  if (isProduction && hasPer) {
    const baseResourceType = resourceType.replace("-production", "");

    let perIcon = null;
    if (hasPer.tag) {
      perIcon = getTagIconPath(hasPer.tag);
    } else if (hasPer.type) {
      perIcon = getIconPath(hasPer.type);
    }

    if (perIcon) {
      // Special handling for credits-production - use GameIcon with amount inside
      if (baseResourceType === "credit") {
        const itemClasses = !isAffordable
          ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
          : "flex items-center gap-px relative";

        return (
          <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
            <div className={itemClasses}>
              <GameIcon iconType="credit" amount={Math.abs(amount)} size="small" />
            </div>
            <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              /
            </span>
            <img
              src={perIcon}
              alt={hasPer.tag || hasPer.type}
              className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                hasPer.target === "any-player"
                  ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                  : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
              }`}
            />
          </div>
        );
      } else {
        // For other resources, use regular icon with amount overlay
        const productionIcon = (
          <BehaviorIcon
            resourceType={baseResourceType}
            isProduction={false}
            isAttack={isAttack}
            context="production"
            isAffordable={isAffordable}
            tileScaleInfo={tileScaleInfo}
          />
        );
        if (productionIcon) {
          return (
            <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
              <div className="flex items-center gap-px relative">
                {amount > 1 && (
                  <span className="text-[20px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none flex items-center ml-0.5 max-md:text-xs">
                    {amount}
                  </span>
                )}
                {productionIcon}
              </div>
              <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
              <img
                src={perIcon}
                alt={hasPer.tag || hasPer.type}
                className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                  hasPer.target === "any-player"
                    ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                    : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                }`}
              />
            </div>
          );
        }
      }
    }
  }

  // Handle production WITHOUT per condition in ACTION context only (e.g., energy-production input in Equatorial Magnetizer)
  // Other contexts (standalone, production) already have parent components that wrap in brown boxes
  if (isProduction && !hasPer && context === "action") {
    const baseResourceType = resourceType.replace("-production", "");

    const itemClasses = !isAffordable
      ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
      : "flex items-center gap-px relative";

    return (
      <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.7)_0%,rgba(139,89,42,0.65)_100%)] border border-[rgba(160,110,60,0.7)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
        <div className={itemClasses}>
          {amount > 1 && (
            <span className="text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
              {amount}
            </span>
          )}
          <BehaviorIcon
            resourceType={baseResourceType}
            isProduction={false}
            isAttack={isAttack}
            context="production"
            isAffordable={isAffordable}
            tileScaleInfo={tileScaleInfo}
          />
        </div>
      </div>
    );
  }

  // Handle regular resources with per condition (e.g., 1 floater per jovian tag)
  if (!isProduction && hasPer) {
    let perIcon = null;
    if (hasPer.tag) {
      perIcon = getTagIconPath(hasPer.tag);
    } else if (hasPer.type) {
      perIcon = getIconPath(hasPer.type);
    }

    if (perIcon) {
      // Special handling for credits - use GameIcon with amount inside
      if (isCredits) {
        const itemClasses = !isAffordable
          ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
          : "flex items-center gap-px relative";

        return (
          <div className="flex items-center gap-[3px]">
            <div className={itemClasses}>
              <GameIcon iconType="credit" amount={Math.abs(amount)} size="small" />
            </div>
            <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              /
            </span>
            <img
              src={perIcon}
              alt={hasPer.tag || hasPer.type}
              className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                hasPer.target === "any-player"
                  ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                  : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
              }`}
            />
          </div>
        );
      } else {
        // For other resources, use BehaviorIcon with amount overlay
        const iconElement = (
          <BehaviorIcon
            resourceType={resourceType}
            isProduction={false}
            isAttack={isAttack}
            context={context}
            isAffordable={isAffordable}
            tileScaleInfo={tileScaleInfo}
          />
        );

        if (iconElement) {
          return (
            <div className="flex items-center gap-[3px]">
              <div className="flex items-center gap-px relative">
                {amount > 1 && (
                  <span className="text-[20px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none flex items-center ml-0.5 max-md:text-xs">
                    {amount}
                  </span>
                )}
                {iconElement}
              </div>
              <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
              <img
                src={perIcon}
                alt={hasPer.tag || hasPer.type}
                className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                  hasPer.target === "any-player"
                    ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                    : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                }`}
              />
            </div>
          );
        }
      }
    }
  }

  if (isCredits) {
    const creditsClasses = `flex items-center gap-0.5 relative ${isAttack ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulseCredits_2s_ease-in-out_infinite]" : ""}`;

    // Show minus inside icon if not grouped with other negative resources
    const showMinusInside = amount < 0 && !isGroupedWithOtherNegatives;
    // Never show minus outside - the group minus is handled at the row level

    const finalCreditsClasses = !isAffordable
      ? `${creditsClasses} opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`
      : creditsClasses;

    return (
      <div className={finalCreditsClasses}>
        <GameIcon
          iconType="credit"
          amount={showMinusInside ? amount : Math.abs(amount)}
          size="small"
        />
      </div>
    );
  }

  if (isDiscount) {
    const discountClasses = !isAffordable
      ? "flex items-center gap-0.5 relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
      : "flex items-center gap-0.5 relative";

    return (
      <div className={discountClasses}>
        <GameIcon iconType="credit" amount={-amount} size="small" />
      </div>
    );
  }

  // Handle global-parameter-lenience (special edge case)
  if (resourceType === "global-parameter-lenience") {
    return (
      <div className="flex items-center gap-[2px]">
        <BehaviorIcon
          resourceType="temperature"
          isProduction={false}
          isAttack={false}
          context={context}
          isAffordable={isAffordable}
          tileScaleInfo={tileScaleInfo}
        />
        <BehaviorIcon
          resourceType="oxygen"
          isProduction={false}
          isAttack={false}
          context={context}
          isAffordable={isAffordable}
          tileScaleInfo={tileScaleInfo}
        />
        <BehaviorIcon
          resourceType="ocean-tile"
          isProduction={false}
          isAttack={false}
          context={context}
          isAffordable={isAffordable}
          tileScaleInfo={tileScaleInfo}
        />
        <span className="text-base font-bold text-white mx-1 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
          :
        </span>
        <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
          +/- {amount}
        </span>
      </div>
    );
  }

  // Use the passed context or determine based on production status
  let iconContext = context;
  if (iconContext === "default" && isProduction) {
    iconContext = "production";
  }

  // Check if this is a tile placement with restrictions (show asterisk)
  const isTilePlacement =
    resourceType === "city-placement" ||
    resourceType === "greenery-placement" ||
    resourceType === "ocean-placement" ||
    resourceType === "volcano-placement" ||
    resourceType === "land-claim";

  const hasTileRestrictions =
    isTilePlacement &&
    resource?.tileRestrictions &&
    (resource.tileRestrictions.adjacency ||
      resource.tileRestrictions.onTileType ||
      (resource.tileRestrictions.boardTags?.length ?? 0) > 0);

  const baseIconElement = (
    <BehaviorIcon
      resourceType={resourceType}
      isProduction={false}
      isAttack={isAttack}
      context={iconContext}
      isAffordable={isAffordable}
      tileScaleInfo={tileScaleInfo}
    />
  );

  if (!baseIconElement) {
    return (
      <span className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
        {isInput && "-"}
        {amount} {resourceType}
      </span>
    );
  }

  if (displayMode === "individual") {
    const absoluteAmount = Math.abs(amount);
    return (
      <div className="flex items-center gap-px relative">
        {(isInput || amount < 0) && !isGroupedWithOtherNegatives && context !== "action" && (
          <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[24px] flex items-center justify-center leading-none [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
            -
          </span>
        )}
        {Array.from({ length: absoluteAmount }, (_, i) => (
          <React.Fragment key={i}>{baseIconElement}</React.Fragment>
        ))}
        {hasTileRestrictions && (
          <span className="text-white font-bold text-sm ml-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            *
          </span>
        )}
      </div>
    );
  } else {
    return (
      <div className="flex items-center gap-0.5 relative">
        {isInput && !isGroupedWithOtherNegatives && context !== "action" && (
          <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[24px] flex items-center justify-center leading-none [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
            -
          </span>
        )}
        <span className="text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] w-[20px] flex items-center justify-center">
          {isGroupedWithOtherNegatives ? Math.abs(amount) : amount}
        </span>
        {baseIconElement}
        {hasTileRestrictions && (
          <span className="text-white font-bold text-sm ml-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            *
          </span>
        )}
      </div>
    );
  }
};

export default ResourceDisplay;
