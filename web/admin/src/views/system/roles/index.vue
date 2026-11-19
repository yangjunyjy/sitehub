<template>
  <AdminListPage
    :columns="columns"
    preference-key="roles"
    title="角色管理"
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
      ><el-form-item label="角色名称"
        ><el-input v-model="searchForm.name" placeholder="请输入角色名称" clearable /></el-form-item
      ><el-form-item label="角色编码"
        ><el-input v-model="searchForm.code" placeholder="请输入角色编码" clearable /></el-form-item
      ><el-form-item label="角色状态"
        ><el-select v-model="searchForm.status" placeholder="全部状态" clearable
          ><el-option label="启用" :value="1" /><el-option
            label="禁用"
            :value="0" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" @click="handleCreate"
        >新增角色</el-button
      ></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable v-bind="tableProps" :max-height="tableHeight" :data="roleList">
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction action="edit" @click="handleEdit(row)" /><AdminRowAction
              action="assignPermissions"
              @click="handleAssign(row)" /><AdminRowAction
              action="delete"
              @click="handleDelete(row.id)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
    <template #dialogs
      ><el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <el-form-item label="角色名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入角色名称" />
          </el-form-item>
          <el-form-item label="角色编码" prop="code">
            <el-input v-model="form.code" placeholder="请输入角色编码" />
          </el-form-item>
          <el-form-item label="描述" prop="description">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="请输入描述"
            />
          </el-form-item>
          <el-form-item label="状态" prop="status">
            <el-radio-group v-model="form.status">
              <el-radio :value="1">启用</el-radio>
              <el-radio :value="0">禁用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
        </template>
      </el-dialog>

      <!-- 分配权限对话框 -->
      <el-dialog
        v-model="assignDialogVisible"
        title="分配权限"
        width="600px"
        @open="GetRoleMenus(currentRole.id)"
        @close="DestroyLastRoleData"
      >
        <el-form :model="assignForm" ref="assignFormRef" label-width="100px">
          <el-form-item label="角色名称">
            <span>{{ currentRole?.roleName }}</span>
          </el-form-item>
          <el-form-item label="权限">
            <el-tree
              ref="menuTreeRef"
              :data="menuTree"
              :props="{ label: 'title', children: 'children' }"
              node-key="id"
              show-checkbox
              :default-checked-keys="checkedMenuIds"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="assignDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleAssignSubmit" :loading="assigning"
            >确定</el-button
          >
        </template>
      </el-dialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import { useAdminList } from '@/composables/useAdminList'
import AdminRowAction from '@/components/admin/AdminRowAction.vue'
import AdminRowActions from '@/components/admin/AdminRowActions.vue'
import AdminListPage from '@/components/admin/AdminListPage.vue'
import AdminDataTable from '@/components/admin/AdminDataTable.vue'
const columns = [
  {
    key: 'id',
    prop: 'id',
    label: 'ID',
    width: '70',
    align: 'center',
  },
  {
    key: 'roleName',
    prop: 'roleName',
    label: '角色名称',
    minWidth: '120',
  },
  {
    key: 'roleCode',
    prop: 'roleCode',
    label: '角色编码',
    minWidth: '120',
  },
  {
    key: 'roleDesc',
    prop: 'roleDesc',
    label: '描述',
    minWidth: '200',
  },
  {
    key: 'status',
    label: '状态',
    width: '80',
  },
  {
    key: 'actions',
    label: '操作',
    width: '184',
    fixed: 'right',
    utility: true,
  },
]

defineOptions({
  name: 'RoleList',
})

import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  assignRoleMenus,
  getAllMenus,
  GetMenusByRoleID,
  getRoleList,
  createRole,
  updateRole,
  deleteRole,
} from '@/api'

const dialogVisible = ref(false)
const assignDialogVisible = ref(false)
const submitting = ref(false)
const assigning = ref(false)
const formRef = ref(null)
const assignFormRef = ref(null)
const menuTree = ref([])
const menuTreeRef = ref(null)
const roleList = ref([])
const currentRole = ref(null)
const checkedMenuIds = ref([])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const searchForm = reactive({
  name: '',
  code: '',
  status: '',
})
const appliedFilters = reactive({ ...searchForm })

const dialogTitle = ref('新增角色')
const form = reactive({
  id: null,
  name: '',
  code: '',
  description: '',
  status: 1,
})

const assignForm = reactive({
  menuIds: [],
})

const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
}

const { loading, loadError, loadList } = useAdminList({
  fetcher: getRoleList,
  rows: roleList,
  pagination,
  params: () => ({
    roleName: appliedFilters.name.trim() || undefined,
    roleCode: appliedFilters.code.trim() || undefined,
    status: appliedFilters.status === '' ? undefined : appliedFilters.status,
    page: pagination.page,
    pageSize: pagination.pageSize,
  }),
  selectRows: (data) => (Array.isArray(data.list) ? data.list : []),
  errorMessage: '加载失败，请重试',
})

const handleSearch = () => {
  Object.assign(appliedFilters, searchForm)
  pagination.page = 1
  loadList()
}

const handleReset = () => {
  Object.assign(searchForm, { name: '', code: '', status: '' })
  handleSearch()
}

const handleCreate = () => {
  dialogTitle.value = '新增角色'
  Object.assign(form, {
    id: null,
    name: '',
    code: '',
    description: '',
    status: 1,
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑角色'
  Object.assign(form, {
    id: row.id,
    name: row.roleName,
    code: row.roleCode,
    description: row.roleDesc,
    status: row.status,
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
        const params = {
          id: form.id,
          role_name: form.name,
          role_code: form.code,
          status: form.status,
          role_desc: form.description,
        }
        const data = await updateRole(params)
        if (data.code == 0) {
          ElMessage.success('更新成功')
        }
      } else {
        const params = {
          role_code: form.code,
          role_name: form.name,
          role_desc: form.description,
          status: form.status,
        }
        const data = await createRole(params)

        if (data.code == 0) {
          ElMessage.success('创建成功')
        }
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
    await ElMessageBox.confirm('确定删除该角色吗？', '提示', { type: 'warning' })
    const data = await deleteRole(id)
    if (!data) return
    ElMessage.success('删除成功')
    loadList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleAssign = (row) => {
  currentRole.value = row
  checkedMenuIds.value = [] // 先清空，等待对话框打开时重新获取
  assignDialogVisible.value = true
}

const handleAssignSubmit = async () => {
  assigning.value = true
  try {
    // 获取所有完全选中的节点（包括父节点自动选中的子节点）
    const checkedKeys = menuTreeRef.value.getCheckedKeys()
    // 获取半选中的节点（父节点下只有部分子节点被选中）
    const halfCheckedKeys = menuTreeRef.value.getHalfCheckedKeys()
    // 合并：完全选中的节点 + 半选中的父节点
    const allMenuIds = [...checkedKeys, ...halfCheckedKeys]


    const roleId = currentRole.value.id
    const params = {
      role_id: roleId,
      menu_ids: allMenuIds,
    }
    const data = await assignRoleMenus(params)

    if (data.code == 0) {
      ElMessage.success('分配权限成功')
    }
    assignDialogVisible.value = false
  } catch (error) {
    ElMessage.error('分配权限失败')
  } finally {
    assigning.value = false
  }
}

const GetRoleMenus = async (id) => {
  const dataRole = await GetMenusByRoleID(id)
  const dataAll = await getAllMenus()
  // 转化菜单数据为树形结构
  menuTree.value = buildMenusTree(dataAll.data)

  // 只保留叶子节点（没有子菜单的节点）在选中列表中
  // 这样父菜单会根据子菜单的选中状态自动显示为半选中或全选中
  const isLeafNode = (menuId, allMenus) => {
    return !allMenus.some((m) => m.parent_id === menuId)
  }

  checkedMenuIds.value = dataRole.data
    .filter((menu) => isLeafNode(menu.id, dataAll.data))
    .map((menu) => menu.id)


  // 等待树形结构渲染完后设置选中状态
  nextTick(() => {
    if (menuTreeRef.value) {
      menuTreeRef.value.setCheckedKeys(checkedMenuIds.value)
    }
  })
}

// 把扁平化的菜单数组转化成菜单树
const buildMenusTree = (menus, parent_id = 0) => {
  return menus
    .filter((menu) => menu.parent_id == parent_id && menu.status == 1)
    .map((menu) => ({
      ...menu,
      title: menu.menu_name,
      children: buildMenusTree(menus, menu.id),
    }))
}

const DestroyLastRoleData = () => {
  if (menuTreeRef.value) {
    menuTreeRef.value.setCheckedKeys([])
  }
  menuTree.value = []
  checkedMenuIds.value = []
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
