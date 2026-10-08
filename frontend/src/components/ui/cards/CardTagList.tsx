import type { CSSProperties } from "react";
import { CardDto, PlayerCardDto, TagWild } from "@/types/generated/api-types.ts";
import { getTagIconPath } from "@/utils/iconStore.ts";
import RevealTrigger from "../display/RevealTrigger.tsx";

export type TagIconSize = "sm" | "md" | "card";

const SIZE_CLASS: Record<TagIconSize, string> = {
  sm: "w-4 h-4",
  md: "w-5 h-5",
  card: "w-8 h-8",
};

export function TagIcon({ tag, size }: { tag: string; size: TagIconSize }) {
  const tagIcon = getTagIconPath(tag.toLowerCase());
  if (!tagIcon) {
    return null;
  }
  const image = (
    <img
      src={tagIcon}
      alt={tag}
      className={`${SIZE_CLASS[size]} object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`}
    />
  );
  const className =
    "flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]";
  if (tag.toLowerCase() !== TagWild) {
    return <div className={className}>{image}</div>;
  }
  return (
    <RevealTrigger
      as="div"
      className={className}
      content="A wild tag counts as any tag"
      gap={0}
      cornerSize={10}
    >
      {image}
    </RevealTrigger>
  );
}

interface CardTagListProps {
  card: CardDto | PlayerCardDto;
  size?: TagIconSize;
  className?: string;
  style?: CSSProperties;
}

/** The card's tags (at most three, two plus the event marker for events) as icons. */
export default function CardTagList({
  card,
  size = "sm",
  className = "flex items-center gap-1",
  style,
}: CardTagListProps) {
  const isEvent = card.type === "event";
  const tags = card.tags?.slice(0, isEvent ? 2 : 3) ?? [];
  if (tags.length === 0 && !isEvent) {
    return null;
  }
  return (
    <div className={className} style={style}>
      {tags.map((tag, index) => (
        <TagIcon key={`${tag}-${index}`} tag={tag} size={size} />
      ))}
      {isEvent && <TagIcon tag="event" size={size} />}
    </div>
  );
}
