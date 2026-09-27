import React, { useCallback, useEffect, useMemo, useState } from "react";
import { BehaviorSectionProps, ClassifiedBehavior, BehaviorPresentation } from "./types.ts";
import { CalculatedOutputDto } from "@/types/generated/api-types.ts";
import { classifyBehaviors } from "./utils/behaviorClassifier.ts";
import { detectTilePlacementScale } from "./utils/tileScaling.ts";
import { isResourceAffordable } from "./utils/resourceValidation.ts";
import { analyzeResourceDisplayWithConstraints } from "./utils/displayAnalysis.ts";
import { mergeAutoProductionBehaviors, mergeTriggeredEffects } from "./utils/behaviorMerger.ts";
import { analyzeCardLayout, optimizeBehaviorsForSpace } from "./utils/spaceOptimizer.ts";
import BehaviorContainer from "./components/BehaviorContainer.tsx";
import ManualActionLayout from "./components/ManualActionLayout.tsx";
import TriggeredEffectLayout from "./components/TriggeredEffectLayout.tsx";
import ImmediateResourceLayout from "./components/ImmediateResourceLayout.tsx";
import DiscountLayout from "./components/DiscountLayout.tsx";
import PaymentSubstituteLayout from "./components/PaymentSubstituteLayout.tsx";
import StoragePaymentSubstituteLayout from "./components/StoragePaymentSubstituteLayout.tsx";
import ValueModifierLayout from "./components/ValueModifierLayout.tsx";
import DefenseLayout from "./components/DefenseLayout.tsx";
import BehaviorIcon from "./components/BehaviorIcon.tsx";

function buildBehaviorLayout(
  behaviors: NonNullable<BehaviorSectionProps["behaviors"]>,
  presentation: BehaviorPresentation,
) {
  const classified = classifyBehaviors(behaviors);
  const merged = mergeTriggeredEffects(mergeAutoProductionBehaviors(classified));
  const layout = analyzeCardLayout(merged, presentation === "inspection" ? Infinity : undefined);
  return {
    tileScaleInfo: detectTilePlacementScale(merged),
    cardLayoutPlan: layout,
    optimizedBehaviors: optimizeBehaviorsForSpace(merged, layout),
  };
}

const emptyBehaviors: NonNullable<BehaviorSectionProps["behaviors"]> = [];
const layoutCache = new WeakMap<
  NonNullable<BehaviorSectionProps["behaviors"]>,
  Map<BehaviorPresentation, ReturnType<typeof buildBehaviorLayout>>
>();

function getBehaviorLayout(
  behaviors = emptyBehaviors,
  presentation: BehaviorPresentation = "compact",
) {
  let layouts = layoutCache.get(behaviors);
  if (!layouts) {
    layouts = new Map();
    layoutCache.set(behaviors, layouts);
  }
  let layout = layouts.get(presentation);
  if (!layout) {
    layout = buildBehaviorLayout(behaviors, presentation);
    layouts.set(presentation, layout);
  }
  return layout;
}

const BehaviorSection: React.FC<BehaviorSectionProps> = ({
  behaviors,
  computedValues,
  playerResources,
  resourceStorage,
  cardId,
  greyOutAll = false,
  hideActionChip = false,
  noContainer = false,
  presentation = "compact",
  showTooltips = true,
}) => {
  const [hoveredBehaviorIndex, setHoveredBehaviorIndex] = useState<number | null>(null);
  const tooltipsEnabled = showTooltips && presentation === "compact";
  useEffect(() => {
    setHoveredBehaviorIndex(null);
  }, [tooltipsEnabled]);
  const handleBehaviorHover = useCallback((index: number | null) => {
    setHoveredBehaviorIndex(index);
  }, []);

  const computedValuesByIndex = useMemo(() => {
    const map = new Map<number, CalculatedOutputDto[]>();
    if (computedValues) {
      for (const cv of computedValues) {
        const match = cv.target.match(/^behaviors::(\d+)$/);
        if (match) {
          map.set(parseInt(match[1], 10), cv.outputs);
        }
      }
    }
    return map;
  }, [computedValues]);

  const { tileScaleInfo, cardLayoutPlan, optimizedBehaviors } = getBehaviorLayout(
    behaviors,
    presentation,
  );

  if (!behaviors || behaviors.length === 0) {
    return null;
  }

  // Helper function to check if a resource is affordable (bound to current context)
  const checkResourceAffordable = (resource: any, isInput: boolean = true): boolean => {
    return isResourceAffordable(
      resource,
      isInput,
      playerResources,
      resourceStorage,
      cardId,
      greyOutAll,
    );
  };

  // Helper function to render icons (for ImmediateResourceLayout)
  const renderIcon = (
    resourceType: string,
    isProduction: boolean,
    isAttack: boolean,
    context: "standalone" | "action" | "production" | "default",
    isAffordable: boolean,
  ): React.ReactNode => {
    return (
      <BehaviorIcon
        resourceType={resourceType}
        isProduction={isProduction}
        isAttack={isAttack}
        context={context}
        isAffordable={isAffordable}
        tileScaleInfo={tileScaleInfo}
      />
    );
  };

  // Render individual behavior based on its type
  const renderBehavior = (
    classifiedBehavior: ClassifiedBehavior,
    index: number,
  ): React.ReactNode => {
    const { behavior, type } = classifiedBehavior;
    const layoutPlan = cardLayoutPlan.behaviors[index]?.layoutPlan;
    const behaviorComputedOutputs =
      classifiedBehavior.originalIndex !== undefined
        ? computedValuesByIndex.get(classifiedBehavior.originalIndex)
        : undefined;

    let content: React.ReactNode = null;

    switch (type) {
      case "manual-action":
        content = (
          <ManualActionLayout
            behavior={behavior}
            layoutPlan={layoutPlan}
            isResourceAffordable={checkResourceAffordable}
            analyzeResourceDisplayWithConstraints={analyzeResourceDisplayWithConstraints}
            tileScaleInfo={tileScaleInfo}
            hideActionChip={hideActionChip}
            computedOutputs={behaviorComputedOutputs}
          />
        );
        break;

      case "triggered-effect":
        content = (
          <TriggeredEffectLayout
            behavior={behavior}
            mergedBehaviors={classifiedBehavior.mergedBehaviors}
            layoutPlan={layoutPlan}
            isResourceAffordable={checkResourceAffordable}
            analyzeResourceDisplayWithConstraints={analyzeResourceDisplayWithConstraints}
            tileScaleInfo={tileScaleInfo}
            computedOutputs={behaviorComputedOutputs}
          />
        );
        break;

      case "immediate-production":
      case "immediate-effect":
      case "auto-no-background":
        content = (
          <ImmediateResourceLayout
            behavior={behavior}
            layoutPlan={layoutPlan}
            isResourceAffordable={checkResourceAffordable}
            analyzeResourceDisplayWithConstraints={analyzeResourceDisplayWithConstraints}
            tileScaleInfo={tileScaleInfo}
            renderIcon={renderIcon}
            computedOutputs={behaviorComputedOutputs}
          />
        );
        break;

      case "discount":
        content = <DiscountLayout behavior={behavior} />;
        break;

      case "payment-substitute":
        content = <PaymentSubstituteLayout behavior={behavior} />;
        break;

      case "storage-payment-substitute":
        content = <StoragePaymentSubstituteLayout behavior={behavior} />;
        break;

      case "value-modifier":
        content = <ValueModifierLayout behavior={behavior} />;
        break;

      case "defense":
        content = <DefenseLayout behavior={behavior} />;
        break;
    }

    return (
      <BehaviorContainer
        key={`behavior-${index}`}
        classifiedBehavior={classifiedBehavior}
        index={index}
        description={classifiedBehavior.description}
        isHovered={tooltipsEnabled && hoveredBehaviorIndex === index}
        onHover={tooltipsEnabled ? handleBehaviorHover : undefined}
        noContainer={noContainer}
      >
        {content}
      </BehaviorContainer>
    );
  };

  // Render behaviors with overflow handling if needed
  const containerClass = cardLayoutPlan.needsOverflowHandling
    ? "flex flex-col gap-[3px] items-center w-full max-h-[120px] overflow-y-auto scroll-smooth [scrollbar-width:thin] [&::-webkit-scrollbar]:w-0.5 [&::-webkit-scrollbar-track]:bg-white/10 [&::-webkit-scrollbar-track]:rounded-px [&::-webkit-scrollbar-thumb]:bg-white/30 [&::-webkit-scrollbar-thumb]:rounded-px max-md:gap-px"
    : "flex flex-col gap-[3px] items-center w-full max-md:gap-px";

  return (
    <div className={`behavior-section ${containerClass}`}>
      {optimizedBehaviors.map((classifiedBehavior, index) =>
        renderBehavior(classifiedBehavior, index),
      )}
    </div>
  );
};

export default BehaviorSection;
