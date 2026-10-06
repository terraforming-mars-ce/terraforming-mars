import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React, { useEffect, useState } from "react";
import { globalWebSocketManager } from "../../../../services/globalWebSocketManager.ts";
import {
  AdminCommandTypeSetActionsRemaining,
  AdminCommandTypeSetCurrentTurn,
  GamePhaseAction,
  type AdminCommandRequest,
  type GameDto,
  type OtherPlayerDto,
  type PlayerDto,
  type SetActionsRemainingAdminCommand,
} from "../../../../types/generated/api-types.ts";

interface TurnCommandsPageProps {
  gameState: GameDto;
}

const labelStyle = {
  color: "#3b82f6",
  fontSize: "11px",
  fontWeight: "bold",
  display: "block",
  textAlign: "left" as const,
  marginBottom: "8px",
};

const buttonStyle = {
  padding: "4px 10px",
  background: "linear-gradient(135deg, rgba(59, 130, 246, 0.8), rgba(59, 130, 246, 0.6))",
  border: "1px solid rgba(59, 130, 246, 0.5)",
  borderRadius: "6px",
  color: "white",
  fontSize: "11px",
  cursor: "pointer" as const,
  fontWeight: "500" as const,
};

const disabledButtonStyle = {
  ...buttonStyle,
  background: "rgba(80, 80, 90, 0.4)",
  border: "1px solid rgba(80, 80, 90, 0.5)",
  color: "#777",
  cursor: "default" as const,
};

const tagStyle = (color: string) => ({
  fontSize: "9px",
  padding: "1px 5px",
  borderRadius: "3px",
  border: `1px solid ${color}`,
  color,
  letterSpacing: "0.5px",
});

const sendCommand = async (request: AdminCommandRequest) => {
  try {
    await globalWebSocketManager.sendAdminCommand(request);
  } catch (error) {
    console.error("Failed to send admin command:", error);
  }
};

const moveTurn = async (playerId: string) => {
  await sendCommand({ commandType: AdminCommandTypeSetCurrentTurn, payload: { playerId } });
};

const TurnCommandsPage: React.FC<TurnCommandsPageProps> = ({ gameState }) => {
  const byId = new Map<string, PlayerDto | OtherPlayerDto>();
  if (gameState.currentPlayer?.id) {
    byId.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  for (const p of gameState.otherPlayers ?? []) {
    byId.set(p.id, p);
  }
  const players = (gameState.turnOrder ?? [])
    .map((id) => byId.get(id))
    .filter((p): p is PlayerDto | OtherPlayerDto => p !== undefined);

  const currentTurnId = gameState.currentTurn ?? "";
  const currentTurnPlayer = byId.get(currentTurnId);
  const isActionPhase = gameState.currentPhase === GamePhaseAction;

  const [actions, setActions] = useState(String(currentTurnPlayer?.availableActions ?? 2));

  useEffect(() => {
    setActions(String(currentTurnPlayer?.availableActions ?? 2));
  }, [currentTurnId, currentTurnPlayer?.availableActions]);

  const handleSetActions = async () => {
    const value = parseInt(actions, 10);
    if (isNaN(value) || value < 1) {
      return;
    }
    const command: SetActionsRemainingAdminCommand = { actions: value };
    await sendCommand({ commandType: AdminCommandTypeSetActionsRemaining, payload: command });
  };

  return (
    <div>
      <label style={labelStyle}>Players</label>
      <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
        {players.map((p) => {
          const isCurrent = p.id === currentTurnId;
          const unavailable = p.passed || p.isExited;
          const canMove = isActionPhase && !unavailable;
          return (
            <div
              key={p.id}
              style={{ display: "flex", alignItems: "center", gap: "8px", minHeight: "26px" }}
            >
              <span style={{ width: "8px", height: "8px", background: p.color, flexShrink: 0 }} />
              <span
                style={{
                  color: "white",
                  fontSize: "12px",
                  minWidth: "90px",
                  textAlign: "left",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {p.name}
              </span>
              {isCurrent && <span style={tagStyle("#3b82f6")}>CURRENT</span>}
              {p.passed && <span style={tagStyle("#888")}>PASSED</span>}
              <div style={{ flex: 1 }} />
              {isCurrent ? (
                <>
                  <input
                    type="number"
                    min={1}
                    value={actions}
                    aria-label={`Actions remaining for ${p.name}`}
                    onChange={(e) => setActions(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        void handleSetActions();
                      }
                    }}
                    style={{
                      width: "50px",
                      padding: "3px 6px",
                      background: "rgba(0, 0, 0, 0.8)",
                      border: "1px solid rgba(59, 130, 246, 0.3)",
                      borderRadius: "4px",
                      color: "white",
                      fontSize: "12px",
                      outline: "none",
                      textAlign: "center",
                    }}
                  />
                  <GameButton
                    emphasis="quiet"
                    onClick={() => void handleSetActions()}
                    style={buttonStyle}
                  >
                    Set
                  </GameButton>
                </>
              ) : (
                <GameButton
                  emphasis="quiet"
                  disabled={!canMove}
                  onClick={() => void moveTurn(p.id)}
                  style={canMove ? buttonStyle : disabledButtonStyle}
                >
                  Move turn here
                </GameButton>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default TurnCommandsPage;
