import { useState } from "react";
import { CardDto, ColonyDto, PendingEffectSelectionDto } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameButton from "../buttons/GameButton.tsx";
import BehaviorSection from "../cards/BehaviorSection";
import { OVERLAY_ROOT_SAFE_AREA_CLASS } from "./overlayStyles.ts";

interface Props {
  selection: PendingEffectSelectionDto;
  cards: CardDto[];
  colonies: ColonyDto[];
  players: { id: string; name: string }[];
  onConfirm: (optionIndex: number) => Promise<unknown>;
}

export default function EffectSelectionOverlay({
  selection,
  cards,
  colonies,
  players,
  onConfirm,
}: Props) {
  const [selected, setSelected] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  return (
    <div
      className={`fixed inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm ${OVERLAY_ROOT_SAFE_AREA_CLASS}`}
      style={{ zIndex: Z_INDEX.SELECTION_POPOVER }}
    >
      <div
        className="game-panel game-panel-clipped game-window w-[540px] max-w-[95vw] max-h-[80dvh] compact:max-h-[calc(100%-16px)] flex flex-col overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-label={selection.source}
      >
        <h2 className="font-orbitron text-white text-base px-4 py-3 border-b border-white/10 compact:shrink-0 compact:py-2.5 compact:m-0">
          {selection.source}
        </h2>
        <div className="overflow-y-auto flex-1 p-3 space-y-2">
          {selection.options.map((option, index) => (
            <GameButton
              key={index}
              emphasis="quiet"
              selected={selected === index}
              aria-pressed={selected === index}
              className="w-full p-3 text-left"
              disabled={submitting}
              onClick={() => setSelected(index)}
            >
              <div className="flex flex-col gap-2 w-full">
                {option.cardId && (
                  <span>
                    {cards.find((card) => card.id === option.cardId)?.name ?? option.cardId}
                  </span>
                )}
                {option.targetPlayerId && (
                  <span>{players.find((player) => player.id === option.targetPlayerId)?.name}</span>
                )}
                {option.colonyIds?.map((id, position) => (
                  <span key={`${position}-${id}`} className="flex justify-between gap-3">
                    <span>{colonies.find((colony) => colony.id === id)?.name ?? id}</span>
                    <BehaviorSection
                      behaviors={[{ outputs: [selection.outputs[position]] }]}
                      noContainer
                      showTooltips={false}
                    />
                  </span>
                ))}
                {option.outputs && (
                  <BehaviorSection
                    behaviors={[{ outputs: option.outputs }]}
                    noContainer
                    showTooltips={false}
                  />
                )}
              </div>
            </GameButton>
          ))}
        </div>
        {error && (
          <p className="px-4 text-red-300 text-sm" role="alert">
            {error}
          </p>
        )}
        <div className="flex justify-end p-3 border-t border-white/10 compact:shrink-0 compact:py-1.5">
          <GameButton
            className="compact:min-h-11 compact:px-6"
            disabled={selected === null || submitting}
            onClick={() => {
              if (selected === null) {
                return;
              }
              setSubmitting(true);
              setError(null);
              void onConfirm(selected)
                .catch((reason: unknown) => {
                  setError(
                    reason instanceof Error ? reason.message : "Could not confirm selection",
                  );
                  setSubmitting(false);
                })
                .finally(() => setSubmitting(false));
            }}
          >
            Confirm
          </GameButton>
        </div>
      </div>
    </div>
  );
}
