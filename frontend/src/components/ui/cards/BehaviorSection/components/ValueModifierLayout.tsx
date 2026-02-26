import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import Slash from "./Slash.tsx";

interface ValueModifierLayoutProps {
  behavior: any;
}

const getResourcesFromSelectors = (selectors: any[]): string[] => {
  const resources: string[] = [];
  const seen = new Set<string>();
  selectors.forEach((selector: any) => {
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

const ValueModifierLayout: React.FC<ValueModifierLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const valueModifierOutput = behavior.outputs.find(
    (output: any) => output.type === "value-modifier",
  );
  if (!valueModifierOutput) return null;

  const amount = valueModifierOutput.amount ?? 1;
  const selectors: any[] = valueModifierOutput.selectors || [];
  const affectedResources = getResourcesFromSelectors(selectors);

  if (affectedResources.length === 0) return null;

  return (
    <div className="flex gap-[3px] items-center justify-center">
      {/* Left side: affected resources */}
      <div className="flex gap-[3px] items-center">
        {affectedResources.map((resourceType: string, resIndex: number) => (
          <React.Fragment key={`res-${resIndex}`}>
            {resIndex > 0 && <Slash />}
            <GameIcon iconType={resourceType} size="small" />
          </React.Fragment>
        ))}
      </div>

      {/* Separator: colon */}
      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {/* Plus sign */}
      <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        +
      </span>

      {/* Right side: Credits icon with amount inside */}
      <GameIcon iconType="credit" amount={amount} size="small" />
    </div>
  );
};

export default ValueModifierLayout;
