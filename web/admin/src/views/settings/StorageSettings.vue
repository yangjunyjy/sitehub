<template>
  <AdminPage title="文件存储" :loading="loading" :error="error">
    <template #actions
      ><el-button @click="reload">重新加载</el-button
      ><el-button type="primary" :loading="saving" :disabled="!data || !!error" @click="submit"
        >保存设置</el-button
      ></template
    >
    <div class="settings-layout">
      <AdminFormSection>
        <template #header>
          <div>
            <h2>存储配置</h2>
            <p>保存后应用于后续文件上传</p>
          </div>
          <el-tag effect="plain" type="success">{{ providerName }}</el-tag>
        </template>

        <el-form label-position="top" class="settings-form">
          <el-form-item label="存储方式"
            ><div class="provider-options">
              <button
                v-for="item in providers"
                :key="item.value"
                type="button"
                :class="{ selected: form.provider === item.value }"
                :aria-pressed="form.provider === item.value"
                @click="form.provider = item.value"
              >
                <el-icon><component :is="item.icon" /></el-icon><strong>{{ item.label }}</strong
                ><small>{{ item.description }}</small
                ><el-icon v-if="form.provider === item.value" class="provider-check"
                  ><CircleCheckFilled
                /></el-icon>
              </button></div
          ></el-form-item>
          <template v-if="form.provider === 'local'"
            ><el-form-item label="本地目录" required
              ><el-input v-model="form.localPath" placeholder="/data/files" />
              <p class="form-note">使用后端可写入的目录，例如 ./data/files。</p></el-form-item
            ><el-form-item label="公开访问前缀"
              ><el-input v-model="form.urlPrefix" placeholder="https://deploy.example.com" />
              <p class="form-note">
                留空时生成相对路径，填写时请使用可访问的后端域名。
              </p></el-form-item
            ></template
          >
          <template v-else
            ><div class="form-columns">
              <el-form-item label="Endpoint" required
                ><el-input
                  v-model="form.endpoint"
                  placeholder="https://s3.example.com" /></el-form-item
              ><el-form-item label="Bucket" required
                ><el-input v-model="form.bucket" placeholder="deploy-releases"
              /></el-form-item>
            </div>
            <div class="form-columns">
              <el-form-item label="Region"
                ><el-input v-model="form.region" placeholder="us-east-1" /></el-form-item
              ><el-form-item label="Access Key" required
                ><el-input v-model="form.accessKey" autocomplete="off"
              /></el-form-item>
            </div>
            <el-form-item label="Secret Key"
              ><el-input
                v-model="form.secretKey"
                type="password"
                show-password
                autocomplete="new-password"
                placeholder="保存时请输入存储密钥"
              />
              <p class="form-note">密钥不回显；保存对象存储配置时需要重新填写。</p></el-form-item
            ><el-form-item label="CDN / 公开域名"
              ><el-input
                v-model="form.cdn"
                placeholder="https://downloads.example.com" /></el-form-item
            ><el-form-item label="寻址模式"
              ><el-switch v-model="form.pathStyle" active-text="Path Style" />
              <p class="form-note">
                MinIO 通常启用 Path Style；OSS 通常使用 Virtual Host。
              </p></el-form-item
            ></template
          >
          <div class="storage-divider">上传规则</div>
          <div class="form-columns">
            <el-form-item label="安装包大小上限（MB）"
              ><el-input-number
                v-model="maxMB"
                :min="1"
                :max="2048"
                controls-position="right" /></el-form-item
            ><el-form-item label="允许的文件扩展名"
              ><el-input v-model="extensions" placeholder=".zip, .tar.gz, .exe, .dmg" />
              <p class="form-note">使用英文逗号分隔，如 .zip, .exe。</p></el-form-item
            >
          </div>
        </el-form>

        <template #footer>
          <span>更换存储后，已有文件仍保留在原位置。</span
          ><el-button type="primary" :loading="saving" :disabled="!data || !!error" @click="submit"
            >保存设置</el-button
          >
        </template>
      </AdminFormSection>
      <aside class="help-card">
        <el-icon><Files /></el-icon>
        <h3>文件与版本分开管理</h3>
        <p>本页管理文件存储位置和上传规则。</p>
        <p>安装包的版本、平台、下载地址与发布状态，请到下载管理页面维护。</p>
        <p>更换存储不会自动搬迁已有文件，原下载地址需要保持可访问。</p>
        <router-link to="/dashboard/downloads">前往下载管理 →</router-link>
      </aside>
    </div>
  </AdminPage>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { computed, reactive, ref } from 'vue'
import { Files, FolderOpened, Cloudy, Box, CircleCheckFilled } from '@element-plus/icons-vue'
import AdminPage from '@/components/AdminPage.vue'
import { getWebsiteStorage, updateWebsiteStorage } from '@/api/website'
import { useAdminResource, useUnsavedChanges } from '@/composables/useAdminResource'
import { clone } from '@/utils/website'
const { data, loading, saving, error, load, save } = useAdminResource(getWebsiteStorage)
const form = reactive({
    provider: 'local',
    localPath: '',
    urlPrefix: '',
    endpoint: '',
    bucket: '',
    region: '',
    accessKey: '',
    secretKey: '',
    cdn: '',
    pathStyle: true,
    maxSize: 2147483648,
    allowedExts: [],
  }),
  extensions = ref(''),
  baseline = ref('')
const providers = [
  { value: 'local', label: '本地存储', description: '服务器文件目录', icon: FolderOpened },
  { value: 'minio', label: 'MinIO', description: '自托管对象存储', icon: Box },
  { value: 'oss', label: 'OSS / S3', description: '兼容 S3 的对象存储', icon: Cloudy },
]
const providerName = computed(() => providers.find((item) => item.value === form.provider)?.label)
const maxMB = computed({
  get: () => Math.round(form.maxSize / 1024 / 1024),
  set: (value) => (form.maxSize = (value || 1) * 1024 * 1024),
})
const snapshot = () => JSON.stringify({ form, extensions: extensions.value })
const confirmLeave = useUnsavedChanges(() => !!baseline.value && snapshot() !== baseline.value)
function reset() {
  Object.assign(form, clone(data.value.storage), { secretKey: '' })
  extensions.value = (form.allowedExts || []).join(', ')
  baseline.value = snapshot()
}
async function initialize() {
  await load()
  if (!error.value) reset()
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
async function submit() {
  await save(async () => {
    const payload = clone(form)
    payload.secretKey = form.secretKey
    if (payload.provider !== 'local' && !payload.secretKey) throw new Error('保存对象存储配置时请填写访问密钥')
    payload.allowedExts = extensions.value
      .split(/[,，]/)
      .map((value) => value.trim())
      .filter(Boolean)
    const result = await updateWebsiteStorage(payload)
    data.value.storage = result.data
    reset()
  }, '存储配置已保存并生效')
}
initialize()
</script>
<style scoped>
.provider-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
}
.provider-options button {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 7px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  text-align: left;
  padding: 18px;
  color: #606266;
  cursor: pointer;
  transition: 0.15s;
}
.provider-options button:hover {
  border-color: #ebeef5;
}
.provider-options button.selected {
  border-color: #ebeef5;
  background: #f5f7fa;
  box-shadow: 0 0 0 1px #409eff;
}
.provider-options button > .el-icon {
  font-size: 23px;
  color: #409eff;
  margin-bottom: 3px;
}
.provider-options strong {
  font-size: 13px;
  color: #303133;
}
.provider-options small {
  font-size: 11px;
  color: #909399;
}
.provider-options .provider-check {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 16px;
  color: #409eff;
}
.storage-divider {
  padding-top: 20px;
  margin: 12px 0 22px;
  border-top: 1px solid #ebeef5;
  font-size: 14px;
  font-weight: 600;
}
@media (max-width: 760px) {
  .provider-options {
    grid-template-columns: minmax(0, 1fr);
  }
  .provider-options button {
    padding: 14px 18px;
  }
  .provider-options small {
    display: none;
  }
}
</style>
