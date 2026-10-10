export interface PerformanceSnapshot {
  timestamp: number;
  fps: number;
  frameTimeMs: number;
  jsHeapUsedMB: number;
  jsHeapTotalMB: number;
  drawCalls: number;
  triangles: number;
  textureCount: number;
  geometryCount: number;
}

export interface SceneGroupStats {
  group: string;
  objects: number;
  drawCalls: number;
  triangles: number;
}

export type StatSection = Record<string, number | string>;

export interface GpuStats {
  drawCalls: number;
  triangles: number;
  textureCount: number;
  geometryCount: number;
}

const SAMPLE_INTERVAL_MS = 250;
const DETAIL_INTERVAL_MS = 500;
const MAX_SAMPLES = 120; // ~30s at 4Hz

const hasMemoryApi =
  typeof performance !== "undefined" &&
  "memory" in performance &&
  (performance as any).memory != null;

class PerformanceStoreService {
  private static instance: PerformanceStoreService;

  private buffer: PerformanceSnapshot[] = [];
  private listeners: Set<(snapshots: PerformanceSnapshot[]) => void> = new Set();

  private rafId: number | null = null;
  private frameCount = 0;
  private lastSampleTime = 0;
  private lastFrameTimestamp = 0;
  private latestFrameTimeMs = 0;

  private latestGpu: GpuStats = {
    drawCalls: 0,
    triangles: 0,
    textureCount: 0,
    geometryCount: 0,
  };

  private refCount = 0;
  private census: SceneGroupStats[] = [];
  private sections = new Map<string, StatSection>();
  private lastDetail = 0;
  private sectionTimes = new Map<string, number>();

  static getInstance(): PerformanceStoreService {
    if (!PerformanceStoreService.instance) {
      PerformanceStoreService.instance = new PerformanceStoreService();
    }
    return PerformanceStoreService.instance;
  }

  start() {
    this.refCount++;
    if (this.rafId !== null) return;

    this.frameCount = 0;
    this.lastSampleTime = performance.now();
    this.lastFrameTimestamp = performance.now();
    this.tick();
  }

  stop() {
    this.refCount = Math.max(0, this.refCount - 1);
    if (this.refCount > 0) return;

    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
  }

  updateGpuStats(stats: GpuStats) {
    this.latestGpu = stats;
  }

  // Detailed stats are only gathered while someone is watching, at a low rate.
  get watching() {
    return this.refCount > 0;
  }

  detailDue(now: number) {
    if (!this.watching || now - this.lastDetail < DETAIL_INTERVAL_MS) {
      return false;
    }
    this.lastDetail = now;
    return true;
  }

  // True at most twice a second per section, and only while the window is open.
  sectionDue(name: string) {
    if (!this.watching) {
      return false;
    }
    const now = performance.now();
    if (now - (this.sectionTimes.get(name) ?? 0) < DETAIL_INTERVAL_MS) {
      return false;
    }
    this.sectionTimes.set(name, now);
    return true;
  }

  updateCensus(census: SceneGroupStats[]) {
    this.census = census;
  }

  setSection(name: string, values: StatSection) {
    this.sections.set(name, values);
  }

  getCensus() {
    return this.census;
  }

  // Scene groups hidden from the performance window for A/B measurements.
  readonly hiddenGroups = new Set<string>();

  toggleGroup(group: string) {
    if (!this.hiddenGroups.delete(group)) {
      this.hiddenGroups.add(group);
    }
  }

  getSections() {
    return this.sections;
  }

  subscribe(listener: (snapshots: PerformanceSnapshot[]) => void): () => void {
    this.listeners.add(listener);
    listener([...this.buffer]);
    return () => {
      this.listeners.delete(listener);
    };
  }

  getSnapshots(): PerformanceSnapshot[] {
    return [...this.buffer];
  }

  private tick = () => {
    const now = performance.now();

    // Track per-frame time
    if (this.lastFrameTimestamp > 0) {
      this.latestFrameTimeMs = now - this.lastFrameTimestamp;
    }
    this.lastFrameTimestamp = now;
    this.frameCount++;

    // Sample at fixed interval
    const elapsed = now - this.lastSampleTime;
    if (elapsed >= SAMPLE_INTERVAL_MS) {
      const fps = (this.frameCount / elapsed) * 1000;

      let jsHeapUsedMB = 0;
      let jsHeapTotalMB = 0;
      if (hasMemoryApi) {
        const mem = (performance as any).memory;
        jsHeapUsedMB = mem.usedJSHeapSize / 1048576;
        jsHeapTotalMB = mem.totalJSHeapSize / 1048576;
      }

      const snapshot: PerformanceSnapshot = {
        timestamp: now,
        fps: Math.round(fps * 10) / 10,
        frameTimeMs: Math.round(this.latestFrameTimeMs * 100) / 100,
        jsHeapUsedMB: Math.round(jsHeapUsedMB * 10) / 10,
        jsHeapTotalMB: Math.round(jsHeapTotalMB * 10) / 10,
        drawCalls: this.latestGpu.drawCalls,
        triangles: this.latestGpu.triangles,
        textureCount: this.latestGpu.textureCount,
        geometryCount: this.latestGpu.geometryCount,
      };

      this.buffer.push(snapshot);
      if (this.buffer.length > MAX_SAMPLES) {
        this.buffer.shift();
      }

      this.frameCount = 0;
      this.lastSampleTime = now;

      this.notifyListeners();
    }

    this.rafId = requestAnimationFrame(this.tick);
  };

  private notifyListeners() {
    const copy = [...this.buffer];
    this.listeners.forEach((listener) => listener(copy));
  }
}

export const performanceStore = PerformanceStoreService.getInstance();

type TraceDetail = Record<string, unknown>;
const TRACE_PREFIX = "openmars:cold-start";
const NOOP = () => {};

function traceEnabled() {
  try {
    return localStorage.getItem("openmars.perfTrace") === "1";
  } catch {
    return false;
  }
}

// Opt in before reloading: localStorage.setItem("openmars.perfTrace", "1").
class ColdStartTrace {
  readonly enabled = traceEnabled();
  private sequence = 0;
  private runs = new Map<string, number>();
  private current: {
    id: string;
    start: number;
    rows: { stage: string; offsetMs: number; durationMs: number; detail: TraceDetail }[];
    maxFrameGapMs: number;
    frameCount: number;
    resources: PerformanceResourceTiming[];
  } | null = null;
  private frame = 0;
  private timer: ReturnType<typeof setTimeout> | undefined;
  private observer: PerformanceObserver | undefined;
  private entries: string[] = [];

  get active() {
    return this.current !== null;
  }

  begin(name: string, detail: TraceDetail = {}) {
    if (!this.enabled) {
      return;
    }
    this.finish();
    for (const entry of this.entries) {
      performance.clearMarks(entry);
      performance.clearMeasures(entry);
    }
    this.entries = [];
    const run = (this.runs.get(name) ?? 0) + 1;
    this.runs.set(name, run);
    this.current = {
      id: `${TRACE_PREFIX}:${name}:${++this.sequence}`,
      start: performance.now(),
      rows: [],
      maxFrameGapMs: 0,
      frameCount: 0,
      resources: [],
    };
    this.mark("requested", { ...detail, run, visibility: document.visibilityState });
    let previous = this.current.start;
    const tick = () => {
      const trace = this.current;
      if (!trace) {
        return;
      }
      const now = performance.now();
      const gap = now - previous;
      trace.maxFrameGapMs = Math.max(trace.maxFrameGapMs, gap);
      trace.frameCount++;
      if (trace.frameCount === 1 || gap >= 25) {
        this.record("browser-frame-gap", previous, gap, {
          frame: trace.frameCount,
          visibility: document.visibilityState,
        });
      }
      previous = now;
      this.frame = requestAnimationFrame(tick);
    };
    this.frame = requestAnimationFrame(tick);
    if (typeof PerformanceObserver !== "undefined") {
      const entryTypes = ["longtask", "resource"].filter((type) =>
        PerformanceObserver.supportedEntryTypes.includes(type),
      );
      if (entryTypes.length) {
        this.observer = new PerformanceObserver((list) => this.recordEntries(list.getEntries()));
        this.observer.observe({ entryTypes });
      }
    }
    this.timer = setTimeout(() => this.finish(), 4000);
  }

  mark(stage: string, detail: TraceDetail = {}) {
    if (!this.current) {
      return;
    }
    this.record(stage, performance.now(), 0, detail);
  }

  span(stage: string, detail: TraceDetail = {}) {
    const trace = this.current;
    if (!trace) {
      return NOOP;
    }
    const start = performance.now();
    return () => {
      if (this.current === trace) {
        this.record(stage, start, performance.now() - start, detail);
      }
    };
  }

  record(stage: string, start: number, duration: number, detail: TraceDetail = {}) {
    const trace = this.current;
    if (!trace || trace.rows.length >= 250) {
      return;
    }
    const name = `${trace.id}:${stage}:${trace.rows.length}`;
    if (duration === 0) {
      performance.mark(name, { startTime: start, detail });
    } else {
      performance.measure(name, { start, duration, detail });
    }
    this.entries.push(name);
    trace.rows.push({
      stage,
      offsetMs: Math.round((start - trace.start) * 100) / 100,
      durationMs: Math.round(duration * 100) / 100,
      detail,
    });
  }

  reactRender = (
    id: string,
    phase: string,
    actualDuration: number,
    _baseDuration: number,
    startTime: number,
    commitTime: number,
  ) => {
    this.record(`react:${id}`, startTime, actualDuration, { phase, commitTime });
  };

  private recordEntries(entries: PerformanceEntry[]) {
    for (const entry of entries) {
      if (this.current && entry.startTime + entry.duration >= this.current.start) {
        if (entry.entryType === "longtask") {
          this.record("long-task", entry.startTime, entry.duration);
        } else if (this.current.resources.length < 40) {
          this.current.resources.push(entry as PerformanceResourceTiming);
        }
      }
    }
  }

  private finish() {
    const trace = this.current;
    if (!trace) {
      return;
    }
    clearTimeout(this.timer);
    cancelAnimationFrame(this.frame);
    if (this.observer) {
      this.recordEntries(this.observer.takeRecords());
      this.observer.disconnect();
      this.observer = undefined;
    }
    const resources = trace.resources.map((entry) => ({
      url: entry.name,
      startMs: Math.round(entry.startTime - trace.start),
      durationMs: Math.round(entry.duration),
      transferBytes: entry.transferSize,
    }));
    this.current = null;
    // Log after the capture so DevTools console rendering cannot cause the measured stall.
    console.groupCollapsed(`[${trace.id}] max frame gap ${trace.maxFrameGapMs.toFixed(1)}ms`);
    console.table(trace.rows);
    console.log("Capture", {
      durationMs: performance.now() - trace.start,
      frames: trace.frameCount,
      maxFrameGapMs: trace.maxFrameGapMs,
      rowLimitReached: trace.rows.length === 250,
      longTasksSupported:
        typeof PerformanceObserver !== "undefined" &&
        PerformanceObserver.supportedEntryTypes.includes("longtask"),
    });
    if (resources.length) {
      console.table(resources);
    }
    console.log(
      "CPU submission timings are not GPU execution timings. Frame gaps include browser rendering; use a Performance recording for paint/raster/GPU attribution.",
    );
    console.groupEnd();
  }
}

export const coldStartTrace = new ColdStartTrace();
