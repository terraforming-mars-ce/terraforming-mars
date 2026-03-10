import {
  CardDto,
  CardPaymentDto,
  PaymentConstantsDto,
  PaymentSubstituteDto,
  ResourcesDto,
  StoragePaymentSubstituteDto,
  TagBuilding,
  TagSpace,
} from "../types/generated/api-types.ts";

/**
 * Checks if a card matches any of a storage substitute's selectors.
 * Empty selectors means it applies to all cards.
 */
export function cardMatchesStorageSubstitute(
  card: CardDto,
  substitute: StoragePaymentSubstituteDto,
): boolean {
  if (substitute.selectors.length === 0) {
    return true;
  }
  return substitute.selectors.some((selector) => {
    if (selector.tags && selector.tags.length > 0) {
      return selector.tags.some((tag) => card.tags?.includes(tag));
    }
    return false;
  });
}

/**
 * Calculates the total MC value of a payment
 * This is for UI display only - backend validates actual payment
 */
export function calculatePaymentValue(
  payment: CardPaymentDto,
  constants: PaymentConstantsDto,
  playerSubstitutes?: PaymentSubstituteDto[],
  storageSubstitutes?: StoragePaymentSubstituteDto[],
): number {
  let total =
    payment.credits +
    payment.steel * constants.steelValue +
    payment.titanium * constants.titaniumValue;

  if (payment.substitutes && playerSubstitutes) {
    for (const [resourceType, amount] of Object.entries(payment.substitutes)) {
      const substitute = playerSubstitutes.find((sub) => sub.resourceType === resourceType);
      if (substitute) {
        total += amount * substitute.conversionRate;
      }
    }
  }

  if (payment.storageSubstitutes && storageSubstitutes) {
    for (const [cardId, amount] of Object.entries(payment.storageSubstitutes)) {
      const substitute = storageSubstitutes.find((sub) => sub.cardId === cardId);
      if (substitute) {
        total += amount * substitute.conversionRate;
      }
    }
  }

  return total;
}

/**
 * Creates a default all-credits payment for a card cost
 * Backend will validate if player can afford this
 */
export function createDefaultPayment(cardCost: number): CardPaymentDto {
  return {
    credits: cardCost,
    steel: 0,
    titanium: 0,
    substitutes: undefined,
    storageSubstitutes: undefined,
  };
}

/**
 * Gets applicable storage payment substitutes for a card, filtered by selectors
 * and available storage resources.
 */
export function getApplicableStorageSubstitutes(
  card: CardDto,
  storagePaymentSubstitutes: StoragePaymentSubstituteDto[],
  resourceStorage: { [key: string]: number },
): StoragePaymentSubstituteDto[] {
  return storagePaymentSubstitutes.filter((sub) => {
    const available = resourceStorage[sub.cardId] ?? 0;
    return available > 0 && cardMatchesStorageSubstitute(card, sub);
  });
}

/**
 * Determines if the payment modal should be shown
 * Show if:
 * 1. Card can use steel/titanium AND player has them, OR
 * 2. Player has any payment substitutes, OR
 * 3. Player has any applicable storage payment substitutes
 */
export function shouldShowPaymentModal(
  card: CardDto,
  playerResources: ResourcesDto,
  playerSubstitutes?: PaymentSubstituteDto[],
  storagePaymentSubstitutes?: StoragePaymentSubstituteDto[],
  resourceStorage?: { [key: string]: number },
): boolean {
  if (card.cost === 0) {
    return false;
  }

  const canUseSteel = (card.tags?.includes(TagBuilding) ?? false) && playerResources.steel > 0;
  const canUseTitanium = (card.tags?.includes(TagSpace) ?? false) && playerResources.titanium > 0;

  const hasUsableSubstitutes =
    playerSubstitutes &&
    playerSubstitutes.some((sub) => {
      const resourceType = sub.resourceType;
      switch (resourceType) {
        case "heat":
          return playerResources.heat > 0;
        case "energy":
          return playerResources.energy > 0;
        case "plant":
          return playerResources.plants > 0;
        default:
          return false;
      }
    });

  const hasUsableStorageSubstitutes =
    storagePaymentSubstitutes &&
    resourceStorage &&
    getApplicableStorageSubstitutes(card, storagePaymentSubstitutes, resourceStorage).length > 0;

  return canUseSteel || canUseTitanium || !!hasUsableSubstitutes || !!hasUsableStorageSubstitutes;
}
