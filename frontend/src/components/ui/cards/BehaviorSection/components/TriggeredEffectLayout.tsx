import React from "react";
import ResourceDisplay from "./ResourceDisplay.tsx";
import CardIcon from "./CardIcon.tsx";
import OrChip from "./OrChip.tsx";
import Slash from "./Slash.tsx";
import GameIcon from "../../../display/GameIcon.tsx";
import { CardBehaviorDto, SelectorDto, MinMaxValueDto } from "@/types/generated/api-types.ts";

interface IconDisplayInfo {
  resourceType: string;
  amount: number;
  displayMode: "individual" | "number";
  iconCount: number;
}

interface LayoutPlan {
  rows: IconDisplayInfo[][];
  separators: Array<{ position: number; type: string }>;
  totalRows: number;
}

interface TileScaleInfo {
  scale: 1 | 1.25 | 1.5 | 2;
  tileType: string | null;
}

interface TriggeredEffectLayoutProps {
  behavior: any;
  mergedBehaviors?: CardBehaviorDto[];
  layoutPlan: LayoutPlan;
  isResourceAffordable: (resource: any, isInput: boolean) => boolean;
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo;
  tileScaleInfo: TileScaleInfo;
}

// Extract requiredOriginalCost from selectors (new location) or condition level (legacy)
const getRequiredOriginalCost = (
  selectors: SelectorDto[] | undefined,
  conditionLevelCost: MinMaxValueDto | undefined,
): MinMaxValueDto | undefined => {
  if (selectors) {
    for (const sel of selectors) {
      if (sel.requiredOriginalCost) {
        return sel.requiredOriginalCost;
      }
    }
  }
  return conditionLevelCost;
};

// Render a single selector (AND logic: tags together, then card type)
const renderSelector = (
  selector: any,
  selectorIndex: number,
  triggerIndex: number,
  redGlowClass: string,
): React.ReactNode => {
  const elements: React.ReactNode[] = [];

  if (selector.tags && selector.tags.length > 0) {
    selector.tags.forEach((tag: string, tagIndex: number) => {
      elements.push(
        <GameIcon
          key={`tag-${triggerIndex}-${selectorIndex}-${tagIndex}`}
          iconType={`${tag}-tag`}
          size="small"
        />,
      );
    });
  }

  if (selector.cardTypes && selector.cardTypes.length > 0) {
    selector.cardTypes.forEach((cardType: string, typeIndex: number) => {
      if (cardType === "event") {
        elements.push(
          <GameIcon
            key={`type-${triggerIndex}-${selectorIndex}-${typeIndex}`}
            iconType="event"
            size="small"
          />,
        );
      } else {
        elements.push(
          <span
            key={`type-${triggerIndex}-${selectorIndex}-${typeIndex}`}
            className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]"
          >
            {cardType}
          </span>,
        );
      }
    });
  }

  if (selector.vp) {
    const vpLabel =
      selector.vp.min === 0 && !selector.vp.max
        ? "+ VP"
        : selector.vp.max != null && !selector.vp.min
          ? `≤${selector.vp.max} VP`
          : `≥${selector.vp.min} VP`;
    elements.push(
      <span
        key={`vp-${triggerIndex}-${selectorIndex}`}
        className="font-orbitron text-[10px] font-semibold text-white bg-black/40 border border-white/15 px-1.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)] leading-[22px]"
      >
        {vpLabel}
      </span>,
    );
  }

  return (
    <div
      key={`selector-${triggerIndex}-${selectorIndex}`}
      className={`flex gap-[2px] items-center ${redGlowClass}`}
    >
      {elements}
    </div>
  );
};

// Render a single trigger icon based on its condition type
const renderTriggerIcon = (trigger: any, triggerIndex: number): React.ReactNode => {
  // Check if trigger has selectors (new system)
  const hasSelectors = trigger.condition?.selectors && trigger.condition.selectors.length > 0;

  // Handle standard-project-played triggers with SP chip
  const isStandardProjectPlayed = trigger.condition?.type === "standard-project-played";
  if (isStandardProjectPlayed) {
    const ALL_STANDARD_PROJECTS = [
      "power-plant",
      "asteroid",
      "aquifer",
      "greenery",
      "city",
      "sell-patents",
    ];
    const selectors = trigger.condition?.selectors || [];
    const specifiedProjects = selectors.flatMap((s: any) => s.standardProjects || []);
    const isSubset =
      specifiedProjects.length > 0 && specifiedProjects.length < ALL_STANDARD_PROJECTS.length;

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        <span className="text-[10px] font-semibold text-white bg-black/30 border border-white/15 px-1.5 py-0.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)]">
          SP
        </span>
        {isSubset && (
          <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            *
          </span>
        )}
      </div>
    );
  }

  // Check if trigger is city-placed condition (e.g., Tharsis Republic)
  const isCityPlaced = trigger.condition?.type === "city-placed";

  // Handle global-parameter-raised condition (e.g., Aphrodite for venus)
  const isGlobalParameterRaised = trigger.condition?.type === "global-parameter-raised";

  if (isGlobalParameterRaised) {
    const paramToIcon: Record<string, string> = {
      temperature: "temperature",
      oxygen: "oxygen",
      ocean: "ocean-tile",
      venus: "venus",
    };
    const globalParams: string[] = trigger.condition?.selectors?.[0]?.globalParameters ?? [];

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className="flex items-center gap-[2px]">
          {globalParams.map((param: string) => (
            <GameIcon key={param} iconType={paramToIcon[param] ?? param} size="small" />
          ))}
        </div>
      </div>
    );
  }

  // Handle placement-bonus-gained (e.g., Mining Guild): tile icon + selectors
  const isPlacementBonusGained = trigger.condition?.type === "placement-bonus-gained";
  if (isPlacementBonusGained) {
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <GameIcon iconType="tile-placement" size="small" />
        {hasSelectors && (
          <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            *
          </span>
        )}
      </div>
    );
  }

  // Handle selectors first (new system with AND within selector, OR between selectors)
  if (hasSelectors) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        {trigger.condition.selectors.map((selector: any, selectorIndex: number) => (
          <React.Fragment key={`${triggerIndex}-${selectorIndex}`}>
            {selectorIndex > 0 && <Slash />}
            {renderSelector(selector, selectorIndex, triggerIndex, redGlowClass)}
          </React.Fragment>
        ))}
      </div>
    );
  }

  if (isCityPlaced) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center justify-center ${redGlowClass}`}>
          <GameIcon iconType="city-tile" size="small" />
        </div>
      </div>
    );
  }

  // Check if trigger is greenery-placed condition (e.g., Herbivores)
  const isGreeneryPlaced = trigger.condition?.type === "greenery-placed";

  if (isGreeneryPlaced) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center justify-center ${redGlowClass}`}>
          <GameIcon iconType="greenery-tile" size="small" />
        </div>
      </div>
    );
  }

  // Check if trigger is ocean-placed condition (e.g., Arctic Algae)
  const isOceanPlaced = trigger.condition?.type === "ocean-placed";

  if (isOceanPlaced) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center justify-center ${redGlowClass}`}>
          <GameIcon iconType="ocean-tile" size="small" />
        </div>
      </div>
    );
  }

  // Handle tile-placed condition with onBonusType (e.g., Mining Rights)
  const isTilePlaced = trigger.condition?.type === "tile-placed";
  if (isTilePlaced && trigger.condition?.onBonusType) {
    const bonusTypes: string[] = trigger.condition.onBonusType;
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center gap-[2px] ${redGlowClass}`}>
          {bonusTypes.map((bonusType: string, idx: number) => (
            <React.Fragment key={`bonus-${triggerIndex}-${idx}`}>
              {idx > 0 && <Slash />}
              <GameIcon iconType={bonusType} size="small" />
            </React.Fragment>
          ))}
        </div>
      </div>
    );
  }

  // Check if trigger has requiredOriginalCost (from selectors or legacy condition level)
  const requiredOriginalCost = getRequiredOriginalCost(
    trigger.condition?.selectors,
    trigger.condition?.requiredOriginalCost,
  );
  const hasRequiredOriginalCost = requiredOriginalCost !== undefined;

  if (hasRequiredOriginalCost) {
    const costReq = requiredOriginalCost;
    const hasMin = costReq.min !== undefined;
    const hasMax = costReq.max !== undefined;
    const value = (hasMin ? costReq.min : hasMax ? costReq.max : 0) ?? 0;
    const isMax = hasMax && !hasMin;

    return (
      <div key={triggerIndex} className="flex gap-[3px] items-center">
        {isMax && (
          <span className="text-xs font-semibold text-[#e0e0e0] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
            Max
          </span>
        )}
        <div className="flex items-center gap-0.5">
          <span className="relative z-10 text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            -
          </span>
          <GameIcon iconType="credit" amount={value} size="small" />
        </div>
      </div>
    );
  }

  // Handle trading trigger (e.g., Trade Envoys, Trading Colony)
  if (trigger.condition?.type === "before-colony-trade") {
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <GameIcon iconType="trade" size="small" />
      </div>
    );
  }

  // Handle production-increased trigger (e.g., Manutech partial group)
  const isProductionIncreased = trigger.condition?.type === "production-increased";
  if (isProductionIncreased) {
    const resourceTypes: string[] = trigger.condition?.resourceTypes ?? [];
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        <div className="bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1 py-[2px] shadow-[0_1px_3px_rgba(0,0,0,0.2)] flex items-center gap-[2px]">
          {resourceTypes.map((rt: string, idx: number) => {
            const iconType = PRODUCTION_TYPE_TO_RESOURCE[rt] ?? rt;
            return (
              <React.Fragment key={`prod-${triggerIndex}-${idx}`}>
                {idx > 0 && <Slash />}
                <GameIcon iconType={iconType} size="small" />
              </React.Fragment>
            );
          })}
        </div>
      </div>
    );
  }

  // Fallback to text display for other trigger types
  return (
    <span
      key={triggerIndex}
      className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]"
    >
      {trigger.description || trigger.type || "trigger"}
    </span>
  );
};

// Render a single behavior row (trigger : outputs)
const renderBehaviorRow = (
  behavior: any,
  rowIndex: number,
  isResourceAffordable: (resource: any, isInput: boolean) => boolean,
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo,
  tileScaleInfo: TileScaleInfo,
): React.ReactNode => {
  // Check if this is a global-parameter-lenience effect (special case)
  const isGlobalParameterLenience =
    behavior.outputs?.some((output: any) => output.type === "global-parameter-lenience") ?? false;

  const hasTriggers =
    !isGlobalParameterLenience && behavior.triggers && behavior.triggers.length > 0;

  const hasChoices = behavior.choices && behavior.choices.length > 0;
  const hasOutputs = behavior.outputs && behavior.outputs.length > 0;

  // When we have triggered choices, each choice row renders its own trigger,
  // so skip the main row if there are no behavior-level outputs
  const skipMainRow = hasTriggers && hasChoices && !hasOutputs;

  return (
    <React.Fragment key={`behavior-${rowIndex}`}>
      {!skipMainRow && (
        <div className="flex gap-[3px] items-center justify-center">
          {/* Trigger conditions - hide for global-parameter-lenience */}
          {!isGlobalParameterLenience && behavior.triggers && behavior.triggers.length > 0 && (
            <>
              <div className="flex gap-[3px] items-center">
                {(() => {
                  // Check if any trigger has requiredOriginalCost (from selectors or condition level)
                  const triggersWithCost = behavior.triggers.filter(
                    (trigger: any) =>
                      getRequiredOriginalCost(
                        trigger.condition?.selectors,
                        trigger.condition?.requiredOriginalCost,
                      ) !== undefined,
                  );

                  // If we have cost-based triggers, deduplicate and show once
                  if (triggersWithCost.length > 0) {
                    // Get unique cost requirements
                    const uniqueCosts: string[] = Array.from(
                      new Set(
                        triggersWithCost.map((trigger: any) => {
                          const costReq = getRequiredOriginalCost(
                            trigger.condition?.selectors,
                            trigger.condition?.requiredOriginalCost,
                          );
                          const hasMin = costReq?.min !== undefined;
                          const hasMax = costReq?.max !== undefined;
                          const value = hasMin ? costReq?.min : costReq?.max;
                          const prefix = hasMax && !hasMin ? "Max-" : "";
                          return `${prefix}${value}`;
                        }),
                      ),
                    );

                    // Render unique cost requirements
                    return uniqueCosts.map((costKey: string, idx: number) => {
                      const isMax = costKey.startsWith("Max-");
                      const value = parseInt(costKey.replace("Max-", ""), 10);

                      return (
                        <div key={`cost-${idx}`} className="flex gap-[3px] items-center">
                          {isMax && (
                            <span className="text-xs font-semibold text-[#e0e0e0] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
                              Max
                            </span>
                          )}
                          <div className="flex items-center gap-0.5">
                            <span className="relative z-10 text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                              -
                            </span>
                            <GameIcon iconType="credit" amount={value} size="small" />
                          </div>
                        </div>
                      );
                    });
                  }

                  // Otherwise, render other trigger types normally
                  return behavior.triggers.map((trigger: any, triggerIndex: number) =>
                    renderTriggerIcon(trigger, triggerIndex),
                  );
                })()}
              </div>
              <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                :
              </span>
            </>
          )}

          {/* Outputs in same row if they fit */}
          {behavior.outputs &&
            behavior.outputs.map((output: any, index: number) => {
              const resourceType = output.type || output.resourceType;
              const cardOutputTypes = [
                "card-draw",
                "card-peek",
                "card-take",
                "card-buy",
                "card-discard",
              ];
              if (cardOutputTypes.includes(resourceType)) {
                const badgeType =
                  resourceType === "card-peek"
                    ? "peek"
                    : resourceType === "card-take"
                      ? "take"
                      : resourceType === "card-buy"
                        ? "buy"
                        : resourceType === "card-discard"
                          ? "discard"
                          : "none";
                const isAttack =
                  output.target === "any-player" ||
                  output.target === "all-opponents" ||
                  output.target?.startsWith("steal-");
                return (
                  <CardIcon
                    key={`triggered-output-${rowIndex}-${index}`}
                    amount={Math.abs(output.amount || 1)}
                    badgeType={badgeType}
                    isAffordable={isResourceAffordable(output, false)}
                    isAttack={isAttack}
                    totalCardTypes={1}
                  />
                );
              }
              const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
              return (
                <React.Fragment key={`triggered-output-${rowIndex}-${index}`}>
                  <ResourceDisplay
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="default"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            })}

          {behavior.generationalEventRequirements?.length > 0 && (
            <span className="text-white font-bold text-sm ml-1">*</span>
          )}
        </div>
      )}

      {/* Triggered effect with choices */}
      {behavior.choices &&
        behavior.choices.length > 0 &&
        hasTriggers &&
        (() => {
          const cardOutputTypes = ["card-draw", "card-peek", "card-take", "card-buy"];
          const isSimpleChoices = behavior.choices.every(
            (c: any) =>
              (!c.inputs || c.inputs.length === 0) &&
              c.outputs?.every((o: any) => !cardOutputTypes.includes(o.type)),
          );

          if (isSimpleChoices) {
            // Simple output-only choices: two rows — triggers on top, outputs below
            return (
              <>
                <div className="flex gap-[3px] items-center justify-center">
                  <div className="flex gap-[3px] items-center">
                    {behavior.triggers.map((trigger: any, triggerIndex: number) =>
                      renderTriggerIcon(trigger, triggerIndex),
                    )}
                  </div>
                  <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                    :
                  </span>
                </div>
                <div className="flex gap-[3px] items-center justify-center">
                  {behavior.choices.map((choice: any, idx: number) => (
                    <React.Fragment key={`choice-${rowIndex}-${idx}`}>
                      {idx > 0 && <Slash />}
                      {choice.outputs?.map((output: any, outputIndex: number) => {
                        const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
                        return (
                          <ResourceDisplay
                            key={`choice-${rowIndex}-${idx}-output-${outputIndex}`}
                            displayInfo={displayInfo}
                            isInput={false}
                            resource={output}
                            isGroupedWithOtherNegatives={false}
                            context="default"
                            isAffordable={isResourceAffordable(output, false)}
                            tileScaleInfo={tileScaleInfo}
                          />
                        );
                      })}
                    </React.Fragment>
                  ))}
                </div>
              </>
            );
          }

          // Complex choices (has inputs or card outputs): each choice on its own row
          return behavior.choices.map((choice: any, idx: number) => (
            <div
              key={`choice-row-${rowIndex}-${idx}`}
              className="flex gap-[3px] items-center justify-center"
            >
              <div className="flex gap-[3px] items-center">
                {behavior.triggers.map((trigger: any, triggerIndex: number) =>
                  renderTriggerIcon(trigger, triggerIndex),
                )}
              </div>
              <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                :
              </span>
              {choice.inputs?.map((input: any, inputIndex: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(input, 6, false);
                return (
                  <ResourceDisplay
                    key={`choice-${rowIndex}-${idx}-input-${inputIndex}`}
                    displayInfo={displayInfo}
                    isInput={true}
                    resource={input}
                    isGroupedWithOtherNegatives={false}
                    context="action"
                    isAffordable={isResourceAffordable(input, true)}
                    tileScaleInfo={tileScaleInfo}
                  />
                );
              })}
              {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
                <span className="text-white text-sm font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  →
                </span>
              )}
              {choice.outputs?.map((output: any, outputIndex: number) => {
                const resourceType = output.type || output.resourceType;
                const isCardResource = cardOutputTypes.includes(resourceType);

                if (isCardResource) {
                  const badgeType =
                    resourceType === "card-peek"
                      ? "peek"
                      : resourceType === "card-take"
                        ? "take"
                        : resourceType === "card-buy"
                          ? "buy"
                          : resourceType === "card-discard"
                            ? "discard"
                            : "none";
                  const isAttack =
                    output.target === "any-player" ||
                    output.target === "all-opponents" ||
                    output.target?.startsWith("steal-");
                  return (
                    <CardIcon
                      key={`choice-${rowIndex}-${idx}-output-${outputIndex}`}
                      amount={Math.abs(output.amount || 1)}
                      badgeType={badgeType}
                      isAffordable={isResourceAffordable(output, false)}
                      isAttack={isAttack}
                      totalCardTypes={1}
                    />
                  );
                }

                const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
                return (
                  <ResourceDisplay
                    key={`choice-${rowIndex}-${idx}-output-${outputIndex}`}
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="default"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                );
              })}
              {idx < behavior.choices.length - 1 && <OrChip />}
            </div>
          ));
        })()}

      {/* Non-triggered choices: keep existing single-row layout */}
      {behavior.choices && behavior.choices.length > 0 && !hasTriggers && (
        <div className="flex gap-[6px] items-center justify-center">
          {behavior.choices.map((choice: any, idx: number) => (
            <React.Fragment key={`choice-${rowIndex}-${idx}`}>
              {idx > 0 && <Slash />}
              {choice.inputs?.map((input: any, inputIndex: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(input, 6, false);
                return (
                  <ResourceDisplay
                    key={`choice-${rowIndex}-${idx}-input-${inputIndex}`}
                    displayInfo={displayInfo}
                    isInput={true}
                    resource={input}
                    isGroupedWithOtherNegatives={false}
                    context="action"
                    isAffordable={isResourceAffordable(input, true)}
                    tileScaleInfo={tileScaleInfo}
                  />
                );
              })}
              {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
                <span className="text-white text-sm font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  →
                </span>
              )}
              {choice.outputs?.map((output: any, outputIndex: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
                return (
                  <ResourceDisplay
                    key={`choice-${rowIndex}-${idx}-output-${outputIndex}`}
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="default"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                );
              })}
            </React.Fragment>
          ))}
        </div>
      )}
    </React.Fragment>
  );
};

const isTilePlacementWithBonusProductionPattern = (allBehaviors: any[]): boolean => {
  const hasTilePlacement = allBehaviors.some((b) =>
    b.outputs?.some((o: any) => o.type === "tile-placement"),
  );
  const hasTilePlacedTrigger = allBehaviors.some(
    (b) => b.triggers?.[0]?.condition?.type === "tile-placed",
  );
  return hasTilePlacement && hasTilePlacedTrigger;
};

const ALL_PRODUCTION_TYPES = new Set([
  "credits-production",
  "steel-production",
  "titanium-production",
  "plant-production",
  "energy-production",
  "heat-production",
]);

const isAllStandardResourceProductionPattern = (allBehaviors: any[]): boolean => {
  if (allBehaviors.length < 6) {
    return false;
  }
  const coveredTypes = new Set<string>();
  for (const b of allBehaviors) {
    const trigger = b.triggers?.[0];
    if (trigger?.condition?.type !== "production-increased") {
      return false;
    }
    const resourceTypes: string[] = trigger.condition.resourceTypes ?? [];
    if (resourceTypes.length !== 1) {
      return false;
    }
    if (!ALL_PRODUCTION_TYPES.has(resourceTypes[0])) {
      return false;
    }
    coveredTypes.add(resourceTypes[0]);
  }
  return coveredTypes.size === 6;
};

const PRODUCTION_TYPE_TO_RESOURCE: Record<string, string> = {
  "credits-production": "credit",
  "steel-production": "steel",
  "titanium-production": "titanium",
  "plant-production": "plant",
  "energy-production": "energy",
  "heat-production": "heat",
};

const TriggeredEffectLayout: React.FC<TriggeredEffectLayoutProps> = ({
  behavior,
  mergedBehaviors,
  layoutPlan: _layoutPlan,
  isResourceAffordable,
  analyzeResourceDisplayWithConstraints,
  tileScaleInfo,
}) => {
  // Collect all behaviors to render (primary + merged)
  const allBehaviors = [behavior, ...(mergedBehaviors || [])];

  // Compact rendering for all-6-standard-resource production-increased pattern (Manutech)
  if (isAllStandardResourceProductionPattern(allBehaviors)) {
    return (
      <div className="flex gap-[3px] items-center justify-center">
        <div className="bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)] flex items-center">
          <span className="bg-[rgba(255,255,255,0.9)] text-black text-[10px] font-bold rounded px-1 py-[1px] leading-tight">
            SR
          </span>
        </div>
        <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
          :
        </span>
        <span className="bg-[rgba(255,255,255,0.9)] text-black text-[10px] font-bold rounded px-1 py-[1px] leading-tight">
          SR
        </span>
      </div>
    );
  }

  // Compact rendering for tile-placement + conditional production pattern (Mining Area/Rights)
  if (isTilePlacementWithBonusProductionPattern(allBehaviors)) {
    const tileBehavior = allBehaviors.find((b) =>
      b.outputs?.some((o: any) => o.type === "tile-placement"),
    );
    const tileOutput = tileBehavior?.outputs?.find((o: any) => o.type === "tile-placement");
    const hasBonusRestriction = tileOutput?.tileRestrictions?.onBonusType?.length > 0;
    const triggeredBehaviors = allBehaviors.filter(
      (b) => b.triggers?.[0]?.condition?.type === "tile-placed",
    );

    return (
      <div className="flex flex-col gap-1 items-center justify-center">
        <div className="flex gap-[2px] items-center justify-center">
          <GameIcon iconType="tile-placement" size="small" />
          {hasBonusRestriction && (
            <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              *
            </span>
          )}
        </div>
        {triggeredBehaviors.map((b, index) =>
          renderBehaviorRow(
            b,
            index,
            isResourceAffordable,
            analyzeResourceDisplayWithConstraints,
            tileScaleInfo,
          ),
        )}
      </div>
    );
  }

  // For production-increased triggers, render resource icons inside production box
  return (
    <div className="flex flex-col gap-2 items-center justify-center">
      {allBehaviors.map((b, index) =>
        renderBehaviorRow(
          b,
          index,
          isResourceAffordable,
          analyzeResourceDisplayWithConstraints,
          tileScaleInfo,
        ),
      )}
    </div>
  );
};

export default TriggeredEffectLayout;
