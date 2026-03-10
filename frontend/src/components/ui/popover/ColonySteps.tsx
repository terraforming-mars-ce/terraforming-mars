import React from "react";
import { ColonyStepDto, ColonyOutputDto } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";

interface ColonyStepsProps {
  steps: ColonyStepDto[];
  markerPosition: number;
  playerColonies: string[];
  maxSlots: number;
  getPlayerColor: (id: string) => string;
}

type StepPattern =
  | "same-resource-varying"
  | "same-resource-all-one"
  | "mixed-all-one"
  | "mixed-varying";

function analyzeSteps(steps: ColonyStepDto[]): {
  pattern: StepPattern;
  sameType: string | null;
} {
  const types = new Set<string>();
  const amounts = new Set<number>();

  for (const step of steps) {
    for (const output of step.outputs) {
      types.add(output.type);
      amounts.add(output.amount);
    }
  }

  const sameType = types.size === 1 ? [...types][0] : null;
  const allOne = amounts.size === 1 && amounts.has(1);

  if (sameType && allOne) {
    return { pattern: "same-resource-all-one", sameType };
  }
  if (sameType) {
    return { pattern: "same-resource-varying", sameType };
  }
  if (allOne) {
    return { pattern: "mixed-all-one", sameType: null };
  }
  return { pattern: "mixed-varying", sameType: null };
}

function mapOutputTypeToIcon(outputType: string): string {
  const mapping: Record<string, string> = {
    credit: "credit",
    steel: "steel",
    titanium: "titanium",
    plant: "plant",
    energy: "energy",
    heat: "heat",
    "credit-production": "credit-production",
    "steel-production": "steel-production",
    "titanium-production": "titanium-production",
    "plant-production": "plant-production",
    "energy-production": "energy-production",
    "heat-production": "heat-production",
    "card-draw": "card",
    microbe: "microbe",
    animal: "animal",
    floater: "floater",
    "ocean-placement": "ocean",
  };
  return mapping[outputType] ?? outputType;
}

function getStepAmount(outputs: ColonyOutputDto[]): number {
  return outputs.reduce((sum, o) => sum + o.amount, 0);
}

const ColonySteps: React.FC<ColonyStepsProps> = ({
  steps,
  markerPosition,
  playerColonies,
  maxSlots,
  getPlayerColor,
}) => {
  const { pattern, sameType } = analyzeSteps(steps);

  const stepCount = steps.length;
  const markerLeftPercent = (markerPosition / stepCount) * 100;

  return (
    <div className="w-full">
      {/* Colony slots above first N step positions + MAX indicator */}
      <div className="flex w-full mb-1">
        {steps.map((_, i) => (
          <div key={i} className="flex-1 flex justify-center">
            {i < maxSlots && (
              <div
                className={`w-4 h-4 rounded-sm ${
                  playerColonies[i] ? "" : "border border-white/20"
                }`}
                style={{
                  backgroundColor: playerColonies[i]
                    ? getPlayerColor(playerColonies[i])
                    : "transparent",
                }}
              />
            )}
            {i === steps.length - 1 && markerPosition === steps.length - 1 && (
              <span className="text-[9px] font-orbitron font-bold text-white mt-2">MAX</span>
            )}
          </div>
        ))}
      </div>

      {/* Step boxes with sliding marker overlay */}
      <div className="relative flex w-full">
        {/* Animated light region covering steps up to marker */}
        <div
          className="absolute top-0 h-full pointer-events-none z-[5] bg-white/5 rounded-l"
          style={{
            width: `${((markerPosition + 1) / stepCount) * 100}%`,
            transition: "width 500ms cubic-bezier(0.4, 0, 0.2, 1)",
          }}
        />

        {/* Sliding marker highlight */}
        <div
          className="absolute top-0 h-full pointer-events-none z-10 ring-1 ring-white bg-white/25 rounded-sm"
          style={{
            width: `${100 / stepCount}%`,
            left: `${markerLeftPercent}%`,
            transition: "left 500ms cubic-bezier(0.4, 0, 0.2, 1)",
          }}
        />

        {steps.map((step, i) => {
          const isFirst = i === 0;
          const isLast = i === steps.length - 1;
          const roundingClass = isFirst ? "rounded-l" : isLast ? "rounded-r" : "";

          return (
            <div
              key={i}
              className={`flex-1 flex items-center justify-center py-1 text-[10px] font-orbitron font-bold min-h-[24px] ${roundingClass} bg-white/[0.02] text-white/50`}
            >
              {pattern === "same-resource-all-one" && sameType && (
                <GameIcon iconType={mapOutputTypeToIcon(sameType)} size="small" />
              )}
              {pattern === "mixed-all-one" &&
                step.outputs.map((o, j) => (
                  <GameIcon key={j} iconType={mapOutputTypeToIcon(o.type)} size="small" />
                ))}
              {pattern === "mixed-varying" &&
                step.outputs.map((o, j) => (
                  <span key={j} className="inline-flex items-center gap-0.5">
                    <span>{o.amount}</span>
                    <GameIcon iconType={mapOutputTypeToIcon(o.type)} size="small" />
                  </span>
                ))}
            </div>
          );
        })}
      </div>

      {/* Numbers underneath (only for same-resource-varying) */}
      {pattern === "same-resource-varying" && (
        <div className="flex w-full mt-0.5">
          {steps.map((step, i) => (
            <div
              key={i}
              className={`flex-1 text-center text-[11px] font-orbitron font-bold ${
                i === markerPosition ? "text-white" : "text-white/70"
              }`}
            >
              {getStepAmount(step.outputs)}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export function getTradeExpression(steps: ColonyStepDto[]): { type: string; icon: string } | null {
  const { pattern, sameType } = analyzeSteps(steps);

  if (sameType) {
    if (pattern === "same-resource-all-one") {
      return { type: "icon-only", icon: mapOutputTypeToIcon(sameType) };
    }
    return { type: "x-icon", icon: mapOutputTypeToIcon(sameType) };
  }
  return null;
}

export { mapOutputTypeToIcon };
export default ColonySteps;
