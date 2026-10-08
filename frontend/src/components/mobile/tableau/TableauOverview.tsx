import type { ReactNode } from "react";
import { CardTypePrelude, ResourceTypeTR, type CardDto } from "@/types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import GameIcon from "../../ui/display/GameIcon.tsx";
import VictoryPointIcon from "../../ui/display/VictoryPointIcon.tsx";
import CardTagList from "../../ui/cards/CardTagList.tsx";
import ResourceProductionGrid from "./ResourceProductionGrid.tsx";
import type { TableauPlayer } from "./tableauPlayers.ts";

function SectionLabel({ children }: { children: ReactNode }) {
  return (
    <div className="font-orbitron text-[11px] font-bold uppercase tracking-wider text-white/50 mb-1.5">
      {children}
    </div>
  );
}

interface TableauOverviewProps {
  player: TableauPlayer;
  totalVP: number;
  onCardTap: (card: CardDto) => void;
}

export default function TableauOverview({ player, totalVP, onCardTap }: TableauOverviewProps) {
  const corporation = player.corporation;
  const preludes = (player.playedCards ?? []).filter((card) => card.type === CardTypePrelude);

  return (
    <div className="h-full flex gap-4 p-3">
      <div className="w-[200px] h-full shrink-0 flex flex-col gap-1 overflow-y-auto overscroll-contain">
        {corporation ? (
          <button
            type="button"
            aria-label={`Show ${corporation.name}`}
            className="shrink-0 flex items-center justify-center h-[72px] cursor-pointer"
            onClick={() => onCardTap(corporation)}
          >
            {getCorporationLogo(corporation.name, "w-[180px] h-[72px]", "180px") ?? (
              <span className="font-orbitron text-[13px] font-bold uppercase tracking-wider text-white/80">
                {corporation.name}
              </span>
            )}
          </button>
        ) : (
          <div className="shrink-0 h-[72px] flex items-center justify-center border border-dashed border-white/20 text-[13px] text-white/50">
            No corporation
          </div>
        )}
        {preludes.map((prelude) => (
          <button
            key={prelude.id}
            type="button"
            aria-label={`Show ${prelude.name}`}
            className="shrink-0 min-h-11 flex items-center gap-2 px-2 border-t border-white/10 text-left cursor-pointer"
            onClick={() => onCardTap(prelude)}
          >
            <span className="flex-1 min-w-0 truncate font-orbitron text-[12px] text-white/90">
              {prelude.name}
            </span>
            <CardTagList card={prelude} size="sm" className="shrink-0 flex items-center gap-0.5" />
          </button>
        ))}
      </div>

      <div className="flex-1 min-w-0 overflow-y-auto overscroll-contain flex flex-col gap-3 pb-2">
        <div className="flex items-center gap-5 h-9 shrink-0">
          <span className="flex items-center gap-1.5" aria-label={`TR ${player.terraformRating}`}>
            <GameIcon iconType={ResourceTypeTR} size="small" />
            <span className="font-orbitron text-[16px] font-bold leading-none tabular-nums text-white">
              {player.terraformRating}
            </span>
          </span>
          <span aria-label={`${totalVP} victory points`}>
            <VictoryPointIcon value={String(totalVP)} bare />
          </span>
        </div>

        <div>
          <SectionLabel>Resources and production</SectionLabel>
          <ResourceProductionGrid player={player} size="regular" />
        </div>
      </div>
    </div>
  );
}
