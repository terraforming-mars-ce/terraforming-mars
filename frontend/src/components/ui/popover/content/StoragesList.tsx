import React, { useEffect, useState } from "react";
import { CardTag } from "@/types/generated/api-types.ts";
import { fetchAllCards } from "@/utils/cardPlayabilityUtils.ts";
import { getTagIconPath } from "@/utils/iconStore.ts";
import GameIcon from "../../display/GameIcon.tsx";
import { GamePopoverItem, getThemeStyles } from "../../GamePopover";
import type { ContentDensity } from "./density.ts";

export interface StorageItem {
  cardId: string;
  cardName: string;
  resourceType: string;
  count: number;
  tags: CardTag[];
}

export function useStorageItems(
  resourceStorage: Record<string, number> | undefined,
  enabled: boolean,
): StorageItem[] {
  const [storageItems, setStorageItems] = useState<StorageItem[]>([]);

  useEffect(() => {
    const fetchStorageCards = async () => {
      if (!resourceStorage) {
        setStorageItems([]);
        return;
      }

      try {
        const allCards = await fetchAllCards();
        const items: StorageItem[] = [];

        for (const [cardId, count] of Object.entries(resourceStorage)) {
          const card = allCards.get(cardId);
          if (card && card.resourceStorage) {
            items.push({
              cardId,
              cardName: card.name,
              resourceType: card.resourceStorage.type,
              count,
              tags: card.tags ?? [],
            });
          }
        }
        setStorageItems(items);
      } catch (error) {
        console.error("Failed to fetch cards:", error);
        setStorageItems([]);
      }
    };

    if (enabled) {
      void fetchStorageCards();
    }
  }, [resourceStorage, enabled]);

  return storageItems;
}

interface StoragesListProps {
  items: StorageItem[];
  density: ContentDensity;
}

const StoragesList: React.FC<StoragesListProps> = ({ items, density }) => {
  const isScreen = density === "screen";

  if (items.length === 0) {
    return (
      <div className="flex items-center justify-center py-10 px-5">
        <span className="font-orbitron text-sm text-white/50">No storages</span>
      </div>
    );
  }

  const listClass = isScreen
    ? "popover-list p-3 grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-2"
    : "popover-list popover-list-headed p-2 flex flex-col gap-2";
  const nameClass = isScreen
    ? "text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]"
    : "text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] max-[768px]:text-xs";
  const tagClass = isScreen
    ? "w-[20px] h-[20px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))]"
    : "w-[16px] h-[16px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-[768px]:w-[14px] max-[768px]:h-[14px]";
  const countClass = isScreen
    ? "text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right"
    : "text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm";

  return (
    <div className={listClass} style={isScreen ? getThemeStyles("storages") : undefined}>
      {items.map((storage, index) => (
        <GamePopoverItem
          className={isScreen ? "popover-list-item min-h-11" : "popover-list-item"}
          key={storage.cardId}
          state="available"
          animationDelay={index * 0.05}
        >
          <div className="flex justify-between items-center flex-1">
            <div className="flex flex-col gap-1">
              <div className={nameClass}>{storage.cardName}</div>
              {storage.tags.length > 0 && (
                <div className="flex items-center gap-1">
                  {storage.tags.map((tag, tagIndex) => {
                    const tagIcon = getTagIconPath(tag);
                    if (!tagIcon) {
                      return null;
                    }
                    return (
                      <img
                        key={`${storage.cardId}-tag-${tagIndex}`}
                        src={tagIcon}
                        alt={tag}
                        className={tagClass}
                      />
                    );
                  })}
                </div>
              )}
            </div>

            <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
              <span className={countClass}>{storage.count}</span>
              <GameIcon iconType={storage.resourceType} size="small" />
            </div>
          </div>
        </GamePopoverItem>
      ))}
    </div>
  );
};

export default StoragesList;
