import React from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredit,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { webSocketService } from "@/services/webSocketService.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { FormattedDescription } from "../display/FormattedDescription";

interface MilestonePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const MilestonePopover: React.FC<MilestonePopoverProps> = ({
  isVisible,
  onClose,
  gameState,
  anchorRef,
}) => {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const canClaimMilestones =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  const milestones = gameState?.currentPlayer?.milestones ?? [];
  const claimedCount = milestones.filter((m) => m.isClaimed).length;

  const getPlayerName = (playerId: string | undefined): string => {
    if (!playerId || !gameState) return "Unknown";
    if (playerId === gameState.currentPlayer.id) return gameState.currentPlayer.name;
    const otherPlayer = gameState.otherPlayers.find((p) => p.id === playerId);
    return otherPlayer?.name ?? "Unknown";
  };

  const handleClaimMilestone = (milestoneId: string) => {
    if (!canClaimMilestones) return;
    void webSocketService.claimMilestone(milestoneId);
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "fixed", top: 60, left: 20 }}
      theme="milestones"
      excludeRef={anchorRef}
      header={{
        title: "Milestones",
        badge: <span>{claimedCount}/3 Claimed</span>,
        showCloseButton: true,
      }}
      width={500}
      maxHeight="calc(100vh - 80px)"
      animation="slideDown"
    >
      <div className="p-2">
        {milestones.map((milestone) => {
          const isClaimed = milestone.isClaimed;
          const isAvailable = milestone.available && !isClaimed;
          const isExecutable = canClaimMilestones && isAvailable;
          const progressMet =
            milestone.progress !== undefined &&
            milestone.required !== undefined &&
            milestone.progress >= milestone.required;

          const getState = () => {
            if (isClaimed) return "claimed" as const;
            if (isAvailable) return "available" as const;
            return "disabled" as const;
          };

          return (
            <GamePopoverItem
              key={milestone.type}
              state={getState()}
              onClick={isExecutable ? () => handleClaimMilestone(milestone.type) : undefined}
              error={
                !isAvailable && !isClaimed && milestone.errors && milestone.errors.length > 0
                  ? { message: milestone.errors[0].message, count: milestone.errors.length }
                  : undefined
              }
              statusBadge={isClaimed ? "Claimed" : undefined}
              hoverEffect="background"
              className="mb-2 last:mb-0"
            >
              <div className="flex-1">
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="text-white text-sm font-bold font-orbitron m-0">
                        {milestone.name}
                      </h3>
                    </div>

                    <div className="flex items-center gap-2">
                      <GameIcon
                        iconType={ResourceTypeCredit}
                        amount={milestone.claimCost}
                        size="small"
                      />
                      <span className="text-white/60 text-xs">→</span>
                      <span className="text-amber-400 text-xs font-semibold">5 VP</span>
                    </div>

                    {milestone.progress !== undefined &&
                      milestone.required !== undefined &&
                      !isClaimed && (
                        <div className="mt-2">
                          <div className="flex justify-between text-xs mb-1">
                            <span className={progressMet ? "text-green-400" : "text-amber-400"}>
                              Progress: {milestone.progress}/{milestone.required}
                            </span>
                            {progressMet && <span className="text-green-400">Ready!</span>}
                          </div>
                          <div className="h-1.5 bg-black/40 rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full transition-all ${progressMet ? "bg-green-500" : "bg-amber-500"}`}
                              style={{
                                width: `${Math.min(100, (milestone.progress / milestone.required) * 100)}%`,
                              }}
                            />
                          </div>
                        </div>
                      )}
                  </div>

                  {canClaimMilestones && !isClaimed && (
                    <button
                      className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all cursor-pointer ${
                        isAvailable
                          ? "bg-[var(--popover-accent)]/80 hover:bg-[var(--popover-accent)] text-white shadow-sm hover:shadow-md"
                          : "bg-gray-600/50 text-gray-400"
                      }`}
                      onClick={(e) => {
                        e.stopPropagation();
                        if (isExecutable) handleClaimMilestone(milestone.type);
                      }}
                      disabled={!isAvailable}
                    >
                      Claim
                    </button>
                  )}
                </div>

                <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                  <FormattedDescription text={milestone.description} />
                </p>

                {isClaimed && milestone.claimedBy && (
                  <div className="mt-2 text-xs text-blue-400/80 italic">
                    Claimed by {getPlayerName(milestone.claimedBy)}
                  </div>
                )}
              </div>
            </GamePopoverItem>
          );
        })}
      </div>
    </GamePopover>
  );
};

export default MilestonePopover;
