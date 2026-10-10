import GameSelect from "./GameSelect.tsx";
import { MAX_PLAYER_NAME_LENGTH } from "@/constants/gameConstants.ts";

interface Props {
  id: string;
  seats: { id: string; name: string }[];
  seatId: string;
  name: string;
  disabled?: boolean;
  onSeatChange: (id: string) => void;
  onNameChange: (name: string) => void;
}

export default function ResumeSeatFields({
  id,
  seats,
  seatId,
  name,
  disabled,
  onSeatChange,
  onNameChange,
}: Props) {
  return (
    <>
      <div className="flex flex-col gap-2 text-sm">
        <label htmlFor={id} className="font-orbitron">
          Choose your seat
        </label>
        <GameSelect
          id={id}
          label="Choose your seat"
          className="w-full"
          value={seatId || "Select a player"}
          options={seats}
          disabled={disabled || seats.length === 0}
          onChange={onSeatChange}
        />
      </div>
      <input
        className="game-input text-lg"
        placeholder="Your name"
        aria-label="Your name"
        autoComplete="nickname"
        value={name}
        maxLength={MAX_PLAYER_NAME_LENGTH}
        disabled={disabled || !seatId}
        onChange={(event) => onNameChange(event.target.value)}
      />
    </>
  );
}
