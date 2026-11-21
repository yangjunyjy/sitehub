<script setup>
import { computed, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { createUUID } from "@/utils/uuid";
import {
  createReleaseGroup,
  updateReleaseGroup,
  deleteReleaseGroup,
} from "@/api/downloads";
import { useUnsavedChanges } from "@/composables/useAdminResource";
const props = defineProps({ groups: { type: Array, required: true } });
const emit = defineEmits(["close", "change"]);
const draft = ref(null),
  original = ref(""),
  busy = ref(false),
  error = ref("");
const dirty = computed(
  () => !!draft.value && JSON.stringify(draft.value) !== original.value,
);
useUnsavedChanges(() => dirty.value);
function edit(group) {
  draft.value = group
    ? { ...group }
    : {
        id: "group-" + createUUID(),
        name: "",
        description: "",
        sortOrder: 0,
        revision: "0",
      };
  original.value = JSON.stringify(draft.value);
  error.value = "";
}
async function close(done) {
  if (busy.value) return;
  if (dirty.value) {
    try {
      await ElMessageBox.confirm("放弃本次分组修改？", "关闭编辑", {
        type: "warning",
      });
    } catch {
      return;
    }
  }
  if (draft.value) {
    draft.value = null;
    error.value = "";
    return;
  }
  done();
}
async function save() {
  if (busy.value || !draft.value.name.trim()) return;
  busy.value = true;
  error.value = "";
  try {
    const action =
      String(draft.value.revision) === "0"
        ? createReleaseGroup
        : updateReleaseGroup;
    await action({ ...draft.value, name: draft.value.name.trim() });
    draft.value = null;
    emit("change");
    ElMessage.success("分组已保存");
  } catch (failure) {
    error.value = failure.message;
  } finally {
    busy.value = false;
  }
}
async function remove(row) {
  try {
    await ElMessageBox.confirm("删除分组「" + row.name + "」？", "删除分组", {
      type: "warning",
    });
  } catch {
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    await deleteReleaseGroup(row);
    emit("change");
    ElMessage.success("分组已删除");
  } catch (failure) {
    error.value = failure.message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <el-drawer
    :model-value="true"
    :title="
      draft
        ? String(draft.revision) === '0'
          ? '新增分组'
          : '编辑分组'
        : '下载分组'
    "
    size="min(560px,100vw)"
    :close-on-click-modal="false"
    :before-close="close"
    @closed="emit('close')"
  >
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      :closable="false"
      show-icon
    />
    <el-form
      v-if="draft"
      label-position="top"
      :disabled="busy"
      @submit.prevent="save"
    >
      <el-form-item label="分组名称" required
        ><el-input v-model="draft.name" maxlength="80"
      /></el-form-item>
      <el-form-item label="介绍"
        ><el-input v-model="draft.description" type="textarea" maxlength="2000"
      /></el-form-item>
      <el-form-item label="排序"
        ><el-input-number
          v-model="draft.sortOrder"
          :min="0"
          :max="99999"
          :precision="0"
      /></el-form-item>
      <el-button @click="close(() => {})">返回</el-button
      ><el-button
        type="primary"
        native-type="submit"
        :loading="busy"
        :disabled="!draft.name.trim()"
        >保存分组</el-button
      >
    </el-form>
    <template v-else>
      <el-button type="primary" :disabled="busy" @click="edit()"
        >新增分组</el-button
      >
      <el-table :data="props.groups" empty-text="尚无下载分组" row-key="id">
        <el-table-column
          prop="name"
          label="名称"
          min-width="100"
        /><el-table-column prop="sortOrder" label="排序" width="65" />
        <el-table-column label="操作" width="110"
          ><template #default="{ row }"
            ><el-button link :disabled="busy" @click="edit(row)">编辑</el-button
            ><el-button link type="danger" :disabled="busy" @click="remove(row)"
              >删除</el-button
            ></template
          ></el-table-column
        >
      </el-table>
    </template>
  </el-drawer>
</template>
<style scoped lang="scss">
.el-table,
.el-form,
.el-alert {
  margin-top: 16px;
}
.el-alert {
  margin-bottom: 12px;
}
</style>
