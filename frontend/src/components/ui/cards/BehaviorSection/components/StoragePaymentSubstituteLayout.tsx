import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import type { CardBehaviorDto, SelectorDto } from "@/types/generated/api-types";
import { isEffect, getSelectors } from "@/types/resourceConditions";

interface StoragePaymentSubstituteLayoutProps {
  behavior: CardBehaviorDto;
  storageResourceType?: string;
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

const StoragePaymentSubstituteLayout: React.FC<StoragePaymentSubstituteLayoutProps> = ({
  behavior,
  storageResourceType = "floater",
}) => {
  if (!behavior.outputs || behavior.outputs.length === 0) {
    return null;
  }

  const subOutput = behavior.outputs.find(
    (output) => isEffect(output) && output.type === "storage-payment-substitute",
  );
  if (!subOutput) {
    return null;
  }

  const conversionRate = subOutput.amount ?? 1;
  const selectors = getSelectors(subOutput);
  const targetResources = selectors ? getResourcesFromSelectors(selectors) : [];
  const targetResource = targetResources.length > 0 ? targetResources[0] : "credit";

  return (
    <div className="flex gap-[3px] items-center justify-center">
      <span className="font-orbitron text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-sm">
        x
      </span>
      <GameIcon iconType={storageResourceType} size="small" />

      <span className="font-orbitron text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {targetResource === "credit" ? (
        <GameIcon iconType="credit" amount={conversionRate} size="small" />
      ) : (
        <>
          {conversionRate > 1 && (
            <span className="font-orbitron text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-sm">
              {conversionRate}
            </span>
          )}
          <GameIcon iconType={targetResource} size="small" />
        </>
      )}
    </div>
  );
};

export default StoragePaymentSubstituteLayout;
