<template>
  <AdminListPage
    :columns="columns"
    preference-key="documents"
    title="文档列表"
    :error="error"
    :loading="loading || saving"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    :total="filtered.length"
    @search="search"
    @reset="resetFilters"
    @refresh="load"
  >
    <template #filters>
      <DocumentVersionSelect
        :model-value="versionID"
        :versions="manifest?.versions"
        :default-version="manifest?.defaultVersion"
        :disabled="saving"
        @update:model-value="selectVersion" /><el-form-item label="文档搜索"
        ><el-input v-model="keyword" placeholder="标题、摘要或访问标识" clearable /></el-form-item
      ><el-form-item label="所属分组"
        ><el-select v-model="group" placeholder="全部分组" clearable
          ><el-option
            v-for="(item, index) in version?.groups"
            :key="index"
            :label="item.title"
            :value="index" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button
        type="primary"
        plain
        :icon="Plus"
        :disabled="!version || !!error"
        @click="router.push(location('product/docs/editor/index'))"
        >新建文档</el-button
      >
      <el-button
        :disabled="version?.status !== 'published'"
        :icon="View"
        @click="
          previewWebsite(
            `/docs?version=${encodeURIComponent(versionID)}&mode=${encodeURIComponent(version?.modeId || 'cli')}`,
          )
        "
        >文档站预览</el-button
      ><el-button @click="router.push(location('product/docs/groups/index'))">管理分组</el-button
      ><el-button @click="router.push(router.page('product/docs/versions/index'))">版本管理</el-button></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        :data="pageItems"
        class="admin-table"
        empty-text="暂无匹配文档，可调整筛选或新建文档"
      >
        <template #title="{ row }"
          ><div class="doc-title-cell">
            <div class="document-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div>
              <div class="table-title">{{ row.title }}</div>
              <div class="table-description">{{ row.summary || '暂无摘要' }}</div>
            </div>
          </div></template
        >
        <template #slug="{ row }"
          ><code class="slug-label">{{ row.slug }}</code></template
        >
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="edit"
              @click="
                router.push(location('product/docs/editor/index', { id: row.id }))
              " /><AdminRowAction
              action="view"
              @click="
                router.push(location('product/docs/editor/index', { id: row.id, mode: 'preview' }))
              " /><AdminRowAction
              action="delete"
              :disabled="saving || !!error"
              @click="remove(row)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
  </AdminListPage>
</template>
<script setup>
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
const columns = [
  {
    key: 'title',
    label: '文档标题',
    minWidth: '300',
  },
  {
    key: 'groupTitle',
    prop: 'groupTitle',
    label: '所属分组',
    minWidth: '150',
  },
  {
    key: 'slug',
    prop: 'slug',
    label: '访问标识',
    minWidth: '170',
  },
  {
    key: 'actions',
    label: '操作',
    width: '184',
    fixed: 'right',
    utility: true,
  },
]

import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Document, View, Plus } from '@element-plus/icons-vue'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import DocumentVersionSelect from './components/DocumentVersionSelect.vue'

import { flattenDocuments } from '@/utils/documents'
import { deleteDocument } from '@/api/documents'
import { previewWebsite } from '@/utils/website'
const router = useRouter()
const {
  data,
  loading,
  saving,
  error,
  load,
  save,
  manifest,
  version,
  versionID,
  location,
  selectVersion,
  mutate,
} = useDocumentWorkspace()
const keyword = ref(''),
  group = ref(''),
  page = ref(1),
  pageSize = ref(10),
  appliedKeyword = ref(''),
  appliedGroup = ref('')
const documents = computed(() => flattenDocuments(version.value))
const filtered = computed(() =>
  documents.value.filter(
    (item) =>
      (appliedGroup.value === '' || item.groupIndex === appliedGroup.value) &&
      `${item.title} ${item.summary} ${item.slug}`
        .toLowerCase()
        .includes(appliedKeyword.value.toLowerCase().trim()),
  ),
)
const pageItems = computed(() =>
  filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value),
)
watch(versionID, resetFilters)
watch([appliedKeyword, appliedGroup], () => (page.value = 1))
watch(
  () => filtered.value.length,
  (count) => (page.value = Math.min(page.value, Math.max(1, Math.ceil(count / pageSize.value)))),
)
function search() {
  appliedKeyword.value = keyword.value
  appliedGroup.value = group.value ?? ''
  page.value = 1
}
function resetFilters() {
  keyword.value = ''
  group.value = ''
  search()
}
async function remove(row) {
  try {
    await ElMessageBox.confirm(`删除“${row.title}”后，文档站将无法访问此文档。`, '删除文档', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  await save(async () => {
    await mutate(deleteDocument, { version: versionID.value, id: row.id, revision: row.revision })
  }, '文档已删除')
}
load()
</script>
<style lang="scss" scoped>
.doc-title-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}
.document-icon {
  display: grid;
  place-items: center;
  width: 38px;
  height: 42px;
  border-radius: 4px;
  background: #f5f7fa;
  color: #409eff;
  font-size: 20px;
  flex-shrink: 0;
}
.slug-label {
  font-size: 12px;
  color: #606266;
  background: #f5f7fa;
  padding: 4px 7px;
  border-radius: 4px;
}
.table-description {
  max-width: 500px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>
