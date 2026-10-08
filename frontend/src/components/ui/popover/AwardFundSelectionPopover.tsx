import React from "react";
import { PendingAwardFundSelectionDto, GameDto } from "@/types/generated/api-types.ts";
import { webSocketService } from "@/services/webSocketService.ts";
import { GameFlowPopover, GameFlowTitle, GameFlowBody } from "./GameFlowPopover.tsx";

interface AwardFundSelectionPopoverProps {
  isOpen: boolean;
  selection: PendingAwardFundSelectionDto;
  gameState?: GameDto;
}

const AWARD_INFO: Record<string, { name: string; description: string }> = {
  landlord: { name: "Landlord", description: "Most tiles on Mars" },
  banker: { name: "Banker", description: "Highest MC production" },
  scientist: { name: "Scientist", description: "Most science tags in play" },
  thermalist: { name: "Thermalist", description: "Most heat resources" },
  miner: { name: "Miner", description: "Most steel and titanium resources" },
};

const AwardFundSelectionPopover: React.FC<AwardFundSelectionPopoverProps> = ({
  isOpen,
  selection,
  gameState,
}) => {
  const handleSelect = (awardType: string) => {
    void webSocketService.confirmAwardFund(awardType);
  };

  const globalAwards = gameState?.awards ?? [];

  return (
    <GameFlowPopover isVisible={isOpen} type="interactive-mandatory">
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          Fund an Award
        </h3>
        <div className="text-white/60 text-xs compact:text-[13px] text-shadow-glow mt-1">
          Choose an award to fund for free
        </div>
      </GameFlowTitle>
      <GameFlowBody>
        {selection.availableAwards.map((awardType, index) => {
          const info = AWARD_INFO[awardType];
          const globalData = globalAwards.find((a) => a.type === awardType);
          const name = info?.name ?? globalData?.name ?? awardType;
          const description = info?.description ?? globalData?.description ?? "";
          const delay = index * 0.05;

          return (
            <div
              key={awardType}
              className="
                game-panel game-panel-clipped game-choice
                px-3.5 py-3
                mb-2 compact:min-h-11
                transition-all duration-[250ms] ease-out
                animate-choiceSlideIn
                cursor-pointer
                hover:brightness-125

              "
              style={{ animationDelay: `${delay}s` }}
              onClick={() => handleSelect(awardType)}
            >
              <h3 className="text-white text-sm font-bold font-orbitron m-0">{name}</h3>
              <p className="text-white/60 text-xs compact:text-[13px] mt-1 m-0">{description}</p>
            </div>
          );
        })}
      </GameFlowBody>
    </GameFlowPopover>
  );
};

export default AwardFundSelectionPopover;
