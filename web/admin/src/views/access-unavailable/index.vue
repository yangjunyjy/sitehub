<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
const route = useRoute(), router = useRouter(), busy = ref(false)
const destination = computed(() => {
  const path = route.query.redirect
  return typeof path === 'string' && /^\/(?!\/)/.test(path) && !path.startsWith('/access-unavailable') ? path : '/dashboard'
})
async function retry() {
  busy.value = true
  try { await router.replace(destination.value) } finally { busy.value = false }
}
</script>
<template>
  <main class="access-error"><el-result icon="warning" title="后台暂时无法加载" sub-title="读取账号或权限数据失败，请重试。">
    <template #extra><el-button type="primary" :loading="busy" @click="retry">重新加载</el-button><el-button @click="router.replace('/login')">返回登录</el-button></template>
  </el-result></main>
</template>
<style scoped lang="scss">.access-error { min-height: 100svh; display: grid; place-items: center; }</style>
