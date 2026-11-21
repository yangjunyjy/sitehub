<template>
  <AdminListPage
    :columns="columns"
    preference-key="tags"
    title="标签管理"
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
      ><el-form-item label="标签名称"
        ><el-input v-model="searchForm.name" placeholder="请输入标签名称" clearable /></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" @click="handleCreate">新增标签</el-button
      ><el-button :icon="Download" @click="handleExport">导出</el-button></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable v-bind="tableProps" :max-height="tableHeight" :data="tagList">
        <template #created_at="{ row }">
          <div class="time-cell">
            <el-icon><Clock /></el-icon>
            <span>{{ formatDate(row.created_at) }}</span>
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
      ><AdminFormDialog :loading="submitting" confirm-text="确定" @submit="handleSubmit"
        v-model="dialogVisible"
        :title="dialogTitle"
        width="500px"
        @closed="handleDialogClosed"
      >
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <TaxonomyFields v-model:name="form.name" v-model:slug="form.slug" kind="tags" />

          <el-form-item label="标签描述" prop="description">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="请输入标签描述"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
        </el-form>
      </AdminFormDialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import TaxonomyFields from '@/components/blog/TaxonomyFields.vue'
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue'
import { useAdminList } from '@/composables/useAdminList'
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
import { columns } from './columns'

defineOptions({
  name: 'TagList',
})

import { ref, reactive, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Download, Clock } from '@element-plus/icons-vue'
import { getTagList, createTag, updateTag, deleteTag } from '@/api/tag'

const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref()
const tagList = ref([])

const searchForm = reactive({
  name: '',
})
const appliedFilters = reactive({ ...searchForm })

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const dialogTitle = ref('新增标签')
const form = reactive({
  id: null,
  name: '',
  slug: '',
  description: '',
})

const formatDate = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// 自定义校验函数
const validateName = (rule, value, callback) => {
  if (!value || value.trim() === '') {
    callback(new Error('标签名称不能为空'))
  } else if (value.length > 50) {
    callback(new Error('标签名称长度不能超过50个字符'))
  } else {
    callback()
  }
}

const validateSlug = (rule, value, callback) => {
  if (!value || value.trim() === '') {
    callback(new Error('URL后缀不能为空'))
  } else if (value.length > 50) {
    callback(new Error('URL后缀长度不能超过50个字符'))
  } else if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value)) {
    callback(new Error('URL后缀只能包含小写字母、数字和连字符(-)，且不能以连字符开头或结尾'))
  } else {
    callback()
  }
}

const validateDescription = (rule, value, callback) => {
  if (value && value.length > 500) {
    callback(new Error('描述长度不能超过500个字符'))
  } else {
    callback()
  }
}

// 校验规则
const rules = {
  name: [
    { required: true, message: '请输入标签名称', trigger: 'blur' },
    { validator: validateName, trigger: 'blur' },
  ],
  slug: [
    { required: true, message: '请输入URL后缀', trigger: 'blur' },
    { validator: validateSlug, trigger: 'blur' },
  ],
  description: [{ validator: validateDescription, trigger: 'blur' }],
}

// 自动从名称生成slug
const autoGenerateSlug = () => {
  if (form.name && (!form.slug || form.slug === '')) {
    form.slug = form.name
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/^-+|-+$/g, '')
      .replace(/-+/g, '-')
  }
}

// 监听名称变化自动生成slug
watch(
  () => form.name,
  () => {
    if (!form.slug || form.slug === '') {
      autoGenerateSlug()
    }
  },
)

// 表单验证方法
const validateForm = () => {
  return new Promise((resolve, reject) => {
    if (!formRef.value) {
      reject(new Error('表单未初始化'))
      return
    }

    formRef.value.validate((valid) => {
      if (valid) {
        resolve({
          id: form.id,
          name: form.name.trim(),
          slug: form.slug.trim(),
          description: form.description ? form.description.trim() : null,
        })
      } else {
        reject(new Error('请检查表单填写是否正确'))
      }
    })
  })
}

// 加载标签列表
const { loading, loadError, loadList } = useAdminList({
  fetcher: getTagList,
  rows: tagList,
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

// 重置搜索
const handleReset = () => {
  searchForm.name = ''
  handleSearch()
}

// 新增标签
const handleCreate = () => {
  dialogTitle.value = '新增标签'
  Object.assign(form, {
    id: null,
    name: '',
    slug: '',
    description: '',
  })
  dialogVisible.value = true

  nextTick(() => {
    if (formRef.value) {
      formRef.value.clearValidate()
    }
  })
}

// 编辑标签
const handleEdit = (row) => {
  dialogTitle.value = '编辑标签'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    slug: row.slug,
    description: row.description || '',
  })
  dialogVisible.value = true

  nextTick(() => {
    if (formRef.value) {
      formRef.value.clearValidate()
    }
  })
}

// 对话框关闭时的处理
const handleDialogClosed = () => {
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    const formData = await validateForm()

    submitting.value = true

    if (form.id) {
      await updateTag(formData)
      ElMessage.success('更新成功')
    } else {
      await createTag(formData)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    loadList()
  } catch (error) {
    if (error.message !== '请检查表单填写是否正确') {
      ElMessage.error(error.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

// 删除标签
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm(
      '确定删除该标签吗？删除后无法恢复，且相关文章的标签将被移除。',
      '警告',
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
      },
    )

    await deleteTag(id)
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
  const dataStr = JSON.stringify(tagList.value, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `tags_${new Date().getTime()}.json`
  link.click()
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

// 组件挂载时加载列表
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
