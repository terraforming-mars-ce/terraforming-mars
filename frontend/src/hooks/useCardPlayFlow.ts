import { useCallback, useRef } from "react";
import { useGameStore } from "@/stores/gameStore.ts";
import { useCardPlayFlowStore } from "@/stores/cardPlayFlowStore.ts";
import { useSpectateStore } from "@/stores/spectateStore.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { shouldShowPaymentModal, createDefaultPayment } from "@/utils/paymentUtils.ts";
import { StandardProject } from "@/types/cards.tsx";
import {
  getAllAnyCardStorageSelections,
  needsCardStorageSelection,
  needsTargetPlayerSelection,
  needsCardResourceInput,
  getVariableAmountInfo,
} from "@/utils/cardPlayUtils.ts";
import type {
  CardDto,
  CardPaymentDto,
  PlayerActionDto,
  ResourceType,
} from "@/types/generated/api-types.ts";

export function useCardPlayFlow() {
  const activeReuseSourceCardId = useRef<string | undefined>(undefined);

  const finalizePlayCard = useCallback(
    async (
      cardId: string,
      payment: CardPaymentDto,
      choiceIndex?: number,
      cardStorageTargets?: string[],
      cardForBehaviors?: CardDto,
      selectedAmount?: number,
    ) => {
      const cp = useGameStore.getState().currentPlayer;
      const card = cardForBehaviors || cp?.cards.find((c) => c.id === cardId);
      const store = useCardPlayFlowStore.getState();

      if (card && selectedAmount === undefined) {
        const autoTriggerBehaviors = card.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        for (const behavior of autoTriggerBehaviors || []) {
          const matchedChoice =
            choiceIndex !== undefined
              ? behavior.choices?.find((c) => c.originalIndex === choiceIndex)
              : undefined;
          const outputs = matchedChoice ? matchedChoice.outputs : behavior.outputs;
          const inputs = matchedChoice ? matchedChoice.inputs : behavior.inputs;
          const variableInfo = getVariableAmountInfo(inputs, outputs, cp);
          if (variableInfo) {
            store.setPendingVariableAmount({
              type: "play-card",
              cardId,
              cardName: card.name,
              payment,
              choiceIndex,
              cardStorageTargets,
              resourceLabel: variableInfo.resourceLabel,
              maxAmount: variableInfo.maxAmount,
            });
            store.setShowAmountSelection(true);
            return;
          }
        }
      }

      if (card) {
        const autoTriggerBehaviors = card.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        for (const behavior of autoTriggerBehaviors || []) {
          const matchedChoice2 =
            choiceIndex !== undefined
              ? behavior.choices?.find((c) => c.originalIndex === choiceIndex)
              : undefined;
          const outputs = matchedChoice2 ? matchedChoice2.outputs : behavior.outputs;
          const g = useGameStore.getState().game;
          const targetInfo = needsTargetPlayerSelection(outputs, g?.otherPlayers);
          if (targetInfo) {
            store.setPendingTargetPlayer({
              cardId,
              payment,
              choiceIndex,
              cardStorageTargets,
              selectedAmount,
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            store.setShowTargetPlayerSelection(true);
            return;
          }
        }
      }

      await globalWebSocketManager.playCard(
        cardId,
        payment,
        choiceIndex,
        cardStorageTargets,
        undefined,
        selectedAmount,
      );
    },
    [],
  );

  const handlePlayCard = useCallback(
    async (cardId: string) => {
      try {
        const g = useGameStore.getState().game;
        const cp = useGameStore.getState().currentPlayer;
        const store = useCardPlayFlowStore.getState();

        if (useSpectateStore.getState().spectatePlayerId) {
          return;
        }

        if (g?.currentTurn !== g?.viewingPlayerId) {
          throw new Error("Not your turn");
        }

        if (cp?.pendingTileSelection) {
          return;
        }

        const card = cp?.cards.find((c) => c.id === cardId);
        if (!card) {
          console.error(`Card ${cardId} not found in player's hand`);
          return;
        }

        const behaviorWithChoices = card.behaviors?.findIndex(
          (b) =>
            b.choices &&
            b.choices.length > 0 &&
            b.triggers?.some((t) => t.type === "auto") &&
            b.choicePolicy?.type !== "auto",
        );

        if (
          behaviorWithChoices !== undefined &&
          behaviorWithChoices >= 0 &&
          card.behaviors?.[behaviorWithChoices]?.choices
        ) {
          store.setCardPendingChoice(card);
          store.setPendingCardBehaviorIndex(behaviorWithChoices);
          store.setShowChoiceSelection(true);
        } else {
          const autoTriggerBehaviors = card.behaviors?.filter((b) =>
            b.triggers?.some((t) => t.type === "auto"),
          );

          const allStorageNeeds: Array<{
            resourceType: ResourceType;
            amount: number;
            selectorTags?: string[];
          }> = [];
          for (const behavior of autoTriggerBehaviors || []) {
            const selections = getAllAnyCardStorageSelections(behavior.outputs);
            for (const sel of selections) {
              allStorageNeeds.push({
                resourceType: sel.resourceType,
                amount: sel.amount,
                selectorTags: sel.selectorTags,
              });
            }
          }

          if (allStorageNeeds.length > 0) {
            const first = allStorageNeeds[0];
            store.setPendingCardStorage({
              cardId: card.id,
              choiceIndex: undefined,
              allStorageNeeds,
              collectedTargets: [],
              currentIndex: 0,
              resourceType: first.resourceType,
              amount: first.amount,
              selectorTags: first.selectorTags,
            });
            store.setShowCardStorageSelection(true);
          } else if (
            cp &&
            shouldShowPaymentModal(
              card,
              cp.resources,
              cp.paymentSubstitutes,
              cp.storagePaymentSubstitutes,
              cp.resourceStorage,
            )
          ) {
            store.setPendingCardPayment({
              card: card,
              choiceIndex: undefined,
            });
            store.setShowPaymentSelection(true);
          } else {
            const payment = createDefaultPayment(card.effectiveCost);
            await finalizePlayCard(cardId, payment, undefined, undefined, card);
          }
        }
      } catch (error) {
        console.error(`Failed to play card ${cardId}:`, error);
        throw error;
      }
    },
    [finalizePlayCard],
  );

  const handleChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      const store = useCardPlayFlowStore.getState();
      const { cardPendingChoice, pendingCardBehaviorIndex } = store;
      const currentPlayer = useGameStore.getState().currentPlayer;

      if (!cardPendingChoice || !currentPlayer) {
        return;
      }

      try {
        store.setShowChoiceSelection(false);

        const behavior = cardPendingChoice.behaviors?.[pendingCardBehaviorIndex];
        const selectedChoice = behavior?.choices?.find((c) => c.originalIndex === choiceIndex);

        const allStorageNeeds: Array<{
          resourceType: ResourceType;
          amount: number;
          selectorTags?: string[];
        }> = [];
        const choiceSelections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
        for (const sel of choiceSelections) {
          allStorageNeeds.push({
            resourceType: sel.resourceType,
            amount: sel.amount,
            selectorTags: sel.selectorTags,
          });
        }
        if (allStorageNeeds.length === 0) {
          const behaviorSelections = getAllAnyCardStorageSelections(behavior?.outputs);
          for (const sel of behaviorSelections) {
            allStorageNeeds.push({
              resourceType: sel.resourceType,
              amount: sel.amount,
              selectorTags: sel.selectorTags,
            });
          }
        }

        if (allStorageNeeds.length > 0) {
          const first = allStorageNeeds[0];
          store.setPendingCardStorage({
            cardId: cardPendingChoice.id,
            choiceIndex: choiceIndex,
            allStorageNeeds,
            collectedTargets: [],
            currentIndex: 0,
            resourceType: first.resourceType,
            amount: first.amount,
            selectorTags: first.selectorTags,
          });
          store.setShowCardStorageSelection(true);
          store.setCardPendingChoice(null);
          store.setPendingCardBehaviorIndex(0);
        } else if (
          shouldShowPaymentModal(
            cardPendingChoice,
            currentPlayer.resources,
            currentPlayer.paymentSubstitutes,
            currentPlayer.storagePaymentSubstitutes,
            currentPlayer.resourceStorage,
          )
        ) {
          store.setPendingCardPayment({
            card: cardPendingChoice,
            choiceIndex: choiceIndex,
          });
          store.setShowPaymentSelection(true);
          store.setCardPendingChoice(null);
          store.setPendingCardBehaviorIndex(0);
        } else {
          const payment = createDefaultPayment(cardPendingChoice.effectiveCost);
          await finalizePlayCard(
            cardPendingChoice.id,
            payment,
            choiceIndex,
            undefined,
            cardPendingChoice,
          );
          store.setCardPendingChoice(null);
          store.setPendingCardBehaviorIndex(0);
        }
      } catch (error) {
        console.error(
          `Failed to play card ${cardPendingChoice.id} with choice ${choiceIndex}:`,
          error,
        );
        store.setCardPendingChoice(null);
        store.setPendingCardBehaviorIndex(0);
      }
    },
    [finalizePlayCard],
  );

  const handleChoiceCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowChoiceSelection(false);
    store.setCardPendingChoice(null);
    store.setPendingCardBehaviorIndex(0);
  }, []);

  const handleActionChoiceSelect = useCallback(async (choiceIndex: number) => {
    const store = useCardPlayFlowStore.getState();
    const { actionPendingChoice } = store;

    if (!actionPendingChoice) {
      return;
    }

    try {
      store.setShowActionChoiceSelection(false);

      const selectedChoice = actionPendingChoice.behavior.choices?.find(
        (c) => c.originalIndex === choiceIndex,
      );

      const storageInfo = needsCardStorageSelection(selectedChoice?.outputs);

      if (storageInfo && storageInfo.target === "self-card") {
        const g = useGameStore.getState().game;
        const targetInfo = needsTargetPlayerSelection(selectedChoice?.outputs, g?.otherPlayers);
        if (targetInfo) {
          store.setPendingActionTargetPlayer({
            cardId: actionPendingChoice.cardId,
            behaviorIndex: actionPendingChoice.behaviorIndex,
            choiceIndex,
            cardStorageTargets: [actionPendingChoice.cardId],
            resourceType: targetInfo.resourceType,
            amount: targetInfo.amount,
            isSteal: targetInfo.isSteal,
          });
          store.setShowActionTargetPlayerSelection(true);
          store.setActionPendingChoice(null);
        } else {
          const reuseId = activeReuseSourceCardId.current;
          activeReuseSourceCardId.current = undefined;
          await globalWebSocketManager.playCardAction(
            actionPendingChoice.cardId,
            actionPendingChoice.behaviorIndex,
            choiceIndex,
            [actionPendingChoice.cardId],
            undefined,
            undefined,
            undefined,
            undefined,
            reuseId,
          );
          store.setActionPendingChoice(null);
        }
      } else {
        const allStorageNeeds: Array<{
          resourceType: ResourceType;
          amount: number;
          selectorTags?: string[];
        }> = [];
        const selections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
        for (const sel of selections) {
          allStorageNeeds.push({
            resourceType: sel.resourceType,
            amount: sel.amount,
            selectorTags: sel.selectorTags,
          });
        }

        if (allStorageNeeds.length > 0) {
          const first = allStorageNeeds[0];
          store.setPendingActionStorage({
            cardId: actionPendingChoice.cardId,
            behaviorIndex: actionPendingChoice.behaviorIndex,
            choiceIndex: choiceIndex,
            allStorageNeeds,
            collectedTargets: [],
            currentIndex: 0,
            resourceType: first.resourceType,
            amount: first.amount,
            selectorTags: first.selectorTags,
          });
          store.setShowActionStorageSelection(true);
          store.setActionPendingChoice(null);
        } else {
          const g = useGameStore.getState().game;

          // Check if choice has trade output - validate fleet/colonies before proceeding
          const hasFreeTrade = selectedChoice?.outputs?.some((o: any) => o.type === "trade");
          if (hasFreeTrade) {
            if (!g?.tradeFleetAvailable) {
              store.setPendingFreeTradeWarning("No trade fleet available");
              store.setShowFreeTradeWarning(true);
              store.setActionPendingChoice(null);
              return;
            }
            const tradeableColonies = (g?.colonies ?? []).filter((c) => !c.tradedThisGen);
            if (tradeableColonies.length === 0) {
              store.setPendingFreeTradeWarning("No colonies available for trading");
              store.setShowFreeTradeWarning(true);
              store.setActionPendingChoice(null);
              return;
            }
          }

          const targetInfo = needsTargetPlayerSelection(selectedChoice?.outputs, g?.otherPlayers);
          if (targetInfo) {
            store.setPendingActionTargetPlayer({
              cardId: actionPendingChoice.cardId,
              behaviorIndex: actionPendingChoice.behaviorIndex,
              choiceIndex,
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            store.setShowActionTargetPlayerSelection(true);
            store.setActionPendingChoice(null);
          } else {
            const reuseId = activeReuseSourceCardId.current;
            activeReuseSourceCardId.current = undefined;
            await globalWebSocketManager.playCardAction(
              actionPendingChoice.cardId,
              actionPendingChoice.behaviorIndex,
              choiceIndex,
              undefined,
              undefined,
              undefined,
              undefined,
              undefined,
              reuseId,
            );
            store.setActionPendingChoice(null);
          }
        }
      }
    } catch (error) {
      console.error(
        `Failed to play action ${actionPendingChoice.cardId} with choice ${choiceIndex}:`,
        error,
      );
      store.setActionPendingChoice(null);
      activeReuseSourceCardId.current = undefined;
    }
  }, []);

  const handleActionChoiceCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowActionChoiceSelection(false);
    store.setActionPendingChoice(null);
  }, []);

  const handleActionReuseSelect = useCallback((targetAction: PlayerActionDto) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingActionReuse } = store;

    if (!pendingActionReuse) {
      return;
    }
    store.setShowActionReuseSelection(false);

    activeReuseSourceCardId.current = pendingActionReuse.cardId;

    if (targetAction.behavior.choices && targetAction.behavior.choices.length > 0) {
      store.setActionPendingChoice(targetAction);
      store.setShowActionChoiceSelection(true);
    } else {
      const storageInfo = needsCardStorageSelection(targetAction.behavior.outputs);
      if (storageInfo && storageInfo.target === "self-card") {
        void globalWebSocketManager.playCardAction(
          targetAction.cardId,
          targetAction.behaviorIndex,
          undefined,
          [targetAction.cardId],
          undefined,
          undefined,
          undefined,
          undefined,
          pendingActionReuse.cardId,
        );
        activeReuseSourceCardId.current = undefined;
      } else {
        const allStorageNeeds: Array<{
          resourceType: ResourceType;
          amount: number;
          selectorTags?: string[];
        }> = [];
        const selections = getAllAnyCardStorageSelections(targetAction.behavior.outputs);
        for (const sel of selections) {
          allStorageNeeds.push({
            resourceType: sel.resourceType,
            amount: sel.amount,
            selectorTags: sel.selectorTags,
          });
        }

        if (allStorageNeeds.length > 0) {
          const first = allStorageNeeds[0];
          store.setPendingActionStorage({
            cardId: targetAction.cardId,
            behaviorIndex: targetAction.behaviorIndex,
            choiceIndex: undefined,
            allStorageNeeds,
            collectedTargets: [],
            currentIndex: 0,
            resourceType: first.resourceType,
            amount: first.amount,
            selectorTags: first.selectorTags,
          });
          store.setShowActionStorageSelection(true);
        } else {
          void globalWebSocketManager.playCardAction(
            targetAction.cardId,
            targetAction.behaviorIndex,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            pendingActionReuse.cardId,
          );
          activeReuseSourceCardId.current = undefined;
        }
      }
    }
    store.setPendingActionReuse(null);
  }, []);

  const handleActionReuseCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowActionReuseSelection(false);
    store.setPendingActionReuse(null);
    activeReuseSourceCardId.current = undefined;
  }, []);

  const handleBehaviorChoiceSelect = useCallback(async (choiceIndex: number) => {
    const g = useGameStore.getState().game;
    const pendingBehaviorChoice = g?.currentPlayer?.pendingBehaviorChoiceSelection;
    if (!pendingBehaviorChoice) {
      return;
    }

    const store = useCardPlayFlowStore.getState();

    try {
      const selectedChoice = pendingBehaviorChoice.choices.find(
        (c) => c.originalIndex === choiceIndex,
      );

      const allStorageNeeds: Array<{
        resourceType: ResourceType;
        amount: number;
        selectorTags?: string[];
      }> = [];
      const selections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
      for (const sel of selections) {
        allStorageNeeds.push({
          resourceType: sel.resourceType,
          amount: sel.amount,
          selectorTags: sel.selectorTags,
        });
      }

      if (allStorageNeeds.length > 0) {
        const first = allStorageNeeds[0];
        store.setPendingBehaviorChoiceStorage({
          choiceIndex,
          allStorageNeeds,
          collectedTargets: [],
          currentIndex: 0,
          resourceType: first.resourceType,
          amount: first.amount,
          selectorTags: first.selectorTags,
        });
        store.setShowBehaviorChoiceStorage(true);
        store.setShowBehaviorChoiceSelection(false);
      } else {
        await globalWebSocketManager.confirmBehaviorChoice(choiceIndex);
      }
    } catch (error) {
      console.error("Failed to confirm behavior choice:", error);
    }
  }, []);

  const handleBehaviorChoiceStorageSelect = useCallback(async (targetCardId: string) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingBehaviorChoiceStorage } = store;

    if (!pendingBehaviorChoiceStorage) {
      return;
    }

    try {
      const newCollected = [...pendingBehaviorChoiceStorage.collectedTargets, targetCardId];
      const nextIndex = pendingBehaviorChoiceStorage.currentIndex + 1;

      if (nextIndex < pendingBehaviorChoiceStorage.allStorageNeeds.length) {
        const next = pendingBehaviorChoiceStorage.allStorageNeeds[nextIndex];
        store.setPendingBehaviorChoiceStorage({
          ...pendingBehaviorChoiceStorage,
          collectedTargets: newCollected,
          currentIndex: nextIndex,
          resourceType: next.resourceType,
          amount: next.amount,
          selectorTags: next.selectorTags,
        });
      } else {
        store.setShowBehaviorChoiceStorage(false);
        await globalWebSocketManager.confirmBehaviorChoice(
          pendingBehaviorChoiceStorage.choiceIndex,
          newCollected,
        );
        store.setPendingBehaviorChoiceStorage(null);
      }
    } catch (error) {
      console.error("Failed to confirm behavior choice with storage target:", error);
      store.setPendingBehaviorChoiceStorage(null);
    }
  }, []);

  const handleBehaviorChoiceStorageCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowBehaviorChoiceStorage(false);
    store.setPendingBehaviorChoiceStorage(null);
    store.setShowBehaviorChoiceSelection(true);
  }, []);

  const handleStealTargetSelect = useCallback(async (targetPlayerId: string) => {
    void globalWebSocketManager.confirmStealTarget(targetPlayerId);
  }, []);

  const handleStealTargetSkip = useCallback(async () => {
    void globalWebSocketManager.confirmStealTarget("");
  }, []);

  const handleColonyResourceSelect = useCallback(async (cardId: string) => {
    void globalWebSocketManager.confirmColonyResource(cardId);
  }, []);

  const handleColonyResourceSkip = useCallback(async () => {
    void globalWebSocketManager.confirmColonyResource("");
  }, []);

  const handlePaymentConfirm = useCallback(
    async (payment: CardPaymentDto) => {
      const store = useCardPlayFlowStore.getState();
      const { pendingCardPayment } = store;

      if (!pendingCardPayment) {
        return;
      }
      const cp = useGameStore.getState().currentPlayer;
      if (!cp) {
        return;
      }

      try {
        store.setShowPaymentSelection(false);

        await finalizePlayCard(
          pendingCardPayment.card.id,
          payment,
          pendingCardPayment.choiceIndex,
          pendingCardPayment.cardStorageTargets,
          pendingCardPayment.card,
        );

        store.setPendingCardPayment(null);
      } catch (error) {
        console.error("Failed to play card with payment:", error);
        store.setPendingCardPayment(null);
      }
    },
    [finalizePlayCard],
  );

  const handlePaymentCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowPaymentSelection(false);
    store.setPendingCardPayment(null);
  }, []);

  const handleCardStorageSelect = useCallback(
    async (targetCardId: string) => {
      const store = useCardPlayFlowStore.getState();
      const { pendingCardStorage } = store;

      if (!pendingCardStorage) {
        return;
      }
      const cp = useGameStore.getState().currentPlayer;
      if (!cp) {
        return;
      }

      try {
        const newCollected = [...pendingCardStorage.collectedTargets, targetCardId];
        const nextIndex = pendingCardStorage.currentIndex + 1;

        if (nextIndex < pendingCardStorage.allStorageNeeds.length) {
          const next = pendingCardStorage.allStorageNeeds[nextIndex];
          store.setPendingCardStorage({
            ...pendingCardStorage,
            collectedTargets: newCollected,
            currentIndex: nextIndex,
            resourceType: next.resourceType,
            amount: next.amount,
            selectorTags: next.selectorTags,
          });
          return;
        }

        store.setShowCardStorageSelection(false);
        const card = cp.cards.find((c) => c.id === pendingCardStorage.cardId);

        if (
          card &&
          shouldShowPaymentModal(
            card,
            cp.resources,
            cp.paymentSubstitutes,
            cp.storagePaymentSubstitutes,
            cp.resourceStorage,
          )
        ) {
          store.setPendingCardPayment({
            card: card,
            choiceIndex: pendingCardStorage.choiceIndex,
            cardStorageTargets: newCollected,
          });
          store.setShowPaymentSelection(true);
          store.setPendingCardStorage(null);
          return;
        }

        const payment = createDefaultPayment(card?.effectiveCost ?? 0);
        await finalizePlayCard(
          pendingCardStorage.cardId,
          payment,
          pendingCardStorage.choiceIndex,
          newCollected,
          card,
        );
        store.setPendingCardStorage(null);
      } catch (error) {
        console.error(
          `Failed to play card ${pendingCardStorage.cardId} with card storage target ${targetCardId}:`,
          error,
        );
        store.setPendingCardStorage(null);
      }
    },
    [finalizePlayCard],
  );

  const handleCardStorageCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowCardStorageSelection(false);
    store.setPendingCardStorage(null);
  }, []);

  const handleActionStorageSelect = useCallback(async (targetCardId: string) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingActionStorage } = store;

    if (!pendingActionStorage) {
      return;
    }

    try {
      const newCollected = [...pendingActionStorage.collectedTargets, targetCardId];
      const nextIndex = pendingActionStorage.currentIndex + 1;

      if (nextIndex < pendingActionStorage.allStorageNeeds.length) {
        const next = pendingActionStorage.allStorageNeeds[nextIndex];
        store.setPendingActionStorage({
          ...pendingActionStorage,
          collectedTargets: newCollected,
          currentIndex: nextIndex,
          resourceType: next.resourceType,
          amount: next.amount,
          selectorTags: next.selectorTags,
        });
        return;
      }

      store.setShowActionStorageSelection(false);

      const cp = useGameStore.getState().currentPlayer;
      const action = cp?.actions?.find(
        (a) =>
          a.cardId === pendingActionStorage.cardId &&
          a.behaviorIndex === pendingActionStorage.behaviorIndex,
      );
      const outputs =
        pendingActionStorage.choiceIndex !== undefined
          ? action?.behavior.choices?.find(
              (c) => c.originalIndex === pendingActionStorage.choiceIndex,
            )?.outputs
          : action?.behavior.outputs;
      const g = useGameStore.getState().game;
      const targetInfo = needsTargetPlayerSelection(outputs, g?.otherPlayers);

      if (targetInfo) {
        store.setPendingActionTargetPlayer({
          cardId: pendingActionStorage.cardId,
          behaviorIndex: pendingActionStorage.behaviorIndex,
          choiceIndex: pendingActionStorage.choiceIndex,
          cardStorageTargets: newCollected,
          resourceType: targetInfo.resourceType,
          amount: targetInfo.amount,
          isSteal: targetInfo.isSteal,
        });
        store.setShowActionTargetPlayerSelection(true);
        store.setPendingActionStorage(null);
        return;
      }

      const reuseId = activeReuseSourceCardId.current;
      activeReuseSourceCardId.current = undefined;
      await globalWebSocketManager.playCardAction(
        pendingActionStorage.cardId,
        pendingActionStorage.behaviorIndex,
        pendingActionStorage.choiceIndex,
        newCollected,
        undefined,
        undefined,
        undefined,
        undefined,
        reuseId,
      );
      store.setPendingActionStorage(null);
    } catch (error) {
      console.error(
        `Failed to play action ${pendingActionStorage.cardId} with card storage target ${targetCardId}:`,
        error,
      );
      store.setPendingActionStorage(null);
      activeReuseSourceCardId.current = undefined;
    }
  }, []);

  const handleActionStorageCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowActionStorageSelection(false);
    store.setPendingActionStorage(null);
    activeReuseSourceCardId.current = undefined;
  }, []);

  const handleTargetPlayerSelect = useCallback(async (targetPlayerId: string) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingTargetPlayer } = store;

    if (!pendingTargetPlayer) {
      return;
    }

    try {
      store.setShowTargetPlayerSelection(false);
      await globalWebSocketManager.playCard(
        pendingTargetPlayer.cardId,
        pendingTargetPlayer.payment,
        pendingTargetPlayer.choiceIndex,
        pendingTargetPlayer.cardStorageTargets,
        targetPlayerId,
        pendingTargetPlayer.selectedAmount,
      );
      store.setPendingTargetPlayer(null);
    } catch (error) {
      console.error(
        `Failed to play card ${pendingTargetPlayer.cardId} with target player ${targetPlayerId}:`,
        error,
      );
      store.setPendingTargetPlayer(null);
    }
  }, []);

  const handleTargetPlayerCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowTargetPlayerSelection(false);
    store.setPendingTargetPlayer(null);
  }, []);

  const handleAmountSelect = useCallback(
    async (amount: number) => {
      const store = useCardPlayFlowStore.getState();
      const { pendingVariableAmount } = store;

      if (!pendingVariableAmount) {
        return;
      }

      try {
        store.setShowAmountSelection(false);
        if (pendingVariableAmount.type === "play-card") {
          await finalizePlayCard(
            pendingVariableAmount.cardId,
            pendingVariableAmount.payment!,
            pendingVariableAmount.choiceIndex,
            pendingVariableAmount.cardStorageTargets,
            undefined,
            amount,
          );
        } else if (pendingVariableAmount.type === "card-action") {
          await globalWebSocketManager.playCardAction(
            pendingVariableAmount.cardId,
            pendingVariableAmount.behaviorIndex!,
            pendingVariableAmount.choiceIndex,
            pendingVariableAmount.cardStorageTargets,
            undefined,
            undefined,
            amount,
          );
        }
        store.setPendingVariableAmount(null);
      } catch (error) {
        console.error(`Failed to execute with amount ${amount}:`, error);
        store.setPendingVariableAmount(null);
      }
    },
    [finalizePlayCard],
  );

  const handleAmountCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowAmountSelection(false);
    store.setPendingVariableAmount(null);
  }, []);

  const handleActionTargetPlayerSelect = useCallback(async (targetPlayerId: string) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingActionTargetPlayer } = store;

    if (!pendingActionTargetPlayer) {
      return;
    }

    try {
      store.setShowActionTargetPlayerSelection(false);
      await globalWebSocketManager.playCardAction(
        pendingActionTargetPlayer.cardId,
        pendingActionTargetPlayer.behaviorIndex,
        pendingActionTargetPlayer.choiceIndex,
        pendingActionTargetPlayer.cardStorageTargets,
        targetPlayerId,
      );
      store.setPendingActionTargetPlayer(null);
    } catch (error) {
      console.error(
        `Failed to play action ${pendingActionTargetPlayer.cardId} with target player ${targetPlayerId}:`,
        error,
      );
      store.setPendingActionTargetPlayer(null);
    }
  }, []);

  const handleActionTargetPlayerCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowActionTargetPlayerSelection(false);
    store.setPendingActionTargetPlayer(null);
  }, []);

  const handleCardResourceSelect = useCallback(async (sourceCardId: string) => {
    const store = useCardPlayFlowStore.getState();
    const { pendingCardResourceInput } = store;

    if (!pendingCardResourceInput) {
      return;
    }

    try {
      store.setShowCardResourceSelection(false);
      await globalWebSocketManager.playCardAction(
        pendingCardResourceInput.cardId,
        pendingCardResourceInput.behaviorIndex,
        pendingCardResourceInput.choiceIndex,
        pendingCardResourceInput.cardStorageTargets,
        undefined,
        sourceCardId,
      );
      store.setPendingCardResourceInput(null);
    } catch (error) {
      console.error(
        `Failed to play action ${pendingCardResourceInput.cardId} with source card ${sourceCardId}:`,
        error,
      );
      store.setPendingCardResourceInput(null);
    }
  }, []);

  const handleCardResourceCancel = useCallback(() => {
    const store = useCardPlayFlowStore.getState();
    store.setShowCardResourceSelection(false);
    store.setPendingCardResourceInput(null);
  }, []);

  const handleActionSelect = useCallback((action: PlayerActionDto) => {
    const cp = useGameStore.getState().currentPlayer;
    if (cp?.pendingTileSelection) {
      return;
    }

    const store = useCardPlayFlowStore.getState();

    const isActionReuse = action.behavior.outputs?.some(
      (o: { type: string }) => o.type === "action-reuse",
    );
    if (isActionReuse) {
      store.setPendingActionReuse({
        cardId: action.cardId,
        behaviorIndex: action.behaviorIndex,
      });
      store.setShowActionReuseSelection(true);
      return;
    }

    if (action.behavior.choices && action.behavior.choices.length > 0) {
      store.setActionPendingChoice(action);
      store.setShowActionChoiceSelection(true);
    } else {
      const cp2 = useGameStore.getState().currentPlayer;
      const variableInfo = getVariableAmountInfo(
        action.behavior.inputs,
        action.behavior.outputs,
        cp2,
      );
      if (variableInfo) {
        store.setPendingVariableAmount({
          type: "card-action",
          cardId: action.cardId,
          cardName: action.cardName,
          behaviorIndex: action.behaviorIndex,
          resourceLabel: variableInfo.resourceLabel,
          maxAmount: variableInfo.maxAmount,
        });
        store.setShowAmountSelection(true);
        return;
      }

      const cardResourceInfo = needsCardResourceInput(
        action.behavior.inputs,
        action.behavior.outputs,
      );

      if (cardResourceInfo) {
        const storageInfo = needsCardStorageSelection(action.behavior.outputs);
        const cardStorageTargets =
          storageInfo?.target === "self-card" ? [action.cardId] : undefined;

        store.setPendingCardResourceInput({
          cardId: action.cardId,
          behaviorIndex: action.behaviorIndex,
          cardStorageTargets,
          resourceType: cardResourceInfo.resourceType,
          amount: cardResourceInfo.amount,
        });
        store.setShowCardResourceSelection(true);
      } else {
        const storageInfo = needsCardStorageSelection(action.behavior.outputs);

        if (storageInfo && storageInfo.target === "self-card") {
          const g = useGameStore.getState().game;
          const targetInfo = needsTargetPlayerSelection(action.behavior.outputs, g?.otherPlayers);
          if (targetInfo) {
            store.setPendingActionTargetPlayer({
              cardId: action.cardId,
              behaviorIndex: action.behaviorIndex,
              cardStorageTargets: [action.cardId],
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            store.setShowActionTargetPlayerSelection(true);
          } else {
            void globalWebSocketManager.playCardAction(
              action.cardId,
              action.behaviorIndex,
              undefined,
              [action.cardId],
            );
          }
        } else {
          const allStorageNeeds: Array<{
            resourceType: ResourceType;
            amount: number;
            selectorTags?: string[];
          }> = [];
          const selections = getAllAnyCardStorageSelections(action.behavior.outputs);
          for (const sel of selections) {
            allStorageNeeds.push({
              resourceType: sel.resourceType,
              amount: sel.amount,
              selectorTags: sel.selectorTags,
            });
          }

          if (allStorageNeeds.length > 0) {
            const first = allStorageNeeds[0];
            store.setPendingActionStorage({
              cardId: action.cardId,
              behaviorIndex: action.behaviorIndex,
              allStorageNeeds,
              collectedTargets: [],
              currentIndex: 0,
              resourceType: first.resourceType,
              amount: first.amount,
              selectorTags: first.selectorTags,
            });
            store.setShowActionStorageSelection(true);
          } else {
            const g = useGameStore.getState().game;
            const targetInfo = needsTargetPlayerSelection(action.behavior.outputs, g?.otherPlayers);
            if (targetInfo) {
              store.setPendingActionTargetPlayer({
                cardId: action.cardId,
                behaviorIndex: action.behaviorIndex,
                resourceType: targetInfo.resourceType,
                amount: targetInfo.amount,
                isSteal: targetInfo.isSteal,
              });
              store.setShowActionTargetPlayerSelection(true);
            } else {
              void globalWebSocketManager.playCardAction(action.cardId, action.behaviorIndex);
            }
          }
        }
      }
    }
  }, []);

  const handleStandardProjectSelect = useCallback((project: StandardProject) => {
    const cp = useGameStore.getState().currentPlayer;
    if (cp?.pendingTileSelection) {
      return;
    }

    void globalWebSocketManager.standardProject(project);
  }, []);

  const handleConvertPlantsToGreenery = useCallback(() => {
    const cp = useGameStore.getState().currentPlayer;
    if (cp?.pendingTileSelection) {
      return;
    }

    void globalWebSocketManager.convertPlantsToGreenery();
  }, []);

  const handleConvertHeatToTemperature = useCallback(() => {
    void globalWebSocketManager.convertHeatToTemperature();
  }, []);

  return {
    handlePlayCard,
    handleChoiceSelect,
    handleChoiceCancel,
    handlePaymentConfirm,
    handlePaymentCancel,
    handleCardStorageSelect,
    handleCardStorageCancel,
    handleTargetPlayerSelect,
    handleTargetPlayerCancel,
    handleAmountSelect,
    handleAmountCancel,
    handleActionChoiceSelect,
    handleActionChoiceCancel,
    handleActionStorageSelect,
    handleActionStorageCancel,
    handleActionTargetPlayerSelect,
    handleActionTargetPlayerCancel,
    handleCardResourceSelect,
    handleCardResourceCancel,
    handleActionReuseSelect,
    handleActionReuseCancel,
    handleBehaviorChoiceSelect,
    handleBehaviorChoiceStorageSelect,
    handleBehaviorChoiceStorageCancel,
    handleStealTargetSelect,
    handleStealTargetSkip,
    handleColonyResourceSelect,
    handleColonyResourceSkip,
    handleStandardProjectSelect,
    handleConvertPlantsToGreenery,
    handleConvertHeatToTemperature,
    handleActionSelect,
    finalizePlayCard,
  };
}
