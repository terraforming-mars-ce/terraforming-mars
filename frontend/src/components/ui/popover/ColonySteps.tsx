import React, { useState } from "react";
import { createPortal } from "react-dom";
import { ColonyStepDto, ColonyOutputDto } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";

interface ColonyStepsProps {
  steps: ColonyStepDto[];
  markerPosition: number;
  tradeStepBonus?: number;
  playerColonies: string[];
  maxSlots: number;
  getPlayerColor: (id: string) => string;
  getPlayerName: (id: string) => string;
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
    "card-draw": "card-draw",
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

function isCreditType(type: string): boolean {
  return type === "credit" || type === "credit-production";
}

const ColonySlotTooltip: React.FC<{
  data: { x: number; y: number; label: string; color?: string } | null;
}> = ({ data }) => {
  if (!data) {
    return null;
  }
  return createPortal(
    <div
      className="fixed pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: data.x + 12, top: data.y + 12, zIndex: 99999 }}
    >
      <div className="bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-1.5 shadow-[0_2px_8px_rgba(0,0,0,0.5)] rounded-sm flex items-center gap-1.5">
        {data.color && (
          <div className="w-2 h-2 rounded-full shrink-0" style={{ backgroundColor: data.color }} />
        )}
        <span className="font-orbitron font-bold text-[10px]">{data.label}</span>
      </div>
    </div>,
    document.body,
  );
};

const ColonySteps: React.FC<ColonyStepsProps> = ({
  steps,
  markerPosition,
  tradeStepBonus = 0,
  playerColonies,
  maxSlots,
  getPlayerColor,
  getPlayerName,
}) => {
  const { pattern, sameType } = analyzeSteps(steps);
  const [slotTooltip, setSlotTooltip] = useState<{
    x: number;
    y: number;
    label: string;
    color?: string;
  } | null>(null);

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
                className={`w-4 h-4 rounded-sm cursor-default ${
                  playerColonies[i] ? "" : "border border-white/20"
                }`}
                style={{
                  backgroundColor: playerColonies[i]
                    ? getPlayerColor(playerColonies[i])
                    : "transparent",
                }}
                onMouseEnter={(e) => {
                  const playerId = playerColonies[i];
                  setSlotTooltip({
                    x: e.clientX,
                    y: e.clientY,
                    label: playerId ? getPlayerName(playerId) : "Empty colony slot",
                    color: playerId ? getPlayerColor(playerId) : undefined,
                  });
                }}
                onMouseMove={(e) => {
                  setSlotTooltip((prev) => (prev ? { ...prev, x: e.clientX, y: e.clientY } : null));
                }}
                onMouseLeave={() => setSlotTooltip(null)}
              />
            )}
            {i === steps.length - 1 && markerPosition === steps.length - 1 && (
              <span className="text-[9px] font-orbitron font-bold text-white mt-2">MAX</span>
            )}
          </div>
        ))}
      </div>
      <ColonySlotTooltip data={slotTooltip} />

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

        {/* Boosted marker from Trade Envoys */}
        {tradeStepBonus > 0 && (
          <div
            className="absolute top-0 h-full pointer-events-none z-[8] ring-1 ring-amber-400/60 bg-amber-400/15 rounded-sm"
            style={{
              width: `${100 / stepCount}%`,
              left: `${(Math.min(markerPosition + tradeStepBonus, stepCount - 1) / stepCount) * 100}%`,
              transition: "left 500ms cubic-bezier(0.4, 0, 0.2, 1)",
            }}
          />
        )}

        {steps.map((step, i) => {
          const isFirst = i === 0;
          const isLast = i === steps.length - 1;
          const roundingClass = isFirst ? "rounded-l" : isLast ? "rounded-r" : "";

          return (
            <div
              key={i}
              className={`flex-1 flex items-center justify-center py-1 text-[10px] font-orbitron font-bold min-h-[24px] ${roundingClass} bg-white/[0.02] text-white/50 relative z-20`}
            >
              {pattern === "same-resource-all-one" && sameType && (
                <GameIcon
                  iconType={mapOutputTypeToIcon(sameType)}
                  amount={isCreditType(sameType) ? 1 : undefined}
                  size="small"
                />
              )}
              {pattern === "same-resource-varying" && sameType && isCreditType(sameType) && (
                <GameIcon
                  iconType={mapOutputTypeToIcon(sameType)}
                  amount={getStepAmount(step.outputs)}
                  size="small"
                />
              )}
              {pattern === "mixed-all-one" &&
                step.outputs.map((o, j) => (
                  <GameIcon
                    key={j}
                    iconType={mapOutputTypeToIcon(o.type)}
                    amount={isCreditType(o.type) ? o.amount : undefined}
                    size="small"
                  />
                ))}
              {pattern === "mixed-varying" &&
                step.outputs.map((o, j) => {
                  const useAmountProp = isCreditType(o.type);
                  return (
                    <span key={j} className="inline-flex items-center gap-0.5">
                      {!useAmountProp && <span>{o.amount}</span>}
                      <GameIcon
                        iconType={mapOutputTypeToIcon(o.type)}
                        amount={useAmountProp ? o.amount : undefined}
                        size="small"
                      />
                    </span>
                  );
                })}
            </div>
          );
        })}
      </div>

      {/* Numbers underneath (only for same-resource-varying with non-credit types) */}
      {pattern === "same-resource-varying" && sameType && !isCreditType(sameType) && (
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

export function getTradeExpression(
  steps: ColonyStepDto[],
): { type: string; icon: string; isCreditType: boolean } | null {
  const { pattern, sameType } = analyzeSteps(steps);

  if (sameType) {
    const isCreditType = sameType === "credit" || sameType === "credit-production";
    if (pattern === "same-resource-all-one") {
      return { type: "icon-only", icon: mapOutputTypeToIcon(sameType), isCreditType };
    }
    return { type: "x-icon", icon: mapOutputTypeToIcon(sameType), isCreditType };
  }
  return null;
}

export { mapOutputTypeToIcon };
export default ColonySteps;
