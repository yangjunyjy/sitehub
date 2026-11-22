<script setup>
import { onMounted, onBeforeUnmount, ref } from "vue";
import BlogIcon from "@/components/BlogIcon.vue";

const props = defineProps({ modalOpen: Boolean });
const dock = ref(null),
  trigger = ref(null),
  expanded = ref(false);
let leaveTimer;
function open() {
  clearTimeout(leaveTimer);
  leaveTimer = undefined;
  expanded.value = true;
}
function close() {
  clearTimeout(leaveTimer);
  leaveTimer = undefined;
  if (!props.modalOpen) expanded.value = false;
}
function move(event) {
  if (event.pointerType !== "mouse" || props.modalOpen) return;
  if (
    event.clientX >= document.documentElement.clientWidth - 10 ||
    dock.value?.contains(event.target)
  ) {
    open();
  } else if (
    expanded.value &&
    !leaveTimer &&
    !dock.value?.querySelector(":focus-visible")
  ) {
    leaveTimer = setTimeout(() => {
      leaveTimer = undefined;
      close();
    }, 180);
  }
}
function pointerOutside(event) {
  if (!dock.value?.contains(event.target)) close();
}
function leave(event) {
  if (event.pointerType === "mouse" && !dock.value?.querySelector(":focus-visible")) close();
}
function toggle() {
  if (expanded.value) close();
  else open();
}
function escape() {
  expanded.value = false;
  trigger.value?.focus({ preventScroll: true });
}
function focusOut(event) {
  if (!dock.value?.contains(event.relatedTarget)) close();
}
onMounted(() => {
  document.addEventListener("pointermove", move, { passive: true });
  document.addEventListener("pointerdown", pointerOutside, { passive: true });
  document.documentElement.addEventListener("pointerleave", leave);
});
onBeforeUnmount(() => {
  clearTimeout(leaveTimer);
  document.removeEventListener("pointermove", move);
  document.removeEventListener("pointerdown", pointerOutside);
  document.documentElement.removeEventListener("pointerleave", leave);
});
</script>
<template>
  <Teleport to="body">
    <aside
      ref="dock"
      class="reading-dock"
      aria-label="文章互动"
      :class="{ 'is-open': expanded }"
      @focusout="focusOut"
      @keydown.esc.stop="escape"
    >
      <button
        ref="trigger"
        class="dock-trigger"
        type="button"
        aria-label="文章互动工具"
        aria-controls="reader-action-panel"
        :aria-expanded="expanded"
        @click="toggle"
      >
        <span class="dock-mark" aria-hidden="true"></span
        ><BlogIcon name="heart" :size="18" />
      </button>
      <Transition name="dock"
        ><div
          v-show="expanded"
          id="reader-action-panel"
          class="dock-panel"
          :inert="!expanded"
        >
          <slot /></div
      ></Transition>
    </aside>
  </Teleport>
</template>
<style scoped>
.reading-dock {
  position: fixed;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 26;
  display: flex;
  align-items: center;
}
.dock-trigger {
  order: 2;
  position: relative;
  width: 15px;
  height: 84px;
  flex: none;
  padding: 0;
  border: 0;
  border-radius: 8px 0 0 8px;
  background: transparent;
  color: var(--journal-accent);
}
.dock-mark {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  background: #b99580;
  width: 4px;
  height: 38px;
  border-radius: 4px 0 0 4px;
}
.dock-trigger:focus-visible {
  outline-offset: -2px;
  background: var(--journal-wash);
}
.dock-trigger svg {
  display: none;
}
.dock-panel {
  display: grid;
  gap: 3px;
  width: 128px;
  max-height: calc(100dvh - 132px);
  overflow-y: auto;
  padding: 5px;
  background: #fff;
  border: 1px solid var(--journal-rule);
  border-radius: 10px;
  box-shadow: 0 4px 18px #30273519;
}
.dock-enter-active,
.dock-leave-active {
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}
.dock-enter-from,
.dock-leave-to {
  opacity: 0;
  transform: translateX(7px);
}
@media (hover: none), (pointer: coarse) {
  .reading-dock {
    top: auto;
    bottom: max(82px, calc(72px + env(safe-area-inset-bottom)));
    transform: none;
    align-items: end;
  }
  .dock-trigger {
    width: max(44px, calc(37px + env(safe-area-inset-right)));
    height: 46px;
    display: grid;
    place-items: center;
    border: 1px solid var(--journal-rule);
    border-right: 0;
    background: var(--journal-paper);
    box-shadow: 0 2px 8px #30273519;
  }
  .dock-trigger svg {
    display: block;
  }
  .dock-trigger .dock-mark {
    display: none;
  }
  .dock-panel {
    margin-right: 4px;
  }
}
@media (max-height: 350px) {
  .reading-dock {
    top: 50%;
    bottom: auto;
    transform: translateY(-50%);
  }
  .dock-panel {
    max-height: 60dvh;
  }
}
</style>
