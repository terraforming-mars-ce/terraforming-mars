import React, { useState, useEffect, useCallback, useMemo } from "react";
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
import { cardMatchesStorageSubstitute } from "@/utils/paymentUtils.ts";
import GameButton from "../buttons/GameButton.tsx";
import { GameFlowPopover, GameFlowTitle, GameFlowFooter } from "./GameFlowPopover.tsx";

export interface GenericPaymentConfig {
  name: string;
  cost: number;
  substitutes: Array<{ resourceType: string; conversionRate: number }>;
  baseResource?: string;
  storageSubstitutes?: StoragePaymentSubstituteDto[];
  resourceStorage?: { [key: string]: number };
}

interface PaymentSelectionPopoverProps {
  cardId?: string;
  card?: PlayerCardDto;
  playerResources: ResourcesDto;
  paymentConstants?: PaymentConstantsDto;
  playerPaymentSubstitutes?: PaymentSubstituteDto[];
  storagePaymentSubstitutes?: StoragePaymentSubstituteDto[];
  resourceStorage?: { [key: string]: number };
  genericPayment?: GenericPaymentConfig;
  onConfirm: (payment: CardPaymentDto) => void;
  onCancel: () => void;
  isVisible: boolean;
}

const PaymentSelectionPopover: React.FC<PaymentSelectionPopoverProps> = ({
  card,
  playerResources,
  paymentConstants,
  playerPaymentSubstitutes,
  storagePaymentSubstitutes,
  resourceStorage,
  genericPayment,
  onConfirm,
  onCancel,
  isVisible,
}) => {
  const [steel, setSteel] = useState(0);
  const [titanium, setTitanium] = useState(0);
  const [substitutes, setSubstitutes] = useState<Record<string, number>>({});
  const [storageSubstitutes, setStorageSubstitutes] = useState<Record<string, number>>({});

  const isGenericMode = !!genericPayment;
  const effectiveCost = genericPayment?.cost ?? card?.effectiveCost ?? 0;
  const effectiveName = genericPayment?.name ?? card?.name ?? "";

  const genericSteelSub = genericPayment?.substitutes.find((s) => s.resourceType === "steel");
  const genericTitaniumSub = genericPayment?.substitutes.find((s) => s.resourceType === "titanium");

  const steelValue = isGenericMode
    ? (genericSteelSub?.conversionRate ?? 0)
    : (playerPaymentSubstitutes?.find((s) => s.resourceType === "steel")?.conversionRate ??
      paymentConstants?.steelValue ??
      2);
  const titaniumValue = isGenericMode
    ? (genericTitaniumSub?.conversionRate ?? 0)
    : (playerPaymentSubstitutes?.find((s) => s.resourceType === "titanium")?.conversionRate ??
      paymentConstants?.titaniumValue ??
      3);

  const applicableStorageSubstitutes = useMemo(() => {
    if (isGenericMode) {
      const genSubs = genericPayment?.storageSubstitutes;
      const genStorage = genericPayment?.resourceStorage;
      if (!genSubs || !genStorage) {
        return [];
      }
      return genSubs.filter((sub) => (genStorage[sub.cardId] ?? 0) > 0);
    }
    if (!storagePaymentSubstitutes || !resourceStorage) {
      return [];
    }
    return storagePaymentSubstitutes.filter((sub) => {
      const available = resourceStorage[sub.cardId] ?? 0;
      return available > 0 && cardMatchesStorageSubstitute(card!, sub);
    });
  }, [isGenericMode, genericPayment, storagePaymentSubstitutes, resourceStorage, card]);

  useEffect(() => {
    if (isVisible) {
      setSteel(0);
      setTitanium(0);
      setSubstitutes({});
      setStorageSubstitutes({});
    }
  }, [isVisible, effectiveCost]);

  let storageSubstitutesValue = 0;
  for (const [cardId, amount] of Object.entries(storageSubstitutes)) {
    const sub = applicableStorageSubstitutes.find((s) => s.cardId === cardId);
    if (sub) {
      storageSubstitutesValue += amount * sub.conversionRate;
    }
  }

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
    effectiveCost -
    steel * steelValue -
    titanium * titaniumValue -
    substitutesValue -
    storageSubstitutesValue;

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

  const canUseSteel = isGenericMode ? !!genericSteelSub : card?.tags?.includes(TagBuilding);
  const canUseTitanium = isGenericMode ? !!genericTitaniumSub : card?.tags?.includes(TagSpace);

  const remainingCostForSteel = effectiveCost - totalOtherPaymentValue(true, false);
  const maxSteelUnits = canUseSteel
    ? Math.min(playerResources.steel, Math.ceil(Math.max(0, remainingCostForSteel) / steelValue))
    : 0;

  const remainingCostForTitanium = effectiveCost - totalOtherPaymentValue(false, true);
  const maxTitaniumUnits = canUseTitanium
    ? Math.min(
        playerResources.titanium,
        Math.ceil(Math.max(0, remainingCostForTitanium) / titaniumValue),
      )
    : 0;

  const baseResource = genericPayment?.baseResource ?? "credit";
  const baseResourceMap: Record<string, number> = {
    heat: playerResources.heat,
    plant: playerResources.plants,
  };
  const baseResourceAvailable = baseResourceMap[baseResource] ?? playerResources.credits;

  const isOverpaying = finalCost < 0;
  const cannotAfford = finalCost > baseResourceAvailable;
  const canConfirm = !cannotAfford;

  const handleConfirm = useCallback(() => {
    if (!canConfirm) {
      return;
    }

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
    <GameFlowPopover
      isVisible={isVisible}
      onClose={onCancel}
      outerClassName="items-start pt-[30vh]"
      className="min-w-[400px] max-h-[600px]"
    >
      <GameFlowTitle>
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          Select Payment Method
        </h3>
        <div className="mt-2 flex items-center justify-between">
          <span className="text-sm text-gray-400">Pay for: {effectiveName}</span>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-400">Cost:</span>
            <GameIcon iconType="credit" amount={effectiveCost} size="small" />
          </div>
        </div>
      </GameFlowTitle>

      <div className="flex-1 overflow-y-auto p-4 space-y-3">
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
              <span className="w-12 text-center text-lg text-white font-semibold">{titanium}</span>
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

              if (available === 0) {
                return null;
              }

              const currentAmount = substitutes[resourceType] || 0;

              const remainingCostForThisSubstitute =
                effectiveCost - totalOtherPaymentValue(false, false, resourceType, undefined);

              const maxUnits = Math.min(
                available,
                Math.ceil(Math.max(0, remainingCostForThisSubstitute) / substitute.conversionRate),
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
                        {substitute.conversionRate}{" "}
                        {baseResource === "credit" ? "MC" : baseResource} each ({available}{" "}
                        available)
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

        {applicableStorageSubstitutes.map((substitute) => {
          const effectiveStorage = isGenericMode
            ? genericPayment?.resourceStorage
            : resourceStorage;
          const available = effectiveStorage?.[substitute.cardId] ?? 0;
          if (available === 0) {
            return null;
          }

          const currentAmount = storageSubstitutes[substitute.cardId] || 0;

          const remainingCostForThis =
            effectiveCost - totalOtherPaymentValue(false, false, undefined, substitute.cardId);

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
                    {substitute.conversionRate} {baseResource === "credit" ? "MC" : baseResource}{" "}
                    each ({available} available)
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

        {!canUseSteel &&
          !canUseTitanium &&
          (!playerPaymentSubstitutes || playerPaymentSubstitutes.length === 0) &&
          applicableStorageSubstitutes.length === 0 && (
            <div className="text-center text-gray-400 py-4">
              No alternative payment methods available
            </div>
          )}
      </div>

      <div className="px-4 pb-4">
        <div className="rounded-md border border-space-blue-500/30 bg-space-black/30 p-4">
          <div className="flex items-center justify-between">
            <span className="text-gray-400">
              {isGenericMode
                ? `${baseResource === "credit" ? "Credits" : baseResource.charAt(0).toUpperCase() + baseResource.slice(1)} needed:`
                : "Card cost:"}
            </span>
            <div
              className={`flex items-center gap-2 ${finalCost < 0 ? "opacity-90" : cannotAfford ? "opacity-90" : ""}`}
            >
              <GameIcon iconType={baseResource} amount={finalCost} size="medium" />
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
              Can't afford: Need {Math.max(finalCost - baseResourceAvailable, 0)} more{" "}
              {baseResource === "credit" ? "MC" : baseResource}
            </div>
          </div>
        </div>
      </div>

      <GameFlowFooter className="justify-end gap-3">
        <GameButton buttonType="secondary" variant="info" size="sm" onClick={onCancel}>
          Cancel
        </GameButton>
        <GameButton
          buttonType="primary"
          variant="info"
          size="sm"
          onClick={handleConfirm}
          disabled={!canConfirm}
        >
          Confirm
        </GameButton>
      </GameFlowFooter>
    </GameFlowPopover>
  );
};

export default PaymentSelectionPopover;
