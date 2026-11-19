<template>
  <AdminPage title="文档站设置" :loading="loading" :error="error">
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
            <h2>文档站信息</h2>
            <p>帮助用户识别当前文档的产品与适用版本</p>
          </div>
          <el-tag effect="plain">公开信息</el-tag>
        </template>

        <el-form label-position="top" class="settings-form"
          ><el-form-item label="文档站标题" required
            ><el-input
              v-model="form.title"
              maxlength="100"
              placeholder="Deploy 产品文档" /></el-form-item
          ><el-form-item label="文档站介绍"
            ><el-input
              v-model="form.description"
              type="textarea"
              :rows="4"
              maxlength="500"
              show-word-limit /></el-form-item
        ></el-form>
      </AdminFormSection>
      <aside class="help-card">
        <el-icon><Reading /></el-icon>
        <h3>让文档保持一致</h3>
        <router-link :to="$router.page('product/docs/modes/index')">管理文档模式 →</router-link>
        <router-link :to="$router.page('product/docs/versions/index')">管理产品版本 →</router-link>
        <p>这里维护文档站的公共信息。产品版本、发布状态与默认版本请在版本管理中维护。</p>
        <el-button link type="primary" @click="previewWebsite('/docs')">打开文档站 →</el-button>
      </aside>
    </div>
  </AdminPage>
</template>
<script setup>
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { reactive, ref } from 'vue'
import { Reading } from '@element-plus/icons-vue'
import AdminPage from '@/components/AdminPage.vue'
import { useUnsavedChanges } from '@/composables/useAdminResource'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import { saveDocSettings } from '@/api/documents'
import { previewWebsite } from '@/utils/website'
const { data, loading, saving, error, load, save, mutate } = useDocumentWorkspace()
const form = reactive({ title: '', description: '' }),
  baseline = ref('')
const confirmLeave = useUnsavedChanges(
  () => !!baseline.value && JSON.stringify(form) !== baseline.value,
)
async function initialize() {
  await load()
  if (!error.value) {
    for (const key in form) form[key] = data.value[key] || ''
    baseline.value = JSON.stringify(form)
  }
}
async function reload() {
  if (await confirmLeave()) await initialize()
}
async function submit() {
  await save(async () => {
    if (!form.title.trim()) throw new Error('请填写文档站标题')
    const submitted = JSON.stringify(form)
    await mutate(saveDocSettings, { ...form, revision: data.value.revision })
    baseline.value = submitted
  }, '文档站设置已保存')
}
initialize()
</script>
