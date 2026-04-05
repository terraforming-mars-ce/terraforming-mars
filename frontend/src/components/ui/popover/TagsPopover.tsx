import React, { useState, useRef } from "react";
import GameIcon from "../display/GameIcon.tsx";
import DecorBoxTooltip from "../display/DecorBoxTooltip.tsx";
import { GamePopover, GamePopoverItem } from "../GamePopover";
import { TagWild } from "@/types/generated/api-types.ts";

interface TagCount {
  tag: string;
  count: number;
}

interface TagsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  tagCounts: TagCount[];
  anchorRef: React.RefObject<HTMLElement>;
}

const WildBadge: React.FC<{ count: number }> = ({ count }) => {
  const badgeRef = useRef<HTMLSpanElement>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);

  const handleMouseEnter = () => {
    if (badgeRef.current) {
      const rect = badgeRef.current.getBoundingClientRect();
      setTooltipPos({ x: rect.left + rect.width / 2, y: rect.top });
    }
  };

  const handleMouseLeave = () => {
    setTooltipPos(null);
  };

  return (
    <>
      <span
        ref={badgeRef}
        className="text-base font-bold font-orbitron text-white/60 [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] cursor-default"
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        +{count}
      </span>
      <DecorBoxTooltip position={tooltipPos} placement="above" cornerSize={10}>
        <div className="flex items-center gap-1.5 whitespace-nowrap">
          <GameIcon iconType="wild-tag" size="small" />
          <span className="font-orbitron text-white font-bold">
            {count} wild {count === 1 ? "tag" : "tags"}
          </span>
        </div>
      </DecorBoxTooltip>
    </>
  );
};

const TagsPopover: React.FC<TagsPopoverProps> = ({ isVisible, onClose, tagCounts, anchorRef }) => {
  const wildCount = tagCounts.find((t) => t.tag === TagWild)?.count || 0;
  const nonWildTags = tagCounts.filter((tag) => tag.tag !== TagWild && tag.count > 0);
  const totalTags = nonWildTags.reduce((sum, tag) => sum + tag.count, 0);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="tags"
      header={{ title: "Tags", badge: `${totalTags} total` }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {nonWildTags.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No tags</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {nonWildTags.map((tagData, index) => (
            <GamePopoverItem
              key={tagData.tag}
              state="available"
              hoverEffect="glow"
              animationDelay={index * 0.05}
            >
              <div className="flex items-center gap-3 flex-1">
                <GameIcon iconType={`${tagData.tag}-tag`} size="medium" />
                <span className="text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {tagData.tag.charAt(0).toUpperCase() + tagData.tag.slice(1)}
                </span>
                <span className="ml-auto flex items-center gap-1.5 text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {tagData.count}
                  {wildCount > 0 && <WildBadge count={wildCount} />}
                </span>
              </div>
            </GamePopoverItem>
          ))}
        </div>
      )}
    </GamePopover>
  );
};

export default TagsPopover;
