<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { useNavigationState } from "@/composables/useNavigationState";
const navigation = useNavigationState();
const visible = ref(false);
let timer;
watch(
  () => navigation.pending,
  (pending) => {
    clearTimeout(timer);
    if (pending)
      timer = setTimeout(() => {
        visible.value = true;
      }, 160);
    else visible.value = false;
  },
);
onBeforeUnmount(() => clearTimeout(timer));
const retry = () =>
  window.location.assign(navigation.target || window.location.href);
</script>

<template>
  <div
    v-if="visible"
    class="route-progress"
    role="progressbar"
    aria-label="正在打开页面"
  >
    <span />
  </div>
  <div v-if="navigation.error" class="route-error" role="alert">
    <span>{{ navigation.error }}</span
    ><button type="button" @click="retry">重新打开</button>
  </div>
</template>

<style scoped>
.route-progress {
  position: fixed;
  inset: 0 0 auto;
  height: 3px;
  z-index: 10000;
  background: var(--deploy-soft);
  overflow: hidden;
}
.route-progress span {
  display: block;
  width: 35%;
  height: 100%;
  background: var(--deploy-primary);
  animation: route-progress 1.1s ease-in-out infinite;
}
.route-error {
  position: fixed;
  left: 50%;
  top: 90px;
  transform: translateX(-50%);
  z-index: 10000;
  display: flex;
  align-items: center;
  gap: 16px;
  width: max-content;
  max-width: calc(100% - 32px);
  padding: 14px 18px;
  border: 1px solid #efcbc5;
  border-radius: 10px;
  background: var(--deploy-paper);
  color: #9c4038;
  box-shadow: var(--deploy-shadow);
  font-size: 14px;
}
.route-error button {
  flex-shrink: 0;
  min-height: 36px;
  padding: 6px 10px;
  border: 1px solid #efcbc5;
  border-radius: 6px;
  background: var(--deploy-white);
  color: inherit;
}
@keyframes route-progress {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(390%);
  }
}
</style>
