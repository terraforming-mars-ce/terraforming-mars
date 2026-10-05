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

const WildTagRow: React.FC<{ count: number }> = ({ count }) => {
  const ref = useRef<HTMLButtonElement>(null);
  const [tooltipPos, setTooltipPos] = useState<{ x: number; y: number } | null>(null);
  const explanation =
    "Each wild tag can represent one tag during your action. The same wild cannot count as multiple tags at once. Wild tags do not count for endgame scoring.";
  const showTooltip = () => {
    const rect = ref.current?.getBoundingClientRect();
    if (rect) setTooltipPos({ x: rect.left + rect.width / 2, y: rect.top });
  };
  return (
    <>
      <button
        ref={ref}
        type="button"
        className="flex items-center gap-3 flex-1 w-full text-white/90 font-orbitron cursor-help rounded focus-visible:outline focus-visible:outline-white/60"
        aria-label={`Wild × ${count}. ${explanation}`}
        onMouseEnter={showTooltip}
        onMouseLeave={() => setTooltipPos(null)}
        onFocus={showTooltip}
        onBlur={() => setTooltipPos(null)}
        onClick={() => (tooltipPos ? setTooltipPos(null) : showTooltip())}
      >
        <GameIcon iconType="wild-tag" size="medium" />
        <span className="text-sm font-semibold">Wild</span>
        <span className="ml-auto text-base font-bold">× {count}</span>
      </button>
      <DecorBoxTooltip position={tooltipPos} placement="above" cornerSize={10}>
        <div className="max-w-64 text-sm leading-relaxed">{explanation}</div>
      </DecorBoxTooltip>
    </>
  );
};

const TagsPopover: React.FC<TagsPopoverProps> = ({ isVisible, onClose, tagCounts, anchorRef }) => {
  const wildCount = tagCounts.find((t) => t.tag === TagWild)?.count || 0;
  const nonWildTags = tagCounts.filter((tag) => tag.tag !== TagWild && tag.count > 0);
  const totalTags = nonWildTags.reduce((sum, tag) => sum + tag.count, 0) + wildCount;

  return (
    <GamePopover
      className="game-popover-list"
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="tags"
      header={{ title: "Tags", badge: `${totalTags} total` }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {totalTags === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No tags</span>
        </div>
      ) : (
        <div className="popover-list popover-list-headed p-2 flex flex-col gap-2">
          {nonWildTags.map((tagData, index) => (
            <GamePopoverItem
              className="popover-list-item"
              key={tagData.tag}
              state="available"
              animationDelay={index * 0.05}
            >
              <div className="flex items-center gap-3 flex-1">
                <GameIcon iconType={`${tagData.tag}-tag`} size="medium" />
                <span className="text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {tagData.tag.charAt(0).toUpperCase() + tagData.tag.slice(1)}
                </span>
                <span className="ml-auto flex items-center gap-1.5 text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {tagData.count}
                </span>
              </div>
            </GamePopoverItem>
          ))}
          {wildCount > 0 && (
            <GamePopoverItem
              className="popover-list-item"
              state="available"
              animationDelay={nonWildTags.length * 0.05}
            >
              <WildTagRow count={wildCount} />
            </GamePopoverItem>
          )}
        </div>
      )}
    </GamePopover>
  );
};

export default TagsPopover;
