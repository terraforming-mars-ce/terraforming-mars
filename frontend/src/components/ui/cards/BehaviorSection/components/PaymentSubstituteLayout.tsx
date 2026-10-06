import { BehaviorArrow } from "./BehaviorIcon";
import GameIcon from "../../../display/GameIcon";
import type { CardBehaviorDto } from "@/types/generated/api-types";

export default function PaymentSubstituteLayout({ behavior }: { behavior: CardBehaviorDto }) {
  const output = behavior.outputs?.find((o) => o.type === "payment-substitute");
  if (!output || output.type !== "payment-substitute") {
    return null;
  }
  const tags = [...new Set(output.selectors?.flatMap((s) => s.tags ?? []) ?? [])];
  return (
    <div className="flex gap-1 items-center justify-center font-orbitron font-bold">
      <GameIcon iconType={output.source.resource} size="behavior" />
      <BehaviorArrow />
      {output.targetResource !== "credit" && output.amount > 1 && <span>{output.amount}</span>}
      <GameIcon
        iconType={output.targetResource}
        amount={output.targetResource === "credit" ? output.amount : undefined}
        size="behavior"
      />
      {tags.map((tag) => (
        <span key={tag} className="ml-1 inline-flex items-center">
          <GameIcon iconType={`${tag}-tag`} size="behavior" />
        </span>
      ))}
    </div>
  );
}
