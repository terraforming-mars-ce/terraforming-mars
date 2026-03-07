import React from "react";
import { RequirementDto } from "@/types/generated/api-types.ts";
import { getTagIconPath, getIconPath } from "@/utils/iconStore.ts";

interface ChoiceRequirementBoxProps {
  requirements?: { items: RequirementDto[] };
  children: React.ReactNode;
}

export function renderRequirementItems(items: RequirementDto[]): React.ReactNode {
  return items.map((req, index) => {
    if (req.type === "tags" && req.tag) {
      const tagIcon = getTagIconPath(req.tag);
      const count = req.min ?? 1;
      return (
        <span key={index} className="flex items-center gap-0.5">
          {count > 1 && (
            <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
              {count}
            </span>
          )}
          {tagIcon && (
            <img
              src={tagIcon}
              alt={req.tag}
              className="w-[18px] h-[18px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
            />
          )}
        </span>
      );
    }

    const globalParamTypes = ["temperature", "oxygen", "ocean", "venus", "tr"];
    if (globalParamTypes.includes(req.type)) {
      const icon = getIconPath(req.type);
      const value = req.min ?? req.max;
      const prefix = req.min != null ? "" : "max ";
      return (
        <span key={index} className="flex items-center gap-0.5">
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
            {prefix}
            {value}
          </span>
          {icon && (
            <img
              src={icon}
              alt={req.type}
              className="w-[18px] h-[18px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
            />
          )}
        </span>
      );
    }

    return null;
  });
}

const ChoiceRequirementBox: React.FC<ChoiceRequirementBoxProps> = ({ requirements, children }) => {
  if (!requirements || !requirements.items || requirements.items.length === 0) {
    return <>{children}</>;
  }

  return (
    <div className="border border-dashed border-white/30 rounded px-1.5 py-1 flex flex-col items-center gap-[3px]">
      <div className="flex items-center gap-1">{renderRequirementItems(requirements.items)}</div>
      {children}
    </div>
  );
};

export default ChoiceRequirementBox;
