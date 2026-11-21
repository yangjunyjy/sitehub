<template>
  <AdminPage
    title="官网内容"
    :loading="loading"
    :error="error"
  >
    <el-button v-if="error && !manifest" @click="reload">重新加载</el-button>
    <div v-if="manifest">
      <div class="content-toolbar">
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
        <div class="content-actions">
          <span class="save-state">{{ dirty ? '有未发布修改' : '内容已同步' }}</span>
          <el-button @click="preview">预览页面</el-button>
          <el-button @click="reload">重新加载</el-button>
          <el-button type="primary" :loading="saving" :disabled="!!error" @click="submit">发布内容</el-button>
        </div>
      </div>
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
            <el-form-item label="粒子起始形态">
              <el-radio-group v-model="manifest.content.home.particleShape" aria-label="粒子起始形态">
                <el-radio-button v-for="(scene, id) in particleScenes" :key="id" :value="id">{{ scene.label }}</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <div class="two-columns">
              <el-form-item label="自动播放">
                <el-switch v-model="manifest.content.home.particleAutoplay" aria-label="自动播放" />
              </el-form-item>
              <el-form-item label="允许旋转交互">
                <el-switch v-model="manifest.content.home.particleInteractive" aria-label="允许旋转交互" />
              </el-form-item>
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
                ><IconPicker :model-value="item.icon || item.key" @update:model-value="item.icon = $event" /></el-form-item
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
import { particleScenes } from '@shared/product/particle-scenes'
import { createUUID } from '@/utils/uuid'
import IconPicker from '@/components/icons/IconPicker.vue'
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
    key: createUUID(),
    icon: 'business/asset-center',
    index: String(manifest.value.content.capabilities.length + 1).padStart(2, '0'),
    title: '',
    description: '',
    tags: [],
  })
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
<style lang="scss" scoped src="./content.scss"></style>
