import React from "react";
import { PlayerActionDto, GameDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import { GamePopover, GamePopoverItem } from "../GamePopover";

interface ActionReusePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  actions: PlayerActionDto[];
  reuseSourceCardId: string;
  onActionSelect: (action: PlayerActionDto) => void;
  gameState?: GameDto;
}

const ActionReusePopover: React.FC<ActionReusePopoverProps> = ({
  isVisible,
  onClose,
  actions,
  reuseSourceCardId,
  onActionSelect,
  gameState,
}) => {
  const usedActions = actions.filter(
    (a) => a.timesUsedThisGeneration > 0 && a.cardId !== reuseSourceCardId,
  );

  const getReuseErrors = (action: PlayerActionDto) =>
    (action.errors || []).filter((e) => e.code !== "action-already-played");

  const availableCount = usedActions.filter((a) => getReuseErrors(a).length === 0).length;

  const handleActionClick = (action: PlayerActionDto) => {
    onActionSelect(action);
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{
        type: "fixed",
        top: window.innerHeight / 2 - 200,
        left: window.innerWidth / 2 - 160,
      }}
      theme="actions"
      header={{
        title: "Reuse Action",
        badge: `${availableCount} available`,
      }}
      width={320}
      maxHeight={400}
    >
      {usedActions.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No used actions to reuse</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {usedActions.map((action, index) => {
            const reuseErrors = getReuseErrors(action);
            const isReuseAvailable = reuseErrors.length === 0;

            return (
              <GamePopoverItem
                key={`${action.cardId}-${action.behaviorIndex}`}
                state={isReuseAvailable ? "available" : "disabled"}
                onClick={isReuseAvailable ? () => handleActionClick(action) : undefined}
                error={
                  !isReuseAvailable && reuseErrors.length > 0
                    ? { message: reuseErrors[0].message, count: reuseErrors.length }
                    : undefined
                }
                hoverEffect="glow"
                animationDelay={index * 0.05}
              >
                <div className="flex flex-col gap-2 flex-1">
                  <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[10px]">
                    {action.cardName}
                  </div>

                  <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                    <BehaviorSection
                      behaviors={[action.behavior]}
                      computedValues={action.computedValues}
                      playerResources={gameState?.currentPlayer?.resources}
                      resourceStorage={gameState?.currentPlayer?.resourceStorage}
                      cardId={action.cardId}
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

export default ActionReusePopover;
