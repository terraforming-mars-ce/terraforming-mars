import React from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import GameCard from "../cards/GameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";

interface CardPreviewPanelProps {
  card: CardDto | null;
}

const noop = () => {};

const CardPreviewPanel: React.FC<CardPreviewPanelProps> = ({ card }) => {
  const isCorporation = card?.type === "corporation";

  return (
    <div
      className={`
        hidden md:block
        absolute left-full top-0 ml-4
        transition-all duration-200 ease-out
        ${card ? "opacity-100 translate-x-0" : "opacity-0 translate-x-2 pointer-events-none"}
      `}
    >
      {card &&
        (isCorporation ? (
          <div
            style={{ transform: "scale(0.55)", transformOrigin: "top left", width: 400, height: 0 }}
          >
            <CorporationCard
              card={card}
              isSelected={false}
              onSelect={noop}
              showCheckbox={false}
              disableInteraction={true}
            />
          </div>
        ) : (
          <GameCard card={card} isSelected={false} onSelect={noop} showCheckbox={false} />
        ))}
    </div>
  );
};

export default CardPreviewPanel;
