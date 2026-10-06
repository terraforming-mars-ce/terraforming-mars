import React from "react";
import {
  ResourceType,
  ResourcesDto,
  ProductionDto,
  ResourceRemovalTargetDto,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import GameButton from "../buttons/GameButton.tsx";
import {
  GameFlowPopover,
  GameFlowTitle,
  GameFlowBody,
  GameFlowFooter,
} from "./GameFlowPopover.tsx";

interface TargetPlayer {
  id: string;
  name: string;
  resources: ResourcesDto;
  production: ProductionDto;
}

interface TargetPlayerSelectionPopoverProps {
  removalTargets: ResourceRemovalTargetDto[];
  resourceType: ResourceType;
  amount: number;
  isSteal?: boolean;
  players: TargetPlayer[];
  currentPlayerId?: string;
  onPlayerSelect: (playerId: string) => void;
  onCancel: () => void;
  isVisible: boolean;
  mandatory?: boolean;
}

function getPlayerResourceAmount(player: TargetPlayer, resourceType: ResourceType): number {
  switch (resourceType) {
    case "credit":
      return player.resources.credits;
    case "steel":
      return player.resources.steel;
    case "titanium":
      return player.resources.titanium;
    case "plant":
      return player.resources.plants;
    case "energy":
      return player.resources.energy;
    case "heat":
      return player.resources.heat;
    case "credit-production":
      return player.production.credits;
    case "steel-production":
      return player.production.steel;
    case "titanium-production":
      return player.production.titanium;
    case "plant-production":
      return player.production.plants;
    case "energy-production":
      return player.production.energy;
    case "heat-production":
      return player.production.heat;
    default:
      return 0;
  }
}

const TargetPlayerSelectionPopover: React.FC<TargetPlayerSelectionPopoverProps> = ({
  resourceType,
  removalTargets,
  amount,
  isSteal,
  players,
  onPlayerSelect,
  onCancel,
  isVisible,
  mandatory = false,
}) => {
  const isProduction = resourceType.endsWith("-production");
  const displayIconType = resourceType;
  const eligiblePlayers = players.filter((player) =>
    isProduction
      ? getPlayerResourceAmount(player, resourceType) > 0
      : removalTargets.some(
          (t) =>
            t.playerId === player.id &&
            !t.cardId &&
            t.resourceType === resourceType &&
            t.amount > 0,
        ),
  );
  const hasNoTargets = eligiblePlayers.length === 0;
  const canDismiss = !mandatory || hasNoTargets;

  const handleContinueAnyway = () => {
    onPlayerSelect("");
  };

  const handleSkip = () => {
    onPlayerSelect("");
  };

  return (
    <GameFlowPopover
      isVisible={isVisible}
      onClose={onCancel}
      type={canDismiss ? "interactive" : "interactive-mandatory"}
      className="min-w-[280px]"
    >
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          {hasNoTargets ? "No Valid Targets" : "Select Target Player"}
        </h3>
        {!hasNoTargets && (
          <div className="text-white/60 text-xs text-shadow-glow mt-1 flex items-center justify-center gap-1.5">
            <span>{isSteal ? "Steal" : "Remove"} up to</span>
            <GameIcon iconType={displayIconType} amount={amount} size="small" />
            <span>{isProduction ? "production" : ""}</span>
          </div>
        )}
      </GameFlowTitle>

      <GameFlowBody>
        {hasNoTargets ? (
          <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
            <div className="flex items-center justify-center gap-3 mb-4">
              <GameIcon iconType={displayIconType} size="large" />
            </div>
            <div className="text-white/70 text-xs mb-4 max-w-[280px]">
              No eligible players have {isProduction ? `${resourceType} production` : resourceType}{" "}
              available. You can continue without targeting anyone.
            </div>
          </div>
        ) : (
          eligiblePlayers.map((player, index) => {
            const currentAmount = getPlayerResourceAmount(player, resourceType);
            const delay = index * 0.05;

            return (
              <div
                key={player.id}
                className="
                  game-panel game-panel-clipped game-choice
                    px-3.5 py-3
                  mb-2
                  transition-all duration-[250ms] ease-out
                  animate-choiceSlideIn
                  flex items-center justify-between gap-3
                  cursor-pointer hover:brightness-125
                "
                style={{ animationDelay: `${delay}s` }}
                onClick={() => onPlayerSelect(player.id)}
              >
                <div className="text-white font-semibold text-sm">{player.name}</div>
                <div className="flex items-center gap-1.5 flex-shrink-0">
                  <GameIcon iconType={displayIconType} amount={currentAmount} size="small" />
                </div>
              </div>
            );
          })
        )}
      </GameFlowBody>

      {(hasNoTargets || canDismiss) && (
        <GameFlowFooter className="gap-3">
          {hasNoTargets ? (
            <>
              <GameButton
                emphasis="primary"
                tone="warn"
                size="sm"
                clickSound={false}
                onClick={handleContinueAnyway}
              >
                Continue Anyway
              </GameButton>
              <GameButton emphasis="secondary" tone="info" size="sm" onClick={onCancel}>
                Cancel
              </GameButton>
            </>
          ) : (
            <>
              {!isProduction && (
                <GameButton
                  emphasis="primary"
                  tone="warn"
                  size="sm"
                  clickSound={false}
                  onClick={handleSkip}
                >
                  Skip
                </GameButton>
              )}
              <GameButton emphasis="secondary" tone="info" size="sm" onClick={onCancel}>
                Cancel
              </GameButton>
            </>
          )}
        </GameFlowFooter>
      )}
    </GameFlowPopover>
  );
};

export default TargetPlayerSelectionPopover;
