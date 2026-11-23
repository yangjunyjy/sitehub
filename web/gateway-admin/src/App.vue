<script setup>
import { ref, computed, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import Login from './views/Login.vue'
import ConsoleLayout from './layout/ConsoleLayout.vue'
import Workspace from './views/Workspace.vue'
import Dialog from './components/Dialog.vue'
import { api } from './api'
import { changedResources, fieldChanges, rebaseDraft, replaceResource } from './manifest-changes'
import FieldChanges from './components/FieldChanges.vue'
const state = ref('loading'),
  failure = ref(''),
  workspace = shallowRef(null),
  draft = ref(null),
  busy = ref(false),
  logoutDialog = ref(false),
  logoutError = ref('')
const dirty = computed(
  () => workspace.value && changedResources(workspace.value.state.manifest, draft.value).length > 0,
)
const conflicts = ref([]),
  notice = ref('')
// Keep unpublished manifests in this tab's memory only. Clear on explicit
// logout or a different account signing in, never persist credentials.
const suspended = shallowRef(null)
function captureDraft() {
  if (workspace.value?.state.desired_unavailable) return suspended.value
  if (!workspace.value || (!dirty.value && !conflicts.value.length)) return null
  const local = JSON.parse(JSON.stringify(draft.value))
  const base = structuredClone(workspace.value.state.manifest)
  for (const conflict of conflicts.value) {
    replaceResource(base, conflict.key, conflict.base)
    replaceResource(local, conflict.key, conflict.local)
  }
  return { actor: workspace.value.actor, revision: workspace.value.state.revision, base, local }
}
function expired() {
  if (workspace.value) suspended.value = captureDraft()
  signedOut()
}
let generation = 0
function signedOut() {
  generation++
  workspace.value = null
  draft.value = null
  state.value = 'login'
  logoutDialog.value = false
  logoutError.value = ''
  failure.value = ''
  notice.value = ''
  conflicts.value = []
}
async function load(preserve = false) {
  const saved = preserve === true ? captureDraft() : suspended.value
  const current = ++generation
  busy.value = true
  failure.value = ''
  try {
    const result = await api('workspace?history=0')
    if (current !== generation) return
    workspace.value = result
    draft.value = structuredClone(result.state.manifest)
    conflicts.value = []
    notice.value = ''
    suspended.value = null
    if (saved && result.actor === saved.actor && result.state.desired_unavailable) {
      suspended.value = saved
      notice.value = '当前账号的草稿已暂存。配置源恢复后重新载入，即可继续审阅。'
    } else if (saved && result.actor === saved.actor && result.publish) {
      const restored = rebaseDraft(saved.base, saved.local, result.state.manifest)
      draft.value = restored.merged
      conflicts.value = restored.conflicts
      notice.value =
        saved.revision === result.state.revision
          ? '已恢复当前账号的草稿。'
          : '已载入最新修订并合并独立变更，请重新审阅草稿。'
    }
    state.value = 'workspace'
  } catch (e) {
    if (current !== generation) return
    if (e.status === 401) expired()
    else {
      failure.value = e.message
      if (!workspace.value || e.status === 403) {
        workspace.value = null
        draft.value = null
        state.value = 'error'
      }
    }
  } finally {
    if (current === generation || !workspace.value) busy.value = false
  }
}
function resolveConflicts() {
  for (const conflict of conflicts.value) {
    if (conflict.choice === 'local') replaceResource(draft.value, conflict.key, conflict.local)
  }
  conflicts.value = []
  notice.value = '冲突已处理，请核对草稿并重新审阅。'
}
async function logout() {
  if (busy.value) return
  busy.value = true
  logoutError.value = ''
  try {
    await api('auth/logout', {})
    suspended.value = null
    signedOut()
  } catch (e) {
    logoutError.value = e.message
    logoutDialog.value = true
  } finally {
    busy.value = false
  }
}
function leave(e) {
  if (dirty.value || suspended.value || conflicts.value.length) {
    e.preventDefault()
    e.returnValue = ''
  }
}
onMounted(() => {
  window.addEventListener('session-expired', expired)
  window.addEventListener('beforeunload', leave)
  load()
})
onBeforeUnmount(() => {
  window.removeEventListener('session-expired', expired)
  window.removeEventListener('beforeunload', leave)
})
</script>
<template>
  <Login v-if="state === 'login'" @authenticated="load()" />
  <ConsoleLayout
    v-else-if="workspace"
    :actor="workspace.actor"
    :role="workspace.role"
    :busy="busy"
    @logout="logoutDialog = true"
    v-slot="{ section }"
  >
    <p v-if="failure" role="alert" class="error">{{ failure }}</p>
    <p v-if="notice" role="status" class="mode-note">{{ notice }}</p>
    <section v-if="conflicts.length" class="card conflict-panel">
      <h2>处理并发修改</h2>
      <p>下列资源同时被修改。逐项选择要保留的整份资源后继续编辑。</p>
      <article v-for="conflict in conflicts" :key="conflict.key" class="change">
        <h3>{{ conflict.key }}</h3>
        <FieldChanges
          :changes="fieldChanges(conflict.current, conflict.local)"
          left="最新配置"
          right="我的草稿"
        />
        <label
          >采用的版本<select aria-label="采用的版本" v-model="conflict.choice">
            <option value="">请选择</option>
            <option value="remote">最新配置</option>
            <option value="local">我的草稿</option>
          </select></label
        >
      </article>
      <button
        class="primary"
        :disabled="conflicts.some((c) => !c.choice)"
        @click="resolveConflicts"
      >
        应用选择
      </button>
    </section>
    <Workspace
      v-else
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
        dirty || suspended || conflicts.length
          ? '本页尚有未发布的修改。退出会丢弃这些修改并结束当前会话。'
          : '确定退出当前网关账号？'
      }}
    </p>
  </Dialog>
</template>
