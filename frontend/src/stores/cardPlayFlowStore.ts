import { create } from "zustand";
import type {
  CardPaymentDto,
  PlayerActionDto,
  PlayerCardDto,
  ResourceType,
} from "@/types/generated/api-types.ts";

export interface CardStoragePending {
  cardId: string;
  choiceIndex?: number;
  allStorageNeeds: Array<{
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  }>;
  collectedTargets: string[];
  currentIndex: number;
  resourceType: ResourceType;
  amount: number;
  selectorTags?: string[];
}

export interface CardPaymentPending {
  card: PlayerCardDto;
  choiceIndex?: number;
  cardStorageTargets?: string[];
}

export interface TargetPlayerPending {
  cardId: string;
  payment: CardPaymentDto;
  choiceIndex?: number;
  cardStorageTargets?: string[];
  selectedAmount?: number;
  resourceType: ResourceType;
  amount: number;
  isSteal: boolean;
}

export interface ActionStoragePending {
  cardId: string;
  behaviorIndex: number;
  choiceIndex?: number;
  allStorageNeeds: Array<{
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  }>;
  collectedTargets: string[];
  currentIndex: number;
  resourceType: ResourceType;
  amount: number;
  selectorTags?: string[];
}

export interface ActionTargetPlayerPending {
  cardId: string;
  behaviorIndex: number;
  choiceIndex?: number;
  cardStorageTargets?: string[];
  resourceType: ResourceType;
  amount: number;
  isSteal: boolean;
}

export interface CardResourceInputPending {
  cardId: string;
  behaviorIndex: number;
  choiceIndex?: number;
  cardStorageTargets?: string[];
  resourceType: ResourceType;
  amount: number;
}

export interface VariableAmountPending {
  type: "play-card" | "card-action";
  cardId: string;
  cardName: string;
  payment?: CardPaymentDto;
  choiceIndex?: number;
  cardStorageTargets?: string[];
  behaviorIndex?: number;
  resourceLabel: string;
  maxAmount: number;
}

export interface BehaviorChoiceStoragePending {
  choiceIndex: number;
  allStorageNeeds: Array<{
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  }>;
  collectedTargets: string[];
  currentIndex: number;
  resourceType: ResourceType;
  amount: number;
  selectorTags?: string[];
}

interface CardPlayFlowState {
  // Card play flow
  showChoiceSelection: boolean;
  cardPendingChoice: PlayerCardDto | null;
  pendingCardBehaviorIndex: number;

  showCardStorageSelection: boolean;
  pendingCardStorage: CardStoragePending | null;

  showPaymentSelection: boolean;
  pendingCardPayment: CardPaymentPending | null;

  showTargetPlayerSelection: boolean;
  pendingTargetPlayer: TargetPlayerPending | null;

  showAmountSelection: boolean;
  pendingVariableAmount: VariableAmountPending | null;

  // Action flow
  showActionChoiceSelection: boolean;
  actionPendingChoice: PlayerActionDto | null;

  showActionStorageSelection: boolean;
  pendingActionStorage: ActionStoragePending | null;

  showActionTargetPlayerSelection: boolean;
  pendingActionTargetPlayer: ActionTargetPlayerPending | null;

  showCardResourceSelection: boolean;
  pendingCardResourceInput: CardResourceInputPending | null;

  showActionReuseSelection: boolean;
  pendingActionReuse: { cardId: string; behaviorIndex: number } | null;

  showFreeTradeWarning: boolean;
  pendingFreeTradeWarning: string | null;

  // Behavior choice flow (passive triggered)
  showBehaviorChoiceSelection: boolean;
  showBehaviorChoiceStorage: boolean;
  pendingBehaviorChoiceStorage: BehaviorChoiceStoragePending | null;

  // Setters
  setShowChoiceSelection: (show: boolean) => void;
  setCardPendingChoice: (card: PlayerCardDto | null) => void;
  setPendingCardBehaviorIndex: (index: number) => void;
  setShowCardStorageSelection: (show: boolean) => void;
  setPendingCardStorage: (pending: CardStoragePending | null) => void;
  setShowPaymentSelection: (show: boolean) => void;
  setPendingCardPayment: (pending: CardPaymentPending | null) => void;
  setShowTargetPlayerSelection: (show: boolean) => void;
  setPendingTargetPlayer: (pending: TargetPlayerPending | null) => void;
  setShowAmountSelection: (show: boolean) => void;
  setPendingVariableAmount: (pending: VariableAmountPending | null) => void;
  setShowActionChoiceSelection: (show: boolean) => void;
  setActionPendingChoice: (action: PlayerActionDto | null) => void;
  setShowActionStorageSelection: (show: boolean) => void;
  setPendingActionStorage: (pending: ActionStoragePending | null) => void;
  setShowActionTargetPlayerSelection: (show: boolean) => void;
  setPendingActionTargetPlayer: (pending: ActionTargetPlayerPending | null) => void;
  setShowCardResourceSelection: (show: boolean) => void;
  setPendingCardResourceInput: (pending: CardResourceInputPending | null) => void;
  setShowActionReuseSelection: (show: boolean) => void;
  setPendingActionReuse: (pending: { cardId: string; behaviorIndex: number } | null) => void;
  setShowFreeTradeWarning: (show: boolean) => void;
  setPendingFreeTradeWarning: (warning: string | null) => void;
  setShowBehaviorChoiceSelection: (show: boolean) => void;
  setShowBehaviorChoiceStorage: (show: boolean) => void;
  setPendingBehaviorChoiceStorage: (pending: BehaviorChoiceStoragePending | null) => void;

  resetCardPlayFlow: () => void;
  resetActionFlow: () => void;
  resetAll: () => void;
}

const cardPlayFlowInitial = {
  showChoiceSelection: false,
  cardPendingChoice: null,
  pendingCardBehaviorIndex: 0,
  showCardStorageSelection: false,
  pendingCardStorage: null,
  showPaymentSelection: false,
  pendingCardPayment: null,
  showTargetPlayerSelection: false,
  pendingTargetPlayer: null,
  showAmountSelection: false,
  pendingVariableAmount: null,
};

const actionFlowInitial = {
  showActionChoiceSelection: false,
  actionPendingChoice: null,
  showActionStorageSelection: false,
  pendingActionStorage: null,
  showActionTargetPlayerSelection: false,
  pendingActionTargetPlayer: null,
  showCardResourceSelection: false,
  pendingCardResourceInput: null,
  showActionReuseSelection: false,
  pendingActionReuse: null,
  showFreeTradeWarning: false,
  pendingFreeTradeWarning: null,
};

const behaviorChoiceInitial = {
  showBehaviorChoiceSelection: false,
  showBehaviorChoiceStorage: false,
  pendingBehaviorChoiceStorage: null,
};

const allInitial = {
  ...cardPlayFlowInitial,
  ...actionFlowInitial,
  ...behaviorChoiceInitial,
};

export const useCardPlayFlowStore = create<CardPlayFlowState>((set) => ({
  ...allInitial,

  setShowChoiceSelection: (show) => set({ showChoiceSelection: show }),
  setCardPendingChoice: (card) => set({ cardPendingChoice: card }),
  setPendingCardBehaviorIndex: (index) => set({ pendingCardBehaviorIndex: index }),
  setShowCardStorageSelection: (show) => set({ showCardStorageSelection: show }),
  setPendingCardStorage: (pending) => set({ pendingCardStorage: pending }),
  setShowPaymentSelection: (show) => set({ showPaymentSelection: show }),
  setPendingCardPayment: (pending) => set({ pendingCardPayment: pending }),
  setShowTargetPlayerSelection: (show) => set({ showTargetPlayerSelection: show }),
  setPendingTargetPlayer: (pending) => set({ pendingTargetPlayer: pending }),
  setShowAmountSelection: (show) => set({ showAmountSelection: show }),
  setPendingVariableAmount: (pending) => set({ pendingVariableAmount: pending }),
  setShowActionChoiceSelection: (show) => set({ showActionChoiceSelection: show }),
  setActionPendingChoice: (action) => set({ actionPendingChoice: action }),
  setShowActionStorageSelection: (show) => set({ showActionStorageSelection: show }),
  setPendingActionStorage: (pending) => set({ pendingActionStorage: pending }),
  setShowActionTargetPlayerSelection: (show) => set({ showActionTargetPlayerSelection: show }),
  setPendingActionTargetPlayer: (pending) => set({ pendingActionTargetPlayer: pending }),
  setShowCardResourceSelection: (show) => set({ showCardResourceSelection: show }),
  setPendingCardResourceInput: (pending) => set({ pendingCardResourceInput: pending }),
  setShowActionReuseSelection: (show) => set({ showActionReuseSelection: show }),
  setPendingActionReuse: (pending) => set({ pendingActionReuse: pending }),
  setShowFreeTradeWarning: (show) => set({ showFreeTradeWarning: show }),
  setPendingFreeTradeWarning: (warning) => set({ pendingFreeTradeWarning: warning }),
  setShowBehaviorChoiceSelection: (show) => set({ showBehaviorChoiceSelection: show }),
  setShowBehaviorChoiceStorage: (show) => set({ showBehaviorChoiceStorage: show }),
  setPendingBehaviorChoiceStorage: (pending) => set({ pendingBehaviorChoiceStorage: pending }),

  resetCardPlayFlow: () => set(cardPlayFlowInitial),
  resetActionFlow: () => set(actionFlowInitial),
  resetAll: () => set(allInitial),
}));
