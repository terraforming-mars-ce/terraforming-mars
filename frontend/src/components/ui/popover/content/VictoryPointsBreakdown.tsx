import React from "react";
import { VPGranterDto, VPGranterConditionDto } from "@/types/generated/api-types.ts";
import { getIconPath } from "@/utils/iconStore.ts";
import RevealTrigger from "../../display/RevealTrigger.tsx";
import { GamePopoverItem, getThemeStyles } from "../../GamePopover";
import type { ContentDensity } from "./density.ts";

export function totalVictoryPoints(vpGranters: VPGranterDto[]): number {
  return vpGranters.reduce((sum, g) => sum + g.computedValue, 0);
}

const ConditionDisplay: React.FC<{ condition: VPGranterConditionDto; isScreen: boolean }> = ({
  condition,
  isScreen,
}) => {
  if (condition.conditionType === "per" && condition.perType) {
    const icon = getIconPath(condition.perType);
    const perAmount = condition.perAmount ?? 1;
    const separatorClass = isScreen ? "text-[13px] text-white/40" : "text-[10px] text-white/40";
    const iconClass = isScreen
      ? "w-[20px] h-[20px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))]"
      : "w-[16px] h-[16px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-[768px]:w-[14px] max-[768px]:h-[14px]";

    return (
      <div className="flex items-center gap-0.5">
        <span className="text-[13px] font-bold text-white/70">{condition.amount}</span>
        <span className={separatorClass}>/</span>
        {perAmount > 1 && <span className="text-[13px] font-bold text-white/70">{perAmount}</span>}
        {icon && <img src={icon} alt={condition.perType} className={iconClass} />}
        {condition.adjacentToSelfTile && <span className={`${separatorClass} font-bold`}>*</span>}
      </div>
    );
  }

  return null;
};

const hasNonStaticCondition = (conditions: VPGranterConditionDto[]) =>
  conditions.some((c) => c.conditionType === "per" && c.perType);

const getDescription = (granter: VPGranterDto): string | null => {
  if (!hasNonStaticCondition(granter.conditions)) {
    return null;
  }
  const perCondition = granter.conditions.find((c) => c.conditionType === "per");
  return perCondition?.explanation || granter.description || null;
};

interface VictoryPointsBreakdownProps {
  vpGranters: VPGranterDto[];
  density: ContentDensity;
}

const VictoryPointsBreakdown: React.FC<VictoryPointsBreakdownProps> = ({ vpGranters, density }) => {
  const isScreen = density === "screen";

  if (vpGranters.length === 0) {
    return (
      <div className="flex items-center justify-center py-10 px-5">
        <span className="font-orbitron text-sm text-white/50">No VP sources</span>
      </div>
    );
  }

  const listClass = isScreen
    ? "popover-list p-3 grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-2"
    : "popover-list popover-list-headed p-2 flex flex-col gap-2";
  const rowClass = isScreen
    ? "flex justify-between items-center flex-1 min-h-11"
    : "flex justify-between items-center flex-1";
  const nameClass = isScreen
    ? "text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]"
    : "text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] max-[768px]:text-xs";
  const smallLabelClass = isScreen
    ? "text-[13px] font-bold font-orbitron text-white/50"
    : "text-[10px] font-bold font-orbitron text-white/50";
  const valueClass = isScreen
    ? "text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right"
    : "text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm";

  return (
    <div className={listClass} style={isScreen ? getThemeStyles("victoryPoints") : undefined}>
      {vpGranters.map((granter, index) => {
        const showConditions = hasNonStaticCondition(granter.conditions);

        return (
          <GamePopoverItem
            className="popover-list-item"
            key={granter.cardId}
            state="available"
            animationDelay={index * 0.05}
          >
            <RevealTrigger
              as="div"
              className={rowClass}
              content={getDescription(granter)}
              placement="below"
            >
              <div className="flex flex-col gap-1">
                <div className={nameClass}>{granter.cardName}</div>
                {showConditions && (
                  <div className="flex items-center gap-1.5">
                    {granter.conditions.map((cond, i) => (
                      <ConditionDisplay key={i} condition={cond} isScreen={isScreen} />
                    ))}
                    <span
                      className={`${isScreen ? "text-[13px]" : "text-[10px]"} text-white/50 font-semibold font-orbitron tracking-wider`}
                    >
                      VP
                    </span>
                  </div>
                )}
              </div>

              <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
                <span className={valueClass}>{granter.computedValue}</span>
                <span className={smallLabelClass}>VP</span>
              </div>
            </RevealTrigger>
          </GamePopoverItem>
        );
      })}
    </div>
  );
};

export default VictoryPointsBreakdown;
