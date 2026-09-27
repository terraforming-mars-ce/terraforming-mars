import GameButton from "./GameButton.tsx";

interface Props {
  onClick: () => void;
  label?: string;
}

export default function CloseButton({ onClick, label = "Close dialog" }: Props) {
  return (
    <GameButton
      emphasis="quiet"
      aria-label={label}
      onClick={onClick}
      onMouseDown={(event) => event.stopPropagation()}
      className="!p-0 !min-h-11 w-11 h-11 shrink-0 !font-sans !text-3xl !font-normal leading-none"
    >
      ×
    </GameButton>
  );
}
