import { FC, useState, useMemo, useRef, useEffect } from "react";
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from "recharts";
import type { GameHistoryEntryDto } from "../../../types/generated/api-types";
import { useHoverSound } from "../../../hooks/useHoverSound";

type GraphMode =
  | "terraforming"
  | "global-params"
  | "vp"
  | "tr"
  | "greeneries"
  | "cities"
  | "resources"
  | "credits"
  | "steel"
  | "titanium"
  | "plants"
  | "energy"
  | "heat"
  | "cards-played"
  | "resource-value";

const GRAPH_MODE_LABELS: Record<GraphMode, string> = {
  "global-params": "Global Parameters",
  terraforming: "Terraforming",
  vp: "VP",
  tr: "TR",
  greeneries: "Greeneries",
  cities: "Cities",
  resources: "Total Resources",
  credits: "Credits",
  steel: "Steel",
  titanium: "Titanium",
  plants: "Plants",
  energy: "Energy",
  heat: "Heat",
  "cards-played": "Cards Played",
  "resource-value": "Total Resource Value",
};

interface GameGraphsProps {
  entries: GameHistoryEntryDto[];
  playerColors: Record<string, string>;
  playerNames: Record<string, string>;
}

function countPlayerTiles(entry: GameHistoryEntryDto, playerId: string, tileType: string): number {
  return entry.board.tiles.filter((t) => t.occupiedBy?.type === tileType && t.ownerId === playerId)
    .length;
}

function countAllTiles(entry: GameHistoryEntryDto, tileType: string): number {
  return entry.board.tiles.filter((t) => t.occupiedBy?.type === tileType).length;
}

function countCityVP(entry: GameHistoryEntryDto, playerId: string): number {
  let vp = 0;
  for (const tile of entry.board.tiles) {
    if (tile.occupiedBy?.type === "city-tile" && tile.ownerId === playerId) {
      vp += countAdjacentGreeneries(
        entry,
        tile.coordinates.q,
        tile.coordinates.r,
        tile.coordinates.s,
      );
    }
  }
  return vp;
}

function countAdjacentGreeneries(
  entry: GameHistoryEntryDto,
  q: number,
  r: number,
  s: number,
): number {
  const directions = [
    [1, -1, 0],
    [1, 0, -1],
    [0, 1, -1],
    [-1, 1, 0],
    [-1, 0, 1],
    [0, -1, 1],
  ];
  let count = 0;
  for (const [dq, dr, ds] of directions) {
    const neighbor = entry.board.tiles.find(
      (t) => t.coordinates.q === q + dq && t.coordinates.r === r + dr && t.coordinates.s === s + ds,
    );
    if (neighbor?.occupiedBy?.type === "greenery-tile") {
      count++;
    }
  }
  return count;
}

function computeTotalVP(entry: GameHistoryEntryDto, playerId: string): number {
  const player = entry.players[playerId];
  if (!player) return 0;
  const greenery = countPlayerTiles(entry, playerId, "greenery-tile");
  const city = countCityVP(entry, playerId);
  const milestones = entry.milestones.filter((m) => m.playerId === playerId).length * 5;
  return player.terraformRating + greenery + city + milestones;
}

function totalResources(entry: GameHistoryEntryDto, playerId: string): number {
  const p = entry.players[playerId];
  if (!p) return 0;
  return p.credits + p.steel + p.titanium + p.plants + p.energy + p.heat;
}

function formatOffsetMs(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

const TERRAFORMING_LINES: { key: string; color: string; label: string }[] = [
  { key: "cities", color: "#94a3b8", label: "Cities" },
  { key: "greeneries", color: "#22c55e", label: "Greeneries" },
  { key: "oceans", color: "#3b82f6", label: "Oceans" },
];

const GLOBAL_PARAM_LINES: { key: string; color: string; label: string }[] = [
  { key: "temperature", color: "#ef4444", label: "Temperature" },
  { key: "oxygen", color: "#a3e635", label: "Oxygen" },
  { key: "oceans", color: "#3b82f6", label: "Oceans" },
  { key: "venus", color: "#f59e0b", label: "Venus" },
];

const TOOLTIP_STYLE: React.CSSProperties = {
  backgroundColor: "rgba(10,10,15,0.95)",
  border: "1px solid rgba(255,255,255,0.2)",
  borderRadius: 4,
  padding: "6px 10px",
};

interface CustomTooltipProps {
  active?: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  payload?: Array<{ name: string; value: number; color: string; payload?: any }>;
}

const CustomTooltip: FC<CustomTooltipProps> = ({ active, payload }) => {
  if (!active || !payload?.length) {
    return null;
  }

  const dataPoint = payload[0]?.payload;
  const generation = dataPoint?.generation;
  const offsetMs = dataPoint?.offsetMs;

  return (
    <div style={TOOLTIP_STYLE}>
      {generation !== undefined && (
        <div className="text-white/50 text-xs mb-1">
          Gen {generation}
          {offsetMs !== undefined && (
            <span className="ml-2 text-white/30">{formatOffsetMs(offsetMs)}</span>
          )}
        </div>
      )}
      {payload.map((entry, i) => (
        <div key={i} className="flex items-center gap-2 text-sm" style={{ minWidth: 100 }}>
          <span style={{ color: entry.color }}>{entry.name}</span>
          <span className="ml-auto text-white tabular-nums">{entry.value}</span>
        </div>
      ))}
    </div>
  );
};

const TICK_STYLE = { fill: "rgba(255,255,255,0.8)", fontSize: 13 };

const Y_AXIS_PROPS = {
  stroke: "rgba(255,255,255,0.3)",
  tick: TICK_STYLE,
  width: 40,
  domain: ["auto" as const, "auto" as const],
};

function GraphModeDropdown({
  mode,
  onModeChange,
  onMouseEnter,
}: {
  mode: GraphMode;
  onModeChange: (mode: GraphMode) => void;
  onMouseEnter?: () => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  return (
    <div ref={ref} className="relative" onMouseEnter={onMouseEnter}>
      <button
        onClick={() => setOpen(!open)}
        className="bg-black text-white text-[11px] px-3 py-1 border border-white/20 rounded font-orbitron font-bold cursor-pointer flex items-center gap-2"
      >
        {GRAPH_MODE_LABELS[mode]}
        <span className="text-[9px] text-white/50">▼</span>
      </button>
      {open && (
        <div className="absolute top-full right-0 mt-1 bg-black border border-white/20 rounded overflow-hidden z-20 min-w-[140px]">
          {Object.entries(GRAPH_MODE_LABELS).map(([value, label]) => (
            <button
              key={value}
              onClick={() => {
                onModeChange(value as GraphMode);
                setOpen(false);
              }}
              className={`block w-full text-left px-3 py-1.5 text-[11px] font-orbitron font-bold cursor-pointer ${
                mode === value
                  ? "bg-white/20 text-white"
                  : "text-white/60 hover:bg-white/10 hover:text-white"
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

type XAxisMode = "action" | "time";

const GameGraphs: FC<GameGraphsProps> = ({ entries, playerColors, playerNames }) => {
  const [mode, setMode] = useState<GraphMode>("global-params");
  const [xAxisMode, setXAxisMode] = useState<XAxisMode>("action");
  const hoverSound = useHoverSound();

  const playerIds = useMemo(() => Object.keys(playerNames), [playerNames]);

  const sampledEntries = useMemo(() => {
    if (entries.length <= 200) return entries;
    const step = Math.ceil(entries.length / 200);
    return entries.filter((_, i) => i % step === 0 || i === entries.length - 1);
  }, [entries]);

  const startTime = useMemo(() => {
    if (entries.length === 0) return 0;
    return new Date(entries[0].timestamp).getTime();
  }, [entries]);

  const generationTickIndices = useMemo(() => {
    const indices: number[] = [];
    let lastGen = -1;
    for (let i = 0; i < sampledEntries.length; i++) {
      if (sampledEntries[i].generation !== lastGen) {
        lastGen = sampledEntries[i].generation;
        indices.push(i);
      }
    }
    return indices;
  }, [sampledEntries]);

  const genOffsets = useMemo(() => {
    const map = new Map<number, string>();
    for (const e of sampledEntries) {
      if (!map.has(e.generation)) {
        const ms = new Date(e.timestamp).getTime() - startTime;
        map.set(e.generation, formatOffsetMs(ms));
      }
    }
    return map;
  }, [sampledEntries, startTime]);

  const renderXTick = useMemo(() => {
    return ({
      x,
      y,
      payload,
    }: {
      x: number | string;
      y: number | string;
      payload: { value: number };
    }) => {
      const idx = payload.value;
      const entry = sampledEntries[idx];
      if (!entry) return <g />;
      const gen = entry.generation;
      const offset = genOffsets.get(gen) ?? "";
      return (
        <g transform={`translate(${x},${y})`}>
          <text textAnchor="middle" fill="rgba(255,255,255,0.8)" fontSize={13} dy={14}>
            Gen {gen}
          </text>
          <text textAnchor="middle" fill="rgba(255,255,255,0.35)" fontSize={11} dy={30}>
            {offset}
          </text>
        </g>
      );
    };
  }, [sampledEntries, genOffsets]);

  const timeTickValues = useMemo(() => {
    if (sampledEntries.length === 0) return [];
    const maxMs =
      new Date(sampledEntries[sampledEntries.length - 1].timestamp).getTime() - startTime;
    if (maxMs <= 0) return [0];
    const intervals = [
      1000, 2000, 5000, 10000, 15000, 30000, 60000, 120000, 300000, 600000, 1800000, 3600000,
    ];
    let interval = intervals[intervals.length - 1];
    for (const iv of intervals) {
      const tickCount = maxMs / iv;
      if (tickCount >= 3 && tickCount <= 12) {
        interval = iv;
        break;
      }
    }
    const ticks: number[] = [];
    for (let t = 0; t <= maxMs; t += interval) {
      ticks.push(t);
    }
    if (ticks.length < 2) {
      ticks.push(maxMs);
    }
    return ticks;
  }, [sampledEntries, startTime]);

  const renderXTimeTick = useMemo(() => {
    return ({
      x,
      y,
      payload,
    }: {
      x: number | string;
      y: number | string;
      payload: { value: number };
    }) => {
      return (
        <g transform={`translate(${x},${y})`}>
          <text textAnchor="middle" fill="rgba(255,255,255,0.8)" fontSize={13} dy={14}>
            {formatOffsetMs(payload.value)}
          </text>
        </g>
      );
    };
  }, []);

  function addOffsetMs(e: GameHistoryEntryDto): number {
    return new Date(e.timestamp).getTime() - startTime;
  }

  const terraformingData: Record<string, number>[] = useMemo(
    () =>
      sampledEntries.map((e, i) => ({
        idx: i,
        generation: e.generation,
        offsetMs: addOffsetMs(e),
        cities: countAllTiles(e, "city-tile"),
        greeneries: countAllTiles(e, "greenery-tile"),
        oceans: e.oceans,
      })),
    [sampledEntries, startTime],
  );

  const globalParamsData: Record<string, number>[] = useMemo(
    () =>
      sampledEntries.map((e, i) => ({
        idx: i,
        generation: e.generation,
        offsetMs: addOffsetMs(e),
        temperature: e.temperature,
        oxygen: e.oxygen,
        oceans: e.oceans,
        venus: e.venus,
      })),
    [sampledEntries, startTime],
  );

  const multiPlayerData = useMemo(() => {
    const getValueFn = (m: GraphMode): ((e: GameHistoryEntryDto, pid: string) => number) => {
      switch (m) {
        case "vp":
          return (e, pid) => computeTotalVP(e, pid);
        case "tr":
          return (e, pid) => e.players[pid]?.terraformRating ?? 0;
        case "greeneries":
          return (e, pid) => countPlayerTiles(e, pid, "greenery-tile");
        case "cities":
          return (e, pid) => countPlayerTiles(e, pid, "city-tile");
        case "resources":
          return (e, pid) => totalResources(e, pid);
        case "credits":
          return (e, pid) => e.players[pid]?.credits ?? 0;
        case "steel":
          return (e, pid) => e.players[pid]?.steel ?? 0;
        case "titanium":
          return (e, pid) => e.players[pid]?.titanium ?? 0;
        case "plants":
          return (e, pid) => e.players[pid]?.plants ?? 0;
        case "energy":
          return (e, pid) => e.players[pid]?.energy ?? 0;
        case "heat":
          return (e, pid) => e.players[pid]?.heat ?? 0;
        case "cards-played":
          return (e, pid) => e.players[pid]?.playedCardCount ?? 0;
        case "resource-value":
          return (e, pid) => {
            const p = e.players[pid];
            if (!p) {
              return 0;
            }
            return p.credits + p.steel * 2 + p.titanium * 3 + p.plants * 2.87 + p.energy + p.heat;
          };
        default:
          return () => 0;
      }
    };
    const fn = getValueFn(mode);
    return sampledEntries.map((e, i) => {
      const point: Record<string, number> = {
        idx: i,
        generation: e.generation,
        offsetMs: addOffsetMs(e),
      };
      for (const pid of playerIds) {
        point[pid] = fn(e, pid);
      }
      return point;
    });
  }, [sampledEntries, playerIds, mode, startTime]);

  if (entries.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-white/40 text-sm">
        No history data available
      </div>
    );
  }

  const isFixedLines = mode === "terraforming" || mode === "global-params";

  const fixedLinesConfig =
    mode === "terraforming"
      ? TERRAFORMING_LINES
      : mode === "global-params"
        ? GLOBAL_PARAM_LINES
        : [];

  const fixedData =
    mode === "terraforming" ? terraformingData : mode === "global-params" ? globalParamsData : [];

  return (
    <div className="h-full relative">
      <div className="absolute top-6 right-8 z-10 flex items-center gap-3">
        <div className="flex rounded overflow-hidden border border-white/20">
          <button
            onClick={() => {
              hoverSound.onClick?.();
              setXAxisMode("action");
            }}
            onMouseEnter={hoverSound.onMouseEnter}
            className={`px-3 py-1 text-[11px] font-orbitron font-bold cursor-pointer ${
              xAxisMode === "action"
                ? "bg-white/20 text-white"
                : "bg-transparent text-white/40 hover:text-white/60"
            }`}
          >
            Action
          </button>
          <button
            onClick={() => {
              hoverSound.onClick?.();
              setXAxisMode("time");
            }}
            onMouseEnter={hoverSound.onMouseEnter}
            className={`px-3 py-1 text-[11px] font-orbitron font-bold cursor-pointer ${
              xAxisMode === "time"
                ? "bg-white/20 text-white"
                : "bg-transparent text-white/40 hover:text-white/60"
            }`}
          >
            Time
          </button>
        </div>
        <GraphModeDropdown
          mode={mode}
          onModeChange={(m) => {
            hoverSound.onClick?.();
            setMode(m);
          }}
          onMouseEnter={hoverSound.onMouseEnter}
        />
      </div>

      <style>{`.recharts-wrapper svg { outline: none; }`}</style>
      <div className="h-full">
        <ResponsiveContainer width="100%" height="100%">
          {isFixedLines ? (
            <LineChart data={fixedData}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              {xAxisMode === "time" ? (
                <XAxis
                  dataKey="offsetMs"
                  type="number"
                  ticks={timeTickValues}
                  tick={renderXTimeTick}
                  tickLine={false}
                  axisLine={false}
                  height={24}
                  domain={["dataMin", "dataMax"]}
                />
              ) : (
                <XAxis
                  dataKey="idx"
                  type="number"
                  ticks={generationTickIndices}
                  tick={renderXTick}
                  tickLine={false}
                  axisLine={false}
                  height={40}
                  domain={[0, sampledEntries.length - 1]}
                />
              )}
              <YAxis {...Y_AXIS_PROPS} />
              <Tooltip content={<CustomTooltip />} isAnimationActive={false} />
              {fixedLinesConfig.map(({ key, color, label }) => (
                <Line
                  key={key}
                  type="monotone"
                  dataKey={key}
                  stroke={color}
                  strokeWidth={1.5}
                  dot={false}
                  name={label}
                  animationDuration={300}
                />
              ))}
            </LineChart>
          ) : (
            <LineChart data={multiPlayerData}>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
              {xAxisMode === "time" ? (
                <XAxis
                  dataKey="offsetMs"
                  type="number"
                  ticks={timeTickValues}
                  tick={renderXTimeTick}
                  tickLine={false}
                  axisLine={false}
                  height={24}
                  domain={["dataMin", "dataMax"]}
                />
              ) : (
                <XAxis
                  dataKey="idx"
                  type="number"
                  ticks={generationTickIndices}
                  tick={renderXTick}
                  tickLine={false}
                  axisLine={false}
                  height={40}
                  domain={[0, sampledEntries.length - 1]}
                />
              )}
              <YAxis {...Y_AXIS_PROPS} />
              <Tooltip content={<CustomTooltip />} isAnimationActive={false} />
              {playerIds.map((pid) => (
                <Line
                  key={pid}
                  type="monotone"
                  dataKey={pid}
                  stroke={playerColors[pid] ?? "#ffffff"}
                  strokeWidth={1.5}
                  dot={false}
                  name={playerNames[pid]}
                  animationDuration={300}
                />
              ))}
            </LineChart>
          )}
        </ResponsiveContainer>
      </div>
    </div>
  );
};

export default GameGraphs;
