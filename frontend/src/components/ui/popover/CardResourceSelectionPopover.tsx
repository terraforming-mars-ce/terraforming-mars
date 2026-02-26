import React, { useEffect, useRef, useState } from "react";
import { CardDto, ResourceType } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { Z_INDEX } from "@/constants/zIndex";

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
    if (card.id === excludeCardId) continue;
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
    if (card.id === excludeCardId) continue;
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
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      setSelectedPlayerId(null);
      onCancel();
    }, 200);
  };

  const handleBackClick = () => {
    setSelectedPlayerId(null);
  };

  const handlePlayerClick = (playerId: string) => {
    setSelectedPlayerId(playerId);
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
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedPlayerId) {
          setSelectedPlayerId(null);
        } else {
          handleCancelClick();
        }
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        handleCancelClick();
      }
    };

    const preventScroll = (event: WheelEvent | TouchEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      document.body.style.overflow = "hidden";
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      document.body.style.overflow = "";
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel, selectedPlayerId]);

  useEffect(() => {
    if (!isVisible) {
      setSelectedPlayerId(null);
    }
  }, [isVisible]);

  if (!isVisible) return null;

  const selectedPlayer = selectedPlayerId ? players.find((p) => p.id === selectedPlayerId) : null;
  const cardsWithResource = selectedPlayer
    ? getPlayerCardsWithResource(selectedPlayer, resourceType, excludeCardId)
    : [];
  const eligiblePlayers = players.filter(
    (player) => getPlayerTotalResourceOnCards(player, resourceType, excludeCardId) > 0,
  );
  const hasNoTargets = eligiblePlayers.length === 0;

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-auto overflow-hidden"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className={`
          min-w-[280px] w-fit max-w-[90vw] max-h-[500px]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          ${isClosing ? "animate-fadeOut" : "animate-popIn"}
        `}
        ref={popoverRef}
      >
        {/* Header */}
        <div className="py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60">
          {selectedPlayer ? (
            <>
              <div className="flex items-center gap-2">
                <button
                  className="text-white/60 hover:text-white transition-colors text-xs"
                  onClick={handleBackClick}
                >
                  &larr; Back
                </button>
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
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-space-blue-500/50 scrollbar-track-space-blue-900/30">
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
                  onClick={() => handlePlayerClick(player.id)}
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
        </div>

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center gap-3">
          {hasNoTargets && !selectedPlayer ? (
            <>
              <button
                className="
                  bg-yellow-600/50
                  border-2 border-yellow-500/60
                  rounded-md text-white text-xs font-semibold
                  px-6 py-2 cursor-pointer
                  transition-all duration-200
                  text-shadow-glow font-orbitron
                  shadow-[0_0_8px_rgba(180,120,0,0.4)]
                  hover:bg-yellow-500/60
                  hover:border-yellow-500/80
                                   hover:shadow-[0_0_12px_rgba(180,120,0,0.6)]
                "
                onClick={handleContinueAnyway}
              >
                Continue Anyway
              </button>
              <button
                className="
                  bg-space-blue-600/50
                  border-2 border-space-blue-500/60
                  rounded-md text-white text-xs font-semibold
                  px-6 py-2 cursor-pointer
                  transition-all duration-200
                  text-shadow-glow font-orbitron
                  shadow-[0_0_8px_rgba(30,60,150,0.4)]
                  hover:bg-space-blue-500/60
                  hover:border-space-blue-500/80
                                   hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
                "
                onClick={handleCancelClick}
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              className="
                bg-space-blue-600/50
                border-2 border-space-blue-500/60
                rounded-md text-white text-xs font-semibold
                px-6 py-2 cursor-pointer
                transition-all duration-200
                text-shadow-glow font-orbitron
                shadow-[0_0_8px_rgba(30,60,150,0.4)]
                hover:bg-space-blue-500/60
                hover:border-space-blue-500/80
                               hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
              "
              onClick={handleCancelClick}
            >
              Cancel
            </button>
          )}
        </div>
      </div>

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.9) translateY(-20px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }

        @keyframes fadeOut {
          from {
            opacity: 1;
          }
          to {
            opacity: 0;
          }
        }

        @keyframes choiceSlideIn {
          from {
            opacity: 0;
            transform: translateX(-20px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }

        .animate-choiceSlideIn {
          animation: choiceSlideIn 0.3s ease-out both;
        }
      `}</style>
    </div>
  );
};

export default CardResourceSelectionPopover;
