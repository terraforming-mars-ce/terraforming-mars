import React from "react";
import GameIcon from "../display/GameIcon.tsx";
import { GamePopover, GamePopoverItem } from "../GamePopover";

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

const TagsPopover: React.FC<TagsPopoverProps> = ({ isVisible, onClose, tagCounts, anchorRef }) => {
  const visibleTags = tagCounts.filter((tag) => tag.count > 0);
  const totalTags = visibleTags.reduce((sum, tag) => sum + tag.count, 0);

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
      {visibleTags.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No tags</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {visibleTags.map((tagData, index) => (
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
                <span className="ml-auto text-base font-bold text-white font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {tagData.count}
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
