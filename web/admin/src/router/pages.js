// Only pages backed by the current API may be resolved from database menus.
export const pageLoaders = {
  'views/docs/DocumentModes.vue': () => import('../views/docs/DocumentModes.vue'),
  'views/docs/DocumentList.vue': () => import('../views/docs/DocumentList.vue'),
  'views/docs/DocumentGroups.vue': () => import('../views/docs/DocumentGroups.vue'),
  'views/docs/DocumentEditor.vue': () => import('../views/docs/DocumentEditor.vue'),
  'views/docs/DocumentSettings.vue': () => import('../views/docs/DocumentSettings.vue'),
  'views/docs/DocumentVersions.vue': () => import('../views/docs/DocumentVersions.vue'),
  'views/settings/BasicSettings.vue': () => import('../views/settings/BasicSettings.vue'),
  'views/settings/StorageSettings.vue': () => import('../views/settings/StorageSettings.vue'),
  'views/settings/EmailSettings.vue': () => import('../views/settings/EmailSettings.vue'),
  'views/Dashboard.vue': () => import('../views/Dashboard.vue'),
  'views/website/ContentCenter.vue': () => import('../views/website/ContentCenter.vue'),
  'views/website/DownloadCenter.vue': () => import('../views/website/DownloadCenter.vue'),
  'views/articles/ArticleList.vue': () => import('../views/articles/ArticleList.vue'),
  'views/articles/ArticleEditor.vue': () => import('../views/articles/ArticleEditor.vue'),
  'views/categories/CategoryList.vue': () => import('../views/categories/CategoryList.vue'),
  'views/tags/TagList.vue': () => import('../views/tags/TagList.vue'),
  'views/statistics/Content.vue': () => import('../views/statistics/Content.vue'),
  'views/users/UserList.vue': () => import('../views/users/UserList.vue'),
  'views/roles/RoleList.vue': () => import('../views/roles/RoleList.vue'),
  'views/menus/MenuList.vue': () => import('../views/menus/MenuList.vue'),
  'views/profile/Profile.vue': () => import('../views/profile/Profile.vue'),
}

// Sidebar and page search share the persisted menu titles, hierarchy, order and
// visibility. Email settings retain their existing administrator-only boundary.
export function navigationMenus(menus = [], superAdmin = false, parentPath = '') {
  return filterMenuTree(menus, true)
    .flatMap((menu) => {
      if (!superAdmin && normalizeComponent(menu.component) === 'views/settings/EmailSettings.vue')
        return []
      const path = menu.path?.startsWith('/')
        ? menu.path
        : [parentPath, menu.path].filter(Boolean).join('/')
      const children = navigationMenus(menu.children || [], superAdmin, path)
      if (menu.type === 'directory' && !children.length) return []
      return [{ ...menu, path, children }]
    })
    .sort((a, b) => (Number(a.sort) || 0) - (Number(b.sort) || 0))
}

export function normalizeComponent(value = '') {
  if (typeof value !== 'string') return ''
  let component = value
    .trim()
    .replace(/^\/?src\//, '')
    .replace(/^\//, '')
  if (!component) return ''
  if (!component.startsWith('views/')) component = `views/${component}`
  if (!component.endsWith('.vue')) component += '.vue'
  return component
}

// Directory paths may be chosen by administrators. Keep valid children without
// mutating the store while preparing sidebar navigation.
export function filterMenuTree(menus = [], visibleOnly = false) {
  return menus.flatMap((menu) => {
    if (menu.status !== undefined && Number(menu.status) !== 1) return []
    if (visibleOnly && !menu.isVisible) return []
    if (menu.type === 'button') return visibleOnly ? [] : [{ ...menu }]
    const children = filterMenuTree(menu.children || [], visibleOnly)
    if (menu.type === 'directory') return children.length ? [{ ...menu, children }] : []
    if (menu.type !== 'menu' || !pageLoaders[normalizeComponent(menu.component)]) return []
    return [{ ...menu, children }]
  })
}
