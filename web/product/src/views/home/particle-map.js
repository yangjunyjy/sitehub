import outlines from "./china-outline.json";

// DataV GeoAtlas national boundary, simplified with a tolerance of 0.035 degrees.
// https://geo.datav.aliyun.com/areas_v3/bound/100000.json
// Albers equal-area conic, central meridian 105°E and standard parallels 25°N/47°N.
const rad = Math.PI / 180;
const n = (Math.sin(25 * rad) + Math.sin(47 * rad)) / 2;
const c = Math.cos(25 * rad) ** 2 + 2 * n * Math.sin(25 * rad);
const rho = (lat) => Math.sqrt(c - 2 * n * Math.sin(lat * rad)) / n;
const origin = rho(36);
function project(lon, lat) {
  const theta = n * (lon - 105) * rad,
    r = rho(lat);
  return [
    0.08 + 2.84 * r * Math.sin(theta),
    2.84 * (origin - r * Math.cos(theta)),
  ];
}

function southSea(lon, lat) {
  return [1.17 + (lon - 114) * 0.018, -0.72 + (lat - 12) * 0.024];
}

export function mapPoints(count) {
  const points = [],
    edges = [],
    rows = new Map();
  const step = 0.006;
  let perimeter = 0;

  function ring(coords, projectPoint, inset) {
    const vertices = coords.map(([lon, lat]) => projectPoint(lon, lat));
    let [ax, ay] = vertices.at(-1);
    for (const [bx, by] of vertices) {
      const length = Math.hypot(bx - ax, by - ay);
      if (length > 0) {
        perimeter += length;
        edges.push({ ax, ay, bx, by, length, end: perimeter, inset });
      }
      // Pair crossings across each scanline to distribute the interior by land area.
      for (
        let row = Math.ceil(Math.min(ay, by) / step);
        row < Math.ceil(Math.max(ay, by) / step);
        row++
      ) {
        if (!rows.has(row)) rows.set(row, []);
        rows.get(row).push(ax + ((row * step - ay) / (by - ay)) * (bx - ax));
      }
      ax = bx;
      ay = by;
    }
  }
  for (const coords of outlines) {
    const north = coords.some(([, lat]) => lat >= 18);
    if (north) ring(coords, project, false);
    if (!north || coords.every(([, lat]) => lat < 20.25))
      ring(coords, southSea, true);
  }
  function outline(count, depth) {
    let edge = 0;
    for (let i = 0; i < count; i++) {
      const distance = ((i + 0.5) / count) * perimeter;
      while (edges[edge].end < distance) edge++;
      const { ax, ay, bx, by, length, end, inset } = edges[edge];
      const t = 1 - (end - distance) / length;
      const x = ax + (bx - ax) * t,
        y = ay + (by - ay) * t;
      const offset = inset ? 0 : depth;
      points.push([
        x + offset * 0.15,
        y + offset * 0.52,
        0.1 + offset,
        depth ? 0.32 : 0.8,
        depth ? 0.7 : 1.08,
        -1,
      ]);
    }
  }
  outline(2140, 0);
  outline(680, -0.09);
  // A small separate frame locates the southern island groups at their own scale.
  for (const x of [1.0, 1.34])
    for (const y of [-0.96, -0.48]) {
      for (let i = 0; i < 24; i++) {
        const t = (i % 12) / 11;
        points.push([
          x + (i < 12 ? (1.17 - x) * 0.28 * t : 0),
          y + (i >= 12 ? (-0.72 - y) * 0.23 * t : 0),
          0.09,
          0.47,
          0.8,
          -1,
        ]);
      }
    }
  // Beijing, Shanghai, Shenzhen, Chengdu and Urumqi.
  const hubs = [
    [116.4, 39.9],
    [121.47, 31.23],
    [114.06, 22.54],
    [104.07, 30.67],
    [87.62, 43.83],
  ].map(([lon, lat]) => project(lon, lat));
  for (const [route, [from, to]] of [
    [4, 3],
    [3, 0],
    [0, 1],
    [3, 2],
  ].entries()) {
    const [ax, ay] = hubs[from],
      [bx, by] = hubs[to];
    for (let i = 0; i < 120; i++) {
      const t = i / 119,
        arch = Math.sin(Math.PI * t);
      points.push([
        ax + (bx - ax) * t,
        ay + (by - ay) * t + arch * 0.12,
        0.13 + arch * 0.24,
        0.86,
        0.83,
        t + route * 0.19,
      ]);
    }
  }
  hubs.forEach(([x, y], j) => {
    for (let i = 0; i < 45; i++) {
      const angle = (i * 2 * Math.PI) / 36,
        r = i < 36 ? 0.022 : 0.006;
      points.push([
        x + r * Math.cos(angle),
        y + r * Math.sin(angle),
        0.14,
        1.18,
        1.08,
        j * 0.19,
      ]);
    }
  });
  const spans = [];
  let area = 0;
  for (const row of [...rows.keys()].sort((a, b) => a - b)) {
    const xs = rows.get(row).sort((a, b) => a - b);
    for (let j = 0; j + 1 < xs.length; j += 2) {
      const width = xs[j + 1] - xs[j];
      if (width <= 0) continue;
      area += width;
      spans.push({ x: xs[j], y: row * step, width, end: area });
    }
  }
  const interior = count - points.length;
  let span = 0;
  for (let i = 0; i < interior; i++) {
    const target = ((i + 0.5) / interior) * area;
    while (spans[span].end < target) span++;
    const { x, y, width } = spans[span];
    const px = x + (((i + 0.5) * 0.618033989) % 1) * width;
    const wave = Math.sin(px * 6 + y * 4) * Math.sin(y * 7 - px * 3);
    points.push([
      px,
      y,
      0.095 + wave * 0.018,
      0.42 + 0.19 * wave + px * 0.12,
      0.97,
      -1,
    ]);
  }
  return points;
}
