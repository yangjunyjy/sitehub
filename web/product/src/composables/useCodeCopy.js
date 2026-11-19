import { onBeforeUnmount } from "vue";

// A timer belongs to one button; copying a second block must not strand the first.
export function useCodeCopy() {
  const timers = new Map();
  let revision = 0;

  const reset = () => {
    revision++;
    for (const [button, timer] of timers) {
      clearTimeout(timer);
      if (button.isConnected) button.textContent = "复制";
    }
    timers.clear();
  };

  const copy = async (event) => {
    const button = event.target.closest?.(".code-copy-btn");
    if (!button || button.disabled) return;
    const code = button.closest(".code-block-wrapper")?.querySelector("code");
    if (!code) return;
    const current = revision;
    clearTimeout(timers.get(button));
    button.disabled = true;
    let message;
    try {
      await navigator.clipboard.writeText(code.textContent || "");
      message = "已复制";
    } catch {
      message = "复制失败，请手动选择";
    } finally {
      button.disabled = false;
    }
    if (current !== revision || !button.isConnected) return;
    button.textContent = message;
    timers.set(
      button,
      setTimeout(() => {
        if (button.isConnected) button.textContent = "复制";
        timers.delete(button);
      }, 1800),
    );
  };

  onBeforeUnmount(reset);
  return { copyCode: copy, resetCodeCopy: reset };
}
