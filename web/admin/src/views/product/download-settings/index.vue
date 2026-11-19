<template>
  <AdminPage title="下载设置" :loading="loading" :error="error">
    <template #actions>
      <el-button @click="reload">重新加载</el-button>
      <el-button
        type="primary"
        :loading="saving"
        :disabled="!data || !!error"
        @click="submit"
        >保存设置</el-button
      >
    </template>
    <DownloadSettings v-if="data" v-model="form" />
  </AdminPage>
</template>
<script setup>
import { ref } from "vue";
import AdminPage from "@/components/AdminPage.vue";
import DownloadSettings from "../downloads/components/DownloadSettings.vue";
import { getWebsiteDownloads, updateWebsiteDownloads } from "@/api/website";
import {
  useAdminResource,
  useUnsavedChanges,
} from "@/composables/useAdminResource";
const { data, loading, saving, error, load, save } =
  useAdminResource(getWebsiteDownloads);
const form = ref({
  groups: [],
  packages: [],
  installOptions: [],
  requirements: [],
});
const baseline = ref("");
const confirmLeave = useUnsavedChanges(
  () => !!baseline.value && JSON.stringify(form.value) !== baseline.value,
);
async function initialize() {
  await load();
  if (!error.value && data.value) {
    form.value = {
      groups: [],
      packages: [],
      installOptions: [],
      requirements: [],
      ...structuredClone(data.value.downloads),
    };
    baseline.value = JSON.stringify(form.value);
  }
}
async function reload() {
  if (await confirmLeave()) await initialize();
}
async function submit() {
  await save(async () => {
    // Preserve the current file catalog when this page only edits presentation.
    const latest = (await getWebsiteDownloads()).data.downloads;
    await updateWebsiteDownloads({
      ...form.value,
      packages: latest.packages || [],
    });
    baseline.value = JSON.stringify(form.value);
  });
}
initialize();
</script>
