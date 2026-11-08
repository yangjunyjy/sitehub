<template>
  <AdminListPage
    :columns="columns"
    preference-key="downloads"
    class="download-page"
    title="下载管理"
    :filter-model="filters"
    :error="loadError"
    :loading="loading || saving"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    :total="filteredPackages.length"
    :page-sizes="[10, 20, 50]"
    @search="search"
    @reset="resetFilters"
    @refresh="load"
  >
    <template #summary
      ><section class="stats-grid" aria-label="下载版本统计">
        <article v-for="card in statCards" :key="card.label" class="stat-card">
          <div class="stat-icon" :class="card.tone">
            <el-icon><component :is="card.icon" /></el-icon>
          </div>
          <div>
            <span>{{ card.label }}</span
            ><strong>{{ ready ? card.value : '—' }}</strong
            ><small>{{ card.note }}</small>
          </div>
        </article>
      </section></template
    >
    <template #filters
      ><el-form-item label="版本搜索"
        ><el-input
          v-model="filters.keyword"
          placeholder="版本号、文件名或描述"
          clearable /></el-form-item
      ><el-form-item label="所属产品"
        ><el-select v-model="filters.product" placeholder="全部产品" clearable
          ><el-option
            v-for="item in products"
            :key="item"
            :label="item"
            :value="item" /></el-select></el-form-item
      ><el-form-item label="运行平台"
        ><el-select v-model="filters.platform" placeholder="全部平台" clearable
          ><el-option
            v-for="item in platforms"
            :key="item"
            :label="item"
            :value="item" /></el-select></el-form-item></template
    ><template #advanced
      ><el-form-item label="下载分组"
        ><el-select v-model="filters.groupId" placeholder="全部分组" clearable
          ><el-option
            v-for="group in downloads.groups"
            :key="group.id"
            :label="group.name"
            :value="group.id" /></el-select></el-form-item
      ><el-form-item label="版本类型"
        ><el-select v-model="filters.edition" placeholder="全部类型" clearable
          ><el-option
            v-for="edition in editions"
            :key="edition"
            :label="edition"
            :value="edition" /></el-select></el-form-item
      ><el-form-item label="发布状态"
        ><el-select v-model="filters.status" placeholder="全部状态" clearable
          ><el-option
            v-for="(label, state) in releaseStates"
            :key="state"
            :label="label"
            :value="state" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button
        type="primary"
        plain
        :icon="Plus"
        :disabled="!ready || loading || saving"
        @click="openEditor()"
        >新增下载文件</el-button
      >
      <el-button :icon="View" @click="preview">下载页预览</el-button
      ><el-button :icon="Setting" :disabled="!ready" @click="openSettings"
        >下载页设置</el-button
      ></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        :data="pagedPackages"
        row-key="id"
        class="release-table"
        empty-text="暂无匹配版本，可调整筛选或新增下载文件"
      >
        <template #product="{ row }"
          ><div class="product-cell">
            <span class="product-icon"
              ><el-icon><Box /></el-icon
            ></span>
            <div>
              <b>{{ row.product }}</b
              ><small :title="row.fileName || row.description">{{
                row.description || row.fileName || '产品安装包'
              }}</small>
            </div>
          </div></template
        >
        <template #groupId="{ row }"
          ><span>{{
            downloads.groups.find((group) => group.id === row.groupId)?.name || '未分组'
          }}</span
          ><small class="cell-secondary">{{ row.edition || '通用版本' }}</small></template
        >
        <template #version="{ row }"
          ><span class="version">{{ row.version }}</span></template
        >
        <template #platform="{ row }"
          >{{ row.platform
          }}<small class="cell-secondary">{{ row.arch || '架构未填写' }}</small></template
        >
        <template #size="{ row }">{{ formatFileSize(row.size) }}</template>
        <template #releaseDate="{ row }">{{ row.releaseDate || '—' }}</template>
        <template #status="{ row }"
          ><span class="release-status" :class="row.status"
            ><i />{{ releaseStates[row.status] }}</span
          ></template
        >
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="edit"
              :disabled="!ready"
              @click="openEditor(row)"
            /><AdminRowAction
              :action="row.status === 'published' ? 'unpublish' : 'publish'"
              :disabled="!ready"
              @click="togglePublish(row)"
            /><el-dropdown
              trigger="click"
              :popper-options="{
                modifiers: [
                  {
                    name: 'preventOverflow',
                    options: { padding: 8, altAxis: true, tether: false },
                  },
                ],
              }"
              :disabled="!ready"
              @command="(command) => handleMore(command, row)"
              ><AdminRowAction action="more" label="更多版本操作" :disabled="!ready" /><template
                #dropdown
                ><el-dropdown-menu
                  ><el-dropdown-item command="copy">复制为草稿</el-dropdown-item
                  ><el-dropdown-item command="delete" divided
                    >删除版本记录</el-dropdown-item
                  ></el-dropdown-menu
                ></template
              ></el-dropdown
            ></AdminRowActions
          ></template
        >
      </AdminDataTable></template
    >
    <template #footer
      ><p class="page-note">
        发布后安装包会出现在官网下载页；草稿和已下架版本仅在后台可见。
      </p></template
    >
    <template #dialogs
      ><el-dialog
        v-model="editorOpen"
        :title="editingID ? '编辑下载文件' : '新增下载文件'"
        width="min(740px, calc(100vw - 32px))"
        class="release-dialog"
        destroy-on-close
        :close-on-click-modal="false"
        :before-close="closeEditor"
      >
        <el-alert
          v-if="editorError"
          :title="editorError"
          type="error"
          :closable="false"
          show-icon
          class="load-error"
        />
        <el-form label-position="top" :disabled="saving || uploading">
          <div class="field-grid">
            <el-form-item label="产品名称 *"
              ><el-input v-model="draft.product" placeholder="Deploy" /></el-form-item
            ><el-form-item label="版本号 *"
              ><el-input v-model="draft.version" placeholder="v1.0.0"
            /></el-form-item>
          </div>
          <div class="field-grid">
            <el-form-item label="下载分组"
              ><el-select v-model="draft.groupId" clearable placeholder="未分组"
                ><el-option
                  v-for="group in downloads.groups"
                  :key="group.id"
                  :label="group.name"
                  :value="group.id"
              /></el-select>
              <p class="field-note">分组名称和顺序可在“下载页设置”维护。</p></el-form-item
            >
            <el-form-item label="版本类型"
              ><el-select
                v-model="draft.edition"
                filterable
                allow-create
                default-first-option
                clearable
                placeholder="选择或输入，例如单节点 API 版"
                ><el-option
                  v-for="edition in editionOptions"
                  :key="edition"
                  :label="edition"
                  :value="edition" /></el-select
            ></el-form-item>
          </div>
          <div class="field-grid">
            <el-form-item label="展示顺序"
              ><el-input-number v-model="draft.sortOrder" :min="0" :max="99999" :precision="0" />
              <p class="field-note">数值越小越靠前，相同数值保留原顺序。</p></el-form-item
            >
          </div>
          <div class="field-grid">
            <el-form-item label="平台 / 系统 *"
              ><el-select
                v-model="draft.platform"
                filterable
                allow-create
                default-first-option
                placeholder="选择或输入平台"
                ><el-option
                  v-for="name in [
                    'Windows',
                    'macOS',
                    'Linux',
                    'Android',
                    'iOS',
                    '跨平台',
                    '离线部署包',
                  ]"
                  :key="name"
                  :value="name"
                  :label="name" /></el-select></el-form-item
            ><el-form-item label="处理器架构"
              ><el-input v-model="draft.arch" placeholder="x64 / ARM64 / 通用"
            /></el-form-item>
          </div>
          <el-form-item label="上传文件"
            ><el-upload
              drag
              :show-file-list="false"
              :http-request="uploadFile"
              :disabled="uploading || saving"
              :accept="acceptedFiles"
              class="package-upload"
              ><el-icon class="upload-symbol"><UploadFilled /></el-icon
              ><b>{{
                uploading ? `正在上传 ${uploadProgress}%` : draft.fileName || '点击或拖拽文件到此处'
              }}</b
              ><small>{{
                draft.fileName
                  ? `${formatFileSize(draft.size)} · 点击可替换文件`
                  : `支持 ${acceptedFiles || '当前存储配置允许的格式'}`
              }}</small></el-upload
            ><el-progress v-if="uploading" :percentage="uploadProgress" :stroke-width="4"
          /></el-form-item>
          <el-form-item label="下载文件名">
            <el-input v-model="draft.fileName" placeholder="例如 deploy_arm.tgz" />
            <p class="field-note">
              用户下载后保存的名称，修改名称不会改变已上传文件的存储路径。请保留正确的扩展名。
            </p>
          </el-form-item>
          <el-form-item label="下载地址">
            <el-input
              v-model="draft.url"
              :readonly="localUpload && !editingLocation"
              placeholder="上传后自动填入，或填写已有文件的外部下载链接"
              @input="clearFileMetadata"
            />
            <div class="location-help">
              <p class="field-note">
                {{
                  localUpload
                    ? '上传文件的地址由系统生成。替换安装包请重新上传；只改下载名称请填写上方文件名。'
                    : '地址必须指向已经存在的文件。填写新地址不会移动或重命名原文件。'
                }}
              </p>
              <el-button
                v-if="localUpload && !editingLocation"
                link
                type="primary"
                @click="editingLocation = true"
                >更换文件地址</el-button
              >
            </div>
          </el-form-item>
          <div class="field-grid">
            <el-form-item label="包格式"
              ><el-input v-model="draft.format" placeholder="ZIP / DMG / TAR.GZ" /></el-form-item
            ><el-form-item label="发布日期"
              ><el-date-picker
                v-model="draft.releaseDate"
                value-format="YYYY-MM-DD"
                type="date"
                placeholder="选择日期"
            /></el-form-item>
          </div>
          <el-form-item label="版本说明"
            ><el-input
              v-model="draft.description"
              type="textarea"
              :rows="2"
              placeholder="简要描述更新内容或适用场景"
          /></el-form-item>
          <el-form-item label="SHA256（可选）"
            ><el-input v-model="draft.checksum" placeholder="64 位校验值"
          /></el-form-item>
        </el-form>
        <template #footer
          ><el-button :disabled="uploading || saving" @click="cancelEditor">取消</el-button
          ><el-button :disabled="uploading" :loading="saving" @click="saveDraft('draft')"
            >保存草稿</el-button
          ><el-button
            type="primary"
            :disabled="uploading"
            :loading="saving"
            @click="saveDraft('published')"
            >保存并发布</el-button
          ></template
        >
      </el-dialog>

      <el-drawer
        v-model="settingsOpen"
        title="下载页设置"
        :size="'min(760px, 100vw)'"
        :close-on-click-modal="false"
        :before-close="closeSettings"
      >
        <DownloadSettings v-if="settingsDraft" v-model="settingsDraft" /><template #footer
          ><el-button :disabled="saving" @click="cancelSettings">取消</el-button
          ><el-button type="primary" :loading="saving" @click="saveSettings"
            >保存设置</el-button
          ></template
        >
      </el-drawer>
    </template>
  </AdminListPage>
</template>
<script setup>
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
const columns = [
  {
    key: 'product',
    label: '产品名称',
    minWidth: '225',
  },
  { key: 'groupId', label: '分组 / 版本类型', minWidth: '180' },
  {
    key: 'version',
    label: '版本号',
    width: '115',
  },
  {
    key: 'platform',
    label: '平台 / 系统',
    minWidth: '155',
  },
  {
    key: 'size',
    label: '文件大小',
    width: '110',
  },
  {
    key: 'releaseDate',
    label: '发布日期',
    width: '120',
  },
  {
    key: 'status',
    label: '状态',
    width: '108',
  },
  {
    key: 'actions',
    label: '操作',
    width: '184',
    fixed: 'right',
    utility: true,
  },
]

import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getWebsiteDownloads, updateWebsiteDownloads, uploadDownloadPackage } from '@/api/website'
import {
  Box,
  Document,
  CircleCheck,
  EditPen,
  View,
  Plus,
  Setting,
  UploadFilled,
} from '@element-plus/icons-vue'
import {
  releaseStates,
  normalizePackage,
  filterPackages,
  formatFileSize,
  packageError,
  websiteURL,
  isLocalReleaseURL,
  normalizeDownloadURL,
} from '@/utils/downloads'
import DownloadSettings from './DownloadSettings.vue'
import { createUUID } from '@/utils/uuid'
import { useUnsavedChanges } from '@/composables/useAdminResource'

const loading = ref(false),
  saving = ref(false),
  ready = ref(false),
  loadError = ref('')
const downloads = ref({ packages: [], installOptions: [], requirements: [] })
const productName = ref('Deploy'),
  allowedExts = ref([])
const page = ref(1),
  pageSize = ref(10)
const filters = reactive({
  product: '',
  groupId: '',
  edition: '',
  platform: '',
  status: '',
  keyword: '',
})
const appliedFilters = reactive({ ...filters })
const packages = computed(() => downloads.value.packages)
const products = computed(() => [...new Set(packages.value.map((item) => item.product))])
const editions = computed(() => [
  ...new Set(packages.value.map((item) => item.edition).filter(Boolean)),
])
const editionOptions = computed(() => [
  ...new Set(['单节点 API 版', '分布式版', '命令行版', ...editions.value]),
])
const platforms = computed(() => [...new Set(packages.value.map((item) => item.platform))])
const publishedCount = computed(
  () => packages.value.filter((item) => item.status === 'published').length,
)
const statCards = computed(() => [
  {
    label: '产品数量',
    value: products.value.length,
    note: '已配置下载版本的产品',
    icon: Box,
    tone: 'blue',
  },
  {
    label: '版本文件',
    value: packages.value.length,
    note: `覆盖 ${platforms.value.length} 个平台`,
    icon: Document,
    tone: 'blue',
  },
  {
    label: '已发布版本',
    value: publishedCount.value,
    note: '当前可在官网下载',
    icon: CircleCheck,
    tone: 'green',
  },
  {
    label: '待发布版本',
    value: packages.value.filter((item) => item.status === 'draft').length,
    note: `${packages.value.filter((item) => item.status === 'archived').length} 个版本已下架`,
    icon: EditPen,
    tone: 'amber',
  },
])
const filteredPackages = computed(() => filterPackages(packages.value, appliedFilters))
const pagedPackages = computed(() =>
  filteredPackages.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value),
)
watch([() => filteredPackages.value.length, pageSize], () => {
  page.value = Math.min(
    page.value,
    Math.max(1, Math.ceil(filteredPackages.value.length / pageSize.value)),
  )
})
const search = () => {
  Object.assign(appliedFilters, filters)
  page.value = 1
}
const resetFilters = () => {
  Object.assign(filters, {
    product: '',
    groupId: '',
    edition: '',
    platform: '',
    status: '',
    keyword: '',
  })
  search()
}
const applyDownloads = (value) => {
  downloads.value = {
    ...value,
    title: value.title || '下载中心',
    groups: value.groups || [],
    installOptions: value.installOptions || [],
    requirements: value.requirements || [],
    packages: (value.packages || []).map((item) =>
      normalizePackage(item, { ...value, product: productName.value }),
    ),
  }
}
const load = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await getWebsiteDownloads()
    productName.value = data.content?.brand?.name || 'Deploy'
    allowedExts.value = data.storage?.allowedExts || []
    applyDownloads(data.downloads)
    ready.value = true
  } catch (error) {
    loadError.value = error.message || '下载版本加载失败'
    ready.value = false
  } finally {
    loading.value = false
  }
}
// 只提交后端声明的安装包字段，避免选中态等 UI 数据混入严格 JSON 接口。
const persist = async (value) => {
  if (saving.value || !ready.value) return false
  saving.value = true
  try {
    const clean = { ...value, packages: value.packages.map((item) => normalizePackage(item)) }
    const result = await updateWebsiteDownloads(clean)
    applyDownloads(result.data || clean)
    return true
  } finally {
    saving.value = false
  }
}

const editorOpen = ref(false),
  editingID = ref(''),
  draft = ref({}),
  editorError = ref('')
const uploading = ref(false),
  uploadProgress = ref(0)
const editingLocation = ref(false)
const editorBaseline = ref(''),
  settingsBaseline = ref('')
const localUpload = computed(() => isLocalReleaseURL(draft.value.url || ''))
const clearFileMetadata = () => {
  draft.value.size = 0
  draft.value.checksum = ''
}
const acceptedFiles = computed(() =>
  allowedExts.value
    .map((item) => (item.startsWith('.') ? item.toLowerCase() : `.${item.toLowerCase()}`))
    .join(','),
)
const today = () => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
}
const openEditor = (row) => {
  editingLocation.value = false
  editingID.value = row?.id || ''
  editorError.value = ''
  draft.value = normalizePackage(
    row || { id: `package-${createUUID()}`, status: 'draft', releaseDate: today() },
    { product: productName.value, version: downloads.value.version },
  )
  editingLocation.value = !isLocalReleaseURL(draft.value.url)
  if (!/^\d{4}-\d{2}-\d{2}$/.test(draft.value.releaseDate)) draft.value.releaseDate = ''
  editorBaseline.value = JSON.stringify(draft.value)
  editorOpen.value = true
}
const confirmDiscard = async (changed) => {
  if (!changed) return true
  try {
    await ElMessageBox.confirm('有未保存的修改，关闭后将丢失。', '关闭编辑', {
      confirmButtonText: '放弃修改',
      cancelButtonText: '继续编辑',
      type: 'warning',
    })
    return true
  } catch {
    return false
  }
}
const closeEditor = async (done) => {
  if (
    !uploading.value &&
    !saving.value &&
    (await confirmDiscard(JSON.stringify(draft.value) !== editorBaseline.value))
  )
    done()
}
const cancelEditor = () =>
  closeEditor(() => {
    editorOpen.value = false
  })
const closeSettings = async (done) => {
  if (
    !saving.value &&
    (await confirmDiscard(JSON.stringify(settingsDraft.value) !== settingsBaseline.value))
  )
    done()
}
const cancelSettings = () =>
  closeSettings(() => {
    settingsOpen.value = false
  })
const uploadFile = async ({ file, onSuccess, onError }) => {
  uploading.value = true
  uploadProgress.value = 0
  editorError.value = ''
  try {
    const result = await uploadDownloadPackage(file, (percent) => {
      uploadProgress.value = percent
    })
    draft.value.url = normalizeDownloadURL(result.data.url)
    editingLocation.value = false
    draft.value.fileName = file.name
    draft.value.size = file.size
    draft.value.format = file.name.toLowerCase().endsWith('.tar.gz')
      ? 'TAR.GZ'
      : file.name.split('.').pop().toUpperCase()
    draft.value.checksum = ''
    onSuccess(result)
  } catch (error) {
    editorError.value = error.message || '安装包上传失败'
    onError(error)
  } finally {
    uploading.value = false
  }
}
const saveDraft = async (status) => {
  const row = normalizePackage({ ...draft.value, status })
  row.url = row.url.trim()
  row.checksum = row.checksum.trim()
  if (status === 'published' && !row.releaseDate) row.releaseDate = today()
  editorError.value = packageError(row)
  if (editorError.value) return
  const next = editingID.value
    ? packages.value.map((item) => (item.id === editingID.value ? row : item))
    : [row, ...packages.value]
  try {
    if (await persist({ ...downloads.value, packages: next })) {
      editorOpen.value = false
      ElMessage.success(status === 'published' ? '版本已发布' : '草稿已保存')
    }
  } catch (error) {
    editorError.value = error.message || '保存失败，请重试'
  }
}
const togglePublish = async (row) => {
  if (saving.value) return
  const next = { ...row, status: row.status === 'published' ? 'archived' : 'published' }
  if (next.status === 'published' && !next.releaseDate) next.releaseDate = today()
  const error = packageError(next)
  if (error) {
    openEditor(row)
    editorError.value = error
    return
  }
  try {
    if (next.status === 'archived')
      await ElMessageBox.confirm('下架后官网将不再展示此版本，已有文件会保留。', '下架版本', {
        type: 'warning',
        confirmButtonText: '确认下架',
        cancelButtonText: '取消',
      })
    if (
      await persist({
        ...downloads.value,
        packages: packages.value.map((item) => (item.id === row.id ? next : item)),
      })
    )
      ElMessage.success(next.status === 'published' ? '版本已发布' : '版本已下架')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.message || '操作失败')
  }
}
const handleMore = async (command, row) => {
  if (command === 'copy') {
    openEditor({ ...row, id: `package-${createUUID()}`, status: 'draft' })
    editingID.value = ''
    return
  }
  try {
    await ElMessageBox.confirm(
      '将删除此版本记录并移除官网入口，存储中的文件会保留。',
      '删除版本记录',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' },
    )
    if (
      await persist({
        ...downloads.value,
        packages: packages.value.filter((item) => item.id !== row.id),
      })
    )
      ElMessage.success('版本记录已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.message || '删除失败')
  }
}
const settingsOpen = ref(false),
  settingsDraft = ref(null)
const openSettings = () => {
  settingsDraft.value = JSON.parse(JSON.stringify(downloads.value))
  settingsBaseline.value = JSON.stringify(settingsDraft.value)
  settingsOpen.value = true
}
const saveSettings = async () => {
  if (settingsDraft.value.groups.some((group) => !group.name.trim())) {
    ElMessage.error('请填写每个下载分组的名称')
    return
  }
  try {
    if (await persist({ ...settingsDraft.value, packages: packages.value })) {
      settingsOpen.value = false
      ElMessage.success('下载页设置已保存')
    }
  } catch (error) {
    ElMessage.error(error.message || '保存失败')
  }
}
useUnsavedChanges(
  () =>
    (editorOpen.value && JSON.stringify(draft.value) !== editorBaseline.value) ||
    (settingsOpen.value && JSON.stringify(settingsDraft.value) !== settingsBaseline.value),
)
const preview = () => window.open(websiteURL('/download'), '_blank', 'noopener,noreferrer')
load()
</script>

<style scoped>
.field-note {
  margin: 7px 0 0;
  color: #606266;
  font-size: 12px;
  line-height: 1.7;
}
.location-help {
  width: 100%;
}
.location-help .el-button {
  margin-top: 4px;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 22px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 17px;
  padding: 23px 20px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fff;
}
.stat-icon {
  display: grid;
  place-items: center;
  width: 50px;
  height: 50px;
  flex-shrink: 0;
  border-radius: 15px;
  font-size: 24px;
}
.blue {
  color: #409eff;
  background: #f5f7fa;
}
.green {
  color: #13a477;
  background: #f5f7fa;
}
.amber {
  color: #d79832;
  background: #fff7e9;
}
.stat-card span,
.stat-card small {
  display: block;
  color: #606266;
  font-size: 12px;
}
.stat-card small {
  font-size: 11px;
  color: #909399;
  margin-top: 6px;
}
.stat-card strong {
  display: block;
  margin: 6px 0 0;
  font-size: 28px;
  line-height: 1;
  color: #303133;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}

.list-toolbar .el-button {
  font-size: 12px;
  color: #606266;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 11px;
  min-width: 0;
}
.product-cell > div {
  min-width: 0;
}
.product-cell b,
.product-cell small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.product-cell b {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}
.product-cell small,
.cell-secondary {
  margin-top: 5px;
  color: #909399;
  font-size: 11px;
}
.cell-secondary {
  display: block;
}
.product-icon {
  display: grid;
  place-items: center;
  width: 35px;
  height: 35px;
  flex-shrink: 0;
  border-radius: 4px;
  color: #409eff;
  background: #f5f7fa;
  font-size: 20px;
}
.version {
  font:
    12px ui-monospace,
    monospace;
}
.release-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 5px;
  padding: 4px 7px;
  color: #606266;
  background: #f5f7fa;
  font-size: 11px;
  white-space: nowrap;
}
.release-status i {
  width: 5px;
  height: 5px;
  background: currentColor;
  border-radius: 50%;
}
.release-status.published {
  color: #0b9871;
  background: #f5f7fa;
}
.release-status.archived {
  color: #ba873b;
  background: #fcf5e9;
}
.page-note {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-top: 14px;
  color: #909399;
  font-size: 11px;
}
.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}
.field-grid :deep(.el-select),
.field-grid :deep(.el-date-editor) {
  width: 100%;
}
.package-upload {
  width: 100%;
}
.package-upload b,
.package-upload small {
  display: block;
}
.package-upload b {
  color: #303133;
  font-size: 13px;
}
.package-upload small {
  color: #909399;
  font-size: 11px;
  margin-top: 7px;
  word-break: break-all;
}
.upload-symbol {
  font-size: 30px;
  color: #409eff;
  margin-bottom: 10px;
}
.load-error {
  margin-bottom: 18px;
}
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 700px) {
  .download-page {
    padding: 18px 12px;
  }

  .stats-grid {
    gap: 9px;
  }
  .stat-card {
    padding: 16px 12px;
    gap: 10px;
  }
  .stat-icon {
    width: 35px;
    height: 35px;
    font-size: 19px;
  }
  .stat-card small {
    font-size: 10px;
  }
  .stat-card strong {
    font-size: 23px;
  }

  .field-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }
  .page-note {
    align-items: flex-start;
    line-height: 1.7;
  }
}
</style>
