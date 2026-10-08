import React, { useState, useCallback, useMemo } from "react";
import CorporationCard from "../cards/CorporationCard.tsx";
import CardChoice from "../cards/CardChoice.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_CONTAINER_STYLE,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_FOOTER_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
  OVERLAY_ACTION_BUTTON_CLASS,
  OVERLAY_FOOTER_LEFT_CLASS,
  OVERLAY_ROOT_SAFE_AREA_CLASS,
} from "./overlayStyles.ts";
import OverlayDescription from "./OverlayDescription.tsx";
import GameButton from "../buttons/GameButton.tsx";
import StartingSelectionSteps from "./StartingSelectionSteps.tsx";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

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
  const { isCompact } = useLayoutMode();
  const [selectedCorporationId, setSelectedCorporationId] = useState<string | null>(null);
  const [selectedPreludeIds, setSelectedPreludeIds] = useState<string[]>([]);
  const [stepIndex, setStepIndex] = useState(0);
  const [maxVisitedIndex, setMaxVisitedIndex] = useState(0);

  const handleStepIndexChange = (index: number) => {
    setStepIndex(index);
    setMaxVisitedIndex((visited) => Math.max(visited, index));
  };

  const selectedCorp = useMemo(
    () => availableCorporations.find((c) => c.id === selectedCorporationId),
    [availableCorporations, selectedCorporationId],
  );

  const effectiveCredits = useMemo(() => {
    if (!selectedCorp) {
      return playerCredits;
    }
    if (selectedCorp.startingResources?.credits != null) {
      return selectedCorp.startingResources.credits;
    }
    let credits = 0;
    for (const behavior of selectedCorp.behaviors ?? []) {
      const isCorpStart = behavior.triggers?.some((t) => t.type === "auto-corporation-start");
      if (isCorpStart) {
        for (const output of behavior.outputs ?? []) {
          if (output.type === "credit") {
            credits += output.amount;
          }
        }
      }
    }
    return credits || playerCredits;
  }, [selectedCorp, playerCredits]);

  const costPerCard = useMemo(() => {
    if (!selectedCorp?.behaviors) {
      return 3;
    }
    let discount = 0;
    for (const behavior of selectedCorp.behaviors) {
      for (const output of behavior.outputs ?? []) {
        if (output.type === "discount" && output.selectors) {
          for (const sel of output.selectors) {
            if (sel.actions?.includes("card-buying")) {
              discount += output.amount;
            }
          }
        }
      }
    }
    return Math.max(3 - discount, 0);
  }, [selectedCorp]);

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
    costPerCard,
    minCards: 0,
  });

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
    if (!selectedCorporationId) {
      return;
    }

    handleCardConfirm((cardIds) => {
      onConfirm(selectedCorporationId, selectedPreludeIds, cardIds);
    });
  };

  if (!isOpen) {
    return null;
  }

  if (isCompact) {
    return (
      <div
        className={`fixed inset-0 flex items-center justify-center animate-[fadeIn_0.3s_ease] ${OVERLAY_ROOT_SAFE_AREA_CLASS}`}
        style={{ zIndex: Z_INDEX.CORPORATION_SELECTION }}
      >
        <StartingSelectionSteps
          availableCorporations={availableCorporations}
          availablePreludes={availablePreludes}
          maxSelectablePreludes={maxSelectablePreludes}
          cards={cards}
          selectedCorporation={selectedCorp}
          onSelectCorporation={setSelectedCorporationId}
          selectedPreludeIds={selectedPreludeIds}
          onSelectPrelude={handlePreludeSelect}
          selectedCardIds={selectedCardIds}
          onSelectCard={handleCardSelect}
          credits={effectiveCredits}
          totalCost={totalCost}
          preludesValid={preludesValid}
          cardsValid={isValidCardSelection}
          allValid={allValid}
          showConfirmation={showConfirmation}
          onConfirm={handleConfirm}
          onHide={onHide}
          stepIndex={stepIndex}
          maxVisitedIndex={maxVisitedIndex}
          onStepIndexChange={handleStepIndexChange}
        />
      </div>
    );
  }

  return (
    <div
      className={`fixed inset-0 flex items-center justify-center animate-[fadeIn_0.3s_ease] ${OVERLAY_ROOT_SAFE_AREA_CLASS}`}
      style={{ zIndex: Z_INDEX.CORPORATION_SELECTION }}
    >
      <div className={OVERLAY_CONTAINER_CLASS} style={OVERLAY_CONTAINER_STYLE}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Starting Cards</h2>
          <OverlayDescription>
            Choose your corporation
            {hasPreludes ? ", prelude cards," : ""} and starting project cards. Each project card
            costs {costPerCard} MC.
          </OverlayDescription>
        </div>

        <div className="flex-1 overflow-y-auto p-8 bg-black/20">
          {availableCorporations.length > 0 && (
            <div>
              <h3 className="text-white/60 text-sm font-orbitron font-bold uppercase tracking-widest mb-4">
                Corporation
              </h3>
              <div className="flex gap-8 justify-center flex-wrap">
                {availableCorporations.map((corp) => (
                  <div key={corp.id} className="w-[400px]">
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
                <div className="flex gap-6 justify-center flex-wrap">
                  {availablePreludes.map((card, index) => (
                    <CardChoice
                      key={card.id}
                      card={card}
                      isSelected={selectedPreludeIds.includes(card.id)}
                      disabled={
                        selectedPreludeIds.length >= maxSelectablePreludes &&
                        !selectedPreludeIds.includes(card.id)
                      }
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
              className="grid gap-x-6 gap-y-14 justify-center py-6"
              style={{
                gridTemplateColumns: `repeat(${Math.ceil(cards.length / Math.ceil(cards.length / 6))}, max-content)`,
              }}
            >
              {cards.map((card, index) => {
                const isSelected = selectedCardIds.includes(card.id);

                return (
                  <CardChoice
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
          <div className={OVERLAY_FOOTER_LEFT_CLASS}>
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

          <div className="flex items-center gap-4">
            {onHide && (
              <GameButton emphasis="secondary" size="lg" onClick={onHide}>
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
              className={OVERLAY_ACTION_BUTTON_CLASS}
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
