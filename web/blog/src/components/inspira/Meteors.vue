<!-- Adapted from Inspira UI Meteors. See public/licenses/inspira-ui.txt. -->
<script setup>
import { computed } from "vue";
const props = defineProps({ count: { type: Number, default: 9 }, shower: Boolean });
const trails = computed(() => Array.from({ length: props.count }, (_, index) => props.shower ? {
  left: `${32 + (index * 31) % 65}%`,
  top: `calc(64px + ${(index * 7) % 23}%)`,
  '--cycle': '5.4s',
  '--delay': `${-5.4 * (index + .65) / props.count}s`,
  '--core': index % 4 === 0 ? '3px' : '2px',
  '--angle': `${-30 - index % 7}deg`,
  '--tint': ['#ffe2bb', '#e2deff', '#c7e7ff'][index % 3],
  '--reach': [.96, .72, 1.08, .83][index % 4],
} : {
  left: `${18 + (index * 37) % 91}%`,
  top: `${(index * 13) % 35 - 5}%`,
  '--delay': `${-(index * 1.73 + .6)}s`,
  '--cycle': `${6.8 + (index % 4) * .85}s`,
}));
</script>

<template>
  <div class="meteor-shower" :class="{ 'sky-rain': shower }" aria-hidden="true">
    <i v-for="(trail, index) in trails" :key="index" :style="trail" />
  </div>
</template>

<style lang="scss" scoped>
.meteor-shower {
  pointer-events: none;
  position: absolute;
  inset: 0;
  overflow: hidden;

  i {
    position: absolute;
    width: 2px;
    height: 2px;
    opacity: 0;
    border-radius: 50%;
    background: #fff8eb;
    box-shadow: 0 0 5px #ffefdf;
    animation: meteor var(--cycle) linear var(--delay) infinite;
    &::before {
      content: "";
      position: absolute;
      top: 1px;
      left: 1px;
      width: var(--meteor-tail, clamp(40px, 6vw, 100px));
      height: 1px;
      background: linear-gradient(to right, #fff2e1df, #f5e7fb66 42%, transparent);
    }
  }
}
.sky-rain i {
  width: var(--core);
  height: var(--core);
  background: #fffdf2;
  box-shadow: 0 0 3px white, 0 0 8px 1px var(--tint);
  animation-name: lake-meteor;
  &::before {
    top: 0;
    height: var(--core);
    width: calc(var(--meteor-tail) * var(--reach));
    clip-path: polygon(0 4%, 100% 50%, 0 96%);
    background: linear-gradient(90deg, #fffcf3, var(--tint) 12%, transparent 98%);
  }
  &:nth-child(4n + 1)::after {
    content: "";
    position: absolute;
    inset: -2px -8px;
    border-radius: 50%;
    background: radial-gradient(ellipse, #fff9e5 4%, var(--tint) 18%, transparent 65%);
  }
}
@keyframes lake-meteor {
  0% { transform: rotate(var(--angle)) translateX(0) scaleX(.55); opacity: 0; }
  6%, 23% { opacity: 1; }
  38%, 100% { transform: rotate(var(--angle)) translateX(calc(-1 * var(--meteor-distance) * var(--reach))) scaleX(1.15); opacity: 0; }
}

@keyframes meteor {
  0% { transform: rotate(-33deg) translateX(0); opacity: 0; }
  5%, 15% { opacity: .9; }
  26%, 100% { transform: rotate(-33deg) translateX(calc(-1 * var(--meteor-distance, 48vmax))); opacity: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .meteor-shower { display: none; }
}
</style>
