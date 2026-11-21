<template>
  <AdminListPage
    title="文档模式"
    :columns="columns"
    preference-key="document-modes"
    :loading="loading || saving"
    :error="modeError"
    :total="filtered.length"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    @search="search"
    @reset="reset"
    @refresh="load"
  >
    <template #filters>
      <el-form-item label="模式名称"
        ><el-input v-model="keyword" placeholder="名称、标识或说明" clearable
      /></el-form-item>
      <el-form-item label="展示状态"
        ><el-select v-model="status" clearable placeholder="全部状态">
          <el-option
            v-for="(label, value) in statuses"
            :key="value"
            :label="label"
            :value="value"
          /> </el-select
      ></el-form-item>
    </template>
    <template #actions>
      <el-button type="primary" :icon="Plus" :disabled="!manifest || !!modeError" @click="edit()"
        >新增模式</el-button
      >
      <el-button @click="$router.push($router.page('product/docs/versions/index'))">版本管理</el-button>
      <el-button @click="previewWebsite('/docs')">查看前台</el-button>
    </template>
    <template #notice
      >管理文档页顶部的模式分类。先创建模式和所属版本，发布版本后选择默认版本，即可将模式上线。</template
    >
    <template #default="{ tableHeight, tableProps }">
      <AdminDataTable v-bind="tableProps" :data="pageItems" :max-height="tableHeight">
        <template #label="{ row }"
          ><ResourceIcon :name="row.icon" /> <strong>{{ row.label }}</strong>
          <el-tag v-if="row.id === manifest.defaultMode" size="small" effect="plain">默认</el-tag>
        </template>
        <template #status="{ row }"
          ><el-tag :type="row.status === 'published' ? 'success' : 'info'">{{
            statuses[row.status]
          }}</el-tag></template
        >
        <template #count="{ row }">{{ versionsFor(row.id).length }} 个版本</template>
        <template #actions="{ row }"
          ><AdminRowActions>
            <AdminRowAction action="edit" label="编辑模式" :disabled="saving" @click="edit(row)" />
            <AdminRowAction
              action="view"
              label="管理所属版本"
              @click="$router.push($router.page('product/docs/versions/index', { query: { mode: row.id } }))"
            />
            <AdminRowAction
              action="publish"
              label="设为默认模式"
              :disabled="saving || row.status !== 'published' || row.id === manifest.defaultMode"
              @click="setDefault(row)"
            />
            <AdminRowAction
              action="delete"
              label="删除模式"
              :disabled="saving || versionsFor(row.id).length > 0"
              @click="remove(row)"
            /> </AdminRowActions
        ></template>
      </AdminDataTable>
    </template>
    <template #dialogs>
      <el-dialog
        v-model="dialog"
        :title="editing ? '编辑文档模式' : '新增文档模式'"
        class="admin-dialog"
        :close-on-click-modal="false"
        :show-close="!saving"
        :close-on-press-escape="!saving"
      >
        <el-form label-position="top" @submit.prevent="submit">
          <el-form-item label="模式名称" required
            ><el-input v-model="form.label" maxlength="60" placeholder="例如：单节点 API 模式"
          /></el-form-item>
          <el-form-item label="模式标识" required
            ><el-input
              v-model="form.id"
              :disabled="!!editing"
              maxlength="80"
              placeholder="例如：single-node-api"
            />
            <p class="form-note">用于访问链接，创建后保持不变。</p>
          </el-form-item>
          <el-form-item label="模式说明"
            ><el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              maxlength="500"
              show-word-limit
          /></el-form-item>
          <el-form-item label="展示图标"
            ><IconPicker v-model="form.icon" /></el-form-item>
          <el-form-item label="展示状态"
            ><el-select v-model="form.status"
              ><el-option
                v-for="(label, value) in statuses"
                :key="value"
                :label="label"
                :value="value" /></el-select
          ></el-form-item>
          <el-form-item label="默认版本" :required="form.status === 'published'">
            <el-select
              v-model="form.defaultVersion"
              clearable
              placeholder="选择此模式下已发布的版本"
            >
              <el-option
                v-for="item in publishedVersions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
            <p class="form-note">暂无可选版本时，先保存为“待推出”，再到版本管理创建并发布内容。</p>
          </el-form-item>
          <el-form-item label="显示顺序"
            ><el-input-number v-model="form.sortOrder" :min="0" :max="9999" :precision="0" />
            <p class="form-note">数字越小越靠前。</p></el-form-item
          >
        </el-form>
        <template #footer
          ><el-button :disabled="saving" @click="dialog = false">取消</el-button
          ><el-button type="primary" :loading="saving" @click="submit"
            >保存模式</el-button
          ></template
        >
      </el-dialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
import IconPicker from '@/components/icons/IconPicker.vue'
import { computed, reactive, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import { changeDocMode } from '@/api/documents'
import { previewWebsite } from '@/utils/website'
const { manifest, loading, saving, error, load, save, mutate } = useDocumentWorkspace()
const modeError = computed(
  () =>
    error.value ||
    (manifest.value && !Array.isArray(manifest.value.modes)
      ? '文档模式尚未就绪，请完成数据库迁移并重启后端'
      : ''),
)
const statuses = { published: '已上线', upcoming: '待推出', hidden: '已隐藏' }
import { columns } from './columns'
const keyword = ref(''),
  status = ref(''),
  applied = reactive({ keyword: '', status: '' })
const page = ref(1),
  pageSize = ref(10)
const filtered = computed(() =>
  (manifest.value?.modes || []).filter(
    (item) =>
      (item.label + ' ' + item.id + ' ' + item.description)
        .toLowerCase()
        .includes(applied.keyword) &&
      (!applied.status || item.status === applied.status),
  ),
)
const pageItems = computed(() =>
  filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value),
)
watch(
  () => filtered.value.length,
  (count) => {
    page.value = Math.min(page.value, Math.max(1, Math.ceil(count / pageSize.value)))
  },
)
function search() {
  applied.keyword = keyword.value.trim().toLowerCase()
  applied.status = status.value
  page.value = 1
}
function reset() {
  keyword.value = ''
  status.value = ''
  search()
}
const versionsFor = (id) => (manifest.value?.versions || []).filter((v) => v.modeId === id)
const dialog = ref(false),
  editing = ref(false)
const form = reactive({
  id: '',
  label: '',
  description: '',
  icon: 'playbook',
  status: 'upcoming',
  sortOrder: 0,
  defaultVersion: '',
  revision: 0,
})
const publishedVersions = computed(() =>
  versionsFor(form.id).filter((v) => v.status === 'published'),
)
function edit(row) {
  editing.value = !!row
  for (const [key, value] of Object.entries({
    id: '',
    label: '',
    description: '',
    icon: 'playbook',
    status: 'upcoming',
    sortOrder: 0,
    defaultVersion: '',
    revision: 0,
  }))
    form[key] = row?.[key] ?? value
  dialog.value = true
}
async function submit() {
  const ok = await save(async () => {
    if (!form.label.trim() || !/^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$/.test(form.id))
      throw new Error('请填写名称和有效的模式标识')
    if (form.status === 'published' && !form.defaultVersion)
      throw new Error('上线前请选择已发布的默认版本')
    await mutate(changeDocMode, { action: 'save', ...form, label: form.label.trim() })
  }, '文档模式已保存')
  if (ok) dialog.value = false
}
async function confirm(message, title) {
  try {
    await ElMessageBox.confirm(message, title, {
      type: 'warning',
      confirmButtonText: '确认',
      cancelButtonText: '取消',
    })
    return true
  } catch {
    return false
  }
}
async function remove(row) {
  if (await confirm('删除文档模式“' + row.label + '”？前台将不再展示此分类。', '删除文档模式'))
    await save(
      () => mutate(changeDocMode, { action: 'delete', id: row.id, revision: row.revision }),
      '文档模式已删除',
    )
}
async function setDefault(row) {
  if (await confirm('未指定模式时默认打开“' + row.label + '”？', '设置默认模式'))
    await save(
      () => mutate(changeDocMode, { action: 'default', id: row.id, revision: row.revision }),
      '默认模式已更新',
    )
}
load()
</script>
