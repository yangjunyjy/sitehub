import { defineStore } from 'pinia'
import { isSuperAdmin } from '@/router/pages'
import { ref } from 'vue'
import { getUserInfo, getPermissions, getUserMenus } from '@/api/auth'
import { useWorkspaceStore } from './workspace'

export const useRbacStore = defineStore('rbac', () => {
  const currentUser = ref(null)
  const currentPermissions = ref([])
  const currentMenus = ref([])
  const menusLoaded = ref(false)
  const menuLastFetchTime = ref(0)
  const MENU_CACHE_DURATION = 5 * 60 * 1000
  // 检查是否有权限
  const hasPermission = (permission) => {
    return currentPermissions.value.includes('*') || currentPermissions.value.includes(permission) || isSuperAdmin(currentUser.value)
  }

  let sessionEpoch = 0
  let pendingAccess = null
  const loadAllDataFromBackend = (forceRefresh = false) => {
    if (!forceRefresh && menusLoaded.value && Date.now() - menuLastFetchTime.value < MENU_CACHE_DURATION) {
      return Promise.resolve({ user: currentUser.value, roles: currentUser.value?.roles || [], menus: currentMenus.value, permissions: currentPermissions.value })
    }
    if (pendingAccess) return pendingAccess
    const epoch = sessionEpoch
    const token = sessionStorage.getItem('token')
    pendingAccess = Promise.all([getUserInfo(), getUserMenus(), getPermissions()])
      .then(([profile, menuList, grants]) => {
        if (epoch !== sessionEpoch || token !== sessionStorage.getItem('token')) throw new Error('登录状态已更改')
        if (!profile.data || !Array.isArray(menuList.data) || !Array.isArray(grants.data?.items)) throw new Error('后台权限数据不完整，请重试')
        const user = profile.data
        currentUser.value = user
        currentMenus.value = menuList.data
        currentPermissions.value = grants.data.items
        menusLoaded.value = true
        menuLastFetchTime.value = Date.now()
        return { user, roles: currentUser.value?.roles || [], menus: menuList.data, permissions: currentPermissions.value }
      })
      .finally(() => { if (epoch === sessionEpoch) pendingAccess = null })
    return pendingAccess
  }

  const loadUserMenusFromBackend = async (forceRefresh = false) => {
    const access = await loadAllDataFromBackend(forceRefresh)
    return access.menus
  }

  function resetAllData() {
    sessionEpoch++
    pendingAccess = null
    useWorkspaceStore().reset()
    currentUser.value = null
    currentPermissions.value = []
    currentMenus.value = []
    menusLoaded.value = false
    menuLastFetchTime.value = 0
  }
  const refreshUserMenus = () => loadUserMenusFromBackend(true)
  return { currentUser, currentPermissions, currentMenus, menusLoaded, hasPermission,
    resetAllData, loadAllDataFromBackend, refreshUserMenus, refreshUserPermission: refreshUserMenus }
})
