import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import Slash from "./Slash.tsx";

interface DiscountLayoutProps {
  behavior: any;
}

const getStandardProjectIcon = (project: string): string | null => {
  const mapping: { [key: string]: string } = {
    "power-plant": "power-tag", // Power tag icon for power plant SP
    "convert-plants-to-greenery": "greenery-tile",
    "convert-heat-to-temperature": "heat",
    aquifer: "ocean-tile",
    asteroid: "temperature",
    "air-scrapping": "venus",
  };
  return mapping[project] || null;
};

const getActionIcon = (action: string): string | null => {
  const mapping: { [key: string]: string } = {
    "card-buying": "card-buy",
    "colony-trade": "trade",
  };
  return mapping[action] || null;
};

const IconWithBadge: React.FC<{
  iconType: string;
  showSpBadge?: boolean;
}> = ({ iconType, showSpBadge = false }) => {
  return (
    <div className="relative inline-flex items-center justify-center">
      <GameIcon iconType={iconType} size="small" />
      {showSpBadge && (
        <span className="absolute -bottom-[2px] -right-[2px] text-[8px] font-black text-white bg-[rgba(80,80,80,0.9)] px-[3px] py-[1px] rounded-[2px] leading-none [text-shadow:0_0_2px_rgba(0,0,0,0.8)]">
          SP
        </span>
      )}
    </div>
  );
};

const DiscountAmount: React.FC<{
  amount: number;
  resourceTypes: string[];
  isNegativeDiscount?: boolean;
}> = ({ amount, resourceTypes, isNegativeDiscount = false }) => {
  const sign = isNegativeDiscount ? "+" : "-";

  if (resourceTypes.length === 1 && resourceTypes[0] === "credit") {
    return (
      <div className="flex items-center gap-0.5">
        <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
          {sign}
        </span>
        <GameIcon iconType="credit" amount={amount} size="small" />
      </div>
    );
  }

  return (
    <div className="flex items-center gap-[2px]">
      <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        {sign}
      </span>
      {resourceTypes.map((rt, i) => {
        const isCredit = rt === "credit" || rt === "credit-production";
        return (
          <GameIcon key={i} iconType={rt} size="small" amount={isCredit ? amount : undefined} />
        );
      })}
    </div>
  );
};

const DiscountRow: React.FC<{
  icons: React.ReactNode;
  amount: number;
  resourceTypes: string[];
  isNegativeDiscount?: boolean;
}> = ({ icons, amount, resourceTypes, isNegativeDiscount = false }) => {
  return (
    <div className="flex gap-[3px] items-center justify-center">
      <div className="flex gap-[3px] items-center">{icons}</div>

      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      <DiscountAmount
        amount={amount}
        resourceTypes={resourceTypes}
        isNegativeDiscount={isNegativeDiscount}
      />
    </div>
  );
};

const renderSelectorIcons = (selector: any): React.ReactNode => {
  const elements: React.ReactNode[] = [];

  if (selector.tags && selector.tags.length > 0) {
    selector.tags.forEach((tag: string, tagIndex: number) => {
      elements.push(
        <IconWithBadge
          key={`tag-${tagIndex}`}
          iconType={`${tag.toLowerCase()}-tag`}
          showSpBadge={false}
        />,
      );
    });
  }

  if (selector.cardTypes && selector.cardTypes.length > 0) {
    selector.cardTypes.forEach((cardType: string, typeIndex: number) => {
      if (cardType === "event") {
        elements.push(<GameIcon key={`type-${typeIndex}`} iconType="event" size="small" />);
      } else {
        elements.push(
          <span
            key={`type-${typeIndex}`}
            className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]"
          >
            {cardType}
          </span>,
        );
      }
    });
  }

  if (selector.standardProjects && selector.standardProjects.length > 0) {
    selector.standardProjects.forEach((project: string, spIndex: number) => {
      const iconType = getStandardProjectIcon(project);
      if (iconType) {
        elements.push(
          <IconWithBadge key={`sp-${spIndex}`} iconType={iconType} showSpBadge={true} />,
        );
      }
    });
  }

  if (selector.actions && selector.actions.length > 0) {
    selector.actions.forEach((action: string, actionIndex: number) => {
      const iconType = getActionIcon(action);
      if (iconType) {
        elements.push(<GameIcon key={`action-${actionIndex}`} iconType={iconType} size="small" />);
      }
    });
  }

  return elements;
};

const DiscountLayout: React.FC<DiscountLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const discountOutput = behavior.outputs.find((output: any) => output.type === "discount");
  if (!discountOutput) return null;

  const rawAmount = discountOutput.amount ?? 0;
  const isNegativeDiscount = rawAmount < 0;
  const amount = Math.abs(rawAmount);
  const selectors: any[] = discountOutput.selectors || [];

  // Extract affected resource types from selectors
  const affectedResources: string[] = [];
  for (const sel of selectors) {
    for (const r of sel.resources ?? []) {
      if (!affectedResources.includes(r)) {
        affectedResources.push(r);
      }
    }
  }
  if (affectedResources.length === 0) {
    affectedResources.push("credit");
  }

  // If we have selectors, render them with OR separators between selectors
  if (selectors.length > 0) {
    const selectorIcons = selectors.map((selector: any, selectorIndex: number) => (
      <React.Fragment key={`selector-${selectorIndex}`}>
        {selectorIndex > 0 && <Slash />}
        <div className="flex gap-[2px] items-center">{renderSelectorIcons(selector)}</div>
      </React.Fragment>
    ));

    return (
      <DiscountRow
        icons={selectorIcons}
        amount={amount}
        resourceTypes={affectedResources}
        isNegativeDiscount={isNegativeDiscount}
      />
    );
  }

  // Global discount (no selectors - applies to all cards)
  return (
    <DiscountRow
      icons={
        <span className="text-[10px] font-semibold text-white bg-[rgba(60,60,60,0.8)] px-1.5 py-0.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)]">
          All cards
        </span>
      }
      amount={amount}
      resourceTypes={affectedResources}
    />
  );
};

export default DiscountLayout;
