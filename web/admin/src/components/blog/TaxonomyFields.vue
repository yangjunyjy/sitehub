<script setup>
import { computed } from "vue";
const props = defineProps({ kind: { type: String, required: true } });
const tag = computed(() => props.kind === "tags");
const name = defineModel("name", { type: String });
const slug = defineModel("slug", { type: String });
</script>
<template>
  <el-form-item :label="tag ? '标签名称' : '分类名称'" prop="name" required>
    <el-input
      v-model="name"
      :placeholder="tag ? '请输入标签名称' : '请输入分类名称'"
      :maxlength="tag ? 50 : undefined"
      :show-word-limit="tag"
    />
  </el-form-item>
  <el-form-item :label="tag ? '友好URL' : '分类标识'" prop="slug" required>
    <el-input
      v-model="slug"
      :placeholder="tag ? '请输入友好URL后缀' : '请输入分类标识（英文）'"
      :maxlength="tag ? 50 : undefined"
      :show-word-limit="tag"
    >
      <template v-if="tag" #prepend>tag/</template>
    </el-input>
    <div v-if="tag" class="form-tips">
      将显示为 example.com/tag/{{ slug || "your-slug" }}
    </div>
  </el-form-item>
</template>
<style scoped lang="scss">
.form-tips {
  margin-top: 6px;
  color: #606266;
  font-size: 12px;
  line-height: 1.6;
}
</style>
