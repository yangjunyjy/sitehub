<template>
  <AdminListPage
    title="文档版本"
    :columns="columns"
    preference-key="document-versions"
    :loading="loading || saving"
    :error="error"
    :total="filtered.length"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    @refresh="load"
    @search="search"
    @reset="reset"
  >
    <template #filters
      ><el-form-item label="文档模式"
        ><el-select v-model="modeFilter" clearable placeholder="全部模式"
          ><el-option
            v-for="item in manifest?.modes"
            :key="item.id"
            :label="item.label"
            :value="item.id" /></el-select></el-form-item
      ><el-form-item label="版本名称"
        ><el-input v-model="keyword" placeholder="版本名称或标识" clearable /></el-form-item
      ><el-form-item label="发布状态"
        ><el-select v-model="status" placeholder="全部状态" clearable
          ><el-option label="已发布" value="published" /><el-option
            label="草稿"
            value="draft" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button
        type="primary"
        plain
        :icon="Plus"
        :disabled="!manifest || !!error"
        @click="create()"
        >新建版本</el-button
      ><el-button @click="$router.push($router.page('product/docs/index'))">文档列表</el-button></template
    >
    <template #notice
      >每个版本独立维护正文与分组。新版本先保存为草稿，发布后供读者选择；默认版本用于未指定版本的访问。</template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable v-bind="tableProps" :data="pageItems" :max-height="tableHeight">
        <template #name="{ row }"
          ><strong>{{ row.name }}</strong>
          <el-tag v-if="isDefaultVersion(row)" size="small" effect="plain"
            >默认版本</el-tag
          ></template
        >
        <template #count="{ row }"
          >{{ row.groups.reduce((n, group) => n + group.items.length, 0) }} 篇 ·
          {{ row.groups.length }} 个分组</template
        >
        <template #status="{ row }"
          ><el-tag :type="row.status === 'published' ? 'success' : 'info'">{{
            row.status === 'published' ? '已发布' : '草稿'
          }}</el-tag></template
        >
        <template #mode="{ row }">{{
          manifest?.modes?.find((item) => item.id === row.modeId)?.label || row.modeId
        }}</template>
        <template #actions="{ row }"
          ><AdminRowActions>
            <AdminRowAction
              action="edit"
              label="维护此版本文档"
              @click="$router.push($router.page('product/docs/index', { query: { version: row.id } }))"
            />
            <AdminRowAction
              action="copy"
              label="复制为新版本"
              :disabled="saving"
              @click="create(row)"
            />
            <AdminRowAction
              :action="row.status === 'published' ? 'unpublish' : 'publish'"
              :label="row.status === 'published' ? '转为草稿' : '发布版本'"
              :disabled="saving || isDefaultVersion(row)"
              @click="toggle(row)"
            />
            <el-dropdown trigger="click" @command="(command) => more(command, row)"
              ><AdminRowAction action="more" :disabled="saving" /><template #dropdown
                ><el-dropdown-menu
                  ><el-dropdown-item command="rename">重命名</el-dropdown-item
                  ><el-dropdown-item
                    command="default"
                    :disabled="row.status !== 'published' || isDefaultVersion(row)"
                    >设为默认版本</el-dropdown-item
                  ><el-dropdown-item
                    command="delete"
                    :disabled="row.status !== 'draft' || isDefaultVersion(row)"
                    divided
                    >删除草稿版本</el-dropdown-item
                  ></el-dropdown-menu
                ></template
              ></el-dropdown
            >
          </AdminRowActions></template
        >
      </AdminDataTable></template
    >
    <template #dialogs
      ><el-dialog
        v-model="dialog"
        :title="editing ? '重命名版本' : '新建文档版本'"
        class="admin-dialog"
        :close-on-click-modal="false"
        :show-close="!saving"
        :close-on-press-escape="!saving"
        ><el-form label-position="top" @submit.prevent="submit">
          <el-form-item label="所属模式" required
            ><el-select v-model="form.modeId" :disabled="!!editing" placeholder="请选择文档模式"
              ><el-option
                v-for="item in manifest?.modes"
                :key="item.id"
                :label="item.label"
                :value="item.id" /></el-select
          ></el-form-item>
          <el-form-item label="版本名称" required
            ><el-input v-model="form.name" placeholder="例如：2.0" maxlength="60"
          /></el-form-item>
          <el-form-item label="版本标识" required
            ><el-input
              v-model="form.id"
              placeholder="例如：v2.0"
              :disabled="!!editing"
              maxlength="80"
            />
            <p class="form-note">
              所有模式中的标识需唯一，例如 cli-v2、api-v2；创建后保持不变。
            </p></el-form-item
          >
          <el-form-item v-if="!editing" label="初始内容"
            ><el-select v-model="form.source"
              ><el-option label="空白版本" value="" /><el-option
                v-for="item in manifest?.versions"
                :key="item.id"
                :value="item.id"
                :label="`复制 ${item.modeLabel} / ${item.name} 的文档与分组`"
            /></el-select>
            <p class="form-note">复制后独立保存，后续修改不会同步到来源版本。</p></el-form-item
          > </el-form
        ><template #footer
          ><el-button :disabled="saving" @click="dialog = false">取消</el-button
          ><el-button type="primary" :loading="saving" @click="submit">{{
            editing ? '保存名称' : '创建草稿版本'
          }}</el-button></template
        ></el-dialog
      ></template
    >
  </AdminListPage>
</template>
<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import { changeDocVersion } from '@/api/documents'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
const { manifest, loading, saving, error, load, save, mutate } = useDocumentWorkspace()
const route = useRoute()
const modeFilter = ref(String(route.query.mode || ''))
const isDefaultVersion = (row) =>
  row.id === manifest.value?.modes?.find((m) => m.id === row.modeId)?.defaultVersion ||
  row.id === manifest.value?.defaultVersion
const columns = [
  { key: 'mode', label: '所属模式', minWidth: 170 },
  { key: 'name', label: '产品版本', minWidth: 200 },
  { key: 'id', prop: 'id', label: '版本标识', minWidth: 130 },
  { key: 'count', label: '文档内容', minWidth: 180 },
  { key: 'status', label: '状态', width: 100 },
  { key: 'actions', label: '操作', width: 238, fixed: 'right', utility: true },
]
const keyword = ref(''),
  status = ref(''),
  applied = reactive({ keyword: '', status: '', mode: modeFilter.value }),
  page = ref(1),
  pageSize = ref(10)
const filtered = computed(() =>
  (manifest.value?.versions || []).filter(
    (item) =>
      `${item.name} ${item.id}`.toLowerCase().includes(applied.keyword) &&
      (!applied.mode || item.modeId === applied.mode) &&
      (!applied.status || item.status === applied.status),
  ),
)
const pageItems = computed(() =>
  filtered.value.slice((page.value - 1) * pageSize.value, page.value * pageSize.value),
)
watch(
  () => filtered.value.length,
  (count) => (page.value = Math.min(page.value, Math.max(1, Math.ceil(count / pageSize.value)))),
)
function search() {
  applied.keyword = keyword.value.trim().toLowerCase()
  applied.status = status.value
  applied.mode = modeFilter.value
  page.value = 1
}
watch(
  () => route.query.mode,
  (value) => {
    modeFilter.value = String(value || '')
    search()
  },
)
function reset() {
  keyword.value = ''
  status.value = ''
  modeFilter.value = ''
  search()
}
const dialog = ref(false),
  editing = ref(''),
  form = reactive({ id: '', name: '', source: '', modeId: '' })
function create(source) {
  editing.value = ''
  Object.assign(form, {
    id: '',
    name: '',
    source: source?.id || '',
    modeId: modeFilter.value || source?.modeId || manifest.value?.defaultMode || 'cli',
  })
  dialog.value = true
}
async function submit() {
  const current = manifest.value.versions.find((item) => item.id === editing.value)
  const ok = await save(
    () =>
      mutate(changeDocVersion, {
        action: 'save',
        id: form.id.trim(),
        name: form.name.trim(),
        source: editing.value ? '' : form.source,
        modeId: form.modeId,
        revision: current?.revision || 0,
      }),
    editing.value ? '版本名称已更新' : '草稿版本已创建',
  )
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
async function toggle(row) {
  const publish = row.status !== 'published'
  if (
    !(await confirm(
      publish
        ? `发布 ${row.name} 的全部文档，读者将可访问此版本。`
        : `${row.name} 将不再公开，已有链接将暂时无法访问。`,
      publish ? '发布版本' : '转为草稿',
    ))
  )
    return
  await save(
    () =>
      mutate(changeDocVersion, {
        action: 'status',
        id: row.id,
        revision: row.revision,
        status: publish ? 'published' : 'draft',
      }),
    publish ? '版本已发布' : '版本已转为草稿',
  )
}

async function more(command, row) {
  if (command === 'rename') {
    editing.value = row.id
    Object.assign(form, { id: row.id, name: row.name, modeId: row.modeId, source: '' })
    dialog.value = true
    return
  }
  if (command === 'default') {
    if (await confirm(`未指定版本的访问将默认打开 ${row.name}。`, '切换默认版本'))
      await save(
        () => mutate(changeDocVersion, { action: 'default', id: row.id, revision: row.revision }),
        '默认版本已更新',
      )
  } else if (
    command === 'delete' &&
    (await confirm(`永久删除草稿版本 ${row.name} 及其全部文档？其他版本不受影响。`, '删除版本'))
  )
    await save(
      () => mutate(changeDocVersion, { action: 'delete', id: row.id, revision: row.revision }),
      '草稿版本已删除',
    )
}
load()
</script>
