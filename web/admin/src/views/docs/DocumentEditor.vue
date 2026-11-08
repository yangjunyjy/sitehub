<template>
  <div class="document-studio" :class="{ 'is-immersive': immersive }" v-loading="loading">
    <header ref="header" class="document-toolbar">
      <div class="document-context">
        <el-button
          :icon="ArrowLeft"
          aria-label="返回文档列表"
          @click="router.push(location('/dashboard/docs'))"
        /><span>{{ version?.name || '文档' }}</span
        ><el-tag :type="version?.status === 'published' ? 'success' : 'info'" size="small">{{
          version?.status === 'published' ? '已发布版本' : '草稿版本'
        }}</el-tag
        ><span class="save-state" role="status">{{ dirty ? '有未保存修改' : '已同步' }}</span>
      </div>
      <div class="document-actions">
        <el-button
          :icon="RefreshRight"
          aria-label="重新加载文档"
          :disabled="saving || uploading"
          @click="reload"
        />
        <el-button v-if="route.query.id" @click="openHistory" :disabled="saving || loading"
          >修订记录</el-button
        >
        <el-button :icon="View" @click="togglePreview">{{
          mode === 'preview' ? '返回编辑' : '阅读预览'
        }}</el-button
        ><el-button :icon="Setting" @click="showSettings = true">文档设置</el-button
        ><el-button
          :icon="FullScreen"
          :aria-label="immersive ? '退出沉浸模式' : '沉浸模式'"
          @click="immersive = !immersive"
        /><el-button
          type="primary"
          :icon="Check"
          :loading="saving"
          :disabled="loading || !!error || !version || mode === 'preview'"
          @click="submit"
          >{{ version?.status === 'published' ? '保存并更新' : '保存文档' }}</el-button
        >
      </div>
    </header>
    <div v-if="error" class="document-error">
      <el-alert :title="error" type="error" :closable="false" /><el-button @click="reload"
        >重新加载</el-button
      >
    </div>
    <main v-else-if="version && initialized && !loading" class="document-sheet">
      <div class="document-title">
        <div class="document-eyebrow">
          <span>{{ groupTitle || '未选择分组' }}</span
          ><span>{{ form.content.length.toLocaleString() }} 字符</span>
        </div>
        <template v-if="mode !== 'preview'"
          ><el-input
            v-model="form.title"
            class="title-input"
            placeholder="请输入文档标题"
            maxlength="100"
            aria-label="文档标题" /><el-input
            v-model="form.summary"
            class="summary-input"
            placeholder="添加简短介绍，帮助读者找到所需内容（选填）"
            maxlength="300"
            aria-label="文档摘要"
        /></template>
        <template v-else
          ><h1>{{ form.title || '未命名文档' }}</h1>
          <p>{{ form.summary }}</p></template
        >
        <div class="document-meta">
          <span>{{
            version.status === 'published'
              ? '保存后将更新此公开版本的内容'
              : '草稿仅管理端可见，发布版本后供读者访问'
          }}</span>
          <div v-if="mode !== 'preview'">
            <el-button text :icon="Upload" @click="importInput.click()">导入 Markdown</el-button
            ><el-button text :icon="Picture" :loading="uploading" @click="imageInput.click()"
              >插入图片</el-button
            >
          </div>
          <el-button v-else text :icon="Download" @click="exportMarkdown">导出 Markdown</el-button>
        </div>
      </div>
      <div v-show="mode !== 'preview'" class="document-writing">
        <MarkdownEditor
          :key="editorKey"
          ref="editor"
          v-model="form.content"
          :disabled="saving || uploading"
          @save="submit"
          @upload="uploadImages"
        />
      </div>
      <div v-if="mode === 'preview'" class="document-reading">
        <nav v-if="rendered.headings.length" aria-label="文档目录">
          <button
            v-for="heading in rendered.headings"
            :key="heading.slug"
            @click="jump(heading.slug)"
          >
            {{ heading.text }}
          </button>
        </nav>
        <MarkdownContent :html="rendered.html" @click="copyCode" />
      </div>
    </main>
    <input
      ref="importInput"
      type="file"
      accept=".md,.markdown"
      hidden
      @change="importMarkdown"
    /><input ref="imageInput" type="file" accept="image/*" hidden multiple @change="selectImages" />
    <el-drawer v-model="showSettings" title="文档设置" size="min(420px, 100vw)">
      <el-form label-position="top"
        ><el-form-item label="产品版本"
          ><el-input :model-value="version?.name" disabled />
          <p class="form-note">当前文档属于此版本。复制版本请前往版本管理。</p></el-form-item
        ><el-form-item label="所属分组" required
          ><el-select v-model="groupTitle"
            ><el-option
              v-for="group in version?.groups"
              :key="group.title"
              :value="group.title"
              :label="group.title" /></el-select
          ><router-link v-if="!version?.groups.length" :to="location('/dashboard/docs/groups')"
            >先创建分组</router-link
          ></el-form-item
        ><el-form-item label="访问标识" required
          ><el-input v-model="form.slug" placeholder="quick-start" maxlength="100" />
          <p class="form-note">/docs/{{ form.slug || 'quick-start' }}?version={{ versionID }}</p>
          <p class="form-note">同一产品版本内不能重复，保存时检查是否已被占用。</p></el-form-item
        ></el-form
      >
    </el-drawer>
    <el-drawer v-model="historyVisible" title="文档修订记录" size="min(760px, 100vw)">
      <p class="form-note">
        显示最近 100 次保存。载入历史内容后可继续编辑，点击保存才会生成新的修订。
      </p>
      <el-alert v-if="historyError" :title="historyError" type="error" :closable="false" />
      <div v-loading="historyLoading" class="document-history-list">
        <el-empty
          v-if="!historyLoading && !histories.length && !historyError"
          description="暂无修订记录"
        />
        <el-button
          v-for="entry in histories"
          :key="entry.id"
          :disabled="historyLoading"
          :class="{ selected: selectedHistory?.id === entry.id }"
          @click="previewHistory(entry)"
        >
          #{{ entry.revision }} · {{ new Date(entry.createdAt).toLocaleString() }}
        </el-button>
      </div>
      <template v-if="selectedHistory">
        <h3>{{ selectedHistory.title }}</h3>
        <p class="form-note">
          修订 #{{ selectedHistory.revision }} ·
          {{ selectedHistory.operation === 'migration' ? '迁移时保存的初始内容' : '历史保存内容' }}
        </p>
        <el-button type="primary" @click="loadHistoricalContent">载入此修订</el-button>
        <MarkdownContent :html="renderArticle(selectedHistory.contentMarkdown).html" />
      </template>
    </el-drawer>
  </div>
</template>
<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter, onBeforeRouteUpdate } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft,
  Check,
  Download,
  FullScreen,
  Picture,
  RefreshRight,
  Setting,
  Upload,
  View,
} from '@element-plus/icons-vue'
import MarkdownEditor from '@/components/content/MarkdownEditor.vue'
import MarkdownContent from '@/components/content/MarkdownContent.vue'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import { useUnsavedChanges } from '@/composables/useAdminResource'
import { documentPayload, validateDocument } from '@/utils/documents'
import { renderArticle, validateMarkdownFile } from '@/utils/article-editor'
import { createUUID } from '@/utils/uuid'
import { getAdminDocument, saveDocument, getDocumentHistory } from '@/api/documents'
import { uploadImage } from '@/api/upload'
const route = useRoute(),
  router = useRouter()
const { loading, saving, error, load, save, version, versionID, location } = useDocumentWorkspace()
const form = reactive({ id: '', title: '', slug: '', summary: '', content: '' }),
  groupTitle = ref(''),
  baseline = ref('')
const mode = ref('edit'),
  initialized = ref(false),
  immersive = ref(false),
  showSettings = ref(false),
  editor = ref(),
  editorKey = ref(0),
  header = ref(),
  headerHeight = ref('72px'),
  importInput = ref(),
  imageInput = ref(),
  uploading = ref(false)
let documentRevision = 0,
  initializeID = 0,
  resizeObserver
const snapshot = () => JSON.stringify({ ...form, group: groupTitle.value })
const dirty = computed(() => !!baseline.value && snapshot() !== baseline.value)
const confirmLeave = useUnsavedChanges(() => dirty.value || saving.value || uploading.value)
const rendered = computed(() => renderArticle(form.content))
async function initialize() {
  const current = ++initializeID
  initialized.value = false
  await load()
  if (current !== initializeID || error.value) return
  if (!version.value) {
    error.value = '此文档版本不存在，请返回版本管理'
    return
  }
  let original = null
  try {
    if (route.query.id)
      original = (await getAdminDocument(String(route.query.id), versionID.value)).data
  } catch (e) {
    if (current === initializeID) error.value = e.message || '文档加载失败'
    return
  }
  if (current !== initializeID) return
  documentRevision = original?.revision || 0
  if (route.query.id && !original) {
    error.value = '文档不存在或已被删除'
    return
  }
  Object.assign(
    form,
    original
      ? documentPayload(original)
      : { id: `doc-${createUUID()}`, title: '', slug: '', summary: '', content: '' },
  )
  groupTitle.value =
    version.value.groups.find((g) => g.id === original?.groupId)?.title ||
    version.value.groups[0]?.title ||
    ''
  mode.value = route.query.mode === 'preview' ? 'preview' : 'edit'
  editorKey.value++
  baseline.value = snapshot()
  initialized.value = true
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
onBeforeRouteUpdate(
  async (to, from) =>
    (to.query.id === from.query.id && to.query.version === from.query.version) ||
    (await confirmLeave()),
)
watch(() => [route.query.id, route.query.version], initialize, { immediate: true })
function togglePreview() {
  form.content = editor.value?.flush() ?? form.content
  mode.value = mode.value === 'preview' ? 'edit' : 'preview'
}
async function submit() {
  if (saving.value || uploading.value || loading.value || mode.value === 'preview') return
  form.content = editor.value?.flush() ?? form.content
  const submitted = snapshot()
  const ok = await save(
    async () => {
      if (!form.slug || !form.title.trim() || !groupTitle.value) {
        showSettings.value = true
        throw new Error('请填写文档标题、分组和访问标识')
      }
      const payload = validateDocument(form)
      const target = version.value.groups.find((group) => group.title === groupTitle.value)
      if (!target) throw new Error('分组不存在，请重新加载')
      const result = await saveDocument({
        ...payload,
        version: versionID.value,
        groupId: target.id,
        revision: documentRevision,
      })
      documentRevision = result.data.revision
      baseline.value = submitted
    },
    version.value?.status === 'published' ? '当前版本的公开文档已更新' : '草稿文档已保存',
  )
  if (ok && !route.query.id)
    await router.replace(location('/dashboard/docs/editor', { id: form.id }))
}
const historyVisible = ref(false),
  historyLoading = ref(false),
  historyError = ref(''),
  histories = ref([]),
  selectedHistory = ref(null)
async function openHistory() {
  historyVisible.value = true
  historyLoading.value = true
  historyError.value = ''
  selectedHistory.value = null
  try {
    histories.value = (await getDocumentHistory(form.id, versionID.value)).data
  } catch (e) {
    historyError.value = e.message || '修订记录加载失败'
  } finally {
    historyLoading.value = false
  }
}
async function previewHistory(entry) {
  historyLoading.value = true
  historyError.value = ''
  selectedHistory.value = null
  try {
    selectedHistory.value = (await getDocumentHistory(form.id, versionID.value, entry.id)).data
  } catch (e) {
    historyError.value = e.message || '修订内容加载失败'
  } finally {
    historyLoading.value = false
  }
}
async function loadHistoricalContent() {
  if (!(await confirmLeave())) return
  const item = selectedHistory.value
  Object.assign(form, {
    title: item.title,
    slug: item.slug,
    summary: item.summary,
    content: item.contentMarkdown,
  })
  if (version.value.groups.some((g) => g.title === item.groupTitle))
    groupTitle.value = item.groupTitle
  mode.value = 'edit'
  editorKey.value++
  historyVisible.value = false
}
async function importMarkdown(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  try {
    validateMarkdownFile(file)
    if (form.content.trim())
      await ElMessageBox.confirm('导入将替换当前正文。', '导入 Markdown', {
        confirmButtonText: '替换',
        cancelButtonText: '取消',
      })
    form.content = await file.text()
    if (!form.title) form.title = file.name.replace(/\.(md|markdown)$/i, '')
  } catch (e) {
    if (e instanceof Error) ElMessage.error(e.message)
  }
}
async function uploadImages(files) {
  if (uploading.value) return
  uploading.value = true
  try {
    for (const file of Array.from(files || [])) {
      if (!file.type.startsWith('image/')) throw new Error('请选择图片文件')
      const payload = new FormData()
      payload.append('file', file)
      const result = await uploadImage(payload),
        url = result.data?.url
      if (!url) throw new Error('上传成功但未返回图片地址')
      editor.value?.insert(
        `\n![图片](${String(url).replace(/\(/g, '%28').replace(/\)/g, '%29')})\n`,
      )
    }
  } catch (e) {
    ElMessage.error(e.message || '图片上传失败')
  } finally {
    uploading.value = false
  }
}
function selectImages(event) {
  const files = Array.from(event.target.files || [])
  event.target.value = ''
  uploadImages(files)
}
function exportMarkdown() {
  const url = URL.createObjectURL(
      new Blob([form.content], { type: 'text/markdown;charset=utf-8' }),
    ),
    link = document.createElement('a')
  link.href = url
  link.download = `${form.slug || 'document'}.md`
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
function jump(id) {
  document
    .querySelector(`.document-reading [id="${CSS.escape(id)}"]`)
    ?.scrollIntoView({ block: 'start', behavior: 'smooth' })
}
async function copyCode(event) {
  const button = event.target.closest('.code-copy-btn')
  if (!button) return
  try {
    await navigator.clipboard.writeText(
      button.closest('.code-block-wrapper')?.querySelector('code')?.textContent || '',
    )
    ElMessage.success('代码已复制')
  } catch {
    ElMessage.warning('复制失败，请手动选择代码')
  }
}
const keydown = (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
    event.preventDefault()
    submit()
  }
  if (event.key === 'Escape') immersive.value = false
}
onMounted(async () => {
  window.addEventListener('keydown', keydown)
  await nextTick()
  resizeObserver = new ResizeObserver(
    () => (headerHeight.value = `${header.value?.offsetHeight || 72}px`),
  )
  if (header.value) resizeObserver.observe(header.value)
})
onBeforeUnmount(() => {
  initializeID++
  resizeObserver?.disconnect()
  window.removeEventListener('keydown', keydown)
})
</script>
<style scoped>
.document-history-list {
  display: grid;
  gap: 8px;
  max-height: 280px;
  overflow-y: auto;
  margin: 20px 0;
}
.document-history-list .el-button {
  margin: 0;
  min-height: 40px;
  justify-content: flex-start;
}
.document-history-list .selected {
  color: #409eff;
  border-color: #ebeef5;
  background: #f5f7fa;
}

.document-studio {
  --el-color-primary: #409eff;
  --el-color-primary-light-3: #8494f3;
  --el-color-primary-light-9: #f2f4ff;
  --studio-header-height: v-bind(headerHeight);
  --document-gutter: 40px;
  min-height: 100%;
  background: #f5f7fa;
  color: #303133;
}
.document-toolbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  flex-wrap: wrap;
}
.document-context,
.document-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.document-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
.save-state {
  color: #606266;
  font-size: 12px;
}
.document-sheet {
  margin: 24px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fff;
  min-height: 70vh;
}
.document-title {
  padding: 28px 40px 20px;
}
.document-eyebrow {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #606266;
  margin-bottom: 18px;
}
.title-input :deep(.el-input__wrapper),
.summary-input :deep(.el-input__wrapper) {
  box-shadow: none !important;
  padding: 0;
  background: transparent;
}
.title-input :deep(input) {
  height: 52px;
  font-size: 30px;
  font-weight: 650;
  color: #303133;
}
.summary-input {
  margin-top: 12px;
}
.summary-input :deep(input) {
  color: #606266;
}
.document-title h1 {
  font-size: 30px;
  margin: 12px 0;
  overflow-wrap: anywhere;
}
.document-title p {
  color: #606266;
}
.document-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 18px;
  font-size: 12px;
  color: #606266;
}
.document-reading {
  padding: 0 40px 60px;
  max-width: 1100px;
  margin: auto;
}
.document-reading nav {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin: 0 0 24px;
  border-top: 1px solid #ebeef5;
  padding-top: 18px;
}
.document-reading nav button {
  border: 0;
  background: #f5f7fa;
  color: #409eff;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
}
.document-reading :deep([id]) {
  scroll-margin-top: calc(var(--studio-header-height) + 20px);
}
.document-error {
  padding: 24px;
}
.is-immersive {
  position: fixed;
  inset: 0;
  z-index: 1800;
  overflow-y: auto;
}
@media (max-width: 800px) {
  .document-studio {
    --document-gutter: 20px;
  }
  .document-toolbar {
    padding: 12px;
    gap: 10px;
  }
  .document-actions {
    gap: 6px;
  }
  .document-actions :deep(.el-button) {
    padding-inline: 10px;
  }
  .document-sheet {
    margin: 12px;
  }
  .document-title {
    padding: 22px 20px 16px;
  }
  .document-reading {
    padding: 0 20px 36px;
  }
  .title-input :deep(input) {
    font-size: 24px;
  }
  .save-state {
    display: none;
  }
}
</style>
