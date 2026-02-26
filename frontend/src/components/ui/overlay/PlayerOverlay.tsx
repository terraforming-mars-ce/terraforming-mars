import React from "react";
import { PlayerDto, OtherPlayerDto } from "../../../types/generated/api-types.ts";
import { getPlayerColor } from "@/utils/playerColors.ts";

interface PlayerOverlayProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | OtherPlayerDto | null;
}

const PlayerOverlay: React.FC<PlayerOverlayProps> = ({ players, currentPlayer }) => {
  // Corporation logo mapping from available assets
  const corporationLogos: { [key: string]: string } = {
    polaris: "/assets/pathfinders/corp-logo-polaris.png",
    "mars-direct": "/assets/pathfinders/corp-logo-mars-direct.png",
    "habitat-marte": "/assets/pathfinders/corp-logo-habitat-marte.png",
    aurorai: "/assets/pathfinders/corp-logo-aurorai.png",
    "bio-sol": "/assets/pathfinders/corp-logo-bio-sol.png",
    chimera: "/assets/pathfinders/corp-logo-chimera.png",
    ambient: "/assets/pathfinders/corp-logo-ambient.png",
    odyssey: "/assets/pathfinders/corp-logo-odyssey.png",
    steelaris: "/assets/pathfinders/corp-logo-steelaris.png",
    soylent: "/assets/pathfinders/corp-logo-soylent.png",
    ringcom: "/assets/pathfinders/corp-logo-ringcom.png",
    "mind-set-mars": "/assets/pathfinders/corp-logo-mind-set-mars.png",
  };

  const getCorpLogo = (corporationId?: string) => {
    if (!corporationId) return "/assets/pathfinders/corp-logo-polaris.png"; // Default
    return corporationLogos[corporationId] || "/assets/pathfinders/corp-logo-polaris.png";
  };

  const playersToShow = players.length > 0 ? players : [];

  return (
    <div className="hidden absolute top-[70px] left-1/2 -translate-x-1/2 pointer-events-none">
      <div className="flex gap-2 items-center justify-center">
        {playersToShow.map((player, index) => {
          const isCurrentPlayer = player.id === currentPlayer?.id;
          const playerColor = getPlayerColor(index);
          const corpLogo = getCorpLogo(player.corporation?.id);
          const isPassed = player.passed || false;

          return (
            <div
              key={player.id || index}
              className={`
                relative border-2 rounded-xl py-2 px-4 min-w-[140px] backdrop-blur-space transition-all duration-300 pointer-events-auto cursor-pointer
                before:content-[''] before:absolute before:inset-0 before:rounded-[inherit] before:pointer-events-none
                ${
                  isCurrentPlayer
                    ? "bg-[linear-gradient(135deg,rgba(100,200,255,0.2)_0%,rgba(80,160,220,0.15)_50%,rgba(60,140,200,0.2)_100%)] border-[#64c8ff] shadow-[0_6px_25px_rgba(100,200,255,0.3),0_0_30px_#64c8ff,inset_0_1px_0_rgba(255,255,255,0.2)] before:bg-[linear-gradient(45deg,rgba(100,200,255,0.15)_0%,transparent_50%,rgba(100,200,255,0.08)_100%)]"
                    : isPassed
                      ? "opacity-60 [filter:grayscale(40%)] bg-[linear-gradient(135deg,rgba(80,80,80,0.9)_0%,rgba(60,60,60,0.85)_50%,rgba(40,40,40,0.9)_100%)] border-[rgba(150,150,150,0.5)] hover:opacity-80 hover:-translate-y-px"
                      : "bg-[linear-gradient(135deg,rgba(30,60,90,0.95)_0%,rgba(20,40,70,0.9)_50%,rgba(10,30,60,0.95)_100%)] before:bg-[linear-gradient(45deg,rgba(255,255,255,0.1)_0%,transparent_50%,rgba(255,255,255,0.05)_100%)] hover:-translate-y-0.5"
                }
              `}
              style={{
                borderColor: isCurrentPlayer ? "#64c8ff" : playerColor,
                boxShadow: isCurrentPlayer
                  ? "0 6px 25px rgba(100, 200, 255, 0.3), 0 0 30px #64c8ff, inset 0 1px 0 rgba(255, 255, 255, 0.2)"
                  : isPassed
                    ? undefined
                    : `0 4px 20px rgba(0, 0, 0, 0.4), 0 0 20px ${playerColor}`,
              }}
            >
              <div className="flex items-center gap-3 relative">
                <div className="flex-shrink-0">
                  <img
                    src={corpLogo}
                    alt={`${player.corporation?.name || "Unknown"} Corporation`}
                    className="w-8 h-8 rounded-md object-cover border border-white/20 transition-all duration-200 hover:border-white/40 hover:scale-105"
                  />
                </div>

                <div className="flex flex-col items-start flex-1">
                  <div className="text-xs font-semibold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] mb-0.5 whitespace-nowrap overflow-hidden text-ellipsis max-w-[80px]">
                    {player.name}
                  </div>
                  <div className="flex items-baseline gap-1">
                    <span className="text-[9px] font-medium text-[rgba(255,215,0,0.9)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                      TR
                    </span>
                    <span className="text-sm font-bold text-[#ffd700] [text-shadow:0_1px_2px_rgba(0,0,0,0.9),0_0_8px_rgba(255,215,0,0.4)] font-[Courier_New,monospace]">
                      {player.terraformRating}
                    </span>
                  </div>
                </div>

                {isPassed && (
                  <div className="absolute -top-1 -right-1.5 bg-[rgba(100,255,100,0.9)] text-black text-[8px] font-bold py-px px-1 rounded-sm tracking-wide border border-[rgba(50,200,50,1)]">
                    PASSED
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default PlayerOverlay;
