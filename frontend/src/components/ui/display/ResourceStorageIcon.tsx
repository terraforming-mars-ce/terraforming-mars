import React, { useState, useRef, useEffect } from "react";
import { getIconPath } from "@/utils/iconStore.ts";
import { ResourceStorageDto } from "@/types/generated/api-types.ts";
import DecorBoxTooltip from "./DecorBoxTooltip.tsx";
import DecorBox from "./DecorBox.tsx";

interface ResourceStorageIconProps {
  resourceStorage?: ResourceStorageDto;
  corner?: "bottom-right" | "top-left";
  bare?: boolean;
}

const ResourceStorageIcon: React.FC<ResourceStorageIconProps> = ({
  resourceStorage,
  corner = "bottom-right",
  bare = false,
}) => {
  const [hoverDescription, setHoverDescription] = useState<string | null>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (hoverDescription && ref.current) {
      const rect = ref.current.getBoundingClientRect();
      setTooltipPos({ x: rect.left + rect.width / 2, y: rect.bottom });
    } else {
      setTooltipPos(null);
    }
  }, [hoverDescription]);

  if (!resourceStorage) {
    return null;
  }

  const resourceIcon = getIconPath(resourceStorage.type);

  const content = (
    <div
      className="inline-flex items-center gap-1"
      onMouseEnter={() => {
        if (resourceStorage.description) {
          setHoverDescription(resourceStorage.description);
        }
      }}
      onMouseLeave={() => setHoverDescription(null)}
    >
      {resourceIcon && (
        <img src={resourceIcon} alt={resourceStorage.type} className="w-3.5 h-3.5 object-contain" />
      )}
      <span className="text-[9px] text-white/50 font-semibold tracking-wider uppercase">
        {resourceStorage.type}
      </span>
    </div>
  );

  if (bare) {
    return (
      <div className="relative w-fit" ref={ref}>
        {content}
        <DecorBoxTooltip description={hoverDescription} position={tooltipPos} />
      </div>
    );
  }

  return (
    <div className="relative w-fit" ref={ref}>
      <DecorBox
        corner={corner}
        onMouseEnter={() => {
          if (resourceStorage.description) {
            setHoverDescription(resourceStorage.description);
          }
        }}
        onMouseLeave={() => setHoverDescription(null)}
      >
        {resourceIcon && (
          <img
            src={resourceIcon}
            alt={resourceStorage.type}
            className="w-3.5 h-3.5 object-contain"
          />
        )}
        <span className="text-[9px] text-white/50 font-semibold tracking-wider uppercase">
          {resourceStorage.type}
        </span>
      </DecorBox>
      <DecorBoxTooltip description={hoverDescription} position={tooltipPos} />
    </div>
  );
};

export default ResourceStorageIcon;
