import React, { useState, useRef, useCallback } from "react";
import { VPGranterDto, VPGranterConditionDto } from "@/types/generated/api-types.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { getIconPath } from "@/utils/iconStore.ts";
import DecorBoxTooltip from "../display/DecorBoxTooltip.tsx";

interface VictoryPointsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  vpGranters: VPGranterDto[];
  totalVP: number;
  anchorRef: React.RefObject<HTMLElement>;
}

const ConditionDisplay: React.FC<{ condition: VPGranterConditionDto }> = ({ condition }) => {
  if (condition.conditionType === "per" && condition.perType) {
    const icon = getIconPath(condition.perType);
    const perAmount = condition.perAmount ?? 1;

    return (
      <div className="flex items-center gap-0.5">
        <span className="text-[13px] font-bold text-white/70">{condition.amount}</span>
        <span className="text-[10px] text-white/40">/</span>
        {perAmount > 1 && <span className="text-[13px] font-bold text-white/70">{perAmount}</span>}
        {icon && (
          <img
            src={icon}
            alt={condition.perType}
            className="w-[16px] h-[16px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-[768px]:w-[14px] max-[768px]:h-[14px]"
          />
        )}
        {condition.adjacentToSelfTile && (
          <span className="text-[10px] text-white/40 font-bold">*</span>
        )}
      </div>
    );
  }

  return null;
};

const hasNonStaticCondition = (conditions: VPGranterConditionDto[]) =>
  conditions.some((c) => c.conditionType === "per" && c.perType);

const getDescription = (granter: VPGranterDto): string | null => {
  if (!hasNonStaticCondition(granter.conditions)) return null;
  const perCondition = granter.conditions.find((c) => c.conditionType === "per");
  return perCondition?.explanation || granter.description || null;
};

const VictoryPointsPopover: React.FC<VictoryPointsPopoverProps> = ({
  isVisible,
  onClose,
  vpGranters,
  totalVP,
  anchorRef,
}) => {
  const [tooltipDescription, setTooltipDescription] = useState<string | null>(null);
  const [tooltipPos, setTooltipPos] = useState<{
    x: number;
    y: number;
  } | null>(null);
  const rowRefs = useRef<Map<string, HTMLDivElement>>(new Map());

  const handleMouseEnter = useCallback((granter: VPGranterDto, el: HTMLDivElement) => {
    const desc = getDescription(granter);
    if (desc) {
      const rect = el.getBoundingClientRect();
      setTooltipDescription(desc);
      setTooltipPos({ x: rect.left, y: rect.bottom });
    }
  }, []);

  const handleMouseLeave = useCallback(() => {
    setTooltipDescription(null);
    setTooltipPos(null);
  }, []);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="victoryPoints"
      header={{
        title: `${totalVP} VP`,
        badge: `${vpGranters.length} source${vpGranters.length !== 1 ? "s" : ""}`,
      }}
      width={320}
      maxHeight={400}
    >
      {vpGranters.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No VP sources</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {vpGranters.map((granter, index) => {
            const showConditions = hasNonStaticCondition(granter.conditions);

            return (
              <GamePopoverItem
                key={granter.cardId}
                state="available"
                hoverEffect="glow"
                animationDelay={index * 0.05}
              >
                <div
                  className="flex justify-between items-center flex-1"
                  ref={(el) => {
                    if (el) rowRefs.current.set(granter.cardId, el);
                  }}
                  onMouseEnter={(e) => handleMouseEnter(granter, e.currentTarget as HTMLDivElement)}
                  onMouseLeave={handleMouseLeave}
                >
                  <div className="flex flex-col gap-1">
                    <div className="text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] max-[768px]:text-xs">
                      {granter.cardName}
                    </div>
                    {showConditions && (
                      <div className="flex items-center gap-1.5">
                        {granter.conditions.map((cond, i) => (
                          <ConditionDisplay key={i} condition={cond} />
                        ))}
                        <span className="text-[10px] text-white/50 font-semibold font-orbitron tracking-wider">
                          VP
                        </span>
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
                    <span className="text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm">
                      {granter.computedValue}
                    </span>
                    <span className="text-[10px] font-bold font-orbitron text-white/50">VP</span>
                  </div>
                </div>
              </GamePopoverItem>
            );
          })}
        </div>
      )}
      <DecorBoxTooltip description={tooltipDescription} position={tooltipPos} />
    </GamePopover>
  );
};

export default VictoryPointsPopover;
