import React, { useState } from "react";
import {
  CardDto,
  ResourceType,
  ColonyResourceReason,
  ColonyResourceReasonTrade,
  ColonyResourceReasonColonyTax,
  ColonyResourceReasonBuild,
  ColonyResourceReasonColonyBonus,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import CardPreviewPanel from "./CardPreviewPanel.tsx";
import GameButton from "../buttons/GameButton.tsx";
import {
  GameFlowPopover,
  GameFlowTitle,
  GameFlowBody,
  GameFlowFooter,
} from "./GameFlowPopover.tsx";

interface CardStorageSelectionPopoverProps {
  resourceType: ResourceType;
  amount: number;
  selectorTags?: string[];
  playedCards: CardDto[];
  corporationCard?: CardDto;
  resourceStorage?: { [cardId: string]: number };
  reason?: ColonyResourceReason;
  mandatory?: boolean;
  onCardSelect: (cardId: string) => void;
  onCancel: () => void;
  isVisible: boolean;
}

interface CardStorageOption {
  card: CardDto;
  currentStorage: number;
}

const CardStorageSelectionPopover: React.FC<CardStorageSelectionPopoverProps> = ({
  resourceType,
  amount,
  selectorTags,
  playedCards,
  corporationCard,
  resourceStorage,
  reason,
  mandatory,
  onCardSelect,
  onCancel,
  isVisible,
}) => {
  const [hoveredCard, setHoveredCard] = useState<CardDto | null>(null);

  const allCandidateCards = [
    ...(corporationCard?.resourceStorage ? [corporationCard] : []),
    ...playedCards,
  ];

  const validCards: CardStorageOption[] = allCandidateCards
    .filter((card) => {
      if (!card.resourceStorage) {
        return false;
      }
      if (resourceType === "card-resource") {
        if (!selectorTags || selectorTags.length === 0) {
          return true;
        }
        return selectorTags.some((tag) => card.tags?.includes(tag as any));
      }
      if (card.resourceStorage.type !== resourceType) {
        return false;
      }
      if (selectorTags && selectorTags.length > 0) {
        return selectorTags.some((tag) => card.tags?.includes(tag as any));
      }
      return true;
    })
    .map((card) => ({
      card,
      currentStorage: resourceStorage?.[card.id] || 0,
    }));

  const hasNoStorage = validCards.length === 0;
  const canDismiss = !mandatory || hasNoStorage;

  const handleContinueWithoutStorage = () => {
    onCardSelect("");
  };

  const reasonTitleMap: Record<string, string> = {
    [ColonyResourceReasonTrade]: "Trading Complete",
    [ColonyResourceReasonColonyTax]: "Colony Taxes Received",
    [ColonyResourceReasonBuild]: "Colony Built",
    [ColonyResourceReasonColonyBonus]: "Colony Bonus",
  };
  const title = hasNoStorage
    ? "No Storage Available"
    : (reason && reasonTitleMap[reason]) || "Select Card Storage";

  return (
    <GameFlowPopover
      isVisible={isVisible}
      onClose={onCancel}
      type={canDismiss ? "interactive" : "interactive-mandatory"}
      renderSiblings={<CardPreviewPanel card={hoveredCard} />}
    >
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          {title}
        </h3>
        <div className="text-white/60 text-xs text-shadow-glow mt-1 flex items-center justify-center gap-1.5">
          {hasNoStorage ? (
            resourceType === "card-resource" ? (
              "You have no matching cards with resource storage"
            ) : (
              `You have no cards that can store ${resourceType}`
            )
          ) : (
            <>
              <span>
                Place {amount} resource{amount !== 1 ? "s" : ""}
              </span>
              {resourceType !== "card-resource" && (
                <GameIcon iconType={resourceType} size="small" />
              )}
            </>
          )}
        </div>
      </GameFlowTitle>

      <GameFlowBody>
        {hasNoStorage ? (
          <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
            <div className="flex items-center justify-center gap-3 mb-4">
              <div className="text-yellow-400 text-4xl">⚠️</div>
              <GameIcon
                iconType={resourceType === "card-resource" ? "card" : resourceType}
                size="large"
              />
            </div>
            <div className="text-white text-sm mb-3 font-semibold">
              No {resourceType === "card-resource" ? "matching" : resourceType} storage available
            </div>
            <div className="text-white/70 text-xs mb-4 max-w-[280px]">
              {resourceType === "card-resource"
                ? "You don't have any matching cards with resource storage. If you continue, the resource will be lost."
                : `You don't have any cards with ${resourceType} storage. If you continue, the ${resourceType} will be lost.`}
            </div>
            <div className="text-white/50 text-xs italic">
              Play cards with resource storage first to avoid losing resources
            </div>
          </div>
        ) : (
          validCards.map(({ card, currentStorage }, index) => {
            const delay = index * 0.05;

            return (
              <div
                key={card.id}
                className="
                  bg-black/30
                  border-2 border-space-blue-500/40
                  rounded-[10px] px-3.5 py-3
                  mb-2 cursor-pointer
                  transition-all duration-[250ms] ease-out
                  hover:border-space-blue-500/80
                  hover:bg-black/50
                  hover:shadow-[0_4px_16px_rgba(30,60,150,0.5)]
                  animate-choiceSlideIn
                  flex items-center justify-between gap-3
                "
                style={{ animationDelay: `${delay}s` }}
                onClick={() => onCardSelect(card.id)}
                onMouseEnter={() => setHoveredCard(card)}
                onMouseLeave={() => setHoveredCard(null)}
              >
                <div className="text-white font-semibold text-sm">{card.name}</div>
                <div className="flex items-center gap-1.5 flex-shrink-0">
                  <span className="text-white/60 text-xs font-medium">{currentStorage}</span>
                  <GameIcon iconType={card.resourceStorage?.type || resourceType} size="small" />
                </div>
              </div>
            );
          })
        )}
      </GameFlowBody>

      {(hasNoStorage || canDismiss) && (
        <GameFlowFooter className="gap-3">
          {hasNoStorage ? (
            <>
              <GameButton
                buttonType="primary"
                variant="warn"
                size="sm"
                onClick={handleContinueWithoutStorage}
              >
                Continue Anyway
              </GameButton>
              <GameButton buttonType="secondary" variant="info" size="sm" onClick={onCancel}>
                Cancel
              </GameButton>
            </>
          ) : (
            <GameButton buttonType="secondary" variant="info" size="sm" onClick={onCancel}>
              Cancel
            </GameButton>
          )}
        </GameFlowFooter>
      )}
    </GameFlowPopover>
  );
};

export default CardStorageSelectionPopover;
