import React from "react";
import type { PerConditionDto } from "@/types/generated/api-types";
import { getIconPath, getTagIconPath } from "@/utils/iconStore";

export default function PerConditionIcons({ per }: { per: PerConditionDto }) {
  const tags = [...new Set([...(per.tags ?? []), ...(per.tag ? [per.tag] : [])])];
  const icons = tags.length
    ? tags.map((tag) => ({ name: tag, src: getTagIconPath(tag) }))
    : [{ name: per.type, src: getIconPath(per.type === "card-count" ? "card-draw" : per.type) }];
  // City counts default to all owners, matching the backend tile counter.
  const others = per.target
    ? !["self-player", "self-card"].includes(per.target)
    : per.type === "city-tile";
  return (
    <span className="inline-flex items-center gap-px font-orbitron text-white">
      {per.amount > 1 && (
        <span className="text-[13px] font-bold leading-none max-md:text-[11px]">{per.amount}</span>
      )}
      {icons.map((icon, index) => (
        <React.Fragment key={icon.name}>
          {index > 0 && <span className="mx-0.5 text-xs font-bold">+</span>}
          {icon.src && (
            <img
              src={icon.src}
              alt={icon.name}
              className={`w-[var(--behavior-icon-size,26px)] h-[var(--behavior-icon-size,26px)] object-contain shrink-0 max-md:w-[var(--behavior-icon-small-size,22px)] max-md:h-[var(--behavior-icon-small-size,22px)] ${
                others
                  ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
                  : "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
              }`}
            />
          )}
        </React.Fragment>
      ))}
    </span>
  );
}
