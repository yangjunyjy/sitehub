<template>
  <el-form label-position="top" class="download-settings">
    <AdminFormSection title="页面信息" description="设置标题和介绍，各文件单独维护版本号。">
      <el-form-item label="页面标题"
        ><el-input v-model="value.title" placeholder="下载中心" maxlength="80"
      /></el-form-item>
      <el-form-item label="页面介绍"
        ><el-input v-model="value.description" type="textarea" :rows="3"
      /></el-form-item>
    </AdminFormSection>
    <AdminFormSection
      title="下载分组"
      description="按产品、用途或部署方式分组，排列顺序与官网一致。"
    >
      <template #actions><el-button size="small" @click="addGroup">新增分组</el-button></template>
      <p v-if="!value.groups.length" class="field-note">
        尚未建立分组，已有文件统一展示在全部文件中。
      </p>
      <div v-for="(group, index) in value.groups" :key="group.id" class="repeat-card">
        <div class="section-title">
          <strong>{{ group.name || '新分组' }}</strong>
          <div>
            <el-button link :disabled="index === 0" @click="moveGroup(index, -1)">上移</el-button
            ><el-button
              link
              :disabled="index === value.groups.length - 1"
              @click="moveGroup(index, 1)"
              >下移</el-button
            ><el-button
              link
              type="danger"
              :disabled="inUse(group.id) > 0"
              @click="value.groups.splice(index, 1)"
              >移除</el-button
            >
          </div>
        </div>
        <el-form-item label="分组名称" required
          ><el-input v-model="group.name" placeholder="例如：Deploy、运维工具" maxlength="80"
        /></el-form-item>
        <el-form-item label="分组说明"
          ><el-input
            v-model="group.description"
            type="textarea"
            :rows="2"
            placeholder="简要介绍用途或适用场景"
        /></el-form-item>
        <p v-if="inUse(group.id)" class="field-note">
          {{ inUse(group.id) }} 个文件使用此分组，移除前请先调整文件归属。
        </p>
      </div>
    </AdminFormSection>
    <AdminFormSection title="安装说明" description="需要展示安装命令时添加；留空则不展示此区块。">
      <template #actions
        ><el-button
          size="small"
          @click="
            value.installOptions.push({
              id: makeID(),
              title: '',
              badge: '',
              description: '',
              command: '',
            })
          "
          >新增方式</el-button
        ></template
      >
      <div v-for="(item, index) in value.installOptions" :key="item.id" class="repeat-card">
        <div class="section-title">
          <strong>{{ item.title || `方式 ${index + 1}` }}</strong
          ><el-button link type="danger" @click="value.installOptions.splice(index, 1)"
            >移除</el-button
          >
        </div>
        <div class="field-grid">
          <el-form-item label="标题"><el-input v-model="item.title" /></el-form-item
          ><el-form-item label="标签"><el-input v-model="item.badge" /></el-form-item>
        </div>
        <el-form-item label="说明"><el-input v-model="item.description" /></el-form-item
        ><el-form-item label="安装命令"
          ><el-input v-model="item.command" type="textarea" :rows="2"
        /></el-form-item>
      </div>
    </AdminFormSection>
    <AdminFormSection title="环境要求" description="可选，留空则不展示。">
      <template #actions
        ><el-button size="small" @click="value.requirements.push({ label: '', value: '' })"
          >新增要求</el-button
        ></template
      >
      <div v-for="(item, index) in value.requirements" :key="index" class="requirement">
        <el-form-item label="项目"><el-input v-model="item.label" /></el-form-item
        ><el-form-item label="要求"><el-input v-model="item.value" /></el-form-item
        ><el-button link type="danger" @click="value.requirements.splice(index, 1)">移除</el-button>
      </div>
    </AdminFormSection>
    <AdminFormSection
      title="联系支持"
      description="填写标题后在下载页底部展示，联系邮箱沿用站点基本设置。"
    >
      <el-form-item label="标题"><el-input v-model="value.supportTitle" /></el-form-item
      ><el-form-item label="说明"
        ><el-input v-model="value.supportDescription" type="textarea" :rows="2"
      /></el-form-item>
    </AdminFormSection>
    <el-collapse
      ><el-collapse-item title="发布默认值" name="defaults"
        ><p class="field-note">用于创建下载版本时的默认信息，每个安装包可单独设置。</p>
        <div class="field-grid">
          <el-form-item label="默认版本"><el-input v-model="value.version" /></el-form-item
          ><el-form-item label="默认发布日期"
            ><el-input v-model="value.releaseDate"
          /></el-form-item>
        </div>
        <el-form-item label="默认架构说明"
          ><el-input v-model="value.supportArchitectures" /></el-form-item></el-collapse-item
    ></el-collapse>
  </el-form>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { createUUID } from '@/utils/uuid'
const value = defineModel({ type: Object, required: true })
const makeID = () => `install-${createUUID()}`
const addGroup = () =>
  value.value.groups.push({ id: `group-${createUUID()}`, name: '', description: '' })
const inUse = (id) => value.value.packages.filter((item) => item.groupId === id).length
const moveGroup = (index, direction) => {
  const groups = value.value.groups
  const [group] = groups.splice(index, 1)
  groups.splice(index + direction, 0, group)
}
</script>
<style lang="scss" scoped>
.download-settings {
  display: grid;
  gap: 18px;
}
.field-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}
.section-title strong {
  font-size: 14px;
  color: #303133;
}
.repeat-card {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 18px;
  margin-bottom: 14px;
  background: #fafbfe;
}
.repeat-card:last-child {
  margin-bottom: 0;
}
.requirement {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 10px;
  align-items: center;
}
.field-note {
  margin: 0 0 14px;
  color: #606266;
  font-size: 12px;
  line-height: 1.7;
}
@media (max-width: 600px) {
  .field-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }
  .section-title {
    flex-wrap: wrap;
  }
}
</style>
