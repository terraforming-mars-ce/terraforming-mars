import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import { getIconPath, getTagIconPath } from "@/utils/iconStore.ts";
import BehaviorIcon from "./BehaviorIcon.tsx";
import Slash from "./Slash.tsx";
import { CalculatedOutputDto } from "@/types/generated/api-types.ts";

interface IconDisplayInfo {
  resourceType: string;
  amount: number;
  displayMode: "individual" | "number";
  iconCount: number;
  variableAmount?: boolean;
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
  computedAmount?: number;
  computedOutputs?: CalculatedOutputDto[];
}

const ComputedValueDisplay: React.FC<{
  amount: number;
  resourceType: string;
}> = ({ amount, resourceType }) => {
  const isCredits = resourceType === "credit" || resourceType === "credit-production";
  const isProduction = resourceType.includes("-production");
  const parenClasses = isProduction
    ? "text-[26px] font-normal text-white/70 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]"
    : "text-[22px] font-bold text-white/70 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]";

  return (
    <span className="flex items-center gap-[2px] opacity-80 ml-1">
      <span className={parenClasses}>(</span>
      {isCredits ? (
        <GameIcon iconType={resourceType} amount={amount} size="small" />
      ) : (
        <span className="flex items-center gap-[2px]">
          {amount > 1 && (
            <span className="text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
              {amount}
            </span>
          )}
          <GameIcon iconType={resourceType} size="small" />
        </span>
      )}
      <span className={parenClasses}>)</span>
    </span>
  );
};

const ResourceDisplay: React.FC<ResourceDisplayProps> = ({
  displayInfo,
  isInput = false,
  resource,
  isGroupedWithOtherNegatives = false,
  context = "default",
  isAffordable = true,
  tileScaleInfo,
  computedAmount,
  computedOutputs,
}) => {
  const { resourceType, amount, displayMode } = displayInfo;
  const isVariableAmount = !!displayInfo.variableAmount;

  const isCredits = resourceType === "credit" || resourceType === "credit-production";
  const isDiscount = resourceType === "discount";
  const isVP = resourceType === "vp";
  const isProduction = resourceType?.includes("-production");
  const hasPer = resource?.per;
  const isAttack =
    resource?.target === "any-player" ||
    resource?.target === "all-opponents" ||
    resource?.target?.startsWith("steal-");

  const resolvedComputedAmount =
    computedAmount ??
    (hasPer && computedOutputs
      ? computedOutputs.find((o) => o.resourceType === resourceType)?.amount
      : undefined);

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
      const perAmount = hasPer.amount ?? 1;
      const perIconEl = (
        <div className="flex items-center gap-px">
          {perAmount > 1 && (
            <span className="text-[13px] font-bold font-orbitron text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none flex items-center max-md:text-[11px]">
              {perAmount}
            </span>
          )}
          <img
            src={perIcon}
            alt={hasPer.tag || hasPer.type}
            className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
              hasPer.target !== "self-player"
                ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
            }`}
          />
        </div>
      );

      // Special handling for credits-production - use GameIcon with amount inside
      if (baseResourceType === "credit") {
        const itemClasses = !isAffordable
          ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
          : "flex items-center gap-px relative";

        return (
          <div className="flex items-center gap-[3px]">
            <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
              <div className={itemClasses}>
                <GameIcon iconType="credit" amount={Math.abs(amount)} size="small" />
              </div>
              <Slash />
              {perIconEl}
            </div>
            {resolvedComputedAmount !== undefined && (
              <ComputedValueDisplay amount={resolvedComputedAmount} resourceType={resourceType} />
            )}
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
            <div className="flex items-center gap-[3px]">
              <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
                <div className="flex items-center gap-px relative">
                  {amount > 1 && (
                    <span className="text-[20px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none flex items-center ml-0.5 max-md:text-xs">
                      {amount}
                    </span>
                  )}
                  {productionIcon}
                </div>
                <Slash />
                {perIconEl}
              </div>
              {resolvedComputedAmount !== undefined && (
                <ComputedValueDisplay amount={resolvedComputedAmount} resourceType={resourceType} />
              )}
            </div>
          );
        }
      }
    }
  }

  // Handle production WITHOUT per condition - wrap in brown production box
  // "action" context: e.g., energy-production input in Equatorial Magnetizer
  // "default" context: e.g., production outputs in triggered effects (Tharsis Republic, Mining Area)
  if (isProduction && !hasPer && (context === "action" || context === "default")) {
    const baseResourceType = resourceType.replace("-production", "");
    const baseIsCredits = baseResourceType === "credit";

    const itemClasses = !isAffordable
      ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
      : "flex items-center gap-px relative";

    return (
      <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.7)_0%,rgba(139,89,42,0.65)_100%)] border border-[rgba(160,110,60,0.7)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
        <div className={itemClasses}>
          {baseIsCredits ? (
            <GameIcon iconType="credit" amount={Math.abs(amount)} size="small" />
          ) : (
            <>
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
            </>
          )}
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
            <Slash />
            <img
              src={perIcon}
              alt={hasPer.tag || hasPer.type}
              className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                hasPer.target !== "self-player"
                  ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                  : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
              }`}
            />
            {resolvedComputedAmount !== undefined && (
              <ComputedValueDisplay amount={resolvedComputedAmount} resourceType={resourceType} />
            )}
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
              <Slash />
              <img
                src={perIcon}
                alt={hasPer.tag || hasPer.type}
                className={`w-[26px] h-[26px] object-contain max-md:w-[22px] max-md:h-[22px] ${
                  hasPer.target !== "self-player"
                    ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                    : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                }`}
              />
              {resolvedComputedAmount !== undefined && (
                <ComputedValueDisplay amount={resolvedComputedAmount} resourceType={resourceType} />
              )}
            </div>
          );
        }
      }
    }
  }

  if (isCredits) {
    const creditsClasses = `flex items-center gap-0.5 relative ${isAttack ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulseCredits_2s_ease-in-out_infinite]" : ""}`;

    const isNegative = amount < 0 && !isGroupedWithOtherNegatives;

    const finalCreditsClasses = !isAffordable
      ? `${creditsClasses} opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`
      : creditsClasses;

    return (
      <div className={finalCreditsClasses}>
        {isNegative && (
          <span className="relative z-10 text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            -
          </span>
        )}
        <GameIcon
          iconType="credit"
          amount={isVariableAmount ? "X" : Math.abs(amount)}
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

  if (isVP) {
    return (
      <span className="font-orbitron font-bold text-white text-sm leading-none [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        {amount} VP
      </span>
    );
  }

  // Handle global-parameter-lenience with selector-based global params
  if (resourceType === "global-parameter-lenience") {
    const globalParams: string[] = resource?.selectors?.[0]?.globalParameters ?? [
      "temperature",
      "oxygen",
      "ocean",
      "venus",
    ];
    const paramToIcon: Record<string, string> = {
      temperature: "temperature",
      oxygen: "oxygen",
      ocean: "ocean-tile",
      venus: "venus",
    };

    return (
      <div className="flex items-center gap-[2px]">
        {globalParams.map((param: string) => (
          <BehaviorIcon
            key={param}
            resourceType={paramToIcon[param] ?? param}
            isProduction={false}
            isAttack={false}
            context={context}
            isAffordable={isAffordable}
            tileScaleInfo={tileScaleInfo}
          />
        ))}
        <span className="text-base font-bold text-white mx-1 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
          :
        </span>
        <div className="flex items-center gap-[3px]">
          <div className="flex flex-col items-center leading-none -space-y-[11px] translate-y-px">
            <span className="text-sm font-bold font-orbitron text-[#c8e6c9] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              +
            </span>
            <span className="text-sm font-bold font-orbitron text-[#ffcdd2] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              -
            </span>
          </div>
          <span className="text-base font-bold font-orbitron text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            {amount}
          </span>
        </div>
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
    resourceType === "land-claim" ||
    resourceType === "tile-placement";

  const hasTileRestrictions =
    isTilePlacement &&
    resource?.tileRestrictions &&
    (resource.tileRestrictions.adjacency ||
      resource.tileRestrictions.onTileType ||
      (resource.tileRestrictions.boardTags?.length ?? 0) > 0);

  // Check for selector tag badges (e.g., venus badge on animal icon)
  const selectorTags: string[] = [];
  if (resource?.selectors) {
    for (const selector of resource.selectors) {
      if (selector.tags) {
        for (const tag of selector.tags) {
          if (!selectorTags.includes(tag)) selectorTags.push(tag);
        }
      }
    }
  }

  const renderSelectorBadges = () => {
    if (selectorTags.length === 0) return null;
    return selectorTags.map((tag: string) => {
      const tagIcon = getTagIconPath(tag);
      if (!tagIcon) return null;
      return (
        <img
          key={`badge-${tag}`}
          src={tagIcon}
          alt={tag}
          className="w-[14px] h-[14px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-[12px] max-md:h-[12px]"
        />
      );
    });
  };

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
        {(isInput || isAttack || amount < 0) &&
          !isGroupedWithOtherNegatives &&
          context !== "action" && (
            <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[24px] flex items-center justify-center leading-none [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
              -
            </span>
          )}
        {Array.from({ length: absoluteAmount }, (_, i) => (
          <React.Fragment key={i}>{baseIconElement}</React.Fragment>
        ))}
        {renderSelectorBadges()}
        {hasTileRestrictions && (
          <span className="text-white font-bold text-sm ml-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            *
          </span>
        )}
      </div>
    );
  } else {
    return (
      <div className="flex items-center gap-[3px] relative">
        {(isInput || isAttack || amount < 0) &&
          !isGroupedWithOtherNegatives &&
          context !== "action" && (
            <span className="text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-[#ffcdd2] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] flex items-center justify-center">
              -
            </span>
          )}
        <span className="text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] flex items-center justify-center">
          {isVariableAmount ? "X" : Math.abs(amount)}
        </span>
        {baseIconElement}
        {renderSelectorBadges()}
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
