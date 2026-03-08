import React, { useState, useRef, useEffect, useCallback } from "react";
import { createPortal } from "react-dom";
import GameIcon from "../display/GameIcon.tsx";
import CardDecorBar from "../display/CardDecorBar.tsx";
import BehaviorSection from "./BehaviorSection";
import RequirementsBox from "./RequirementsBox.tsx";
import { FormattedDescription } from "../display/FormattedDescription.tsx";
import { getTagIconPath } from "@/utils/iconStore.ts";
import { CardDto, PlayerCardDto, ResourceTypeCredit } from "@/types/generated/api-types.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

interface GameCardProps {
  card: CardDto | PlayerCardDto;
  isSelected: boolean;
  onSelect: (cardId: string) => void;
  animationDelay?: number;
  showCheckbox?: boolean;
}

// Type guard to check if card is a PlayerCardDto (has state information)
function isPlayerCardDto(card: CardDto | PlayerCardDto): card is PlayerCardDto {
  return "available" in card && "effectiveCost" in card;
}

const CARD_CLIP_PATH = "polygon(0 0, calc(100% - 28px) 0, 100% 28px, 100% 100%, 0 100%)";

const CardDescriptionPortal: React.FC<{
  description: string;
  anchorRef: React.RefObject<HTMLDivElement | null>;
}> = ({ description, anchorRef }) => {
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);

  useEffect(() => {
    if (anchorRef.current) {
      const rect = anchorRef.current.getBoundingClientRect();
      setPos({ x: rect.left + rect.width / 2, y: rect.bottom });
    }
  }, [anchorRef]);

  if (!pos) {
    return null;
  }

  return createPortal(
    <div
      className="fixed w-[184px] max-md:w-[148px] -translate-x-1/2 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: pos.x, top: pos.y, zIndex: 99999 }}
    >
      <div
        className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
        style={{
          clipPath:
            "polygon(0 0, calc(100% - 14px) 0, 100% 14px, 100% 100%, 14px 100%, 0 calc(100% - 14px))",
        }}
      >
        <FormattedDescription text={description} />
        <svg
          className="absolute top-0 right-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
        <svg
          className="absolute bottom-0 left-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
      </div>
    </div>,
    document.body,
  );
};

const GameCard: React.FC<GameCardProps> = ({
  card,
  isSelected,
  onSelect,
  animationDelay = 0,
  showCheckbox = false,
}) => {
  const [imageError, setImageError] = useState(false);
  const [imageLoaded, setImageLoaded] = useState(false);
  const [cardHovered, setCardHovered] = useState(false);
  const descRef = useRef<HTMLDivElement>(null);
  const { playCardHoverSound } = useSoundEffects();
  const pendingSoundRef = useRef(false);

  useEffect(() => {
    if (pendingSoundRef.current) {
      pendingSoundRef.current = false;
      void playCardHoverSound();
    }
  }, [isSelected, playCardHoverSound]);

  const hasDescription = !!card.description;

  // Determine if card has state information (PlayerCardDto)
  const hasState = isPlayerCardDto(card);
  const isAvailable = hasState ? card.available : true;

  // Use effectiveCost from PlayerCardDto state calculation
  const displayCost = card.cost;
  const effectiveCost = hasState ? card.effectiveCost : card.cost;
  const actualDiscountAmount = displayCost - effectiveCost;

  const handleClick = useCallback(() => {
    pendingSoundRef.current = true;
    onSelect(card.id);
  }, [onSelect, card.id]);

  const handleImageLoad = () => {
    setImageLoaded(true);
  };

  const handleImageError = () => {
    setImageError(true);
  };

  const cardImagePath = `/assets/cards/${card.id}.webp`;

  // Card type accent colors (used for left stripe and selected checkbox)
  const accentColors = {
    automated: "#4caf50",
    active: "#2196f3",
    event: "#f44336",
    corporation: "#ffc107",
    prelude: "#e91e63",
  };

  const titleStyles = {
    automated:
      "bg-[linear-gradient(135deg,#0a1a0d_0%,#050f08_100%)] border border-[rgba(60,60,70,0.7)]",
    active:
      "bg-[linear-gradient(135deg,#0a1520_0%,#050a15_100%)] border border-[rgba(60,60,70,0.7)]",
    event:
      "bg-[linear-gradient(135deg,#1a0a0a_0%,#0f0505_100%)] border border-[rgba(60,60,70,0.7)]",
    corporation:
      "bg-[linear-gradient(135deg,#1a1508_0%,#0f0a04_100%)] border border-[rgba(60,60,70,0.7)]",
    prelude:
      "bg-[linear-gradient(135deg,#1a0a15_0%,#0f050a_100%)] border border-[rgba(60,60,70,0.7)]",
  };

  const cardBackgrounds = {
    automated: "bg-black",
    active: "bg-black",
    event: "bg-black",
    corporation: "bg-black",
    prelude: "bg-black",
  };

  // Card type specific checkbox colors (darker background, matching accent)
  const checkboxColors = {
    automated: { bg: "bg-[#1f3322]", border: `border-[${accentColors.automated}]` },
    active: { bg: "bg-[#152d4a]", border: `border-[${accentColors.active}]` },
    event: { bg: "bg-[#3a1f1f]", border: `border-[${accentColors.event}]` },
    corporation: { bg: "bg-[#3a2f0d]", border: `border-[${accentColors.corporation}]` },
    prelude: { bg: "bg-[#3a152c]", border: `border-[${accentColors.prelude}]` },
  };

  const hasTags = (card.tags && card.tags.length > 0) || card.type === "event";
  const cardType = card.type as keyof typeof accentColors;
  const cardBg =
    cardType && cardBackgrounds[cardType] ? cardBackgrounds[cardType] : "bg-[rgba(0,0,0,0.9)]";
  const checkboxColor =
    cardType && checkboxColors[cardType]
      ? checkboxColors[cardType]
      : { bg: "bg-[#4a90e2]", border: "border-[#4a90e2]" };

  return (
    <div
      className={`relative w-[200px] min-h-[280px] p-4 transition-all duration-200 z-[1] max-md:w-[160px] max-md:min-h-[240px] max-md:p-3 group select-none ${animationDelay >= 0 ? "opacity-0 translate-y-5 animate-[fadeInUp_0.5s_ease_forwards]" : ""} ${!isAvailable ? "grayscale-[0.6] brightness-[0.65] saturate-[0.2]" : ""}`}
      style={animationDelay >= 0 ? { animationDelay: `${animationDelay}ms` } : undefined}
      onClick={handleClick}
      onMouseEnter={() => setCardHovered(true)}
      onMouseLeave={() => setCardHovered(false)}
    >
      {/* Inner card body with clip-path for angled top-right corner */}
      <div
        className={`absolute inset-0 ${cardBg} shadow-[0_4px_12px_rgba(0,0,0,0.3)]`}
        style={{ clipPath: CARD_CLIP_PATH }}
      >
        {/* Card border - neutral, matching BottomResourceBar */}
        <div
          className="absolute inset-0 border border-[rgba(60,60,70,0.7)] pointer-events-none transition-colors duration-200 group-hover:border-[rgba(120,120,140,0.8)]"
          style={{ clipPath: CARD_CLIP_PATH }}
        ></div>
        {/* Diagonal border line at the angled corner */}
        <svg
          className="absolute top-0 right-0 w-[28px] h-[28px] pointer-events-none transition-colors duration-200"
          viewBox="0 0 28 28"
        >
          <line
            x1="0"
            y1="0"
            x2="28"
            y2="28"
            className="stroke-[rgba(60,60,70,0.7)] group-hover:stroke-[rgba(120,120,140,0.8)] transition-all duration-200"
            strokeWidth="2"
          />
        </svg>
      </div>

      {/* Cost in top-left (hidden for prelude cards) */}
      {card.type !== "prelude" && (
        <div
          className={`absolute top-2 left-2 flex items-center justify-start z-[2] shrink-0 group/cost ${actualDiscountAmount > 0 ? "animate-[goldenPulse_2.5s_ease-in-out_infinite]" : ""}`}
        >
          <div className={actualDiscountAmount > 0 ? "group-hover/cost:hidden" : ""}>
            <GameIcon iconType={ResourceTypeCredit} amount={effectiveCost} size="medium" />
          </div>
          {actualDiscountAmount > 0 && (
            <div className="hidden group-hover/cost:block [filter:none]">
              <GameIcon iconType={ResourceTypeCredit} amount={displayCost} size="medium" />
            </div>
          )}
        </div>
      )}

      {/* Card type accent stripe on left side */}
      {cardType && accentColors[cardType] && (
        <div
          className="absolute -left-[5px] top-[2.5%] bottom-[2.5%] w-[5px] z-[0] transition-all duration-300"
          style={{
            filter: isSelected
              ? `drop-shadow(0 0 6px ${accentColors[cardType]}) drop-shadow(0 0 12px ${accentColors[cardType]}80)`
              : "none",
          }}
        >
          <div
            className="w-full h-full"
            style={{
              backgroundColor: accentColors[cardType],
              clipPath: "polygon(0 4px, 100% 0, 100% 100%, 0 calc(100% - 4px))",
            }}
          />
        </div>
      )}

      {/* Requirements box */}
      <RequirementsBox requirements={card.requirements} />

      {/* Tags as vertical stack on right side */}
      {hasTags && (
        <div className="absolute bottom-[58%] right-0 flex flex-col gap-1 items-center z-[5] pointer-events-auto max-md:right-0.5">
          {card.tags &&
            card.tags.slice(0, card.type === "event" ? 2 : 3).map((tag, index) => {
              const tagIcon = getTagIconPath(tag.toLowerCase());
              if (!tagIcon) return null;
              return (
                <div
                  key={index}
                  className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
                >
                  <img
                    src={tagIcon}
                    alt={tag}
                    className="w-6 h-6 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
                  />
                </div>
              );
            })}
          {card.type === "event" && (
            <div className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]">
              <GameIcon iconType="event" size="small" />
            </div>
          )}
        </div>
      )}

      {/* Image area */}
      <div className="absolute top-5 left-4 right-4 h-[35%] bg-white/5 rounded border border-dashed border-white/20 z-[1] overflow-hidden max-md:top-4 max-md:left-3 max-md:right-3">
        {!imageError && (
          <img
            src={cardImagePath}
            alt={card.name}
            className={`w-full h-full object-cover rounded border border-[rgba(60,60,70,0.7)] opacity-0 transition-opacity duration-300 ${imageLoaded ? "opacity-100" : ""}`}
            onLoad={handleImageLoad}
            onError={handleImageError}
          />
        )}
        {imageError && (
          <div className="w-full h-full bg-white/5 rounded border border-dashed border-white/20"></div>
        )}
      </div>

      {/* Card title at 38% from top */}
      <div className={"absolute top-[38%] left-0 right-2 z-[4] max-md:px-0.5"}>
        <div className="relative w-full">
          <h3
            className={`${card.name.length > 19 ? "text-[11px]" : card.name.length > 14 ? "text-[13px]" : "text-base"} font-orbitron font-semibold text-white leading-[1.2] text-left flex items-center justify-start w-full h-[44px] rounded-none p-1 pl-3 ${hasTags ? "pr-[30px]" : "pr-3"} shadow-[0_3px_6px_rgba(0,0,0,0.4)] my-0 mx-auto ${card.name.length > 19 ? "max-md:text-[9px]" : card.name.length > 14 ? "max-md:text-[11px]" : "max-md:text-sm"} max-md:h-[36px] max-md:pl-2 ${hasTags ? "max-md:pr-[25px]" : "max-md:pr-2"} ${cardType && titleStyles[cardType] ? titleStyles[cardType] : ""}`}
            style={{
              clipPath:
                "polygon(0 0, 100% 0, 100% calc(100% - 12px), calc(100% - 12px) 100%, 0 100%)",
            }}
          >
            {card.name}
          </h3>
          <svg
            className="absolute bottom-0 right-0 w-[12px] h-[12px] pointer-events-none"
            viewBox="0 0 12 12"
          >
            <line x1="12" y1="0" x2="0" y2="12" stroke="rgba(60,60,70,0.7)" strokeWidth="2" />
          </svg>
        </div>
        {/* Victory Points + Resource Storage label below title */}
        <div className="relative pointer-events-auto">
          <CardDecorBar vpConditions={card.vpConditions} resourceStorage={card.resourceStorage} />
        </div>
      </div>

      {/* Content section - takes up roughly half the card height and vertically centers content */}
      <div
        className={`absolute left-2 right-2 flex items-center justify-center z-[3] max-md:left-1.5 max-md:right-1.5 ${hasDescription ? "bottom-10 max-md:bottom-9" : "bottom-4 max-md:bottom-3"} ${(card.vpConditions && card.vpConditions.length > 0) || card.resourceStorage ? "top-[calc(50%+38px)] max-md:top-[calc(50%+40px)]" : "top-[calc(50%+20px)] max-md:top-[calc(50%+25px)]"}`}
      >
        <BehaviorSection behaviors={card.behaviors} />
      </div>

      {/* Card description at the bottom */}
      {hasDescription && (
        <div
          ref={descRef}
          className="absolute bottom-1.5 left-2 right-2 z-[3] text-[9px] leading-tight text-white/50 line-clamp-2 max-md:left-1.5 max-md:right-1.5 max-md:bottom-1"
        >
          <FormattedDescription text={card.description} />
          {cardHovered && (
            <CardDescriptionPortal description={card.description} anchorRef={descRef} />
          )}
        </div>
      )}

      {/* Selection indicator at bottom center, peeking out (only shown when showCheckbox is true) */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2] max-md:-bottom-2.5">
          <div
            className={`w-6 h-6 rounded-full bg-[#2a3142] border-2 border-[rgba(100,150,200,0.3)] flex items-center justify-center transition-all duration-300 ${isSelected ? `${checkboxColor.bg} ${checkboxColor.border}` : ""}`}
          >
            {isSelected && <span className="text-white text-sm font-bold">✓</span>}
          </div>
        </div>
      )}
    </div>
  );
};

export default GameCard;
