<script setup>
import { computed, onBeforeUnmount, ref } from "vue";
const selected = ref(0),
  step = ref(0),
  running = ref(false),
  copied = ref(false);
const examples = [
  {
    label: "应用发布",
    file: "release.yaml",
    title: "一次发布，有序完成。",
    code: [
      "name: web-release",
      "environment: production",
      "",
      "pipeline:",
      "  - checkout: main",
      "  - build: container",
      "  - deploy: kubernetes",
      "  - verify: health-check",
    ],
    stages: ["获取代码与构建产物", "执行部署编排", "检查应用健康状态"],
  },
  {
    label: "自动化运维",
    file: "playbook.yaml",
    title: "重复工作，交给自动化。",
    code: [
      "name: server-maintenance",
      "targets: web-servers",
      "",
      "tasks:",
      "  - check: system-status",
      "  - backup: configuration",
      "  - update: application",
      "  - report: execution-log",
    ],
    stages: ["检查目标环境", "执行运维任务", "生成执行记录"],
  },
  {
    label: "权限与审计",
    file: "policy.yaml",
    title: "每次操作，都有迹可循。",
    code: [
      "name: production-policy",
      "scope: production",
      "",
      "access:",
      "  role: maintainer",
      "  approval: required",
      "",
      "audit: enabled",
    ],
    stages: ["校验访问权限", "确认审批策略", "记录操作审计"],
  },
];
const current = computed(() => examples[selected.value]);
let timer, copyTimer;
function select(index) {
  clearTimeout(timer);
  clearTimeout(copyTimer);
  copied.value = false;
  selected.value = index;
  step.value = 0;
  running.value = false;
}
function navigateTabs(event) {
  const keys = ["ArrowLeft", "ArrowRight", "Home", "End"];
  if (!keys.includes(event.key)) return;
  event.preventDefault();
  select(
    event.key === "Home"
      ? 0
      : event.key === "End"
        ? examples.length - 1
        : (selected.value +
            (event.key === "ArrowRight" ? 1 : -1) +
            examples.length) %
          examples.length,
  );
  event.currentTarget.querySelectorAll('[role="tab"]')[selected.value]?.focus();
}
function run() {
  clearTimeout(timer);
  step.value = 0;
  running.value = true;
  const advance = () => {
    step.value++;
    if (step.value < 3) timer = setTimeout(advance, 650);
    else running.value = false;
  };
  timer = setTimeout(advance, 450);
}
async function copy() {
  try {
    await navigator.clipboard.writeText(current.value.code.join("\n"));
    copied.value = true;
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => (copied.value = false), 1600);
  } catch {
    copied.value = false;
  }
}
onBeforeUnmount(() => {
  clearTimeout(timer);
  clearTimeout(copyTimer);
});
</script>
<template>
  <div class="home-demo" aria-label="Deploy 交互演示">
    <div class="demo-editor">
      <header class="demo-window-bar">
        <span class="window-dots" aria-hidden="true"><i /><i /><i /></span
        ><span>./{{ current.file }}</span
        ><button
          type="button"
          @click="copy"
          :aria-label="copied ? '已复制配置示例' : '复制配置示例'"
        >
          {{ copied ? "已复制" : "复制" }}
        </button>
      </header>
      <div
        class="demo-tabs"
        role="tablist"
        aria-label="选择演示场景"
        @keydown="navigateTabs"
      >
        <button
          v-for="(example, index) in examples"
          :key="example.file"
          :id="`demo-tab-${index}`"
          type="button"
          role="tab"
          :tabindex="selected === index ? 0 : -1"
          :aria-selected="selected === index"
          aria-controls="demo-code-panel"
          :class="{ active: selected === index }"
          @click="select(index)"
        >
          {{ example.label }}
        </button>
      </div>
      <div
        id="demo-code-panel"
        role="tabpanel"
        :aria-labelledby="`demo-tab-${selected}`"
        class="demo-code"
      >
        <div v-for="(line, index) in current.code" :key="index">
          <span class="line-number" aria-hidden="true">{{ index + 1 }}</span
          ><code
            ><span class="code-key">{{
              line.includes(":") ? line.slice(0, line.indexOf(":") + 1) : line
            }}</span
            >{{
              line.includes(":") ? line.slice(line.indexOf(":") + 1) : ""
            }}</code
          >
        </div>
      </div>
      <footer>
        <span><i />配置示例</span><span>YAML</span>
      </footer>
    </div>
    <div class="demo-result">
      <header>
        <span class="demo-result-label">DEPLOY WORKFLOW</span
        ><span class="demo-status" :class="{ 'is-complete': step === 3 }">{{
          running ? "演示进行中" : step === 3 ? "演示完成" : "准备就绪"
        }}</span>
      </header>
      <h2>{{ current.title }}</h2>
      <ol>
        <li
          v-for="(label, index) in current.stages"
          :key="label"
          :class="{ done: step > index, pending: running && step === index }"
        >
          <span>{{
            step > index ? "✓" : String(index + 1).padStart(2, "0")
          }}</span
          >{{ label
          }}<small>{{
            step > index ? "完成" : running && step === index ? "执行中" : ""
          }}</small>
        </li>
      </ol>
      <footer>
        <span>本地交互演示</span
        ><button type="button" :disabled="running" @click="run">
          <svg viewBox="0 0 16 16" aria-hidden="true">
            <path d="m5 3 8 5-8 5Z" /></svg
          >{{ running ? "正在演示" : step === 3 ? "重新演示" : "运行演示" }}
        </button>
      </footer>
      <span class="demo-announcement" aria-live="polite">{{
        step === 3 ? "当前流程演示完成" : ""
      }}</span>
    </div>
  </div>
</template>
<style scoped>
.home-demo {
  position: relative;
  width: 100%;
  padding: 0 48px 170px 0;
  isolation: isolate;
}
.demo-editor,
.demo-result {
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  border-radius: 8px;
  box-shadow: 0 8px 32px #00000012;
  overflow: hidden;
}
.demo-window-bar {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  color: var(--deploy-muted);
  font: 12px var(--site-mono);
}
.window-dots {
  display: flex;
  gap: 7px;
}
.window-dots i {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ff5f57;
  border: 1px solid #0001;
}
.window-dots i:nth-child(2) {
  background: #ffbd2e;
}
.window-dots i:nth-child(3) {
  background: #28c840;
}
.demo-window-bar button {
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  padding: 6px;
  font-size: 11px;
}
.demo-tabs {
  display: flex;
  padding: 0 20px;
  border-bottom: 1px solid var(--deploy-line);
  gap: 20px;
}
.demo-tabs button {
  padding: 12px 0;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 12px;
  white-space: nowrap;
}
.demo-tabs .active {
  color: var(--deploy-primary);
  border-bottom-color: var(--deploy-primary);
}
.demo-code {
  padding: 20px 16px 24px;
  min-height: 264px;
  font: 13px/27px var(--site-mono);
  color: var(--deploy-ink);
  overflow: auto;
}
.demo-code > div {
  display: flex;
  min-height: 27px;
  white-space: pre;
}
.line-number {
  color: var(--deploy-faint);
  width: 32px;
  flex-shrink: 0;
  text-align: right;
  padding-right: 14px;
  user-select: none;
}
.code-key {
  color: var(--deploy-primary);
}
.demo-editor > footer {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--deploy-paper);
  font: 11px var(--site-mono);
  color: var(--deploy-muted);
}
.demo-editor > footer span:first-child {
  display: flex;
  align-items: center;
  gap: 6px;
}
.demo-editor > footer i {
  width: 6px;
  height: 6px;
  background: #6fcce1;
  border-radius: 50%;
}
.demo-result {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 86%;
  padding: 22px 26px;
  z-index: 1;
}
.demo-result > header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
.demo-result-label {
  font: 10px var(--site-mono);
  letter-spacing: 0.09em;
  color: var(--deploy-muted);
}
.demo-status {
  font-size: 10px;
  padding: 3px 8px;
  border-radius: 12px;
  background: var(--deploy-paper);
  color: var(--deploy-muted);
}
.demo-status.is-complete {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
h2 {
  font-size: 21px;
  line-height: 1.5;
  font-weight: 500;
  margin: 18px 0;
}
ol {
  padding: 0;
  margin: 0;
  list-style: none;
}
li {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  font-size: 12px;
  color: var(--deploy-muted);
}
li > span {
  width: 22px;
  height: 22px;
  border: 1px solid var(--deploy-line);
  border-radius: 50%;
  display: grid;
  place-items: center;
  font: 10px var(--site-mono);
}
li.done > span {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  border-color: transparent;
}
li.pending > span {
  border-color: var(--deploy-primary);
  color: var(--deploy-primary);
}
li small {
  margin-left: auto;
  font-size: 10px;
  color: var(--deploy-primary);
}
.demo-result > footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--deploy-line);
  color: var(--deploy-faint);
  font-size: 10px;
}
.demo-result > footer button {
  display: flex;
  align-items: center;
  gap: 5px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  border: 0;
  border-radius: 5px;
  padding: 6px 10px;
  font-size: 11px;
}
.demo-result > footer svg {
  width: 12px;
  height: 12px;
  fill: currentColor;
}
.demo-announcement {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
}
@media (max-width: 760px) {
  .home-demo {
    max-width: 620px;
    margin: auto;
    padding-right: 24px;
    padding-bottom: 172px;
  }
  .demo-result {
    width: 90%;
    padding: 18px 20px;
  }
  .demo-code {
    font-size: 12px;
  }
  .demo-tabs {
    gap: 14px;
    padding-inline: 14px;
  }
  h2 {
    font-size: 19px;
  }
}
@media (max-width: 380px) {
  .home-demo {
    padding-right: 12px;
  }
  .demo-result {
    width: 95%;
    padding: 16px;
  }
  .demo-tabs {
    gap: 12px;
  }
  .demo-result-label {
    font-size: 9px;
  }
}
</style>
