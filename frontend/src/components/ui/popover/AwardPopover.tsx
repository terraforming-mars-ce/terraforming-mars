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
import GameButton from "../buttons/GameButton.tsx";
import BehaviorSection from "../cards/BehaviorSection/BehaviorSection.tsx";

interface AwardPopoverProps {
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

const AwardPopover: React.FC<AwardPopoverProps> = ({
  isVisible,
  onClose,
  gameState,
  anchorRef,
}) => {
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const canFundAwards =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  const playerAwards = gameState?.currentPlayer?.awards ?? [];
  const globalAwards = gameState?.awards ?? [];

  // For spectators, use global awards as the item source
  const awards =
    playerAwards.length > 0
      ? playerAwards
      : globalAwards.map((a) => ({
          type: a.type,
          name: a.name,
          description: a.description,
          fundingCost: a.fundingCost,
          isFunded: a.isFunded,
          fundedBy: a.fundedBy,
          available: false,
          errors: [] as import("@/types/generated/api-types.ts").StateErrorDto[],
        }));
  const fundedCount = awards.filter((a) => a.isFunded).length;

  const longestNameLength = useMemo(() => {
    if (!gameState) return 0;
    const players = [
      gameState.currentPlayer?.name ?? "",
      ...gameState.otherPlayers.map((p) => p.name),
    ];
    return players.reduce((max, name) => Math.max(max, name.length), 0);
  }, [gameState]);

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

  const getPlayerName = (playerId: string | undefined): string => {
    if (!playerId) return "Unknown";
    return allPlayers.find((p) => p.id === playerId)?.name ?? "Unknown";
  };

  const handleFundAward = (awardId: string) => {
    if (!canFundAwards) return;
    void webSocketService.fundAward(awardId);
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "fixed", top: 60, left: 20 }}
      theme="colonies"
      excludeRef={anchorRef}
      header={{
        title: "Awards",
        badge: <span>{fundedCount}/3 Funded</span>,
        rightContent: (
          <GameButton buttonType="textonly" size="xs" onClick={onClose}>
            ✕
          </GameButton>
        ),
      }}
      width={500}
      maxHeight="80vh"
      animation="slideDown"
    >
      <div className="p-2 flex flex-col gap-2">
        {awards.map((award) => {
          const isFunded = award.isFunded;
          const isAvailable = award.available && !isFunded;
          const isExecutable = canFundAwards && isAvailable;

          const globalData = globalAwards.find((a) => a.type === award.type);
          const styleColor = globalData?.style?.color ?? "#f39c12";
          const playerProgress = globalData?.playerProgress ?? {};

          const sortedPlayers = [...allPlayers].sort(
            (a, b) => (playerProgress[b.id] ?? 0) - (playerProgress[a.id] ?? 0),
          );

          const getState = () => {
            if (isFunded) return "claimed" as const;
            if (isAvailable) return "available" as const;
            return "disabled" as const;
          };

          return (
            <GamePopoverItem
              key={award.type}
              state={getState()}
              onClick={isExecutable ? () => handleFundAward(award.type) : undefined}
              error={
                !isAvailable && !isFunded && award.errors && award.errors.length > 0
                  ? { message: award.errors[0].message, count: award.errors.length }
                  : undefined
              }
              statusBadge={isFunded ? "Funded" : undefined}
              borderColor={styleColor}
              style={
                isFunded
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
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">{award.name}</h3>
                  {isFunded && award.fundedBy && (
                    <span className="text-white/50 text-xs">
                      Funded by{" "}
                      <span
                        style={{ color: allPlayers.find((p) => p.id === award.fundedBy)?.color }}
                      >
                        {getPlayerName(award.fundedBy)}
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
                          isFunded
                            ? { maxWidth: 0, opacity: 0, gap: 0 }
                            : { maxWidth: "100px", opacity: 1 }
                        }
                      >
                        <GameIcon
                          iconType={ResourceTypeCredit}
                          amount={award.fundingCost}
                          size="small"
                        />
                        <span className="text-white/60 text-xs">→</span>
                      </div>
                      {(globalData?.rewards ?? []).map((reward: AwardRewardDto) => (
                        <div key={reward.place} className="flex items-center gap-1">
                          <span className="text-white/50 text-[10px] font-orbitron">
                            {reward.place === 1 ? "1st" : `${reward.place}nd`}:
                          </span>
                          <div className="[&>div]:items-center [&_div]:justify-start">
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
                        </div>
                      ))}
                    </div>

                    <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                      <FormattedDescription text={award.description} />
                    </p>
                  </div>

                  <div
                    className="flex-shrink-0 grid grid-cols-[auto_auto] gap-x-3 gap-y-1"
                    style={{ minWidth: `${longestNameLength + 5}ch` }}
                  >
                    {sortedPlayers.map((player) => {
                      const score = playerProgress[player.id] ?? 0;
                      return (
                        <React.Fragment key={player.id}>
                          <div className="flex items-center gap-2 text-sm">
                            <span
                              className="w-2 h-2 rounded-full flex-shrink-0"
                              style={{ backgroundColor: player.color }}
                            />
                            <span className="text-white/80">{player.name}</span>
                          </div>
                          <span className="text-sm font-orbitron font-semibold text-white/50">
                            {score}
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

export default AwardPopover;
