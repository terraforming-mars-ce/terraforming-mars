import React, { useEffect, useState } from "react";
import { CardDto, ResourceType } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import CardPreviewPanel from "./CardPreviewPanel.tsx";
import GameButton from "../buttons/GameButton.tsx";
import {
  GameFlowPopover,
  GameFlowTitle,
  GameFlowBody,
  GameFlowFooter,
} from "./GameFlowPopover.tsx";

interface CardResourcePlayer {
  id: string;
  name: string;
  playedCards: CardDto[];
  resourceStorage: { [cardId: string]: number };
}

interface CardResourceSelectionPopoverProps {
  resourceType: ResourceType;
  amount: number;
  excludeCardId?: string;
  players: CardResourcePlayer[];
  onCardSelect: (cardId: string) => void;
  onCancel: () => void;
  isVisible: boolean;
}

function getPlayerTotalResourceOnCards(
  player: CardResourcePlayer,
  resourceType: ResourceType,
  excludeCardId?: string,
): number {
  let total = 0;
  for (const card of player.playedCards) {
    if (card.id === excludeCardId) {
      continue;
    }
    if (card.resourceStorage?.type === resourceType && player.resourceStorage[card.id] > 0) {
      total += player.resourceStorage[card.id];
    }
  }
  return total;
}

function getPlayerCardsWithResource(
  player: CardResourcePlayer,
  resourceType: ResourceType,
  excludeCardId?: string,
): { card: CardDto; count: number }[] {
  const result: { card: CardDto; count: number }[] = [];
  for (const card of player.playedCards) {
    if (card.id === excludeCardId) {
      continue;
    }
    if (card.resourceStorage?.type === resourceType && player.resourceStorage[card.id] > 0) {
      result.push({ card, count: player.resourceStorage[card.id] });
    }
  }
  return result;
}

const CardResourceSelectionPopover: React.FC<CardResourceSelectionPopoverProps> = ({
  resourceType,
  amount,
  excludeCardId,
  players,
  onCardSelect,
  onCancel,
  isVisible,
}) => {
  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);
  const [hoveredCard, setHoveredCard] = useState<CardDto | null>(null);

  const handleBackClick = () => {
    setSelectedPlayerId(null);
    setHoveredCard(null);
  };

  const handleCardClick = (cardId: string) => {
    setSelectedPlayerId(null);
    onCardSelect(cardId);
  };

  const handleContinueAnyway = () => {
    setSelectedPlayerId(null);
    onCardSelect("");
  };

  useEffect(() => {
    if (!isVisible) {
      setSelectedPlayerId(null);
    }
  }, [isVisible]);

  // Custom escape handling: back navigation when player is selected, close otherwise
  useEffect(() => {
    if (!isVisible) {
      return;
    }

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedPlayerId) {
          setSelectedPlayerId(null);
        } else {
          onCancel();
        }
      }
    };

    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isVisible, selectedPlayerId, onCancel]);

  const selectedPlayer = selectedPlayerId ? players.find((p) => p.id === selectedPlayerId) : null;
  const cardsWithResource = selectedPlayer
    ? getPlayerCardsWithResource(selectedPlayer, resourceType, excludeCardId)
    : [];
  const eligiblePlayers = players.filter(
    (player) => getPlayerTotalResourceOnCards(player, resourceType, excludeCardId) > 0,
  );
  const hasNoTargets = eligiblePlayers.length === 0;

  return (
    <GameFlowPopover
      isVisible={isVisible}
      onClose={onCancel}
      handleEscapeKey={false}
      className="min-w-[280px]"
      renderSiblings={<CardPreviewPanel card={hoveredCard} />}
    >
      <GameFlowTitle>
        {selectedPlayer ? (
          <>
            <div className="flex items-center gap-2">
              <GameButton
                buttonType="textonly"
                size="sm"
                onClick={handleBackClick}
                className="!py-0 !px-0 flex items-center gap-1"
              >
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polyline points="15 18 9 12 15 6" />
                </svg>
                Back
              </GameButton>
              <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
                {selectedPlayer.name}&apos;s Cards
              </h3>
            </div>
            <div className="text-white/60 text-xs text-shadow-glow mt-1 flex items-center justify-center gap-1.5">
              <span>Select card to remove {amount}</span>
              <GameIcon iconType={resourceType} size="small" />
              <span>from</span>
            </div>
          </>
        ) : hasNoTargets ? (
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            No Valid Targets
          </h3>
        ) : (
          <>
            <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
              Select Target
            </h3>
            <div className="text-white/60 text-xs text-shadow-glow mt-1 flex items-center justify-center gap-1.5">
              <span>Remove {amount}</span>
              <GameIcon iconType={resourceType} size="small" />
              <span>from any card</span>
            </div>
          </>
        )}
      </GameFlowTitle>

      <GameFlowBody>
        {selectedPlayer ? (
          cardsWithResource.map(({ card, count }, index) => {
            const delay = index * 0.05;
            return (
              <div
                key={card.id}
                className="
                  bg-black/30
                  border-2 border-space-blue-500/40
                  rounded-[10px] px-3.5 py-3
                  mb-2
                  transition-all duration-[250ms] ease-out
                  animate-choiceSlideIn
                  flex items-center justify-between gap-3
                  cursor-pointer hover:border-space-blue-500/80 hover:bg-black/50 hover:shadow-[0_4px_16px_rgba(30,60,150,0.5)]
                "
                style={{ animationDelay: `${delay}s` }}
                onClick={() => handleCardClick(card.id)}
                onMouseEnter={() => setHoveredCard(card)}
                onMouseLeave={() => setHoveredCard(null)}
              >
                <div className="text-white font-semibold text-sm">{card.name}</div>
                <div className="flex items-center gap-1.5 flex-shrink-0">
                  <span className="text-white/60 text-xs font-medium">{count}</span>
                  <GameIcon iconType={resourceType} size="small" />
                </div>
              </div>
            );
          })
        ) : hasNoTargets ? (
          <div className="flex flex-col items-center justify-center py-8 px-4 text-center">
            <div className="flex items-center justify-center gap-3 mb-4">
              <GameIcon iconType={resourceType} size="large" />
            </div>
            <div className="text-white/70 text-xs mb-4 max-w-[280px]">
              No cards have {resourceType} resources available. You can continue without targeting
              any card.
            </div>
          </div>
        ) : (
          eligiblePlayers.map((player, index) => {
            const totalResources = getPlayerTotalResourceOnCards(
              player,
              resourceType,
              excludeCardId,
            );
            const delay = index * 0.05;

            return (
              <div
                key={player.id}
                className="
                  bg-black/30
                  border-2 border-space-blue-500/40
                  rounded-[10px] px-3.5 py-3
                  mb-2
                  transition-all duration-[250ms] ease-out
                  animate-choiceSlideIn
                  flex items-center justify-between gap-3
                  cursor-pointer hover:border-space-blue-500/80 hover:bg-black/50 hover:shadow-[0_4px_16px_rgba(30,60,150,0.5)]
                "
                style={{ animationDelay: `${delay}s` }}
                onClick={() => setSelectedPlayerId(player.id)}
              >
                <div className="text-white font-semibold text-sm">{player.name}</div>
                <div className="flex items-center gap-1.5 flex-shrink-0">
                  <span className="text-white/60 text-xs font-medium">{totalResources}</span>
                  <GameIcon iconType={resourceType} size="small" />
                </div>
              </div>
            );
          })
        )}
      </GameFlowBody>

      <GameFlowFooter className="gap-3">
        {hasNoTargets && !selectedPlayer ? (
          <>
            <GameButton
              buttonType="primary"
              variant="warn"
              size="sm"
              onClick={handleContinueAnyway}
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
    </GameFlowPopover>
  );
};

export default CardResourceSelectionPopover;
