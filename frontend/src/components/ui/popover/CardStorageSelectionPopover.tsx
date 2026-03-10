import React, { useEffect, useRef, useState } from "react";
import { CardDto, ResourceType } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { Z_INDEX } from "@/constants/zIndex";

interface CardStorageSelectionPopoverProps {
  resourceType: ResourceType;
  amount: number;
  selectorTags?: string[];
  playedCards: CardDto[];
  corporationCard?: CardDto;
  resourceStorage?: { [cardId: string]: number }; // Map of cardId to current storage count
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
  onCardSelect,
  onCancel,
  isVisible,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  // Include corporation card (if it has resource storage) alongside played cards
  const allCandidateCards = [
    ...(corporationCard?.resourceStorage ? [corporationCard] : []),
    ...playedCards,
  ];

  // Filter cards to only those with matching resource storage
  // For "card-resource" type, match by selector tags instead of storage type
  // For specific resource types (floater, animal, etc.), match storage type AND selector tags if present
  const validCards: CardStorageOption[] = allCandidateCards
    .filter((card) => {
      if (!card.resourceStorage) return false;
      if (resourceType === "card-resource") {
        // Match cards that have resource storage AND match the selector tags
        if (!selectorTags || selectorTags.length === 0) return true;
        return selectorTags.some((tag) => card.tags?.includes(tag as any));
      }
      // Match by storage type
      if (card.resourceStorage.type !== resourceType) return false;
      // Also filter by selector tags if present (e.g., "any venus card" for floaters)
      if (selectorTags && selectorTags.length > 0) {
        return selectorTags.some((tag) => card.tags?.includes(tag as any));
      }
      return true;
    })
    .map((card) => ({
      card,
      currentStorage: resourceStorage?.[card.id] || 0,
    }));

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200); // Match fadeOut animation duration
  };

  const handleCardClick = (cardId: string) => {
    onCardSelect(cardId);
  };

  const handleContinueWithoutStorage = () => {
    // Play the card without selecting any storage (resources will be lost)
    onCardSelect(""); // Empty string means no storage target
  };

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleCancelClick();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      // Only close on left click (button 0), ignore right click (button 2) and middle click (button 1)
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
      // Prevent body scroll
      document.body.style.overflow = "hidden";

      // Add event listeners
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      // Restore body scroll
      document.body.style.overflow = "";

      // Remove event listeners
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel]);

  if (!isVisible) return null;

  const hasNoStorage = validCards.length === 0;

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-auto overflow-hidden"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className={`
          min-w-[240px] w-fit max-w-[90vw] max-h-[500px]
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
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            {hasNoStorage ? "No Storage Available" : "Select Card Storage"}
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
        </div>

        {/* Cards Container or Warning */}
        <div className="flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-space-blue-500/50 scrollbar-track-space-blue-900/30">
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
                  onClick={() => handleCardClick(card.id)}
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
        </div>

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center gap-3">
          {hasNoStorage ? (
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
                onClick={handleContinueWithoutStorage}
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

        /* Media queries */
        @media (max-width: 768px) {
          .min-w-\\[240px\\] {
            min-width: 180px;
          }
          .max-w-\\[90vw\\] {
            max-width: 95vw;
          }
        }
      `}</style>
    </div>
  );
};

export default CardStorageSelectionPopover;
