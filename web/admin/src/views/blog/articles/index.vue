<template>
  <AdminListPage
    :columns="columns"
    preference-key="articles"
    title="文章管理"
    :filter-model="filters"
    :error="loadError"
    :loading="loading"
    v-model:current-page="pagination.page"
    v-model:page-size="pagination.size"
    :total="pagination.total"
    :page-sizes="[10, 20, 50]"
    @search="search"
    @reset="reset"
    @refresh="refresh"
    @page-change="loadList"
  >
    <template #filters
      ><el-form-item label="文章标题"
        ><el-input
          v-model="filters.title"
          placeholder="请输入文章标题"
          aria-label="文章标题搜索"
          clearable /></el-form-item
      ><el-form-item label="文章分类"
        ><el-select v-model="filters.categoryId" placeholder="全部分类" clearable
          ><el-option
            v-for="item in categories"
            :key="item.id"
            :label="item.name"
            :value="item.id" /></el-select></el-form-item
      ></template>
    <template #advanced>
      <el-form-item label="文章标签"
        ><el-select v-model="filters.tagId" placeholder="全部标签" clearable
          ><el-option
            v-for="item in tags"
            :key="item.id"
            :label="item.name"
            :value="item.id" /></el-select></el-form-item></template
    ><template #actions>
      <el-button type="primary" plain :icon="Plus" v-if="can('create')" @click="create"
        >新建文章</el-button
      >
      <el-button
        :icon="Download"
        :loading="exporting"
        :disabled="!selected.length || !!loadError"
        @click="exportSelected"
        >导出所选{{ selected.length ? ` (${selected.length})` : '' }}</el-button
      ></template
    >

    <template #tabs
      ><div class="article-status-tabs" role="tablist" aria-label="文章发布状态">
        <button
          v-for="tab in statuses"
          :key="tab.value"
          role="tab"
          :aria-selected="filters.status === tab.value"
          :class="{ active: filters.status === tab.value }"
          @click="changeStatus(tab.value)"
        >
          {{ tab.label }}
        </button>
      </div></template
    >
    <template #notice v-if="optionsError"
      >{{ optionsError
      }}<el-button link type="primary" @click="loadOptions">重试</el-button></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        ref="tableRef"
        :data="articles"
        row-key="id"
        class="admin-table article-table"
        @selection-change="selected = $event"
        :empty-text="
          loadError ? '文章加载失败，请点击刷新重试' : '暂无匹配文章，可以调整筛选或新建文章'
        "
      >
        <template #content="{ row }"
          ><div class="article-cell">
            <el-image v-if="row.cover" :src="row.cover" fit="cover" class="article-cover"
              ><template #error
                ><div class="cover-fallback">
                  <el-icon><Document /></el-icon></div></template
            ></el-image>
            <div v-else class="article-cover cover-fallback">
              <el-icon><Document /></el-icon>
            </div>
            <div class="article-copy">
              <button
                v-if="can('update') || can('select') || can('list')"
                @click="edit(row)"
                class="article-title"
              >
                {{ row.title || '未命名草稿' }}
              </button>
              <span v-else class="article-title">{{ row.title || '未命名草稿' }}</span>
              <p>{{ row.summary || '暂无摘要' }}</p>
            </div>
          </div></template
        >
        <template #taxonomy="{ row }"
          ><div class="category-name">{{ row.category?.name || '未分类' }}</div>
          <div class="article-tags">
            <span v-for="(tag, index) in row.tags.slice(0, 2)" :key="tag.id || index">{{
              tag.name
            }}</span
            ><small v-if="row.tags.length > 2">+{{ row.tags.length - 2 }}</small>
          </div></template
        >
        <template #status="{ row }"
          ><span class="article-state" :class="`status-${row.status}`"
            ><i />{{ row.deleted_at ? '回收站' : statusLabel(row.status) }}</span
          ></template
        >
        <template #views="{ row }">{{ Number(row.views || 0).toLocaleString() }}</template>
        <template #updatedAt="{ row }"
          ><div class="article-date">
            {{ formatDate(row.updatedAt) }}<small>{{ formatTime(row.updatedAt) }}</small>
          </div></template
        >
        <template #isTop="{ row }"
          ><el-switch
            :model-value="row.isTop"
            :loading="busy.has(row.id)"
            :disabled="!!loadError || !can('audit') || !!row.deleted_at"
            :aria-label="`置顶 ${row.title}`"
            @change="(value) => toggleTop(row, value)"
        /></template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              v-if="can('update') || can('select') || can('list')"
              :action="can('update') ? 'edit' : 'view'"
              @click="edit(row)" /><AdminRowAction
              action="export"
              :disabled="busy.has(row.id)"
              @click="exportOne(row)" /><AdminRowAction
              v-if="can('delete') && !row.deleted_at"
              action="delete"
              :disabled="busy.has(row.id) || !!loadError"
              @click="remove(row)" /><AdminRowAction
              v-if="can('delete') && row.deleted_at"
              action="restore"
              :disabled="busy.has(row.id)"
              @click="restore(row)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
  </AdminListPage>
</template>
<script setup>
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
import { columns } from './columns'

import { ref, reactive, onMounted, onActivated, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Download, Search, Refresh, RefreshLeft, Document } from '@element-plus/icons-vue'
import { getArticleList, getArticleDetail, deleteArticle, toggleArticleTop } from '@/api/article'
import { useRbacStore } from '@/store/rbac'
import { getArticleOptions, articleAction } from '@/api/article'
import { createUUID } from '@/utils/uuid'
import {
  articleListParams,
  normalizeArticle,
  articleError,
  articleMarkdown,
  articleFilename,
  saveBlob,
} from '@/utils/articles'
defineOptions({ name: 'ArticleList' })
const rbac = useRbacStore()
const can = (action) => rbac.hasPermission(`blog:article:${action}`)
const router = useRouter(),
  route = useRoute(),
  loading = ref(false),
  exporting = ref(false),
  loadError = ref(''),
  optionsError = ref(''),
  articles = ref([]),
  categories = ref([]),
  tags = ref([]),
  selected = ref([]),
  tableRef = ref(),
  busy = reactive(new Set())
const statuses = [
  { label: '全部文章', value: '' },
  { label: '已发布', value: '1' },
  { label: '草稿箱', value: '0' },
  { label: '定时发布', value: '2' },
  { label: '回收站', value: 'trash' },
]
const filters = reactive({
  title: String(route.query.keyword || ''),
  status: ['0', '1', '2'].includes(route.query.status) ? route.query.status : '',
  categoryId: '',
  tagId: '',
})
const appliedFilters = reactive({ ...filters })
const pagination = reactive({ page: 1, size: 10, total: 0 })
let requestID = 0,
  mounted = false
const statusLabel = (value) => ({ 0: '草稿', 1: '已发布', 2: '定时发布' })[value] || '未知状态'
const formatDate = (value) =>
  value ? new Date(value).toLocaleDateString('zh-CN').replaceAll('/', '-') : '—'
const formatTime = (value) =>
  value ? new Date(value).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) : ''
async function loadList() {
  const id = ++requestID
  loading.value = true
  loadError.value = ''
  selected.value = []
  try {
    const result = await getArticleList(
      articleListParams(appliedFilters, pagination.page, pagination.size),
      { silentError: true },
    )
    if (id !== requestID) return
    articles.value = (Array.isArray(result.data?.items) ? result.data.items : [])
      .filter(Boolean)
      .map(normalizeArticle)
    pagination.total = Number(result.data?.total || 0)
    tableRef.value?.clearSelection()
  } catch (error) {
    if (id === requestID) {
      loadError.value = articleError(error, '文章列表加载失败')
      articles.value = []
      pagination.total = 0
    }
  } finally {
    if (id === requestID) loading.value = false
  }
}
async function loadOptions() {
  optionsError.value = ''
  try {
    const { data } = await getArticleOptions()
    categories.value = data.categories
    tags.value = data.tags
  } catch (e) {
    optionsError.value = articleError(e)
  }
}
function search() {
  Object.assign(appliedFilters, filters)
  pagination.page = 1
  loadList()
}
function changeStatus(status) {
  filters.status = status
  search()
}
function reset() {
  Object.assign(filters, { title: '', status: '', categoryId: '', tagId: '' })
  search()
}
function refresh() {
  loadOptions()
  loadList()
}
const create = () => router.push(router.page('blog/articles/create/index'))
const edit = (row) => router.push(router.page('blog/articles/editor/index', { params: { id: row.id } }))
async function toggleTop(row, value) {
  if (busy.has(row.id)) return
  busy.add(row.id)
  try {
    const response = await toggleArticleTop(row.id, value, row.version)
    Object.assign(row, normalizeArticle(response.data))
    ElMessage.success(value ? '文章已置顶' : '已取消置顶')
  } catch (error) {
    ElMessage.error(articleError(error))
  } finally {
    busy.delete(row.id)
  }
}
async function remove(row) {
  try {
    await ElMessageBox.confirm(`确定删除“${row.title}”？文章将移入回收站，可以恢复。`, '删除文章', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  busy.add(row.id)
  try {
    await deleteArticle(row.id, row.version)
    if (articles.value.length === 1 && pagination.page > 1) pagination.page--
    await loadList()
    ElMessage.success('文章已删除')
  } catch (error) {
    ElMessage.error(articleError(error))
  } finally {
    busy.delete(row.id)
  }
}
async function restore(row) {
  if (busy.has(row.id)) return
  busy.add(row.id)
  try {
    await articleAction('restore', {
      id: row.id,
      version: row.version,
      request_id: createUUID(),
    })
    if (articles.value.length === 1 && pagination.page > 1) pagination.page--
    await loadList()
    ElMessage.success('文章已恢复到草稿箱')
  } catch (error) {
    ElMessage.error(articleError(error))
  } finally {
    busy.delete(row.id)
  }
}
async function exportOne(row) {
  busy.add(row.id)
  try {
    const detail = (await getArticleDetail(row.id)).data
    saveBlob(
      new Blob([articleMarkdown(detail)], { type: 'text/markdown;charset=utf-8' }),
      articleFilename(detail),
    )
    ElMessage.success('文章已导出')
  } catch (error) {
    ElMessage.error(articleError(error, '导出失败'))
  } finally {
    busy.delete(row.id)
  }
}
async function exportSelected() {
  if (!selected.value.length || exporting.value) return
  exporting.value = true
  const rows = [...selected.value]
  try {
    const JSZip = (await import('jszip')).default,
      zip = new JSZip()
    for (const row of rows) {
      const detail = (await getArticleDetail(row.id)).data
      zip.file(articleFilename(detail), articleMarkdown(detail))
    }
    saveBlob(
      await zip.generateAsync({ type: 'blob' }),
      `articles-${new Date().toISOString().slice(0, 10)}.zip`,
    )
    ElMessage.success(`已导出 ${rows.length} 篇完整文章`)
  } catch (error) {
    ElMessage.error(articleError(error, '导出失败'))
  } finally {
    exporting.value = false
  }
}
watch(
  () => [route.query.keyword, route.query.status],
  () => {
    if (route.meta.componentPath !== 'blog/articles/index') return
    filters.title = String(route.query.keyword || '')
    filters.status = ['0', '1', '2'].includes(route.query.status) ? route.query.status : ''
    search()
  },
)
onMounted(() => {
  mounted = true
  refresh()
})
onActivated(() => {
  if (mounted && !loading.value) loadList()
})
</script>
<style lang="scss" scoped src="./articles.scss"></style>
