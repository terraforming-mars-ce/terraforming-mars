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
      theme="awards"
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
      maxHeight="calc(100vh - 80px)"
      animation="slideDown"
    >
      <div className="p-2">
        {awards.map((award) => {
          const isFunded = award.isFunded;
          const isAvailable = award.available && !isFunded;
          const isExecutable = canFundAwards && isAvailable;

          const globalData = globalAwards.find((a) => a.type === award.type);
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
              hoverEffect="background"
              className="mb-2 last:mb-0"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h3 className="text-white text-sm font-bold font-orbitron m-0">{award.name}</h3>
                </div>

                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <GameIcon
                        iconType={ResourceTypeCredit}
                        amount={award.fundingCost}
                        size="small"
                      />
                      <span className="text-white/60 text-xs">→</span>
                      <span className="text-amber-400 text-xs font-semibold">
                        5 VP (1st), 2 VP (2nd)
                      </span>
                    </div>

                    <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                      <FormattedDescription text={award.description} />
                    </p>

                    {isFunded && award.fundedBy && (
                      <div className="mt-2 text-xs text-blue-400/80 italic">
                        Funded by {getPlayerName(award.fundedBy)}
                      </div>
                    )}
                  </div>

                  <div
                    className="flex-shrink-0 grid grid-cols-[auto_auto] gap-x-3 gap-y-1"
                    style={{ minWidth: `${longestNameLength + 5}ch` }}
                  >
                    {sortedPlayers.map((player) => {
                      const score = playerProgress[player.id] ?? 0;
                      return (
                        <React.Fragment key={player.id}>
                          <div className="flex items-center gap-2 text-xs">
                            <span
                              className="w-2 h-2 rounded-full flex-shrink-0"
                              style={{ backgroundColor: player.color }}
                            />
                            <span className="text-white/80">{player.name}</span>
                          </div>
                          <span className="text-xs font-orbitron font-semibold text-white/50">
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
