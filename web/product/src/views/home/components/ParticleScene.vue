<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import { particleScenes } from "@shared/product/particle-scenes";
import { createParticleField } from "../particle-field";

const props = defineProps({
  shape: { type: String, required: true },
  autoplay: { type: Boolean, required: true },
  interactive: { type: Boolean, required: true },
});
const canvas = ref(null);
const currentShape = ref(props.shape);
const scene = computed(() => particleScenes[currentShape.value] || particleScenes.artifact);
const reducedMotion = ref(false);
const playing = computed(() => props.autoplay && !reducedMotion.value);
const available = ref(true);
const dragging = ref(false);
let field, observer, sizeObserver, themeObserver, preference;
let frame = 0,
  last = 0,
  inView = true,
  drag;

function paint(time) {
  frame = 0;
  if (!field || !inView || document.hidden) return;
  const elapsed = last ? Math.min((time - last) / 1000, 0.06) : 1 / 60;
  last = time;
  currentShape.value = field.draw(elapsed, playing.value);
  if (playing.value) frame = requestAnimationFrame(paint);
}

function invalidate() {
  if (!frame && field && inView && !document.hidden)
    frame = requestAnimationFrame(paint);
}

function suspend() {
  cancelAnimationFrame(frame);
  frame = 0;
  last = 0;
}

function visibility() {
  if (document.hidden || !inView) suspend();
  else invalidate();
}

function setPreference() {
  reducedMotion.value = preference.matches;
}

watch(playing, () => {
  suspend();
  field?.clearPointer();
  invalidate();
});

function startDrag(event) {
  if (!props.interactive || !field || !event.isPrimary || event.button !== 0) return;
  drag = { id: event.pointerId, x: event.clientX, y: event.clientY };
  dragging.value = true;
  canvas.value.setPointerCapture(event.pointerId);
}

function move(event) {
  if (!props.interactive || !field || !event.isPrimary) return;
  if (drag) {
    field.turn(
      (event.clientX - drag.x) * 0.0065,
      (event.clientY - drag.y) * 0.006,
    );
    drag.x = event.clientX;
    drag.y = event.clientY;
  }
  if (playing.value) field.pointer(event.offsetX, event.offsetY);
  invalidate();
}

function leave() {
  field?.clearPointer();
  invalidate();
}

function stopDrag(event) {
  if (drag?.id !== event.pointerId) return;
  drag = undefined;
  dragging.value = false;
  leave();
}

function rotate(x, y) {
  if (!props.interactive) return;
  field?.turn(x, y);
  invalidate();
}

onMounted(() => {
  field = createParticleField(canvas.value, props.shape);
  if (!field) {
    available.value = false;
    return;
  }
  preference = window.matchMedia("(prefers-reduced-motion: reduce)");
  setPreference();
  preference.addEventListener("change", setPreference);
  const readTheme = () => {
    field.theme(document.documentElement.classList.contains("dark"));
    invalidate();
  };
  readTheme();
  themeObserver = new MutationObserver(readTheme);
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["class"],
  });
  sizeObserver = new ResizeObserver(() => {
    field.resize();
    invalidate();
  });
  sizeObserver.observe(canvas.value);
  observer = new IntersectionObserver(([entry]) => {
    inView = entry.isIntersecting;
    visibility();
  });
  observer.observe(canvas.value);
  document.addEventListener("visibilitychange", visibility);
});

onBeforeUnmount(() => {
  suspend();
  preference?.removeEventListener("change", setPreference);
  document.removeEventListener("visibilitychange", visibility);
  sizeObserver?.disconnect();
  observer?.disconnect();
  themeObserver?.disconnect();
});
</script>

<template>
  <figure class="particle-scene">
    <div class="plot">
      <div v-if="!available" class="scene-fallback" role="img" :aria-label="scene.description">
        <ResourceIcon :name="scene.icon" :size="144" />
      </div>
      <canvas
        v-show="available"
        ref="canvas"
        :tabindex="available && interactive ? 0 : undefined"
        role="img"
        :aria-label="scene.description + (interactive ? '，可拖动或使用方向键旋转' : '')"
        :class="{ dragging, interactive: interactive && available }"
        @pointerdown="startDrag"
        @pointermove="move"
        @pointerleave="leave"
        @pointerup="stopDrag"
        @pointercancel="stopDrag"
        @lostpointercapture="stopDrag"
        @keydown.left.prevent="rotate(-0.24, 0)"
        @keydown.right.prevent="rotate(0.24, 0)"
        @keydown.up.prevent="rotate(0, -0.18)"
        @keydown.down.prevent="rotate(0, 0.18)"
      />
    </div>
  </figure>
</template>

<style scoped lang="scss">
.particle-scene {
  min-width: 0;
  margin: 0;
  display: grid;
  align-content: center;
}
.plot {
  position: relative;
  height: clamp(330px, 38vw, 560px);
}
.scene-fallback {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: var(--exhibit-accent);
}
canvas {
  width: 100%;
  height: 100%;
  display: block;
}
canvas.interactive {
  touch-action: pan-y;
  cursor: grab;
}
canvas.dragging {
  cursor: grabbing;
}
@media (max-width: 720px) {
  .plot {
    height: clamp(224px, 70vw, 350px);
  }
}
</style>
