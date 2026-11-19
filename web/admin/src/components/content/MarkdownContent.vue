<template><div class="article-prose" v-html="html" /></template>
<script setup>
// Only pass HTML from the shared, sanitizing Markdown renderer.
defineProps({ html: { type: String, default: '' } })
</script>
<style lang="scss" scoped>
.article-prose {
  font-size: 14px;
  line-height: 1.9;
  overflow-wrap: anywhere;
}
.article-prose :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}
.article-prose :deep(h1),
.article-prose :deep(h2),
.article-prose :deep(h3) {
  color: #303133;
  margin-block: 1.4em 0.7em;
}
.article-prose :deep(a) {
  color: var(--studio-accent, #409eff);
  text-decoration: none;
}
.article-prose :deep(blockquote) {
  margin-inline: 0;
  padding: 10px 16px;
  border-left: 3px solid #ebeef5;
  background: #f5f7fa;
  color: #606266;
}
.article-prose :deep(.code-block-wrapper) {
  border: 1px solid #ebeef5;
  border-radius: 7px;
  box-shadow: none;
  overflow: hidden;
  margin-block: 18px;
}
.article-prose :deep(.code-header) {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f5f7fa;
  color: #606266;
  font-size: 11px;
}
.article-prose :deep(.code-copy-btn) {
  border: 0;
  background: none;
  color: #409eff;
  cursor: pointer;
}
.article-prose :deep(pre) {
  overflow: auto;
  background: #f5f7fa;
  padding: 16px;
  border-radius: 0;
  margin: 0;
}
.article-prose :deep(pre code) {
  padding: 0;
  background: transparent;
  font-size: 12px;
}
.article-prose :deep(table) {
  display: block;
  overflow: auto;
  border-collapse: collapse;
  font-size: 13px;
}
.article-prose :deep(td),
.article-prose :deep(th) {
  border: 1px solid #ebeef5;
  padding: 8px 14px;
}
.article-prose :deep(th) {
  background: #f5f7fa;
}
</style>
