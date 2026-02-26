import React, { useEffect, useState } from "react";
import { CardTag, PlayerDto } from "../../../types/generated/api-types.ts";
import { fetchAllCards } from "../../../utils/cardPlayabilityUtils.ts";
import GameIcon from "../display/GameIcon.tsx";
import { getTagIconPath } from "../../../utils/iconStore.ts";
import { GamePopover, GamePopoverItem } from "../GamePopover";

interface StorageItem {
  cardId: string;
  cardName: string;
  resourceType: string;
  count: number;
  tags: CardTag[];
}

interface StoragesPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  player: PlayerDto;
  anchorRef: React.RefObject<HTMLElement>;
}

const StoragesPopover: React.FC<StoragesPopoverProps> = ({
  isVisible,
  onClose,
  player,
  anchorRef,
}) => {
  const [storageItems, setStorageItems] = useState<StorageItem[]>([]);

  useEffect(() => {
    const fetchStorageCards = async () => {
      if (!player.resourceStorage) {
        setStorageItems([]);
        return;
      }

      try {
        const allCards = await fetchAllCards();
        const items: StorageItem[] = [];

        for (const [cardId, count] of Object.entries(player.resourceStorage)) {
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

    if (isVisible) {
      void fetchStorageCards();
    }
  }, [player.resourceStorage, isVisible]);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="storages"
      header={{
        title: "Card Storages",
        badge: `${storageItems.length} card${storageItems.length !== 1 ? "s" : ""}`,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {storageItems.length === 0 ? (
        <div className="flex items-center justify-center py-10 px-5">
          <span className="font-orbitron text-sm text-white/50">No storages</span>
        </div>
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {storageItems.map((storage, index) => (
            <GamePopoverItem
              key={storage.cardId}
              state="available"
              hoverEffect="glow"
              animationDelay={index * 0.05}
            >
              <div className="flex justify-between items-center flex-1">
                <div className="flex flex-col gap-1">
                  <div className="text-white/90 text-sm font-semibold font-orbitron [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] max-[768px]:text-xs">
                    {storage.cardName}
                  </div>
                  {storage.tags.length > 0 && (
                    <div className="flex items-center gap-1">
                      {storage.tags.map((tag, tagIndex) => {
                        const tagIcon = getTagIconPath(tag);
                        if (!tagIcon) return null;
                        return (
                          <img
                            key={`${storage.cardId}-tag-${tagIndex}`}
                            src={tagIcon}
                            alt={tag}
                            className="w-[16px] h-[16px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-[768px]:w-[14px] max-[768px]:h-[14px]"
                          />
                        );
                      })}
                    </div>
                  )}
                </div>

                <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
                  <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm">
                    {storage.count}
                  </span>
                  <GameIcon iconType={storage.resourceType} size="small" />
                </div>
              </div>
            </GamePopoverItem>
          ))}
        </div>
      )}
    </GamePopover>
  );
};

export default StoragesPopover;
