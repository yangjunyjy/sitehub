<script setup>
import { ref } from "vue";
const dialog = ref(null);
const screenshot = `${import.meta.env.BASE_URL}images/product/workbench.png`;
</script>
<template>
  <figure class="product-showcase">
    <div class="showcase-bar">
      <span class="showcase-dots" aria-hidden="true"><i></i><i></i><i></i></span
      ><span>Deploy / 统一工作台</span
      ><span class="showcase-live"><i></i> 产品实景</span>
    </div>
    <button
      class="showcase-image"
      type="button"
      aria-label="放大查看产品工作台截图"
      @click="dialog.showModal()"
    >
      <img
        :src="screenshot"
        width="1600"
        height="632"
        alt="Deploy 实际工作台：执行概览、近七日执行趋势、任务状态分布与平台资源统计"
        fetchpriority="high"
      /><span>放大查看 </span>
    </button>
    <figcaption>
      <span>一个工作台，掌握执行进度与资源状态。</span
      ><span>产品界面截图 · 非实时数据</span>
    </figcaption>
  </figure>
  <dialog
    ref="dialog"
    class="showcase-dialog"
    aria-label="产品工作台截图"
    @click="
      (event) => {
        if (event.target === dialog) dialog.close();
      }
    "
  >
    <button type="button" autofocus @click="dialog.close()">关闭预览 ×</button
    ><img
      :src="screenshot"
      width="1600"
      height="632"
      alt="Deploy 产品工作台完整截图"
    />
    <p>产品界面实景 · 数据为截图时状态</p>
  </dialog>
</template>
<style scoped>
.product-showcase {
  margin: 0;
  overflow: hidden;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
  box-shadow:
    0 24px 70px #00000012,
    0 3px 12px #00000012;
  text-align: left;
}
.showcase-bar {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 18px;
  border-bottom: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  color: var(--deploy-muted);
  font-size: 10px;
}
.showcase-dots {
  display: flex;
  gap: 6px;
  min-width: 90px;
}
.showcase-dots i {
  width: 7px;
  height: 7px;
  border-radius: 100%;
  background: var(--deploy-paper);
}
.showcase-dots i:first-child {
  background: #edc7c7;
}
.showcase-dots i:nth-child(2) {
  background: #efdfb8;
}
.showcase-dots i:last-child {
  background: #bcdccd;
}
.showcase-live {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 90px;
  justify-content: flex-end;
}
.showcase-live i {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #65b098;
}
.showcase-image {
  display: block;
  width: 100%;
  position: relative;
  padding: 0;
  border: 0;
  background: var(--deploy-paper);
}
.showcase-image img {
  width: 100%;
  height: auto;
}
.showcase-image > span {
  position: absolute;
  right: 16px;
  bottom: 16px;
  background: var(--deploy-white);
  padding: 6px 12px;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  color: var(--deploy-primary);
  opacity: 0;
  transition: opacity 0.2s;
  font-size: 12px;
}
.showcase-image:hover > span,
.showcase-image:focus-visible > span {
  opacity: 1;
}
.product-showcase figcaption {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 13px 18px;
  border-top: 1px solid var(--deploy-line);
  color: var(--deploy-muted);
  font-size: 11px;
}
.product-showcase figcaption span:last-child {
  color: var(--deploy-faint);
}
.showcase-dialog {
  position: fixed;
  max-width: calc(100vw - 40px);
  width: 1500px;
  max-height: 90dvh;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  padding: 16px;
  background: var(--deploy-paper);
}
.showcase-dialog::backdrop {
  background: var(--deploy-soft);
  backdrop-filter: blur(5px);
}
.showcase-dialog > button {
  display: block;
  margin: 0 0 16px auto;
  background: var(--deploy-white);
  border: 1px solid var(--deploy-line);
  border-radius: 7px;
  padding: 8px 16px;
  color: var(--deploy-muted);
}
.showcase-dialog img {
  width: 100%;
  height: auto;
}
.showcase-dialog p {
  text-align: center;
  font-size: 12px;
  color: var(--deploy-muted);
  margin: 16px 0 0;
}
@media (max-width: 600px) {
  .showcase-bar {
    padding-inline: 10px;
    font-size: 9px;
  }
  .showcase-dots,
  .showcase-live {
    min-width: 50px;
  }
  .product-showcase {
    border-radius: 9px;
  }
  .product-showcase figcaption {
    padding: 10px;
    font-size: 10px;
    flex-direction: column;
    gap: 2px;
  }
  .showcase-image > span {
    right: 8px;
    bottom: 8px;
    opacity: 1;
    font-size: 10px;
    padding: 3px 7px;
  }
  .showcase-dialog {
    max-width: calc(100vw - 16px);
    padding: 10px;
  }
}
</style>
