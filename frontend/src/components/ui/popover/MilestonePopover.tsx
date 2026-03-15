import React, { useMemo } from "react";
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
import GameButton from "../buttons/GameButton.tsx";

interface MilestonePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

interface PlayerInfo {
  id: string;
  name: string;
  color: string;
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

  const playerMilestones = gameState?.currentPlayer?.milestones ?? [];
  const globalMilestones = gameState?.milestones ?? [];

  // For spectators, use global milestones as the item source
  const milestones =
    playerMilestones.length > 0
      ? playerMilestones
      : globalMilestones.map((m) => ({
          type: m.type,
          name: m.name,
          description: m.description,
          claimCost: m.claimCost,
          isClaimed: m.isClaimed,
          claimedBy: m.claimedBy,
          available: false,
          progress: 0,
          required: m.required,
          errors: [] as import("@/types/generated/api-types.ts").StateErrorDto[],
        }));
  const claimedCount = milestones.filter((m) => m.isClaimed).length;

  const allPlayers: PlayerInfo[] = useMemo(() => {
    if (!gameState) return [];
    const players: PlayerInfo[] = [];
    if (gameState.currentPlayer?.id) {
      players.push({
        id: gameState.currentPlayer.id,
        name: gameState.currentPlayer.name,
        color: gameState.currentPlayer.color,
      });
    }
    for (const p of gameState.otherPlayers) {
      players.push({ id: p.id, name: p.name, color: p.color });
    }
    return players;
  }, [gameState]);

  const longestNameLength = useMemo(() => {
    return allPlayers.reduce((max, p) => Math.max(max, p.name.length), 0);
  }, [allPlayers]);

  const getPlayerName = (playerId: string | undefined): string => {
    if (!playerId) return "Unknown";
    return allPlayers.find((p) => p.id === playerId)?.name ?? "Unknown";
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
        rightContent: (
          <GameButton buttonType="textonly" size="xs" onClick={onClose}>
            ✕
          </GameButton>
        ),
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

          const globalData = globalMilestones.find((m) => m.type === milestone.type);
          const playerProgress = globalData?.playerProgress ?? {};
          const required = globalData?.required ?? 0;

          const sortedPlayers = [...allPlayers].sort(
            (a, b) => (playerProgress[b.id] ?? 0) - (playerProgress[a.id] ?? 0),
          );

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
                <div className="flex items-center gap-2 mb-2">
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">
                    {milestone.name}
                  </h3>
                </div>

                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <GameIcon
                        iconType={ResourceTypeCredit}
                        amount={milestone.claimCost}
                        size="small"
                      />
                      <span className="text-white/60 text-xs">→</span>
                      <span className="text-amber-400 text-xs font-semibold">5 VP</span>
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

                  {!isClaimed && required > 0 && (
                    <div
                      className="flex-shrink-0 grid grid-cols-[auto_auto] gap-x-3 gap-y-1"
                      style={{ minWidth: `${longestNameLength + 5}ch` }}
                    >
                      {sortedPlayers.map((player) => {
                        const progress = playerProgress[player.id] ?? 0;
                        const met = progress >= required;
                        return (
                          <React.Fragment key={player.id}>
                            <div className="flex items-center gap-2 text-xs">
                              <span
                                className="w-2 h-2 rounded-full flex-shrink-0"
                                style={{ backgroundColor: player.color }}
                              />
                              <span className="text-white/80">{player.name}</span>
                            </div>
                            <span
                              className={`text-xs font-orbitron font-semibold ${met ? "text-green-400" : "text-white/50"}`}
                            >
                              {progress}/{required}
                            </span>
                          </React.Fragment>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            </GamePopoverItem>
          );
        })}
      </div>
    </GamePopover>
  );
};

export default MilestonePopover;
