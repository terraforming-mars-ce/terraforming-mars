import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import Slash from "./Slash.tsx";
import type { CardBehaviorDto, SelectorDto } from "@/types/generated/api-types";
import { isEffect, getSelectors } from "@/types/resourceConditions";

interface PaymentSubstituteLayoutProps {
  behavior: CardBehaviorDto;
}

const getResourcesFromSelectors = (selectors: SelectorDto[]): string[] => {
  const resources: string[] = [];
  const seen = new Set<string>();
  selectors.forEach((selector: SelectorDto) => {
    if (selector.resources) {
      selector.resources.forEach((r: string) => {
        if (!seen.has(r)) {
          seen.add(r);
          resources.push(r);
        }
      });
    }
  });
  return resources;
};

const PaymentSubstituteLayout: React.FC<PaymentSubstituteLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const paymentSubOutput = behavior.outputs.find(
    (output) => isEffect(output) && output.type === "payment-substitute",
  );
  if (!paymentSubOutput) return null;

  const amount = paymentSubOutput.amount ?? 1;
  const paymentSelectors = getSelectors(paymentSubOutput);
  const affectedResources = paymentSelectors ? getResourcesFromSelectors(paymentSelectors) : [];

  // Calculate ratio display
  // If amount < 1, invert the ratio and show multiplier on left side
  // E.g., amount=0.5 means 2x resources for X credits
  let leftMultiplier = null;
  let creditsText = "X";

  if (amount < 1 && amount > 0) {
    // Invert ratio: 0.5 → 2x, 0.25 → 4x, etc.
    leftMultiplier = Math.round(1 / amount);
  } else if (amount > 1) {
    // Show multiplier inside credits icon: 2 → "2X"
    creditsText = `${amount}X`;
  }

  return (
    <div className="flex gap-[3px] items-center justify-center">
      {/* Left side: multiplier (if applicable) + x + affected resources */}
      <div className="flex gap-[3px] items-center">
        {leftMultiplier && (
          <span className="font-orbitron text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-sm">
            {leftMultiplier}
          </span>
        )}
        <span className="font-orbitron text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-sm">
          x
        </span>
        {affectedResources.map((resourceType: string, resIndex: number) => (
          <React.Fragment key={`res-${resIndex}`}>
            {resIndex > 0 && <Slash />}
            <GameIcon iconType={resourceType} size="small" />
          </React.Fragment>
        ))}
      </div>

      {/* Separator */}
      <span className="font-orbitron text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {/* Right side: Credits icon with X or ratio inside */}
      <div className="relative flex items-center justify-center">
        <GameIcon iconType="credit" size="small" />
        <span className="absolute inset-0 flex items-center justify-center text-[13px] font-black font-orbitron text-black [text-shadow:0_0_2px_rgba(255,255,255,0.3)] tracking-[0.5px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] pointer-events-none max-md:text-[11px] translate-y-px">
          {creditsText}
        </span>
      </div>
    </div>
  );
};

export default PaymentSubstituteLayout;
