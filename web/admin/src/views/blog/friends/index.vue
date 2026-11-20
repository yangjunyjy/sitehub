<script setup>
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import AdminPage from '@/components/AdminPage.vue'
import request from '@/utils/request'
import { uint64 } from '@/utils/native-api.mjs'
const kind = ref('friends'), status = ref('pending'), page = ref(1)
const items = ref([]), total = ref(0), loading = ref(false), saving = ref(false), error = ref('')
let generation = 0
async function load() {
  const current = ++generation
  loading.value = true
  error.value = ''
  items.value = []
  try {
    const { data } = await request.get('/blog/admin/community', { params: { kind: kind.value, status: status.value, page: page.value } })
    if (current === generation) { items.value = data.items; total.value = data.total }
  } catch (failure) { if (current === generation) error.value = failure.message }
  finally { if (current === generation) loading.value = false }
}
async function review(row, next) {
  saving.value = true
  try {
    await request.post('/blog/admin/community', { kind: kind.value, id: uint64(row.id), status: next })
    ElMessage.success('审核已保存')
    if (page.value > 1 && items.value.length === 1) page.value--
    else await load()
  } catch (failure) { if (failure?.message) error.value = failure.message }
  finally { saving.value = false }
}
watch([kind, status], () => { page.value = 1 }, { flush: 'sync' })
watch([kind, status, page], load, { immediate: true })
</script>
<template>
  <AdminPage title="友链管理" :loading="loading" :error="error">
    <template #actions><el-button @click="load">刷新</el-button></template>
    <el-tabs v-model="kind"><el-tab-pane label="友链申请" name="friends"/><el-tab-pane label="留言板" name="messages"/><el-tab-pane label="文章评论" name="comments"/></el-tabs>
    <el-radio-group v-model="status" aria-label="审核状态"><el-radio-button value="pending">待审核</el-radio-button><el-radio-button value="approved">已公开</el-radio-button><el-radio-button value="rejected">未通过</el-radio-button></el-radio-group>
    <el-table :data="items" empty-text="暂无记录">
      <el-table-column prop="name" :label="kind === 'friends' ? '网站名称' : '昵称'" />
      <el-table-column v-if="kind === 'friends'" prop="url" label="网站地址" />
      <el-table-column v-if="kind === 'comments'" prop="articleId" label="文章 ID" />
      <el-table-column prop="content" :label="kind === 'friends' ? '网站描述' : '内容'" min-width="180" />
      <el-table-column label="操作" width="170"><template #default="{ row }">
        <el-button v-if="status !== 'approved'" link type="primary" :disabled="saving" @click="review(row,'approved')">通过</el-button>
        <el-button v-if="status !== 'rejected'" link type="danger" :disabled="saving" @click="review(row,'rejected')">{{ status === 'approved' ? '撤下' : '拒绝' }}</el-button>
      </template></el-table-column>
    </el-table>
    <el-pagination v-if="total > 20" v-model:current-page="page" :total="total" :page-size="20" layout="prev, pager, next" />
  </AdminPage>
</template>
