import { computed, nextTick, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { isWorkspaceRoute, workspacePath } from '../utils/workspace-route.js'

const cacheable = new Set([
  'Dashboard',
  'ArticleList',
  'CategoryList',
  'TagList',
  'UserList',
  'RoleList',
  'MenuList',
  'Content',
])
const home = () => ({
  path: '/dashboard/index',
  fullPath: '/dashboard/index',
  title: '工作台',
  componentName: 'Dashboard',
  affix: true,
})

// Navigation and KeepAlive share one source of truth. Editors retain their
// existing navigation guards and are deliberately outside the cache whitelist.
export const useWorkspaceStore = defineStore('workspace', () => {
  const views = ref([home()])
  const revisions = ref({})
  const refreshing = ref('')
  let sessionEpoch = 0
  let storageKey = ''
  watch(
    views,
    () => {
      if (!storageKey) return
      try {
        sessionStorage.setItem(storageKey, JSON.stringify(views.value.map((view) => view.fullPath)))
      } catch {
        /* Restricted storage keeps an in-memory workspace. */
      }
    },
    { deep: true, flush: 'sync' },
  )
  const cachedNames = computed(() => [
    ...new Set(
      views.value
        .map((view) => view.componentName)
        .filter((name) => cacheable.has(name) && name !== refreshing.value),
    ),
  ])
  function visit(route) {
    if (!isWorkspaceRoute(route)) return
    const path = workspacePath(route)
    const component = route.matched.at(-1)?.components?.default
    const previous = views.value.find((view) => view.path === path)
    const view = {
      path,
      fullPath: route.fullPath,
      title:
        route.path.startsWith('/dashboard/docs') && route.query.version
          ? `${route.meta.title} · ${route.query.version}`
          : route.meta.title,
      componentName: component?.name || component?.__name,
      affix: path === '/dashboard/index',
    }
    if (previous) Object.assign(previous, view)
    else views.value.push(view)
  }
  function remove(path) {
    views.value = views.value.filter((view) => view.affix || view.path !== path)
    delete revisions.value[path]
  }
  function keepOnly(path) {
    retain((view) => view.path === path)
  }
  function retain(canOpen) {
    views.value = views.value.filter((view) => view.affix || canOpen(view))
    for (const key of Object.keys(revisions.value)) {
      if (!views.value.some((view) => view.path === key)) delete revisions.value[key]
    }
  }
  function canRefresh(path) {
    return cacheable.has(views.value.find((view) => view.path === path)?.componentName)
  }
  function reset() {
    sessionEpoch++
    try {
      if (storageKey) sessionStorage.removeItem(storageKey)
    } catch {}
    storageKey = ''
    views.value = [home()]
    revisions.value = {}
    refreshing.value = ''
  }
  function restore(userId, resolve, canOpen) {
    const nextKey = userId ? `deploy-admin-tabs:${userId}` : ''
    if (nextKey === storageKey) return
    reset()
    if (!nextKey) return
    let paths = []
    try {
      const saved = JSON.parse(sessionStorage.getItem(nextKey) || '[]')
      if (Array.isArray(saved)) paths = saved.slice(0, 40)
    } catch {}
    // Resolve against today's menu and permissions; never trust saved titles or components.
    for (const path of paths) {
      if (typeof path !== 'string' || !path.startsWith('/') || path.startsWith('//')) continue
      const route = resolve(path)
      if (isWorkspaceRoute(route) && canOpen(route)) visit(route)
    }
    storageKey = nextKey
  }
  async function refresh(path) {
    if (!canRefresh(path) || refreshing.value) return
    const epoch = sessionEpoch
    refreshing.value = views.value.find((view) => view.path === path).componentName
    await nextTick()
    if (epoch !== sessionEpoch) return
    if (!canRefresh(path)) {
      refreshing.value = ''
      return
    }
    revisions.value[path] = (revisions.value[path] || 0) + 1
    await nextTick()
    if (epoch === sessionEpoch) refreshing.value = ''
  }
  return {
    views,
    revisions,
    cachedNames,
    visit,
    remove,
    keepOnly,
    retain,
    canRefresh,
    refresh,
    reset,
    restore,
  }
})
