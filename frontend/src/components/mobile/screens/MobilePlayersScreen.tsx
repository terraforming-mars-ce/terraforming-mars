import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useMobileUiStore } from "@/stores/mobileUiStore.ts";
import { useConfirmDialogStore } from "@/stores/confirmDialogStore.ts";
import { isPlayerActionPhase } from "@/utils/actionUtils.ts";
import type { PlayerHostAction } from "@/hooks/usePlayerHostActions.ts";
import MobileScreen from "../MobileScreen.tsx";
import { PlayersIcon } from "../screenIcons.tsx";
import { useMobileGame } from "../MobileGameContext.tsx";
import PlayerRow from "../tableau/PlayerRow.tsx";
import PlayerActionSheet from "../tableau/PlayerActionSheet.tsx";
import { playerColor, playersInTurnOrder, type TableauPlayer } from "../tableau/tableauPlayers.ts";

const HIGHLIGHT_DURATION_MS = 1600;

interface ActionSheetState {
  player: TableauPlayer;
  actions: PlayerHostAction[];
}

export default function MobilePlayersScreen() {
  const close = useMobileUiStore((s) => s.close);
  const focusedPlayerId = useMobileUiStore((s) => s.focusedPlayerId);
  const { gameState, currentPlayer, isSpectator, isReplay, playerColorMap } = useMobileGame();
  const [highlightedId, setHighlightedId] = useState<string | null>(focusedPlayerId);
  const [actionSheet, setActionSheet] = useState<ActionSheetState | null>(null);
  const rowRefs = useRef(new Map<string, HTMLDivElement>());

  const players = useMemo(() => playersInTurnOrder(gameState), [gameState]);
  const selfId = isSpectator ? null : (currentPlayer?.id ?? null);
  const isHost = !isSpectator && !isReplay && !!selfId && selfId === gameState.hostPlayerId;
  const isActionPhase = isPlayerActionPhase(gameState.currentPhase);
  const turnPlayerId = gameState.currentTurn ?? "";

  useEffect(() => {
    if (!focusedPlayerId) {
      return undefined;
    }
    rowRefs.current.get(focusedPlayerId)?.scrollIntoView({ block: "nearest" });
    setHighlightedId(focusedPlayerId);
    const timer = setTimeout(() => setHighlightedId(null), HIGHLIGHT_DURATION_MS);
    return () => clearTimeout(timer);
  }, [focusedPlayerId]);

  const findName = useCallback(
    (playerId: string) => players.find((p) => p.id === playerId)?.name || "Unknown",
    [players],
  );

  const handleKickPlayer = useCallback(
    (playerId: string) => {
      useConfirmDialogStore
        .getState()
        .request({ kind: "kick", playerId, playerName: findName(playerId) });
    },
    [findName],
  );

  const handleConvertToBot = useCallback(
    (playerId: string) => {
      useConfirmDialogStore
        .getState()
        .request({ kind: "convertToBot", playerId, playerName: findName(playerId) });
    },
    [findName],
  );

  const handleOpenTableau = useCallback((playerId: string) => {
    useMobileUiStore.getState().open("tableau", { tableauPlayerId: playerId });
  }, []);

  const handleOpenActions = useCallback((player: TableauPlayer, actions: PlayerHostAction[]) => {
    setActionSheet({ player, actions });
  }, []);

  return (
    <MobileScreen icon={PlayersIcon} title="Players" onClose={close}>
      <div className="flex flex-col">
        {players.map((player) => (
          <PlayerRow
            key={player.id}
            ref={(element) => {
              if (element) {
                rowRefs.current.set(player.id, element);
              } else {
                rowRefs.current.delete(player.id);
              }
            }}
            player={player}
            color={playerColor(player, playerColorMap)}
            isMe={player.id === selfId}
            isTurn={player.id === turnPlayerId}
            isActionPhase={isActionPhase}
            isHost={isHost}
            highlighted={player.id === highlightedId}
            onOpenTableau={handleOpenTableau}
            onOpenActions={handleOpenActions}
            onKickPlayer={handleKickPlayer}
            onConvertToBot={
              gameState.settings?.hasClaudeOAuthToken ? handleConvertToBot : undefined
            }
          />
        ))}
      </div>
      {actionSheet && (
        <PlayerActionSheet
          playerName={actionSheet.player.name}
          actions={actionSheet.actions}
          onClose={() => setActionSheet(null)}
        />
      )}
    </MobileScreen>
  );
}
