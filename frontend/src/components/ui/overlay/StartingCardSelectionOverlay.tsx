import React, { useEffect, useState } from "react";
import GameCard from "../cards/GameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
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
import GameMenuButton from "../buttons/GameMenuButton.tsx";
import MainMenuSettingsButton from "../buttons/MainMenuSettingsButton.tsx";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  availableCorporations: CardDto[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[], corporationId: string) => void;
}

const StartingCardSelectionOverlay: React.FC<StartingCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  availableCorporations,
  playerCredits,
  onSelectCards,
}) => {
  const [selectedCorporationId, setSelectedCorporationId] = useState<string | null>(null);

  const {
    selectedCardIds,
    totalCost,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
  } = useCardSelection({
    cards,
    isOpen,
    playerCredits,
    costPerCard: 3,
    minCards: 0,
  });

  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCorporationId(null);
    }
  }, [isOpen, cards]);

  if (!isOpen || cards.length === 0) return null;

  const handleCorporationSelect = (corporationId: string) => {
    setSelectedCorporationId(corporationId);
  };

  const handleConfirm = () => {
    if (!selectedCorporationId) {
      return;
    }

    handleCardConfirm((cardIds) => {
      onSelectCards(cardIds, selectedCorporationId);
    });
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <MainMenuSettingsButton />
      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Starting Cards</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Choose your corporation and starting project cards. Each project card costs 3 MC.
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
                      corporation={{
                        id: corp.id,
                        name: corp.name,
                        description: corp.description,
                        startingMegaCredits: corp.startingResources?.credits ?? 0,
                        startingProduction: corp.startingProduction
                          ? {
                              credits: corp.startingProduction.credits,
                              steel: corp.startingProduction.steel,
                              titanium: corp.startingProduction.titanium,
                              plants: corp.startingProduction.plants,
                              energy: corp.startingProduction.energy,
                              heat: corp.startingProduction.heat,
                            }
                          : undefined,
                        startingResources: corp.startingResources
                          ? {
                              credits: corp.startingResources.credits,
                              steel: corp.startingResources.steel,
                              titanium: corp.startingResources.titanium,
                              plants: corp.startingResources.plants,
                              energy: corp.startingResources.energy,
                              heat: corp.startingResources.heat,
                            }
                          : undefined,
                        behaviors: corp.behaviors,
                        logoPath: undefined,
                      }}
                      isSelected={selectedCorporationId === corp.id}
                      onSelect={handleCorporationSelect}
                      borderColor={getCorporationBorderColor(corp.name)}
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="border-t border-white/10 my-6" />

          <div>
            <h3 className="text-white/60 text-sm font-orbitron font-bold uppercase tracking-widest mb-4">
              Project Cards
            </h3>
            <div className="flex gap-6 justify-center flex-wrap max-[768px]:gap-4">
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
              <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
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

            <GameMenuButton
              variant="primary"
              size="lg"
              onClick={handleConfirm}
              disabled={!selectedCorporationId || !isValidSelection}
              className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
            >
              {showConfirmation ? "Confirm Skip" : "Confirm Selection"}
            </GameMenuButton>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;
