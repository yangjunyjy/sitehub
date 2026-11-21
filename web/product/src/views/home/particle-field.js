import {
  createTargets,
  createTerminalTarget,
  GRID,
  PARTICLES,
  STRIDE,
} from "./particle-targets";
import { createParticleLight } from "./particle-light";

const TAU = Math.PI * 2;
const strandWidth = GRID / 8;
const order = ["artifact", "cluster", "terminal"];
const hold = 1.65,
  passage = 3.35,
  beat = hold + passage;
const light = [
  "#464fce",
  "#6868ff",
  "#9d82ff",
  "#aeacff",
  "#548dff",
  "#45b6ff",
  "#28dafe",
  "#6af5ff",
  "#ccffff",
  "#ffad69",
  "#ffe5b2",
];
const night = [
  "#4048ad",
  "#5556e7",
  "#a380ff",
  "#a29dff",
  "#4c79ff",
  "#249ef9",
  "#17d6fc",
  "#55ebff",
  "#c4ffff",
  "#ffa369",
  "#ffdb9f",
];
const clamp = (x, lo, hi) => Math.min(hi, Math.max(lo, x));
const mix = (a, b, t) => a + (b - a) * t;
function ease(t) {
  t = clamp(t, 0, 1);
  return t * t * (3 - 2 * t);
}
const opacity = [0.09, 0.18, 0.3, 0.44, 0.58, 0.7, 0.83, 0.95];

// A signal has a bright head and a longer, soft wake. The phase belongs to the
// geometry, so sorting or turning the cloud cannot change its direction.
function pulse(pathPhase, clock) {
  if (pathPhase < 0) return 0;
  const position = clock - pathPhase + 0.025;
  const age = position - Math.floor(position);
  if (age > 0.2) return 0;
  return ease(age / 0.025) * (1 - ease((age - 0.025) / 0.175));
}

function flowPaths() {
  const paths = order.map(() => new Float32Array(PARTICLES * 3));
  for (let i = 0; i < PARTICLES; i++) {
    const strand = Math.floor((i % GRID) / strandWidth);
    const v = (i % GRID) / (GRID - 1);
    const u =
      (Math.floor(i / GRID) + ((i % strandWidth) + strand / 8) / strandWidth) /
      GRID;
    const spread = ((((i + 0.5) * 0.754877666) % 1) - 0.5) * 0.05;
    const sweep = u * TAU;
    const theta = (strand / 8) * TAU + sweep * 1.05 + spread;
    const x = (u - 0.5) * 2.68;
    // Eight narrow filaments twist in flight, rise off the map and open like a fan.
    paths[0].set(
      [
        x,
        0.19 * Math.sin(sweep - 1.1) + (0.43 + spread) * Math.sin(theta),
        0.63 * Math.cos(theta) + 0.21 * Math.sin(sweep),
      ],
      i * 3,
    );
    const orbit = (u - 0.5) * TAU * 1.14 + (v - 0.5) * 1.38;
    paths[1].set(
      [
        (0.96 + spread) * Math.sin(orbit),
        0.55 - 1.22 * u + 0.22 * Math.sin(orbit) + (strand / 7 - 0.5) * 0.32,
        0.9 * Math.cos(orbit),
      ],
      i * 3,
    );
    paths[2].set(
      [
        x,
        0.38 * Math.sin(sweep - 0.7) +
          (strand / 7 - 0.5) * (0.18 + 1.29 * Math.sin(Math.PI * u)) +
          spread,
        0.55 * Math.cos(theta * 0.8),
      ],
      i * 3,
    );
  }
  return paths;
}

export function createParticleField(canvas, shape) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return null;
  let states;
  const lighting = createParticleLight();
  const first = Math.max(0, order.indexOf(shape));
  const paths = flowPaths();
  const delays = new Float32Array(PARTICLES);
  const grain = new Float32Array(PARTICLES);
  const pixels = new Float32Array(PARTICLES * 3);
  const next = new Int32Array(PARTICLES);
  const heads = new Int32Array(light.length * opacity.length);
  const trailSegments = new Float32Array(Math.ceil(PARTICLES / 5) * 4);
  for (let i = 0; i < PARTICLES; i++) {
    const u = Math.floor(i / GRID) / (GRID - 1),
      v = (i % GRID) / (GRID - 1);
    const seed = Math.sin(i * 127.1 + 17.47) * 43758.5453;
    grain[i] = seed - Math.floor(seed);
    delays[i] =
      0.16 * u +
      0.05 * (0.5 + 0.5 * Math.sin(2 * TAU * v + 2 * u)) +
      0.01 * grain[i];
  }
  let width = 0,
    height = 0,
    scale = 0,
    time = 0,
    yaw = 0,
    pitch = 0,
    wantedYaw = 0,
    wantedPitch = 0;
  let pointerX = 0,
    pointerY = 0,
    pointerStrength = 0,
    hover = false,
    dark = false;
  let hasPreviousFrame = false;
  let terminalColumns = 0;
  function resize() {
    width = canvas.clientWidth;
    height = canvas.clientHeight;
    // Supersample ordinary displays as well as retaining fine edges on high-DPI phones.
    const ratio = Math.min(
      Math.max(1.5, window.devicePixelRatio || 1),
      3,
      Math.sqrt(2400000 / Math.max(1, width * height)),
    );
    canvas.width = Math.round(width * ratio);
    canvas.height = Math.round(height * ratio);
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    scale = Math.min(width * 0.345, height * 0.47);
    const columns = scale < 170 ? 40 : 64;
    if (!states) states = createTargets(columns);
    else if (columns !== terminalColumns) {
      states.terminal = createTerminalTarget(columns);
    }
    terminalColumns = columns;
    hasPreviousFrame = false;
    lighting.resize(ctx, width, height, scale, dark);
  }

  function draw(dt, animate) {
    if (!width || !height) return order[first];
    if (animate) time = (time + dt) % (beat * order.length);
    else time = 0;
    const section = Math.floor(time / beat);
    const sourceIndex = (first + section) % order.length,
      destIndex = (sourceIndex + 1) % order.length;
    const from = states[order[sourceIndex]],
      to = states[order[destIndex]];
    const progress = clamp(((time % beat) - hold) / passage, 0, 1);
    const follow = animate ? 1 - Math.exp(-dt * 10) : 1;
    yaw += (wantedYaw - yaw) * follow;
    pitch += (wantedPitch - pitch) * follow;
    pointerStrength += ((hover && animate ? 1 : 0) - pointerStrength) * follow;
    const pan = yaw + 0.12 * Math.sin((TAU * time) / (beat * order.length));
    const tilt =
      pitch + 0.045 * Math.sin((2 * TAU * time) / (beat * order.length));
    const cy = Math.cos(pan),
      sy = Math.sin(pan),
      cp = Math.cos(tilt),
      sp = Math.sin(tilt);
    const mapAmount =
      (sourceIndex === 1 ? 1 - ease(progress * 2) : 0) +
      (destIndex === 1 ? ease(progress * 2 - 1) : 0);
    const typed = !animate
      ? 1
      : sourceIndex === 2
        ? clamp(0.18 + (time % beat) / 1.35, 0, 1)
        : 0.18 * ease((progress - 0.76) / 0.24);
    const caret = !animate
      ? 1
      : 0.45 + 0.55 * (0.5 + 0.5 * Math.cos(TAU * time)) ** 2;
    const stream = paths[sourceIndex];
    const palette = dark ? night : light;
    const pulseClock = time * 0.8 + 0.18;
    const dotRadius = Math.max(0.38, scale * 0.0028);
    const twist = 1.52 * (progress - 0.5);
    const roll = Math.sin(twist),
      unroll = Math.cos(twist);
    let trailLength = 0;
    heads.fill(-1);
    ctx.clearRect(0, 0, width, height);
    lighting.background(ctx, width, height, pan, tilt);
    ctx.globalCompositeOperation = "lighter";
    for (let i = 0; i < PARTICLES; i++) {
      const k = i * STRIDE,
        n = i * 3;
      const t = clamp((progress - delays[i]) / 0.78, 0, 1);
      const depart = ease(t / 0.48),
        arrive = ease((t - 0.49) / 0.51);
      const source = 1 - depart,
        ribbon = depart - arrive;
      const x = source * from[k] + ribbon * stream[n] + arrive * to[k];
      const y =
        source * from[k + 1] +
        ribbon * (unroll * stream[n + 1] - roll * stream[n + 2]) +
        arrive * to[k + 1];
      const z =
        source * from[k + 2] +
        ribbon * (roll * stream[n + 1] + unroll * stream[n + 2]) +
        arrive * to[k + 2];
      const side = x * cy + z * sy,
        depth = z * cy - x * sy;
      const front = y * sp + depth * cp;
      const perspective = 1 + 0.095 * front;
      let px = width * 0.5 + scale * side * perspective;
      let py = height * 0.5 - scale * (y * cp - depth * sp) * perspective;
      const dx = px - pointerX,
        dy = py - pointerY,
        reach = scale * 0.34;
      let contact = 0;
      if (pointerStrength > 0.01 && dx * dx + dy * dy < reach * reach) {
        const proximity = 1 - Math.sqrt(dx * dx + dy * dy) / reach;
        contact = pointerStrength * proximity;
        px += dx * contact * proximity * 1.8;
        py += dy * contact * proximity * 1.8;
      }
      const signal =
        source * pulse(from[k + 5], pulseClock) +
        arrive * pulse(to[k + 5], pulseClock) +
        ribbon *
          pulse(
            i / PARTICLES + Math.floor((i % GRID) / strandWidth) * 0.08,
            pulseClock,
          );
      // A thin travelling wave crosses the surface and picks out individual grains.
      const scan = Math.max(
        contact,
        pulse(0.63 + x * 0.28 + y * 0.12, time * 0.4 + 0.65),
      );
      const weight =
        source * from[k + 4] +
        arrive * to[k + 4] +
        ribbon * (0.6 + 0.45 * grain[i]);
      const baseHue =
        source * from[k + 3] +
        arrive * to[k + 3] +
        ribbon * (0.1 + Math.floor((i % GRID) / strandWidth) * 0.103);
      const hue = clamp(
        Math.round(
          8 * mix(baseHue, 1 + mapAmount * 0.24, signal * 0.95) + scan * 0.85,
        ),
        0,
        palette.length - 1,
      );
      const terminalWeight =
        sourceIndex === 2 ? source : destIndex === 2 ? arrive : 0;
      const onset = sourceIndex === 2 ? from[k + 6] : to[k + 6];
      let ink = 1;
      if (terminalWeight && onset !== -1) {
        const typedInk =
          onset === -2
            ? caret * ease((typed - 0.78) / 0.07)
            : ease((typed - onset) / 0.055);
        ink = mix(1, typedInk, terminalWeight);
      }
      const visibility =
        ink *
        clamp(
          0.53 +
            front * 0.4 +
            (grain[i] - 0.5) * 0.37 +
            0.36 * (weight - 1) +
            0.34 * signal +
            0.2 * scan,
          0,
          1,
        );
      const band = Math.round(visibility * (opacity.length - 1));
      const bucket = band * palette.length + hue;
      if (
        animate &&
        hasPreviousFrame &&
        t > 0 &&
        t < 1 &&
        i % 5 === 0 &&
        dt > 0
      ) {
        const vx = px - pixels[n],
          vy = py - pixels[n + 1];
        const speed = Math.hypot(vx, vy);
        if (speed > 0.15) {
          const length =
            Math.min(0.12 / dt, (scale * 0.19) / speed) * Math.sin(Math.PI * t);
          trailSegments[trailLength++] = px - length * vx;
          trailSegments[trailLength++] = py - length * vy;
          trailSegments[trailLength++] = px;
          trailSegments[trailLength++] = py;
        }
      }
      pixels[n] = px;
      pixels[n + 1] = py;
      pixels[n + 2] =
        dotRadius *
        (0.67 + grain[i] * 0.81 + visibility * 0.34) *
        weight *
        (1 + 0.5 * signal);
      if (ink < 0.02) continue;
      const emission = Math.max(signal, contact, scan * 0.54) * visibility;
      if (grain[i] > 0.92 && (emission > 0.18 || weight > 1.08)) {
        const radius = dotRadius * (9 + 17 * emission + 4 * weight);
        lighting.glow(
          ctx,
          px,
          py,
          radius,
          hue,
          (0.15 * visibility + 0.56 * emission) * (1 - terminalWeight * 0.62),
        );
      }
      next[i] = heads[bucket];
      heads[bucket] = i;
    }
    if (trailLength) {
      ctx.beginPath();
      ctx.lineWidth = Math.max(0.6, scale * 0.0035);
      ctx.lineCap = "round";
      ctx.globalAlpha = 1;
      ctx.strokeStyle = "rgba(80,163,255,.43)";
      for (let j = 0; j < trailLength; j += 4) {
        ctx.moveTo(trailSegments[j], trailSegments[j + 1]);
        ctx.lineTo(trailSegments[j + 2], trailSegments[j + 3]);
      }
      ctx.stroke();
    }
    for (let b = 0; b < heads.length; b++) {
      if (heads[b] < 0) continue;
      ctx.fillStyle = palette[b % palette.length];
      ctx.globalAlpha = opacity[Math.floor(b / palette.length)];
      ctx.beginPath();
      for (let i = heads[b]; i >= 0; i = next[i]) {
        const n = i * 3,
          x = pixels[n],
          y = pixels[n + 1],
          r = pixels[n + 2];
        ctx.moveTo(x + r, y);
        ctx.arc(x, y, r, 0, TAU);
      }
      ctx.fill();
    }
    ctx.globalAlpha = 1;
    ctx.globalCompositeOperation = "source-over";
    hasPreviousFrame = animate;
    return order[progress < 0.5 ? sourceIndex : destIndex];
  }
  return {
    draw,
    resize,
    theme(value) {
      dark = value;
      lighting.resize(ctx, width, height, scale, dark);
    },
    turn(x, y) {
      wantedYaw = clamp(wantedYaw + x, -1.35, 1.35);
      wantedPitch = clamp(wantedPitch - y, -0.82, 0.82);
    },
    pointer(x, y) {
      pointerX = x;
      pointerY = y;
      hover = true;
    },
    clearPointer() {
      hover = false;
    },
  };
}
