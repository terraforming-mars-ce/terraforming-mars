import type { GameDto } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import GameSetupControls from "./GameSetupControls.tsx";

interface LobbyMapInfoPanelProps {
  game: GameDto;
  playerId: string;
  layout?: "stacked" | "split";
}

export default function LobbyMapInfoPanel({ game, playerId, layout }: LobbyMapInfoPanelProps) {
  return (
    <GameSetupControls
      layout={layout}
      maps={game.settings.availableMaps ?? []}
      mapId={game.settings.mapId}
      cardPacks={game.settings.cardPacks ?? []}
      venusNextEnabled={game.settings.venusNextEnabled}
      onChange={
        game.hostPlayerId === playerId
          ? (patch) => {
              void globalWebSocketManager.updateGameSettings(patch);
            }
          : undefined
      }
    />
  );
}
