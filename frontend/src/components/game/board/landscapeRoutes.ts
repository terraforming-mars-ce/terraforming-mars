import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";
import { boardCenter, polar, type Point2, type RoadPath } from "./landscapeGeometry";
import type { LandscapeTile, LandscapeConnection } from "./landscapeNetwork";

export interface LandscapeSpace {
  coordinate: HexCoordinate;
  blocked: boolean;
}
const keyOf = HexGrid2D.coordinateToKey;
const distance = (a: Point2, b: Point2) => Math.hypot(a.x - b.x, a.y - b.y);
const mix = (a: Point2, b: Point2, t: number): Point2 => ({
  x: a.x + (b.x - a.x) * t,
  y: a.y + (b.y - a.y) * t,
});
const add = (a: Point2, b: Point2, scale: number): Point2 => ({
  x: a.x + b.x * scale,
  y: a.y + b.y * scale,
});
const length = (points: Point2[]) =>
  points.slice(1).reduce((sum, p, i) => sum + distance(points[i], p), 0);
const hexDistance = (a: HexCoordinate, b: HexCoordinate) =>
  Math.max(Math.abs(a.q - b.q), Math.abs(a.r - b.r), Math.abs(a.s - b.s));

export function coordinateAt(p: Point2): HexCoordinate {
  const r = -p.y / (0.45 * Math.PI * 0.2);
  const q = p.x / (0.3 * Math.sqrt(3) * Math.PI * 0.2) - r / 2;
  const s = -q - r;
  let rq = Math.round(q),
    rr = Math.round(r),
    rs = Math.round(s);
  const dq = Math.abs(q - rq),
    dr = Math.abs(r - rr),
    ds = Math.abs(s - rs);
  if (dq > dr && dq > ds) {
    rq = -rr - rs;
  } else if (dr > ds) {
    rr = -rq - rs;
  } else {
    rs = -rq - rr;
  }
  return { q: rq, r: rr, s: rs };
}
interface Port {
  point: Point2;
  outward: Point2;
}
function ports(tile: LandscapeTile): Port[] {
  const center = boardCenter(tile.coordinate);
  return (tile.layout?.entrances ?? []).map((angle, index) => {
    const road = tile.layout!.roads.find((r) => r.exit === index)!;
    const gate = road.points.at(-1)!;
    const previous = road?.points.at(-2) ?? polar(0.1, angle);
    const span = distance(gate, previous);
    return {
      point: add(center, gate, 1),
      outward: { x: (gate.x - previous.x) / span, y: (gate.y - previous.y) / span },
    };
  });
}

function cubic(a: Point2, b: Point2, c: Point2, d: Point2): Point2[] {
  const count = Math.max(12, Math.ceil((distance(a, b) + distance(b, c) + distance(c, d)) / 0.003));
  return Array.from({ length: count + 1 }, (_, i) => {
    const t = i / count,
      u = 1 - t;
    return {
      x: u ** 3 * a.x + 3 * u * u * t * b.x + 3 * u * t * t * c.x + t ** 3 * d.x,
      y: u ** 3 * a.y + 3 * u * u * t * b.y + 3 * u * t * t * c.y + t ** 3 * d.y,
    };
  });
}
function sampleLine(a: Point2, b: Point2): Point2[] {
  return Array.from({ length: Math.max(1, Math.ceil(distance(a, b) / 0.004)) + 1 }, (_, i) =>
    mix(a, b, i / Math.max(1, Math.ceil(distance(a, b) / 0.004))),
  );
}

function routingSpace(tiles: LandscapeTile[], spaces: LandscapeSpace[], endpoints: Set<string>) {
  const domain = new Set(spaces.map((s) => keyOf(s.coordinate)));
  const blocked = spaces.filter((s) => s.blocked).map((s) => boardCenter(s.coordinate));
  const cities = tiles
    .filter((t) => t.kind === "city")
    .map((t) => ({
      ...boardCenter(t.coordinate),
      radius: endpoints.has(keyOf(t.coordinate)) ? 0.118 : 0.14,
    }));
  const special = tiles
    .filter((t) => t.kind === "special-greenery")
    .map((t) => boardCenter(t.coordinate));
  const obstacles = [...blocked, ...special];
  const forests = tiles.filter((t) => t.kind === "greenery").map((t) => boardCenter(t.coordinate));
  return (p: Point2, clearance = 0.008) => {
    if (domain.size && !domain.has(keyOf(coordinateAt(p)))) {
      return false;
    }
    if (cities.some((c) => distance(c, p) < c.radius + clearance * 0.3)) {
      return false;
    }
    for (const c of forests) {
      const x = p.x - c.x,
        y = p.y - c.y;
      const side = Math.max(
        Math.abs(x),
        Math.abs(0.5 * x + 0.866025404 * y),
        Math.abs(-0.5 * x + 0.866025404 * y),
      );
      if (side < 0.139 + Math.min(clearance, 0.008)) {
        return false;
      }
    }
    for (const c of obstacles) {
      const x = p.x - c.x,
        y = p.y - c.y;
      const side = Math.max(
        Math.abs(x),
        Math.abs(0.5 * x + 0.866025404 * y),
        Math.abs(-0.5 * x + 0.866025404 * y),
      );
      if (side < 0.163241943 + clearance) {
        return false;
      }
    }
    return true;
  };
}

// The grid is only an obstacle search aid. Visibility simplification and Bezier
// rounding replace its steps before any road geometry is produced.
function detour(
  start: Point2,
  end: Point2,
  free: ReturnType<typeof routingSpace>,
): Point2[] | null {
  const step = 0.018,
    padding = 0.4;
  const minX = Math.min(start.x, end.x) - padding,
    minY = Math.min(start.y, end.y) - padding;
  const cols = Math.ceil((Math.abs(start.x - end.x) + padding * 2) / step) + 1;
  const rows = Math.ceil((Math.abs(start.y - end.y) + padding * 2) / step) + 1;
  const nodes: Point2[] = [start, end];
  for (let y = 0; y < rows; y++) {
    for (let x = 0; x < cols; x++) {
      nodes.push({ x: minX + x * step, y: minY + y * step });
    }
  }
  const allowed = new Uint8Array(nodes.length);
  const available = (i: number) => {
    if (!allowed[i]) {
      allowed[i] = free(nodes[i], 0.018) ? 1 : 2;
    }
    return allowed[i] === 1;
  };
  const visible = (a: Point2, b: Point2) => sampleLine(a, b).every((p) => free(p, 0.018));
  const nearest = (p: Point2) => {
    const cx = Math.round((p.x - minX) / step),
      cy = Math.round((p.y - minY) / step);
    const result: number[] = [];
    for (let y = cy - 2; y <= cy + 2; y++) {
      for (let x = cx - 2; x <= cx + 2; x++) {
        const i = 2 + y * cols + x;
        if (x >= 0 && x < cols && y >= 0 && y < rows && available(i) && visible(p, nodes[i])) {
          result.push(i);
        }
      }
    }
    return result;
  };
  const endNeighbors = new Set(nearest(end));
  const costs = new Float64Array(nodes.length).fill(Infinity),
    parents = new Int32Array(nodes.length).fill(-1);
  const closed = new Uint8Array(nodes.length);
  const heap: { id: number; score: number }[] = [];
  const push = (id: number, score: number) => {
    heap.push({ id, score });
    let i = heap.length - 1;
    while (i > 0) {
      const parent = Math.floor((i - 1) / 2);
      if (heap[parent].score <= score) {
        break;
      }
      [heap[i], heap[parent]] = [heap[parent], heap[i]];
      i = parent;
    }
  };
  const pop = () => {
    const first = heap[0],
      last = heap.pop()!;
    if (heap.length) {
      heap[0] = last;
      let i = 0;
      while (i * 2 + 1 < heap.length) {
        let child = i * 2 + 1;
        if (child + 1 < heap.length && heap[child + 1].score < heap[child].score) {
          child++;
        }
        if (heap[i].score <= heap[child].score) {
          break;
        }
        [heap[i], heap[child]] = [heap[child], heap[i]];
        i = child;
      }
    }
    return first.id;
  };
  costs[0] = 0;
  push(0, distance(start, end));
  while (heap.length) {
    const id = pop();
    if (closed[id]) {
      continue;
    }
    closed[id] = 1;
    if (id === 1) {
      const path: Point2[] = [];
      for (let i = 1; i !== -1; i = parents[i]) {
        path.push(nodes[i]);
      }
      path.reverse();
      const simple = [path[0]];
      let anchor = 0;
      while (anchor < path.length - 1) {
        let next = path.length - 1;
        while (next > anchor + 1 && !visible(path[anchor], path[next])) {
          next--;
        }
        simple.push(path[next]);
        anchor = next;
      }
      return simple;
    }
    const neighbors: number[] = [];
    if (id === 0) {
      neighbors.push(...nearest(start));
    } else {
      const x = (id - 2) % cols,
        y = Math.floor((id - 2) / cols);
      for (let dy = -1; dy <= 1; dy++) {
        for (let dx = -1; dx <= 1; dx++) {
          if ((dx || dy) && x + dx >= 0 && x + dx < cols && y + dy >= 0 && y + dy < rows) {
            const next = id + dx + dy * cols;
            if (available(next) && visible(nodes[id], nodes[next])) {
              neighbors.push(next);
            }
          }
        }
      }
      if (endNeighbors.has(id)) {
        neighbors.push(1);
      }
    }
    for (const next of neighbors) {
      const cost = costs[id] + distance(nodes[id], nodes[next]);
      if (cost < costs[next]) {
        costs[next] = cost;
        parents[next] = id;
        push(next, cost + distance(nodes[next], end));
      }
    }
  }
  return null;
}

function roundPath(points: Point2[], free: ReturnType<typeof routingSpace>): Point2[] | null {
  const rounded: Point2[] = [points[0]];
  for (let i = 1; i < points.length - 1; i++) {
    const previous = points[i - 1],
      corner = points[i],
      next = points[i + 1];
    const incoming = distance(previous, corner),
      outgoing = distance(corner, next);
    let radius = Math.min(0.11, incoming * 0.45, outgoing * 0.45);
    let curve: Point2[] | null = null;
    for (let attempt = 0; attempt < 5; attempt++) {
      const a = mix(corner, previous, radius / incoming),
        b = mix(corner, next, radius / outgoing);
      const candidate = cubic(a, mix(a, corner, 2 / 3), mix(b, corner, 2 / 3), b);
      if (candidate.every((p) => free(p))) {
        curve = candidate;
        break;
      }
      radius *= 0.7;
    }
    if (!curve) {
      return null;
    }
    rounded.push(...sampleLine(rounded.at(-1)!, curve[0]).slice(1), ...curve.slice(1));
  }
  rounded.push(...sampleLine(rounded.at(-1)!, points.at(-1)!).slice(1));
  return rounded.every((p) => free(p)) ? rounded : null;
}

export function acceptableRoad(points: Point2[]): boolean {
  if (points.length < 3) {
    return false;
  }
  const start = points[0],
    end = points.at(-1)!;
  const chord = distance(start, end);
  if (chord < 0.018 || length(points) > chord * 1.3 + 0.008) {
    return false;
  }
  const dx = (end.x - start.x) / chord,
    dy = (end.y - start.y) / chord;
  let turning = 0;
  for (let i = 1; i < points.length - 1; i++) {
    const a = points[i - 1],
      b = points[i],
      c = points[i + 1];
    const ax = b.x - a.x,
      ay = b.y - a.y,
      bx = c.x - b.x,
      by = c.y - b.y;
    const al = Math.hypot(ax, ay),
      bl = Math.hypot(bx, by);
    if (al < 1e-8 || bl < 1e-8) {
      continue;
    }
    if ((ax * dx + ay * dy) / al < 0.38) {
      return false;
    }
    const angle = Math.acos(Math.max(-1, Math.min(1, (ax * bx + ay * by) / (al * bl))));
    turning += angle;
    if (angle > 0.001 && (al + bl) / (2 * angle) < 0.035) {
      return false;
    }
  }
  return turning <= 1.9;
}

function connectPorts(
  a: Port[],
  b: Port[],
  free: ReturnType<typeof routingSpace>,
  allowDetour = true,
): Point2[] | null {
  const pairs = a
    .flatMap((from) => b.map((to) => ({ from, to, score: distance(from.point, to.point) })))
    .filter(({ from, to, score }) => {
      const dx = (to.point.x - from.point.x) / score,
        dy = (to.point.y - from.point.y) / score;
      return (
        from.outward.x * dx + from.outward.y * dy > 0.35 &&
        -to.outward.x * dx - to.outward.y * dy > 0.35
      );
    });
  pairs.sort((x, y) => x.score - y.score);
  let best: Point2[] | null = null,
    bestLength = Infinity;
  for (const { from, to } of pairs) {
    for (const fraction of [0.2, 0.32, 0.45]) {
      const handle = Math.max(0.018, Math.min(0.12, distance(from.point, to.point) * fraction));
      const curve = cubic(
        from.point,
        add(from.point, from.outward, handle),
        add(to.point, to.outward, handle),
        to.point,
      );
      const span = length(curve);
      if (span < bestLength && acceptableRoad(curve) && curve.every((p) => free(p))) {
        best = curve;
        bestLength = span;
      }
    }
  }
  if (best) {
    return best;
  }
  if (!allowDetour) {
    return null;
  }
  for (const { from, to } of pairs.slice(0, 4)) {
    const start = add(from.point, from.outward, 0.042),
      end = add(to.point, to.outward, 0.042);
    if (
      !sampleLine(from.point, start).every((p) => free(p)) ||
      !sampleLine(to.point, end).every((p) => free(p))
    ) {
      continue;
    }
    const path = detour(start, end, free);
    if (!path) {
      continue;
    }
    const curve = roundPath([from.point, ...path, to.point], free);
    if (curve && acceptableRoad(curve)) {
      return curve;
    }
  }
  return null;
}

function surfaces(points: Point2[], byKey: Map<string, LandscapeTile>): RoadPath[] {
  const result: RoadPath[] = [];
  const kindAt = (p: Point2) =>
    byKey.get(keyOf(coordinateAt(p)))?.kind === "greenery" ? "path" : "road";
  for (let i = 1; i < points.length; i++) {
    const a = points[i - 1],
      b = points[i],
      kind = kindAt(mix(a, b, 0.5));
    const previous = result.at(-1);
    if (previous?.kind === kind) {
      previous.points.push(b);
    } else {
      result.push({ kind, width: 0.007, points: [a, b] });
    }
  }
  // Keep continuous edges at the paving/dirt transitions; the unpaved route is
  // broad enough for access, without adding urban curbs through forests.
  return result;
}

function sharedJunction(
  a: LandscapeTile,
  b: LandscapeTile,
  connections: LandscapeConnection[],
  tiles: LandscapeTile[],
  spaces: LandscapeSpace[],
): Point2[] | null {
  let best: Point2[] | null = null;
  for (const [source, shared] of [
    [a, b],
    [b, a],
  ]) {
    const sk = keyOf(shared.coordinate),
      sourceKey = keyOf(source.coordinate);
    const center = boardCenter(shared.coordinate);
    const free = routingSpace(tiles, spaces, new Set([sourceKey]));
    for (const trunk of connections.filter((c) => c.a === sk || c.b === sk)) {
      const path = trunk.routes.flatMap((r, i) => (i ? r.points.slice(1) : r.points));
      const sharedAtStart = distance(path[0], center) < distance(path.at(-1)!, center);
      for (let i = 3; i < path.length - 3; i += 6) {
        const point = path[i],
          radius = distance(point, center);
        if (radius < 0.163 || radius > 0.28) {
          continue;
        }
        const tangent = { x: path[i + 1].x - path[i - 1].x, y: path[i + 1].y - path[i - 1].y };
        const span = Math.hypot(tangent.x, tangent.y),
          sign = sharedAtStart ? 1 : -1;
        const to = {
          point,
          outward: { x: (tangent.x / span) * sign, y: (tangent.y / span) * sign },
        };
        const sourceCenter = boardCenter(source.coordinate),
          approachLength = distance(sourceCenter, point);
        const crossing = {
          point,
          outward: {
            x: (sourceCenter.x - point.x) / approachLength,
            y: (sourceCenter.y - point.y) / approachLength,
          },
        };
        const branch = connectPorts(ports(source), [to, crossing], free, false);
        if (branch && (!best || length(branch) < length(best))) {
          best = branch;
        }
      }
    }
  }
  return best;
}

export function routeLandscape(
  tiles: LandscapeTile[],
  spaces: LandscapeSpace[],
): LandscapeConnection[] {
  const cities = tiles.filter((t) => t.kind === "city" && t.layout);
  const byKey = new Map(tiles.map((t) => [keyOf(t.coordinate), t]));
  const connections: LandscapeConnection[] = [];
  const degree = new Map<string, number>();
  const components = new Map(
    cities.map((city) => {
      const key = keyOf(city.coordinate);
      return [key, key];
    }),
  );
  const component = (key: string): string => {
    let root = key;
    while (components.get(root) !== root) {
      root = components.get(root)!;
    }
    return root;
  };
  const pairs = cities.flatMap((a, i) =>
    cities
      .slice(i + 1)
      .filter((b) => hexDistance(a.coordinate, b.coordinate) <= 3)
      .map((b) => ({
        a,
        b,
        distance: distance(boardCenter(a.coordinate), boardCenter(b.coordinate)),
      })),
  );
  pairs.sort(
    (a, b) => a.distance - b.distance || keyOf(a.a.coordinate).localeCompare(keyOf(b.a.coordinate)),
  );
  for (const pair of pairs) {
    const { a, b } = pair,
      ak = keyOf(a.coordinate),
      bk = keyOf(b.coordinate);
    if (
      hexDistance(a.coordinate, b.coordinate) > 1 &&
      ((degree.get(ak) ?? 0) >= 3 || (degree.get(bk) ?? 0) >= 3)
    ) {
      continue;
    }
    if (component(ak) === component(bk)) {
      continue;
    }
    const direct = connectPorts(ports(a), ports(b), routingSpace(tiles, spaces, new Set([ak, bk])));
    const joined = sharedJunction(a, b, connections, tiles, spaces);
    const points = joined && (!direct || length(joined) + 0.025 < length(direct)) ? joined : direct;
    if (!points) {
      continue;
    }
    connections.push({ key: `${ak}|${bk}`, a: ak, b: bk, routes: surfaces(points, byKey) });
    components.set(component(ak), component(bk));
    degree.set(ak, (degree.get(ak) ?? 0) + 1);
    degree.set(bk, (degree.get(bk) ?? 0) + 1);
  }
  return connections;
}
