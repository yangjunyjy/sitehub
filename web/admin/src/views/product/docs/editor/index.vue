<template>
  <div class="document-studio" :class="{ 'is-immersive': immersive }" v-loading="loading">
    <header ref="header" class="document-toolbar">
      <div class="document-context">
        <el-button
          :icon="ArrowLeft"
          aria-label="返回文档列表"
          @click="router.push(location('product/docs/index', { group: groupId }))"
        /><span>{{ version?.name || '文档' }}</span
        ><el-tag :type="version?.status === 'published' ? 'success' : 'info'" size="small">{{
          version?.status === 'published' ? '已发布版本' : '草稿版本'
        }}</el-tag
        ><span class="save-state" role="status">{{ dirty ? '有未保存修改' : documentStatus[recordStatus] || '新文档' }}</span>
      </div>
      <div class="document-actions">
        <el-button
          :icon="RefreshRight"
          aria-label="重新加载文档"
          :disabled="saving || uploading"
          @click="reload"
        />
        <el-button v-if="documentRevision" @click="openHistory" :disabled="saving || loading"
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
          :disabled="loading || !!error || !!conflict || !version || mode === 'preview'"
          @click="submit"
          >保存草稿</el-button>
        <el-button v-if="documentRevision && recordStatus !== 'published'" :loading="saving"
          :disabled="!!conflict || dirty || !form.content.trim()" @click="changePublication(true)">发布文档</el-button>
        <el-button v-if="recordStatus && recordStatus !== 'draft'" :disabled="!!conflict || dirty || saving" @click="changePublication(false)">撤回文档</el-button>
      </div>
    </header>
    <div v-if="error" class="document-error">
      <el-alert :title="error" type="error" :closable="false" /><el-button @click="reload"
        >重新加载</el-button
      >
    </div>
    <div v-if="conflict" class="document-error" role="alert">
      <el-alert title="此文档已有更新，当前编辑内容已保留。" type="warning" :closable="false" />
      <el-button @click="compareConflict">查看服务器版本</el-button>
    </div>
    <main v-if="!error && version && initialized && !loading" class="document-sheet">
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
          <span>{{ recordStatus === 'modified' ? '当前草稿有修改，公开页面继续显示已发布版本。' : '保存到草稿后，可单独发布此文档。' }}</span>
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
          <p class="form-note">可在文档工作区复制版本。</p></el-form-item
        ><el-form-item label="所属分组" required
          ><el-select v-model="groupId"
            ><el-option
              v-for="group in groups"
              :key="group.id"
              :value="group.id"
              :label="group.title" /></el-select
          ><router-link v-if="!groups.length" :to="location('product/docs/index')"
            >先创建分组</router-link
          ></el-form-item
        ><el-form-item label="访问标识" required
          ><el-input v-model="form.slug" placeholder="quick-start" maxlength="100" />
          <p class="form-note">同一产品版本内不能重复，保存时检查是否已被占用。</p></el-form-item
        ></el-form
      >
    </el-drawer>
    <el-drawer v-model="conflictVisible" title="对照服务器版本" size="min(760px, 100vw)">
      <p class="form-note">先对照服务器修订，再决定如何保存当前编辑。</p>
      <template v-if="remote">
        <h3>{{ remote.title }}</h3><p>{{ remote.summary }}</p>
        <pre class="remote-source">{{ remote.content }}</pre>
      </template>
      <template #footer>
        <el-button @click="conflictVisible = false">继续编辑</el-button>
        <el-button type="primary" :disabled="!remote" @click="rebase">保留本地稿并采用此修订号</el-button>
      </template>
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
import { documentPayload, documentStatus, validateDocument } from '@/utils/documents'
import { renderArticle, validateMarkdownFile } from '@/utils/article-editor'
import { createUUID } from '@/utils/uuid'
import { getAdminDocument, saveDocument, getDocumentHistory, getDocumentRevision, publishDocument, withdrawDocument } from '@/api/documents'
import { uploadImage } from '@/api/upload'
const route = useRoute(),
  router = useRouter()
const { loading, saving, error, load, version, groups, location } = useDocumentWorkspace()
const versionID = computed(() => version.value?.id)
const conflict = ref(false), conflictVisible = ref(false), remote = ref(null), recordStatus = ref('')
const form = reactive({ id: '', title: '', slug: '', summary: '', content: '' }),
  groupId = ref(''),
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
const documentRevision = ref(0)
const groupTitle = computed(() => groups.value.find(g => g.id === groupId.value)?.title)
let initializeID = 0,
  resizeObserver
const snapshot = () => JSON.stringify({ ...form, group: groupId.value })
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
  documentRevision.value = original?.revision || 0
  recordStatus.value = original?.status || ''
  conflict.value = false
  conflictVisible.value = false
  historyVisible.value = false
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
  groupId.value = original?.groupId || groups.value.find(group => String(group.id) === route.query.group)?.id || groups.value[0]?.id || ''
  mode.value = route.query.view === 'preview' ? 'preview' : 'edit'
  editorKey.value++
  baseline.value = snapshot()
  initialized.value = true
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
let promoting = false
onBeforeRouteUpdate(async (to, from) =>
  promoting || (to.query.id === from.query.id && to.query.version === from.query.version && to.query.mode === from.query.mode) || await confirmLeave())
watch([() => route.query.id, () => route.query.version, () => route.query.mode], () => {
  if (!promoting && route.meta.componentPath === 'product/docs/editor/index') initialize()
}, { immediate: true })
function togglePreview() {
  form.content = editor.value?.flush() ?? form.content
  mode.value = mode.value === 'preview' ? 'edit' : 'preview'
}
async function submit() {
  if (saving.value || uploading.value || loading.value || conflict.value || mode.value === 'preview') return
  editor.value?.flush()
  const submitted = snapshot()
  saving.value = true
  try {
    if (!form.slug || !form.title.trim() || !groupId.value) {
      showSettings.value = true
      throw new Error('请填写文档标题、分组和访问标识')
    }
    const result = await saveDocument({
      ...validateDocument(form), version: versionID.value, groupId: groupId.value, revision: documentRevision.value,
    })
    documentRevision.value = result.data.revision
    recordStatus.value = result.data.status
    baseline.value = submitted
    ElMessage.success('草稿已保存')
    if (!route.query.id) {
      promoting = true
      try { await router.replace(location('product/docs/editor/index', { id: form.id, group: groupId.value })) }
      finally { promoting = false }
    }
  } catch (failure) {
    conflict.value = failure.response?.status === 409 && !!documentRevision.value
    ElMessage.error(failure.message || '保存失败')
  } finally { saving.value = false }
}
async function changePublication(publish) {
  editor.value?.flush()
  if (saving.value || conflict.value || dirty.value) return
  try {
    await ElMessageBox.confirm(publish ? '将当前保存的正文发布给读者？' : '将此文档从公开目录撤回？', publish ? '发布文档' : '撤回文档',
      { confirmButtonText: '确认', cancelButtonText: '取消' })
  } catch { return }
  if (dirty.value || saving.value) return
  saving.value = true
  try {
    const action = publish ? publishDocument : withdrawDocument
    const result = await action({ id: form.id, version: versionID.value, revision: documentRevision.value })
    documentRevision.value = result.data.revision
    recordStatus.value = result.data.status
    ElMessage.success(publish ? '文档已发布' : '文档已撤回')
  } catch (failure) {
    conflict.value = failure.response?.status === 409
    ElMessage.error(failure.message)
  } finally { saving.value = false }
}
async function compareConflict() {
  try {
    remote.value = (await getAdminDocument(form.id, versionID.value)).data
    conflictVisible.value = true
  } catch (failure) { ElMessage.error(failure.message) }
}
function rebase() {
  documentRevision.value = remote.value.revision
  recordStatus.value = remote.value.status
  baseline.value = JSON.stringify({ ...documentPayload(remote.value), group: remote.value.groupId })
  conflict.value = false
  conflictVisible.value = false
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
    selectedHistory.value = (await getDocumentRevision(form.id, versionID.value, entry.id)).data
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
  const group = groups.value.find(g => g.title === item.groupTitle)
  if (group) groupId.value = group.id
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
<style lang="scss" scoped src="./editor.scss"></style>
