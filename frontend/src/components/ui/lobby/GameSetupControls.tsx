import GameSelect from "../GameSelect.tsx";
import CloseButton from "@/components/ui/buttons/CloseButton.tsx";
import { useState } from "react";
import type { MapInfoDto, UpdateGameSettingsRequest } from "@/types/generated/api-types.ts";
import { CARD_PACKS, VENUS_PACK } from "@/constants/cardPacks.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import GameButton from "../buttons/GameButton.tsx";
import { GameModal } from "../GameModal";
import MapPreview from "./MapPreview.tsx";

interface Props {
  maps: MapInfoDto[];
  mapId: string;
  cardPacks: string[];
  venusNextEnabled: boolean;
  onChange?: (patch: UpdateGameSettingsRequest) => void;
  disabled?: boolean;
}

const packs = [...CARD_PACKS, VENUS_PACK];

export default function GameSetupControls({
  maps,
  mapId,
  cardPacks,
  venusNextEnabled,
  onChange,
  disabled,
}: Props) {
  const [choosingPacks, setChoosingPacks] = useState(false);
  const activeMap = maps.find((map) => map.id === mapId);
  const selected = (id: string) =>
    id === VENUS_PACK.id ? venusNextEnabled : cardPacks.includes(id);
  const togglePack = (id: string) => {
    if (id === VENUS_PACK.id) {
      onChange?.({ venusNextEnabled: !venusNextEnabled });
      return;
    }
    onChange?.({
      cardPacks: selected(id) ? cardPacks.filter((pack) => pack !== id) : [...cardPacks, id],
    });
  };
  return (
    <section className="text-white text-left flex flex-col gap-5 min-w-0">
      <div className="flex items-center justify-between gap-4">
        <label htmlFor="setup-map" className="text-sm text-white/60">
          Map
        </label>
        {onChange ? (
          <GameSelect
            id="setup-map"
            label="Map"
            value={mapId}
            options={maps}
            disabled={disabled}
            onChange={(id) => onChange({ mapId: id })}
          />
        ) : (
          <span className="font-orbitron text-sm">{activeMap?.name ?? mapId}</span>
        )}
      </div>
      <div
        key={mapId}
        className="menu-enter flex justify-center py-3 [&_svg]:w-full [&_svg]:max-w-[340px] [&_svg]:h-auto"
      >
        <MapPreview tiles={activeMap?.tiles ?? []} />
      </div>
      <div className="flex items-center justify-between gap-4 border-t border-white/10 pt-4">
        <div className="min-w-0">
          <span className="text-sm text-white/60">Expansions</span>
          <p className="mt-2 text-sm leading-relaxed">
            {packs
              .filter((pack) => selected(pack.id))
              .map((pack) => pack.label)
              .join(", ")}
          </p>
        </div>
        {onChange && (
          <GameButton
            emphasis="secondary"
            size="sm"
            disabled={disabled}
            onClick={() => setChoosingPacks(true)}
          >
            Edit
          </GameButton>
        )}
      </div>
      <GameModal
        isVisible={choosingPacks}
        onClose={() => setChoosingPacks(false)}
        theme="default"
        size="medium"
        animation="fadeIn"
        zIndex={Z_INDEX.LOBBY_SETTINGS_MODAL}
      >
        <div className="p-6 sm:p-8 overflow-y-auto">
          <div className="flex justify-between items-center gap-4 mb-6">
            <h2 className="font-orbitron text-xl">Expansions</h2>
            <CloseButton onClick={() => setChoosingPacks(false)} label="Close expansions" />
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {packs.map((pack) => (
              <GameButton
                key={pack.id}
                emphasis="secondary"
                selected={selected(pack.id)}
                disabled={disabled || pack.lockedOn}
                aria-pressed={selected(pack.id)}
                onClick={() => togglePack(pack.id)}
                className="text-left !p-5 !justify-start"
              >
                <span className="flex flex-col gap-2">
                  <span>{pack.label}</span>
                  <span className="font-sans font-normal text-sm text-white/60 leading-relaxed">
                    {pack.description}
                  </span>
                  {pack.wip && (
                    <span className="font-sans text-xs text-amber-300">Experimental</span>
                  )}
                </span>
              </GameButton>
            ))}
          </div>
        </div>
      </GameModal>
    </section>
  );
}
