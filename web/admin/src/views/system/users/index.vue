<template>
  <AdminListPage
    :columns="columns"
    preference-key="users"
    title="用户管理"
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
      ><el-form-item label="用户名"
        ><el-input
          v-model="searchForm.username"
          placeholder="请输入用户名"
          clearable /></el-form-item
      ><el-form-item label="邮箱"
        ><el-input v-model="searchForm.email" placeholder="请输入邮箱" clearable /></el-form-item
      ><el-form-item label="用户状态"
        ><el-select v-model="searchForm.status" placeholder="全部状态" clearable
          ><el-option label="启用" :value="1" /><el-option
            label="禁用"
            :value="0" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" @click="handleCreate"
        >新增用户</el-button
      ></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable v-bind="tableProps" :max-height="tableHeight" :data="userList" row-key="id">
        <template #avatar="{ row }">
          <el-avatar :size="40" :src="row.avatar" />
        </template>
        <template #roles="{ row }">
          <div v-if="row.roles.length" class="user-role-tags">
            <el-tag
              v-for="role in row.roles.slice(0, 2)"
              :key="role.id"
              :type="getRoleTagType(role.roleCode)"
              size="small"
              :title="role.roleName"
            >
              {{ role.roleName }}
            </el-tag>
            <el-popover v-if="row.roles.length > 2" trigger="click" placement="bottom" :width="260">
              <template #reference
                ><el-button
                  link
                  type="primary"
                  :aria-label="`查看 ${row.username} 的全部 ${row.roles.length} 个角色`"
                  >+{{ row.roles.length - 2 }}</el-button
                ></template
              >
              <div class="user-all-roles" aria-label="全部角色">
                <el-tag v-for="role in row.roles" :key="role.id" size="small">{{
                  role.roleName
                }}</el-tag>
              </div>
            </el-popover>
          </div>
          <span v-else class="user-role-empty">未分配</span>
        </template>
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </el-tag>
        </template>
        <template #createdAt="{ row }">
          {{ formatDate(row.createTime) }}
        </template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              v-if="isAdmin()"
              action="edit"
              @click="handleEdit(row)" /><AdminRowAction
              action="assignRoles"
              @click="handleAssignRoles(row)" /><AdminRowAction
              v-if="isAdmin()"
              action="delete"
              @click="handleDelete(row.id)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
    <template #dialogs
      ><el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="form.username" placeholder="请输入用户名" :disabled="!!form.id" />
          </el-form-item>
          <el-form-item label="昵称" prop="nickname">
            <el-input v-model="form.nickname" placeholder="请输入昵称" />
          </el-form-item>
          <el-form-item v-if="!form.id" label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              show-password
              autocomplete="new-password"
              placeholder="请输入密码"
              :disabled="!!form.id"
            />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="form.email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="性别" prop="gender">
            <el-radio-group v-model="form.gender">
              <el-radio :value="0">男</el-radio>
              <el-radio :value="1">女</el-radio>
              <el-radio :value="2">未知</el-radio>
            </el-radio-group>
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

      <!-- 分配角色对话框 -->
      <RoleAssignDialog
        v-model="assignDialogVisible"
        :user-id="currentUser?.id"
        :existing-role-ids="currentUser?.roles || []"
        @confirm="handleAssignConfirm"
      />
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
    key: 'avatar',
    prop: 'avatar',
    label: '头像',
    width: '100',
  },
  {
    key: 'username',
    prop: 'username',
    label: '用户名',
    minWidth: '120',
  },
  {
    key: 'nickname',
    prop: 'nickname',
    label: '昵称',
    minWidth: '120',
  },
  {
    key: 'email',
    prop: 'email',
    label: '邮箱',
    minWidth: '180',
  },
  {
    key: 'roles',
    prop: 'roles',
    label: '角色',
    minWidth: '220',
  },
  {
    key: 'status',
    label: '状态',
    width: '80',
  },
  {
    key: 'createdAt',
    label: '创建时间',
    width: '160',
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
  name: 'UserList',
})

import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import RoleAssignDialog from './components/roleAssign'
import { useRbacStore } from '@/store/rbac'
import { getUserList, assignUserRoles, deleteUser, updateUser, createUser } from '@/api/user'
import { normalizeUserRow } from '@/utils/users'

const dialogVisible = ref(false)
const assignDialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const userList = ref([])
const currentUser = ref(null)
const currentLoginUser = ref(null)
const rbacStore = useRbacStore()

// 分页
currentLoginUser.value = rbacStore.currentUser
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const searchForm = reactive({
  username: '',
  email: '',
  status: '',
})
const appliedFilters = reactive({ ...searchForm })

const dialogTitle = ref('新增用户')
const form = reactive({
  id: null,
  nickname: '',
  username: '',
  password: '',
  gender: 0,
  email: '',
  status: 1,
})

const isAdmin = () => {
  return currentLoginUser.value.isAdmin === 1
}
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const getRoleTagType = (roleCode) => {
  if (roleCode == 'admin') {
    return 'danger'
  }
  return 'info'
}

const { loading, loadError, loadList } = useAdminList({
  fetcher: getUserList,
  rows: userList,
  pagination,
  params: () => ({
    username: appliedFilters.username.trim() || undefined,
    email: appliedFilters.email.trim() || undefined,
    status: appliedFilters.status === '' ? undefined : appliedFilters.status,
    page: pagination.page,
    pageSize: pagination.pageSize,
  }),
  selectRows: (data) => (Array.isArray(data.list) ? data.list.map(normalizeUserRow) : []),
  errorMessage: '加载失败，请重试',
})

const handleSearch = () => {
  Object.assign(appliedFilters, searchForm)
  pagination.page = 1
  loadList()
}

const handleReset = () => {
  Object.assign(searchForm, { username: '', email: '', status: '' })
  handleSearch()
}

const handleCreate = () => {
  dialogTitle.value = '新增用户'
  Object.assign(form, {
    id: null,
    username: '',
    nickname: '',
    password: '',
    gender: 0,
    email: '',
    status: 1,
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑用户'
  Object.assign(form, {
    id: row.id,
    username: row.username,
    nickname: row.nickname,
    email: row.email,
    gender: row.gender,
    status: row.status,
    password: '',
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
        const updateParams = {
          id: form.id,
          nickname: form.nickname,
          email: form.email,
          status: Number(form.status),
          gender: Number(form.gender),
          username: form.username,
        }
        const data = await updateUser(updateParams)
        if (data.code != 0) {
          ElMessage.error('更新用户失败')
          return
        }
        ElMessage.success('更新成功')
      } else {
        const createParams = {
          username: form.username,
          nickname: form.nickname,
          password: form.password,
          email: form.email,
          status: Number(form.status),
          gender: Number(form.gender),
        }
        const data = await createUser(createParams)
        if (data.code != 0) {
          ElMessage.error('创建用户失败')
          submitting.value = false
          return
        }
        ElMessage.success('创建用户成功')
      }
      dialogVisible.value = false
      loadList()
    } catch (error) {
    } finally {
      submitting.value = false
    }
  })
}

const handleAssignRoles = (row) => {
  currentUser.value = row
  assignDialogVisible.value = true
}

const handleAssignConfirm = async (data) => {
  try {
    const assignParams = {
      userId: data.userId || data.user_id || data.userID,
      roleIds: data.roleIds || data.role_ids || data.role_IDs,
    }
    const value = await assignUserRoles(assignParams)
    if (value.code != 0) {
      ElMessage.error('分配角色失败')
      return
    }
    ElMessage.success('分配角色成功')
    // 如果分配的是当前登录用户，刷新权限和菜单
    const userInfo = JSON.parse(sessionStorage.getItem('userInfo') || '{}')
    if (currentUser.value.email === userInfo.email) {
      rbacStore.refreshUserPermission()
    }
    loadList()
  } catch (error) {
    ElMessage.error('分配角色失败')
  }
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确定删除该用户吗？', '提示', { type: 'warning' })
  const data = await deleteUser(id)
  if (data.code != 0) {
    ElMessage.success('删除失败')
    return
  }
  ElMessage.success('删除成功')
  loadList()
}

onMounted(() => {
  loadList()
})
</script>

<style lang="scss" scoped>
.user-role-tags {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: nowrap;
}
.user-role-tags .el-tag {
  max-width: 125px;
  flex-shrink: 1;
  min-width: 0;
}
.user-role-tags :deep(.el-tag__content) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.user-role-tags .el-button {
  flex-shrink: 0;
}
.user-role-empty {
  color: #606266;
}
.user-all-roles {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  max-height: 240px;
  overflow: auto;
}
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
