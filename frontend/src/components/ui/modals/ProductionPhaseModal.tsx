import React, { useState, useEffect, useMemo, useCallback, useRef } from "react";
import {
  GameDto,
  ProductionPhaseDto,
  OtherPlayerDto,
  PlayerDto,
  ResourceType,
} from "@/types/generated/api-types.ts";
import { RESOURCE_COLORS, RESOURCE_NAMES } from "@/utils/resourceColors.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import ProductionCardSelectionOverlay from "@/components/ui/overlay/ProductionCardSelectionOverlay.tsx";
import GameIcon from "@/components/ui/display/GameIcon.tsx";
import { GameModal, GameModalContent } from "../GameModal";

interface ProductionPhaseModalProps {
  isOpen: boolean;
  gameState: GameDto | null;
  onClose: () => void;
  onHide?: () => void;
  openDirectlyToCardSelection?: boolean;
}

const ProductionPhaseModal: React.FC<ProductionPhaseModalProps> = ({
  isOpen,
  gameState,
  onClose,
  onHide,
  openDirectlyToCardSelection = false,
}) => {
  const [hasSubmittedCardSelection, setHasSubmittedCardSelection] = useState(false);
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [animationStep, setAnimationStep] = useState<"energyConversion" | "production">(
    "energyConversion",
  );
  const [isAnimating, setIsAnimating] = useState(false);
  const [resourceAnimationState, setResourceAnimationState] = useState<
    "initial" | "fadeInResources" | "showProduction" | "fadeOut" | "fadeIn"
  >("initial");
  const [energyAnimationState, setEnergyAnimationState] = useState<
    "initial" | "fadeOut" | "fadeIn"
  >("initial");
  const [showCardSelection, setShowCardSelection] = useState(false);
  const prevSelectionCompleteRef = useRef<boolean | undefined>(undefined);
  const prevIsOpenRef = useRef(false);

  const handleCardSelection = useCallback(
    async (selectedCardIds: string[]) => {
      try {
        await globalWebSocketManager.confirmProductionCards(selectedCardIds);
        setHasSubmittedCardSelection(true);
        setShowCardSelection(false);
      } catch (error) {
        console.error("Failed to submit card selection:", error);
        onClose();
      }
    },
    [onClose],
  );

  const handleReturnFromCardSelection = useCallback(() => {
    if (onHide) {
      onHide();
    }
  }, [onHide]);

  useEffect(() => {
    const selectionComplete = gameState?.currentPlayer?.productionPhase?.selectionComplete;
    const prev = prevSelectionCompleteRef.current;
    const wasOpen = prevIsOpenRef.current;

    const selectionTransitioned =
      selectionComplete === false && (prev === undefined || prev === true);
    const modalJustOpened = isOpen && !wasOpen && selectionComplete === false;

    if (isOpen && (selectionTransitioned || modalJustOpened)) {
      setHasSubmittedCardSelection(false);
      setShowCardSelection(openDirectlyToCardSelection);
    }

    prevSelectionCompleteRef.current = selectionComplete;
    prevIsOpenRef.current = isOpen;
  }, [
    isOpen,
    gameState?.currentPlayer?.productionPhase?.selectionComplete,
    openDirectlyToCardSelection,
  ]);

  const modalProductionData = useMemo(() => {
    if (!gameState || !gameState.currentPlayer?.productionPhase) {
      return null;
    }

    const allPlayers: (PlayerDto | OtherPlayerDto)[] = [
      gameState.currentPlayer,
      ...gameState.otherPlayers,
    ];

    const playersWithProduction = allPlayers.filter((player) => player.productionPhase);

    if (playersWithProduction.length === 0) {
      return null;
    }

    const playersData = playersWithProduction.map((player) => {
      const productionPhase = player.productionPhase as ProductionPhaseDto;
      if ("cards" in player) {
        const currentPlayer = player as PlayerDto;
        return {
          playerId: currentPlayer.id,
          playerName: currentPlayer.name,
          production: currentPlayer.production,
          terraformRating: currentPlayer.terraformRating,
          ...productionPhase,
        };
      } else {
        const otherPlayer = player as OtherPlayerDto;
        return {
          playerId: otherPlayer.id,
          playerName: otherPlayer.name,
          production: otherPlayer.production,
          terraformRating: otherPlayer.terraformRating,
          ...productionPhase,
        };
      }
    });

    return {
      playersData,
      generation: gameState.generation || 1,
    };
  }, [gameState]);

  const resourceNames = RESOURCE_NAMES;

  useEffect(() => {
    if (modalProductionData && modalProductionData.playersData.length > 0) {
      const currentPlayerData = modalProductionData.playersData[currentPlayerIndex];
      const hasEnergyToConvert = currentPlayerData.energyConverted > 0;

      if (hasEnergyToConvert) {
        setAnimationStep("energyConversion");
      } else {
        setAnimationStep("production");
      }
    }
  }, [modalProductionData, currentPlayerIndex]);

  useEffect(() => {
    if (!isAnimating) return;

    const timer = setTimeout(
      () => {
        if (animationStep === "energyConversion") {
          setAnimationStep("production");
        } else {
          setIsAnimating(false);
        }
      },
      animationStep === "energyConversion" ? 4500 : 4000,
    );

    return () => clearTimeout(timer);
  }, [currentPlayerIndex, animationStep, isAnimating]);

  const handlePlayerSelect = (playerIndex: number) => {
    if (playerIndex !== currentPlayerIndex && modalProductionData) {
      setCurrentPlayerIndex(playerIndex);

      const playerData = modalProductionData.playersData[playerIndex];
      const hasEnergyToConvert = playerData.energyConverted > 0;

      if (hasEnergyToConvert) {
        setAnimationStep("energyConversion");
      } else {
        setAnimationStep("production");
      }

      setIsAnimating(true);
      setResourceAnimationState("initial");
      setEnergyAnimationState("initial");
    }
  };

  useEffect(() => {
    if (!isAnimating || animationStep !== "energyConversion") return;

    let timeoutId: NodeJS.Timeout;

    if (energyAnimationState === "initial") {
      timeoutId = setTimeout(() => {
        setEnergyAnimationState("fadeOut");
      }, 2000);
    } else if (energyAnimationState === "fadeOut") {
      timeoutId = setTimeout(() => {
        setEnergyAnimationState("fadeIn");
      }, 400);
    }

    return () => clearTimeout(timeoutId);
  }, [energyAnimationState, animationStep, isAnimating]);

  useEffect(() => {
    if (!isOpen) return;
    setCurrentPlayerIndex(0);
    setAnimationStep("energyConversion");
    setIsAnimating(true);
    setResourceAnimationState("initial");
    setEnergyAnimationState("initial");
  }, [isOpen]);

  useEffect(() => {
    if (!isAnimating || animationStep !== "production") return;

    let timeoutId: NodeJS.Timeout;

    if (resourceAnimationState === "initial") {
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeInResources");
      }, 500);
    } else if (resourceAnimationState === "fadeInResources") {
      timeoutId = setTimeout(() => {
        setResourceAnimationState("showProduction");
      }, 500);
    } else if (resourceAnimationState === "showProduction") {
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeOut");
      }, 1500);
    } else if (resourceAnimationState === "fadeOut") {
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeIn");
      }, 400);
    }

    return () => clearTimeout(timeoutId);
  }, [resourceAnimationState, animationStep, isAnimating]);

  useEffect(() => {
    if (animationStep === "production") {
      setResourceAnimationState("initial");
    }
  }, [currentPlayerIndex, animationStep]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Enter" && !hasSubmittedCardSelection && !showCardSelection) {
        setShowCardSelection(true);
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }

    return () => {};
  }, [isOpen, hasSubmittedCardSelection, showCardSelection]);

  if (!isOpen) return null;
  if (!modalProductionData) return null;

  const currentPlayerData = modalProductionData.playersData[currentPlayerIndex];
  if (!currentPlayerData) return null;

  const renderResourceAnimation = (
    resourceType: ResourceType,
    beforeAmount: number,
    afterAmount: number,
  ) => {
    let displayBeforeAmount = beforeAmount;
    let displayAfterAmount = afterAmount;
    let change = afterAmount - beforeAmount;
    let shouldAnimate;

    if (animationStep === "energyConversion") {
      if (resourceType === "energy") {
        displayBeforeAmount = currentPlayerData.beforeResources.energy;
        displayAfterAmount = 0;
        change = -currentPlayerData.beforeResources.energy;
        shouldAnimate = true;
      } else if (resourceType === "heat") {
        displayBeforeAmount = currentPlayerData.beforeResources.heat;
        displayAfterAmount =
          currentPlayerData.beforeResources.heat + currentPlayerData.energyConverted;
        change = currentPlayerData.energyConverted;
        shouldAnimate = true;
      } else {
        shouldAnimate = false;
      }
    } else {
      if (resourceType === "energy") {
        displayBeforeAmount = 0;
        displayAfterAmount = currentPlayerData.production.energy;
        change = currentPlayerData.production.energy;
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      } else if (resourceType === "heat") {
        displayBeforeAmount =
          currentPlayerData.beforeResources.heat + currentPlayerData.energyConverted;
        displayAfterAmount = currentPlayerData.afterResources.heat;
        change = displayAfterAmount - displayBeforeAmount;
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      } else {
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      }
    }

    const getAnimationStateClass = () => {
      switch (resourceAnimationState) {
        case "initial":
          return "";
        case "fadeOut":
          return "[&_.beforeAmount]:animate-[fadeOutAnimation_0.4s_ease-out_forwards] [&_.changeIndicator]:animate-[fadeOutAnimation_0.4s_ease-out_forwards]";
        case "fadeIn":
          return "[&_.finalValue]:text-white [&_.finalValue]:[text-shadow:0_1px_3px_rgba(0,0,0,0.8)] [&_.finalValue]:animate-[fadeInAnimation_0.3s_ease-out_forwards]";
        default:
          return "";
      }
    };

    const renderAmounts = () => {
      if (
        animationStep === "production" &&
        (resourceAnimationState === "initial" || resourceAnimationState === "fadeInResources")
      ) {
        return (
          <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px]">
            <div
              className={`beforeAmount text-white/70 ${displayBeforeAmount < 0 ? "!text-red-500 [text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
            >
              {displayBeforeAmount}
            </div>
          </div>
        );
      }

      if (
        animationStep === "production" &&
        resourceAnimationState === "showProduction" &&
        change !== 0
      ) {
        return (
          <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px]">
            <div
              className={`beforeAmount text-white/70 ${displayBeforeAmount < 0 ? "!text-red-500 [text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
            >
              {displayBeforeAmount}
            </div>
            <div
              className={`changeIndicator text-sm font-bold text-green-400 [text-shadow:0_0_10px_#4ade80] animate-[fadeInUp_0.5s_ease-out_0.5s_both] ${change < 0 ? "!text-red-500 ![text-shadow:0_0_10px_#ef4444]" : ""}`}
            >
              {change > 0 ? `+${change}` : change}
            </div>
          </div>
        );
      }

      if (animationStep === "energyConversion" && shouldAnimate) {
        if (energyAnimationState === "fadeIn") {
          return (
            <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px] [&_.finalValue]:text-white [&_.finalValue]:[text-shadow:0_1px_3px_rgba(0,0,0,0.8)] [&_.finalValue]:animate-[fadeInAnimation_0.3s_ease-out_forwards]">
              <div
                className={`finalValue text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] ${displayAfterAmount < 0 ? "!text-red-500 ![text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
              >
                {displayAfterAmount}
              </div>
            </div>
          );
        } else if (energyAnimationState === "fadeOut") {
          return (
            <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px] [&_.beforeAmount]:animate-[fadeOutAnimation_0.4s_ease-out_forwards] [&_.changeIndicator]:animate-[fadeOutAnimation_0.4s_ease-out_forwards]">
              <div
                className={`beforeAmount text-white/70 ${displayBeforeAmount < 0 ? "!text-red-500 [text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
              >
                {displayBeforeAmount}
              </div>
              <div
                className={`changeIndicator text-sm font-bold text-green-400 [text-shadow:0_0_10px_#4ade80] ${change < 0 ? "!text-red-500 ![text-shadow:0_0_10px_#ef4444]" : ""}`}
              >
                {change > 0 ? `+${change}` : change}
              </div>
            </div>
          );
        } else {
          return (
            <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px]">
              <div
                className={`beforeAmount text-white/70 ${displayBeforeAmount < 0 ? "!text-red-500 [text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
              >
                {displayBeforeAmount}
              </div>
              <div
                className={`changeIndicator text-sm font-bold text-green-400 [text-shadow:0_0_10px_#4ade80] ${change < 0 ? "!text-red-500 ![text-shadow:0_0_10px_#ef4444]" : ""}`}
              >
                {change > 0 ? `+${change}` : change}
              </div>
            </div>
          );
        }
      }

      if (!shouldAnimate) {
        return (
          <div className="flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px]">
            <div
              className={`beforeAmount text-white/70 ${displayBeforeAmount < 0 ? "!text-red-500 [text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
            >
              {displayBeforeAmount}
            </div>
          </div>
        );
      }

      if (resourceAnimationState === "fadeIn") {
        return (
          <div
            className={`flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px] ${getAnimationStateClass()}`}
          >
            <div
              className={`finalValue text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] ${displayAfterAmount < 0 ? "!text-red-500 ![text-shadow:0_1px_3px_rgba(0,0,0,0.8),0_0_8px_#ef4444]" : ""}`}
            >
              {displayAfterAmount}
            </div>
          </div>
        );
      }

      return (
        <div
          className={`flex items-center justify-center gap-2 mb-2 text-base font-bold relative min-h-[24px] ${getAnimationStateClass()}`}
        >
          <div className="beforeAmount text-white/70">{displayBeforeAmount}</div>
          <div
            className={`changeIndicator text-sm font-bold text-green-400 [text-shadow:0_0_10px_#4ade80] ${change < 0 ? "!text-red-500 ![text-shadow:0_0_10px_#ef4444]" : ""}`}
          >
            {change > 0 ? `+${change}` : change}
          </div>
        </div>
      );
    };

    return (
      <div
        key={resourceType}
        className={`bg-[linear-gradient(135deg,rgba(0,0,0,0.4)_0%,rgba(0,0,0,0.2)_100%)] border-2 border-[var(--resource-color,rgba(255,255,255,0.2))] rounded-xl p-5 text-center transition-all duration-300 relative overflow-hidden before:content-[''] before:absolute before:top-0 before:left-0 before:right-0 before:bottom-0 before:bg-[var(--resource-color,transparent)] before:opacity-10 before:pointer-events-none ${shouldAnimate ? "!border-[var(--resource-color,#00d4ff)] scale-[1.02]" : ""} ${
          (animationStep === "energyConversion" && !shouldAnimate) ||
          (animationStep === "production" &&
            resourceAnimationState === "initial" &&
            !shouldAnimate &&
            resourceType !== "energy" &&
            resourceType !== "heat")
            ? "opacity-30"
            : ""
        } ${
          animationStep === "energyConversion" &&
          resourceType === "energy" &&
          energyAnimationState === "initial"
            ? "[&_.resourceIcon_img]:animate-[energyShake_0.1s_ease-in-out_infinite_alternate,energyGlow_2s_ease-in-out_infinite_alternate] [&_.resourceIcon_img]:[filter:drop-shadow(0_0_15px_#ef4444)_drop-shadow(0_0_30px_#ef4444)]"
            : ""
        }`}
        style={
          {
            "--resource-color": RESOURCE_COLORS[resourceType],
          } as React.CSSProperties
        }
      >
        <div className="resourceIcon mb-3 flex justify-center">
          <GameIcon iconType={resourceType} size="medium" />
        </div>
        {renderAmounts()}
        <div className="text-xs text-white/80 uppercase font-semibold tracking-[0.5px]">
          {resourceNames[resourceType]}
        </div>
      </div>
    );
  };

  const renderProductionPhase = () => {
    return (
      <div className="w-full">
        <div className="grid grid-cols-3 gap-5 mb-5">
          {renderResourceAnimation(
            "credit" as ResourceType,
            currentPlayerData.beforeResources.credits,
            currentPlayerData.afterResources.credits,
          )}
          {renderResourceAnimation(
            "steel" as ResourceType,
            currentPlayerData.beforeResources.steel,
            currentPlayerData.afterResources.steel,
          )}
          {renderResourceAnimation(
            "titanium" as ResourceType,
            currentPlayerData.beforeResources.titanium,
            currentPlayerData.afterResources.titanium,
          )}
          {renderResourceAnimation(
            "plant" as ResourceType,
            currentPlayerData.beforeResources.plants,
            currentPlayerData.afterResources.plants,
          )}
          {renderResourceAnimation(
            "energy" as ResourceType,
            currentPlayerData.beforeResources.energy,
            currentPlayerData.afterResources.energy,
          )}
          {renderResourceAnimation(
            "heat" as ResourceType,
            currentPlayerData.beforeResources.heat,
            currentPlayerData.afterResources.heat,
          )}
        </div>
      </div>
    );
  };

  const nextButton =
    !hasSubmittedCardSelection && !showCardSelection ? (
      <button
        className="ml-5 flex-shrink-0 bg-[linear-gradient(135deg,rgba(30,60,150,0.8)_0%,rgba(20,40,120,0.9)_100%)] border-2 border-space-blue-400 rounded-full text-white text-[32px] font-bold w-[60px] h-[60px] cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_15px_rgba(0,0,0,0.4)] flex items-center justify-center p-0 hover:bg-[linear-gradient(135deg,rgba(40,70,160,0.9)_0%,rgba(30,50,130,1)_100%)] hover:border-space-blue-500 hover:shadow-[0_6px_20px_rgba(0,0,0,0.5)] active:scale-95 active:shadow-[0_2px_10px_rgba(0,0,0,0.3)]"
        onClick={() => setShowCardSelection(true)}
      >
        →
      </button>
    ) : null;

  return (
    <>
      <GameModal
        isVisible={isOpen && !showCardSelection}
        onClose={() => {}}
        theme="production"
        size="medium"
        preventClose
        closeOnBackdrop={false}
        closeOnEscape={false}
        className="!max-w-[800px] !min-w-[600px]"
        outerContent={nextButton}
      >
        <div className="text-center py-[30px] px-[30px] pb-5 bg-black/40 border-b border-[var(--modal-accent)]/60 relative">
          <h2 className="text-[28px] font-orbitron text-white font-bold text-shadow-glow tracking-wider m-0 mb-2">
            Production
          </h2>
          <div className="text-base text-[var(--modal-accent)] font-semibold text-shadow-dark">
            Generation {modalProductionData.generation}
          </div>
        </div>

        {modalProductionData.playersData.length > 1 && (
          <div className="flex justify-center gap-3 p-5 border-b border-[var(--modal-accent)]/30">
            {modalProductionData.playersData.map((player, index) => (
              <button
                key={player.playerId}
                className={`bg-space-black-darker/60 border-2 border-[var(--modal-accent)]/40 rounded-lg text-white/80 text-sm font-semibold py-2 px-4 cursor-pointer transition-all duration-300 text-shadow-dark hover:border-[var(--modal-accent)]/60 hover:text-white/90 hover:-translate-y-px ${
                  index === currentPlayerIndex
                    ? "!bg-[var(--modal-accent)]/20 !border-[var(--modal-accent)] !text-white shadow-[0_0_15px_rgba(var(--modal-accent-rgb),0.4)]"
                    : ""
                }`}
                onClick={() => handlePlayerSelect(index)}
              >
                {player.playerName}
              </button>
            ))}
          </div>
        )}

        <GameModalContent padding="large">
          <div className="min-h-[300px] flex items-center justify-center">
            {renderProductionPhase()}
          </div>
        </GameModalContent>
      </GameModal>

      {showCardSelection && (
        <ProductionCardSelectionOverlay
          isOpen={showCardSelection}
          cards={gameState?.currentPlayer?.productionPhase?.availableCards || []}
          playerCredits={gameState?.currentPlayer?.resources.credits || 0}
          onSelectCards={handleCardSelection}
          onReturn={handleReturnFromCardSelection}
        />
      )}
    </>
  );
};

export default ProductionPhaseModal;
