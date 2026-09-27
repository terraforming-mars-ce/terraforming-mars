import { cardImage } from "@/assets";
import { useRef, useState, type ReactNode } from "react";
import GameIcon from "../display/GameIcon.tsx";
import CardDecorBar from "../display/CardDecorBar.tsx";
import BehaviorSection from "./BehaviorSection";
import RequirementsBox from "./RequirementsBox.tsx";
import { getTagIconPath } from "@/utils/iconStore.ts";
import {
  CardDto,
  PlayerCardDto,
  ResourceTypeCredit,
  TagWild,
} from "@/types/generated/api-types.ts";
import DecorBoxTooltip from "../display/DecorBoxTooltip.tsx";
import { FormattedDescription } from "../display/FormattedDescription.tsx";
import { CARD_TYPE_COLORS } from "@/utils/cardTypeColors.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

export interface GameCardProps {
  card: CardDto | PlayerCardDto;
  isSelected?: boolean;
  showCheckbox?: boolean;
  presentation?: "compact" | "inspection";
  description?: ReactNode;
}

const TagIcon: React.FC<{ tagIcon: string; tag: string; isWild: boolean }> = ({
  tagIcon,
  tag,
  isWild,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);

  const handleMouseEnter = () => {
    if (ref.current) {
      const rect = ref.current.getBoundingClientRect();
      setTooltipPos({ x: rect.left + rect.width / 2, y: rect.top });
    }
  };

  return (
    <div
      ref={isWild ? ref : undefined}
      className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
      onMouseEnter={isWild ? handleMouseEnter : undefined}
      onMouseLeave={isWild ? () => setTooltipPos(null) : undefined}
    >
      <img
        src={tagIcon}
        alt={tag}
        className="w-8 h-8 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
      />
      {isWild && (
        <DecorBoxTooltip
          description="A wild tag counts as any tag"
          position={tooltipPos}
          placement="above"
          cornerSize={10}
        />
      )}
    </div>
  );
};

export default function GameCard({
  card,
  isSelected = false,
  showCheckbox = false,
  presentation = "compact",
  description,
}: GameCardProps) {
  const hasState = "available" in card && "effectiveCost" in card;
  const effectiveCost = hasState ? card.effectiveCost : card.cost;
  const discounted = effectiveCost < card.cost;
  const accent = CARD_TYPE_COLORS[card.type as keyof typeof CARD_TYPE_COLORS] ?? "#4a90e2";
  const hasTags = !!card.tags?.length || card.type === "event";
  const artwork = cardImage(card.id);

  return (
    <div className="game-card" data-presentation={presentation} data-card-name={card.name}>
      <div className="game-card-requirements">
        <RequirementsBox requirements={card.requirements} inFlow />
      </div>
      <div
        className={`game-card-body group ${presentation === "compact" && hasState && !card.available ? "grayscale-[0.6] brightness-[0.65] saturate-[0.2]" : ""}`}
        style={{ "--card-accent": accent } as React.CSSProperties}
      >
        <div className="game-card-frame" style={{ zIndex: Z_INDEX.CONTROL_DECORATION }} />
        <div
          className="game-card-stripe"
          style={{
            filter: isSelected
              ? `drop-shadow(0 0 6px ${accent}) drop-shadow(0 0 12px ${accent}80)`
              : undefined,
          }}
        />
        <div className="game-card-heading">
          <div className="game-card-artwork">
            <CardArtwork key={card.id} artwork={artwork} name={card.name} />
          </div>
          {card.type !== "prelude" && (
            <div
              className="game-card-cost absolute top-2 left-2 group/cost"
              style={{ zIndex: Z_INDEX.GAME_BOARD_EFFECTS }}
            >
              <div
                className={
                  discounted
                    ? "animate-[goldenPulse_2.5s_ease-in-out_infinite] transition-opacity duration-400 group-hover/cost:opacity-0"
                    : ""
                }
              >
                <GameIcon iconType={ResourceTypeCredit} amount={effectiveCost} size="medium" />
              </div>
              {discounted && (
                <div className="absolute inset-0 opacity-0 transition-opacity duration-400 group-hover/cost:opacity-100">
                  <GameIcon iconType={ResourceTypeCredit} amount={card.cost} size="medium" />
                </div>
              )}
            </div>
          )}
          <h3 className="game-card-title font-orbitron" data-card-type={card.type}>
            <span>{card.name}</span>
          </h3>
          {hasTags && (
            <div className="game-card-tags" style={{ zIndex: Z_INDEX.GAME_BOARD_EFFECTS }}>
              {card.tags?.slice(0, card.type === "event" ? 2 : 3).map((tag, index) => {
                const tagIcon = getTagIconPath(tag.toLowerCase());
                if (!tagIcon) {
                  return null;
                }
                return (
                  <TagIcon
                    key={`${tag}-${index}`}
                    tagIcon={tagIcon}
                    tag={tag}
                    isWild={tag.toLowerCase() === TagWild}
                  />
                );
              })}
              {card.type === "event" && <GameIcon iconType="event" size="medium" />}
            </div>
          )}
        </div>
        <div className="game-card-labels">
          <CardDecorBar vpConditions={card.vpConditions} resourceStorage={card.resourceStorage} />
        </div>
        <div className="game-card-behaviors">
          <BehaviorSection
            behaviors={card.behaviors}
            computedValues={hasState ? card.computedValues : undefined}
            presentation={presentation}
          />
        </div>
        {presentation === "inspection" && card.description && (
          <div className="game-card-description">
            {description ?? (
              <div>
                <FormattedDescription text={card.description} />
              </div>
            )}
          </div>
        )}
        {showCheckbox && (
          <div
            className="absolute -bottom-3 left-1/2 -translate-x-1/2"
            style={{ zIndex: Z_INDEX.GAME_BOARD_EFFECTS }}
          >
            <div
              className="w-6 h-6 rounded-full border-2 flex items-center justify-center transition-colors duration-200"
              style={{
                backgroundColor: isSelected
                  ? `color-mix(in srgb, ${accent} 25%, #15202f)`
                  : "#2a3142",
                borderColor: isSelected ? accent : "rgba(100,150,200,0.3)",
              }}
            >
              {isSelected && <span className="text-white text-sm font-bold">✓</span>}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function CardArtwork({ artwork, name }: { artwork: ReturnType<typeof cardImage>; name: string }) {
  const [failed, setFailed] = useState(false);
  const [loaded, setLoaded] = useState(false);
  if (!artwork || failed) {
    return null;
  }
  return (
    <img
      {...artwork}
      decoding="async"
      alt={name}
      className={`w-full h-full object-cover transition-opacity duration-300 ${loaded ? "opacity-100" : "opacity-0"}`}
      onLoad={() => setLoaded(true)}
      onError={() => setFailed(true)}
    />
  );
}
