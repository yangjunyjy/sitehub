<template>
  <AdminListPage
    :columns="columns"
    preference-key="document-groups"
    title="分组管理"
    :error="error"
    :loading="loading || saving"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    :total="filteredGroups.length"
    @search="search"
    @reset="reset"
    @refresh="load"
  >
    <template #filters>
      <DocumentVersionSelect
        :model-value="versionID"
        :versions="manifest?.versions"
        :default-version="manifest?.defaultVersion"
        :disabled="saving"
        @update:model-value="selectVersion" /><el-form-item label="分组名称"
        ><el-input v-model="keyword" placeholder="请输入分组名称" clearable /></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" :disabled="!version || !!error" @click="edit()"
        >新建分组</el-button
      ><el-button @click="$router.push(location('/dashboard/docs'))">文档列表</el-button></template
    >
    <template #notice>上下移动调整文档站导航顺序；删除分组前需先移走或删除组内文档。</template>
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable
        v-bind="tableProps"
        :data="pageGroups"
        :max-height="tableHeight"
        empty-text="暂无匹配分组"
      >
        <template #order="{ row }">{{ row.index + 1 }}</template>
        <template #count="{ row }">{{ row.items.length }}</template>
        <template #documents="{ row }">{{
          row.items
            .slice(0, 3)
            .map((item) => item.title)
            .join('、') || '暂无文档'
        }}</template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="moveUp"
              :label="'上移 ' + row.title"
              :disabled="row.index === 0 || saving || !!error"
              @click="move(row.index, -1)" /><AdminRowAction
              action="moveDown"
              :label="'下移 ' + row.title"
              :disabled="row.index === version.groups.length - 1 || saving || !!error"
              @click="move(row.index, 1)" /><AdminRowAction
              action="rename"
              :disabled="saving || !!error"
              @click="edit(row.index)" /><AdminRowAction
              action="delete"
              :label="row.items.length ? '分组内有文档，不能删除' : '删除'"
              :disabled="!!row.items.length || saving || !!error"
              @click="remove(row.index)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
    <template #dialogs
      ><el-dialog
        v-model="dialog"
        :title="editingIndex === null ? '新建分组' : '重命名分组'"
        class="admin-dialog"
        :close-on-click-modal="false"
        :close-on-press-escape="!saving"
        :show-close="!saving"
      >
        <el-form label-position="top" @submit.prevent="saveGroup"
          ><el-form-item label="分组名称" required
            ><el-input
              v-model="groupTitle"
              maxlength="60"
              show-word-limit
              placeholder="例如：快速开始"
              autofocus /></el-form-item
        ></el-form>
        <template #footer
          ><el-button :disabled="saving" @click="dialog = false">取消</el-button
          ><el-button type="primary" :loading="saving" @click="saveGroup"
            >保存分组</el-button
          ></template
        >
      </el-dialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import { ref, computed, watch } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
const columns = [
  {
    key: 'order',
    label: '顺序',
    width: '90',
  },
  {
    key: 'title',
    prop: 'title',
    label: '分组名称',
    minWidth: '180',
  },
  {
    key: 'count',
    label: '文档数量',
    width: '120',
  },
  {
    key: 'documents',
    label: '组内文档',
    minWidth: '220',
  },
  {
    key: 'actions',
    label: '操作',
    width: '238',
    fixed: 'right',
    utility: true,
  },
]

import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import { changeDocGroup } from '@/api/documents'
import DocumentVersionSelect from './DocumentVersionSelect.vue'

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
  appliedKeyword = ref(''),
  page = ref(1),
  pageSize = ref(10)
const filteredGroups = computed(() =>
  (version.value?.groups || [])
    .map((item, index) => ({ ...item, index }))
    .filter((item) => item.title.toLowerCase().includes(appliedKeyword.value.toLowerCase().trim())),
)
const pageGroups = computed(() =>
  filteredGroups.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value),
)
watch(
  () => filteredGroups.value.length,
  (count) => {
    page.value = Math.min(page.value, Math.max(1, Math.ceil(count / pageSize.value)))
  },
)
function search() {
  appliedKeyword.value = keyword.value
  page.value = 1
}
function reset() {
  keyword.value = ''
  search()
}
const dialog = ref(false),
  editingIndex = ref(null),
  groupTitle = ref('')
function edit(index = null) {
  editingIndex.value = index
  groupTitle.value = index === null ? '' : version.value.groups[index].title
  dialog.value = true
}
watch(versionID, () => {
  reset()
  dialog.value = false
})
async function saveGroup() {
  const title = groupTitle.value.trim()
  if (!title) {
    ElMessage.warning('请填写分组名称')
    return
  }
  if (
    version.value.groups.some((item, index) => item.title === title && index !== editingIndex.value)
  ) {
    ElMessage.warning('分组名称已存在')
    return
  }
  const selected = editingIndex.value === null ? null : version.value.groups[editingIndex.value]
  const ok = await save(
    () =>
      mutate(changeDocGroup, {
        action: 'save',
        version: versionID.value,
        title,
        id: selected?.id || 0,
        revision: selected?.revision || 0,
      }),
    '分组已保存',
  )
  if (ok) dialog.value = false
}
function move(index, delta) {
  const group = version.value.groups[index]
  return save(
    () =>
      mutate(changeDocGroup, {
        action: 'move',
        version: versionID.value,
        id: group.id,
        revision: group.revision,
        direction: delta,
      }),
    '分组顺序已更新',
  )
}
async function remove(index) {
  if (version.value.groups[index].items.length) return
  try {
    await ElMessageBox.confirm(
      `确认删除空分组“${version.value.groups[index].title}”？`,
      '删除分组',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  const group = version.value.groups[index]
  await save(
    () =>
      mutate(changeDocGroup, {
        action: 'delete',
        version: versionID.value,
        id: group.id,
        revision: group.revision,
      }),
    '分组已删除',
  )
}
load()
</script>
