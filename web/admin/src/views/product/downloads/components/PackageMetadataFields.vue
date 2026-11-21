<script setup>
import IconPicker from "@/components/icons/IconPicker.vue";
const draft = defineModel({ type: Object, required: true });
defineProps({ groups: Array, editionOptions: Array });
</script>
<template>
  <div class="field-grid">
    <el-form-item label="产品名称 *"
      ><el-input v-model="draft.product" placeholder="Deploy" /></el-form-item
    ><el-form-item label="版本号 *"
      ><el-input v-model="draft.version" placeholder="v1.0.0"
    /></el-form-item>
  </div>
  <div class="field-grid">
    <el-form-item label="图标"
      ><IconPicker v-model="draft.icon"
    /></el-form-item>
    <el-form-item label="下载分组"
      ><el-select v-model="draft.groupId" clearable placeholder="未分组"
        ><el-option
          v-for="group in groups"
          :key="group.id"
          :label="group.name"
          :value="group.id"
      /></el-select>
      <p class="field-note">
        分组名称和顺序可在“下载页设置”维护。
      </p></el-form-item
    >
    <el-form-item label="版本类型"
      ><el-select
        v-model="draft.edition"
        filterable
        allow-create
        default-first-option
        clearable
        placeholder="选择或输入，例如单节点 API 版"
        ><el-option
          v-for="edition in editionOptions"
          :key="edition"
          :label="edition"
          :value="edition" /></el-select
    ></el-form-item>
  </div>
  <div class="field-grid">
    <el-form-item label="展示顺序"
      ><el-input-number
        v-model="draft.sortOrder"
        :min="0"
        :max="99999"
        :precision="0"
      />
      <p class="field-note">
        数值越小越靠前，相同数值保留原顺序。
      </p></el-form-item
    >
  </div>
  <div class="field-grid">
    <el-form-item label="平台 / 系统 *"
      ><el-select
        v-model="draft.platform"
        filterable
        allow-create
        default-first-option
        placeholder="选择或输入平台"
        ><el-option
          v-for="name in [
            'Windows',
            'macOS',
            'Linux',
            'Android',
            'iOS',
            '跨平台',
            '离线部署包',
          ]"
          :key="name"
          :value="name"
          :label="name" /></el-select></el-form-item
    ><el-form-item label="处理器架构"
      ><el-input v-model="draft.arch" placeholder="x64 / ARM64 / 通用"
    /></el-form-item>
  </div>
</template>
<style scoped lang="scss">
.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}
.field-note {
  margin: 7px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
@media (max-width: 640px) {
  .field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
