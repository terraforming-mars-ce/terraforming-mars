import React, { useState, useCallback, useEffect, useMemo } from "react";
import CorporationCard from "../cards/CorporationCard.tsx";
import GameCard from "../cards/GameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
} from "./overlayStyles.ts";
import GameButton from "../buttons/GameButton.tsx";
import MainMenuSettingsButton from "../buttons/MainMenuSettingsButton.tsx";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  availableCorporations: CardDto[];
  availablePreludes: CardDto[];
  maxSelectablePreludes: number;
  cards: CardDto[];
  playerCredits: number;
  onConfirm: (corporationId: string, preludeIds: string[], cardIds: string[]) => void;
  onHide?: () => void;
}

const StartingCardSelectionOverlay: React.FC<StartingCardSelectionOverlayProps> = ({
  isOpen,
  availableCorporations,
  availablePreludes,
  maxSelectablePreludes,
  cards,
  playerCredits,
  onConfirm,
  onHide,
}) => {
  const [selectedCorporationId, setSelectedCorporationId] = useState<string | null>(null);
  const [selectedPreludeIds, setSelectedPreludeIds] = useState<string[]>([]);

  const selectedCorp = useMemo(
    () => availableCorporations.find((c) => c.id === selectedCorporationId),
    [availableCorporations, selectedCorporationId],
  );

  const effectiveCredits = selectedCorp?.startingResources?.credits ?? playerCredits;

  const {
    selectedCardIds,
    totalCost,
    showConfirmation,
    isValidSelection: isValidCardSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
  } = useCardSelection({
    cards,
    isOpen,
    playerCredits: effectiveCredits,
    costPerCard: 3,
    minCards: 0,
  });

  useEffect(() => {
    if (isOpen) {
      setSelectedCorporationId(null);
      setSelectedPreludeIds([]);
    }
  }, [isOpen]);

  const handlePreludeSelect = useCallback(
    (cardId: string) => {
      setSelectedPreludeIds((prev) => {
        if (prev.includes(cardId)) {
          return prev.filter((id) => id !== cardId);
        }
        if (prev.length >= maxSelectablePreludes) {
          return prev;
        }
        return [...prev, cardId];
      });
    },
    [maxSelectablePreludes],
  );

  const hasPreludes = availablePreludes.length > 0;
  const preludesValid = !hasPreludes || selectedPreludeIds.length === maxSelectablePreludes;
  const allValid = !!selectedCorporationId && preludesValid && isValidCardSelection;

  const handleConfirm = () => {
    if (!selectedCorporationId) return;

    handleCardConfirm((cardIds) => {
      onConfirm(selectedCorporationId, selectedPreludeIds, cardIds);
    });
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <MainMenuSettingsButton />
      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Starting Cards</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Choose your corporation
            {hasPreludes ? ", prelude cards," : ""} and starting project cards. Each project card
            costs 3 MC.
          </p>
        </div>

        <div className="flex-1 overflow-y-auto p-8 bg-black/20 max-[768px]:p-5">
          {availableCorporations.length > 0 && (
            <div>
              <h3 className="text-white/60 text-sm font-orbitron font-bold uppercase tracking-widest mb-4">
                Corporation
              </h3>
              <div className="flex gap-4 justify-center flex-wrap">
                {availableCorporations.map((corp) => (
                  <div key={corp.id} className="w-[400px] max-[768px]:w-full">
                    <CorporationCard
                      card={corp}
                      isSelected={selectedCorporationId === corp.id}
                      onSelect={setSelectedCorporationId}
                      borderColor={getCorporationBorderColor(corp.name)}
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          {hasPreludes && (
            <>
              <div className="border-t border-white/10 my-6" />
              <div>
                <h3 className="text-white/60 text-sm font-orbitron font-bold uppercase tracking-widest mb-4">
                  Prelude Cards
                  <span className="ml-2 text-white/40 text-xs normal-case">
                    {selectedPreludeIds.length} / {maxSelectablePreludes} selected
                  </span>
                </h3>
                <div className="flex gap-6 justify-center flex-wrap max-[768px]:gap-4">
                  {availablePreludes.map((card, index) => (
                    <GameCard
                      key={card.id}
                      card={card}
                      isSelected={selectedPreludeIds.includes(card.id)}
                      onSelect={handlePreludeSelect}
                      animationDelay={index * 100}
                      showCheckbox={true}
                    />
                  ))}
                </div>
              </div>
            </>
          )}

          <div className="border-t border-white/10 my-6" />

          <div>
            <h3 className="text-white/60 text-sm font-orbitron font-bold uppercase tracking-widest mb-4">
              Project Cards
            </h3>
            <div
              className="grid gap-x-6 gap-y-14 justify-center py-6 max-[768px]:gap-x-4 max-[768px]:gap-y-8"
              style={{
                gridTemplateColumns: `repeat(${Math.ceil(cards.length / Math.ceil(cards.length / 6))}, max-content)`,
              }}
            >
              {cards.map((card, index) => {
                const isSelected = selectedCardIds.includes(card.id);

                return (
                  <GameCard
                    key={card.id}
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={true}
                  />
                );
              })}
            </div>
          </div>
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between">
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
              <GameIcon iconType={ResourceTypeCredit} amount={effectiveCredits} size="large" />
            </div>
            <div className={RESOURCE_DISPLAY_CLASS}>
              <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
              {totalCost > 0 ? (
                <>
                  <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="large" />
                  <span className="text-white/70 text-sm">
                    ({selectedCardIds.length} card
                    {selectedCardIds.length !== 1 ? "s" : ""} selected)
                  </span>
                </>
              ) : (
                <span className="!text-[#4caf50] font-bold tracking-[1px]">FREE</span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-4 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            {onHide && (
              <GameButton buttonType="secondary" size="lg" onClick={onHide}>
                Hide
              </GameButton>
            )}

            {!selectedCorporationId && (
              <span className="text-sm text-white/70">Select a corporation to continue</span>
            )}

            {showConfirmation && (
              <div className="text-sm">
                <span className="text-[#ff9800]">
                  Are you sure you don't want to select any cards?
                </span>
              </div>
            )}

            <GameButton
              size="lg"
              onClick={handleConfirm}
              disabled={!allValid}
              className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
            >
              {showConfirmation ? "Confirm Skip" : "Confirm Selection"}
            </GameButton>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;
