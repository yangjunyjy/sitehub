import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAdminResource } from './useAdminResource'
import { getDocumentWorkspace } from '@/api/documents'

export function useDocumentWorkspace() {
  const route = useRoute()
  const router = useRouter()
  const resource = useAdminResource(() =>
    getDocumentWorkspace({
      mode: route.query.mode || undefined,
      version: route.query.version || undefined,
    }),
  )
  const workspace = resource.data
  const mode = computed(() =>
    workspace.value?.modes?.find((item) => item.id === workspace.value.modeId),
  )
  const version = computed(() =>
    workspace.value?.versions?.find((item) => item.id === workspace.value.versionId),
  )
  const groups = computed(() => workspace.value?.groups || [])
  const location = (component, query = {}) =>
    router.page(component, {
      query: { mode: workspace.value?.modeId, version: workspace.value?.versionId, ...query },
    })
  return { ...resource, workspace, mode, version, groups, location }
}
