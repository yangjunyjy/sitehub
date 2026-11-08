<template>
  <div class="markdown-editor" :class="{ disabled }">
    <div v-if="error" class="editor-fallback">
      <el-alert :title="error" type="error" :closable="false" />
      <textarea
        :value="modelValue"
        :disabled="disabled"
        aria-label="Markdown 源码"
        @input="emit('update:modelValue', $event.target.value)"
      />
    </div>
    <div v-show="!error" ref="host" />
  </div>
</template>
<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import Vditor from 'vditor'
import DOMPurify from 'dompurify'
import 'vditor/dist/index.css'
const props = defineProps({ modelValue: { type: String, default: '' }, disabled: Boolean })
const emit = defineEmits(['update:modelValue', 'save', 'upload', 'ready'])
const host = ref(),
  error = ref('')
let editor,
  startupTimer,
  savedRange,
  disposed = false,
  ready = false,
  applying = false,
  renderedValue,
  lastEmitted = props.modelValue
function flush() {
  if (!ready) return
  const value = editor.getValue()
  // Vditor delays its input callback. Flush pending edits before saving, but preserve
  // the original imported Markdown if rendering only normalized its whitespace.
  if (value !== renderedValue) {
    renderedValue = value
    lastEmitted = value
    emit('update:modelValue', value)
  }
}
function rememberSelection() {
  const selection = window.getSelection()
  if (selection?.rangeCount && host.value?.contains(selection.anchorNode))
    savedRange = selection.getRangeAt(0).cloneRange()
}
onMounted(() => {
  document.addEventListener('selectionchange', rememberSelection)
  startupTimer = setTimeout(() => {
    if (!ready && !disposed) error.value = '编辑器资源加载超时，已切换源码输入。内容仍可保存。'
  }, 15000)
  try {
    editor = new Vditor(host.value, {
      cdn: '/vendor/vditor',
      mode: 'ir',
      lang: 'zh_CN',
      value: props.modelValue,
      cache: { enable: false },
      height: '100%',
      minHeight: 360,
      placeholder: '开始书写，或粘贴 Markdown…',
      outline: { enable: false },
      toolbar: [
        'headings',
        'bold',
        'italic',
        'strike',
        '|',
        'list',
        'ordered-list',
        'check',
        'quote',
        '|',
        'code',
        'inline-code',
        'link',
        'table',
        '|',
        'undo',
        'redo',
        '|',
        'edit-mode',
        'outline',
      ].map((name) => (name === '|' ? name : { name, tipPosition: 's' })),
      toolbarConfig: { pin: true },
      preview: {
        delay: 150,
        markdown: { sanitize: true, mathBlockPreview: false },
        transform: (html) => DOMPurify.sanitize(html),
        theme: { path: '/vendor/vditor/dist/css/content-theme' },
        render: { media: { enable: false } },
      },
      hint: { emojiPath: '/vendor/vditor/dist/images/emoji' },
      upload: {
        accept: 'image/png,image/jpeg,image/gif,image/webp',
        handler: async (files) => {
          emit('upload', files)
          return null
        },
      },
      input(value) {
        if (!applying) {
          renderedValue = value
          lastEmitted = value
          emit('update:modelValue', value)
        }
      },
      after() {
        clearTimeout(startupTimer)
        if (disposed || error.value) {
          editor.destroy()
          return
        }
        ready = true
        renderedValue = editor.getValue()
        if (props.disabled) editor.disabled()
        emit('ready')
      },
    })
  } catch {
    clearTimeout(startupTimer)
    error.value = '编辑器加载失败，已切换源码输入。内容仍可保存。'
  }
})
watch(
  () => props.modelValue,
  (value) => {
    if (ready && value !== lastEmitted) {
      applying = true
      editor.setValue(value, true)
      renderedValue = editor.getValue()
      lastEmitted = value
      applying = false
    }
  },
)
watch(
  () => props.disabled,
  (value) => {
    if (ready) value ? editor.disabled() : editor.enable()
  },
)
onBeforeUnmount(() => {
  document.removeEventListener('selectionchange', rememberSelection)
  disposed = true
  clearTimeout(startupTimer)
  // Vditor bootstraps asynchronously. Destroying it before `after` invalidates
  // its pending initialization; let that callback clean up an abandoned mount.
  if (ready) editor?.destroy()
  ready = false
})
defineExpose({
  flush,
  insert: (text) => {
    if (ready) {
      editor.enable()
      const pane = host.value.querySelector(
        '.vditor-ir [contenteditable], .vditor-wysiwyg [contenteditable]',
      )
      if (
        savedRange?.startContainer.isConnected &&
        pane?.contains(savedRange.startContainer) &&
        savedRange.startContainer !== pane
      ) {
        editor.focus()
        const selection = window.getSelection()
        selection.removeAllRanges()
        selection.addRange(savedRange)
        editor.insertMD(text)
        lastEmitted = editor.getValue()
      } else {
        // Without an active text cursor, append the source as a new block. Bare nodes at the IR root are omitted by Lute.
        lastEmitted = props.modelValue + text
        editor.setValue(lastEmitted)
      }
      renderedValue = editor.getValue()
      emit('update:modelValue', lastEmitted)
      if (props.disabled) editor.disabled()
    } else emit('update:modelValue', props.modelValue + text)
  },
  focus: () => editor?.focus(),
})
</script>
<style scoped>
.markdown-editor {
  min-width: 0;
  flex: 1;
  min-height: 360px;
}
.markdown-editor :deep(.vditor) {
  border: 0;
  border-radius: 0;
  --border-color: #ebeef5;
  --toolbar-background-color: #fafbfd;
  --toolbar-icon-color: #606266;
  --toolbar-icon-hover-color: #409eff;
  --panel-background-color: #fff;
  --textarea-background-color: #fff;
  min-height: 360px;
}
.markdown-editor :deep(.vditor-reset) {
  box-sizing: border-box;
  width: 100%;
  max-width: 1240px;
  margin-inline: auto;
  font-size: 16px;
  line-height: 1.9;
  padding: 28px var(--document-gutter, 48px) 48px !important;
  color: #303133;
}
.markdown-editor :deep(.vditor-toolbar) {
  position: sticky;
  top: var(--studio-header-height, 0px);
  z-index: 10;
  display: flex;
  align-items: center;
  border-top: 1px solid #ebeef5;
  border-bottom: 1px solid #ebeef5;
  padding: 7px 20px !important;
  flex-wrap: wrap;
  gap: 2px;
  background: #fafbfd;
  box-shadow: 0 3px 6px #0000000d;
}
.markdown-editor :deep(.vditor-toolbar__item) {
  float: none;
  flex: 0 0 auto;
  padding: 0;
}
.markdown-editor :deep(.vditor-toolbar__item > button) {
  border-radius: 5px;
  height: 30px;
  width: 30px;
  padding: 7px;
  color: #606266;
  transition:
    background 0.15s,
    color 0.15s;
}
.markdown-editor :deep(.vditor-toolbar__item > button:hover) {
  background: #f5f7fa;
  color: #409eff;
}
.markdown-editor :deep(.vditor-toolbar__item svg) {
  width: 15px;
  height: 15px;
}
.markdown-editor :deep(.vditor-toolbar__divider) {
  float: none;
  flex: 0 0 auto;
  margin: 8px 6px;
  height: 14px;
  background: #f5f7fa;
}
.markdown-editor :deep(.vditor-ir pre.vditor-reset[contenteditable]:empty:before) {
  color: #909399;
  font-size: 15px;
}
.markdown-editor :deep(.vditor-reset h1),
.markdown-editor :deep(.vditor-reset h2) {
  color: #303133;
}
.markdown-editor :deep(.vditor-outline) {
  background: #fbfcfe;
  border-left: 1px solid #ebeef5;
  font-size: 12px;
}
.markdown-editor :deep(.vditor-content) {
  min-height: 360px;
}
.disabled {
  pointer-events: none;
}
.editor-fallback textarea {
  box-sizing: border-box;
  width: 100%;
  min-height: 360px;
  border: 0;
  padding: 24px;
  font: 15px/1.8 monospace;
}
@media (max-width: 700px) {
  .markdown-editor :deep(.vditor-toolbar) {
    padding: 6px 10px !important;
    gap: 0;
  }
  .markdown-editor :deep(.vditor-toolbar__item > button) {
    width: 28px;
    height: 29px;
    padding: 7px;
  }
  .markdown-editor :deep(.vditor-toolbar__divider) {
    margin-inline: 4px;
  }
  .markdown-editor :deep(.vditor-reset) {
    padding: 22px var(--document-gutter, 22px) 36px !important;
    font-size: 15px;
  }
  .markdown-editor :deep(.vditor-outline) {
    display: none !important;
  }
}
</style>
