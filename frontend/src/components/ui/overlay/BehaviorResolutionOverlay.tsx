import { useEffect, useRef, useState } from "react";
import type {
  CardDto,
  PendingBehaviorResolutionDto,
  PlayerCardDto,
} from "@/types/generated/api-types";
import { globalWebSocketManager } from "@/services/globalWebSocketManager";
import { webSocketService } from "@/services/webSocketService";
import BehaviorSection from "../cards/BehaviorSection";
import CardChoice from "../cards/CardChoice";
import GameButton from "../buttons/GameButton";
import {
  GameFlowPopover,
  GameFlowTitle,
  GameFlowBody,
  GameFlowFooter,
} from "../popover/GameFlowPopover";

interface Props {
  resolutions: PendingBehaviorResolutionDto[];
  handCards: PlayerCardDto[];
  playedCards: CardDto[];
  corporation?: CardDto;
}

export default function BehaviorResolutionOverlay({
  resolutions,
  handCards,
  playedCards,
  corporation,
}: Props) {
  const [activeId, setActiveId] = useState(resolutions[0]?.id);
  const active = resolutions.find((item) => item.id === activeId) ?? resolutions[0];
  const [selectedCards, setSelectedCards] = useState<string[]>([]);
  const [choiceIndex, setChoiceIndex] = useState<number | null>(null);
  const [targets, setTargets] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const submission = useRef<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setActiveId(active?.id);
    setSelectedCards([]);
    setChoiceIndex(null);
    setTargets([]);
    setSubmitting(false);
    submission.current = null;
    setError(null);
  }, [active?.id]);

  useEffect(() => {
    const failed = (payload: unknown) => {
      const result = payload as { resolutionId?: string; error?: string };
      if (submission.current && result.resolutionId === submission.current) {
        setError(result.error ?? "Could not resolve effect");
        submission.current = null;
        setSubmitting(false);
      }
    };
    const disconnected = () => {
      if (submission.current) {
        setError("Connection lost. Reconnect to check whether this effect was resolved.");
        submission.current = null;
        setSubmitting(false);
      }
    };
    webSocketService.on("error", failed);
    webSocketService.on("disconnect", disconnected);
    return () => {
      webSocketService.off("error", failed);
      webSocketService.off("disconnect", disconnected);
    };
  }, []);

  if (!active) {
    return null;
  }

  const choice = active.choices?.find((item) => item.originalIndex === choiceIndex);
  const storageTargets = choice?.storageTargets ?? [];
  const validCards = selectedCards.filter((id) => handCards.some((card) => card.id === id));
  const allCards = corporation ? [...playedCards, corporation] : playedCards;
  const canConfirmChoice =
    !!choice?.available &&
    storageTargets.every((eligible, index) => eligible.includes(targets[index]));
  const canConfirmDiscard =
    validCards.length >= active.minCards && validCards.length <= active.maxCards;
  const groups = new Map<string, PendingBehaviorResolutionDto[]>();
  for (const resolution of resolutions) {
    const key = `${resolution.sourceCardId}:${resolution.sourceBehaviorIndex}:${resolution.triggeringCardId ?? ""}:${resolution.kind}`;
    const entries = groups.get(key) ?? [];
    entries.push(resolution);
    groups.set(key, entries);
  }

  const submit = async (skip = false) => {
    if (submission.current || (active.kind === "choice" && !canConfirmChoice)) {
      return;
    }
    submission.current = active.id;
    setSubmitting(true);
    setError(null);
    try {
      if (active.kind === "card-discard") {
        await globalWebSocketManager.confirmCardDiscard(active.id, skip ? [] : validCards);
      } else if (choice && canConfirmChoice) {
        await globalWebSocketManager.confirmBehaviorChoice(
          active.id,
          choice.originalIndex,
          targets,
        );
      }
    } catch (reason) {
      submission.current = null;
      setSubmitting(false);
      setError(reason instanceof Error ? reason.message : "Could not resolve effect");
    }
  };

  return (
    <GameFlowPopover
      isVisible
      type="interactive-mandatory"
      className="w-[min(920px,95vw)]! max-h-[90vh]! text-white"
    >
      <GameFlowTitle>Resolve effects · {resolutions.length} remaining</GameFlowTitle>
      <div className="flex flex-wrap gap-2 px-4 pb-3">
        {[...groups.entries()].map(([key, entries]) => (
          <GameButton
            key={key}
            size="sm"
            emphasis="secondary"
            aria-pressed={entries.some((item) => item.id === active.id)}
            selected={entries.some((item) => item.id === active.id)}
            disabled={submitting}
            onClick={() => setActiveId(entries[0].id)}
          >
            {entries[0].source} · {entries.length}
          </GameButton>
        ))}
      </div>
      <GameFlowBody>
        <div className="flex flex-col gap-4 p-1">
          <div>
            <p className="font-orbitron text-sm">{active.source}</p>
            {active.triggeringCardName && (
              <p className="text-sm text-white/70">Triggered by {active.triggeringCardName}</p>
            )}
          </div>
          {active.kind === "card-discard" ? (
            <>
              <p className="text-sm">
                Select{" "}
                {active.minCards === active.maxCards ? active.minCards : `up to ${active.maxCards}`}{" "}
                card(s) to discard.
              </p>
              <BehaviorSection
                behaviors={[{ outputs: active.outputs }]}
                noContainer
                showTooltips={false}
              />
              {handCards.length === 0 && (
                <p className="text-white/70">
                  No cards in hand. You can skip or resolve another effect first.
                </p>
              )}
              <div className="flex flex-wrap gap-3 justify-center">
                {handCards.map((card) => (
                  <div key={card.id} className="w-[var(--card-width)]">
                    <CardChoice
                      card={card}
                      isSelected={validCards.includes(card.id)}
                      disabled={submitting}
                      showCheckbox
                      onSelect={(id) => {
                        if (submitting) {
                          return;
                        }
                        setSelectedCards((previous) => {
                          const current = previous.filter((item) =>
                            handCards.some((candidate) => candidate.id === item),
                          );
                          if (current.includes(id)) {
                            return current.filter((item) => item !== id);
                          }
                          return current.length < active.maxCards ? [...current, id] : current;
                        });
                      }}
                    />
                  </div>
                ))}
              </div>
            </>
          ) : (
            <div className="flex flex-col gap-2">
              {active.choices?.map((item) => (
                <div key={item.originalIndex}>
                  <GameButton
                    emphasis="secondary"
                    aria-pressed={choiceIndex === item.originalIndex}
                    className="w-full p-3"
                    selected={choiceIndex === item.originalIndex}
                    disabled={!item.available || submitting}
                    onClick={() => {
                      setChoiceIndex(item.originalIndex);
                      setTargets([]);
                      setError(null);
                    }}
                  >
                    <div className="flex flex-col gap-2">
                      <BehaviorSection
                        behaviors={[
                          {
                            triggers: [{ type: "manual" }],
                            inputs: item.inputs,
                            outputs: item.outputs,
                          },
                        ]}
                        hideActionChip
                        noContainer
                        showTooltips={false}
                      />
                      {item.outputs
                        ?.filter((output) => output.target === "triggering-card")
                        .map((output, index) => (
                          <span key={index} className="text-sm">
                            Add {output.amount} {output.type} to {active.triggeringCardName}
                          </span>
                        ))}
                    </div>
                  </GameButton>
                  {!item.available &&
                    item.errors?.map((issue, index) => (
                      <p key={index} className="text-sm text-white/60 mt-1">
                        {issue.message}
                      </p>
                    ))}
                </div>
              ))}
              {storageTargets.map((eligible, index) => (
                <label key={index} className="flex flex-col gap-1 text-sm">
                  Resource destination {index + 1}
                  <select
                    className="bg-black border border-white/30 rounded p-2"
                    value={targets[index] ?? ""}
                    disabled={submitting}
                    onChange={(event) =>
                      setTargets((previous) => {
                        const next = [...previous];
                        next[index] = event.target.value;
                        return next;
                      })
                    }
                  >
                    <option value="">Choose a card</option>
                    {eligible.map((id) => (
                      <option key={id} value={id}>
                        {allCards.find((card) => card.id === id)?.name ?? id}
                      </option>
                    ))}
                  </select>
                </label>
              ))}
            </div>
          )}
          {error && (
            <p role="alert" className="text-sm text-red-300">
              {error}
            </p>
          )}
        </div>
      </GameFlowBody>
      <GameFlowFooter>
        <div className="flex justify-end gap-2">
          {active.kind === "card-discard" && active.minCards === 0 && (
            <GameButton emphasis="quiet" disabled={submitting} onClick={() => void submit(true)}>
              Skip
            </GameButton>
          )}
          <GameButton
            disabled={
              submitting || (active.kind === "choice" ? !canConfirmChoice : !canConfirmDiscard)
            }
            onClick={() => void submit()}
          >
            {submitting ? "Resolving…" : "Confirm"}
          </GameButton>
        </div>
      </GameFlowFooter>
    </GameFlowPopover>
  );
}
