import type { GameDto } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import GameSetupControls from "./GameSetupControls.tsx";

export default function LobbyMapInfoPanel({ game, playerId }: { game: GameDto; playerId: string }) {
  return (
    <GameSetupControls
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
