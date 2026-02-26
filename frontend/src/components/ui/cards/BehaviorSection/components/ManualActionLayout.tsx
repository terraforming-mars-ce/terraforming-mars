import React from "react";
import ResourceDisplay from "./ResourceDisplay.tsx";
import CardIcon from "./CardIcon.tsx";
import OrChip from "./OrChip.tsx";
import { analyzeCardOutputs } from "../utils/displayAnalysis.ts";

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

interface ManualActionLayoutProps {
  behavior: any;
  layoutPlan: LayoutPlan;
  isResourceAffordable: (resource: any, isInput: boolean) => boolean;
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo;
  tileScaleInfo: TileScaleInfo;
  hideActionChip?: boolean;
}

const isCardResourceType = (type: string): boolean =>
  type === "card-draw" ||
  type === "card-peek" ||
  type === "card-take" ||
  type === "card-buy" ||
  type === "card-discard";

const getCardBadgeType = (type: string): "peek" | "take" | "buy" | "discard" | "none" =>
  type === "card-peek"
    ? "peek"
    : type === "card-take"
      ? "take"
      : type === "card-buy"
        ? "buy"
        : type === "card-discard"
          ? "discard"
          : "none";

const isAttackTarget = (target: string | undefined): boolean =>
  target === "any-player" || target === "all-opponents" || (target?.startsWith("steal-") ?? false);

const renderChoiceOutputs = (
  outputs: any[],
  choiceIndex: number,
  isResourceAffordable: (resource: any, isInput: boolean) => boolean,
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo,
  tileScaleInfo: TileScaleInfo,
): React.ReactNode => {
  if (!outputs) return null;
  return outputs.map((output: any, outputIndex: number) => {
    const resourceType = output.resourceType || output.type || "";
    if (isCardResourceType(resourceType)) {
      return (
        <CardIcon
          key={`choice-${choiceIndex}-output-${outputIndex}`}
          amount={Math.abs(output.amount ?? 1)}
          badgeType={getCardBadgeType(resourceType)}
          isAffordable={isResourceAffordable(output, false)}
          isAttack={isAttackTarget(output.target)}
          totalCardTypes={1}
        />
      );
    }
    const displayInfo = analyzeResourceDisplayWithConstraints(output, 3, false);
    return (
      <ResourceDisplay
        key={`choice-${choiceIndex}-output-${outputIndex}`}
        displayInfo={displayInfo}
        isInput={false}
        resource={output}
        isGroupedWithOtherNegatives={false}
        context="action"
        isAffordable={isResourceAffordable(output, false)}
        tileScaleInfo={tileScaleInfo}
      />
    );
  });
};

const ManualActionLayout: React.FC<ManualActionLayoutProps> = ({
  behavior,
  layoutPlan: _layoutPlan,
  isResourceAffordable,
  analyzeResourceDisplayWithConstraints,
  tileScaleInfo,
  hideActionChip = false,
}) => {
  // Handle choice-based behaviors
  if (behavior.choices && behavior.choices.length > 0) {
    // Check if choices only have inputs (no outputs) and behavior has outputs
    const choicesOnlyHaveInputs = behavior.choices.every(
      (choice: any) => !choice.outputs || choice.outputs.length === 0,
    );
    const behaviorHasOutputs = behavior.outputs && behavior.outputs.length > 0;

    // Special case: choices with only inputs + behavior-level outputs
    // Format: <input1> / <input2> -> <outputs>
    if (choicesOnlyHaveInputs && behaviorHasOutputs) {
      return (
        <div className="flex items-center justify-center gap-2 w-full">
          {/* All choice inputs with "/" separator */}
          <div className="flex items-center gap-[6px]">
            {behavior.choices.map((choice: any, choiceIndex: number) => (
              <React.Fragment key={`choice-input-${choiceIndex}`}>
                {choiceIndex > 0 && <OrChip />}
                <div className="flex gap-[3px] items-center">
                  {choice.inputs &&
                    choice.inputs.map((input: any, inputIndex: number) => {
                      const displayInfo = analyzeResourceDisplayWithConstraints(input, 3, false);
                      return (
                        <React.Fragment key={`choice-${choiceIndex}-input-${inputIndex}`}>
                          <ResourceDisplay
                            displayInfo={displayInfo}
                            isInput={true}
                            resource={input}
                            isGroupedWithOtherNegatives={false}
                            context="action"
                            isAffordable={isResourceAffordable(input, true)}
                            tileScaleInfo={tileScaleInfo}
                          />
                        </React.Fragment>
                      );
                    })}
                </div>
              </React.Fragment>
            ))}
          </div>

          {/* Arrow separator */}
          <div className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
            →
          </div>

          {/* Behavior-level outputs */}
          <div className="flex flex-col gap-0.5 items-center min-w-0">
            {behavior.outputs &&
              behavior.outputs.map((output: any, outputIndex: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 3, false);
                return (
                  <React.Fragment key={`output-${outputIndex}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="action"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
          </div>

          {behavior.generationalEventRequirements?.length > 0 && (
            <span className="text-white font-bold text-sm ml-1">*</span>
          )}
        </div>
      );
    }

    // Default case: each choice has its own outputs
    // Format: <input1> -> <output1> OR <input2> -> <output2>

    // Calculate total icon count across all choices to determine layout
    const totalIconCount = behavior.choices.reduce((total: number, choice: any) => {
      const inputCount = choice.inputs?.length || 0;
      const outputCount = choice.outputs?.length || 0;
      const hasArrow = inputCount > 0 && outputCount > 0 ? 1 : 0;
      return total + inputCount + outputCount + hasArrow;
    }, 0);
    // Add OR separators between choices
    const totalWithSeparators = totalIconCount + (behavior.choices.length - 1);
    const canFitOnSingleRow = totalWithSeparators < 4;

    // Horizontal layout for small choice-based actions
    if (canFitOnSingleRow) {
      return (
        <div className="flex items-center justify-center gap-1 w-full">
          {behavior.choices.map((choice: any, choiceIndex: number) => (
            <React.Fragment key={`choice-${choiceIndex}`}>
              {/* OR separator between choices */}
              {choiceIndex > 0 && <OrChip />}

              {/* Choice content (inputs -> outputs) */}
              <div className="flex items-center gap-1">
                {/* Inputs */}
                {choice.inputs?.map((input: any, inputIndex: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(input, 3, false);
                  return (
                    <ResourceDisplay
                      key={`choice-${choiceIndex}-input-${inputIndex}`}
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

                {/* Arrow if both inputs and outputs */}
                {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
                  <span className="text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                    →
                  </span>
                )}

                {/* Outputs */}
                {renderChoiceOutputs(
                  choice.outputs,
                  choiceIndex,
                  isResourceAffordable,
                  analyzeResourceDisplayWithConstraints,
                  tileScaleInfo,
                )}
              </div>
            </React.Fragment>
          ))}

          {behavior.generationalEventRequirements?.length > 0 && (
            <span className="text-white font-bold text-sm ml-1">*</span>
          )}
        </div>
      );
    }

    // Vertical layout for larger choice-based actions
    return (
      <div className="flex flex-col gap-1.5 items-center w-full">
        {behavior.choices.map((choice: any, choiceIndex: number) => (
          <div
            key={`choice-${choiceIndex}`}
            className="flex items-center gap-1 w-full justify-center"
          >
            {/* Input side for this choice */}
            <div className="flex flex-col gap-0.5 items-center min-w-0">
              {choice.inputs &&
                choice.inputs.map((input: any, inputIndex: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(input, 3, false);
                  return (
                    <React.Fragment key={`choice-${choiceIndex}-input-${inputIndex}`}>
                      <ResourceDisplay
                        displayInfo={displayInfo}
                        isInput={true}
                        resource={input}
                        isGroupedWithOtherNegatives={false}
                        context="action"
                        isAffordable={isResourceAffordable(input, true)}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </React.Fragment>
                  );
                })}
            </div>

            {/* Arrow separator for this choice */}
            {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
              <div className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                →
              </div>
            )}

            {/* Output side for this choice */}
            <div className="flex flex-col gap-0.5 items-center min-w-0">
              {renderChoiceOutputs(
                choice.outputs,
                choiceIndex,
                isResourceAffordable,
                analyzeResourceDisplayWithConstraints,
                tileScaleInfo,
              )}
            </div>

            {behavior.generationalEventRequirements?.length > 0 && (
              <span className="text-white font-bold text-sm ml-1">*</span>
            )}

            {/* Add "OR" separator between choices (except for the last one) */}
            {choiceIndex < behavior.choices.length - 1 && <OrChip />}
          </div>
        ))}
      </div>
    );
  }

  // Regular behavior handling
  // Analyze and consolidate card outputs (card-draw, card-peek, card-take, card-buy)
  const consolidatedCards = behavior.outputs ? analyzeCardOutputs(behavior.outputs) : [];

  // Helper to check if an output is a card resource
  const isCardResource = (output: any): boolean => {
    const type = output.resourceType || output.type || "";
    return (
      type === "card-draw" ||
      type === "card-peek" ||
      type === "card-take" ||
      type === "card-buy" ||
      type === "card-discard"
    );
  };

  // Filter out card resources from regular outputs (they'll be rendered via consolidatedCards)
  const nonCardOutputs = behavior.outputs
    ? behavior.outputs.filter((output: any) => !isCardResource(output))
    : [];

  const hasInputs = behavior.inputs && behavior.inputs.length > 0;

  return (
    <div className="flex items-center justify-center gap-1 w-full">
      {/* Input side */}
      <div className="flex flex-col gap-0.5 items-center min-w-0">
        {hasInputs ? (
          behavior.inputs.map((input: any, inputIndex: number) => {
            const displayInfo = analyzeResourceDisplayWithConstraints(input, 3, false);
            return (
              <React.Fragment key={`input-${inputIndex}`}>
                <ResourceDisplay
                  displayInfo={displayInfo}
                  isInput={true}
                  resource={input}
                  isGroupedWithOtherNegatives={false}
                  context="action"
                  isAffordable={isResourceAffordable(input, true)}
                  tileScaleInfo={tileScaleInfo}
                />
              </React.Fragment>
            );
          })
        ) : !hideActionChip ? (
          <span className="text-[10px] font-semibold text-white bg-[rgba(33,150,243,0.5)] px-1.5 py-0.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)]">
            Action
          </span>
        ) : null}
      </div>

      {/* Arrow separator - only show when there are inputs */}
      {hasInputs && (
        <div className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
          →
        </div>
      )}

      {/* Output side */}
      <div className="flex flex-col gap-0.5 items-center min-w-0">
        {/* Regular non-card outputs */}
        {nonCardOutputs.map((output: any, outputIndex: number) => {
          const displayInfo = analyzeResourceDisplayWithConstraints(output, 3, false);
          return (
            <React.Fragment key={`output-${outputIndex}`}>
              <ResourceDisplay
                displayInfo={displayInfo}
                isInput={false}
                resource={output}
                isGroupedWithOtherNegatives={false}
                context="action"
                isAffordable={isResourceAffordable(output, false)}
                tileScaleInfo={tileScaleInfo}
              />
            </React.Fragment>
          );
        })}

        {/* Consolidated card icons (card-draw, card-peek, card-take, card-buy) */}
        {consolidatedCards.map((cardItem, index) => (
          <React.Fragment key={`card-${index}`}>
            <CardIcon
              amount={cardItem.amount}
              badgeType={cardItem.badgeType}
              isAffordable={true}
              isAttack={cardItem.isAttack}
              totalCardTypes={consolidatedCards.length}
            />
          </React.Fragment>
        ))}
      </div>

      {behavior.generationalEventRequirements?.length > 0 && (
        <span className="text-white font-bold text-sm ml-1">*</span>
      )}
    </div>
  );
};

export default ManualActionLayout;
