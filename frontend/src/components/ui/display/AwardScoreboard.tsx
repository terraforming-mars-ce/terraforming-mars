import React, { useMemo } from "react";

export interface ScoreboardPlayer {
  id: string;
  name: string;
  color: string;
}

interface AwardScoreboardProps {
  players: ScoreboardPlayer[];
  playerProgress: { [playerId: string]: number };
  className?: string;
}

const AwardScoreboard: React.FC<AwardScoreboardProps> = ({
  players,
  playerProgress,
  className = "",
}) => {
  const sortedPlayers = useMemo(
    () => [...players].sort((a, b) => (playerProgress[b.id] ?? 0) - (playerProgress[a.id] ?? 0)),
    [players, playerProgress],
  );

  const longestNameLength = useMemo(
    () => players.reduce((max, p) => Math.max(max, p.name.length), 0),
    [players],
  );

  return (
    <div
      className={`flex-shrink-0 grid grid-cols-[auto_auto] gap-x-3 gap-y-1 ${className}`}
      style={{ minWidth: `${longestNameLength + 5}ch` }}
    >
      {sortedPlayers.map((player) => {
        const score = playerProgress[player.id] ?? 0;
        return (
          <React.Fragment key={player.id}>
            <div className="flex items-center gap-2 text-sm">
              <span
                className="w-2 h-2 rounded-full flex-shrink-0"
                style={{ backgroundColor: player.color }}
              />
              <span className="text-white/80">{player.name}</span>
            </div>
            <span className="text-sm font-orbitron font-semibold text-white/50">{score}</span>
          </React.Fragment>
        );
      })}
    </div>
  );
};

export default AwardScoreboard;
