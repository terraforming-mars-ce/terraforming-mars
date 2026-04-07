import React, { useState, useEffect, useMemo, useCallback, useRef } from "react";
import {
  GameDto,
  ProductionPhaseDto,
  OtherPlayerDto,
  PlayerDto,
  ResourceType,
  ResourcesDto,
  ProductionDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import ProductionCardSelectionOverlay from "@/components/ui/overlay/ProductionCardSelectionOverlay.tsx";
import GameIcon from "@/components/ui/display/GameIcon.tsx";
import GameButton from "@/components/ui/buttons/GameButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { audioService } from "@/services/audioService.ts";
import {
  OVERLAY_BACKDROP_BLUR_CLASS,
  OVERLAY_BACKDROP_TINT_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
} from "@/components/ui/overlay/overlayStyles.ts";

interface ProductionPhaseModalProps {
  isOpen: boolean;
  gameState: GameDto | null;
  onClose: () => void;
  onHide?: () => void;
  openDirectlyToCardSelection?: boolean;
}

type AnimationPhase = "initial" | "energyTransfer" | "productionTransfer" | "final";

const RESOURCE_KEYS: { key: ResourceType; resField: keyof ResourcesDto }[] = [
  { key: "credit", resField: "credits" },
  { key: "steel", resField: "steel" },
  { key: "titanium", resField: "titanium" },
  { key: "plant", resField: "plants" },
  { key: "energy", resField: "energy" },
  { key: "heat", resField: "heat" },
];

const RESOURCE_ICON_TYPES: Record<string, string> = {
  credit: ResourceTypeCredit,
  steel: ResourceTypeSteel,
  titanium: ResourceTypeTitanium,
  plant: ResourceTypePlant,
  energy: ResourceTypeEnergy,
  heat: ResourceTypeHeat,
};

const ENERGY_INDEX = RESOURCE_KEYS.findIndex(({ key }) => key === "energy");

const PHASE_INITIAL_DELAY = 1000;
const COUNTER_DURATION = 1500;
const PHASE_FADE_BUFFER = 200;
const PHASE_PAUSE_BETWEEN = 500;

function useCounterAnimation(
  startValue: number,
  endValue: number,
  active: boolean,
  durationMs: number = COUNTER_DURATION,
): { value: number; done: boolean } {
  const [displayValue, setDisplayValue] = useState(startValue);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!active) {
      setDisplayValue(startValue);
      setDone(false);
      return;
    }

    const delta = endValue - startValue;
    if (delta === 0) {
      setDisplayValue(endValue);
      setDone(true);
      return;
    }

    setDone(false);
    const steps = Math.abs(delta);
    const interval = durationMs / steps;
    const direction = delta > 0 ? 1 : -1;
    let current = startValue;

    const timer = setInterval(() => {
      current += direction;
      setDisplayValue(current);
      if (current === endValue) {
        clearInterval(timer);
        setDone(true);
      }
    }, interval);

    return () => clearInterval(timer);
  }, [startValue, endValue, active, durationMs]);

  return { value: displayValue, done };
}

const ProductionPhaseModal: React.FC<ProductionPhaseModalProps> = ({
  isOpen,
  gameState,
  onClose,
  onHide,
  openDirectlyToCardSelection = false,
}) => {
  const soundHandlesRef = useRef<HTMLAudioElement[]>([]);
  const [hasSubmittedCardSelection, setHasSubmittedCardSelection] = useState(false);
  const persistedSelectionRef = useRef<string[]>([]);
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [animationPhase, setAnimationPhase] = useState<AnimationPhase>("final");
  const [showCardSelection, setShowCardSelection] = useState(false);
  const prevSelectionCompleteRef = useRef<boolean | undefined>(undefined);
  const prevIsOpenRef = useRef(false);
  const animationPlayedRef = useRef(false);
  const animationTimersRef = useRef<NodeJS.Timeout[]>([]);

  const stopAllSounds = useCallback(() => {
    for (const handle of soundHandlesRef.current) {
      handle.pause();
      handle.currentTime = 0;
    }
    soundHandlesRef.current = [];
  }, []);

  const playScoreSound = useCallback(() => {
    const handle = audioService.playSoundWithHandle("production-score");
    if (handle) {
      soundHandlesRef.current.push(handle);
    }
  }, []);

  const clearAnimationTimers = useCallback(() => {
    for (const t of animationTimersRef.current) {
      clearTimeout(t);
    }
    animationTimersRef.current = [];
    stopAllSounds();
  }, [stopAllSounds]);

  // Clean up sounds on unmount
  useEffect(() => {
    return () => {
      for (const handle of soundHandlesRef.current) {
        handle.pause();
      }
    };
  }, []);

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

  const isLastRound = gameState?.isLastRound ?? false;

  const handleNextClick = useCallback(() => {
    clearAnimationTimers();
    animationPlayedRef.current = true;
    setAnimationPhase("final");

    if (isLastRound) {
      void handleCardSelection([]);
    } else {
      setShowCardSelection(true);
    }
  }, [isLastRound, handleCardSelection, clearAnimationTimers]);

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
      persistedSelectionRef.current = [];
      animationPlayedRef.current = false;
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
      return {
        playerId: player.id,
        playerName: player.name,
        playerColor: player.color,
        production: player.production,
        terraformRating: player.terraformRating,
        ...productionPhase,
      };
    });

    return {
      playersData,
      generation: gameState.generation || 1,
    };
  }, [gameState]);

  useEffect(() => {
    if (!isOpen || !modalProductionData || animationPlayedRef.current) {
      return;
    }

    const currentPlayerData = modalProductionData.playersData[0];
    const hasEnergyToConvert = currentPlayerData && currentPlayerData.energyConverted > 0;

    setAnimationPhase("initial");

    const timers: NodeJS.Timeout[] = [];

    const energyPhaseDuration = COUNTER_DURATION + PHASE_FADE_BUFFER;
    const productionPhaseDuration = COUNTER_DURATION + PHASE_FADE_BUFFER;

    const t1 = setTimeout(() => {
      if (hasEnergyToConvert) {
        setAnimationPhase("energyTransfer");
        playScoreSound();
      } else {
        setAnimationPhase("productionTransfer");
        playScoreSound();
      }
    }, PHASE_INITIAL_DELAY);
    timers.push(t1);

    const stopAfterPhase1 = setTimeout(() => {
      stopAllSounds();
    }, PHASE_INITIAL_DELAY + COUNTER_DURATION);
    timers.push(stopAfterPhase1);

    const t2Delay = hasEnergyToConvert
      ? PHASE_INITIAL_DELAY + energyPhaseDuration + PHASE_PAUSE_BETWEEN
      : PHASE_INITIAL_DELAY;

    if (hasEnergyToConvert) {
      const t2 = setTimeout(() => {
        setAnimationPhase("productionTransfer");
        playScoreSound();
      }, t2Delay);
      timers.push(t2);

      const stopAfterPhase2 = setTimeout(() => {
        stopAllSounds();
      }, t2Delay + COUNTER_DURATION);
      timers.push(stopAfterPhase2);
    }

    const t3 = setTimeout(() => {
      setAnimationPhase("final");
      animationPlayedRef.current = true;
    }, t2Delay + productionPhaseDuration);
    timers.push(t3);

    animationTimersRef.current = timers;

    return () => {
      for (const t of timers) {
        clearTimeout(t);
      }
    };
  }, [isOpen, modalProductionData, playScoreSound, stopAllSounds]);

  const handlePlayerSelect = (playerIndex: number) => {
    if (playerIndex === currentPlayerIndex) {
      return;
    }

    if (!animationPlayedRef.current) {
      clearAnimationTimers();
      animationPlayedRef.current = true;
      setAnimationPhase("final");
    }

    setCurrentPlayerIndex(playerIndex);
  };

  useEffect(() => {
    if (!isOpen) {
      return;
    }
    setCurrentPlayerIndex(0);
  }, [isOpen]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Enter" && !hasSubmittedCardSelection && !showCardSelection) {
        handleNextClick();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }

    return () => {};
  }, [isOpen, hasSubmittedCardSelection, showCardSelection, handleNextClick]);

  if (!isOpen) {
    return null;
  }
  if (!modalProductionData) {
    return null;
  }

  const currentPlayerData = modalProductionData.playersData[currentPlayerIndex];
  if (!currentPlayerData) {
    return null;
  }

  const effectivePhase: AnimationPhase = currentPlayerIndex === 0 ? animationPhase : "final";

  return (
    <>
      {isOpen && !showCardSelection && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ zIndex: Z_INDEX.CORPORATION_SELECTION }}
        >
          <div className={OVERLAY_BACKDROP_BLUR_CLASS} />
          <div className={OVERLAY_BACKDROP_TINT_CLASS} />

          <div className="relative z-[1]">
            <div className="max-w-[1050px] min-w-[850px] flex flex-col bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_60px_rgba(30,60,150,0.5)]">
              <div className={OVERLAY_HEADER_CLASS}>
                <h2 className={`${OVERLAY_TITLE_CLASS} text-center`}>Production</h2>
                <p className="mt-2 mb-0 text-base text-white/60 text-center">
                  Generation {modalProductionData.generation}
                </p>
              </div>

              {modalProductionData.playersData.length > 1 && (
                <div className="flex justify-center gap-3 px-8 py-4">
                  {modalProductionData.playersData.map((player, index) => (
                    <button
                      key={player.playerId}
                      className={`bg-space-black-darker/60 border-2 border-space-blue-400/40 rounded-lg text-white/80 text-sm font-semibold py-2 px-4 cursor-pointer transition-all duration-300 text-shadow-dark hover:border-space-blue-400/60 hover:text-white/90 hover:-translate-y-px ${
                        index === currentPlayerIndex
                          ? "!bg-space-blue-400/20 !border-space-blue-400 !text-white shadow-[0_0_15px_rgba(30,60,150,0.4)]"
                          : ""
                      }`}
                      onClick={() => handlePlayerSelect(index)}
                    >
                      <span
                        className="inline-block w-2.5 h-2.5 rounded-full mr-2 flex-shrink-0"
                        style={{ backgroundColor: player.playerColor }}
                      />
                      {player.playerName}
                    </button>
                  ))}
                </div>
              )}

              <div className="px-8 pt-2 pb-14">
                <ResourceGrid playerData={currentPlayerData} animationPhase={effectivePhase} />
              </div>
            </div>

            {!hasSubmittedCardSelection && !showCardSelection && (
              <div className="absolute left-full top-1/2 -translate-y-1/2 ml-5">
                <GameButton
                  buttonType="secondary"
                  size="lg"
                  onClick={handleNextClick}
                  className="whitespace-nowrap"
                >
                  Buy cards
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    className="inline-block ml-2"
                  >
                    <path
                      fill="currentColor"
                      d="m15.06 5.283l5.657 5.657a1.5 1.5 0 0 1 0 2.12l-5.656 5.658a1.5 1.5 0 0 1-2.122-2.122l3.096-3.096H4.5a1.5 1.5 0 0 1 0-3h11.535L12.94 7.404a1.5 1.5 0 0 1 2.122-2.121Z"
                    />
                  </svg>
                </GameButton>
              </div>
            )}
          </div>
        </div>
      )}

      {showCardSelection && (
        <ProductionCardSelectionOverlay
          isOpen={showCardSelection}
          cards={gameState?.currentPlayer?.productionPhase?.availableCards || []}
          playerCredits={gameState?.currentPlayer?.resources.credits || 0}
          costPerCard={
            gameState?.currentPlayer?.actionCosts?.find((a) => a.actionType === "card-buying")
              ?.costs[0]?.effectiveCost ?? 3
          }
          onSelectCards={handleCardSelection}
          onReturn={handleReturnFromCardSelection}
          initialSelectedCardIds={persistedSelectionRef.current}
          onSelectionChange={(ids) => {
            persistedSelectionRef.current = ids;
          }}
        />
      )}
    </>
  );
};

interface ResourceGridProps {
  playerData: {
    beforeResources: ResourcesDto;
    afterResources: ResourcesDto;
    production: ProductionDto;
    energyConverted: number;
    creditsIncome: number;
    terraformRating: number;
  };
  animationPhase: AnimationPhase;
}

const ResourceGrid: React.FC<ResourceGridProps> = ({ playerData, animationPhase }) => {
  const { beforeResources, afterResources, production, energyConverted } = playerData;

  const isEnergyPhaseActive = animationPhase === "energyTransfer";
  const hasEnergy = beforeResources.energy > 0;
  const shouldAnimateEnergyTransfer = isEnergyPhaseActive && hasEnergy && energyConverted > 0;

  const energyDrainCounter = useCounterAnimation(
    beforeResources.energy,
    0,
    shouldAnimateEnergyTransfer,
  );
  const energyTransferDone = energyDrainCounter.done;

  const getResourceValues = (resField: keyof ResourcesDto, key: ResourceType) => {
    const beforeVal = beforeResources[resField];

    let afterEnergyVal: number;
    if (key === "energy") {
      afterEnergyVal = 0;
    } else if (key === "heat") {
      afterEnergyVal = beforeResources.heat + energyConverted;
    } else {
      afterEnergyVal = beforeVal;
    }

    const afterProdVal = afterResources[resField];

    return { beforeVal, afterEnergyVal, afterProdVal };
  };

  return (
    <div className="flex flex-col items-center gap-2 py-4 scale-125 my-4">
      <div className="flex items-start justify-center gap-3 relative">
        {RESOURCE_KEYS.map(({ key, resField }, index) => {
          const { beforeVal, afterEnergyVal, afterProdVal } = getResourceValues(resField, key);
          const prodVal = production[resField];

          const startForEnergy = beforeVal;
          const endForEnergy = afterEnergyVal;

          const startForProduction =
            key === "energy" ? 0 : key === "heat" ? afterEnergyVal : beforeVal;
          const endForProduction = afterProdVal;

          return (
            <React.Fragment key={key}>
              <div className="relative">
                <ResourceColumn
                  resourceKey={key}
                  productionValue={prodVal}
                  startForEnergy={startForEnergy}
                  endForEnergy={endForEnergy}
                  startForProduction={startForProduction}
                  endForProduction={endForProduction}
                  animationPhase={animationPhase}
                  energyConverted={energyConverted}
                  terraformRating={key === "credit" ? playerData.terraformRating : undefined}
                />
                {key === "energy" && (
                  <EnergyToHeatArrows active={shouldAnimateEnergyTransfer && !energyTransferDone} />
                )}
              </div>
              {index < RESOURCE_KEYS.length - 1 && (
                <div
                  className={`w-[1px] h-[90px] self-center mt-2 transition-opacity duration-300 ${index === ENERGY_INDEX && shouldAnimateEnergyTransfer && !energyTransferDone ? "opacity-0" : ""}`}
                  style={{
                    background: "linear-gradient(transparent, rgba(60,60,70,0.8), transparent)",
                  }}
                />
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
};

interface ResourceColumnProps {
  resourceKey: ResourceType;
  productionValue: number;
  startForEnergy: number;
  endForEnergy: number;
  startForProduction: number;
  endForProduction: number;
  animationPhase: AnimationPhase;
  energyConverted: number;
  terraformRating?: number;
}

const ResourceColumn: React.FC<ResourceColumnProps> = ({
  resourceKey,
  productionValue,
  startForEnergy,
  endForEnergy,
  startForProduction,
  endForProduction,
  animationPhase,
  energyConverted,
  terraformRating,
}) => {
  const isEnergyPhaseActive = animationPhase === "energyTransfer";
  const isProductionPhaseActive = animationPhase === "productionTransfer";

  const isEnergyResource = resourceKey === "energy" || resourceKey === "heat";
  const shouldAnimateEnergy = isEnergyPhaseActive && isEnergyResource && energyConverted > 0;
  const shouldAnimateProduction =
    isProductionPhaseActive && endForProduction - startForProduction !== 0;

  const energyCounter = useCounterAnimation(startForEnergy, endForEnergy, shouldAnimateEnergy);

  const productionCounter = useCounterAnimation(
    startForProduction,
    endForProduction,
    shouldAnimateProduction,
  );

  const getDisplayedResourceValue = (): number => {
    if (animationPhase === "initial") {
      return startForEnergy;
    }
    if (animationPhase === "energyTransfer") {
      if (shouldAnimateEnergy) {
        return energyCounter.value;
      }
      return startForEnergy;
    }
    if (animationPhase === "productionTransfer") {
      if (shouldAnimateProduction) {
        return productionCounter.value;
      }
      return startForProduction;
    }
    return endForProduction;
  };

  const displayedValue = getDisplayedResourceValue();

  const isDimmedDuringEnergy = isEnergyPhaseActive && !isEnergyResource;

  const hasProdDelta = endForProduction - startForProduction !== 0;
  const showDownArrows = isProductionPhaseActive && hasProdDelta && !productionCounter.done;

  const prodBadgeLift =
    animationPhase === "productionTransfer" && hasProdDelta && !productionCounter.done
      ? "-translate-y-2"
      : "";

  return (
    <div
      className={`flex flex-col items-center gap-0 px-3 py-2 min-w-[70px] transition-opacity duration-300 ${
        isDimmedDuringEnergy ? "opacity-30" : "opacity-100"
      }`}
    >
      <div className="relative">
        {resourceKey === "credit" && terraformRating !== undefined && (
          <div
            className={`absolute bottom-full left-1/2 -translate-x-1/2 mb-0.5 inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(80,80,120,0.5)_0%,rgba(60,60,100,0.45)_100%)] border border-[rgba(100,100,160,0.6)] py-0.5 w-[36px] transition-transform duration-500 ease-in-out ${prodBadgeLift}`}
          >
            <span className="text-[10px] font-bold font-orbitron text-white/90 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none tabular-nums">
              {terraformRating}
            </span>
          </div>
        )}
        <div
          className={`inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.5)_0%,rgba(139,89,42,0.45)_100%)] border border-[rgba(160,110,60,0.6)] px-3 py-0.5 w-[36px] transition-transform duration-500 ease-in-out ${prodBadgeLift}`}
        >
          <span className="text-[11px] font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none tabular-nums">
            {productionValue}
          </span>
        </div>
      </div>

      <div className="h-[20px] flex items-center justify-center overflow-hidden">
        {showDownArrows && <VerticalArrows />}
      </div>

      <div className="flex flex-col items-center gap-0.5">
        {resourceKey === "credit" ? (
          <div className="flex items-center gap-1 w-[60px] justify-center scale-110">
            <GameIcon iconType={ResourceTypeCredit} amount={displayedValue} size="small" />
          </div>
        ) : (
          <div className="flex items-center gap-1 w-[60px] justify-center scale-[0.85]">
            <GameIcon iconType={RESOURCE_ICON_TYPES[resourceKey]} size="small" />
            <span className="text-sm font-bold font-orbitron text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] tabular-nums min-w-[32px] text-left">
              {displayedValue}
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

const VerticalArrows: React.FC = () => {
  return (
    <div className="flex flex-col items-center leading-none overflow-hidden h-[20px] w-[16px] relative">
      <span className="text-[10px] text-white/80 font-bold absolute left-1/2 -translate-x-1/2 top-0 animate-[slideDown_0.6s_linear_infinite]">
        ▼
      </span>
      <span className="text-[10px] text-white/50 font-bold absolute left-1/2 -translate-x-1/2 top-0 animate-[slideDown_0.6s_linear_0.2s_infinite]">
        ▼
      </span>
      <span className="text-[10px] text-white/30 font-bold absolute left-1/2 -translate-x-1/2 top-0 animate-[slideDown_0.6s_linear_0.4s_infinite]">
        ▼
      </span>
    </div>
  );
};

interface EnergyToHeatArrowsProps {
  active: boolean;
}

const EnergyToHeatArrows: React.FC<EnergyToHeatArrowsProps> = ({ active }) => {
  return (
    <div
      className={`absolute right-0 translate-x-[100%] bottom-[13px] w-[30px] overflow-hidden transition-opacity duration-500 pointer-events-none ${
        active ? "opacity-100" : "opacity-0"
      }`}
    >
      <div className="relative h-[16px] w-[30px] overflow-hidden">
        <span className="text-[10px] text-white/80 font-bold absolute top-1/2 -translate-y-1/2 left-0 animate-[slideRight_0.6s_linear_infinite]">
          ▶
        </span>
        <span className="text-[10px] text-white/50 font-bold absolute top-1/2 -translate-y-1/2 left-0 animate-[slideRight_0.6s_linear_0.2s_infinite]">
          ▶
        </span>
        <span className="text-[10px] text-white/30 font-bold absolute top-1/2 -translate-y-1/2 left-0 animate-[slideRight_0.6s_linear_0.4s_infinite]">
          ▶
        </span>
      </div>
    </div>
  );
};

export default ProductionPhaseModal;
