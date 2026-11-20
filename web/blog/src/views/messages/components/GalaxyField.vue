<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue';
const canvas = ref(null);
let dispose;
onMounted(() => {
  const element = canvas.value;
  const context = element.getContext('2d');
  if (!context) return;
  const motion = matchMedia('(prefers-reduced-motion: reduce)');
  const palette = ['#d7deff', '#ffebce', '#b9b8ff', '#e3b8df', '#fcf8ff'];
  let seed = 7126;
  function random() {
    seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
    return seed / 4294967296;
  }
  const dust = Array.from({ length: 3300 }, (_, index) => {
    const radius = Math.pow(random(), .8);
    const angle = (index % 3) * Math.PI * 2 / 3 + radius * 5.9 + (random() + random() - 1) * (1.12 - radius * .58);
    return {
      x: Math.cos(angle) * radius, z: Math.sin(angle) * radius,
      y: (random() + random() - 1) * (.11 - .07 * radius),
      ring: Math.min(11, Math.floor(radius * 12)), phase: angle,
      size: .5 + random() ** 4 * 1.8, color: palette[Math.floor(random() * palette.length)],
      alpha: .35 + random() * .62, bright: index % 64 === 0,
    };
  });
  const flare = document.createElement('canvas');
  flare.width = flare.height = 48;
  const brush = flare.getContext('2d');
  const halo = brush.createRadialGradient(24, 24, 0, 24, 24, 24);
  halo.addColorStop(0, '#fff8dd'); halo.addColorStop(.11, '#ffeddb'); halo.addColorStop(.24, '#d5c0ffaa'); halo.addColorStop(1, '#9481d800');
  brush.fillStyle = halo; brush.fillRect(0, 0, 48, 48);
  brush.fillStyle = '#fff7'; brush.fillRect(9, 23.5, 30, 1); brush.fillRect(23.5, 12, 1, 24);
  const stars = Array.from({ length: 260 }, () => ({ x: random(), y: random(), depth: random(), phase: random() * 6.28 }));
  let width, height, frame = 0, visible = true, time = 0, previous = 0;
  function draw(now) {
    frame = 0;
    if (!document.hidden && visible && !motion.matches) {
      if (previous) time += Math.min(80, now - previous) / 1000;
      previous = now;
      frame = requestAnimationFrame(draw);
    }
    context.clearRect(0, 0, width, height);
    for (const star of stars) {
      context.globalAlpha = .38 + .36 * Math.sin(star.phase + time * .38);
      context.fillStyle = '#dee7ff';
      const point = .6 + star.depth * 1.4;
      context.fillRect(star.x * width, star.y * height, point, point);
    }
    context.save();
    context.translate(width * .53, height * .48);
    context.rotate(-.29);
    const scale = Math.min(Math.max(width * .53, height * .46), height * 1.24);
    context.globalAlpha = 1;
    context.save();
    context.scale(scale, scale * .32);
    const glow = context.createRadialGradient(0, 0, 0, 0, 0, 1);
    glow.addColorStop(0, '#fff1cfcd'); glow.addColorStop(.09, '#f8c8a9a0'); glow.addColorStop(.25, '#c88cb665'); glow.addColorStop(.64, '#7875bd32'); glow.addColorStop(1, 'transparent');
    context.fillStyle = glow;
    context.fillRect(-1, -1, 2, 2);
    context.restore();
    const rings = Array.from({ length: 12 }, (_, ring) => {
      const angle = time * .125 + .23 * Math.sin(time * .43 - ring * .3);
      return { sine: Math.sin(angle), cosine: Math.cos(angle) };
    });
    for (const grain of dust) {
      const { sine, cosine } = rings[grain.ring];
      const depth = grain.x * sine + grain.z * cosine;
      const perspective = 1 / (1.35 - depth * .32);
      const horizontal = (grain.x * cosine - grain.z * sine) * perspective * scale;
      const vertical = (grain.y + depth * .47) * perspective * scale;
      const size = grain.size * perspective * Math.min(1.6, scale / 520 + .5);
      context.globalAlpha = grain.alpha * (.8 + depth * .19);
      context.fillStyle = grain.color;
      context.fillRect(horizontal, vertical, size, size);
      if (grain.bright) {
        const diameter = (8 + 4 * Math.sin(time * 1.5 + grain.phase)) * perspective;
        context.drawImage(flare, horizontal - diameter / 2, vertical - diameter / 2, diameter, diameter);
      }
    }
    context.restore();
    context.globalAlpha = 1;
  }
  function refresh() {
    cancelAnimationFrame(frame);
    previous = 0;
    draw(performance.now());
  }
  function resize() {
    width = element.clientWidth; height = element.clientHeight;
    const ratio = Math.min(devicePixelRatio || 1, 2);
    element.width = Math.round(width * ratio); element.height = Math.round(height * ratio);
    context.setTransform(ratio, 0, 0, ratio, 0, 0);
    refresh();
  }
  const resizeObserver = new ResizeObserver(resize);
  const observer = new IntersectionObserver(([entry]) => { visible = entry.isIntersecting; refresh(); });
  resize(); resizeObserver.observe(element); observer.observe(element);
  motion.addEventListener('change', refresh); document.addEventListener('visibilitychange', refresh);
  dispose = () => { cancelAnimationFrame(frame); resizeObserver.disconnect(); observer.disconnect(); motion.removeEventListener('change', refresh); document.removeEventListener('visibilitychange', refresh); };
});
onBeforeUnmount(() => dispose?.());
</script>
<template><canvas ref="canvas" aria-hidden="true" /></template>
<style scoped>canvas { position: absolute; inset: 0; width: 100%; height: 100%; pointer-events: none; }</style>
