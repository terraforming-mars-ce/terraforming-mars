import type { PlayerCardDto } from "@/types/generated/api-types.ts";

interface CardStatusMessagesProps {
  card: Pick<PlayerCardDto, "available" | "errors" | "warnings">;
  size?: "small" | "large";
}

export default function CardStatusMessages({ card, size = "small" }: CardStatusMessagesProps) {
  const sizing = size === "large" ? "text-base py-3 px-4" : "text-xs py-2 px-2.5";
  return (
    <div className={`flex flex-col ${size === "large" ? "gap-2" : "gap-1"}`} data-card-status>
      {!card.available &&
        card.errors.map((error, index) => (
          <div
            key={`error-${index}`}
            className={`bg-[rgba(10,10,15,0.95)] border border-[rgba(231,76,60,0.6)] border-l-[3px] border-l-[#e74c3c] text-white/90 leading-[1.4] whitespace-normal ${sizing}`}
          >
            {error.message}
          </div>
        ))}
      {card.warnings?.map((warning, index) => (
        <div
          key={`warning-${index}`}
          className={`bg-[rgba(10,10,15,0.95)] border border-[rgba(255,193,7,0.6)] border-l-[3px] border-l-[#ffc107] text-white/90 leading-[1.4] whitespace-normal ${sizing}`}
        >
          {warning.message}
        </div>
      ))}
    </div>
  );
}
