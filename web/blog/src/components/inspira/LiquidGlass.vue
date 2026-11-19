<!-- Adapted from Inspira UI LiquidGlass. See public/licenses/inspira-ui.txt. -->
<script setup>
import { computed, onMounted, onBeforeUnmount, ref, useId } from "vue";
const props = defineProps({ radius: { type: Number, default: 22 } });
const surface = ref(null);
const size = ref(null);
const id = "glass-" + useId();
let observer;

onMounted(() => {
  // Chromium supports displacement backdrop filters. Other engines use frosted glass.
  if (!window.ResizeObserver || !/Chrome\/|Chromium\//.test(navigator.userAgent)) return;
  observer = new ResizeObserver(([entry]) => {
    const { width, height } = entry.contentRect;
    if (width > 0 && height > 0) size.value = { width, height };
  });
  observer.observe(surface.value);
});
onBeforeUnmount(() => observer?.disconnect());

const displacement = computed(() => {
  if (!size.value) return "";
  const { width, height } = size.value;
  const rim = Math.min(width, height) * .035;
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}">
    <defs>
      <linearGradient id="red" x1="100%" x2="0%"><stop stop-color="red" stop-opacity="0"/><stop offset="1" stop-color="red"/></linearGradient>
      <linearGradient id="blue" x2="0%" y2="100%"><stop stop-color="blue" stop-opacity="0"/><stop offset="1" stop-color="blue"/></linearGradient>
    </defs>
    <rect width="100%" height="100%" fill="black"/>
    <rect width="100%" height="100%" rx="${props.radius}" fill="url(#red)"/>
    <rect width="100%" height="100%" rx="${props.radius}" fill="url(#blue)" style="mix-blend-mode:difference"/>
    <rect x="${rim}" y="${rim}" width="${width - 2 * rim}" height="${height - 2 * rim}" rx="${props.radius}" fill="#808080" style="filter:blur(3px)"/>
  </svg>`;
  return "data:image/svg+xml," + encodeURIComponent(svg);
});
</script>

<template>
  <div ref="surface" class="liquid-glass" aria-hidden="true"
    :style="{ borderRadius: radius + 'px', '--refraction': displacement ? 'url(#' + id + ')' : 'blur(12px)' }">
    <svg v-if="displacement" width="0" height="0" focusable="false">
      <defs>
        <filter :id="id" color-interpolation-filters="sRGB">
          <feImage :href="displacement" width="100%" height="100%" result="map" />
          <feDisplacementMap in="SourceGraphic" in2="map" scale="-24" xChannelSelector="R" yChannelSelector="B" />
        </filter>
      </defs>
    </svg>
  </div>
</template>

<style lang="scss" scoped>
.liquid-glass {
  position: absolute;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  overflow: hidden;
  background: var(--glass-tint, rgb(30 34 62 / 52%));
  border: 1px solid rgb(255 255 255 / 38%);
  box-shadow: inset 0 1px 0 #ffffff40, 0 8px 32px #1d163622;
  -webkit-backdrop-filter: blur(12px) saturate(1.25);
  backdrop-filter: var(--refraction) blur(9px) saturate(1.25);
}
</style>
