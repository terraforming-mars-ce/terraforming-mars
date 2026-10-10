import { useEffect, useRef, useState } from "react";
import type { ErrorPayload, PendingResourceRemovalSelectionDto } from "@/types/generated/api-types";
import { globalWebSocketManager } from "@/services/globalWebSocketManager";
import { webSocketService } from "@/services/webSocketService";
import GameButton from "../buttons/GameButton";
import GameIcon from "../display/GameIcon";
import {
  GameFlowPopover,
  GameFlowTitle,
  GameFlowBody,
  GameFlowFooter,
} from "../popover/GameFlowPopover";

interface Props {
  selection: PendingResourceRemovalSelectionDto;
  players: { id: string; name: string }[];
  currentPlayerId: string;
}

export default function ResourceRemovalOverlay({ selection, players, currentPlayerId }: Props) {
  const [target, setTarget] = useState("");
  const [amount, setAmount] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pending = useRef(false);
  const maximum = selection.maxAmounts[target] ?? 0;

  useEffect(() => {
    const failed = (payload: unknown) => {
      const result = payload as Partial<ErrorPayload>;
      if (pending.current && result.selectionId === selection.id) {
        pending.current = false;
        setSubmitting(false);
        setError(result.message ?? "Could not remove resources");
      }
    };
    const disconnected = () => {
      pending.current = false;
      setSubmitting(false);
      setError("Connection lost. Reconnect to check whether this selection was resolved.");
    };
    webSocketService.on("error", failed);
    webSocketService.on("disconnect", disconnected);
    return () => {
      webSocketService.off("error", failed);
      webSocketService.off("disconnect", disconnected);
    };
  }, [selection.id]);

  const submit = async (skip: boolean) => {
    if (
      pending.current ||
      (!skip && (!target || !Number.isInteger(amount) || amount < 1 || amount > maximum))
    )
      return;
    pending.current = true;
    setSubmitting(true);
    setError(null);
    try {
      await globalWebSocketManager.confirmResourceRemoval(
        selection.id,
        skip ? "" : target,
        skip ? 0 : amount,
      );
    } catch (reason) {
      pending.current = false;
      setSubmitting(false);
      setError(reason instanceof Error ? reason.message : "Could not remove resources");
    }
  };

  return (
    <GameFlowPopover isVisible type="interactive-mandatory" className="min-w-[280px]">
      <GameFlowTitle>
        <h3 className="font-orbitron text-white text-base font-bold">{selection.source}</h3>
        <div className="flex items-center justify-center gap-2 text-white/70 text-sm">
          Remove up to{" "}
          <GameIcon iconType={selection.resourceType} amount={selection.amount} size="small" />
        </div>
      </GameFlowTitle>
      <GameFlowBody>
        {players
          .filter((player) => selection.eligiblePlayerIds.includes(player.id))
          .map((player) => (
            <GameButton
              key={player.id}
              className="w-full mb-2 compact:min-h-11"
              emphasis={target === player.id ? "primary" : "secondary"}
              disabled={submitting}
              onClick={() => {
                setTarget(player.id);
                setAmount(selection.maxAmounts[player.id]);
              }}
            >
              {player.name}
              {player.id === currentPlayerId ? " (you)" : ""}
            </GameButton>
          ))}
        {target && (
          <label className="flex items-center justify-between gap-3 text-white text-sm compact:min-h-11">
            Amount
            <input
              aria-label="Amount to remove"
              className="w-20 compact:w-24 compact:min-h-11 rounded border border-white/30 bg-black/40 p-2"
              type="number"
              min={1}
              max={maximum}
              step={1}
              value={amount}
              disabled={submitting}
              onChange={(event) => setAmount(Number(event.target.value))}
            />
          </label>
        )}
        {error && (
          <p role="alert" className="text-red-300 text-sm mt-2">
            {error}
          </p>
        )}
      </GameFlowBody>
      <GameFlowFooter className="compact:gap-3">
        <GameButton
          emphasis="primary"
          disabled={
            submitting || !target || !Number.isInteger(amount) || amount < 1 || amount > maximum
          }
          onClick={() => void submit(false)}
        >
          Remove
        </GameButton>
        <GameButton emphasis="secondary" disabled={submitting} onClick={() => void submit(true)}>
          Skip
        </GameButton>
      </GameFlowFooter>
    </GameFlowPopover>
  );
}
