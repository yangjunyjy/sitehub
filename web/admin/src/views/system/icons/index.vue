<script setup>
import { onMounted, ref } from "vue";
import { ElMessageBox } from "element-plus";
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import IconEditor from "@/components/icons/IconEditor.vue";
import AdminListPage from "@/components/admin/AdminListPage.vue";
import AdminDataTable from "@/components/admin/AdminDataTable.vue";
import { listIcons, removeIcon } from "@/api/icons";

const rows = ref([]),
  error = ref(""),
  loading = ref(false),
  open = ref(false),
  selected = ref(null);
import { columns } from "./columns";
let generation = 0;
async function load() {
  const current = ++generation;
  loading.value = true;
  error.value = "";
  try {
    const icons = await listIcons();
    if (current === generation) rows.value = icons;
  } catch (failure) {
    if (current === generation) error.value = failure.message;
  } finally {
    if (current === generation) loading.value = false;
  }
}
function edit(icon) {
  selected.value = icon;
  open.value = true;
}
async function archive(icon) {
  try {
    await ElMessageBox.confirm(
      "停用后，已发布内容仍可显示此图标。",
      "停用图标",
    );
  } catch {
    return;
  }
  try {
    await removeIcon(icon.id);
    await load();
  } catch (failure) {
    error.value = failure.message;
  }
}
onMounted(load);
</script>
<template>
  <AdminListPage
    title="图标管理"
    :columns="columns"
    preference-key="system-icons"
    :pageable="false"
    :loading="loading"
    :error="error"
    @refresh="load"
  >
    <template #actions
      ><el-button
        v-permission="'system:icon:create'"
        type="primary"
        @click="edit(null)"
        >创建图标</el-button
      ></template
    >
    <template #default="{ tableProps, tableHeight }">
      <AdminDataTable
        v-bind="tableProps"
        :data="rows"
        :max-height="tableHeight"
      >
        <template #icon="{ row }"
          ><ResourceIcon :key="generation" :name="row.name"
        /></template>
        <template #actions="{ row }">
          <el-button
            v-permission="'system:icon:update'"
            link
            type="primary"
            @click="edit(row)"
            >编辑</el-button
          >
          <el-button
            v-permission="'system:icon:delete'"
            link
            type="danger"
            @click="archive(row)"
            >停用</el-button
          >
        </template>
      </AdminDataTable>
    </template>
    <template #dialogs
      ><IconEditor v-model="open" :record="selected" @saved="load"
    /></template>
  </AdminListPage>
</template>
