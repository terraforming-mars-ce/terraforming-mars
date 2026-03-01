import React, { useState } from "react";
import CorporationCard from "../cards/CorporationCard.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
} from "./overlayStyles.ts";
import GameMenuButton from "../buttons/GameMenuButton.tsx";

interface CorporationSelectionOverlayProps {
  isOpen: boolean;
  availableCorporations: CardDto[];
  onSelectCorporation: (corporationId: string) => void;
}

const CorporationSelectionOverlay: React.FC<CorporationSelectionOverlayProps> = ({
  isOpen,
  availableCorporations,
  onSelectCorporation,
}) => {
  const [selectedCorporationId, setSelectedCorporationId] = useState<string | null>(null);

  if (!isOpen || availableCorporations.length === 0) return null;

  const handleConfirm = () => {
    if (!selectedCorporationId) return;
    onSelectCorporation(selectedCorporationId);
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      <div className={OVERLAY_CONTAINER_CLASS}>
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Select Your Corporation</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>Choose your corporation to begin the game</p>
        </div>

        <div className="flex-1 overflow-y-auto p-8 bg-black/20 max-[768px]:p-5">
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
                  onSelect={setSelectedCorporationId}
                  borderColor={getCorporationBorderColor(corp.name)}
                />
              </div>
            ))}
          </div>
        </div>

        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-sm text-white/70">
            {selectedCorporationId ? "Corporation selected" : "Please select a corporation"}
          </div>
          <GameMenuButton
            variant="primary"
            size="lg"
            onClick={handleConfirm}
            disabled={!selectedCorporationId}
            className="whitespace-nowrap max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
          >
            Confirm Corporation
          </GameMenuButton>
        </div>
      </div>
    </div>
  );
};

export default CorporationSelectionOverlay;
