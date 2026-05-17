import React, { useEffect, useMemo, useRef, useState } from "react";
import { useGameStore } from "@/stores/gameStore.ts";
import { AwardDto, MilestoneDto } from "@/types/generated/api-types.ts";
import GameIcon from "../../ui/display/GameIcon.tsx";
import DecorBoxTooltip from "../../ui/display/DecorBoxTooltip.tsx";
import AwardScoreboard from "../../ui/display/AwardScoreboard.tsx";
import { FormattedDescription } from "../../ui/display/FormattedDescription.tsx";
import ParallelogramButton, { ANGLE_INDENT, BUTTON_SPACING } from "./ParallelogramButton.tsx";

const SLOTS_PER_SIDE = 3;
const CHIP_WIDTH = 84;
const INNER_CHIP_WIDTH = CHIP_WIDTH - 12;
const CHIP_HEIGHT = 40;
const PLACEHOLDER_COLOR = "#374151";
const MILESTONE_FALLBACK_COLOR = "#ff6b35";
const AWARD_FALLBACK_COLOR = "#f39c12";

interface PlayerInfo {
  id: string;
  name: string;
  color: string;
}

type ChipKind = "milestone" | "award";
type HoverTarget = { kind: ChipKind; slot: number } | null;

const useSlotRefs = () => {
  const ref0 = useRef<HTMLButtonElement>(null);
  const ref1 = useRef<HTMLButtonElement>(null);
  const ref2 = useRef<HTMLButtonElement>(null);
  return [ref0, ref1, ref2] as const;
};

const MilestoneAwardStatusStrip: React.FC = () => {
  const game = useGameStore((state) => state.game);

  const milestoneRefs = useSlotRefs();
  const awardRefs = useSlotRefs();

  const [hovered, setHovered] = useState<HoverTarget>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);

  const allPlayers: PlayerInfo[] = useMemo(() => {
    if (!game) {
      return [];
    }
    const players: PlayerInfo[] = [];
    if (game.currentPlayer?.id) {
      players.push({
        id: game.currentPlayer.id,
        name: game.currentPlayer.name,
        color: game.currentPlayer.color,
      });
    }
    for (const p of game.otherPlayers ?? []) {
      players.push({ id: p.id, name: p.name, color: p.color });
    }
    return players;
  }, [game]);

  const claimedMilestones = useMemo(
    () => (game?.milestones ?? []).filter((m) => m.isClaimed),
    [game?.milestones],
  );
  const fundedAwards = useMemo(
    () => (game?.awards ?? []).filter((a) => a.isFunded),
    [game?.awards],
  );

  useEffect(() => {
    if (!hovered) {
      setTooltipPos(null);
      return;
    }
    const refList = hovered.kind === "milestone" ? milestoneRefs : awardRefs;
    const el = refList[hovered.slot].current;
    if (!el) {
      setTooltipPos(null);
      return;
    }
    const rect = el.getBoundingClientRect();
    setTooltipPos({ x: rect.left + rect.width / 2, y: rect.bottom });
  }, [hovered, milestoneRefs, awardRefs]);

  if (!game) {
    return null;
  }

  const hoveredItem = (() => {
    if (!hovered) {
      return null;
    }
    return hovered.kind === "milestone"
      ? (claimedMilestones[hovered.slot] ?? null)
      : (fundedAwards[hovered.slot] ?? null);
  })();
  const hoveredPlayerId = hoveredItem
    ? hovered?.kind === "milestone"
      ? (hoveredItem as MilestoneDto).claimedBy
      : (hoveredItem as AwardDto).fundedBy
    : undefined;
  const hoveredPlayer = allPlayers.find((p) => p.id === hoveredPlayerId);
  const hoveredLabel = hovered?.kind === "milestone" ? "Claimed by " : "Funded by ";

  // Slot index meaning: 0 = innermost (next to where the two groups meet), SLOTS_PER_SIDE-1 = outermost.
  // Render order for milestones: outermost first (left-to-right places slot 2 → slot 0).
  // Render order for awards: innermost first (left-to-right places slot 0 → slot 2).
  type Connect = "none" | "overlap" | "gap";
  const renderSlot = (kind: ChipKind, slot: number, connect: Connect) => {
    const isMilestone = kind === "milestone";
    const filled = isMilestone ? (claimedMilestones[slot] ?? null) : (fundedAwards[slot] ?? null);
    const ref = isMilestone ? milestoneRefs[slot] : awardRefs[slot];
    const isFilled = filled !== null;
    const isInnermost = slot === 0;

    const fallbackColor = isMilestone ? MILESTONE_FALLBACK_COLOR : AWARD_FALLBACK_COLOR;
    const ownerId = isMilestone
      ? (filled as MilestoneDto | null)?.claimedBy
      : (filled as AwardDto | null)?.fundedBy;
    const ownerColor = allPlayers.find((p) => p.id === ownerId)?.color;
    const chipColor = isFilled ? (ownerColor ?? fallbackColor) : PLACEHOLDER_COLOR;

    const slopedLeft = isMilestone ? "slope-right" : "slope-left";
    const slopedRight = isMilestone ? "slope-left" : "slope-right";
    const leftEdge = isMilestone ? slopedLeft : isInnermost ? "flat" : slopedLeft;
    const rightEdge = isMilestone ? (isInnermost ? "flat" : slopedRight) : slopedRight;
    const overlapMargin = -ANGLE_INDENT + BUTTON_SPACING;
    const marginLeft = connect === "none" ? 0 : connect === "gap" ? BUTTON_SPACING : overlapMargin;

    const handleEnter = () => {
      if (!isFilled) {
        return;
      }
      setHovered({ kind, slot });
    };
    const handleLeave = () => {
      setHovered((current) => {
        if (!current) {
          return current;
        }
        if (current.kind === kind && current.slot === slot) {
          return null;
        }
        return current;
      });
    };

    const wrapperClass = isFilled ? "" : "pointer-events-none opacity-30";
    const iconType = filled?.style?.icon;
    const zIndex = isMilestone ? slot + 1 : SLOTS_PER_SIDE - slot;

    return (
      <div
        key={`${kind}-${slot}`}
        className={`relative ${wrapperClass}`}
        onMouseEnter={handleEnter}
        onMouseLeave={handleLeave}
        style={{
          marginLeft,
          zIndex,
        }}
      >
        <ParallelogramButton
          buttonRef={ref}
          width={isInnermost ? INNER_CHIP_WIDTH : CHIP_WIDTH}
          height={CHIP_HEIGHT}
          color={chipColor}
          leftEdge={leftEdge}
          rightEdge={rightEdge}
        >
          {iconType ? <GameIcon iconType={iconType} size="small" /> : null}
        </ParallelogramButton>
      </div>
    );
  };

  const milestoneOrder = [2, 1, 0];
  const awardOrder = [0, 1, 2];

  return (
    <>
      <div className="flex items-center pointer-events-auto">
        {milestoneOrder.map((slot, i) =>
          renderSlot("milestone", slot, i === 0 ? "none" : "overlap"),
        )}
        {awardOrder.map((slot, i) => renderSlot("award", slot, i === 0 ? "gap" : "overlap"))}
      </div>
      <DecorBoxTooltip position={tooltipPos} maxWidth={hovered?.kind === "award" ? 320 : 260}>
        {hoveredItem ? (
          <div className="flex flex-col gap-2">
            <div className="flex flex-col gap-2">
              <div className="flex items-baseline justify-between gap-3">
                <div className="font-orbitron font-bold text-white text-[12px] leading-tight">
                  {hoveredItem.name}
                </div>
                <div className="text-white/70 text-[10px] leading-tight text-right shrink-0">
                  {hoveredLabel}
                  <span style={{ color: hoveredPlayer?.color ?? "#ffffff" }}>
                    {hoveredPlayer?.name ?? "Unknown"}
                  </span>
                </div>
              </div>
              {hoveredItem.description && (
                <div className="text-white/70 text-[10px] leading-tight">
                  <FormattedDescription text={hoveredItem.description} />
                </div>
              )}
            </div>
            {hovered?.kind === "award" && (
              <AwardScoreboard
                players={allPlayers}
                playerProgress={(hoveredItem as AwardDto).playerProgress ?? {}}
              />
            )}
          </div>
        ) : null}
      </DecorBoxTooltip>
    </>
  );
};

export default MilestoneAwardStatusStrip;
