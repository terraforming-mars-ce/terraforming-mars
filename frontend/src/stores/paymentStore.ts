import { useCardPlayFlowStore } from "./cardPlayFlowStore";
import { create } from "zustand";
import type { PaymentDto, PaymentIntentDto, PaymentQuoteDto } from "@/types/generated/api-types";

interface PendingPayment {
  intent: PaymentIntentDto;
  quote: PaymentQuoteDto;
  refresh: () => Promise<PaymentQuoteDto>;
  resolve: (payment: PaymentDto) => void;
  reject: (error: Error) => void;
}
export const usePaymentStore = create<{ pending: PendingPayment | null }>(() => ({
  pending: null,
}));

export function selectPayment(
  intent: PaymentIntentDto,
  quote: PaymentQuoteDto,
  refresh: () => Promise<PaymentQuoteDto>,
): Promise<PaymentDto> {
  usePaymentStore.getState().pending?.reject(new Error("Payment cancelled"));
  if (intent.action === "play-card" && intent.cardId) {
    useCardPlayFlowStore.getState().showPlayPrompt(intent.cardId);
  }
  return new Promise((resolve, reject) => {
    const pending: PendingPayment = {
      intent,
      quote,
      refresh,
      resolve: (payment) => {
        if (usePaymentStore.getState().pending !== pending) {
          return;
        }
        usePaymentStore.setState({ pending: null });
        resolve(payment);
      },
      reject: (error) => {
        if (usePaymentStore.getState().pending !== pending) {
          return;
        }
        usePaymentStore.setState({ pending: null });
        reject(error);
      },
    };
    usePaymentStore.setState({ pending });
  });
}
