<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import request from '@/utils/request'
const form = ref(null)
const busy = ref(false)
const error = ref('')
const original = ref('')
const dirty = computed(() => !!form.value && JSON.stringify(form.value) !== original.value)
async function load() {
  error.value = ''
  try {
    const { data } = await request.get('/blog/about')
    form.value = { homeVerse: data.homeVerse, homeAnimation: data.homeAnimation }
    original.value = JSON.stringify(form.value)
  } catch (failure) { error.value = failure.message }
}
async function persist() {
  if (!dirty.value) return
  const settings = { ...form.value, homeVerse: form.value.homeVerse.trim() }
  if (!settings.homeVerse) throw new Error('请填写首页短句')
  await request.post('/blog/home/settings', settings, { silentError: true })
  form.value = settings
  original.value = JSON.stringify(settings)
}
defineExpose({ persist, dirty, load })
async function save() {
  error.value = ''
  busy.value = true
  try {
    await persist()
    ElMessage.success('博客首页设置已保存')
  } catch (failure) {
    if (failure !== 'cancel' && failure !== 'close') error.value = failure.message || '保存失败，请重试'
  } finally { busy.value = false }
}
onMounted(load)
</script>
<template>
  <AdminFormSection>
    <template #header><h2>博客首页</h2></template>
    <el-form v-if="form" label-position="top" :disabled="busy" @submit.prevent="save">
      <el-form-item label="播放首页动画"><el-switch v-model="form.homeAnimation" aria-label="播放首页动画" /></el-form-item>
      <el-form-item label="首页短句"><el-input v-model="form.homeVerse" maxlength="120" show-word-limit /></el-form-item>
      <el-button native-type="submit" :disabled="!form.homeVerse?.trim()" :loading="busy">保存博客首页设置</el-button>
      <p v-if="error" role="alert">{{ error }}</p>
    </el-form>
    <div v-else-if="error" role="alert">{{ error }} <el-button @click="load">重试</el-button></div>
  </AdminFormSection>
</template>
