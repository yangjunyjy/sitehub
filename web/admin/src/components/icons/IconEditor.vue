<script setup>
import { ref, reactive, watch } from "vue";
import { readIcon, saveIcon, validateIcon } from "@/api/icons";
import AdminFormDialog from "@/components/admin/AdminFormDialog.vue";

const props = defineProps({ record: Object });
const open = defineModel({ type: Boolean });
const emit = defineEmits(["saved"]);
const form = reactive({ name: "", label: "", category: "", svg: "" });
const loading = ref(false),
  saving = ref(false),
  error = ref(""),
  preview = ref("");
let generation = 0;

watch(open, async (visible) => {
  const current = ++generation;
  error.value = preview.value = "";
  if (!visible) return;
  const record = props.record;
  Object.assign(form, {
    name: record?.name || "",
    label: record?.label || "",
    category: record?.category || "",
    svg: "",
  });
  loading.value = !!record;
  if (!record) return;
  try {
    const svg = await readIcon(record.name);
    if (current === generation) form.svg = svg;
  } catch (failure) {
    if (current === generation) error.value = failure.message;
  } finally {
    if (current === generation) loading.value = false;
  }
});
watch(
  () => form.svg,
  () => {
    preview.value = "";
  },
);

async function submit(validateOnly = false) {
  if (loading.value || saving.value) return;
  if (!form.name.trim() || !form.svg.trim()) {
    error.value = "请填写标识和 SVG 源码";
    return;
  }
  saving.value = true;
  error.value = "";
  try {
    if (validateOnly) {
      const { data } = await validateIcon({ ...form });
      preview.value =
        "data:image/svg+xml;charset=utf-8," + encodeURIComponent(data.svg);
    } else {
      const { data } = await saveIcon({ ...form }, props.record?.id);
      open.value = false;
      emit("saved", data);
    }
  } catch (failure) {
    error.value = failure.message;
  } finally {
    saving.value = false;
  }
}
</script>
<template>
  <AdminFormDialog
    v-model="open"
    :title="record ? '编辑图标' : '创建图标'"
    width="min(560px, 94vw)"
    :loading="saving || loading"
    @submit="submit()"
  >
    <el-form
      label-position="top"
      :disabled="saving || loading"
      @submit.prevent="submit()"
    >
      <el-form-item label="标识" required
        ><el-input
          v-model.trim="form.name"
          :disabled="!!record"
          placeholder="例如 site-home"
      /></el-form-item>
      <el-form-item label="名称"
        ><el-input v-model.trim="form.label" maxlength="80"
      /></el-form-item>
      <el-form-item label="分类"
        ><el-input v-model.trim="form.category" maxlength="40"
      /></el-form-item>
      <el-form-item label="SVG 源码" required
        ><el-input v-model="form.svg" type="textarea" :rows="6"
      /></el-form-item>
      <el-button v-if="!record" @click="submit(true)">校验并预览</el-button>
      <img
        v-if="preview"
        :src="preview"
        alt="图标预览"
        width="32"
        height="32"
      />
      <el-alert v-if="error" type="error" :title="error" :closable="false" />
    </el-form>
  </AdminFormDialog>
</template>
