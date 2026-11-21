<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="min(900px, calc(100vw - 24px))"
    :close-on-click-modal="false"
    :destroy-on-close="true"
    class="role-assign-dialog"
  >
    <!-- 中间区域 -->
    <div class="dialog-content">
      <!-- 左侧区域 -->
      <div class="left-section">
        <AdminSearchForm :model="searchForm" :loading="loading" @search="handleSearch" @reset="resetSearch">
          <el-form-item label="角色名称">
            <el-input v-model="searchForm.roleName" placeholder="角色名称" clearable />
          </el-form-item>
          <template #advanced>
            <el-form-item label="角色编码">
              <el-input v-model="searchForm.roleCode" placeholder="角色编码" clearable />
            </el-form-item>
          </template>
        </AdminSearchForm>
        <el-alert v-if="loadError" :title="loadError" type="error" :closable="false" />
        <!-- 左中：角色列表 -->
        <div class="role-list-area">
          <el-table
            ref="table"
            :data="rows"
            row-key="id"
            v-loading="loading"
            height="320"
            @select="toggleRole"
            @select-all="syncVisibleSelection"
            class="role-table"
          >
            <el-table-column type="selection" width="55" align="center" />
            <el-table-column prop="roleName" label="角色名称" min-width="150" />
            <el-table-column
              prop="roleDesc"
              label="角色描述"
              min-width="200"
              show-overflow-tooltip
            />
          </el-table>
        </div>

        <!-- 左下：分页器 -->
        <div class="pagination-area">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[10, 20, 50]"
            size="small"
            layout="total, prev, pager, next, jumper"
            @change="loadList"
          />
        </div>
      </div>

      <!-- 右侧区域：已选角色 -->
      <div class="right-section">
        <div class="selected-header">
          <span class="selected-title">已选角色</span>
          <span class="selected-count">({{ selectedRoles.length }})</span>
        </div>
        <div class="selected-list">
          <el-empty v-if="selectedRoles.length === 0" description="暂未选择角色" :image-size="80" />
          <div v-else class="selected-items">
            <div v-for="role in selectedRoles" :key="role.id" class="selected-item">
              <div class="role-info">
                <el-tag :type="getRoleTagType(role)" size="default">
                  {{ role.roleName }}
                </el-tag>
                <span class="role-code">{{ role.roleCode }}</span>
              </div>
              <el-button
                type="danger"
                link
                size="small"
                :icon="Close"
                @click="handleRemoveRole(role.id)"
              >
                移除
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel" size="default">取消</el-button>
        <el-button type="primary" @click="handleConfirm" :loading="submitting" size="default">
          确认
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, nextTick, onMounted } from 'vue'
import AdminSearchForm from '@/components/admin/AdminSearchForm.vue'
import { useAdminList } from '@/composables/useAdminList'
import { Close } from '@element-plus/icons-vue'
import { getRoleList } from '@/api/role'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '分配角色',
  },
  userId: {
    type: [Number, String],
    default: null,
  },
  userName: {
    type: String,
    default: '',
  },
  existingRoleIds: {
    type: Array,
    default: () => [],
  },
})

// Emits
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
const table = ref()
const searchForm = reactive({ roleName: '', roleCode: '' })
const applied = reactive({ ...searchForm })
const submitting = ref(false)
const selectedRoles = ref([])
const rows = ref([])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const dialogTitle = computed(() => props.title)
const getRoleTagType = role => ({ admin: 'danger', editor: 'success', auditor: 'info' })[role.roleCode] || 'info'
const { loading, loadError, loadList } = useAdminList({
  fetcher: getRoleList,
  rows,
  pagination,
  params: () => ({ ...applied, status: 1, page: pagination.page, pageSize: pagination.pageSize }),
  selectRows: data => data.items || [],
  errorMessage: '角色加载失败，请重新查询',
  onLoaded: async () => {
    await nextTick()
    for (const role of rows.value) {
      if (selectedRoles.value.some(selected => selected.id === role.id)) table.value?.toggleRowSelection(role, true)
    }
  },
})

onMounted(() => {
  selectedRoles.value = props.existingRoleIds.map(role =>
    typeof role === 'object' ? role : { id: role, roleName: String(role) })
  resetSearch()
})

function handleSearch() {
  Object.assign(applied, { roleName: searchForm.roleName.trim(), roleCode: searchForm.roleCode.trim() })
  pagination.page = 1
  loadList()
}

function resetSearch() {
  Object.assign(searchForm, { roleName: '', roleCode: '' })
  handleSearch()
}

function toggleRole(selection, row) {
  const others = selectedRoles.value.filter(role => role.id !== row.id)
  selectedRoles.value = selection.some(role => role.id === row.id) ? [...others, row] : others
}

function syncVisibleSelection(selection) {
  const visibleIds = new Set(rows.value.map(role => role.id))
  selectedRoles.value = [...selectedRoles.value.filter(role => !visibleIds.has(role.id)), ...selection]
}

function handleRemoveRole(id) {
  selectedRoles.value = selectedRoles.value.filter(role => role.id !== id)
  const row = rows.value.find(role => role.id === id)
  if (row) table.value?.toggleRowSelection(row, false)
}

// 取消
const handleCancel = () => {
  visible.value = false
  emit('cancel')
}

// 确认
const handleConfirm = async () => {
  submitting.value = true
  try {
    const roleIds = selectedRoles.value.map((role) => role.id)
    emit('confirm', {
      userId: props.userId,
      roleIds: roleIds,
    })
    visible.value = false
  } finally {
    submitting.value = false
  }
}

</script>

<style scoped lang="scss">
.role-assign-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px;
    border-bottom: 1px solid #ebeef5;
    margin: 0;

    .el-dialog__title {
      font-size: 18px;
      font-weight: 600;
      color: #303133;
    }

    .el-dialog__headerbtn {
      top: 20px;

      .el-dialog__close {
        font-size: 20px;
        color: #606266;

        &:hover {
          color: #409eff;
        }
      }
    }
  }

  :deep(.el-dialog__body) {
    padding: 24px;
  }

  :deep(.el-dialog__footer) {
    padding: 16px 24px;
    border-top: 1px solid #ebeef5;
    margin: 0;
  }
}

.dialog-content {
  display: flex;
  gap: 20px;
  height: 420px;
}

.left-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;

  min-width: 0;
  .admin-search-card { padding: 0; }
  .role-list-area {
    flex: 1;
    overflow: hidden;

    .role-table {
      :deep(.el-table__header) {
        th {
          background-color: #f5f7fa;
        }
      }

      .disabled-text {
        color: #909399;
        font-size: 13px;
      }
    }
  }

  .pagination-area {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    padding: 8px 0;

    :deep(.el-pagination) {
      font-size: 13px;

      .el-pager li {
        min-width: 28px;
        height: 28px;
        line-height: 26px;
        margin: 0 2px;
        font-size: 12px;
      }

      .btn-prev,
      .btn-next {
        padding: 0 6px;
        min-width: 28px;
        height: 28px;
      }

      .el-pagination__jump {
        margin-left: 8px;

        .el-input__wrapper {
          padding: 0 6px;
          width: 50px;
        }

        .el-input__inner {
          text-align: center;
        }
      }

      .el-pagination__total,
      .el-pagination__jump {
        font-size: 12px;
        color: #303133;
      }
    }
  }
}

.right-section {
  width: 280px;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #ebeef5;
  padding-left: 20px;

  .selected-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-bottom: 16px;
    border-bottom: 1px solid #ebeef5;
    margin-bottom: 16px;

    .selected-title {
      font-size: 15px;
      font-weight: 600;
      color: #303133;
    }

    .selected-count {
      font-size: 14px;
      color: #606266;
    }
  }

  .selected-list {
    flex: 1;
    overflow-y: auto;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-thumb {
      background: #ecf5ff;
      border-radius: 3px;

      &:hover {
        background: #ecf5ff;
      }
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }
  }

  .selected-items {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .selected-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    background: #f5f7fa;
    border-radius: 6px;
    transition: all 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    .role-info {
      display: flex;
      flex-direction: column;
      gap: 6px;
      flex: 1;

      .role-code {
        font-size: 12px;
        color: #606266;
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;

  .el-button {
    padding: 10px 28px;
  }
}
@media (max-width: 760px) {
  .dialog-content { flex-direction: column; height: auto; }
  .right-section { width: auto; padding: 0; border-left: 0; }
  .selected-list { max-height: 150px; }
}
</style>
