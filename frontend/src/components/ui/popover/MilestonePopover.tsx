import React, { useMemo } from "react";
import {
  AwardRewardDto,
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
import BehaviorSection from "../cards/BehaviorSection/BehaviorSection.tsx";

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
      position={{ type: "anchor", anchorRef, placement: "below" }}
      theme="colonies"
      excludeRef={anchorRef}
      header={undefined}
      width={500}
      maxHeight="80vh"
      animation="slideDown"
    >
      <div className="p-2 flex flex-col gap-2">
        {milestones.map((milestone) => {
          const isClaimed = milestone.isClaimed;
          const isAvailable = milestone.available && !isClaimed;
          const isExecutable = canClaimMilestones && isAvailable;

          const globalData = globalMilestones.find((m) => m.type === milestone.type);
          const styleColor = globalData?.style?.color ?? "#ff6b35";
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
              error={(() => {
                if (isAvailable || isClaimed || !milestone.errors?.length) {
                  return undefined;
                }
                const realErrors = milestone.errors.filter((e) => e.category !== "requirement");
                if (realErrors.length > 0) {
                  return { message: realErrors[0].message, count: realErrors.length };
                }
                return undefined;
              })()}
              info={(() => {
                if (isAvailable || isClaimed || !milestone.errors?.length) {
                  return undefined;
                }
                const reqError = milestone.errors.find((e) => e.category === "requirement");
                if (reqError) {
                  return { message: reqError.message };
                }
                return undefined;
              })()}
              statusBadge={isClaimed ? "Claimed" : undefined}
              borderColor={styleColor}
              style={
                isClaimed
                  ? {
                      borderColor: styleColor + "BB",
                      background: "rgba(255,255,255,0.06)",
                    }
                  : undefined
              }
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  {globalData?.style?.icon && (
                    <div className="opacity-70 flex items-center">
                      <GameIcon iconType={globalData.style.icon} size="small" />
                    </div>
                  )}
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">
                    {milestone.name}
                  </h3>
                  {isClaimed && milestone.claimedBy && (
                    <span className="text-white/50 text-xs">
                      Claimed by{" "}
                      <span
                        style={{
                          color: allPlayers.find((p) => p.id === milestone.claimedBy)?.color,
                        }}
                      >
                        {getPlayerName(milestone.claimedBy)}
                      </span>
                    </span>
                  )}
                </div>

                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <div
                        className="flex items-center gap-2 overflow-hidden transition-all duration-700 ease-in-out"
                        style={
                          isClaimed
                            ? { maxWidth: 0, opacity: 0, gap: 0 }
                            : { maxWidth: "100px", opacity: 1 }
                        }
                      >
                        <GameIcon
                          iconType={ResourceTypeCredit}
                          amount={milestone.claimCost}
                          size="small"
                        />
                        <span className="text-white/60 text-xs">→</span>
                      </div>
                      {(globalData?.rewards ?? []).map((reward: AwardRewardDto, idx: number) => (
                        <div key={idx} className="[&>div]:items-center [&_div]:justify-start">
                          <BehaviorSection
                            behaviors={[
                              {
                                triggers: [],
                                inputs: [],
                                outputs: reward.outputs,
                              },
                            ]}
                            noContainer
                          />
                        </div>
                      ))}
                    </div>

                    <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                      <FormattedDescription text={milestone.description} />
                    </p>
                  </div>

                  <div
                    className="flex-shrink-0 grid grid-cols-[auto_auto] gap-x-3 gap-y-1"
                    style={{ minWidth: `${longestNameLength + 5}ch` }}
                  >
                    {sortedPlayers.map((player) => {
                      const progress = playerProgress[player.id] ?? 0;
                      const met = required > 0 && progress >= required;
                      return (
                        <React.Fragment key={player.id}>
                          <div className="flex items-center gap-2 text-sm">
                            <span
                              className="w-2 h-2 rounded-full flex-shrink-0"
                              style={{ backgroundColor: player.color }}
                            />
                            <span className="text-white/80">{player.name}</span>
                          </div>
                          <span
                            className={`text-sm font-orbitron font-semibold ${met ? "text-green-400" : "text-white/50"}`}
                          >
                            {progress}/{required}
                          </span>
                        </React.Fragment>
                      );
                    })}
                  </div>
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
