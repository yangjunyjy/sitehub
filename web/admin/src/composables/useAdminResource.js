import { ref, onBeforeUnmount } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

export function useAdminResource(fetcher) {
  let requestID = 0
  onBeforeUnmount(() => requestID++)
  const data = ref(null),
    loading = ref(false),
    saving = ref(false),
    error = ref('')
  async function load() {
    const current = ++requestID
    loading.value = true
    error.value = ''
    try {
      const result = await fetcher()
      if (current === requestID) data.value = result.data
    } catch (e) {
      if (current === requestID) error.value = e.message || '加载失败，请重试'
    } finally {
      if (current === requestID) loading.value = false
    }
  }
  async function save(action, message = '保存成功') {
    if (saving.value || loading.value || error.value || !data.value) return false
    saving.value = true
    try {
      await action()
      ElMessage.success(message)
      return true
    } catch (e) {
      ElMessage.error(e.message || '保存失败，请重试')
      return false
    } finally {
      saving.value = false
    }
  }
  return { data, loading, saving, error, load, save }
}

export function useUnsavedChanges(isDirty) {
  async function confirmLeave() {
    if (!isDirty()) return true
    try {
      await ElMessageBox.confirm('尚有未保存的修改，离开后将丢失。', '离开当前页面', {
        confirmButtonText: '放弃修改',
        cancelButtonText: '继续编辑',
        type: 'warning',
      })
      return true
    } catch {
      return false
    }
  }
  onBeforeRouteLeave(confirmLeave)
  const beforeUnload = (event) => {
    if (isDirty()) {
      event.preventDefault()
      event.returnValue = ''
    }
  }
  window.addEventListener('beforeunload', beforeUnload)
  onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
  return confirmLeave
}
