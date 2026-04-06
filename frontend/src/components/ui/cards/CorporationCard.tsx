import React, { useState, useRef, useEffect, useCallback } from "react";
import GameIcon from "../display/GameIcon.tsx";
import CardDecorBar from "../display/CardDecorBar.tsx";
import DecorBoxTooltip from "../display/DecorBoxTooltip.tsx";
import BehaviorSection from "./BehaviorSection";

import {
  CardDto,
  CardBehaviorDto,
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
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

interface CorporationCardProps {
  card: CardDto;
  isSelected: boolean;
  onSelect: (cardId: string) => void;
  showCheckbox?: boolean;
  borderColor?: string;
  disableInteraction?: boolean;
}

const CARD_CLIP_PATH = "polygon(0 0, calc(100% - 28px) 0, 100% 28px, 100% 100%, 0 100%)";

const ACCENT_COLOR = "#ffc107";

const CorporationCard: React.FC<CorporationCardProps> = ({
  card,
  isSelected,
  onSelect,
  showCheckbox = false,
  borderColor,
  disableInteraction = false,
}) => {
  const [startingDescription, setStartingDescription] = useState<string | null>(null);
  const [startingTooltipPos, setStartingTooltipPos] = useState<{
    x: number;
    y: number;
  } | null>(null);
  const startingRef = useRef<HTMLDivElement>(null);
  const { playCardHoverSound } = useSoundEffects();
  const pendingSoundRef = useRef(false);

  useEffect(() => {
    if (pendingSoundRef.current) {
      pendingSoundRef.current = false;
      void playCardHoverSound();
    }
  }, [isSelected, playCardHoverSound]);

  useEffect(() => {
    if (startingDescription && startingRef.current) {
      const rect = startingRef.current.getBoundingClientRect();
      setStartingTooltipPos({ x: rect.left + rect.width / 2, y: rect.bottom });
    } else {
      setStartingTooltipPos(null);
    }
  }, [startingDescription]);

  const effectiveBorderColor = borderColor || ACCENT_COLOR;

  const handleClick = useCallback(() => {
    pendingSoundRef.current = true;
    onSelect(card.id);
  }, [onSelect, card.id]);

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
    if (!resourceType) {
      return null;
    }
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
    if (!resourceType) {
      return null;
    }
    return <GameIcon iconType={`${resourceType}-production`} amount={amount} size="medium" />;
  };

  const getAutoCorporationFirstAction = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors) {
      return null;
    }
    return behaviors.find((behavior) =>
      behavior.triggers?.some((t) => t.type === "auto-corporation-first-action"),
    );
  };

  const getAutoCorporationStart = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors) {
      return null;
    }
    return behaviors.find((behavior) =>
      behavior.triggers?.some((t) => t.type === "auto-corporation-start" && !t.condition),
    );
  };

  const getStartingResourcesFromBehaviors = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors) {
      return null;
    }
    const startBehavior = getAutoCorporationStart(behaviors);
    if (!startBehavior?.outputs) {
      return null;
    }
    const resources: Record<string, number> = {};
    for (const output of startBehavior.outputs) {
      if (output.type === "credit") {
        resources.credits = output.amount;
      } else if (output.type === "steel") {
        resources.steel = output.amount;
      } else if (output.type === "titanium") {
        resources.titanium = output.amount;
      } else if (output.type === "plant") {
        resources.plants = output.amount;
      } else if (output.type === "energy") {
        resources.energy = output.amount;
      } else if (output.type === "heat") {
        resources.heat = output.amount;
      }
    }
    return Object.keys(resources).length > 0 ? resources : null;
  };

  const getStartingProductionFromBehaviors = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors) {
      return null;
    }
    const startBehavior = getAutoCorporationStart(behaviors);
    if (!startBehavior?.outputs) {
      return null;
    }
    const production: Record<string, number> = {};
    for (const output of startBehavior.outputs) {
      if (output.type === "credit-production") {
        production.credits = output.amount;
      } else if (output.type === "steel-production") {
        production.steel = output.amount;
      } else if (output.type === "titanium-production") {
        production.titanium = output.amount;
      } else if (output.type === "plant-production") {
        production.plants = output.amount;
      } else if (output.type === "energy-production") {
        production.energy = output.amount;
      } else if (output.type === "heat-production") {
        production.heat = output.amount;
      }
    }
    return Object.keys(production).length > 0 ? production : null;
  };

  const renderAutoCorporationFirstAction = (behavior: CardBehaviorDto) => {
    if (!behavior.outputs || behavior.outputs.length === 0) {
      return null;
    }
    const output = behavior.outputs[0];
    return <GameIcon iconType={output.type} amount={output.amount} size="large" />;
  };

  const filterBehaviors = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors || behaviors.length === 0) {
      return [];
    }
    return behaviors.filter((behavior) => {
      const isAutoCorporationStart = behavior.triggers?.some(
        (t) => t.type === "auto-corporation-start",
      );
      const isAutoCorporationFirstAction = behavior.triggers?.some(
        (t) => t.type === "auto-corporation-first-action",
      );
      if (isAutoCorporationStart) {
        return behavior.triggers?.some((t) => t.condition !== undefined) ?? false;
      }
      if (isAutoCorporationFirstAction) {
        return false;
      }
      return true;
    });
  };

  const startingResources =
    card.startingResources || getStartingResourcesFromBehaviors(card.behaviors);
  const startingProduction =
    card.startingProduction || getStartingProductionFromBehaviors(card.behaviors);
  const firstAction = getAutoCorporationFirstAction(card.behaviors);
  const startBehavior = getAutoCorporationStart(card.behaviors);
  const filteredBehaviors = filterBehaviors(card.behaviors);
  const hasStartingSection = startingResources || startingProduction || firstAction;
  const hasBehaviors = filteredBehaviors.length > 0;
  const hasTags = card.tags && card.tags.length > 0;
  const hasVpOrStorage =
    (card.vpConditions && card.vpConditions.length > 0) || card.resourceStorage;

  const startingHoverText = [startBehavior?.description, firstAction?.description]
    .filter(Boolean)
    .join(" ");

  return (
    <div
      className={`relative w-[400px] min-h-[380px] p-4 transition-all duration-200 z-[1] group select-none ${disableInteraction ? "" : "cursor-pointer"}`}
      onClick={disableInteraction ? undefined : handleClick}
    >
      {/* Inner card body with clip-path */}
      <div
        className="absolute inset-0 bg-black shadow-[0_4px_12px_rgba(0,0,0,0.3)]"
        style={{ clipPath: CARD_CLIP_PATH }}
      >
        <div
          className="absolute inset-0 border border-[rgba(60,60,70,0.7)] pointer-events-none transition-colors duration-200 group-hover:border-[rgba(120,120,140,0.8)]"
          style={{ clipPath: CARD_CLIP_PATH }}
        />
        <svg
          className="absolute top-0 right-0 w-[28px] h-[28px] pointer-events-none transition-colors duration-200"
          viewBox="0 0 28 28"
        >
          <line
            x1="0"
            y1="0"
            x2="28"
            y2="28"
            className="stroke-[rgba(60,60,70,0.7)] group-hover:stroke-[rgba(120,120,140,0.8)] transition-all duration-200"
            strokeWidth="2"
          />
        </svg>
      </div>

      {/* Left accent stripe */}
      <div
        className="absolute -left-[5px] top-[2.5%] bottom-[2.5%] w-[5px] z-[0] transition-all duration-300"
        style={{
          filter: isSelected
            ? `drop-shadow(0 0 6px ${effectiveBorderColor}) drop-shadow(0 0 12px ${effectiveBorderColor}80)`
            : "none",
        }}
      >
        <div
          className="w-full h-full"
          style={{
            backgroundColor: effectiveBorderColor,
            clipPath: "polygon(0 4px, 100% 0, 100% 100%, 0 calc(100% - 4px))",
          }}
        />
      </div>

      {/* VP + Resource Storage - bottom right */}
      {hasVpOrStorage && (
        <div className="absolute bottom-0 right-0 z-[5]">
          <CardDecorBar
            vpConditions={card.vpConditions}
            resourceStorage={card.resourceStorage}
            corner="top-left"
          />
        </div>
      )}

      {/* Corporation logo area */}
      <div className="relative z-[1] mb-1 p-4 bg-black/30 rounded-lg flex justify-center items-center min-h-[110px] [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.3))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px]">
        <div className="transform scale-[1.3] origin-center">
          {getCorporationLogo(card.name.toLowerCase())}
        </div>
      </div>

      {/* Tags on right side */}
      {hasTags && (
        <div className="absolute top-[38%] right-3 flex flex-col gap-1 items-center z-[5] pointer-events-auto">
          {card.tags!.slice(0, 3).map((tag, index) => {
            const tagIcon = getTagIconPath(tag.toLowerCase());
            if (!tagIcon) {
              return null;
            }
            return (
              <div
                key={index}
                className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
              >
                <img
                  src={tagIcon}
                  alt={tag}
                  className="w-8 h-8 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                />
              </div>
            );
          })}
        </div>
      )}

      {/* Content sections */}
      <div className="relative z-[3] mt-2">
        {/* Section 1: Starting resources/production */}
        {hasStartingSection && (
          <div
            className="flex flex-wrap gap-2 justify-center items-center py-2"
            ref={startingRef}
            onMouseEnter={() => {
              if (startingHoverText) {
                setStartingDescription(startingHoverText);
              }
            }}
            onMouseLeave={() => setStartingDescription(null)}
          >
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
            <DecorBoxTooltip description={startingDescription} position={startingTooltipPos} />
          </div>
        )}

        {/* Faded divider */}
        {hasStartingSection && hasBehaviors && (
          <div
            className="h-px mx-4 my-1"
            style={{
              background:
                "linear-gradient(to right, transparent, rgba(255,255,255,0.2), transparent)",
            }}
          />
        )}

        {/* Section 2: Behaviors */}
        {hasBehaviors && (
          <div className="py-2">
            <div className="relative [&>div]:static [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto">
              <BehaviorSection behaviors={filteredBehaviors} />
            </div>
          </div>
        )}

        {/* Description */}
        {card.description && (
          <div className="text-xs text-white/80 leading-[1.4] text-center mt-1 px-2">
            <FormattedDescription text={card.description} />
          </div>
        )}
      </div>

      {/* Selection checkbox */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2]">
          <div
            className="w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all duration-300"
            style={{
              backgroundColor: isSelected ? `${effectiveBorderColor}33` : "#1a1508",
              borderColor: isSelected ? effectiveBorderColor : `${effectiveBorderColor}4d`,
            }}
          >
            {isSelected && <span className="text-white text-sm font-bold">✓</span>}
          </div>
        </div>
      )}
    </div>
  );
};

export default CorporationCard;
