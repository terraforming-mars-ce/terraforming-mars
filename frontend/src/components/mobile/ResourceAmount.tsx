import { ResourceTypeCredit, type ResourceType } from "@/types/generated/api-types.ts";
import GameIcon from "../ui/display/GameIcon.tsx";

export type ResourceAmountSize = "sm" | "md";

const SIZE_CLASSES: Record<
  ResourceAmountSize,
  { icon: string; credit: string; amount: string; badge: string; gap: string }
> = {
  sm: {
    icon: "!w-[18px] !h-[18px]",
    credit: "!w-[22px] !h-[22px]",
    amount: "text-[13px]",
    badge: "min-w-[22px] text-[11px] leading-[13px]",
    gap: "gap-1",
  },
  md: {
    icon: "!w-6 !h-6",
    credit: "",
    amount: "text-[16px]",
    badge: "min-w-[28px] text-[12px] leading-[16px]",
    gap: "gap-1.5",
  },
};

interface ResourceAmountProps {
  resource: ResourceType;
  amount: number;
  production: number;
  size?: ResourceAmountSize;
  className?: string;
}

/** One resource as the desktop bar shows it: the amount, then its production box to the right. */
export default function ResourceAmount({
  resource,
  amount,
  production,
  size = "sm",
  className = "",
}: ResourceAmountProps) {
  const classes = SIZE_CLASSES[size];

  return (
    <div className={`flex items-center ${classes.gap} ${className}`}>
      {resource === ResourceTypeCredit ? (
        <GameIcon
          iconType={ResourceTypeCredit}
          amount={amount}
          size="small"
          className={classes.credit}
        />
      ) : (
        <span className="flex items-center gap-0.5">
          <GameIcon iconType={resource} size="small" className={classes.icon} />
          <span
            className={`font-orbitron font-bold leading-none tabular-nums text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] ${classes.amount}`}
          >
            {amount}
          </span>
        </span>
      )}
      <span
        className={`inline-flex items-center justify-center px-1 py-0.5 border border-[rgba(160,110,60,0.6)] bg-[linear-gradient(135deg,rgba(160,110,60,0.5)_0%,rgba(139,89,42,0.45)_100%)] font-orbitron font-bold text-white tabular-nums [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] ${classes.badge}`}
      >
        {production}
      </span>
    </div>
  );
}
