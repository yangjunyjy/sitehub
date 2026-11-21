<script setup>
import { computed, ref } from "vue";
import { ElMessage } from "element-plus";
import { UploadFilled } from "@element-plus/icons-vue";
import {
  normalizePackage,
  packageError,
  isLocalReleaseURL,
  normalizeDownloadURL,
  formatFileSize,
} from "@/utils/downloads";
import {
  createReleaseFile,
  updateReleaseFile,
  uploadDownloadPackage,
} from "@/api/downloads";
import { createUUID } from "@/utils/uuid";
import { useUnsavedChanges } from "@/composables/useAdminResource";
import { confirmDiscard } from "../confirmDiscard";
import PackageMetadataFields from "./PackageMetadataFields.vue";
const props = defineProps({
  record: Object,
  existing: Boolean,
  defaults: Object,
  groups: Array,
  editionOptions: Array,
});
const emit = defineEmits(["close", "saved"]);
const visible = ref(true),
  saving = ref(false),
  uploading = ref(false),
  uploadProgress = ref(0),
  editorError = ref("");
const today = () => {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
};
const draft = ref(
  normalizePackage(
    props.record || {
      id: "package-" + createUUID(),
      releaseDate: props.defaults?.releaseDate || today(),
    },
    { version: props.defaults?.version },
  ),
);
if (!/^\d{4}-\d{2}-\d{2}$/.test(draft.value.releaseDate))
  draft.value.releaseDate = "";
const editingLocation = ref(!isLocalReleaseURL(draft.value.url));
const localUpload = computed(() => isLocalReleaseURL(draft.value.url));
const editorBaseline = ref(JSON.stringify(draft.value));
const clearFileMetadata = () => {
  draft.value.size = 0;
  draft.value.checksum = "";
};
useUnsavedChanges(
  () => visible.value && JSON.stringify(draft.value) !== editorBaseline.value,
);
const closeEditor = async (done) => {
  if (
    !uploading.value &&
    !saving.value &&
    (await confirmDiscard(JSON.stringify(draft.value) !== editorBaseline.value))
  )
    done();
};
const cancelEditor = () =>
  closeEditor(() => {
    visible.value = false;
  });
const uploadFile = async ({ file, onSuccess, onError }) => {
  uploading.value = true;
  uploadProgress.value = 0;
  editorError.value = "";
  try {
    const result = await uploadDownloadPackage(file, (percent) => {
      uploadProgress.value = percent;
    });
    draft.value.url = normalizeDownloadURL(result.data.url);
    editingLocation.value = false;
    draft.value.fileName = file.name;
    draft.value.size = file.size;
    draft.value.format = file.name.toLowerCase().endsWith(".tar.gz")
      ? "TAR.GZ"
      : file.name.split(".").pop().toUpperCase();
    draft.value.checksum = "";
    onSuccess(result);
  } catch (error) {
    editorError.value = error.message || "安装包上传失败";
    onError(error);
  } finally {
    uploading.value = false;
  }
};
const saveDraft = async () => {
  if (saving.value || uploading.value) return;
  const normalized = normalizePackage(draft.value);
  editorError.value = packageError(normalized);
  if (editorError.value) return;
  const { status, ...input } = normalized;
  input.revision = props.record?.revision || "0";
  saving.value = true;
  try {
    if (props.existing) await updateReleaseFile(input);
    else await createReleaseFile(input);
    visible.value = false;
    ElMessage.success("已保存下载文件");
    emit("saved");
  } catch (error) {
    editorError.value = error.message;
  } finally {
    saving.value = false;
  }
};
</script>
<template>
  <el-dialog
    v-model="visible"
    @closed="$emit('close')"
    :title="existing ? '编辑下载文件' : '新增下载文件'"
    width="min(740px, calc(100vw - 32px))"
    class="release-dialog"
    destroy-on-close
    :close-on-click-modal="false"
    :before-close="closeEditor"
  >
    <el-alert
      v-if="editorError"
      :title="editorError"
      type="error"
      :closable="false"
      show-icon
      class="load-error"
    />
    <el-form label-position="top" :disabled="saving || uploading">
      <PackageMetadataFields
        v-model="draft"
        :groups="groups"
        :edition-options="editionOptions"
      />
      <el-form-item label="上传文件"
        ><el-upload
          drag
          :show-file-list="false"
          :http-request="uploadFile"
          :disabled="uploading || saving"
          class="package-upload"
          ><el-icon class="upload-symbol"><UploadFilled /></el-icon
          ><b>{{
            uploading
              ? `正在上传 ${uploadProgress}%`
              : draft.fileName || "点击或拖拽文件到此处"
          }}</b
          ><small>{{
            draft.fileName
              ? `${formatFileSize(draft.size)} · 点击可替换文件`
              : "支持当前存储配置允许的格式"
          }}</small></el-upload
        ><el-progress
          v-if="uploading"
          :percentage="uploadProgress"
          :stroke-width="4"
      /></el-form-item>
      <el-form-item label="下载文件名">
        <el-input v-model="draft.fileName" placeholder="例如 deploy_arm.tgz" />
        <p class="field-note">
          用户下载后保存的名称，修改名称不会改变已上传文件的存储路径。请保留正确的扩展名。
        </p>
      </el-form-item>
      <el-form-item label="下载地址">
        <el-input
          v-model="draft.url"
          :readonly="localUpload && !editingLocation"
          placeholder="上传后自动填入，或填写已有文件的外部下载链接"
          @input="clearFileMetadata"
        />
        <div class="location-help">
          <p class="field-note">
            {{
              localUpload
                ? "上传文件的地址由系统生成。替换安装包请重新上传；只改下载名称请填写上方文件名。"
                : "地址必须指向已经存在的文件。填写新地址不会移动或重命名原文件。"
            }}
          </p>
          <el-button
            v-if="localUpload && !editingLocation"
            link
            type="primary"
            @click="editingLocation = true"
            >更换文件地址</el-button
          >
        </div>
      </el-form-item>
      <div class="field-grid">
        <el-form-item label="包格式"
          ><el-input
            v-model="draft.format"
            placeholder="ZIP / DMG / TAR.GZ" /></el-form-item
        ><el-form-item label="发布日期"
          ><el-date-picker
            v-model="draft.releaseDate"
            value-format="YYYY-MM-DD"
            type="date"
            placeholder="选择日期"
        /></el-form-item>
      </div>
      <el-form-item label="版本说明"
        ><el-input
          v-model="draft.description"
          type="textarea"
          :rows="2"
          placeholder="简要描述更新内容或适用场景"
      /></el-form-item>
      <el-form-item label="SHA256（可选）"
        ><el-input v-model="draft.checksum" placeholder="64 位校验值"
      /></el-form-item>
    </el-form>
    <template #footer
      ><el-button :disabled="uploading || saving" @click="cancelEditor"
        >取消</el-button
      ><el-button
        type="primary"
        :disabled="uploading"
        :loading="saving"
        @click="saveDraft"
        >保存文件</el-button
      ></template
    >
  </el-dialog>
</template>
<style scoped lang="scss">
.field-note {
  margin: 7px 0 0;
  color: #606266;
  font-size: 12px;
  line-height: 1.7;
}
.location-help {
  width: 100%;
}
.location-help .el-button {
  margin-top: 4px;
}
.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}
.field-grid :deep(.el-select),
.field-grid :deep(.el-date-editor) {
  width: 100%;
}
.package-upload {
  width: 100%;
}
.package-upload b,
.package-upload small {
  display: block;
}
.package-upload b {
  color: #303133;
  font-size: 13px;
}
.package-upload small {
  color: #909399;
  font-size: 11px;
  margin-top: 7px;
  word-break: break-all;
}
.upload-symbol {
  font-size: 30px;
  color: #409eff;
  margin-bottom: 10px;
}
.load-error {
  margin-bottom: 18px;
}
@media (max-width: 700px) {
  .field-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }
}
</style>
