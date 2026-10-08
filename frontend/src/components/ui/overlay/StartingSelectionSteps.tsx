import { useRef, useState, type ReactNode } from "react";
import { CardDto, ResourceTypeCredit } from "@/types/generated/api-types.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import { useLongPress } from "@/hooks/useLongPress.ts";
import CorporationCard from "../cards/CorporationCard.tsx";
import CardChoice from "../cards/CardChoice.tsx";
import FittedCard from "../cards/FittedCard.tsx";
import CardTagList from "../cards/CardTagList.tsx";
import GameIcon from "../display/GameIcon.tsx";
import GameButton from "../buttons/GameButton.tsx";
import { CurrentGameMenuButton } from "../buttons/MainMenuHamburger.tsx";
import CardDetailPortal from "./CardDetailPortal.tsx";
import {
  OVERLAY_ACTION_BUTTON_CLASS,
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_CONTAINER_STYLE,
  OVERLAY_FOOTER_CLASS,
  OVERLAY_FOOTER_LEFT_CLASS,
  OVERLAY_FOOTER_RIGHT_CLASS,
  OVERLAY_TITLE_CLASS,
  RESOURCE_DISPLAY_CLASS,
  RESOURCE_LABEL_CLASS,
} from "./overlayStyles.ts";
import CardSelectionRow from "./CardSelectionRow.tsx";

type StepId = "corporation" | "preludes" | "cards" | "review";

const STEP_LABELS: Record<StepId, string> = {
  corporation: "Corp",
  preludes: "Prelude",
  cards: "Cards",
  review: "Review",
};

const CORPORATION_NATURAL_WIDTH = 320;

interface StartingSelectionStepsProps {
  availableCorporations: CardDto[];
  availablePreludes: CardDto[];
  maxSelectablePreludes: number;
  cards: CardDto[];
  selectedCorporation: CardDto | undefined;
  onSelectCorporation: (corporationId: string) => void;
  selectedPreludeIds: string[];
  onSelectPrelude: (cardId: string) => void;
  selectedCardIds: string[];
  onSelectCard: (cardId: string) => void;
  credits: number;
  totalCost: number;
  preludesValid: boolean;
  cardsValid: boolean;
  allValid: boolean;
  showConfirmation: boolean;
  onConfirm: () => void;
  onHide?: () => void;
  stepIndex: number;
  maxVisitedIndex: number;
  onStepIndexChange: (index: number) => void;
}

const JUSTIFY_CLASS = {
  start: "justify-start",
  center: "justify-center",
  end: "justify-end",
} as const;

function corporationJustify(index: number, count: number): "start" | "center" | "end" {
  if (count < 2) {
    return "center";
  }
  if (index === 0) {
    return "end";
  }
  if (index === count - 1) {
    return "start";
  }
  return "center";
}

function FittedCorporation({
  corporation,
  isSelected,
  onSelect,
  onInspect,
  justify,
}: {
  corporation: CardDto;
  isSelected: boolean;
  onSelect: (corporationId: string) => void;
  onInspect: () => void;
  justify: "start" | "center" | "end";
}) {
  const boxRef = useRef<HTMLDivElement>(null);
  const longPress = useLongPress<HTMLDivElement>(onInspect);

  return (
    <div
      ref={boxRef}
      className={`flex-1 min-w-0 h-full flex items-center overflow-hidden ${JUSTIFY_CLASS[justify]}`}
    >
      <div {...longPress}>
        <FittedCard naturalWidth={CORPORATION_NATURAL_WIDTH} boundsRef={boxRef}>
          <CorporationCard
            card={corporation}
            isSelected={isSelected}
            onSelect={onSelect}
            borderColor={getCorporationBorderColor(corporation.name)}
            catalog
          />
        </FittedCard>
      </div>
    </div>
  );
}

function ReviewItem({ card, onInspect }: { card: CardDto; onInspect: (card: CardDto) => void }) {
  return (
    <li>
      <button
        type="button"
        className="w-full min-h-11 flex items-center gap-2 px-2 text-left text-[13px] text-white/90 border-b border-white/10 cursor-pointer"
        onClick={() => onInspect(card)}
      >
        <span className="flex-1 min-w-0 truncate">{card.name}</span>
        <CardTagList card={card} size="sm" className="flex items-center gap-1 shrink-0" />
      </button>
    </li>
  );
}

function ReviewSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="min-w-0 flex flex-col">
      <h3 className="m-0 mb-1 font-orbitron text-[11px] font-bold uppercase tracking-wider text-white/50">
        {title}
      </h3>
      {children}
    </section>
  );
}

export default function StartingSelectionSteps({
  availableCorporations,
  availablePreludes,
  maxSelectablePreludes,
  cards,
  selectedCorporation,
  onSelectCorporation,
  selectedPreludeIds,
  onSelectPrelude,
  selectedCardIds,
  onSelectCard,
  credits,
  totalCost,
  preludesValid,
  cardsValid,
  allValid,
  showConfirmation,
  onConfirm,
  onHide,
  stepIndex,
  maxVisitedIndex,
  onStepIndexChange,
}: StartingSelectionStepsProps) {
  const steps: StepId[] = [];
  if (availableCorporations.length > 0) {
    steps.push("corporation");
  }
  if (availablePreludes.length > 0) {
    steps.push("preludes");
  }
  steps.push("cards", "review");

  const [inspected, setInspected] = useState<CardDto | null>(null);
  const currentIndex = Math.min(stepIndex, steps.length - 1);
  const step = steps[currentIndex];
  const remainingCredits = credits - totalCost;

  const stepValid: Record<StepId, boolean> = {
    corporation: !!selectedCorporation,
    preludes: preludesValid,
    cards: cardsValid,
    review: allValid,
  };

  const selectedPreludes = availablePreludes.filter((card) => selectedPreludeIds.includes(card.id));
  const selectedCards = cards.filter((card) => selectedCardIds.includes(card.id));

  let confirmLabel = "Confirm Selection";
  if (showConfirmation) {
    confirmLabel = "Confirm Skip";
  }

  const creditsSummary = (
    <>
      <div className={RESOURCE_DISPLAY_CLASS}>
        <span className={RESOURCE_LABEL_CLASS}>Cost</span>
        <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="medium" />
      </div>
      <div className={RESOURCE_DISPLAY_CLASS}>
        <span className={RESOURCE_LABEL_CLASS}>Left</span>
        <GameIcon iconType={ResourceTypeCredit} amount={remainingCredits} size="medium" />
      </div>
    </>
  );

  let footerStatus: ReactNode = null;
  if (step === "corporation") {
    footerStatus = (
      <span className="truncate text-[13px] text-white/70">
        {selectedCorporation ? selectedCorporation.name : "Select a corporation to continue"}
      </span>
    );
  } else if (step === "preludes") {
    footerStatus = (
      <span className="font-orbitron text-[13px] text-white/70">
        {selectedPreludeIds.length} / {maxSelectablePreludes} selected
      </span>
    );
  } else if (step === "cards") {
    footerStatus = (
      <>
        {creditsSummary}
        <span className="font-orbitron text-[12px] text-white/60 whitespace-nowrap">
          {selectedCardIds.length} {selectedCardIds.length === 1 ? "card" : "cards"}
        </span>
      </>
    );
  } else if (showConfirmation) {
    footerStatus = (
      <span className="text-[13px] text-[#ff9800] leading-tight">
        No project cards selected. Confirm to skip buying cards.
      </span>
    );
  } else {
    footerStatus = creditsSummary;
  }

  return (
    <div className={OVERLAY_CONTAINER_CLASS} style={OVERLAY_CONTAINER_STYLE}>
      <div className="shrink-0 min-h-11 py-1 pl-1 pr-3 grid grid-cols-[1fr_auto_1fr] items-center gap-3 bg-black/40 border-b border-white/15">
        <div className="col-start-1 justify-self-start">
          <CurrentGameMenuButton />
        </div>
        <div className="col-start-2 min-w-0 flex flex-col items-center justify-center text-center">
          <h2 className={OVERLAY_TITLE_CLASS}>Select Starting Cards</h2>
        </div>
        <ol
          className="col-start-3 justify-self-end m-0 p-0 list-none flex items-center gap-1"
          aria-label="Steps"
        >
          {steps.map((id, index) => {
            const reached = index <= maxVisitedIndex;
            let stateClass = "border-white/15 text-white/40";
            if (index === currentIndex) {
              stateClass = "border-white/70 bg-white/10 text-white";
            } else if (reached) {
              stateClass = "border-white/30 text-white/70 hover:border-white/50 hover:text-white";
            }
            return (
              <li key={id}>
                <button
                  type="button"
                  aria-current={index === currentIndex ? "step" : undefined}
                  disabled={!reached}
                  className={`min-h-11 flex items-center ${reached ? "cursor-pointer" : "cursor-default"}`}
                  onClick={() => onStepIndexChange(index)}
                >
                  <span
                    className={`font-orbitron text-[11px] font-bold uppercase tracking-wide whitespace-nowrap px-2 py-1 border transition-colors ${stateClass}`}
                  >
                    {index + 1} {STEP_LABELS[id]}
                  </span>
                </button>
              </li>
            );
          })}
        </ol>
      </div>

      {step === "corporation" && (
        <div className="flex-1 min-h-0 flex gap-8 px-4 py-3">
          {availableCorporations.map((corporation, index) => (
            <FittedCorporation
              key={corporation.id}
              corporation={corporation}
              isSelected={selectedCorporation?.id === corporation.id}
              onSelect={onSelectCorporation}
              onInspect={() => setInspected(corporation)}
              justify={corporationJustify(index, availableCorporations.length)}
            />
          ))}
        </div>
      )}

      {step === "preludes" && (
        <CardSelectionRow>
          {availablePreludes.map((card, index) => (
            <CardChoice
              key={card.id}
              card={card}
              isSelected={selectedPreludeIds.includes(card.id)}
              disabled={
                selectedPreludeIds.length >= maxSelectablePreludes &&
                !selectedPreludeIds.includes(card.id)
              }
              onSelect={onSelectPrelude}
              animationDelay={index * 100}
              showCheckbox={true}
            />
          ))}
        </CardSelectionRow>
      )}

      {step === "cards" && (
        <CardSelectionRow>
          {cards.map((card, index) => (
            <CardChoice
              key={card.id}
              card={card}
              isSelected={selectedCardIds.includes(card.id)}
              onSelect={onSelectCard}
              animationDelay={index * 60}
              showCheckbox={true}
            />
          ))}
        </CardSelectionRow>
      )}

      {step === "review" && (
        <div
          className={`flex-1 min-h-0 overflow-y-auto overscroll-contain px-4 py-3 grid gap-4 ${selectedPreludes.length > 0 ? "grid-cols-3" : "grid-cols-2"}`}
        >
          <ReviewSection title="Corporation">
            {selectedCorporation ? (
              <button
                type="button"
                className="w-full min-w-0 min-h-11 flex items-center justify-center px-2 py-1 cursor-pointer"
                onClick={() => setInspected(selectedCorporation)}
                aria-label={selectedCorporation.name}
              >
                {getCorporationLogo(selectedCorporation.name, "max-w-full h-12 object-contain") ?? (
                  <span className="max-w-full truncate text-[13px] text-white/90">
                    {selectedCorporation.name}
                  </span>
                )}
              </button>
            ) : (
              <span className="text-[13px] text-white/50">None</span>
            )}
          </ReviewSection>
          {selectedPreludes.length > 0 && (
            <ReviewSection title="Preludes">
              <ul className="m-0 p-0 list-none">
                {selectedPreludes.map((card) => (
                  <ReviewItem key={card.id} card={card} onInspect={setInspected} />
                ))}
              </ul>
            </ReviewSection>
          )}
          <ReviewSection title="Project cards">
            {selectedCards.length > 0 ? (
              <ul className="m-0 p-0 list-none">
                {selectedCards.map((card) => (
                  <ReviewItem key={card.id} card={card} onInspect={setInspected} />
                ))}
              </ul>
            ) : (
              <span className="text-[13px] text-white/50">No cards</span>
            )}
          </ReviewSection>
        </div>
      )}

      <div className={OVERLAY_FOOTER_CLASS}>
        <div className={`${OVERLAY_FOOTER_LEFT_CLASS} flex-1`}>{footerStatus}</div>
        <div className={OVERLAY_FOOTER_RIGHT_CLASS}>
          {onHide && (
            <GameButton emphasis="quiet" className={OVERLAY_ACTION_BUTTON_CLASS} onClick={onHide}>
              Hide
            </GameButton>
          )}
          {currentIndex > 0 && (
            <GameButton
              emphasis="secondary"
              className={OVERLAY_ACTION_BUTTON_CLASS}
              onClick={() => onStepIndexChange(currentIndex - 1)}
            >
              Back
            </GameButton>
          )}
          {step === "review" ? (
            <GameButton
              className={OVERLAY_ACTION_BUTTON_CLASS}
              disabled={!allValid}
              onClick={onConfirm}
            >
              {confirmLabel}
            </GameButton>
          ) : (
            <GameButton
              className={OVERLAY_ACTION_BUTTON_CLASS}
              disabled={!stepValid[step]}
              onClick={() => onStepIndexChange(currentIndex + 1)}
            >
              Next
            </GameButton>
          )}
        </div>
      </div>

      {inspected && <CardDetailPortal card={inspected} onClose={() => setInspected(null)} />}
    </div>
  );
}
