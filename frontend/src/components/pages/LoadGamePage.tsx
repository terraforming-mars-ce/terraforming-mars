import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import GameButton from "../ui/buttons/GameButton.tsx";
import ResumeSeatFields from "../ui/ResumeSeatFields.tsx";
import SaveGameLayout, { SaveFeedback } from "../ui/SaveGameLayout.tsx";
import { apiService } from "@/services/apiService.ts";
import { saveGameSession } from "@/utils/sessionStorage.ts";
import type { GameSaveSummaryDto } from "@/types/generated/api-types.ts";

const MAX_SAVE_BYTES = 128 * 1024 * 1024;
const MAX_FILENAME_BYTES = 255;
const TOO_LARGE = "File too large (max 128 MiB)";

function exceedsBytes(text: string, limit: number): boolean {
  if (text.length > limit) return true;
  let size = 0;
  for (const character of text) {
    const point = character.codePointAt(0)!;
    size += point < 128 ? 1 : point < 2048 ? 2 : point < 65536 ? 3 : 4;
    if (size > limit) return true;
  }
  return false;
}

export default function LoadGamePage() {
  const navigate = useNavigate();
  const [raw, setRaw] = useState("");
  const [paste, setPaste] = useState(false);
  const [filename, setFilename] = useState("");
  const [summary, setSummary] = useState<GameSaveSummaryDto | null>(null);
  const [mapNames, setMapNames] = useState<Record<string, string>>({});
  const [seatId, setSeatId] = useState("");
  const [name, setName] = useState("");
  const [token, setToken] = useState("");
  const [status, setStatus] = useState<
    "idle" | "reading" | "validating" | "ready" | "importing" | "error"
  >("idle");
  const [error, setError] = useState("");
  const input = useRef<HTMLInputElement>(null);
  const request = useRef(0);
  const abort = useRef<AbortController | null>(null);
  const importing = status === "importing";
  const busy = status === "reading" || status === "validating" || importing;

  useEffect(() => {
    const controller = new AbortController();
    void apiService
      .getGameOptions(controller.signal)
      .then((options) => {
        if (!controller.signal.aborted) {
          setMapNames(Object.fromEntries(options.availableMaps.map((map) => [map.id, map.name])));
        }
      })
      .catch(() => {});
    return () => controller.abort();
  }, []);

  useEffect(
    () => () => {
      request.current++;
      abort.current?.abort();
    },
    [],
  );

  const invalidate = () => {
    abort.current?.abort();
    request.current++;
    setSummary(null);
    setSeatId("");
    setName("");
    setToken("");
    setError("");
    setStatus("idle");
    return request.current;
  };
  const reject = (message: string) => {
    setError(message);
    setStatus("error");
    setSummary(null);
    setFilename("");
  };
  const validate = async (text: string, sequence: number, selectedName = "") => {
    if (sequence !== request.current) return;
    if (exceedsBytes(text, MAX_SAVE_BYTES)) {
      reject(TOO_LARGE);
      return;
    }
    if (!text.trim()) {
      reject("Empty save");
      return;
    }
    setStatus("validating");
    abort.current = new AbortController();
    try {
      const result = await apiService.validateGameSave(text, abort.current.signal);
      if (sequence !== request.current) return;
      setRaw(text);
      setFilename(selectedName);
      setSummary(result);
      const seat = result.seats.find((s) => s.playerType === "human" && !s.exited);
      setSeatId(seat?.id ?? "");
      setName(seat?.name ?? "");
      setStatus("ready");
    } catch (err) {
      if (sequence === request.current)
        reject(err instanceof Error ? err.message : "Couldn’t check save. Try again.");
    }
  };
  const selectFile = async (file: File) => {
    const sequence = invalidate();
    setFilename("");
    setRaw("");
    if (file.size > MAX_SAVE_BYTES) {
      reject(TOO_LARGE);
      return;
    }
    if (file.size === 0) {
      reject("Empty file");
      return;
    }
    if (exceedsBytes(file.name, MAX_FILENAME_BYTES)) {
      reject("Filename too long (max 255 bytes)");
      return;
    }
    if (!file.name.trim() || /[\p{Cc}/\\\u202a-\u202e\u2066-\u2069]/u.test(file.name)) {
      reject("Invalid filename");
      return;
    }
    if (!file.name.toLowerCase().endsWith(".json")) {
      reject("Select a .json file");
      return;
    }
    setStatus("reading");
    try {
      await validate(await file.text(), sequence, file.name);
    } catch {
      if (sequence === request.current) reject("Couldn’t read file");
    }
  };
  const load = async () => {
    if (!summary || importing) return;
    setStatus("importing");
    setError("");
    try {
      const result = await apiService.importGameSave(raw, seatId, name.trim(), token.trim());
      saveGameSession({
        gameId: result.gameId,
        playerId: result.playerId,
        playerName: result.playerName,
      });
      navigate(`/resume/${result.gameId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t load game. Try again.");
      setStatus("ready");
    }
  };
  const seats = summary?.seats.filter((s) => s.playerType === "human" && !s.exited) ?? [];
  return (
    <SaveGameLayout title="Load game" expanded={paste || !!summary} onBack={() => navigate("/")}>
      <div
        className={
          summary
            ? "grid grid-cols-1 lg:grid-cols-2 gap-8 compact:gap-5"
            : `flex flex-col ${paste ? "compact:h-full compact:min-h-0" : ""}`
        }
      >
        <div
          className={`flex flex-col gap-5 min-w-0 compact:gap-3 ${paste && !summary ? "compact:flex-1 compact:min-h-0" : ""}`}
        >
          {!paste ? (
            <>
              <input
                ref={input}
                type="file"
                accept=".json,application/json"
                className="hidden"
                aria-label="Select an exported Game file"
                disabled={importing}
                onChange={(event) => {
                  const file = event.target.files?.[0];
                  event.target.value = "";
                  if (file) void selectFile(file);
                }}
              />
              <GameButton
                emphasis="secondary"
                disabled={importing}
                className="w-full !font-orbitron !whitespace-normal break-all !min-h-14"
                onClick={() => input.current?.click()}
              >
                {filename || "Select an exported Game file"}
              </GameButton>
            </>
          ) : (
            <>
              <textarea
                aria-label="Game save JSON"
                spellCheck={false}
                className={`game-input w-full h-[clamp(160px,30dvh,320px)] compact:h-40 text-base font-mono resize-y compact:resize-none ${!summary ? "compact:flex-1 compact:!h-0 compact:min-h-0" : ""}`}
                value={raw}
                disabled={importing}
                onPaste={(event) => {
                  const text = event.clipboardData.getData("text");
                  const target = event.currentTarget;
                  const retained =
                    raw.slice(0, target.selectionStart) + raw.slice(target.selectionEnd);
                  if (
                    text.length + retained.length > MAX_SAVE_BYTES ||
                    exceedsBytes(text + retained, MAX_SAVE_BYTES)
                  ) {
                    event.preventDefault();
                    invalidate();
                    reject(TOO_LARGE);
                  }
                }}
                onChange={(event) => {
                  invalidate();
                  if (exceedsBytes(event.target.value, MAX_SAVE_BYTES)) {
                    reject(TOO_LARGE);
                    return;
                  }
                  setRaw(event.target.value);
                }}
              />
            </>
          )}
          <div className="flex flex-col gap-5 compact:flex-row compact:gap-3 compact:shrink-0">
            {paste && (
              <GameButton
                emphasis="secondary"
                className="compact:flex-1 compact:min-h-11"
                disabled={!raw.trim() || busy}
                onClick={() => void validate(raw, invalidate())}
              >
                Check save
              </GameButton>
            )}
            <GameButton
              emphasis="quiet"
              className="compact:flex-1 compact:min-h-11"
              disabled={importing}
              onClick={() => {
                invalidate();
                setPaste(!paste);
                setFilename("");
                setRaw("");
              }}
            >
              {paste ? "Use file" : "Paste JSON"}
            </GameButton>
          </div>
          <SaveFeedback
            className="text-center"
            message={
              error ||
              (status === "reading"
                ? "Reading save…"
                : status === "validating"
                  ? "Checking save…"
                  : "")
            }
            error={!!error}
          />
        </div>
        {summary && (
          <div className="flex flex-col gap-5 min-w-0">
            <div className="rounded-lg border border-white/15 p-4 text-sm text-white/80">
              <p className="font-orbitron flex flex-wrap items-center gap-x-6 gap-y-1">
                <span>Generation {summary.generation}</span>
                <span>
                  {mapNames[summary.mapId] ??
                    summary.mapId
                      .replace(/[-_]+/g, " ")
                      .replace(/\b\w/g, (letter) => letter.toUpperCase())}
                </span>
              </p>
              <p className="mt-3 text-left text-xs text-white/45">
                {new Date(summary.savedAt).toLocaleString()}
              </p>
            </div>
            <ResumeSeatFields
              id="load-game-seat"
              seats={seats}
              seatId={seatId}
              name={name}
              disabled={importing}
              onSeatChange={(id) => {
                setSeatId(id);
                setName(seats.find((s) => s.id === id)?.name ?? "");
              }}
              onNameChange={setName}
            />
            {summary.seats.some((s) => s.playerType === "bot" && !s.exited) && (
              <label className="flex flex-col gap-2 text-sm">
                Claude OAuth token for saved bots
                <input
                  className="game-input text-base"
                  type="password"
                  autoComplete="off"
                  value={token}
                  onChange={(event) => setToken(event.target.value)}
                />
                <span className="text-white/60">
                  Credentials are not included in save files. You can also enter them in the resume
                  lobby.
                </span>
              </label>
            )}
            <GameButton
              disabled={!seatId || !name.trim() || importing}
              loading={importing}
              onClick={() => void load()}
            >
              Load game
            </GameButton>
          </div>
        )}
      </div>
    </SaveGameLayout>
  );
}
