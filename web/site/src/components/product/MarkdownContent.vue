<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { renderMermaidBlocks } from "@/utils/mermaid";
import { useCodeCopy } from "@/composables/useCodeCopy";

// 只接受 utils/markdown 生成的、已经净化的 HTML。
const props = defineProps({ html: { type: String, default: "" } });
const content = ref(null);
const { copyCode, resetCodeCopy } = useCodeCopy();
let revision = 0;
watch(
  [content, () => props.html],
  () => {
    const current = ++revision;
    resetCodeCopy();
    if (content.value) {
      for (const image of content.value.querySelectorAll("img")) {
        image.loading = "lazy";
        image.decoding = "async";
      }
      void renderMermaidBlocks(content.value, () => current === revision);
    }
  },
  { flush: "post" },
);
onBeforeUnmount(() => revision++);
</script>

<template>
  <article ref="content" @click="copyCode" v-html="html" />
</template>
