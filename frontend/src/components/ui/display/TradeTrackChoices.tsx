import type { ColonyTradeOptionDto } from "@/types/generated/api-types.ts";
import GameButton from "../buttons/GameButton.tsx";

export function selectedTradeOption(options: ColonyTradeOptionDto[], selected?: number) {
  return options.find((option) => option.trackSteps === selected) ?? options.at(-1);
}

interface TradeTrackChoicesProps {
  options: ColonyTradeOptionDto[];
  selected: number;
  disabled?: boolean;
  onSelect: (steps: number) => void;
}

export default function TradeTrackChoices({
  options,
  selected,
  disabled,
  onSelect,
}: TradeTrackChoicesProps) {
  if (options.length < 2) {
    return null;
  }
  return (
    <div className="flex items-center gap-2 my-2">
      <span className="font-orbitron text-[10px] text-white/60">Track increase</span>
      <div className="flex" role="group" aria-label="Trade track increase">
        {options.map((option) => (
          <GameButton
            key={option.trackSteps}
            size="xs"
            shape="toolbar"
            emphasis="secondary"
            selected={selected === option.trackSteps}
            aria-pressed={selected === option.trackSteps}
            disabled={disabled}
            onClick={() => onSelect(option.trackSteps)}
          >
            {option.trackSteps === 0 ? "0" : `+${option.trackSteps}`}
          </GameButton>
        ))}
      </div>
    </div>
  );
}
