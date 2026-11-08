<script setup>
import { useTheme } from "@/composables/useTheme";
const { dark, toggleTheme } = useTheme();
</script>
<template>
  <button
    class="theme-toggle"
    type="button"
    role="switch"
    :aria-checked="dark"
    :aria-label="dark ? '切换浅色主题' : '切换深色主题'"
    @click="toggleTheme"
  >
    <span class="theme-toggle__thumb">
      <svg v-if="dark" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M20.8 13.4A9 9 0 0 1 10.6 3.2 9 9 0 1 0 20.8 13.4Z" />
      </svg>
      <svg v-else viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="12" cy="12" r="4" />
        <path
          d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5"
        />
      </svg>
    </span>
  </button>
</template>
<style scoped>
.theme-toggle {
  width: 42px;
  height: 24px;
  padding: 2px;
  border: 1px solid var(--deploy-line);
  border-radius: 20px;
  background: var(--deploy-paper);
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.theme-toggle__thumb {
  display: grid;
  place-items: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--deploy-white);
  color: var(--deploy-muted);
  box-shadow: 0 1px 3px #0002;
  transition: transform 0.2s;
}
.theme-toggle[aria-checked="true"] .theme-toggle__thumb {
  transform: translateX(17px);
}
svg {
  width: 13px;
  height: 13px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;
}
</style>
