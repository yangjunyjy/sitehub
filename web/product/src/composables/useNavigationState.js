import { reactive, readonly } from "vue";

const state = reactive({ pending: false, error: "", target: "" });

export function beginNavigation(target) {
  state.target = target;
  state.pending = true;
  state.error = "";
}

export function finishNavigation(target) {
  if (state.target === target) state.pending = false;
}

export function failNavigation(target) {
  if (state.target !== target) return;
  state.pending = false;
  state.error = "页面资源加载失败，请检查连接后重新打开。";
}

export const useNavigationState = () => readonly(state);
