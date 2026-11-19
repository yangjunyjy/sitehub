<template>
  <AdminListPage
    class="menu-list-page"
    :columns="columns"
    preference-key="menus"
    title="菜单管理"
    :filter-model="searchForm"
    :loading="loading"
    :error="loadError"
    v-model:current-page="pagination.page"
    v-model:page-size="pagination.pageSize"
    :total="allMenuList.length"
    @search="handleSearch"
    @reset="handleReset"
    @refresh="loadList"
    @page-change="handlePageChange"
  >
    <template #filters
      ><el-form-item label="所属系统"
        ><el-select v-model="searchForm.systemCode" placeholder="选择系统"
          ><el-option
            v-for="system in systems"
            :key="system.code"
            :value="system.code"
            :label="system.name" /></el-select></el-form-item
      ><el-form-item label="菜单名称"
        ><el-input
          v-model="searchForm.title"
          placeholder="请输入菜单名称"
          clearable /></el-form-item
      ><el-form-item label="菜单类型"
        ><el-select v-model="searchForm.type" placeholder="全部类型" clearable
          ><el-option label="目录" :value="1" /><el-option
            label="菜单"
            :value="2" /><el-option
            label="按钮"
            :value="3" /></el-select></el-form-item
    ></template>
    <template #actions
      ><el-button type="primary" plain :icon="Plus" @click="handleCreate"
        >新增菜单</el-button
      ><el-button @click="handleExpandAll">{{
        isExpanded ? "折叠菜单" : "展开菜单"
      }}</el-button
      ><span class="menu-tree-count"
        >共 {{ menuCount }} 项 · 按顶层菜单分页</span
      ></template
    >
    <template #default="{ tableHeight, tableProps, fullscreen }"
      ><AdminDataTable
        v-bind="tableProps"
        :max-height="fullscreen ? tableHeight : ''"
        :data="displayedMenuList"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        ref="tableRef"
        :default-expand-all="isExpanded"
        scrollbar-always-on
      >
        <template #icon="{ row }">
          <el-icon v-if="row.icon">
            <component :is="row.icon" />
          </el-icon>
        </template>
        <template #menuType="{ row }">
          <el-tag :type="getTypeTagType(row.menu_type)" size="small">
            {{ getTypeLabel(row.menu_type) }}
          </el-tag>
        </template>
        <template #visibility="{ row }">
          <el-tag :type="!row.is_visible ? 'warning' : 'success'" size="small">
            {{ !row.is_visible ? "隐藏" : "显示" }}
          </el-tag>
        </template>
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? "启用" : "禁用" }}
          </el-tag>
        </template>
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="edit"
              @click="handleEdit(row)" /><AdminRowAction
              action="delete"
              @click="handleDelete(row.id)" /></AdminRowActions
        ></template> </AdminDataTable
    ></template>
    <template #dialogs
      ><el-dialog
        v-model="dialogVisible"
        :title="dialogTitle"
        class="menu-editor-dialog"
        width="min(760px, calc(100vw - 32px))"
        top="6vh"
      >
        <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
          <el-form-item label="所属系统" prop="systemCode"
            ><el-select v-model="form.systemCode" @change="changeFormSystem"
              ><el-option
                v-for="system in systems"
                :key="system.code"
                :value="system.code"
                :label="system.name" /></el-select
          ></el-form-item>
          <el-form-item label="菜单类型" prop="type">
            <el-radio-group v-model="form.type">
              <el-radio :value="1">目录</el-radio>
              <el-radio :value="2">菜单</el-radio>
              <el-radio :value="3">按钮</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="上级菜单" prop="parent_id">
            <el-tree-select
              v-model="form.parent_id"
              :data="menuTreeSelectData"
              :props="{ label: 'menu_name', value: 'id' }"
              clearable
              placeholder="请选择上级菜单"
              check-strictly
            />
          </el-form-item>
          <el-form-item label="菜单名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入菜单名称" />
          </el-form-item>
          <el-form-item label="图标" prop="icon" v-if="form.type !== 3">
            <el-input v-model="form.icon" placeholder="请输入图标名称" />
          </el-form-item>
          <el-form-item label="路径" prop="path" v-if="form.type !== 3">
            <el-input v-model="form.path" placeholder="请输入路由地址" />
          </el-form-item>
          <el-form-item
            label="组件路径"
            prop="component"
            v-if="form.type === 2"
          >
            <el-input
              v-model="form.component"
              placeholder="例如 system/users/index（相对 views）"
            />
          </el-form-item>
          <el-form-item label="是否隐藏" prop="isVisible">
            <el-radio-group v-model="form.isVisible">
              <el-radio :value="true">否</el-radio>
              <el-radio :value="false">是</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="菜单编码" prop="code">
            <el-input
              v-model="form.code"
              placeholder="菜单唯一编码，例如 admin-system-users"
            />
          </el-form-item>
          <el-form-item
            label="权限标识"
            prop="permission"
            v-if="form.type !== 1"
            :required="form.type === 3"
          >
            <el-input
              v-model="form.permission"
              placeholder="例如 system:user:list；页面可留空"
            />
          </el-form-item>
          <el-form-item label="菜单描述" prop="description">
            <el-input
              v-model="form.description"
              placeholder="请输入描述信息"
              type="textarea"
              :rows="3"
            />
          </el-form-item>
          <el-form-item label="排序" prop="sort">
            <el-input-number v-model="form.sort" :min="0" :max="999" />
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
          <el-button type="primary" @click="handleSubmit" :loading="submitting"
            >确定</el-button
          >
        </template>
      </el-dialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import AdminRowAction from "@/components/admin/AdminRowAction.vue";
import AdminRowActions from "@/components/admin/AdminRowActions.vue";
import AdminListPage from "@/components/admin/AdminListPage.vue";
import AdminDataTable from "@/components/admin/AdminDataTable.vue";
import { useRbacStore } from "@/store/rbac";
import { useRouter } from "vue-router";
import {
  ref,
  reactive,
  computed,
  nextTick,
  onMounted,
  onBeforeUnmount,
  onActivated,
  onDeactivated,
} from "vue";
import { createLatestRequest } from "@/utils/latest-request";

const rbacStore = useRbacStore();
const router = useRouter();
async function refreshNavigation() {
  const menus = await rbacStore.refreshUserMenus();
  router.installMenus(menus);
  await router.replace(router.currentRoute.value.fullPath);
}

// On smaller screens keep the tree column scrollable so fixed columns do not overlap.
const narrowViewport = window.matchMedia("(max-width: 1100px)");
const isNarrow = ref(narrowViewport.matches);
const updateViewport = (event) => {
  isNarrow.value = event.matches;
};
narrowViewport.addEventListener("change", updateViewport);
onBeforeUnmount(() =>
  narrowViewport.removeEventListener("change", updateViewport),
);

const columns = computed(() => [
  {
    key: "menu_name",
    prop: "menu_name",
    label: "菜单名称",
    minWidth: "260",
    fixed: isNarrow.value ? false : "left",
    showOverflowTooltip: true,
    locked: true,
  },
  {
    key: "icon",
    label: "图标",
    width: "64",
    align: "center",
  },
  {
    key: "menuType",
    label: "类型",
    width: "84",
    align: "center",
  },
  {
    key: "path",
    prop: "path",
    label: "路由地址",
    minWidth: "280",
    showOverflowTooltip: true,
  },
  {
    key: "component",
    prop: "component",
    label: "组件地址",
    minWidth: "260",
    showOverflowTooltip: true,
  },
  {
    key: "menu_code",
    prop: "menu_code",
    label: "菜单编码",
    minWidth: "230",
    showOverflowTooltip: true,
  },
  {
    key: "permission",
    prop: "permission",
    label: "权限标识",
    minWidth: "230",
    showOverflowTooltip: true,
  },
  {
    key: "visibility",
    label: "是否隐藏",
    width: "96",
    align: "center",
  },
  {
    key: "remark",
    prop: "remark",
    label: "菜单描述",
    minWidth: "220",
    showOverflowTooltip: true,
  },
  {
    key: "sort",
    prop: "sort",
    label: "排序",
    width: "80",
    align: "center",
  },
  {
    key: "status",
    label: "状态",
    width: "88",
    align: "center",
  },
  {
    key: "actions",
    label: "操作",
    width: "132",
    fixed: "right",
    utility: true,
  },
]);

defineOptions({
  name: "MenuList",
});

import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import {
  createMenu,
  updateMenu,
  deleteMenu as deleteMenuApi,
  getAllMenus,
} from "@/api/menu";
import { getSystems } from "@/api/system";

const loading = ref(false);
const loadError = ref("");
const dialogVisible = ref(false);
const submitting = ref(false);
const formRef = ref(null);
const tableRef = ref(null);
const isExpanded = ref(true);
const allMenus = ref([]);
const systems = ref([]);
const parentMenus = ref([]);

const allMenuList = ref([]); // 存储所有顶级目录及其子菜单
// Page whole root branches so children always remain with their parent menu.
const pagination = reactive({ page: 1, pageSize: 10 });
const displayedMenuList = computed(() => {
  const start = (pagination.page - 1) * pagination.pageSize;
  return allMenuList.value.slice(start, start + pagination.pageSize);
});
const menuCount = computed(() => {
  const count = (nodes) =>
    nodes.reduce((total, node) => total + 1 + count(node.children || []), 0);
  return count(allMenuList.value);
});

const searchForm = reactive({
  systemCode: "admin",
  title: "",
  type: null,
});
const appliedFilters = reactive({ ...searchForm });

const dialogTitle = ref("新增菜单");
const form = reactive({
  systemCode: "admin",
  id: null,
  parent_id: 0,
  name: "",
  icon: "",
  type: "menu",
  path: "",
  component: "",
  code: "",
  permission: "",
  sort: 0,
  status: 1,
  isVisible: true,
  description: "",
});

const rules = {
  systemCode: [
    { required: true, message: "请选择所属系统", trigger: "change" },
  ],
  parent_id: [{ required: true, message: "请选择上级菜单", trigger: "change" }],
  name: [{ required: true, message: "请输入菜单名称", trigger: "blur" }],
  type: [{ required: true, message: "请选择菜单类型", trigger: "change" }],
  path: [{ required: true, message: "请输入路由地址", trigger: "blur" }],
  permission: [
    {
      validator: (_rule, value, callback) =>
        callback(
          form.type === 3 && !String(value || "").trim()
            ? new Error("请输入按钮权限标识")
            : undefined,
        ),
      trigger: "blur",
    },
  ],
};

// 构建菜单树选择数据（排除按钮，用于上级菜单选择器）
const menuTreeSelectData = computed(() => {
  const buildTree = (parentId) => {
    if (!Array.isArray(parentMenus.value)) {
      return [];
    }
    return parentMenus.value
      .filter((m) => {
        // 后端 menu_type 是数字: 1=目录 2=菜单 3=按钮
        const isNotButton = m.menu_type !== 3;
        const isEnabled = m.status === 1;
        const isChild = String(m.parent_id) === String(parentId);
        return isChild && isNotButton && isEnabled;
      })
      .sort((a, b) => a.sort - b.sort)
      .map((m) => {
        const children = buildTree(m.id);
        return {
          ...m,
          children,
        };
      });
  };
  return buildTree(0);
});

// 通用构建菜单树函数（用于表格展示）
const buildMenuTree = (parentId, list) => {
  if (!list || !Array.isArray(list)) {
    return [];
  }
  return list
    .filter((m) => String(m.parent_id) === String(parentId))
    .sort((a, b) => a.sort - b.sort)
    .map((m) => {
      const children = buildMenuTree(m.id, list);
      return {
        ...m,
        hasChildren: children.length > 0,
        children,
      };
    });
};

// 递归搜索树：检查节点及其子节点是否匹配搜索条件
const matchTree = (node, searchTitle, searchType) => {
  const matchTitle =
    !searchTitle || (node.menu_name && node.menu_name.includes(searchTitle));
  const matchType = !searchType || node.menu_type === searchType;
  const isMatch = matchTitle && matchType;

  if (node.children && node.children.length > 0) {
    // 递归过滤子节点
    const matchedChildren = node.children
      .map((child) => matchTree(child, searchTitle, searchType))
      .filter(Boolean);
    if (matchedChildren.length > 0) {
      return { ...node, children: matchedChildren, hasChildren: true };
    }
  }
  return isMatch
    ? {
        ...node,
        children: node.children || [],
        hasChildren: node.children && node.children.length > 0,
      }
    : null;
};

const getTypeLabel = (type) => {
  const map = {
    directory: "目录",
    menu: "菜单",
    button: "按钮",
    1: "目录",
    2: "菜单",
    3: "按钮",
  };
  return map[type] || type;
};

const getTypeTagType = (type) => {
  const map = {
    directory: "warning",
    menu: "primary",
    button: "success",
    1: "warning",
    2: "primary",
    3: "success",
  };
  return map[type] || "info";
};

const menuRequest = createLatestRequest();
const loadList = () =>
  menuRequest.run(
    (signal) =>
      getAllMenus({
        params: { system: appliedFilters.systemCode },
        signal,
        silentError: true,
      }),
    {
      start() {
        loading.value = true;
        loadError.value = "";
      },
      success(res) {
        allMenus.value = res.data || [];
        applyFilters();
      },
      failure() {
        loadError.value = "菜单列表加载失败，请刷新重试";
        allMenus.value = [];
        allMenuList.value = [];
      },
      finish() {
        loading.value = false;
      },
    },
  );

function applyFilters() {
  const tree = buildMenuTree(0, allMenus.value);
  allMenuList.value =
    appliedFilters.title || appliedFilters.type
      ? tree
          .map((node) =>
            matchTree(node, appliedFilters.title, appliedFilters.type),
          )
          .filter(Boolean)
      : tree;
  pagination.page = Math.min(
    pagination.page,
    Math.max(1, Math.ceil(allMenuList.value.length / pagination.pageSize)),
  );
}
const handleSearch = () => {
  Object.assign(appliedFilters, searchForm, { title: searchForm.title.trim() });
  pagination.page = 1;
  loadList();
};
const handleReset = () => {
  Object.assign(searchForm, { title: "", type: null, systemCode: "admin" });
  handleSearch();
};

const handleCreate = () => {
  dialogTitle.value = "新增菜单";
  Object.assign(form, {
    systemCode: appliedFilters.systemCode,
    id: null,
    parent_id: 0,
    name: "",
    icon: "",
    type: 2,
    path: "",
    component: "",
    code: "",
    permission: "",
    sort: 0,
    status: 1,
    isVisible: true,
    description: "",
  });
  parentMenus.value = allMenus.value;
  dialogVisible.value = true;
};

const handleEdit = (row) => {
  dialogTitle.value = "编辑菜单";
  Object.assign(form, {
    systemCode: row.systemCode || row.system_code || appliedFilters.systemCode,
    id: row.id,
    parent_id: row.parent_id,
    path: row.path,
    status: row.status,
    sort: row.sort,
    isVisible: row.is_visible,
    component: row.component,
    description: row.remark,
    type: row.menu_type,
    name: row.menu_name,
    code: row.menu_code,
    permission: row.permission || row.perms || "",
    icon: row.icon,
  });
  parentMenus.value = allMenus.value;
  dialogVisible.value = true;
};

async function changeFormSystem() {
  form.parent_id = 0;
  parentMenus.value = [];
  const system = form.systemCode;
  try {
    const response = await getAllMenus({ params: { system } });
    if (form.systemCode === system) parentMenus.value = response.data || [];
  } catch {
    /* Request layer reports the error; no stale parents are offered. */
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      // 构建提交数据，确保布尔值正确
      const submitData = {
        ...form,
        isVisible:
          form.isVisible === true || form.isVisible === "true" ? true : false,
      };

      if (form.id) {
        const data = await updateMenu(submitData);
        if (data.code == 0) {
          ElMessage.success("更新成功");
        }
      } else {
        const data = await createMenu(submitData);
        if (data.code == 0) {
          ElMessage.success("创建成功");
        }
      }
      dialogVisible.value = false;
      loadList();
      await refreshNavigation();
    } catch {
      // 请求层已展示具体错误，保留表单内容供修改后重试。
    } finally {
      submitting.value = false;
    }
  });
};

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm(
      "确定删除该菜单吗？如果有子菜单，子菜单也会被删除。",
      "提示",
      {
        type: "warning",
      },
    );

    // 调用删除API
    const data = await deleteMenuApi(id);

    if (data.code == 0) {
      ElMessage.success("删除成功");
      loadList();
      await refreshNavigation();
    }
  } catch {
    // 取消确认无需提示；接口失败由请求层统一提示。
  }
};

const handleExpandAll = () => {
  isExpanded.value = !isExpanded.value;

  // 使用公开的 toggleRowExpansion 方法递归展开/折叠当前页的菜单分支
  if (tableRef.value) {
    // 先全部折叠，再根据需要展开
    const rows = displayedMenuList.value;
    const toggleRows = (rowList) => {
      rowList.forEach((row) => {
        if (row.children && row.children.length > 0) {
          tableRef.value.toggleRowExpansion(row, isExpanded.value);
          toggleRows(row.children);
        }
      });
    };
    toggleRows(rows);
  }
};

const handlePageChange = async () => {
  await nextTick();
  tableRef.value?.$el
    ?.closest(".admin-table-panel")
    ?.scrollIntoView({ block: "start" });
};

onMounted(() => {
  getSystems()
    .then((response) => {
      systems.value = response.data.items || [];
    })
    .catch(() => {});
  loadList();
});
let activated = false;
onActivated(() => {
  if (activated) loadList();
  activated = true;
});
onDeactivated(() => {
  menuRequest.cancel();
  loading.value = false;
});
onBeforeUnmount(() => menuRequest.cancel());
</script>
