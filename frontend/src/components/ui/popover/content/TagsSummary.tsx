import React from "react";
import GameIcon from "../../display/GameIcon.tsx";
import RevealTrigger from "../../display/RevealTrigger.tsx";
import { GamePopoverItem, getThemeStyles } from "../../GamePopover";
import { CardDto, TagWild } from "@/types/generated/api-types.ts";
import type { ContentDensity } from "./density.ts";

export interface TagCount {
  tag: string;
  count: number;
}

const ALL_TAGS = [
  "space",
  "earth",
  "science",
  "power",
  "building",
  "microbe",
  "animal",
  "plant",
  "event",
  "city",
  "venus",
  "jovian",
  "wild",
  "mars",
  "moon",
  "clone",
  "crime",
];

export function countPlayerTags(
  corporation: CardDto | null | undefined,
  playedCards: CardDto[],
): TagCount[] {
  if (playedCards.length === 0 && !corporation) {
    return [];
  }
  const counts: Record<string, number> = {};
  const add = (tags: string[] | undefined) => {
    tags?.forEach((tag) => {
      const tagKey = tag.toLowerCase();
      counts[tagKey] = (counts[tagKey] || 0) + 1;
    });
  };
  add(corporation?.tags);
  playedCards.forEach((card) => {
    if (card.type !== "event") {
      add(card.tags);
    }
  });
  return ALL_TAGS.map((tag) => ({ tag, count: counts[tag] || 0 }));
}

export function summarizeTags(tagCounts: TagCount[]) {
  const wildCount = tagCounts.find((t) => t.tag === TagWild)?.count || 0;
  const nonWildTags = tagCounts.filter((tag) => tag.tag !== TagWild && tag.count > 0);
  const totalTags = nonWildTags.reduce((sum, tag) => sum + tag.count, 0) + wildCount;
  return { wildCount, nonWildTags, totalTags };
}

const WILD_TAG_EXPLANATION =
  "Each wild tag can represent one tag during your action. The same wild cannot count as multiple tags at once. Wild tags do not count for endgame scoring.";

const WildTagRow: React.FC<{ count: number; isScreen: boolean }> = ({ count, isScreen }) => (
  <RevealTrigger
    as="button"
    className={`flex items-center gap-3 flex-1 w-full text-white/90 font-orbitron cursor-default rounded focus-visible:outline focus-visible:outline-white/60 ${isScreen ? "min-h-11" : ""}`}
    aria-label={`Wild × ${count}. ${WILD_TAG_EXPLANATION}`}
    content={<div className="max-w-64 text-sm leading-relaxed">{WILD_TAG_EXPLANATION}</div>}
    gap={0}
    cornerSize={10}
  >
    <GameIcon iconType="wild-tag" size="medium" />
    <span className="text-sm font-semibold">Wild</span>
    <span className="ml-auto text-base font-bold">× {count}</span>
  </RevealTrigger>
);

interface TagsSummaryProps {
  tagCounts: TagCount[];
  density: ContentDensity;
}

const TagsSummary: React.FC<TagsSummaryProps> = ({ tagCounts, density }) => {
  const isScreen = density === "screen";
  const { wildCount, nonWildTags, totalTags } = summarizeTags(tagCounts);

  if (totalTags === 0) {
    return (
      <div className="flex items-center justify-center py-10 px-5">
        <span className="font-orbitron text-sm text-white/50">No tags</span>
      </div>
    );
  }

  const listClass = isScreen
    ? "popover-list p-3 grid grid-cols-[repeat(auto-fill,minmax(180px,1fr))] gap-2"
    : "popover-list popover-list-headed p-2 flex flex-col gap-2";
  const rowClass = isScreen
    ? "flex items-center gap-3 flex-1 min-h-11"
    : "flex items-center gap-3 flex-1";

  return (
    <div className={listClass} style={isScreen ? getThemeStyles("tags") : undefined}>
      {nonWildTags.map((tagData, index) => (
        <GamePopoverItem
          className="popover-list-item"
          key={tagData.tag}
          state="available"
          animationDelay={index * 0.05}
        >
          <div className={rowClass}>
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
          <WildTagRow count={wildCount} isScreen={isScreen} />
        </GamePopoverItem>
      )}
    </div>
  );
};

export default TagsSummary;
