const RING = 4;
const WINDOW = 120;

interface TimerQueryExtension {
  TIME_ELAPSED_EXT: number;
  GPU_DISJOINT_EXT: number;
}

// Measures GPU time of the scene render with EXT_disjoint_timer_query_webgl2. Results arrive a few
// frames late, so queries rotate through a small ring and are read without stalling.
export class GpuTimer {
  private readonly ext: TimerQueryExtension | null;
  private readonly queries: WebGLQuery[] = [];
  private readonly pending = new Uint8Array(RING);
  private readonly samples = new Float32Array(WINDOW);
  private sampleCount = 0;
  private next = 0;
  private active = false;

  constructor(private readonly gl: WebGL2RenderingContext) {
    this.ext = gl.getExtension("EXT_disjoint_timer_query_webgl2") as TimerQueryExtension | null;
    if (this.ext) {
      for (let i = 0; i < RING; i++) {
        this.queries.push(gl.createQuery()!);
      }
    }
  }

  get available() {
    return this.ext !== null;
  }

  begin() {
    if (!this.ext) {
      return;
    }
    this.collect();
    if (this.pending[this.next]) {
      return;
    }
    this.gl.beginQuery(this.ext.TIME_ELAPSED_EXT, this.queries[this.next]);
    this.active = true;
  }

  end() {
    if (!this.ext || !this.active) {
      return;
    }
    this.gl.endQuery(this.ext.TIME_ELAPSED_EXT);
    this.pending[this.next] = 1;
    this.active = false;
    this.next = (this.next + 1) % RING;
  }

  // Average and worst GPU milliseconds over the recent window.
  stats() {
    const count = Math.min(this.sampleCount, WINDOW);
    let total = 0;
    let worst = 0;
    for (let i = 0; i < count; i++) {
      total += this.samples[i];
      worst = Math.max(worst, this.samples[i]);
    }
    return { average: count ? total / count : 0, worst, count };
  }

  dispose() {
    for (const query of this.queries) {
      this.gl.deleteQuery(query);
    }
    this.queries.length = 0;
  }

  private collect() {
    const gl = this.gl;
    const disjoint = gl.getParameter(this.ext!.GPU_DISJOINT_EXT) as boolean;
    for (let i = 0; i < RING; i++) {
      if (!this.pending[i] || !gl.getQueryParameter(this.queries[i], gl.QUERY_RESULT_AVAILABLE)) {
        continue;
      }
      this.pending[i] = 0;
      if (!disjoint) {
        const nanoseconds = gl.getQueryParameter(this.queries[i], gl.QUERY_RESULT) as number;
        this.samples[this.sampleCount % WINDOW] = nanoseconds / 1e6;
        this.sampleCount++;
      }
    }
  }
}
