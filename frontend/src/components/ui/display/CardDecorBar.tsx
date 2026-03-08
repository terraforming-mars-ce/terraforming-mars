import React, { useState, useRef, useEffect } from "react";
import DecorBox from "./DecorBox.tsx";
import DecorBoxTooltip from "./DecorBoxTooltip.tsx";
import VictoryPointIcon from "./VictoryPointIcon.tsx";
import ResourceStorageIcon from "./ResourceStorageIcon.tsx";
import { ResourceStorageDto } from "@/types/generated/api-types.ts";

interface CardDecorBarProps {
  vpConditions?: any[];
  resourceStorage?: ResourceStorageDto;
  corner?: "bottom-right" | "top-left";
}

const CardDecorBar: React.FC<CardDecorBarProps> = ({
  vpConditions,
  resourceStorage,
  corner = "bottom-right",
}) => {
  const [vpDescription, setVpDescription] = useState<string | null>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (vpDescription && ref.current) {
      const rect = ref.current.getBoundingClientRect();
      setTooltipPos({ x: rect.left + rect.width / 2, y: rect.bottom });
    } else {
      setTooltipPos(null);
    }
  }, [vpDescription]);

  const hasVp = vpConditions && vpConditions.length > 0;
  const hasStorage = !!resourceStorage;

  if (!hasVp && !hasStorage) {
    return null;
  }

  return (
    <div className="relative w-fit" ref={ref}>
      <DecorBox corner={corner}>
        {hasVp && (
          <VictoryPointIcon
            vpConditions={vpConditions}
            onHoverDescription={setVpDescription}
            bare
          />
        )}
        {hasVp && hasStorage && <div className="w-px h-3 bg-white/20 mx-1" />}
        {hasStorage && <ResourceStorageIcon resourceStorage={resourceStorage} bare />}
      </DecorBox>
      <DecorBoxTooltip description={vpDescription} position={tooltipPos} />
    </div>
  );
};

export default CardDecorBar;
