<template>
  <AdminPage title="邮件服务" :loading="loading" :error="error">
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
            <h2>SMTP 发件服务</h2>
            <p>保存后，下次发信使用新的配置</p>
          </div>
          <el-tag effect="plain">{{
            data?.source === 'database' ? '后台配置' : '启动配置'
          }}</el-tag>
        </template>

        <el-form
          label-position="top"
          class="settings-form"
          :model="form"
          :rules="rules"
          ref="formRef"
        >
          <div class="form-columns">
            <el-form-item label="SMTP 主机" prop="smtpHost"
              ><el-input v-model="form.smtpHost" placeholder="smtp.qq.com" /></el-form-item
            ><el-form-item label="SMTP 端口" prop="smtpPort"
              ><el-input-number
                v-model="form.smtpPort"
                :min="1"
                :max="65535"
                controls-position="right"
            /></el-form-item>
          </div>
          <div class="transport-note">
            <el-icon><Lock /></el-icon
            ><span>{{
              form.smtpPort === 465 ? 'SSL / TLS 加密连接' : '连接后优先使用 STARTTLS 加密'
            }}</span
            ><small>{{
              form.smtpPort === 465 ? '常用于 465 端口' : '以 SMTP 服务商提供的端口为准'
            }}</small>
          </div>
          <div class="form-columns">
            <el-form-item label="发件邮箱" prop="fromEmail"
              ><el-input
                v-model="form.fromEmail"
                placeholder="name@qq.com"
                autocomplete="off" /></el-form-item
            ><el-form-item label="发件名称" prop="fromName"
              ><el-input v-model="form.fromName" placeholder="SiteHub 管理平台" maxlength="60"
            /></el-form-item>
          </div>
          <el-form-item label="SMTP 密码 / 授权码" prop="password"
            ><el-input
              v-model="form.password"
              type="password"
              show-password
              autocomplete="new-password"
              :placeholder="
                data?.hasPassword ? '已配置，留空保留原授权码' : '请输入邮箱服务提供的 SMTP 授权码'
              "
            />
            <p class="form-note">
              {{
                data?.hasPassword
                  ? '已有授权码已安全隐藏，只有填写新值时才会替换。'
                  : 'QQ 邮箱请使用开启 SMTP 服务后获得的授权码。'
              }}
            </p></el-form-item
          >
          <el-form-item label="验证码邮件主题" prop="subject"
            ><el-input v-model="form.subject" placeholder="SiteHub 登录验证码" maxlength="100"
          /></el-form-item>
        </el-form>

        <template #footer>
          <span>保存配置后，可通过下一次登录发信验证连接。</span
          ><el-button type="primary" :loading="saving" :disabled="!data || !!error" @click="submit"
            >保存设置</el-button
          >
        </template>
      </AdminFormSection>
      <aside class="help-card">
        <el-icon><Message /></el-icon>
        <h3>连接你的邮箱服务</h3>
        <p>QQ 邮箱常用主机 smtp.qq.com，SSL 端口为 465。请先在邮箱设置中开启 SMTP 服务。</p>
        <p>授权码由邮箱服务商生成，用于系统发信。更换授权码后请在此更新。</p>
        <p>首次进入时会读取后端已加载的邮箱配置，保存后由后台配置接管。</p>
      </aside>
    </div>
  </AdminPage>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { reactive, ref } from 'vue'
import { Message, Lock } from '@element-plus/icons-vue'
import AdminPage from '@/components/AdminPage.vue'
import { getEmailSettings, updateEmailSettings } from '@/api/settings'
import { useAdminResource, useUnsavedChanges } from '@/composables/useAdminResource'
const { data, loading, saving, error, load, save } = useAdminResource(getEmailSettings)
const form = reactive({
    smtpHost: '',
    smtpPort: 465,
    fromEmail: '',
    fromName: '',
    subject: '',
    password: '',
  }),
  baseline = ref(''),
  formRef = ref()
const rules = {
  smtpHost: [{ required: true, message: '请输入 SMTP 主机', trigger: 'blur' }],
  smtpPort: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  fromEmail: [
    { required: true, message: '请输入发件邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  fromName: [{ required: true, message: '请输入发件名称', trigger: 'blur' }],
  subject: [{ required: true, message: '请输入邮件主题', trigger: 'blur' }],
  password: [
    {
      validator: (_, value, callback) =>
        !value && !data.value?.hasPassword ? callback(new Error('请输入 SMTP 授权码')) : callback(),
      trigger: 'blur',
    },
  ],
}
const confirmLeave = useUnsavedChanges(
  () => !!baseline.value && JSON.stringify(form) !== baseline.value,
)
function reset() {
  for (const key in form) form[key] = key === 'password' ? '' : (data.value[key] ?? '')
  baseline.value = JSON.stringify(form)
  formRef.value?.clearValidate()
}
async function initialize() {
  await load()
  if (!error.value) reset()
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
async function submit() {
  if (!(await formRef.value?.validate().catch(() => false))) return
  await save(async () => {
    const payload = { ...form }
    if (!payload.password) delete payload.password
    data.value = (await updateEmailSettings(payload)).data
    reset()
  }, '邮件配置已保存，下次发信生效')
}
initialize()
</script>
<style scoped>
.transport-note {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 12px 15px;
  margin: 0 0 24px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #f5f7fa;
  color: #409eff;
  font-size: 12px;
  flex-wrap: wrap;
}
.transport-note .el-icon {
  font-size: 16px;
}
.transport-note small {
  color: #909399;
  margin-left: auto;
  font-size: 11px;
}
</style>
