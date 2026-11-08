<template>
  <AdminPage title="基本设置" :loading="loading" :error="error">
    <template #actions
      ><el-button @click="reload">重新加载</el-button
      ><el-button type="primary" :loading="saving" :disabled="!data || !!error" @click="submit"
        >保存设置</el-button
      ></template
    >
    <div class="settings-layout">
      <div class="basic-sections">
        <AdminFormSection>
          <template #header>
            <div>
              <h2>品牌与联系信息</h2>
              <p>显示在网站导航、页脚和联系入口</p>
            </div>
            <el-tag effect="plain">站点信息</el-tag>
          </template>

          <el-form label-position="top" class="settings-form"
            ><div class="form-columns">
              <el-form-item label="产品简称" required
                ><el-input v-model="form.brand.name" placeholder="Deploy" /></el-form-item
              ><el-form-item label="产品全称"
                ><el-input v-model="form.brand.fullName"
              /></el-form-item>
            </div>
            <el-form-item label="一句话介绍"
              ><el-input v-model="form.brand.description" type="textarea" :rows="3"
            /></el-form-item>
            <div class="form-columns">
              <el-form-item label="联系邮箱"
                ><el-input
                  v-model="form.brand.contactEmail"
                  placeholder="contact@example.com" /></el-form-item
              ><el-form-item label="备案号"
                ><el-input v-model="form.brand.recordNumber"
              /></el-form-item></div
          ></el-form>
        </AdminFormSection>
        <AdminFormSection>
          <template #header>
            <div>
              <h2>页脚信息</h2>
              <p>官网底部的版权、标语与说明</p>
            </div>
          </template>

          <el-form label-position="top" class="settings-form"
            ><el-form-item label="版权文字"
              ><el-input v-model="form.brand.copyright" /></el-form-item
            ><el-form-item label="页脚标语"><el-input v-model="form.footer.slogan" /></el-form-item
            ><el-form-item label="页脚声明"
              ><el-input v-model="form.footer.statement" /></el-form-item
          ></el-form>
        </AdminFormSection>
      </div>
      <aside class="help-card">
        <el-icon><Setting /></el-icon>
        <h3>站点基础信息</h3>
        <p>此处的联系邮箱用于官网展示。</p>
        <p>登录验证码的发件邮箱和授权码，请在“邮件服务”页面配置。</p>
        <p>首页首屏、产品能力与关于页内容，仍在“官网内容”页面维护。</p>
        <router-link to="/dashboard/website">前往官网内容 →</router-link>
      </aside>
    </div>
  </AdminPage>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { reactive, ref } from 'vue'
import { Setting } from '@element-plus/icons-vue'
import AdminPage from '@/components/AdminPage.vue'
import { getWebsiteContent, updateWebsiteContent } from '@/api/website'
import { useAdminResource, useUnsavedChanges } from '@/composables/useAdminResource'
import { clone } from '@/utils/website'
const { data, loading, saving, error, load, save } = useAdminResource(getWebsiteContent)
const form = reactive({ brand: {}, footer: {} }),
  baseline = ref('')
const confirmLeave = useUnsavedChanges(
  () => !!baseline.value && JSON.stringify(form) !== baseline.value,
)
async function initialize() {
  await load()
  if (!error.value) {
    form.brand = clone(data.value.content.brand)
    form.footer = clone(data.value.content.footer)
    baseline.value = JSON.stringify(form)
  }
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
async function submit() {
  await save(async () => {
    if (!form.brand.name?.trim()) throw new Error('请填写产品简称')
    const content = (await getWebsiteContent()).data.content
    content.brand = clone(form.brand)
    content.footer = clone(form.footer)
    await updateWebsiteContent(content)
    baseline.value = JSON.stringify(form)
  }, '基本设置已保存')
}
initialize()
</script>
<style scoped>
.basic-sections {
  display: grid;
  gap: 20px;
}
</style>
