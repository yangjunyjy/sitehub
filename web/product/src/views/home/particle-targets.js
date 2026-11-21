import { mapPoints } from "./particle-map";
import { terminalPoints } from "./particle-terminal";

const TAU = Math.PI * 2;
export const GRID = 96;
export const PARTICLES = GRID * GRID;
export const STRIDE = 7;
const mix = (a, b, t) => a + (b - a) * t;

// A common spatial ordering keeps neighbouring particles in roughly the same
// flow when changing shape. Target attributes are xyz, hue, weight, signal phase
// and terminal reveal time. A negative signal phase leaves the point unlit.
function pack(points) {
  points.sort((a, b) => a[0] - b[0]);
  const result = new Float32Array(PARTICLES * STRIDE);
  for (let start = 0; start < PARTICLES; start += GRID) {
    const column = points
      .slice(start, start + GRID)
      .sort((a, b) => a[1] - b[1]);
    column.forEach((point, i) => {
      const offset = (start + i) * STRIDE;
      result.set(point, offset);
      result[offset + 6] = point[6] ?? -1;
    });
  }
  return result;
}

function plane() {
  const points = [];
  const nose = [1.17, 0.72, 0.02];
  const edge = [-1.08, 0.14, -0.12];
  const ridge = [-0.36, -0.15, 0.32];
  const keel = [-0.28, -0.66, -0.1];
  const fold = [-0.08, -0.32, 0.25];
  const tip = [0.39, -0.83, -0.08];

  function wing(a, b, count, rows, hue, weight) {
    for (let row = 0; row < rows; row++) {
      const start = Math.round(count * (row / rows) ** 2);
      const dots = Math.round(count * ((row + 1) / rows) ** 2) - start;
      const along = (row + 0.6) / rows;
      for (let col = 0; col < dots; col++) {
        const across = (col + (row % 2 ? 0.3 : 0.7)) / dots;
        const camber = Math.sin(Math.PI * across);
        const bow = along + 0.074 * camber * along * (1 - along);
        const p = nose.map((n, k) => mix(n, mix(a[k], b[k], across), bow));
        p[2] += 0.08 * camber * along;
        const edgeLight = (1 - across) ** 4;
        const sheen = Math.exp(-((across - 0.24 - 0.3 * along) ** 2) / 0.007);
        points.push([
          ...p,
          hue + 0.15 * camber + 0.1 * along + 0.22 * sheen,
          weight + edgeLight * 0.26 + 0.2 * sheen,
          sheen > 0.83 ? 1 - along : -1,
        ]);
      }
    }
  }
  wing(edge, ridge, 2300, 66, 0.02, 0.72);
  wing(ridge, keel, 1035, 51, 0.56, 1.12);
  wing(keel, fold, 518, 45, 0.04, 0.92);
  wing(fold, tip, 2457, 68, 0.44, 0.94);
  // Run optical traces over the actual folds. Each line has a short angled jog.
  for (let trace = 0; trace < 7; trace++) {
    const near = trace >= 3;
    const a = near ? fold : edge,
      b = near ? tip : ridge;
    for (let i = 0; i < 126; i++) {
      const u = i / 125;
      const along = 0.15 + 0.8 * u;
      const across =
        0.11 +
        (trace % 4) * 0.19 +
        0.1 * Math.min(1, Math.max(0, (u - 0.46) * 8));
      const p = nose.map((n, k) => mix(n, mix(a[k], b[k], across), along));
      p[2] += 0.08 * Math.sin(Math.PI * across) * along + 0.012;
      points.push([...p, near ? 0.86 : 0.26, 0.94, 1 - u + trace * 0.062]);
    }
  }
  // Thin seams and a sweeping wake are part of the cloud, so they also join the next shape.
  for (const [end, count, hue] of [
    [edge, 189, 0.19],
    [ridge, 288, 0.87],
    [fold, 207, 0.74],
    [tip, 216, 0.3],
  ]) {
    for (let i = 0; i < count; i++) {
      const t = i / count;
      points.push([...nose.map((n, k) => mix(n, end[k], t)), hue, 1.24, 1 - t]);
    }
  }
  const wakeCount = PARTICLES - points.length;
  for (let i = 0; i < wakeCount; i++) {
    const u = i / wakeCount,
      strand = i % 7;
    const theta = (strand * TAU) / 7 + 5.8 * u;
    const spread = (1 - u) * (0.08 + 0.11 * Math.sin(Math.PI * u));
    points.push([
      -1.42 + 1.12 * u,
      -0.72 + 0.45 * u + 0.13 * Math.sin(5 * u) + spread * Math.sin(theta),
      -0.1 + spread * Math.cos(theta) + 0.24 * Math.sin(4.3 * u),
      0.29 + 0.077 * strand,
      0.44 + 0.58 * u,
      u + strand * 0.038,
    ]);
  }
  return points;
}

export function createTerminalTarget(columns) {
  return pack(terminalPoints(PARTICLES, columns));
}

export function createTargets(columns = 64) {
  return {
    artifact: pack(plane()),
    cluster: pack(mapPoints(PARTICLES)),
    terminal: createTerminalTarget(columns),
  };
}
