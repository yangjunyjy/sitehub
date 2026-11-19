<!-- Adapted from Inspira UI Meteors. See public/licenses/inspira-ui.txt. -->
<script setup>
import { computed } from "vue";
const props = defineProps({ count: { type: Number, default: 9 } });
// Stable positions across renders keep a typing visitor's sky still.
const trails = computed(() => Array.from({ length: props.count }, () => ({
  left: Math.random() * 115 - 15 + "%",
  top: Math.random() * 48 + "%",
  animationDelay: -(Math.random() * 12) + "s",
  animationDuration: 7 + Math.random() * 7 + "s",
})));
</script>

<template>
  <div class="meteor-shower" aria-hidden="true">
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
    width: 3px;
    height: 3px;
    opacity: 0;
    border-radius: 50%;
    background: #fbfcff;
    box-shadow: 0 0 7px #e1f5ff;
    animation: meteor linear infinite;
    &::before {
      content: "";
      position: absolute;
      top: 1px;
      width: clamp(40px, 6vw, 100px);
      height: 1px;
      background: linear-gradient(to right, #ecf5ff, transparent);
    }
  }
}

@keyframes meteor {
  0%, 62% { transform: rotate(-33deg) translateX(0); opacity: 0; }
  65% { opacity: .85; }
  78%, 100% { transform: rotate(-33deg) translateX(-48vmax); opacity: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .meteor-shower { display: none; }
}
</style>
