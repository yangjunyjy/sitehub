<script setup>
import { ref, computed, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import Login from './views/Login.vue'
import ConsoleLayout from './layout/ConsoleLayout.vue'
import Workspace from './views/Workspace.vue'
import Dialog from './components/Dialog.vue'
import { api } from './api'
const state = ref('loading'),
  failure = ref(''),
  workspace = shallowRef(null),
  draft = ref(null),
  busy = ref(false),
  logoutDialog = ref(false),
  logoutError = ref('')
const dirty = computed(
  () =>
    workspace.value &&
    JSON.stringify(workspace.value.state.manifest) !== JSON.stringify(draft.value),
)
let generation = 0
function signedOut() {
  generation++
  workspace.value = null
  draft.value = null
  state.value = 'login'
  logoutDialog.value = false
  logoutError.value = ''
  failure.value = ''
}
async function load() {
  const current = ++generation
  busy.value = true
  failure.value = ''
  try {
    const result = await api('workspace')
    if (current !== generation) return
    workspace.value = result
    draft.value = structuredClone(result.state.manifest)
    state.value = 'workspace'
  } catch (e) {
    if (current !== generation) return
    if (e.status === 401) signedOut()
    else {
      failure.value = e.message
      if (!workspace.value || e.status === 403) {
        workspace.value = null
        draft.value = null
        state.value = 'error'
      }
    }
  } finally {
    busy.value = false
  }
}
async function logout() {
  if (busy.value) return
  busy.value = true
  logoutError.value = ''
  try {
    await api('auth/logout', {})
    signedOut()
  } catch (e) {
    logoutError.value = e.message
    logoutDialog.value = true
  } finally {
    busy.value = false
  }
}
function leave(e) {
  if (dirty.value) {
    e.preventDefault()
    e.returnValue = ''
  }
}
onMounted(() => {
  window.addEventListener('session-expired', signedOut)
  window.addEventListener('beforeunload', leave)
  load()
})
onBeforeUnmount(() => {
  window.removeEventListener('session-expired', signedOut)
  window.removeEventListener('beforeunload', leave)
})
</script>
<template>
  <Login v-if="state === 'login'" @authenticated="load" />
  <ConsoleLayout
    v-else-if="workspace"
    :actor="workspace.actor"
    :role="workspace.role"
    :busy="busy"
    @logout="logoutDialog = true"
    v-slot="{ section }"
  >
    <p v-if="failure" role="alert" class="error">{{ failure }}</p>
    <Workspace
      :section="section"
      :workspace="workspace"
      :draft="draft"
      :dirty="!!dirty"
      :loading="busy"
      @reload="load"
    />
  </ConsoleLayout>
  <section v-else class="load-panel">
    <strong>网关控制台</strong>
    <p role="status">{{ failure || '正在载入工作区…' }}</p>
    <template v-if="failure"
      ><button @click="load">重新连接</button>
      <button @click="logoutDialog = true">退出登录</button></template
    >
  </section>
  <Dialog
    v-if="logoutDialog"
    title="退出网关控制台"
    confirm="确认退出"
    :busy="busy"
    :error="logoutError"
    @close="logoutDialog = false"
    @submit="logout"
  >
    <p>
      {{
        dirty
          ? '本页尚有未发布的修改。退出会丢弃这些修改并结束当前会话。'
          : '确定退出当前网关账号？'
      }}
    </p>
  </Dialog>
</template>
