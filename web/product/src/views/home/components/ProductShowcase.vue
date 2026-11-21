<script setup>
import { ref } from 'vue'
defineProps({ brand: { type: String, required: true } })
const dialog = ref(null)
const fullSize = ref(false)
const screenshot = import.meta.env.BASE_URL + 'images/product/workbench.png'
function openPreview() {
  fullSize.value = false
  dialog.value.showModal()
}
</script>
<template>
  <figure class="product-showcase">
    <button class="showcase-image" type="button" aria-label="放大查看产品工作台截图" @click="openPreview">
      <img :src="screenshot" width="1600" height="632" loading="lazy" decoding="async"
        :alt="brand + ' 工作台，展示执行趋势、任务状态与平台资源统计'" />
      <span class="showcase-expand" aria-hidden="true">⤢</span>
    </button>
  </figure>
  <dialog ref="dialog" class="showcase-dialog" aria-labelledby="preview-heading"
    @click="event => { if (event.target === dialog) dialog.close() }">
    <div class="showcase-dialog__body">
      <header>
        <h2 id="preview-heading">工作台截图</h2>
        <button type="button" :aria-pressed="fullSize" @click="fullSize = !fullSize">{{ fullSize ? '适合窗口' : '原始尺寸' }}</button>
        <button type="button" autofocus @click="dialog.close()">关闭</button>
      </header>
      <div class="showcase-dialog__image" tabindex="0" aria-label="工作台截图预览">
        <img :src="screenshot" width="1600" height="632" :class="{ original: fullSize }" :alt="brand + ' 产品工作台完整截图'" />
      </div>
    </div>
  </dialog>
</template>
<style scoped>
.product-showcase { margin: 0; }
.showcase-image {
  display: block;
  position: relative;
  width: 100%;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--workbench-line);
  border-radius: 7px;
  background: var(--workbench-paper);
  cursor: zoom-in;
}
.showcase-image img { width: 100%; height: auto; }
.showcase-expand {
  position: absolute;
  bottom: 10px;
  right: 10px;
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--workbench-line);
  border-radius: 5px;
  background: var(--workbench-paper);
  color: var(--workbench-ink);
  font: 20px/1 system-ui;
}
.showcase-image:hover, .showcase-image:focus-visible { border-color: var(--workbench-accent); }
.showcase-dialog {
  width: min(1440px, calc(100vw - 32px));
  max-width: none;
  max-height: calc(100dvh - 32px);
  margin: auto;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--workbench-line);
  border-radius: 8px;
  background: var(--workbench-paper);
  color: var(--workbench-ink);
}
.showcase-dialog::backdrop { background: #181712bf; }
.showcase-dialog__body header {
  display: flex; align-items: center; gap: 10px;
  border-bottom: 1px solid var(--workbench-line);
  padding: 10px 14px;
}
.showcase-dialog h2 { flex: 1; margin: 0; font-size: 15px; }
.showcase-dialog button {
  border: 1px solid var(--workbench-line);
  border-radius: 4px;
  padding: 5px 10px;
  background: transparent;
  color: inherit;
  font-size: 13px;
}
.showcase-dialog button:hover { border-color: var(--workbench-accent); }
.showcase-dialog__image { max-height: calc(100dvh - 91px); overflow: auto; }
.showcase-dialog__image img { width: 100%; height: auto; }
.showcase-dialog__image img.original { width: 1600px; max-width: none; }
</style>
