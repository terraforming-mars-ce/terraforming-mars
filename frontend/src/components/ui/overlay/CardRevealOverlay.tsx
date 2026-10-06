import { useState } from "react";
import type { PendingCardRevealDto } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameCard from "../cards/GameCard.tsx";
import GameButton from "../buttons/GameButton.tsx";
import GameIcon from "../display/GameIcon.tsx";

interface Props {
  reveal: PendingCardRevealDto;
  onConfirm: () => Promise<unknown>;
}

export default function CardRevealOverlay({ reveal, onConfirm }: Props) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  return (
    <div
      className="fixed inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={`${reveal.source}: revealed cards`}
        className="game-panel game-panel-clipped game-window max-w-[95vw] max-h-[90vh] flex flex-col overflow-hidden"
      >
        <h2 className="font-orbitron text-white text-base px-5 py-3 border-b border-white/10">
          {reveal.source}
        </h2>
        <div className="overflow-auto p-5">
          <div className="flex gap-5 w-max mx-auto">
            {reveal.cards.map((card) => (
              <div key={card.id} className="w-[min(var(--card-inspection-width),80vw)] shrink-0">
                <GameCard card={card} presentation="inspection" />
                <p className="mt-2 text-center text-sm text-white/80">
                  {reveal.results.find((result) => result.cardId === card.id)?.matched
                    ? "Match"
                    : "No match"}
                </p>
              </div>
            ))}
          </div>
          <p className="text-white text-sm mt-4">Revealed and discarded.</p>
          {reveal.rewards.length > 0 ? (
            <div
              className="flex items-center justify-center gap-3 mt-3 text-white"
              aria-label="Reward"
            >
              {reveal.rewards.map((reward, i) => (
                <GameIcon
                  key={i}
                  iconType={reward.resourceType}
                  amount={reward.amount}
                  size="small"
                />
              ))}
            </div>
          ) : (
            <p className="text-white/70 text-sm mt-2">No reward.</p>
          )}
        </div>
        {error && (
          <p className="px-5 text-red-300 text-sm" role="alert">
            {error}
          </p>
        )}
        <div className="flex justify-end p-3 border-t border-white/10">
          <GameButton
            disabled={submitting}
            onClick={() => {
              setSubmitting(true);
              setError(null);
              void onConfirm()
                .catch((reason: unknown) => {
                  setError(reason instanceof Error ? reason.message : "Could not confirm reveal");
                })
                .finally(() => setSubmitting(false));
            }}
          >
            Continue
          </GameButton>
        </div>
      </div>
    </div>
  );
}
