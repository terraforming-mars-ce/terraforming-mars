import React from "react";
import { PlayerActionDto, GameDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import { canPerformActions } from "../../../utils/actionUtils.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";

interface ActionsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  actions: PlayerActionDto[];
  playerName?: string;
  onActionSelect?: (action: PlayerActionDto) => void;
  anchorRef: React.RefObject<HTMLElement>;
  gameState?: GameDto;
}

const ActionsPopover: React.FC<ActionsPopoverProps> = ({
  isVisible,
  onClose,
  actions,
  onActionSelect,
  anchorRef,
  gameState,
}) => {
  const hasPendingTileSelection = gameState?.currentPlayer?.pendingTileSelection;
  const canPlayActions = canPerformActions(gameState) && !hasPendingTileSelection;

  const handleActionClick = (action: PlayerActionDto) => {
    if (onActionSelect) {
      onActionSelect(action);
      onClose();
    }
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="actions"
      header={{
        title: "Card Actions",
        badge: `${actions.filter((a) => a.available).length} available`,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {actions.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No actions</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {actions.map((action, index) => {
            const isAvailable = action.available;
            const isActionPlayable = canPlayActions && isAvailable;
            const isPlayed = action.timesUsedThisGeneration > 0;

            return (
              <GamePopoverItem
                key={`${action.cardId}-${action.behaviorIndex}`}
                state={isAvailable ? "available" : "disabled"}
                onClick={isActionPlayable ? () => handleActionClick(action) : undefined}
                error={
                  !isAvailable &&
                  action.errors &&
                  action.errors.length > 0 &&
                  !action.errors.some((e) => e.code === "action-already-played")
                    ? { message: action.errors[0].message, count: action.errors.length }
                    : undefined
                }
                hoverEffect="glow"
                animationDelay={index * 0.05}
                className={`${!isActionPlayable && isAvailable ? "cursor-default" : ""} ${isPlayed ? "grayscale saturate-0" : ""}`}
              >
                <div className="flex flex-col gap-2 flex-1">
                  <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[10px]">
                    {action.cardName}
                    {isPlayed && (
                      <span className="bg-[linear-gradient(135deg,rgba(120,120,120,0.8)_0%,rgba(80,80,80,0.9)_100%)] text-white/90 text-[8px] font-semibold uppercase tracking-[0.3px] py-0.5 px-1.5 rounded-lg border border-[rgba(120,120,120,0.6)] [text-shadow:none] opacity-100">
                        played
                      </span>
                    )}
                  </div>

                  <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                    <BehaviorSection
                      behaviors={[action.behavior]}
                      playerResources={gameState?.currentPlayer?.resources}
                      resourceStorage={gameState?.currentPlayer?.resourceStorage}
                      cardId={action.cardId}
                      greyOutAll={isPlayed}
                    />
                  </div>
                </div>
              </GamePopoverItem>
            );
          })}
        </div>
      )}
    </GamePopover>
  );
};

export default ActionsPopover;
