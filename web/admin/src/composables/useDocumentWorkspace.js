import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminResource } from './useAdminResource'
import { getDocumentWorkspace } from '@/api/documents'

export function useDocumentWorkspace() {
  const route = useRoute(),
    router = useRouter()
  const resource = useAdminResource(async () => {
    const result = await getDocumentWorkspace()
    if (
      !Array.isArray(result.data?.versions) ||
      !result.data.defaultVersion ||
      !result.data.revision
    )
      throw new Error('文档接口格式不匹配，请完成数据库迁移并重启后端')
    return result
  })
  const manifest = computed(() => {
    const data = resource.data.value
    if (!data) return data
    return {
      ...data,
      versions: data.versions.map((version) => ({
        ...version,
        modeLabel: data.modes?.find((mode) => mode.id === version.modeId)?.label || '',
      })),
    }
  })
  const versionID = computed(() =>
    String(route.query.version || manifest.value?.defaultVersion || ''),
  )
  const version = computed(() =>
    manifest.value?.versions.find((item) => item.id === versionID.value),
  )
  const location = (path, query = {}) => ({ path, query: { version: versionID.value, ...query } })
  const selectVersion = (id) => router.replace(location(route.path, { version: id }))
  async function mutate(action, payload) {
    await action(payload)
    await resource.load()
    if (resource.error.value) throw new Error('操作已保存，但列表刷新失败，请重新加载')
  }
  return { ...resource, manifest, version, versionID, location, selectVersion, mutate }
}
