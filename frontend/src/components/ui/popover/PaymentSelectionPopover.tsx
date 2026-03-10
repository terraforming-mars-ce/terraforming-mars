import React, { useState, useEffect, useCallback, useRef, useMemo } from "react";
import {
  CardPaymentDto,
  PaymentConstantsDto,
  PlayerCardDto,
  PaymentSubstituteDto,
  ResourcesDto,
  StoragePaymentSubstituteDto,
  TagBuilding,
  TagSpace,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { Z_INDEX } from "@/constants/zIndex";
import { cardMatchesStorageSubstitute } from "@/utils/paymentUtils.ts";

interface PaymentSelectionPopoverProps {
  cardId: string;
  card: PlayerCardDto;
  playerResources: ResourcesDto;
  paymentConstants: PaymentConstantsDto;
  playerPaymentSubstitutes?: PaymentSubstituteDto[];
  storagePaymentSubstitutes?: StoragePaymentSubstituteDto[];
  resourceStorage?: { [key: string]: number };
  onConfirm: (payment: CardPaymentDto) => void;
  onCancel: () => void;
  isVisible: boolean;
}

const PaymentSelectionPopover: React.FC<PaymentSelectionPopoverProps> = ({
  cardId: _cardId,
  card,
  playerResources,
  paymentConstants,
  playerPaymentSubstitutes,
  storagePaymentSubstitutes,
  resourceStorage,
  onConfirm,
  onCancel,
  isVisible,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  // Payment state - only steel and titanium, credits are calculated
  const [steel, setSteel] = useState(0);
  const [titanium, setTitanium] = useState(0);

  // Payment substitutes state (dynamic based on player's available substitutes)
  const [substitutes, setSubstitutes] = useState<Record<string, number>>({});

  // Storage payment substitutes state (cardId -> amount)
  const [storageSubstitutes, setStorageSubstitutes] = useState<Record<string, number>>({});

  // Get dynamic steel/titanium values from paymentSubstitutes (includes value modifiers like Phobolog)
  // Falls back to paymentConstants for backwards compatibility
  const steelValue =
    playerPaymentSubstitutes?.find((s) => s.resourceType === "steel")?.conversionRate ??
    paymentConstants.steelValue;
  const titaniumValue =
    playerPaymentSubstitutes?.find((s) => s.resourceType === "titanium")?.conversionRate ??
    paymentConstants.titaniumValue;

  // Filter storage substitutes applicable to this card
  const applicableStorageSubstitutes = useMemo(() => {
    if (!storagePaymentSubstitutes || !resourceStorage) {
      return [];
    }
    return storagePaymentSubstitutes.filter((sub) => {
      const available = resourceStorage[sub.cardId] ?? 0;
      return available > 0 && cardMatchesStorageSubstitute(card, sub);
    });
  }, [storagePaymentSubstitutes, resourceStorage, card]);

  // Reset payment when modal opens
  useEffect(() => {
    if (isVisible) {
      setSteel(0);
      setTitanium(0);
      setSubstitutes({});
      setStorageSubstitutes({});
    }
  }, [isVisible, card.effectiveCost]);

  // Calculate storage substitutes value
  let storageSubstitutesValue = 0;
  for (const [cardId, amount] of Object.entries(storageSubstitutes)) {
    const sub = applicableStorageSubstitutes.find((s) => s.cardId === cardId);
    if (sub) {
      storageSubstitutesValue += amount * sub.conversionRate;
    }
  }

  // Calculate final cost after applying steel, titanium, and substitute discounts
  let substitutesValue = 0;
  if (playerPaymentSubstitutes) {
    for (const [resourceType, amount] of Object.entries(substitutes)) {
      const substitute = playerPaymentSubstitutes.find((sub) => sub.resourceType === resourceType);
      if (substitute) {
        substitutesValue += amount * substitute.conversionRate;
      }
    }
  }

  const finalCost =
    card.effectiveCost -
    steel * steelValue -
    titanium * titaniumValue -
    substitutesValue -
    storageSubstitutesValue;

  // Helper to calculate total non-credit payment value (excluding a specific source)
  const totalOtherPaymentValue = (
    excludeSteel: boolean,
    excludeTitanium: boolean,
    excludeSubResourceType?: string,
    excludeStorageCardId?: string,
  ) => {
    let total = 0;
    if (!excludeSteel) {
      total += steel * steelValue;
    }
    if (!excludeTitanium) {
      total += titanium * titaniumValue;
    }
    if (playerPaymentSubstitutes) {
      for (const [resourceType, amount] of Object.entries(substitutes)) {
        if (resourceType !== excludeSubResourceType) {
          const sub = playerPaymentSubstitutes.find((s) => s.resourceType === resourceType);
          if (sub) {
            total += amount * sub.conversionRate;
          }
        }
      }
    }
    for (const [cardId, amount] of Object.entries(storageSubstitutes)) {
      if (cardId !== excludeStorageCardId) {
        const sub = applicableStorageSubstitutes.find((s) => s.cardId === cardId);
        if (sub) {
          total += amount * sub.conversionRate;
        }
      }
    }
    return total;
  };

  // Check which resources to show (only if card has appropriate tag)
  const canUseSteel = card.tags?.includes(TagBuilding);
  const canUseTitanium = card.tags?.includes(TagSpace);

  // Calculate max units dynamically based on current payment state
  const remainingCostForSteel = card.effectiveCost - totalOtherPaymentValue(true, false);
  const maxSteelUnits = canUseSteel
    ? Math.min(playerResources.steel, Math.ceil(Math.max(0, remainingCostForSteel) / steelValue))
    : 0;

  const remainingCostForTitanium = card.effectiveCost - totalOtherPaymentValue(false, true);
  const maxTitaniumUnits = canUseTitanium
    ? Math.min(
        playerResources.titanium,
        Math.ceil(Math.max(0, remainingCostForTitanium) / titaniumValue),
      )
    : 0;

  // Check warnings
  const isOverpaying = finalCost < 0;
  const cannotAfford = finalCost > playerResources.credits;
  const canConfirm = !cannotAfford;

  const handleCancelClick = useCallback(() => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200);
  }, [onCancel]);

  const handleConfirm = useCallback(() => {
    if (!canConfirm) return;

    const payment: CardPaymentDto = {
      credits: Math.max(0, finalCost),
      steel,
      titanium,
      substitutes: Object.keys(substitutes).length > 0 ? substitutes : undefined,
      storageSubstitutes:
        Object.keys(storageSubstitutes).length > 0 ? storageSubstitutes : undefined,
    };

    onConfirm(payment);
  }, [steel, titanium, substitutes, storageSubstitutes, finalCost, canConfirm, onConfirm]);

  // Escape key handler
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isVisible) {
        handleCancelClick();
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
  }, [isVisible, handleCancelClick]);

  // Early return if modal is not visible
  if (!isVisible) return null;

  // Increment/decrement handlers
  const incrementSteel = () => {
    if (steel < maxSteelUnits) {
      setSteel(steel + 1);
    }
  };

  const decrementSteel = () => {
    if (steel > 0) {
      setSteel(steel - 1);
    }
  };

  const incrementTitanium = () => {
    if (titanium < maxTitaniumUnits) {
      setTitanium(titanium + 1);
    }
  };

  const decrementTitanium = () => {
    if (titanium > 0) {
      setTitanium(titanium - 1);
    }
  };

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-start justify-center pt-[30vh] pointer-events-auto overflow-hidden"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className={`
          min-w-[400px] w-fit max-w-[90vw] max-h-[600px]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          transition-all duration-300 ease-in-out
          ${isClosing ? "animate-fadeOut" : "animate-popIn"}
        `}
        ref={popoverRef}
      >
        {/* Header */}
        <div className="py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Select Payment Method
          </h3>
          <div className="mt-2 flex items-center justify-between">
            <span className="text-sm text-gray-400">Pay for: {card.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-400">Cost:</span>
              <GameIcon iconType="credit" amount={card.effectiveCost} size="small" />
            </div>
          </div>
        </div>

        {/* Payment options */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {/* Steel - shown only if card has building tag and player has steel */}
          {canUseSteel && playerResources.steel > 0 && (
            <div className="flex items-center justify-between rounded-md bg-black/30 border-2 border-space-blue-500/40 p-4">
              <div className="flex items-center gap-3">
                <GameIcon iconType="steel" size="medium" />
                <div className="flex flex-col">
                  <span className="text-white">Steel</span>
                  <span className="text-xs text-gray-400">
                    {steelValue} MC each ({playerResources.steel} available)
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={decrementSteel}
                  disabled={steel === 0}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                >
                  −
                </button>
                <span className="w-12 text-center text-lg text-white font-semibold">{steel}</span>
                <button
                  onClick={incrementSteel}
                  disabled={steel >= maxSteelUnits}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                >
                  +
                </button>
              </div>
            </div>
          )}

          {/* Titanium - shown only if card has space tag and player has titanium */}
          {canUseTitanium && playerResources.titanium > 0 && (
            <div className="flex items-center justify-between rounded-md bg-black/30 border-2 border-space-blue-500/40 p-4">
              <div className="flex items-center gap-3">
                <GameIcon iconType="titanium" size="medium" />
                <div className="flex flex-col">
                  <span className="text-white">Titanium</span>
                  <span className="text-xs text-gray-400">
                    {titaniumValue} MC each ({playerResources.titanium} available)
                  </span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={decrementTitanium}
                  disabled={titanium === 0}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                >
                  −
                </button>
                <span className="w-12 text-center text-lg text-white font-semibold">
                  {titanium}
                </span>
                <button
                  onClick={incrementTitanium}
                  disabled={titanium >= maxTitaniumUnits}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                >
                  +
                </button>
              </div>
            </div>
          )}

          {/* Payment Substitutes - shown for each substitute player has (excluding steel/titanium which are handled above) */}
          {playerPaymentSubstitutes &&
            playerPaymentSubstitutes
              .filter((sub) => sub.resourceType !== "steel" && sub.resourceType !== "titanium")
              .map((substitute) => {
                const resourceType = substitute.resourceType;
                let available = 0;

                switch (resourceType) {
                  case "heat":
                    available = playerResources.heat;
                    break;
                  case "energy":
                    available = playerResources.energy;
                    break;
                  case "plant":
                    available = playerResources.plants;
                    break;
                }

                if (available === 0) return null;

                const currentAmount = substitutes[resourceType] || 0;

                const remainingCostForThisSubstitute =
                  card.effectiveCost -
                  totalOtherPaymentValue(false, false, resourceType, undefined);

                const maxUnits = Math.min(
                  available,
                  Math.ceil(
                    Math.max(0, remainingCostForThisSubstitute) / substitute.conversionRate,
                  ),
                );

                const incrementSubstitute = () => {
                  if (currentAmount < maxUnits) {
                    setSubstitutes((prev) => ({
                      ...prev,
                      [resourceType]: currentAmount + 1,
                    }));
                  }
                };

                const decrementSubstitute = () => {
                  if (currentAmount > 0) {
                    setSubstitutes((prev) => {
                      const newSubs = { ...prev };
                      if (currentAmount === 1) {
                        delete newSubs[resourceType];
                      } else {
                        newSubs[resourceType] = currentAmount - 1;
                      }
                      return newSubs;
                    });
                  }
                };

                return (
                  <div
                    key={resourceType}
                    className="flex items-center justify-between rounded-md bg-black/30 border-2 border-space-blue-500/40 p-4"
                  >
                    <div className="flex items-center gap-3">
                      <GameIcon iconType={resourceType} size="medium" />
                      <div className="flex flex-col">
                        <span className="text-white capitalize">{resourceType}</span>
                        <span className="text-xs text-gray-400">
                          {substitute.conversionRate} MC each ({available} available)
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button
                        onClick={decrementSubstitute}
                        disabled={currentAmount === 0}
                        className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                      >
                        −
                      </button>
                      <span className="w-12 text-center text-lg text-white font-semibold">
                        {currentAmount}
                      </span>
                      <button
                        onClick={incrementSubstitute}
                        disabled={currentAmount >= maxUnits}
                        className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                      >
                        +
                      </button>
                    </div>
                  </div>
                );
              })}

          {/* Storage Payment Substitutes (e.g., Dirigibles floaters) */}
          {applicableStorageSubstitutes.map((substitute) => {
            const available = resourceStorage?.[substitute.cardId] ?? 0;
            if (available === 0) return null;

            const currentAmount = storageSubstitutes[substitute.cardId] || 0;

            const remainingCostForThis =
              card.effectiveCost -
              totalOtherPaymentValue(false, false, undefined, substitute.cardId);

            const maxUnits = Math.min(
              available,
              Math.ceil(Math.max(0, remainingCostForThis) / substitute.conversionRate),
            );

            const incrementStorage = () => {
              if (currentAmount < maxUnits) {
                setStorageSubstitutes((prev) => ({
                  ...prev,
                  [substitute.cardId]: currentAmount + 1,
                }));
              }
            };

            const decrementStorage = () => {
              if (currentAmount > 0) {
                setStorageSubstitutes((prev) => {
                  const newSubs = { ...prev };
                  if (currentAmount === 1) {
                    delete newSubs[substitute.cardId];
                  } else {
                    newSubs[substitute.cardId] = currentAmount - 1;
                  }
                  return newSubs;
                });
              }
            };

            return (
              <div
                key={`storage-${substitute.cardId}`}
                className="flex items-center justify-between rounded-md bg-black/30 border-2 border-space-blue-500/40 p-4"
              >
                <div className="flex items-center gap-3">
                  <GameIcon iconType={substitute.resourceType} size="medium" />
                  <div className="flex flex-col">
                    <span className="text-white capitalize">{substitute.resourceType}</span>
                    <span className="text-xs text-gray-400">
                      {substitute.conversionRate} MC each ({available} available)
                    </span>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    onClick={decrementStorage}
                    disabled={currentAmount === 0}
                    className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                  >
                    −
                  </button>
                  <span className="w-12 text-center text-lg text-white font-semibold">
                    {currentAmount}
                  </span>
                  <button
                    onClick={incrementStorage}
                    disabled={currentAmount >= maxUnits}
                    className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black transition-all"
                  >
                    +
                  </button>
                </div>
              </div>
            );
          })}

          {/* Show message if no alternative payment options */}
          {!canUseSteel &&
            !canUseTitanium &&
            (!playerPaymentSubstitutes || playerPaymentSubstitutes.length === 0) &&
            applicableStorageSubstitutes.length === 0 && (
              <div className="text-center text-gray-400 py-4">
                This card cannot use alternative payment methods
              </div>
            )}
        </div>

        {/* Payment summary */}
        <div className="px-4 pb-4">
          <div className="rounded-md border border-space-blue-500/30 bg-space-black/30 p-4">
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Card cost:</span>
              <div
                className={`flex items-center gap-2 ${finalCost < 0 ? "opacity-90" : cannotAfford ? "opacity-90" : ""}`}
              >
                <GameIcon iconType="credit" amount={finalCost} size="medium" />
              </div>
            </div>
            <div
              className={`overflow-hidden transition-all duration-300 ease-in-out ${
                isOverpaying ? "max-h-[100px] opacity-100" : "max-h-0 opacity-0"
              }`}
            >
              <div
                className={`mt-2 text-sm text-yellow-500 border-t border-space-blue-500/20 pt-2 transition-opacity duration-150 ${
                  isOverpaying ? "opacity-100" : "opacity-0"
                }`}
              >
                Overpaying by {Math.max(-finalCost, 0)} MC (excess will be lost)
              </div>
            </div>
            <div
              className={`overflow-hidden transition-all duration-300 ease-in-out ${
                cannotAfford ? "max-h-[100px] opacity-100" : "max-h-0 opacity-0"
              }`}
            >
              <div
                className={`mt-2 text-sm text-error-red border-t border-space-blue-500/20 pt-2 transition-opacity duration-150 ${
                  cannotAfford ? "opacity-100" : "opacity-0"
                }`}
              >
                Can't afford: Need {Math.max(finalCost - playerResources.credits, 0)} more MC
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-end gap-3">
          <button
            onClick={handleCancelClick}
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
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={!canConfirm}
            className={`
              border-2 rounded-md text-white text-xs font-semibold
              px-6 py-2 cursor-pointer
              transition-all duration-200
              text-shadow-glow font-orbitron
              ${
                canConfirm
                  ? "bg-green-600/80 border-green-500/60 shadow-[0_0_8px_rgba(34,197,94,0.4)] hover:bg-green-500/80 hover:border-green-500/80 hover:shadow-[0_0_12px_rgba(34,197,94,0.6)]"
                  : "bg-gray-600/50 border-gray-500/40 opacity-50 cursor-default"
              }
            `}
          >
            Confirm
          </button>
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

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }

        /* Media queries */
        @media (max-width: 768px) {
          .min-w-\\[400px\\] {
            min-width: 320px;
          }
          .max-w-\\[90vw\\] {
            max-width: 95vw;
          }
        }
      `}</style>
    </div>
  );
};

export default PaymentSelectionPopover;
