<script setup>
import { computed, ref } from "vue";
import { builtInIcons } from "@shared/icons/catalog";
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import { listIcons } from "@/api/icons";
import { useRbacStore } from "@/store/rbac";
import IconEditor from "./IconEditor.vue";

const value = defineModel({ type: String });
const rbac = useRbacStore();
const custom = ref([]),
  loading = ref(false),
  error = ref(""),
  editing = ref(false);
const options = computed(() => [...custom.value, ...builtInIcons]);
async function refresh(open) {
  if (!open || loading.value || !rbac.hasPermission("system:icon:list")) return;
  loading.value = true;
  error.value = "";
  try {
    custom.value = await listIcons();
  } catch (failure) {
    error.value = failure.message;
  } finally {
    loading.value = false;
  }
}
function selectCreated(icon) {
  custom.value = [
    icon,
    ...custom.value.filter((item) => item.name !== icon.name),
  ];
  value.value = icon.name;
}
</script>
<template>
  <div class="icon-picker">
    <ResourceIcon v-if="value" :name="value" />
    <el-select
      v-model="value"
      clearable
      filterable
      :loading="loading"
      placeholder="选择图标"
      @visible-change="refresh"
    >
      <el-option
        v-for="icon in options"
        :key="icon.name"
        :value="icon.name"
        :label="icon.label + ' · ' + icon.name"
      >
        <ResourceIcon :name="icon.name" /> {{ icon.label }}
        <small>{{ icon.category }}</small>
      </el-option>
    </el-select>
    <el-button
      v-if="rbac.hasPermission('system:icon:create')"
      @click="editing = true"
      >创建</el-button
    >
    <span v-if="error" role="alert">{{ error }}</span>
    <IconEditor v-model="editing" @saved="selectCreated" />
  </div>
</template>
<style scoped lang="scss">
.icon-picker {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;
  flex-wrap: wrap;
}
.icon-picker .el-select {
  flex: 1;
  min-width: 8em;
}
.icon-picker [role="alert"] {
  width: 100%;
  color: var(--el-color-danger);
}
</style>
