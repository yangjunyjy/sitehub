// Sample the ink once at high resolution. The letters are ordinary morphing
// particles, including their counters and the spaces between characters.
function textSampler(points) {
  const mask = document.createElement("canvas");
  mask.width = 960;
  mask.height = 116;
  const ctx = mask.getContext("2d", { willReadFrequently: true });
  ctx.font = "600 88px Consolas, Menlo, monospace";
  ctx.fillStyle = "#fff";
  const advance = ctx.measureText("M").width;
  function text(
    value,
    x,
    baseline,
    cell,
    count,
    hue,
    start = -1,
    duration = 0,
  ) {
    ctx.clearRect(0, 0, mask.width, mask.height);
    ctx.fillText(value, 3, 88);
    const width = Math.ceil(advance * value.length) + 6;
    const { data } = ctx.getImageData(0, 0, width, mask.height);
    const pixels = [];
    for (let i = 0; i < data.length / 4; i++)
      if (data[i * 4 + 3] > 100) pixels.push(i);
    const pixelSize = cell / advance;
    for (let i = 0; i < count; i++) {
      const pixel = pixels[Math.floor(((i + 0.5) * pixels.length) / count)];
      const col = pixel % width,
        row = Math.floor(pixel / width);
      const column = Math.max(0, Math.floor((col - 3) / advance));
      const onset =
        start < 0
          ? -1
          : start + (duration * column) / Math.max(1, value.length - 1);
      points.push([
        x + (col - 3) * pixelSize,
        baseline + (88 - row) * pixelSize,
        0.12,
        hue,
        0.74,
        -1,
        onset,
      ]);
    }
  }
  return text;
}

export function terminalPoints(count, columns) {
  const points = [],
    text = textSampler(points);
  const halfWidth = 1.17,
    halfHeight = 0.76,
    corner = 0.095;
  function rim(samples, depth, hue, weight) {
    const w = halfWidth - corner,
      h = halfHeight - corner;
    const quarter = (Math.PI * corner) / 2;
    const lengths = [
      2 * w,
      quarter,
      2 * h,
      quarter,
      2 * w,
      quarter,
      2 * h,
      quarter,
    ];
    const perimeter = lengths.reduce((sum, len) => sum + len, 0);
    for (let i = 0; i < samples; i++) {
      let d = ((i + 0.5) / samples) * perimeter,
        edge = 0;
      while (d > lengths[edge]) d -= lengths[edge++];
      let x, y;
      if (edge % 2) {
        const angle = (Math.PI / 2) * (1 - Math.floor(edge / 2)) - d / corner;
        x = (edge < 4 ? w : -w) + corner * Math.cos(angle);
        y = (edge === 1 || edge === 7 ? h : -h) + corner * Math.sin(angle);
      } else {
        [x, y] = [
          [-w + d, halfHeight],
          [halfWidth, h - d],
          [w - d, -halfHeight],
          [-halfWidth, -h + d],
        ][edge / 2];
      }
      points.push([x + depth * 0.25, y + depth * 0.48, depth, hue, weight, -1]);
    }
  }
  rim(1176, 0.06, 0.7, 1.05);
  rim(610, -0.09, 0.31, 0.75);
  const titleBarHeight = 0.16;
  const divider = halfHeight - titleBarHeight;
  const titleCenter = halfHeight - titleBarHeight / 2;
  for (let i = 0; i < 440; i++) {
    points.push([
      -halfWidth + (2 * halfWidth * i) / 439,
      divider,
      0.06,
      0.51,
      0.92,
      -1,
    ]);
  }
  for (let lamp = 0; lamp < 3; lamp++) {
    for (let i = 0; i < 90; i++) {
      const angle = i * 2.39996323,
        r = 0.018 * Math.sqrt((i + 0.5) / 90);
      points.push([
        -halfWidth + 0.095 + lamp * 0.068 + r * Math.cos(angle),
        titleCenter + r * Math.sin(angle),
        0.1,
        [1.14, 0.42, 0.87][lamp],
        1.02,
        -1,
      ]);
    }
  }
  const padding = 0.09;
  const cell = (2 * (halfWidth - padding)) / columns;
  const leading = cell * 2.55;
  const left = -halfWidth + padding;
  const firstLine = divider - padding - cell;
  const commandX = left + 2 * cell;
  const lastLine = firstLine - 2 * leading;
  const titleCell = Math.min(cell, 0.043);
  text(
    "bash",
    -2 * titleCell,
    titleCenter - 0.54 * titleCell,
    titleCell,
    300,
    0.56,
  );
  text("$", left, firstLine, cell, 115, 0.81);
  text("hostname", commandX, firstLine, cell, 800, 0.94, 0.04, 0.51);
  text("deploy-01", left, firstLine - leading, cell, 860, 0.63, 0.68);
  text("$", left, lastLine, cell, 115, 0.81, 0.78);
  for (let i = 0; i < 84; i++) {
    points.push([
      commandX + ((i * 0.618033989) % 1) * cell * 0.78,
      lastLine + (i / 84) * cell * 1.22,
      0.12,
      0.91,
      0.72,
      -1,
      -2,
    ]);
  }
  // The remaining points sit behind the text as a low-intensity phosphor surface.
  const glass = count - points.length;
  for (let i = 0; i < glass; i++) {
    points.push([
      -1.1 + 2.2 * ((i * 0.618033989) % 1),
      -0.7 + (divider + 0.66) * (i / glass),
      -0.04,
      0.24,
      0.2,
      -1,
    ]);
  }
  return points;
}
