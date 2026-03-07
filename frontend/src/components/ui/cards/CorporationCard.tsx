import React, { useState } from "react";
import GameIcon from "../display/GameIcon.tsx";
import VictoryPointIcon from "../display/VictoryPointIcon.tsx";
import BehaviorSection from "./BehaviorSection";

import {
  CardBehaviorDto,
  VPConditionDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "../../../types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import { getTagIconPath } from "@/utils/iconStore.ts";
import { FormattedDescription } from "../display/FormattedDescription";

interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingResources?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  behaviors?: CardBehaviorDto[];
  tags?: string[];
  vpConditions?: VPConditionDto[];
  specialEffects?: string[];
  expansion?: string;
  logoPath?: string;
}

interface CorporationCardProps {
  corporation: Corporation;
  isSelected: boolean;
  onSelect: (corporationId: string) => void;
  showCheckbox?: boolean; // Whether to show the selection checkbox (default: false)
  borderColor?: string; // Custom border color (defaults to gold)
  disableInteraction?: boolean; // When true, removes cursor-pointer and hover effects
}

const CorporationCard: React.FC<CorporationCardProps> = ({
  corporation,
  isSelected,
  onSelect,
  showCheckbox = false,
  borderColor,
  disableInteraction = false,
}) => {
  const renderResource = (type: string, amount: number) => {
    const resourceTypeMap: { [key: string]: string } = {
      credits: ResourceTypeCredit,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlant,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };

    const resourceType = resourceTypeMap[type];
    if (!resourceType) return null;

    // Use GameIcon for credits (shows amount inside icon)
    if (type === "credit") {
      return <GameIcon iconType={resourceType} amount={amount} size="large" />;
    }

    // Regular resource display with icon and number
    return <GameIcon iconType={resourceType} amount={amount} size="large" />;
  };

  const renderProduction = (type: string, amount: number) => {
    const resourceTypeMap: { [key: string]: string } = {
      credits: ResourceTypeCredit,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlant,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };

    const resourceType = resourceTypeMap[type];
    if (!resourceType) return null;

    // Use GameIcon with -production suffix for automatic brown background
    return <GameIcon iconType={`${resourceType}-production`} amount={amount} size="medium" />;
  };

  // Extract auto-corporation-first-action behavior for display in starting resources section
  const getAutoCorporationFirstAction = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors) return null;

    return behaviors.find((behavior) =>
      behavior.triggers?.some((t) => t.type === "auto-corporation-first-action"),
    );
  };

  // Extract starting resources from auto-corporation-start behaviors
  const getStartingResourcesFromBehaviors = (
    behaviors: CardBehaviorDto[] | undefined,
  ): {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  } | null => {
    if (!behaviors) return null;

    const startBehavior = behaviors.find((behavior) =>
      behavior.triggers?.some((t) => t.type === "auto-corporation-start"),
    );

    if (!startBehavior?.outputs) return null;

    const resources: {
      credits?: number;
      steel?: number;
      titanium?: number;
      plants?: number;
      energy?: number;
      heat?: number;
    } = {};

    for (const output of startBehavior.outputs) {
      // Only extract immediate resource outputs (not production)
      if (output.type === "credit") resources.credits = output.amount;
      else if (output.type === "steel") resources.steel = output.amount;
      else if (output.type === "titanium") resources.titanium = output.amount;
      else if (output.type === "plant") resources.plants = output.amount;
      else if (output.type === "energy") resources.energy = output.amount;
      else if (output.type === "heat") resources.heat = output.amount;
    }

    return Object.keys(resources).length > 0 ? resources : null;
  };

  // Extract starting production from auto-corporation-start behaviors
  const getStartingProductionFromBehaviors = (
    behaviors: CardBehaviorDto[] | undefined,
  ): {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  } | null => {
    if (!behaviors) return null;

    const startBehavior = behaviors.find((behavior) =>
      behavior.triggers?.some((t) => t.type === "auto-corporation-start"),
    );

    if (!startBehavior?.outputs) return null;

    const production: {
      credits?: number;
      steel?: number;
      titanium?: number;
      plants?: number;
      energy?: number;
      heat?: number;
    } = {};

    for (const output of startBehavior.outputs) {
      // Only extract production outputs
      if (output.type === "credit-production") production.credits = output.amount;
      else if (output.type === "steel-production") production.steel = output.amount;
      else if (output.type === "titanium-production") production.titanium = output.amount;
      else if (output.type === "plant-production") production.plants = output.amount;
      else if (output.type === "energy-production") production.energy = output.amount;
      else if (output.type === "heat-production") production.heat = output.amount;
    }

    return Object.keys(production).length > 0 ? production : null;
  };

  // Render auto-corporation-first-action as simple icons (e.g., card icon with "3" inside)
  const renderAutoCorporationFirstAction = (behavior: CardBehaviorDto) => {
    if (!behavior.outputs || behavior.outputs.length === 0) return null;

    const output = behavior.outputs[0];

    return <GameIcon iconType={output.type} amount={output.amount} size="large" />;
  };

  // Filter out starting bonuses and auto-corporation-first-action (shown separately)
  const filterBehaviors = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors || behaviors.length === 0) return [];

    return behaviors.filter((behavior) => {
      const isAutoCorporationStart = behavior.triggers?.some(
        (t) => t.type === "auto-corporation-start",
      );
      const isAutoCorporationFirstAction = behavior.triggers?.some(
        (t) => t.type === "auto-corporation-first-action",
      );

      // Skip corporation starting bonuses (not an effect, shown in starting resources section)
      if (isAutoCorporationStart) {
        return false;
      }

      // Skip auto-corporation-first-action (shown in starting resources section)
      if (isAutoCorporationFirstAction) {
        return false;
      }

      return true;
    });
  };

  const [isHovered, setIsHovered] = useState(false);

  // Determine border color
  const effectiveBorderColor = borderColor || "#ffc107";

  // Build className based on state
  const getClassName = () => {
    const base =
      "w-[400px] h-[380px] relative bg-[rgba(5,4,2,0.98)] border-2 rounded-xl p-3 transition-all duration-200";

    if (disableInteraction) {
      return base;
    }

    return `${base} cursor-pointer`;
  };

  // Build inline style for dynamic border color and shadow
  const getStyle = (): React.CSSProperties => {
    const color30 = `${effectiveBorderColor}4d`; // 30% opacity
    const color20 = `${effectiveBorderColor}33`; // 20% opacity

    if (disableInteraction || isSelected) {
      return {
        borderColor: effectiveBorderColor,
        boxShadow: `0 4px 20px ${color30}, 0 0 40px ${color20}`,
      };
    }

    return {
      borderColor: isHovered ? effectiveBorderColor : color30,
    };
  };

  return (
    <div
      className={getClassName()}
      style={getStyle()}
      onClick={disableInteraction ? undefined : () => onSelect(corporation.id)}
      onMouseEnter={() => !disableInteraction && setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Logo centered at top */}
      {corporation.logoPath && (
        <div className="flex justify-center mb-2">
          <img
            src={corporation.logoPath}
            alt={corporation.name}
            className="w-20 h-20 rounded-lg object-cover [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.6))]"
          />
        </div>
      )}

      <div className="mb-3 p-3 bg-black/30 rounded-lg flex justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
        {getCorporationLogo(corporation.name.toLowerCase())}
      </div>

      {/* Starting resources, production, and auto-corporation-first-action - compact, no headers */}
      {(() => {
        // Get starting resources from explicit properties or extract from behaviors
        const startingResources =
          corporation.startingResources || getStartingResourcesFromBehaviors(corporation.behaviors);
        const startingProduction =
          corporation.startingProduction ||
          getStartingProductionFromBehaviors(corporation.behaviors);
        const firstAction = getAutoCorporationFirstAction(corporation.behaviors);

        if (!startingResources && !startingProduction && !firstAction) return null;

        return (
          <div className="flex flex-wrap gap-2 justify-center items-center mb-3 pb-3 border-b border-white/20">
            {startingResources &&
              Object.entries(startingResources).map(([type, amount]) =>
                amount && amount > 0 ? (
                  <div key={type} className="flex items-center">
                    {renderResource(type, amount)}
                  </div>
                ) : null,
              )}
            {startingProduction &&
              Object.entries(startingProduction).map(([type, amount]) =>
                amount && amount > 0 ? (
                  <div key={type} className="flex items-center">
                    {renderProduction(type, amount)}
                  </div>
                ) : null,
              )}
            {firstAction && (
              <div className="flex items-center">
                {renderAutoCorporationFirstAction(firstAction)}
              </div>
            )}
          </div>
        );
      })()}

      {/* Behaviors - using BehaviorSection component */}
      {filterBehaviors(corporation.behaviors).length > 0 && (
        <div className="mb-3 border-b border-white/20 pb-3">
          <div className="relative [&>div]:static [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto">
            <BehaviorSection behaviors={filterBehaviors(corporation.behaviors)} />
          </div>
        </div>
      )}

      {/* Description at bottom */}
      <div className="text-xs text-white/80 leading-[1.4] text-center">
        <FormattedDescription text={corporation.description} />
      </div>

      {/* Tags at bottom right */}
      {corporation.tags && corporation.tags.length > 0 && (
        <div className="absolute bottom-3 right-3 flex flex-col gap-1 items-center z-[5]">
          {corporation.tags.map((tag, index) => {
            const tagIcon = getTagIconPath(tag.toLowerCase());
            if (!tagIcon) return null;
            return (
              <div
                key={index}
                className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
              >
                <img
                  src={tagIcon}
                  alt={tag}
                  className="w-6 h-6 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                />
              </div>
            );
          })}
        </div>
      )}

      {/* VP at bottom left */}
      {corporation.vpConditions && corporation.vpConditions.length > 0 && (
        <div className="absolute bottom-3 left-3 z-[5]">
          <VictoryPointIcon vpConditions={corporation.vpConditions} />
        </div>
      )}

      {/* Expansion badge */}
      {corporation.expansion && (
        <div className="absolute top-2 right-2 bg-[rgba(100,150,255,0.3)] text-white/80 py-0.5 px-1.5 rounded text-[9px] uppercase tracking-[0.5px]">
          {corporation.expansion}
        </div>
      )}

      {/* Selection indicator at bottom center (only shown when showCheckbox is true) */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2]">
          <div
            className={`w-6 h-6 rounded-full bg-[#1a1508] border-2 border-[rgba(255,193,7,0.3)] flex items-center justify-center transition-all duration-300 ${isSelected ? "bg-[#3a2f0d] border-[#ffc107]" : ""}`}
          >
            {isSelected && <span className="text-white text-sm font-bold">✓</span>}
          </div>
        </div>
      )}
    </div>
  );
};

export default CorporationCard;
