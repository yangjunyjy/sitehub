<template>
  <main
    class="article-studio"
    :class="{ immersive, 'with-preview': showPreview }"
    :style="{ '--studio-header-height': `${headerHeight}px` }"
  >
    <header ref="studioHeader" class="studio-header" aria-label="文章编辑操作">
      <div
        class="save-state"
        role="status"
        aria-live="polite"
        :class="{ failed: saveError, saving: busy }"
      >
        <el-icon :class="{ 'is-loading': busy }"
          ><Loading v-if="busy" /><Warning v-else-if="saveError" /><CircleCheck v-else
        /></el-icon>
        <span>{{ saveState }}</span>
      </div>
      <div class="studio-actions">
        <div class="view-actions">
          <el-button
            text
            :icon="View"
            :class="{ 'is-selected': showPreview }"
            :disabled="!ready"
            :aria-pressed="showPreview"
            @click="showPreview = !showPreview"
            >{{ showPreview ? '关闭预览' : '发布预览' }}</el-button
          >
          <el-button text :icon="Setting" :disabled="!ready || busy" @click="showSettings = true"
            >文章设置</el-button
          >
          <el-tooltip :content="immersive ? '退出沉浸写作' : '沉浸写作'" :show-after="400">
            <el-button
              text
              class="icon-button focus-button"
              :icon="immersive ? Close : FullScreen"
              :aria-label="immersive ? '退出沉浸写作' : '沉浸写作'"
              :aria-pressed="immersive"
              @click="immersive = !immersive"
            />
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleMoreCommand">
            <el-button
              text
              class="icon-button"
              :icon="MoreFilled"
              aria-label="更多文章操作"
              :disabled="!ready || busy"
            />
            <template #dropdown
              ><el-dropdown-menu>
                <el-dropdown-item command="import" :icon="Upload" :disabled="!editable"
                  >导入 Markdown</el-dropdown-item
                >
                <el-dropdown-item command="export" :icon="Download">导出 Markdown</el-dropdown-item>
                <el-dropdown-item command="history" :icon="Clock" :disabled="!article.id" divided
                  >历史版本与操作记录</el-dropdown-item
                >
              </el-dropdown-menu></template
            >
          </el-dropdown>
        </div>
        <div class="publish-actions">
          <el-button class="draft-button" :disabled="!editable || busy" @click="save(false)"
            >保存草稿</el-button
          >
          <el-button
            v-if="can('publish')"
            class="publish-button"
            type="primary"
            :icon="Promotion"
            :disabled="!actionable || busy"
            @click="act('publish')"
            >{{ article.status === 1 ? '更新线上版本' : '发布文章' }}</el-button
          >
        </div>
      </div>
    </header>
    <input ref="fileInput" type="file" accept=".md,.markdown" hidden @change="importMarkdown" />
    <input ref="assetInput" type="file" :accept="assetAccept" hidden @change="chooseAsset" />
    <div v-if="loadState === 'loading'" class="studio-empty">
      <el-skeleton :rows="8" animated />
    </div>
    <el-result
      v-else-if="loadState === 'error'"
      icon="error"
      title="文章加载失败"
      :sub-title="loadError"
      ><template #extra
        ><el-button @click="load">重新加载</el-button
        ><el-button @click="router.push(router.page('blog/articles/index'))">返回列表</el-button></template
      ></el-result
    >
    <template v-else>
      <div class="studio-notices">
        <el-alert v-if="saveError" :title="saveError" type="error" :closable="false"
          ><el-button @click="downloadLocal">下载本地稿</el-button
          ><el-button v-if="conflict" @click="compareLatest">查看服务器版本</el-button
          ><el-button v-else @click="save(false)">重试保存</el-button></el-alert
        >
        <el-alert
          v-if="article.status === 1"
          title="你正在编辑草稿。保存不会改变线上文章，点击“更新线上版本”后才会公开。"
          type="info"
          :closable="false"
        />
        <el-alert
          v-if="article.status === 2"
          :title="`计划于 ${formatTime(article.scheduled_at)} 发布已安排的版本；之后保存的修改不会自动替换该版本。`"
          type="info"
          :closable="false"
        />
        <el-alert
          v-if="article.deleted_at"
          title="文章位于回收站，恢复后才能继续编辑。"
          type="warning"
          :closable="false"
          ><el-button v-if="can('delete')" @click="act('restore')">恢复文章</el-button></el-alert
        >
      </div>
      <section class="studio-content">
        <div class="writing-sheet">
          <div v-if="article.cover" class="document-cover">
            <img :src="article.cover" alt="文章封面" />
            <div class="cover-actions">
              <el-button size="small" :disabled="!editable || busy" @click="pickAsset('cover')"
                >更换封面</el-button
              ><el-button size="small" :disabled="!editable || busy" @click="article.cover = ''"
                >移除</el-button
              >
            </div>
          </div>
          <div class="document-heading">
            <div class="document-overline">
              <span class="document-status" :class="`state-${article.status}`"
                ><i />{{ statusLabel }}</span
              >
              <el-button
                v-if="!article.cover"
                text
                size="small"
                :icon="Picture"
                :disabled="!editable || busy"
                @click="pickAsset('cover')"
                >添加封面</el-button
              >
            </div>
            <textarea
              v-model="article.title"
              class="article-title-input"
              rows="1"
              :disabled="!editable || busy"
              maxlength="200"
              placeholder="请输入文章标题"
              aria-label="文章标题"
              @keydown.enter.prevent="editor?.focus()"
            />
            <textarea
              v-model="article.summary"
              class="article-summary-input"
              rows="1"
              maxlength="500"
              :disabled="!editable || busy"
              placeholder="添加一段摘要，让读者快速了解文章（选填）"
              aria-label="文章摘要"
            />
            <div class="writing-meta">
              <span class="author-avatar">{{
                (article.author?.name || currentUserName).slice(0, 1).toUpperCase()
              }}</span>
              <span class="author-name">{{ article.author?.name || currentUserName }}</span>
              <span class="meta-divider" />
              <button
                class="taxonomy-button"
                :disabled="!ready || busy"
                @click="showSettings = true"
              >
                <el-icon><FolderOpened /></el-icon
                >{{
                  options.categories.find((item) => item.id === article.category_id)?.name ||
                  '选择分类'
                }}
              </button>
              <button
                class="taxonomy-button tags-button"
                :disabled="!ready || busy"
                @click="showSettings = true"
              >
                <el-icon><CollectionTag /></el-icon
                >{{ article.tag_ids.length ? `${article.tag_ids.length} 个标签` : '添加标签' }}
              </button>
              <div class="asset-actions">
                <el-button
                  text
                  :icon="Picture"
                  :disabled="!editable || busy"
                  @click="pickAsset('image')"
                  >插入图片</el-button
                ><el-button
                  text
                  :icon="Paperclip"
                  :disabled="!editable || busy"
                  @click="pickAsset('attachment')"
                  >插入附件</el-button
                >
              </div>
            </div>
          </div>
          <el-progress
            v-if="uploadProgress !== null"
            :percentage="uploadProgress"
            :text-inside="true"
            :stroke-width="18"
          />
          <MarkdownEditor
            :key="editorKey"
            ref="editor"
            v-model="article.content"
            :disabled="!editable || busy"
            @upload="(files) => uploadFiles(files, 'image')"
          />
          <footer class="writing-footer">
            <div class="document-count">
              <span>{{ charCount.toLocaleString() }} 字符</span><span class="footer-dot">·</span
              ><span>约 {{ Math.max(1, Math.ceil(charCount / 500)) }} 分钟阅读</span>
            </div>
            <span class="keyboard-hint"
              ><span>即时渲染</span><span class="footer-dot">·</span><kbd>{{ saveShortcut }}</kbd
              ><span>保存</span></span
            >
          </footer>
        </div>
        <aside v-if="showPreview" class="publication-preview">
          <div class="preview-heading">
            <div>
              <span class="preview-label">阅读效果</span>
              <h2>发布预览</h2>
            </div>
            <el-button
              text
              :icon="Close"
              class="icon-button"
              aria-label="关闭发布预览"
              @click="showPreview = false"
            />
          </div>
          <img v-if="article.cover" :src="article.cover" class="preview-cover" alt="文章封面" />
          <h1>{{ article.title || '未命名文章' }}</h1>
          <p class="preview-summary">{{ article.summary }}</p>
          <p class="preview-meta">
            {{ article.author?.name || currentUserName }} ·
            {{
              options.categories.find((item) => item.id === article.category_id)?.name || '未分类'
            }}
            · {{ article.created_at ? formatTime(article.created_at) : '待发布' }}
          </p>
          <nav v-if="rendered.headings.length" class="preview-outline" aria-label="文章目录">
            <button
              v-for="heading in rendered.headings"
              :key="heading.slug"
              :style="{ paddingLeft: `${(heading.level - 1) * 12}px` }"
              @click="jump(heading.slug)"
            >
              {{ heading.text }}
            </button>
          </nav>
          <MarkdownContent :html="rendered.html" @click="copyCode" />
        </aside>
      </section>
    </template>
    <el-drawer
      v-model="showSettings"
      title="文章设置"
      class="article-settings-drawer"
      size="min(420px, 100vw)"
    >
      <template #header
        ><div class="settings-heading">
          <span class="settings-icon"
            ><el-icon><Setting /></el-icon
          ></span>
          <div>
            <h2>文章设置</h2>
            <p>内容信息与发布管理</p>
          </div>
        </div></template
      >
      <el-form label-position="top" :disabled="!editable || busy">
        <div class="settings-section-heading">
          <h3>内容信息</h3>
          <span>让文章更完整</span>
        </div>
        <el-form-item label="文章封面">
          <div class="cover-settings">
            <button
              class="cover-picker"
              :class="{ 'has-cover': article.cover }"
              type="button"
              :disabled="!editable || busy"
              @click="pickAsset('cover')"
            >
              <img v-if="article.cover" :src="article.cover" alt="封面预览" />
              <template v-else
                ><span class="cover-picker-icon"
                  ><el-icon><Picture /></el-icon></span
                ><strong>上传文章封面</strong><span>建议使用横向图片</span></template
              >
              <span v-if="article.cover" class="cover-replace">点击更换封面</span>
            </button>
            <div class="cover-help">
              <small>PNG / JPEG / GIF / WebP · 最大 10 MB</small
              ><el-button v-if="article.cover" link type="danger" @click="article.cover = ''"
                >移除</el-button
              >
            </div>
          </div>
        </el-form-item>
        <el-form-item label="摘要"
          ><el-input
            v-model="article.summary"
            placeholder="简要介绍文章内容，显示在文章列表中"
            type="textarea"
            :rows="4"
            maxlength="500"
            show-word-limit
        /></el-form-item>
        <el-form-item label="URL 别名"
          ><el-input v-model="article.slug" maxlength="200" placeholder="留空由服务器生成" /><small
            >文章公开地址使用固定 ID，修改别名不会改变访问地址。</small
          ></el-form-item
        >
        <el-form-item label="分类"
          ><el-select
            :model-value="article.category_id || undefined"
            @update:model-value="article.category_id = $event || 0"
            filterable
            clearable
            placeholder="选择文章分类"
            ><el-option
              v-for="item in options.categories"
              :key="item.id"
              :label="item.name"
              :value="item.id" /></el-select
        ></el-form-item>
        <el-form-item label="标签"
          ><el-select
            v-model="article.tag_ids"
            multiple
            filterable
            :multiple-limit="100"
            placeholder="选择相关标签"
            ><el-option
              v-for="item in options.tags"
              :key="item.id"
              :label="item.name"
              :value="item.id" /></el-select
        ></el-form-item>
        <el-alert v-if="optionsError" :title="optionsError" type="warning" :closable="false"
          ><el-button @click="loadOptions">重新加载选项</el-button></el-alert
        >
      </el-form>
      <section class="publish-settings">
        <div class="settings-section-heading">
          <h3>发布管理</h3>
          <span class="document-status" :class="`state-${article.status}`"
            ><i />{{ statusLabel }}</span
          >
        </div>
        <template v-if="can('publish') && !article.deleted_at"
          ><label>定时发布时间（{{ timezone }}）</label
          ><el-date-picker
            v-model="scheduleTime"
            type="datetime"
            placeholder="至少一分钟后，最多一年"
            :disabled="busy || article.status === 1"
          /><el-button
            :disabled="!actionable || busy || article.status === 1"
            @click="act('schedule')"
            >安排此版本发布</el-button
          ><el-button v-if="article.status === 2" :disabled="busy" @click="act('cancel-schedule')"
            >取消定时</el-button
          ><el-button
            v-if="article.status === 1"
            type="warning"
            :disabled="busy"
            @click="act('unpublish')"
            >下架文章</el-button
          ></template
        >
        <el-button :icon="Clock" :disabled="!article.id || busy" @click="openHistory"
          >历史版本与操作记录</el-button
        ><el-button :icon="Download" @click="downloadLocal">导出当前 Markdown</el-button>
      </section>
      <template #footer
        ><el-button type="primary" @click="showSettings = false">完成设置</el-button></template
      >
    </el-drawer>
    <el-dialog v-model="historyVisible" title="历史版本与操作记录" width="min(860px, 95vw)">
      <el-table :data="history"
        ><el-table-column prop="version" label="版本" width="80" /><el-table-column label="操作"
          ><template #default="{ row }">{{ actionLabel(row.action) }}</template></el-table-column
        ><el-table-column prop="actor_id" label="操作人 ID" /><el-table-column label="时间"
          ><template #default="{ row }">{{ formatTime(row.created_at) }}</template></el-table-column
        ><el-table-column label="操作"
          ><template #default="{ row }"
            ><el-button link @click="inspectRevision(row)">查看 / 比较</el-button></template
          ></el-table-column
        ></el-table
      >
    </el-dialog>
    <el-dialog v-model="comparisonVisible" title="版本比较" width="min(1100px, 96vw)"
      ><div class="revision-comparison">
        <section>
          <h3>本地编辑内容</h3>
          <pre>{{ article.title }}{{ '\n\n' }}{{ article.content }}</pre>
        </section>
        <section>
          <h3>{{ comparisonLatest ? '服务器最新版本' : `历史版本 ${comparison.version}` }}</h3>
          <pre>{{ comparison.title }}{{ '\n\n' }}{{ comparison.content }}</pre>
        </section>
      </div>
      <template #footer
        ><el-button @click="downloadLocal">下载本地稿</el-button
        ><el-button @click="downloadComparison">下载对比版本原文</el-button
        ><el-button v-if="comparisonLatest" type="primary" @click="useLatest"
          >使用服务器版本</el-button
        ><el-button
          v-else
          :disabled="busy || comparison.content_format !== 'markdown' || !can('update')"
          type="primary"
          @click="restoreRevision"
          >恢复为新草稿版本</el-button
        ></template
      ></el-dialog
    >
  </main>
</template>
<script setup>
import { computed, ref, reactive, watch, onBeforeUnmount, onMounted } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  CircleCheck,
  Loading,
  Warning,
  View,
  Setting,
  FullScreen,
  Close,
  MoreFilled,
  Upload,
  Download,
  Clock,
  Promotion,
  Picture,
  Paperclip,
  FolderOpened,
  CollectionTag,
} from '@element-plus/icons-vue'
import MarkdownContent from '@/components/content/MarkdownContent.vue'
import MarkdownEditor from '@/components/content/MarkdownEditor.vue'
import { useRbacStore } from '@/store/rbac'
import {
  getArticleDetail,
  getArticleOptions,
  createArticle,
  updateArticle,
  articleAction,
  getArticleHistory,
  getArticleRevision,
  uploadArticleAsset,
} from '@/api/article'
import {
  draftFields,
  draftFingerprint,
  newRequestID,
  renderArticle,
  validateMarkdownFile,
  MAX_ARTICLE_LENGTH,
} from '@/utils/article-editor'
import { articleError, articleFilename, saveBlob } from '@/utils/articles'
import 'highlight.js/styles/github.css'
const router = useRouter(),
  route = useRoute(),
  rbac = useRbacStore()
const can = (action) => rbac.hasPermission(`blog:article:${action}`)
const currentUserName = computed(
  () => rbac.currentUser?.nickname || rbac.currentUser?.username || '编辑',
)
const empty = () => ({
  id: null,
  version: 0,
  title: '',
  slug: '',
  summary: '',
  content: '',
  content_format: 'markdown',
  cover: '',
  category_id: 0,
  tag_ids: [],
  status: 0,
})
const article = reactive(empty()),
  options = reactive({ categories: [], tags: [] })
const loadState = ref('loading'),
  loadError = ref(''),
  optionsError = ref(''),
  busy = ref(false),
  saveError = ref(''),
  conflict = ref(false),
  lastSaved = ref(''),
  savedFingerprint = ref('')
const editor = ref(),
  editorKey = ref(0),
  fileInput = ref(),
  assetInput = ref(),
  assetKind = ref('image'),
  uploadProgress = ref(null)
const studioHeader = ref(),
  headerHeight = ref(66)
let headerObserver
const showSettings = ref(false),
  showPreview = ref(false),
  immersive = ref(false),
  scheduleTime = ref(),
  historyVisible = ref(false),
  comparisonVisible = ref(false),
  comparisonLatest = ref(false),
  history = ref([]),
  comparison = ref({})
const ready = computed(() => loadState.value === 'ready'),
  editable = computed(
    () =>
      ready.value &&
      !article.deleted_at &&
      !conflict.value &&
      (article.id ? can('update') : can('create')),
  )
const dirty = computed(() => ready.value && draftFingerprint(article) !== savedFingerprint.value)
const actionable = computed(
  () =>
    ready.value &&
    !article.deleted_at &&
    !conflict.value &&
    (article.id || editable.value) &&
    (!dirty.value || editable.value),
)
const statusLabel = computed(() =>
  article.deleted_at ? '回收站' : { 0: '草稿', 1: '已发布', 2: '定时发布' }[article.status],
)
const charCount = computed(() => Array.from(article.content).length)
const rendered = computed(() =>
  renderArticle(article.content, 'markdown'),
)
const saveState = computed(() =>
  busy.value
    ? '正在处理…'
    : saveError.value
      ? '保存失败，本地内容已保留'
      : dirty.value
        ? '尚有未保存修改'
        : !article.id
          ? '自动保存已开启'
          : lastSaved.value
            ? `已保存 ${lastSaved.value}`
            : '所有更改已保存',
)
const saveShortcut = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘ S' : 'Ctrl S'
function handleMoreCommand(command) {
  if (command === 'import') fileInput.value?.click()
  else if (command === 'export') downloadLocal()
  else if (command === 'history') openHistory()
}
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
const assetAccept = computed(() =>
  assetKind.value === 'attachment'
    ? '.pdf,.txt,.md,.zip,.csv,.json,.yaml,.yml'
    : 'image/png,image/jpeg,image/gif,image/webp',
)
const formatTime = (value) => (value ? new Date(value).toLocaleString() : '—')
const actionLabel = (value) =>
  ({
    save: '保存草稿',
    publish: '发布',
    schedule: '安排发布',
    'scheduled-publish': '定时发布',
    'cancel-schedule': '取消定时',
    unpublish: '下架',
    delete: '移入回收站',
    restore: '恢复文章',
    'restore-revision': '恢复历史版本',
    baseline: '历史原文备份',
    top: '调整置顶',
  })[value] || value
const recoveryOwner =
  rbac.currentUser?.id || JSON.parse(sessionStorage.getItem('userInfo') || '{}').id || 'unknown'
const recoveryKey = (id = article.id) => `article-draft:${recoveryOwner}:${id || 'new'}`
let pending = null,
  pendingAction = null,
  timer,
  generation = 0,
  applying = false,
  disposed = false,
  internalNavigation = false
function persist() {
  try {
    localStorage.setItem(
      recoveryKey(),
      JSON.stringify({
        id: article.id,
        version: article.version,
        fields: draftFields(article),
        pending,
        at: Date.now(),
      }),
    )
  } catch {
    saveError.value = '本地备份空间不足，请下载本地稿并保存到服务器'
  }
}
function applyServer(value) {
  if (value.content_format !== 'markdown') throw new Error('文章格式必须为 Markdown')
  applying = true
  Object.assign(article, empty(), value, {
    tag_ids: value.tag_ids || value.tags?.map((t) => t.id) || [],
  })
  savedFingerprint.value = draftFingerprint(article)
  lastSaved.value = new Date().toLocaleTimeString()
  saveError.value = ''
  conflict.value = false
  pending = null
  applying = false
}
async function loadOptions() {
  try {
    Object.assign(options, (await getArticleOptions()).data)
    optionsError.value = ''
  } catch (e) {
    optionsError.value = articleError(e, '分类和标签加载失败')
  }
}
async function load() {
  const seq = ++generation
  clearTimeout(timer)
  loadState.value = 'loading'
  loadError.value = ''
  try {
    if (
      route.params.id &&
      (!/^[1-9]\d*$/.test(String(route.params.id)) ||
        BigInt(route.params.id) > 18446744073709551615n)
    )
      throw new Error('文章 ID 无效，请从文章列表重新打开')
    const id = route.params.id ? String(route.params.id) : null
    const value = id ? (await getArticleDetail(id)).data : empty()
    if (seq !== generation || disposed) return
    applyServer(value)
    loadState.value = 'ready'
    editorKey.value++
    const raw = localStorage.getItem(recoveryKey())
    if (raw) {
      const backup = JSON.parse(raw)
      if (JSON.stringify(backup.fields) !== savedFingerprint.value) {
        try {
          await ElMessageBox.confirm(
            '发现此账号的本地未保存稿，是否恢复？服务器内容不会立即被覆盖。',
            '恢复草稿',
            { confirmButtonText: '恢复本地稿', cancelButtonText: '保留服务器版本' },
          )
          if (seq !== generation) return
          Object.assign(article, backup.fields)
          pending = backup.pending || null
          if (id && backup.version !== article.version) {
            conflict.value = true
            saveError.value = '本地草稿基于旧版本，请比较后处理'
          }
          editorKey.value++
        } catch {
          localStorage.removeItem(recoveryKey())
        }
      }
    }
  } catch (e) {
    if (seq === generation) {
      loadState.value = 'error'
      loadError.value = articleError(e)
    }
  }
  loadOptions()
}
async function save(automatic = false) {
  if (!editable.value || busy.value) return false
  editor.value?.flush()
  if (!dirty.value && article.id) return true
  if (charCount.value > MAX_ARTICLE_LENGTH) {
    saveError.value = '正文最多 500,000 字符'
    return false
  }
  if (automatic && !article.title.trim() && !article.content.trim()) return false
  busy.value = true
  try {
    // Preparing the request can fail before any HTTP request is sent.
    // Keep it inside finally's scope so the editor always becomes usable again.
    saveError.value = ''
    clearTimeout(timer)
    const fields = draftFields(article),
      oldKey = recoveryKey()
    const payload = {
      ...fields,
      ...(article.id ? { id: article.id, version: article.version } : {}),
    }
    const signature = JSON.stringify(payload)
    if (!pending || pending.signature !== signature)
      pending = { signature, request_id: newRequestID() }
    persist()
    const res = await (article.id ? updateArticle : createArticle)({
      ...payload,
      request_id: pending.request_id,
    })
    if (disposed) return false
    applyServer(res.data)
    localStorage.removeItem(oldKey)
    localStorage.removeItem(recoveryKey())
    if (!route.params.id) {
      internalNavigation = true
      try {
        await router.replace(router.page('blog/articles/editor/index', { params: { id: article.id } }))
      } finally {
        internalNavigation = false
      }
    }
    if (!automatic) ElMessage.success('草稿已保存')
    return true
  } catch (e) {
    conflict.value = e.response?.status === 409
    saveError.value = articleError(e)
    persist()
    return false
  } finally {
    busy.value = false
  }
}
async function act(action, extra = {}) {
  if (busy.value) return
  if (editable.value) editor.value?.flush()
  if (
    ['publish', 'schedule'].includes(action) &&
    (!article.title.trim() || !article.content.trim())
  )
    return ElMessage.warning('发布前请填写标题和正文')
  if (
    action === 'schedule' &&
    (!scheduleTime.value || new Date(scheduleTime.value).getTime() <= Date.now() + 60000)
  )
    return ElMessage.warning('请选择至少一分钟后的时间')
  if (action === 'unpublish') {
    try {
      await ElMessageBox.confirm('下架后访客将无法访问这篇文章，是否继续？', '下架文章', {
        type: 'warning',
      })
    } catch {
      return
    }
  }
  if (action !== 'restore' && dirty.value && !(await save(false))) return
  if (!article.id && !(await save(false))) return
  busy.value = true
  try {
    saveError.value = ''
    const payload = {
      id: article.id,
      version: article.version,
      ...(action === 'schedule'
        ? { scheduled_at: new Date(scheduleTime.value).toISOString() }
        : {}),
      ...extra,
    }
    const signature = JSON.stringify({ action, ...payload })
    if (pendingAction?.signature !== signature)
      pendingAction = { signature, request_id: newRequestID() }
    const res = await articleAction(action, {
      ...payload,
      request_id: pendingAction.request_id,
    })
    pendingAction = null
    applyServer(res.data)
    editorKey.value++
    localStorage.removeItem(recoveryKey())
    ElMessage.success('操作成功')
  } catch (e) {
    saveError.value = articleError(e)
    conflict.value = e.response?.status === 409
  } finally {
    busy.value = false
  }
}
async function importMarkdown(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file || !editable.value) return
  try {
    validateMarkdownFile(file)
    const source = new TextDecoder('utf-8', { fatal: true }).decode(await file.arrayBuffer())
    if (Array.from(source).length > MAX_ARTICLE_LENGTH || source.includes('\0'))
      throw new Error('正文过长或文件不是有效的文本')
    if (article.content.trim())
      await ElMessageBox.confirm('导入会替换当前正文；可先保存或下载当前稿。', '导入 Markdown')
    article.content = source
    if (!article.title) article.title = file.name.replace(/\.(md|markdown)$/i, '')
    editorKey.value++
    ElMessage.success('Markdown 已导入，请检查后保存')
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') ElMessage.error(e.message || '导入失败')
  }
}
function downloadLocal() {
  saveBlob(
    new Blob([article.content], { type: 'text/markdown;charset=utf-8' }),
    articleFilename(article),
  )
}
function downloadComparison() {
  const version = comparison.value
  const extension = version.content_format === 'markdown' ? 'md' : 'txt'
  saveBlob(
    new Blob([version.content || ''], { type: 'text/plain;charset=utf-8' }),
    `article-${article.id}-v${version.version}.${extension}`,
  )
}
function pickAsset(kind) {
  assetKind.value = kind
  assetInput.value.click()
}
async function chooseAsset(event) {
  const files = [...(event.target.files || [])]
  event.target.value = ''
  await uploadFiles(files)
}
async function uploadFiles(files, kind = assetKind.value) {
  if (!editable.value || busy.value) return
  busy.value = true
  try {
    for (const file of files) {
      uploadProgress.value = 0
      const res = await uploadArticleAsset(file, (p) => (uploadProgress.value = p))
      const asset = res.data
      const name = asset.name.replace(/[\[\]\\\r\n]/g, '_')
      if (kind === 'cover') article.cover = asset.url
      else
        editor.value.insert(
          `\n${asset.mime.startsWith('image/') ? '!' : ''}[${name}](${asset.url})\n`,
        )
    }
    ElMessage.success('资源已插入，保存文章后完成关联')
  } catch (e) {
    ElMessage.error(articleError(e))
  } finally {
    uploadProgress.value = null
    busy.value = false
  }
}
async function openHistory() {
  try {
    history.value = (await getArticleHistory(article.id)).data
    historyVisible.value = true
  } catch (e) {
    ElMessage.error(articleError(e))
  }
}
async function inspectRevision(row) {
  try {
    comparison.value = {
      ...(await getArticleRevision(article.id, row.id)).data,
      revision_id: row.id,
    }
    comparisonLatest.value = false
    comparisonVisible.value = true
  } catch (e) {
    ElMessage.error(articleError(e))
  }
}
async function compareLatest() {
  try {
    comparison.value = (await getArticleDetail(article.id)).data
    comparisonLatest.value = true
    comparisonVisible.value = true
  } catch (e) {
    ElMessage.error(articleError(e))
  }
}
async function useLatest() {
  try {
    await ElMessageBox.confirm('本地内容将被服务器版本替换，建议先下载本地稿。', '使用服务器版本')
    applyServer(comparison.value)
    localStorage.removeItem(recoveryKey())
    editorKey.value++
    comparisonVisible.value = false
  } catch {}
}
async function restoreRevision() {
  await act('restore-revision', { revision_id: comparison.value.revision_id })
  comparisonVisible.value = false
  historyVisible.value = false
}
async function copyCode(event) {
  const button = event.target.closest('.code-copy-btn')
  if (!button) return
  try {
    await navigator.clipboard.writeText(
      button.closest('.code-block-wrapper').querySelector('code').textContent,
    )
    ElMessage.success('已复制代码')
  } catch {
    ElMessage.error('无法访问剪贴板')
  }
}
function jump(id) {
  document
    .querySelector('.publication-preview')
    ?.querySelector(`[id="${CSS.escape(id)}"]`)
    ?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
async function leave() {
  if (internalNavigation) return true
  if (busy.value) {
    ElMessage.warning('操作进行中，请稍后离开')
    return false
  }
  if (!dirty.value) return true
  persist()
  try {
    await ElMessageBox.confirm('修改尚未保存到服务器，已保留本地恢复稿。仍要离开吗？', '离开编辑页')
    return true
  } catch {
    return false
  }
}
onBeforeRouteLeave((to) => {
  if (to.path === '/login') {
    clearTimeout(timer)
    localStorage.removeItem(recoveryKey())
    return true
  }
  return leave()
})
onBeforeRouteUpdate(async (to) => {
  if (String(to.params.id || '') === String(article.id || '')) return true
  return leave()
})
watch(
  () => route.params.id,
  (id) => {
    if (ready.value && String(id || '') === String(article.id || '')) return
    load()
  },
)
watch(
  () => draftFingerprint(article),
  () => {
    if (applying || !ready.value) return
    clearTimeout(timer)
    if (dirty.value) {
      persist()
      timer = setTimeout(() => {
        if (!saveError.value) save(true)
      }, 2500)
    }
  },
)
function beforeUnload(event) {
  // During the leave animation this editor remains in the DOM, but no longer owns navigation.
  if (studioHeader.value?.closest('[inert]')) return
  if (dirty.value || busy.value) {
    persist()
    event.preventDefault()
    event.returnValue = ''
  }
}
function keydown(event) {
  if (studioHeader.value?.closest('[inert]')) return
  if (
    event.key === 'Escape' &&
    immersive.value &&
    !showSettings.value &&
    !historyVisible.value &&
    !comparisonVisible.value
  )
    immersive.value = false
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
    event.preventDefault()
    save(false)
  }
}
onMounted(() => {
  // The action bar wraps on small screens. Measure it so the editor toolbar never overlaps it.
  const measureHeader = () => {
    headerHeight.value = studioHeader.value?.offsetHeight || 66
  }
  measureHeader()
  headerObserver = new ResizeObserver(measureHeader)
  headerObserver.observe(studioHeader.value, { box: 'border-box' })
  load()
  window.addEventListener('beforeunload', beforeUnload)
  window.addEventListener('keydown', keydown)
})
onBeforeUnmount(() => {
  headerObserver?.disconnect()
  disposed = true
  generation++
  clearTimeout(timer)
  window.removeEventListener('beforeunload', beforeUnload)
  window.removeEventListener('keydown', keydown)
})
</script>
<style scoped lang="scss" src="./article-editor.scss"></style>
<style lang="scss">
.article-settings-drawer .el-drawer__header {
  padding: 24px;
  margin-bottom: 0;
  border-bottom: 1px solid #ebeef5;
}
.article-settings-drawer .el-drawer__body {
  padding: 24px;
}
.article-settings-drawer .el-drawer__footer {
  padding: 16px 24px;
  border-top: 1px solid #ebeef5;
}
.article-settings-drawer .el-drawer__footer .el-button {
  width: 100%;
}
.article-settings-drawer .el-form-item {
  margin-bottom: 22px;
}
.article-settings-drawer .el-form-item__label {
  font-size: 12px;
  color: #606266;
  padding-bottom: 7px;
  line-height: 1.5;
}
.article-settings-drawer .el-form-item__content > small {
  color: #909399;
  line-height: 1.7;
  font-size: 11px;
  margin-top: 6px;
}
.article-settings-drawer .el-input__wrapper,
.article-settings-drawer .el-select__wrapper {
  min-height: 36px;
  border-radius: 6px;
}
.article-settings-drawer .el-textarea__inner {
  padding: 10px 12px;
  border-radius: 6px;
  line-height: 1.8;
}
</style>
