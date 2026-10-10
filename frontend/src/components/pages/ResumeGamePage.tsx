import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import GameButton from "../ui/buttons/GameButton.tsx";
import CopyLinkButton from "../ui/buttons/CopyLinkButton.tsx";
import GameSelect from "../ui/GameSelect.tsx";
import ResumeSeatFields from "../ui/ResumeSeatFields.tsx";
import { PlayerChip } from "../ui/display/BotChips.tsx";
import SaveGameLayout, { SaveFeedback } from "../ui/SaveGameLayout.tsx";
import { apiService, gameSaveErrorMessage } from "@/services/apiService.ts";
import { globalWebSocketManager as socket } from "@/services/globalWebSocketManager.ts";
import { clearGameSession, getGameSession, saveGameSession } from "@/utils/sessionStorage.ts";
import { serverLink } from "@/utils/gateway.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import {
  MessageTypeWatchResumeGame,
  MessageTypeClaimResumeSeat,
  MessageTypeReleaseResumeSeat,
  MessageTypeResumeGame,
  MessageTypeResumeBotToken,
  type GameDto,
  type MessageType,
  type ResumeGameRequest,
  type ResumeSeatDto,
} from "@/types/generated/api-types.ts";

function seatStatus(seat: ResumeSeatDto): string {
  if (seat.exited) {
    return "Exited";
  }
  if (seat.playerType === "bot") {
    if (seat.botError || seat.botStatus === "failed") {
      return "Bot connection failed";
    }
    return `Bot · ${seat.botStatus || "waiting"}`;
  }
  if (seat.connected && seat.claimed) {
    return "Joined";
  }
  return seat.claimed ? "Reconnecting" : "Available";
}

export default function ResumeGamePage() {
  const { gameId = "" } = useParams();
  const navigate = useNavigate();
  const [game, setGame] = useState<GameDto | null>(null);
  const [ownId, setOwnId] = useState("");
  const [seatId, setSeatId] = useState("");
  const [name, setName] = useState("");
  const [token, setToken] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  const [manage, setManage] = useState(false);
  const [releaseId, setReleaseId] = useState("");
  const own = useRef("");
  const selection = useRef("");
  const pending = useRef<{ type: MessageType; seatId?: string } | null>(null);

  useEffect(() => {
    let disposed = false;
    let connecting = false;
    let initializedSelection = false;
    const fail = (err: unknown) => {
      if (disposed) {
        return;
      }
      const message = gameSaveErrorMessage(err, "Connection failed. Try again.");
      setError(message);
      pending.current = null;
      setBusy(false);
      if (typeof err === "object" && err !== null && "code" in err && err.code === "seat_taken") {
        selection.current = "";
        setSeatId("");
        setName("");
        void socket.resumeCommand(MessageTypeWatchResumeGame, { gameId }).catch(() => {});
      }
    };
    const updated = (next: GameDto) => {
      if (disposed || next.id !== gameId) {
        return;
      }
      setGame(next);
      setConnected(true);
      if (!next.resumeLobby) {
        const session = getGameSession();
        if (session?.gameId === gameId && session.playerId && !session.isSpectator) {
          navigate("/game", {
            replace: true,
            state: {
              game: next,
              playerId: session.playerId,
              playerName: session.playerName,
              isReconnection: true,
            },
          });
        } else {
          navigate(`/game/${gameId}`, { replace: true });
        }
        return;
      }
      const seats = next.resumeLobby.seats;
      const id = next.viewingPlayerId ?? "";
      const seat = seats.find((s) => s.id === id && s.claimed);
      const request = pending.current;
      if (
        request &&
        ((request.type === MessageTypeClaimResumeSeat && seat?.id === request.seatId) ||
          (request.type === MessageTypeReleaseResumeSeat &&
            !seats.find((s) => s.id === request.seatId)?.claimed) ||
          request.type === MessageTypeResumeBotToken)
      ) {
        pending.current = null;
        setBusy(false);
        setError("");
        if (request.type === MessageTypeReleaseResumeSeat) {
          setManage(false);
        }
      }
      if (seat) {
        own.current = id;
        setOwnId(id);
        socket.setCurrentPlayerId(id);
        saveGameSession({ gameId, playerId: id, playerName: seat.name });
      } else {
        if (own.current) {
          clearGameSession();
          selection.current = "";
          setSeatId("");
          setName("");
          setManage(false);
          setError("Seat released");
          pending.current = null;
          setBusy(false);
        }
        own.current = "";
        setOwnId("");
        socket.setCurrentPlayerId("");
        const available = seats.filter((s) => !s.claimed && !s.exited && s.playerType === "human");
        if (!initializedSelection) {
          const first = available[0];
          selection.current = first?.id ?? "";
          setSeatId(selection.current);
          setName(first?.savedName ?? "");
          initializedSelection = true;
        } else if (selection.current && !available.some((s) => s.id === selection.current)) {
          selection.current = "";
          setSeatId("");
          setName("");
          setError("Seat already taken");
        }
      }
    };
    const connect = async () => {
      if (connecting || disposed) {
        return;
      }
      connecting = true;
      try {
        const initial = await apiService.getGame(gameId);
        if (disposed) {
          return;
        }
        if (!initial) {
          setError("Game not found");
          return;
        }
        if (!initial.resumeLobby) {
          navigate(`/game/${gameId}`, { replace: true });
          return;
        }
        const session = getGameSession();
        const seat = initial.resumeLobby.seats.find((s) => s.id === session?.playerId);
        if (session?.gameId === gameId && !session.isSpectator && seat?.claimed) {
          own.current = seat.id;
          setOwnId(seat.id);
          setGame(initial);
          await socket.playerConnect(seat.name, gameId, seat.id);
        } else {
          if (session?.gameId === gameId) {
            clearGameSession();
          }
          await socket.resumeCommand(MessageTypeWatchResumeGame, { gameId });
        }
      } catch (err) {
        fail(err);
      } finally {
        connecting = false;
      }
    };
    const disconnected = () => {
      setConnected(false);
      pending.current = null;
      setBusy(false);
    };
    const reconnected = () => {
      void connect();
    };
    socket.on("game-updated", updated);
    socket.on("error", fail);
    socket.on("disconnect", disconnected);
    socket.on("connect", reconnected);
    void connect();
    return () => {
      disposed = true;
      socket.off("game-updated", updated);
      socket.off("error", fail);
      socket.off("disconnect", disconnected);
      socket.off("connect", reconnected);
    };
  }, [gameId, navigate]);

  const command = async (type: MessageType, extra: Partial<ResumeGameRequest> = {}) => {
    if (pending.current) {
      return;
    }
    pending.current = { type, seatId: extra.seatId };
    setBusy(true);
    setError("");
    try {
      await socket.resumeCommand(type, { gameId, ...extra });
    } catch (err) {
      pending.current = null;
      setError(gameSaveErrorMessage(err, "Request failed. Try again."));
      setBusy(false);
    }
  };
  const seats = game?.resumeLobby?.seats ?? [];
  const isHost = !!ownId && ownId === game?.hostPlayerId;
  const joining = !!game && !ownId;
  const available = seats.filter((s) => !s.claimed && !s.exited && s.playerType === "human");
  const releasable = seats.filter(
    (s) => s.id !== ownId && s.claimed && !s.exited && s.playerType === "human",
  );
  const waiting = seats.filter(
    (s) =>
      !s.exited && (s.playerType === "bot" ? s.botStatus !== "ready" : !s.claimed || !s.connected),
  );
  const leave = () => {
    if (busy) {
      return;
    }
    socket.disconnect();
    navigate("/join");
  };
  const closeManage = () => {
    if (busy) {
      return;
    }
    setManage(false);
    setError("");
  };
  let title = "Resume game";
  if (joining) {
    title = "Join game";
  }
  if (manage) {
    title = "Manage seats";
  }
  let footer;
  if (joining) {
    footer = (
      <div className="flex justify-end gap-3">
        <GameButton
          emphasis="secondary"
          className="compact:min-h-11"
          disabled={busy}
          onClick={leave}
        >
          Cancel
        </GameButton>
        <GameButton
          className="compact:min-h-11"
          loading={busy}
          aria-label="Join"
          disabled={busy || !connected || !name.trim() || !available.some((s) => s.id === seatId)}
          onClick={() =>
            void command(MessageTypeClaimResumeSeat, { seatId, playerName: name.trim() })
          }
        >
          Join
        </GameButton>
      </div>
    );
  } else if (manage) {
    footer = (
      <div className="flex justify-end gap-3">
        <GameButton
          emphasis="secondary"
          className="compact:min-h-11"
          disabled={busy}
          onClick={closeManage}
        >
          Cancel
        </GameButton>
        <GameButton
          className="compact:min-h-11"
          loading={busy}
          aria-label="Release seat"
          disabled={busy || !connected || !releasable.some((s) => s.id === releaseId)}
          onClick={() => void command(MessageTypeReleaseResumeSeat, { seatId: releaseId })}
        >
          Release seat
        </GameButton>
      </div>
    );
  } else if (ownId) {
    footer = (
      <div className="flex flex-wrap items-center justify-end gap-3 compact:gap-2">
        {isHost && releasable.length > 0 && (
          <GameButton
            emphasis="quiet"
            size="sm"
            className="mr-auto compact:min-h-11 compact:!px-2"
            disabled={busy || !connected}
            onClick={() => {
              setReleaseId(releasable[0].id);
              setError("");
              setManage(true);
            }}
          >
            Manage seats
          </GameButton>
        )}
        <CopyLinkButton
          textToCopy={serverLink(`/resume/${gameId}`)}
          defaultText="Copy invite link"
          className="compact:min-h-11 compact:!px-2"
        />
        {isHost ? (
          <GameButton
            className="compact:min-h-11 compact:!px-2"
            disabled={busy || !connected || waiting.length > 0}
            onClick={() => void command(MessageTypeResumeGame)}
          >
            Resume game
          </GameButton>
        ) : (
          <p className="text-sm text-white/65">Waiting for the host to resume</p>
        )}
      </div>
    );
  }
  return (
    <SaveGameLayout
      expanded={!!ownId && !manage}
      title={title}
      footer={footer}
      onBack={manage ? closeManage : leave}
    >
      <div className="flex flex-col gap-5 text-left text-white compact:gap-3">
        {joining && (
          <>
            <ResumeSeatFields
              id="resume-game-seat"
              seats={available.map((s) => ({ id: s.id, name: s.savedName }))}
              seatId={seatId}
              name={name}
              disabled={busy || !connected}
              onSeatChange={(id) => {
                selection.current = id;
                setSeatId(id);
                setName(available.find((s) => s.id === id)?.savedName ?? "");
                setError("");
              }}
              onNameChange={setName}
            />
            {available.length === 0 && (
              <p role="status" className="text-center text-sm text-white/65">
                No seats available
              </p>
            )}
          </>
        )}
        {manage && (
          <>
            <p className="text-sm text-white/65">
              Release a claimed seat so someone can join again.
            </p>
            <GameSelect
              id="release-seat"
              label="Seat to release"
              className="w-full"
              value={releaseId}
              options={releasable}
              disabled={busy || !connected || releasable.length === 0}
              onChange={setReleaseId}
            />
            {releasable.length === 0 && (
              <p className="text-sm text-white/65">No seats to release</p>
            )}
          </>
        )}
        {ownId && !manage && (
          <div className="flex flex-col gap-2">
            {seats.map((seat) => (
              <div
                key={seat.id}
                data-testid="resume-player"
                className="flex items-center gap-3 rounded-lg border border-white/15 bg-black/20 p-3 min-h-16 compact:gap-2"
              >
                <span
                  className="h-3 w-3 rounded-full shrink-0"
                  style={{ background: seat.color }}
                />
                <span className="font-orbitron text-sm min-w-0 truncate">{seat.name}</span>
                {seat.id === ownId && (
                  <PlayerChip className="shrink-0 bg-[rgba(60,100,150,0.8)] text-white border border-[rgba(80,130,180,0.7)] !text-[11px]">
                    YOU
                  </PlayerChip>
                )}
                {seat.corporationId && (
                  <div className="shrink-0 w-24 compact:w-20">
                    {getCorporationLogo(seat.corporationName, "w-full h-10", "96px")}
                  </div>
                )}
                <span
                  className={`ml-auto shrink-0 text-sm compact:text-xs ${seat.botError ? "text-red-300" : "text-white/65"}`}
                >
                  {seatStatus(seat)}
                </span>
              </div>
            ))}
          </div>
        )}
        {isHost && !manage && waiting.some((s) => s.playerType === "bot") && (
          <div className="flex flex-col gap-3 rounded-lg border border-white/15 p-3">
            <label className="flex flex-col gap-2 text-sm font-orbitron">
              Claude OAuth token for bots
              <input
                className="game-input text-base"
                type="password"
                autoComplete="off"
                value={token}
                onChange={(event) => setToken(event.target.value)}
              />
            </label>
            <GameButton
              emphasis="secondary"
              disabled={busy || !connected || !token.trim()}
              onClick={() => {
                void command(MessageTypeResumeBotToken, { botToken: token.trim() });
                setToken("");
              }}
            >
              Reconnect bots
            </GameButton>
          </div>
        )}
        {!connected && !error && (
          <p role="status" className="text-sm text-white/60">
            Connecting to server…
          </p>
        )}
        <SaveFeedback message={error} error className="text-center" />
      </div>
    </SaveGameLayout>
  );
}
