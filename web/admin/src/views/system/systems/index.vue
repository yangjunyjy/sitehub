<template>
  <AdminListPage
    title="系统管理"
    preference-key="systems"
    :columns="columns"
    :loading="loading"
    :error="error"
    :filter-model="filters"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    :total="filtered.length"
    @refresh="load"
    @search="search"
    @reset="reset"
  >
    <template #filters
      ><el-form-item label="系统名称"
        ><el-input
          v-model="filters.keyword"
          placeholder="名称或代码"
          clearable /></el-form-item
    ></template>
    <template #actions
      ><el-button
        type="primary"
        :icon="Plus"
        v-permission="'system:system:create'"
        @click="edit()"
        >新增系统</el-button
      ></template
    >
    <template #default="{ tableHeight, tableProps }">
      <AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        :data="rows"
      >
        <template #enabled="{ row }"
          ><el-tag :type="row.enabled ? 'success' : 'info'">{{
            row.enabled ? "启用" : "停用"
          }}</el-tag></template
        >
        <template #public="{ row }"
          ><el-tag :type="row.public ? 'success' : 'info'">{{
            row.public ? "公开" : "内部"
          }}</el-tag></template
        >
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="edit"
              v-permission="'system:system:update'"
              @click="edit(row)" /><AdminRowAction
              action="delete"
              v-permission="'system:system:delete'"
              @click="remove(row)" /></AdminRowActions
        ></template>
      </AdminDataTable>
    </template>
    <template #dialogs>
      <el-dialog
        v-model="visible"
        :title="form.id ? '编辑系统' : '新增系统'"
        width="min(580px, calc(100vw - 32px))"
        :close-on-click-modal="!saving"
      >
        <el-form label-position="top" :model="form" @submit.prevent="submit">
          <el-form-item label="系统代码" required
            ><el-input
              v-model="form.code"
              :disabled="!!form.id"
              placeholder="例如 blog"
              maxlength="32"
          /></el-form-item>
          <el-form-item label="系统名称" required
            ><el-input v-model="form.name" maxlength="80"
          /></el-form-item>
          <el-form-item label="访问地址"
            ><el-input
              v-model="form.origin"
              placeholder="https://blog.example.com"
          /></el-form-item>
          <el-form-item label="系统描述"
            ><el-input
              v-model="form.description"
              type="textarea"
              maxlength="500"
          /></el-form-item>
          <div class="system-flags">
            <el-form-item label="启用"
              ><el-switch v-model="form.enabled" /></el-form-item
            ><el-form-item label="公开系统"
              ><el-switch v-model="form.public"
            /></el-form-item>
          </div>
        </el-form>
        <template #footer
          ><el-button :disabled="saving" @click="visible = false"
            >取消</el-button
          ><el-button type="primary" :loading="saving" @click="submit"
            >保存</el-button
          ></template
        >
      </el-dialog>
    </template>
  </AdminListPage>
</template>
<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import AdminListPage from "@/components/admin/AdminListPage.vue";
import AdminDataTable from "@/components/admin/AdminDataTable.vue";
import AdminRowActions from "@/components/admin/AdminRowActions.vue";
import AdminRowAction from "@/components/admin/AdminRowAction.vue";
import { getSystems, saveSystem, deleteSystem } from "@/api/system";
import { systemPayload } from "@/utils/systems";
import { useAdminResource } from "@/composables/useAdminResource";
defineOptions({ name: "SystemList" });
const { data, loading, error, load } = useAdminResource(getSystems);
const filters = reactive({ keyword: "" }),
  applied = ref(""),
  page = ref(1),
  pageSize = ref(10);
const filtered = computed(() =>
  (data.value?.items || []).filter((item) =>
    `${item.code} ${item.name}`.toLocaleLowerCase().includes(applied.value),
  ),
);
const rows = computed(() =>
  filtered.value.slice(
    (page.value - 1) * pageSize.value,
    page.value * pageSize.value,
  ),
);
const columns = [
  { key: "code", prop: "code", label: "系统代码", minWidth: 130 },
  { key: "name", prop: "name", label: "系统名称", minWidth: 150 },
  { key: "origin", prop: "origin", label: "访问地址", minWidth: 240 },
  { key: "description", prop: "description", label: "描述", minWidth: 200 },
  { key: "enabled", label: "状态", width: 90 },
  { key: "public", label: "可见性", width: 90 },
  { key: "actions", label: "操作", width: 132, fixed: "right", utility: true },
];
const visible = ref(false),
  saving = ref(false),
  originalCode = ref("");
const form = reactive({});
function search() {
  applied.value = filters.keyword.trim().toLocaleLowerCase();
  page.value = 1;
}
function reset() {
  filters.keyword = "";
  search();
}
function edit(row) {
  Object.assign(
    form,
    {
      id: "",
      code: "",
      name: "",
      description: "",
      origin: "",
      enabled: true,
      public: false,
    },
    row || {},
  );
  originalCode.value = row?.code || "";
  visible.value = true;
}
async function submit() {
  if (saving.value) return;
  let payload;
  try {
    payload = systemPayload(form, originalCode.value);
  } catch (e) {
    ElMessage.error(e.message);
    return;
  }
  saving.value = true;
  try {
    await saveSystem(payload);
    visible.value = false;
    ElMessage.success("系统已保存");
    await load();
  } catch {
    /* Request layer reports failures. */
  } finally {
    saving.value = false;
  }
}
async function remove(row) {
  try {
    await ElMessageBox.confirm(`确定删除系统“${row.name}”吗？`, "删除系统", {
      type: "warning",
    });
  } catch {
    return;
  }
  try {
    await deleteSystem(row.id);
    ElMessage.success("系统已删除");
    await load();
  } catch {
    /* Keep the list when deletion fails. */
  }
}
onMounted(load);
</script>
<style lang="scss" scoped>
.system-flags {
  display: flex;
  gap: 40px;
}
</style>
