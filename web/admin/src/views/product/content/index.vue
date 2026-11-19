<template>
  <AdminPage
    title="官网内容"
    description="按页面维护文案，发布后同步到官网。"
    :loading="loading"
    :error="error"
  >
    <template #actions
      ><span class="save-state">{{ dirty ? '有未发布修改' : '内容已同步' }}</span
      ><el-button @click="preview">预览页面</el-button
      ><el-button @click="reload">重新加载</el-button
      ><el-button type="primary" :loading="saving" :disabled="!manifest || !!error" @click="submit"
        >发布内容</el-button
      ></template
    >
    <div v-if="manifest">
      <nav class="content-tabs" aria-label="内容分区">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          :class="{ active: activeSection === tab.key }"
          :aria-pressed="activeSection === tab.key"
          @click="activeSection = tab.key"
        >
          {{ tab.label }}
        </button>
      </nav>
      <div v-show="activeSection === 'home'" class="editor-grid hero-grid">
        <AdminFormSection>
          <template #header>
            <div>
              <h2>首页首屏</h2>
            </div>
          </template>
          <el-form label-position="top">
            <el-form-item label="眉题"
              ><el-input v-model="manifest.content.home.eyebrow"
            /></el-form-item>
            <div class="two-columns">
              <el-form-item label="主标题"
                ><el-input v-model="manifest.content.home.title" /></el-form-item
              ><el-form-item label="高亮词"
                ><el-input v-model="manifest.content.home.highlight"
              /></el-form-item>
            </div>
            <el-form-item label="说明"
              ><el-input v-model="manifest.content.home.description" type="textarea" :rows="3"
            /></el-form-item>
            <div class="two-columns">
              <el-form-item label="主按钮"
                ><el-input v-model="manifest.content.home.primaryAction" /></el-form-item
              ><el-form-item label="次按钮"
                ><el-input v-model="manifest.content.home.secondaryAction"
              /></el-form-item>
            </div>
          </el-form>
        </AdminFormSection>
      </div>

      <AdminFormSection v-show="activeSection === 'capabilities'" class="wide-card">
        <template #header>
          <div>
            <h2>产品能力</h2>
          </div>
          <el-button size="small" @click="addCapability">新增能力</el-button>
        </template>
        <div class="capability-editors">
          <el-form
            v-for="(item, index) in manifest.content.capabilities"
            :key="index"
            label-position="top"
            class="content-item-card"
          >
            <div class="item-heading">
              <strong>{{ item.title || `能力 ${index + 1}` }}</strong
              ><el-button type="danger" link @click="manifest.content.capabilities.splice(index, 1)"
                >移除能力</el-button
              >
            </div>
            <div class="two-columns">
              <el-form-item label="能力标题"><el-input v-model="item.title" /></el-form-item
              ><el-form-item label="显示序号"><el-input v-model="item.index" /></el-form-item>
            </div>
            <el-form-item label="能力说明"
              ><el-input v-model="item.description" type="textarea" :rows="3"
            /></el-form-item>
            <div class="two-columns">
              <el-form-item label="图标"
                ><el-select v-model="item.key"
                  ><el-option
                    v-for="icon in capabilityIcons"
                    :key="icon.value"
                    :label="icon.label"
                    :value="icon.value" /></el-select></el-form-item
              ><el-form-item label="标签"
                ><el-select
                  v-model="item.tags"
                  multiple
                  filterable
                  allow-create
                  default-first-option
                  placeholder="输入后按回车添加"
              /></el-form-item>
            </div>
          </el-form>
        </div>
      </AdminFormSection>

      <div v-show="activeSection === 'home'" class="editor-grid wide-card">
        <AdminFormSection>
          <template #header>
            <div>
              <h2>网站导航</h2>
            </div>
            <el-button size="small" @click="addNavigation">新增导航</el-button>
          </template>
          <div class="repeat-list compact-list">
            <div
              v-for="(item, index) in manifest.content.navigation"
              :key="index"
              class="repeat-row simple-row"
            >
              <el-input v-model="item.label" placeholder="名称" /><el-input
                v-model="item.path"
                placeholder="/path"
              /><el-button type="danger" link @click="manifest.content.navigation.splice(index, 1)"
                >移除</el-button
              >
            </div>
          </div>
        </AdminFormSection>
        <AdminFormSection>
          <template #header>
            <div>
              <h2>首页分区文案</h2>
            </div>
          </template>
          <el-form label-position="top">
            <el-form-item label="首屏能力标签"
              ><el-input
                :model-value="manifest.content.home.meta.join(', ')"
                @update:model-value="manifest.content.home.meta = splitTags($event)"
            /></el-form-item>
            <div class="two-columns">
              <el-form-item label="能力区标题"
                ><el-input v-model="manifest.content.home.capabilityTitle" /></el-form-item
              ><el-form-item label="交付区标题"
                ><el-input v-model="manifest.content.home.flowTitle"
              /></el-form-item>
            </div>
            <div class="two-columns">
              <el-form-item label="能力区说明"
                ><el-input
                  v-model="manifest.content.home.capabilityIntro"
                  type="textarea"
                  :rows="2" /></el-form-item
              ><el-form-item label="交付区说明"
                ><el-input v-model="manifest.content.home.flowIntro" type="textarea" :rows="2"
              /></el-form-item>
            </div>
            <div class="two-columns">
              <el-form-item label="用户区标题"
                ><el-input v-model="manifest.content.home.audienceTitle" /></el-form-item
              ><el-form-item label="博客区标题"
                ><el-input v-model="manifest.content.home.blogTitle"
              /></el-form-item>
            </div>
            <el-form-item label="用户区说明"
              ><el-input v-model="manifest.content.home.audienceIntro"
            /></el-form-item>
            <el-form-item label="底部行动标题"
              ><el-input v-model="manifest.content.home.cta"
            /></el-form-item>
          </el-form>
        </AdminFormSection>
      </div>

      <div v-show="activeSection === 'home'" class="editor-grid wide-card">
        <AdminFormSection>
          <template #header>
            <div>
              <h2>交付步骤</h2>
            </div>
            <el-button size="small" @click="addDeliveryStep">新增步骤</el-button>
          </template>
          <div class="repeat-list">
            <div
              v-for="(item, index) in manifest.content.deliverySteps"
              :key="index"
              class="repeat-row flow-row"
            >
              <el-input v-model="item.index" placeholder="序号" /><el-input
                v-model="item.key"
                placeholder="英文键"
              /><el-input v-model="item.title" placeholder="标题" /><el-input
                v-model="item.description"
                placeholder="说明"
              /><el-button
                type="danger"
                link
                @click="manifest.content.deliverySteps.splice(index, 1)"
                >移除</el-button
              >
            </div>
          </div>
        </AdminFormSection>
        <AdminFormSection>
          <template #header>
            <div>
              <h2>目标用户</h2>
            </div>
            <el-button size="small" @click="addAudience">新增用户组</el-button>
          </template>
          <div class="repeat-list">
            <div
              v-for="(item, index) in manifest.content.audiences"
              :key="index"
              class="audience-editor"
            >
              <div>
                <el-input v-model="item.index" placeholder="序号" /><el-input
                  v-model="item.title"
                  placeholder="用户组"
                /><el-button type="danger" link @click="manifest.content.audiences.splice(index, 1)"
                  >移除</el-button
                >
              </div>
              <el-input v-model="item.headline" placeholder="主张" /><el-input
                v-model="item.description"
                type="textarea"
                :rows="2"
                placeholder="说明"
              />
            </div>
          </div>
        </AdminFormSection>
      </div>
    </div>
  </AdminPage>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { computed, ref } from 'vue'
import { websiteURL } from '@/utils/downloads'
import AdminPage from '@/components/AdminPage.vue'
import { getWebsiteContent, updateWebsiteContent } from '@/api/website'
import { useAdminResource, useUnsavedChanges } from '@/composables/useAdminResource'
const { data: manifest, loading, saving, error, load, save } = useAdminResource(getWebsiteContent)
const baseline = ref('')
const activeSection = ref('home')
const tabs = [
  { key: 'home', label: '首页与导航' },
  { key: 'capabilities', label: '产品能力' },
]
const capabilityIcons = [
  { value: 'assets', label: '资产' },
  { value: 'pipeline', label: '流水线' },
  { value: 'playbook', label: '自动化' },
  { value: 'artifact', label: '制品' },
  { value: 'terminal', label: '终端' },
  { value: 'shield', label: '安全' },
]
const preview = () =>
  window.open(
    websiteURL(activeSection.value === 'capabilities' ? '/#capabilities' : '/'),
    '_blank',
    'noopener,noreferrer',
  )
const dirty = computed(() => !!baseline.value && signature() !== baseline.value)
const signature = () => {
  if (!manifest.value) return ''
  const { brand, footer, ...content } = manifest.value.content
  return JSON.stringify(content)
}
const confirmLeave = useUnsavedChanges(() => !!baseline.value && signature() !== baseline.value)
async function initialize() {
  await load()
  if (!error.value) {
    manifest.value.content.navigation = manifest.value.content.navigation.filter(
      (item) => !['/product', '/about'].includes(item.path.split(/[?#]/)[0]),
    )
    baseline.value = signature()
  }
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
async function submit() {
  await save(async () => {
    const latest = (await getWebsiteContent()).data.content
    await updateWebsiteContent({
      ...manifest.value.content,
      brand: latest.brand,
      footer: latest.footer,
    })
    baseline.value = signature()
  }, '官网内容已发布')
}
const splitTags = (value) =>
  String(value)
    .split(/[,\uFF0C]/)
    .map((item) => item.trim())
    .filter(Boolean)
const addCapability = () =>
  manifest.value.content.capabilities.push({
    key: 'assets',
    index: String(manifest.value.content.capabilities.length + 1).padStart(2, '0'),
    title: '',
    description: '',
    tags: [],
  })
const addNavigation = () => manifest.value.content.navigation.push({ label: '新导航', path: '/' })
const addDeliveryStep = () =>
  manifest.value.content.deliverySteps.push({
    index: String(manifest.value.content.deliverySteps.length + 1).padStart(2, '0'),
    key: 'STEP',
    title: '',
    description: '',
  })
const addAudience = () =>
  manifest.value.content.audiences.push({
    index: String(manifest.value.content.audiences.length + 1).padStart(2, '0'),
    title: '',
    headline: '',
    description: '',
  })
initialize()
</script>
<style lang="scss" scoped>
.content-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 20px;
  padding: 8px;
  background: white;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}
.content-tabs button {
  min-height: 38px;
  padding: 8px 20px;
  border: 0;
  border-radius: 6px;
  color: #606266;
  background: transparent;
  font: inherit;
  font-size: 14px;
  cursor: pointer;
}
.content-tabs button.active {
  color: #409eff;
  background: #f5f7fa;
}
.save-state {
  color: #606266;
  font-size: 12px;
  margin-right: 8px;
}
.capability-editors {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}
.content-item-card {
  min-width: 0;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 20px;
  background: #fafbfe;
}
.item-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
  font-size: 14px;
}
@media (max-width: 1100px) {
  .capability-editors {
    grid-template-columns: 1fr;
  }
}

.editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}
.hero-grid {
  grid-template-columns: 1fr;
}
.wide-card {
  margin-top: 18px;
}
.two-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}
.repeat-list {
  display: grid;
  gap: 10px;
}
.repeat-row {
  display: grid;
  align-items: center;
  gap: 8px;
}
.capability-row {
  grid-template-columns: 60px 90px 140px minmax(150px, 1.5fr) minmax(130px, 1fr) 48px;
}
.simple-row {
  grid-template-columns: 1fr 1.4fr 48px;
}
.flow-row {
  grid-template-columns: 50px 70px 100px minmax(90px, 1fr) 48px;
}
.audience-editor {
  display: grid;
  gap: 8px;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 7px;
  background: #fafcff;
}
.audience-editor > div {
  display: grid;
  grid-template-columns: 60px 1fr 48px;
  gap: 8px;
}
.sub-editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 22px;
  margin-top: 18px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}
.sub-editor-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 6px 0 10px;
  font-size: 13px;
}
.architecture-row {
  grid-template-columns: 100px 1fr 48px;
}
.security-row {
  grid-template-columns: 130px 1fr 48px;
}
.principle-row {
  grid-template-columns: 60px 170px 1fr 48px;
}
.compact-list {
  max-height: 420px;
  overflow: auto;
}
@media (max-width: 1200px) {
  .editor-grid,
  .sub-editor-grid {
    grid-template-columns: 1fr;
  }
  .capability-row {
    grid-template-columns: 60px 1fr 1fr;
  }
  .flow-row {
    grid-template-columns: 50px 90px 1fr;
  }
  .flow-row > *:nth-child(4) {
    grid-column: 1/4;
  }
}
@media (max-width: 760px) {
  .two-columns,
  .capability-row,
  .simple-row,
  .architecture-row,
  .security-row,
  .flow-row,
  .principle-row {
    grid-template-columns: minmax(0, 1fr);
  }
  .flow-row > *:nth-child(4) {
    grid-column: auto;
  }
  .audience-editor > div {
    grid-template-columns: 60px 1fr;
  }
  .repeat-row {
    padding: 14px;
    border: 1px solid #ebeef5;
    border-radius: 6px;
  }
  .repeat-row .el-button {
    justify-self: start;
  }
}
</style>
