<template>
  <AdminListPage
    :columns="columns"
    preference-key="categories"
    title="分类管理"
    :filter-model="searchForm"
    :loading="loading"
    :error="loadError"
    v-model:current-page="pagination.page"
    v-model:page-size="pagination.pageSize"
    :total="pagination.total"
    @search="handleSearch"
    @reset="handleReset"
    @refresh="loadList"
    @page-change="loadList"
  >
    <template #filters
      ><el-form-item label="分类名称"
        ><el-input v-model="searchForm.name" placeholder="请输入分类名称" clearable /></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" @click="handleCreate">新增分类</el-button
      ><el-button :icon="Download" @click="handleExport">导出</el-button></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable v-bind="tableProps" :max-height="tableHeight" :data="categoryList">
        <template #name="{row}"><ResourceIcon :name="row.icon" /> {{ row.name }}</template>
        <template #created_at="{ row }">
          <div class="time-cell">
            <el-icon><Clock /></el-icon>
            <span>{{ formatDate(row.created_at) }}</span>
          </div>
        </template>
        <template #updated_at="{ row }">
          <div class="time-cell">
            <el-icon><Clock /></el-icon>
            <span>{{ formatDate(row.updated_at) }}</span>
          </div>
        </template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction action="edit" @click="handleEdit(row)" /><AdminRowAction
              action="delete"
              @click="handleDelete(row.id)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
    <template #dialogs
      ><AdminFormDialog :loading="submitting" confirm-text="确定" @submit="handleSubmit" v-model="dialogVisible" :title="dialogTitle" width="500px">
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <TaxonomyFields v-model:name="form.name" v-model:slug="form.slug" kind="categories" />
          <el-form-item label="分类图标" prop="icon">
            <IconPicker v-model="form.icon" />
          </el-form-item>
          <el-form-item label="排序" prop="sort_order">
            <el-input-number v-model="form.sort_order" :min="0" placeholder="请输入排序" />
          </el-form-item>
        </el-form>
      </AdminFormDialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import TaxonomyFields from '@/components/blog/TaxonomyFields.vue'
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue'
import IconPicker from '@/components/icons/IconPicker.vue'
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
import { useAdminList } from '@/composables/useAdminList'
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
import { columns } from './columns'

defineOptions({
  name: 'CategoryList',
})

import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Download, Clock } from '@element-plus/icons-vue'
import { getCategoryList, createCategory, updateCategory, deleteCategory } from '@/api/category'

const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const categoryList = ref([])

const searchForm = reactive({
  name: '',
})
const appliedFilters = reactive({ ...searchForm })

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const dialogTitle = ref('新增分类')
const form = reactive({
  id: null,
  name: '',
  slug: '',
  icon: 'data/folder',
  sort_order: 0,
})

const rules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  slug: [{ required: true, message: '请输入分类标识', trigger: 'blur' }],
}

const formatDate = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const { loading, loadError, loadList } = useAdminList({
  fetcher: getCategoryList,
  rows: categoryList,
  pagination,
  params: () => ({
    page: pagination.page,
    pageSize: pagination.pageSize,
    name: appliedFilters.name.trim() || undefined,
  }),
  selectRows: (data) => (Array.isArray(data.items) ? data.items : []),
  errorMessage: '加载失败，请重试',
})

const handleSearch = () => {
  Object.assign(appliedFilters, searchForm)
  pagination.page = 1
  loadList()
}

const handleReset = () => {
  searchForm.name = ''
  handleSearch()
}

const handleCreate = () => {
  dialogTitle.value = '新增分类'
  Object.assign(form, {
    id: null,
    name: '',
    slug: '',
    icon: 'data/folder',
    sort_order: 0,
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑分类'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    slug: row.slug,
    icon: row.icon || '',
    sort_order: row.sort_order || 0,
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (form.id) {
        await updateCategory(form.id, form)
        ElMessage.success('更新成功')
      } else {
        await createCategory(form)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (error) {
      ElMessage.error('操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除该分类吗？', '提示', { type: 'warning' })
    await deleteCategory(id)
    ElMessage.success('删除成功')
    loadList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 导出
const handleExport = () => {
  // 导出分类数据为JSON
  const dataStr = JSON.stringify(categoryList.value, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `categories_${new Date().getTime()}.json`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

onMounted(() => {
  loadList()
})
</script>

<style lang="scss" scoped>
.time-cell,
.action-btns {
  display: flex;
  align-items: center;
  gap: 8px;
}
.time-cell {
  color: #606266;
  font-size: 12px;
  justify-content: center;
}
</style>
