<template>
  <AdminListPage
    :columns="columns"
    preference-key="downloads"
    class="download-page"
    title="下载管理"
    :filter-model="filters"
    :error="loadError"
    :loading="loading || saving"
    v-model:current-page="page"
    v-model:page-size="pageSize"
    :total="filteredPackages.length"
    :page-sizes="[10, 20, 50]"
    @search="search"
    @reset="resetFilters"
    @refresh="load"
  >
    <template #tabs>
      <div class="release-switcher" role="group" aria-label="按上架状态筛选">
        <button
          v-for="item in statusOptions"
          :key="item.status"
          type="button"
          :aria-pressed="appliedFilters.status === item.status"
          @click="selectStatus(item.status)"
        >
          {{ item.label }} <span>{{ ready ? item.count : '—' }}</span>
        </button>
      </div>
    </template>
    <template #filters
      ><el-form-item label="搜索文件"
        ><el-input
          v-model="filters.keyword"
          placeholder="文件名或版本"
          clearable /></el-form-item
      ><el-form-item label="所属产品"
        ><el-select v-model="filters.product" placeholder="全部产品" clearable
          ><el-option
            v-for="item in products"
            :key="item"
            :label="item"
            :value="item" /></el-select></el-form-item
      ><el-form-item label="运行平台"
        ><el-select v-model="filters.platform" placeholder="全部平台" clearable
          ><el-option
            v-for="item in platforms"
            :key="item"
            :label="item"
            :value="item" /></el-select></el-form-item></template
    ><template #advanced
      ><el-form-item label="下载分组"
        ><el-select v-model="filters.groupId" placeholder="全部分组" clearable
          ><el-option
            v-for="group in downloads.groups"
            :key="group.id"
            :label="group.name"
            :value="group.id" /></el-select></el-form-item
      ><el-form-item label="版本类型"
        ><el-select v-model="filters.edition" placeholder="全部类型" clearable
          ><el-option
            v-for="edition in editions"
            :key="edition"
            :label="edition"
            :value="edition" /></el-select></el-form-item
      ></template>
    <template #actions
      ><el-button
        type="primary"
        plain
        :icon="Plus"
        :disabled="!ready || loading || saving"
        @click="openEditor()"
        >新增下载文件</el-button
      >
      <el-button :disabled="!ready" @click="groupsOpen = true"
        >管理分组</el-button
      >
      <el-button :icon="View" @click="preview">下载页预览</el-button
      ><el-button :icon="Setting" :disabled="!ready" @click="openSettings"
        >下载页设置</el-button
      ></template
    >
    <template #default="{ tableHeight, tableProps }"
      ><AdminDataTable
        v-bind="tableProps"
        :max-height="tableHeight"
        :data="pagedPackages"
        row-key="id"
        class="release-table"
        empty-text="暂无匹配的下载文件"
      >
        <template #product="{ row }"
          ><div class="product-cell">
            <span class="product-icon"
              ><ResourceIcon :name="row.icon || 'PKG'"
            /></span>
            <div>
              <b>{{ row.product }}</b
              ><small v-if="row.fileName || row.description" :title="row.fileName || row.description">{{
                row.fileName || row.description
              }}</small>
            </div>
          </div></template
        >
        <template #groupId="{ row }"
          ><span>{{
            downloads.groups.find((group) => group.id === row.groupId)?.name ||
            "未分组"
          }}</span
          ><small class="cell-secondary">{{
            row.edition || "通用版本"
          }}</small></template
        >
        <template #version="{ row }"
          ><span class="version">{{ row.version }}</span></template
        >
        <template #platform="{ row }"
          >{{ row.platform
          }}<small class="cell-secondary">{{
            row.arch
          }}</small></template
        >
        <template #size="{ row }">{{ formatFileSize(row.size) }}</template>
        <template #releaseDate="{ row }">{{ row.releaseDate || "—" }}</template>
        <template #status="{ row }"
          ><span class="release-status" :class="row.status"
            ><i />{{ releaseStates[row.status] }}</span
          ></template
        >
        <template #actions="{ row }"
          ><AdminRowActions
            ><AdminRowAction
              action="edit"
              :disabled="!ready"
              @click="openEditor(row)"
            /><AdminRowAction
              :action="row.status === 'published' ? 'unpublish' : 'publish'"
              :disabled="!ready"
              @click="togglePublish(row)"
            /><el-dropdown
              trigger="click"
              :popper-options="{
                modifiers: [
                  {
                    name: 'preventOverflow',
                    options: { padding: 8, altAxis: true, tether: false },
                  },
                ],
              }"
              :disabled="!ready"
              @command="(command) => handleMore(command, row)"
              ><AdminRowAction
                action="more"
                label="更多文件操作"
                :disabled="!ready"
              /><template #dropdown
                ><el-dropdown-menu
                  ><el-dropdown-item command="copy">复制为草稿</el-dropdown-item
                  ><el-dropdown-item command="delete" divided
                    >删除文件记录</el-dropdown-item
                  ></el-dropdown-menu
                ></template
              ></el-dropdown
            ></AdminRowActions
          ></template
        >
      </AdminDataTable></template
    >
    <template #dialogs
      ><ReleaseFileEditor
        v-if="editorOpen"
        :record="editingRow"
        :existing="editingExisting"
        :defaults="settings"
        :groups="downloads.groups"
        :edition-options="editionOptions"
        @close="editorOpen = false"
        @saved="load"
      />

      <ReleaseGroups
        v-if="groupsOpen"
        :groups="downloads.groups"
        @close="groupsOpen = false"
        @change="load"
      />
      <el-drawer
        v-model="settingsOpen"
        title="下载页设置"
        :size="'min(760px, 100vw)'"
        :close-on-click-modal="false"
        :before-close="closeSettings"
      >
        <DownloadSettings
          v-if="settingsDraft"
          v-model="settingsDraft"
        /><template #footer
          ><el-button :disabled="saving" @click="cancelSettings">取消</el-button
          ><el-button type="primary" :loading="saving" @click="saveSettings"
            >保存设置</el-button
          ></template
        >
      </el-drawer>
    </template>
  </AdminListPage>
</template>
<script setup>
import ReleaseFileEditor from "./components/ReleaseFileEditor.vue";
import { confirmDiscard } from "./confirmDiscard";
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import AdminRowAction from "@/components/admin/AdminRowAction.vue";
import AdminRowActions from "@/components/admin/AdminRowActions.vue";
import AdminListPage from "@/components/admin/AdminListPage.vue";
import AdminDataTable from "@/components/admin/AdminDataTable.vue";
import { columns } from "./columns";

import { computed, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getReleaseFiles,
  listReleaseGroups,
  getReleaseSettings,
  saveReleaseSettings,
  deleteReleaseFile,
  publishReleaseFile,
  withdrawReleaseFile,
} from "@/api/downloads";
import ReleaseGroups from "./components/ReleaseGroups.vue";
import {
  View,
  Plus,
  Setting,
} from "@element-plus/icons-vue";
import {
  releaseStates,
  filterPackages,
  formatFileSize,
} from "@/utils/downloads";
import { previewProductPage } from "@/utils/product-preview";
import DownloadSettings from "./components/DownloadSettings.vue";
import { createUUID } from "@/utils/uuid";
import { useUnsavedChanges } from "@/composables/useAdminResource";

const loading = ref(false),
  saving = ref(false),
  ready = ref(false),
  loadError = ref("");
const downloads = ref({ packages: [], groups: [] });
const page = ref(1),
  pageSize = ref(10);
const filters = reactive({
  product: "",
  groupId: "",
  edition: "",
  platform: "",
  status: "",
  keyword: "",
});
const appliedFilters = reactive({ ...filters });
const packages = computed(() => downloads.value.packages);
const products = computed(() => [
  ...new Set(packages.value.map((item) => item.product)),
]);
const editions = computed(() => [
  ...new Set(packages.value.map((item) => item.edition).filter(Boolean)),
]);
const editionOptions = computed(() => [
  ...new Set(["单节点 API 版", "分布式版", "命令行版", ...editions.value]),
]);
const platforms = computed(() => [
  ...new Set(packages.value.map((item) => item.platform)),
]);
const statusOptions = computed(() => {
  const counts = { published: 0, draft: 0, archived: 0 };
  for (const file of packages.value) counts[file.status]++;
  return [
    { status: "", label: "全部", count: packages.value.length },
    ...Object.entries(releaseStates).map(([status, label]) => ({ status, label, count: counts[status] })),
  ];
});
function selectStatus(status) {
  filters.status = status;
  appliedFilters.status = status;
  page.value = 1;
}
const filteredPackages = computed(() =>
  filterPackages(packages.value, appliedFilters),
);
const pagedPackages = computed(() =>
  filteredPackages.value.slice(
    (page.value - 1) * pageSize.value,
    page.value * pageSize.value,
  ),
);
watch([() => filteredPackages.value.length, pageSize], () => {
  page.value = Math.min(
    page.value,
    Math.max(1, Math.ceil(filteredPackages.value.length / pageSize.value)),
  );
});
const search = () => {
  Object.assign(appliedFilters, filters);
  page.value = 1;
};
const resetFilters = () => {
  Object.assign(filters, {
    product: "",
    groupId: "",
    edition: "",
    platform: "",
    status: "",
    keyword: "",
  });
  search();
};
const groupsOpen = ref(false),
  settings = ref(null);
let loadGeneration = 0;
const load = async () => {
  const generation = ++loadGeneration;
  loading.value = true;
  loadError.value = "";
  try {
    const [files, groups, pageSettings] = await Promise.all([
      getReleaseFiles(),
      listReleaseGroups(),
      getReleaseSettings(),
    ]);
    if (generation === loadGeneration) {
      downloads.value = { packages: files.data.items, groups };
      settings.value = pageSettings;
      ready.value = true;
    }
  } catch (error) {
    if (generation === loadGeneration) {
      loadError.value = error.message;
      ready.value = false;
    }
  } finally {
    if (generation === loadGeneration) loading.value = false;
  }
};
const editorOpen = ref(false),
  editingRow = ref(null),
  editingExisting = ref(false),
  settingsBaseline = ref("");
const openEditor = (row, existing = !!row) => {
  editingRow.value = row;
  editingExisting.value = existing;
  editorOpen.value = true;
};
const closeSettings = async (done) => {
  if (
    !saving.value &&
    (await confirmDiscard(
      JSON.stringify(settingsDraft.value) !== settingsBaseline.value,
    ))
  )
    done();
};
const cancelSettings = () =>
  closeSettings(() => {
    settingsOpen.value = false;
  });
const togglePublish = async (row) => {
  if (saving.value) return;
  const active = row.status === "published";
  try {
    await ElMessageBox.confirm(
      active ? "下架后该文件将无法下载。" : "上架后会员可下载此文件。",
      active ? "下架文件" : "上架文件",
      { type: "warning" },
    );
    saving.value = true;
    if (active) await withdrawReleaseFile(row);
    else await publishReleaseFile(row);
    ElMessage.success(active ? "文件已下架" : "文件已上架");
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error.message || "操作失败");
  } finally {
    saving.value = false;
  }
};
const handleMore = async (command, row) => {
  if (command === "copy") {
    openEditor(
      {
        ...row,
        id: "package-" + createUUID(),
        revision: "0",
        status: "draft",
      },
      false,
    );
    return;
  }
  try {
    await ElMessageBox.confirm(
      "删除该文件的下载记录？已上传的实体文件将保留。",
      "删除文件",
      { type: "warning" },
    );
    saving.value = true;
    await deleteReleaseFile(row);
    ElMessage.success("下载记录已删除");
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error.message || "删除失败");
  } finally {
    saving.value = false;
  }
};
const settingsOpen = ref(false),
  settingsDraft = ref(null);
const openSettings = () => {
  settingsDraft.value = JSON.parse(JSON.stringify(settings.value));
  settingsBaseline.value = JSON.stringify(settingsDraft.value);
  settingsOpen.value = true;
};
const saveSettings = async () => {
  if (saving.value) return;
  saving.value = true;
  try {
    settings.value = await saveReleaseSettings(settingsDraft.value);
    settingsOpen.value = false;
    ElMessage.success("下载页设置已保存");
  } catch (error) {
    ElMessage.error(error.message || "保存失败");
  } finally {
    saving.value = false;
  }
};
useUnsavedChanges(
  () =>
    settingsOpen.value &&
    JSON.stringify(settingsDraft.value) !== settingsBaseline.value,
);
const preview = () => previewProductPage("download");
load();
</script>

<style lang="scss" scoped src="./downloads.scss"></style>
