import { filterMenuTree } from '@/router/pages'
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserMenus } from '@/api/permission'
import { getUserInfo, getPermissions } from '@/api/auth'
import { useWorkspaceStore } from './workspace'

// 菜单类型枚举
export const MenuType = {
  DIRECTORY: 'directory',  // 目录
  MENU: 'menu',            // 菜单
  BUTTON: 'button'         // 按钮
}

// 本地存储键名
const STORAGE_KEY = 'rbac_menus'
const ROLES_KEY = 'rbac_roles'
const USERS_KEY = 'rbac_users'

// 初始化菜单数据
const initMenus = () => {
  try {
    const stored = sessionStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch (error) {
    console.error('解析菜单数据失败:', error)
    sessionStorage.removeItem(STORAGE_KEY)
  }
  
  // 无默认数据，返回空数组
  return []
}

// 初始化角色数据
const initRoles = () => {
  try {
    const stored = sessionStorage.getItem(ROLES_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch (error) {
    console.error('解析角色数据失败:', error)
    sessionStorage.removeItem(ROLES_KEY)
  }
  
  // 无默认数据，返回空数组
  return []
}

// 初始化用户数据
const initUsers = () => {
  try {
    const stored = sessionStorage.getItem(USERS_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch (error) {
    console.error('解析用户数据失败:', error)
    sessionStorage.removeItem(USERS_KEY)
  }
  
  // 无默认数据，返回空数组
  return []
}

export const useRbacStore = defineStore('rbac', () => {
  // 状态
  const menus = ref(initMenus())
  const roles = ref(initRoles())
  const users = ref(initUsers())
  const currentUser = ref(null)
  const currentPermissions = ref([])
  const currentMenus = ref([])
  // 菜单加载状态管理
  const menusLoaded = ref(false)
  const menuLastFetchTime = ref(0)
  const MENU_CACHE_DURATION = 5 * 60 * 1000 // 5分钟缓存

  // 计算属性
  const menuTreeSelectData = computed(() => {
    const buildTree = (parentId) => {
      return menus.value
        .filter(m => m.parentId === parentId && m.type !== MenuType.BUTTON && m.status === 1)
        .map(m => ({
          ...m,
          children: buildTree(m.id)
        }))
    }
    return buildTree(0)
  })

  // 保存菜单
  const saveMenus = () => {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(menus.value))
  }

  // 保存角色
  const saveRoles = () => {
    sessionStorage.setItem(ROLES_KEY, JSON.stringify(roles.value))
  }

  // 保存用户
  const saveUsers = () => {
    sessionStorage.setItem(USERS_KEY, JSON.stringify(users.value))
  }

  // 获取用户的所有权限
  const getUserPermissions = (user) => {
    if (!user) return []

    // 获取用户的角色ID列表（支持两种格式：roleIds 或 roles 数组）
    const getUserRoleIds = (user) => {
      if (user.roleIds && Array.isArray(user.roleIds)) {
        return user.roleIds
      }
      if (user.roles && Array.isArray(user.roles)) {
        return user.roles.map(role => role.id)
      }
      return []
    }
    
    const userRoleIds = getUserRoleIds(user)
    const userRoles = roles.value.filter(r => userRoleIds.includes(r.id))
    const allPermissions = new Set()

    userRoles.forEach(role => {
      // 如果角色的 menuIds 为空，表示没有权限，跳过该角色
      if (!role.menuIds || role.menuIds.length === 0) {
        return
      }
      role.menuIds.forEach(menuId => {
        const menu = menus.value.find(m => m.id === menuId)
        if (menu && menu.permission) {
          allPermissions.add(menu.permission)
        }
      })
    })

    return Array.from(allPermissions)
  }

  // 获取用户的菜单列表（只包含目录和菜单）
  const getUserMenusList = (user) => {
    if (!user) return []

    // 获取用户的角色ID列表（支持两种格式：roleIds 或 roles 数组）
    const getUserRoleIds = (user) => {
      if (user.roleIds && Array.isArray(user.roleIds)) {
        return user.roleIds
      }
      if (user.roles && Array.isArray(user.roles)) {
        return user.roles.map(role => role.id)
      }
      return []
    }


    const allowedMenus = menus.value.filter(m =>
      m.status === 1 &&
      m.isVisible === 1 &&
      (m.type === MenuType.DIRECTORY || m.type === MenuType.MENU)
    )
    // 构建树形结构
    const buildTree = (parentId) => {
      return allowedMenus
        .filter(m => m.parentId === parentId)
        .sort((a, b) => a.sort - b.sort)
        .map(m => ({
          ...m,
          children: buildTree(m.id)
        }))
    }

    return buildTree(0)
  }

  // 获取用户的所有菜单路径（扁平化，用于路由）
  const getUserMenuPaths = (user) => {
    if (!user) return []

    // 获取用户的角色ID列表（支持两种格式：roleIds 或 roles 数组）
    const getUserRoleIds = (user) => {
      if (user.roleIds && Array.isArray(user.roleIds)) {
        return user.roleIds
      }
      if (user.roles && Array.isArray(user.roles)) {
        return user.roles.map(role => role.id)
      }
      return []
    }
    
    const userRoleIds = getUserRoleIds(user)
    const userRoles = roles.value.filter(r => userRoleIds.includes(r.id))
    const allowedMenuIds = new Set()

    userRoles.forEach(role => {
      // 如果角色的 menuIds 为空，表示没有权限，跳过该角色
      if (!role.menuIds || role.menuIds.length === 0) {
        return
      }
      role.menuIds.forEach(menuId => {
        allowedMenuIds.add(menuId)
      })
    })

    // 返回所有有路径的菜单（用于路由）
    return menus.value
      .filter(m => allowedMenuIds.has(m.id) && m.status === 1 && m.path && m.type !== MenuType.BUTTON)
      .map(m => m.path)
  }

  // 获取用户的所有菜单对象（用于动态路由生成）
  const getUserMenuObjects = (user) => {
    if (!user) return []

    // 获取用户的角色ID列表（支持两种格式：roleIds 或 roles 数组）
    const getUserRoleIds = (user) => {
      if (user.roleIds && Array.isArray(user.roleIds)) {
        return user.roleIds
      }
      if (user.roles && Array.isArray(user.roles)) {
        return user.roles.map(role => role.id)
      }
      return []
    }
    
    const userRoleIds = getUserRoleIds(user)
    const userRoles = roles.value.filter(r => userRoleIds.includes(r.id))
    const allowedMenuIds = new Set()

    userRoles.forEach(role => {
      // 如果角色的 menuIds 为空，表示没有权限，跳过该角色
      if (!role.menuIds || role.menuIds.length === 0) {
        return
      }
      role.menuIds.forEach(menuId => {
        allowedMenuIds.add(menuId)
      })
    })

    // 返回所有有路径的菜单对象（用于路由生成）
    return menus.value
      .filter(m => allowedMenuIds.has(m.id) && m.status === 1 && m.path && m.type !== MenuType.BUTTON)
  }

  // 检查是否有权限
  const hasPermission = (permission) => {
    return currentPermissions.value.includes('*') || currentPermissions.value.includes(permission) || (currentUser.value?.roleIds || []).some(id => Number(id) === 1) || (currentUser.value?.roles || []).some(role => Number(role.id) === 1)
  }

  let sessionEpoch = 0
  let pendingAccess = null
  const loadAllDataFromBackend = (forceRefresh = false) => {
    if (!forceRefresh && menusLoaded.value && Date.now() - menuLastFetchTime.value < MENU_CACHE_DURATION) {
      return Promise.resolve({ user: currentUser.value, roles: roles.value, menus: currentMenus.value, permissions: currentPermissions.value })
    }
    if (pendingAccess) return pendingAccess
    const epoch = sessionEpoch
    const token = sessionStorage.getItem('token')
    pendingAccess = Promise.all([getUserInfo(), getUserMenus(), getPermissions()])
      .then(([profile, menuList, grants]) => {
        if (epoch !== sessionEpoch || token !== sessionStorage.getItem('token')) throw new Error('登录状态已更改')
        if (!profile.data || !Array.isArray(menuList.data) || !Array.isArray(grants.data?.items)) throw new Error('后台权限数据不完整，请重试')
        const user = profile.data
        const formatted = formatMenusForFrontend(menuList.data)
        currentUser.value = user
        users.value = [user]
        roles.value = user.roles || []
        menus.value = formatted
        currentMenus.value = formatted
        currentPermissions.value = grants.data.items
        menusLoaded.value = true
        menuLastFetchTime.value = Date.now()
        return { user, roles: roles.value, menus: formatted, permissions: currentPermissions.value }
      })
      .finally(() => { if (epoch === sessionEpoch) pendingAccess = null })
    return pendingAccess
  }

  const loadUserMenusFromBackend = async (forceRefresh = false) => {
    const access = await loadAllDataFromBackend(forceRefresh)
    return access.menus
  }

  // 格式化后端菜单数据为前端格式
  const formatMenusForFrontend = (backendMenus) => {
    const formatMenu = (menu) => {

      const children = (menu.children || []).map(formatMenu).filter(Boolean)
      const formatted = {
        id: menu.id,
        parentId: menu.parent_id || 0,
        title: menu.menu_name,
        type: menu.menu_type === 1 ? MenuType.DIRECTORY :
              menu.menu_type === 2 ? MenuType.MENU :
              MenuType.BUTTON,
        icon: menu.icon,
        path: menu.path,
        component: menu.component,
        pageType: menu.pageType || 'internal',
        externalUrl: menu.externalUrl,
        permission: menu.perms || menu.permission || '',
        menuCode: menu.menu_code,
        sort: menu.sort,
        status: menu.status,
        isVisible: menu.is_visible,
        children
      }

      if (formatted.type === MenuType.DIRECTORY && !formatted.component && children.length === 0) return null
      return formatted
    }
    return filterMenuTree(backendMenus.map(formatMenu).filter(Boolean))
  }

  // 提取菜单权限标识
  const extractMenuPermissions = (menus) => {
    const permissions = []
    const traverse = (menuList) => {
      menuList.forEach(menu => {
        if (menu.permission) {
          permissions.push(menu.permission)
        }
        if (menu.children && menu.children.length > 0) {
          traverse(menu.children)
        }
      })
    }
    traverse(menus)
    return permissions
  }

  // 初始化当前用户权限（返回Promise便于调用者等待）
  const initUserPermission = async (user) => {
    currentUser.value = user
    
    // 从后端加载菜单，强制刷新以确保获取最新数据
    return await loadUserMenusFromBackend(true)
  }

  // 清除权限
  const clearPermission = () => {
    sessionEpoch++
    pendingAccess = null
    useWorkspaceStore().reset()
    currentUser.value = null
    currentPermissions.value = []
    currentMenus.value = []
    menusLoaded.value = false
    menuLastFetchTime.value = 0
  }

  // 重置所有数据到默认值
  const resetAllData = () => {
    sessionEpoch++
    pendingAccess = null
    useWorkspaceStore().reset()

    // 清除 sessionStorage 中的缓存
    sessionStorage.removeItem(STORAGE_KEY)
    sessionStorage.removeItem(ROLES_KEY)
    sessionStorage.removeItem(USERS_KEY)

    // 直接设置为空数组，不要调用 init*() 方法，避免从 sessionStorage 重新读取
    menus.value = []
    currentMenus.value = []
    roles.value = []
    users.value = []
    currentUser.value = null
    currentPermissions.value = []

    // 重置菜单加载状态
    menusLoaded.value = false
    menuLastFetchTime.value = 0

  }

  // 重新初始化菜单（用于菜单结构更新后）
  const reinitMenus = () => {
    sessionStorage.removeItem(STORAGE_KEY)
    menus.value = initMenus()
  }

  // 刷新当前用户权限（用于角色更新后）
  const refreshUserPermission = async (user = null) => {
    const targetUser = user || currentUser.value
    if (targetUser) {
      return await initUserPermission(targetUser)
    }
    return []
  }

  // 手动刷新用户菜单（供外部调用）
  const refreshUserMenus = async () => {
    return await loadUserMenusFromBackend(true)
  }

  // 返回所有状态、计算属性和方法
  return {
    // 常量
    MenuType,
    
    // 状态
    menus,
    roles,
    users,
    currentUser,
    currentPermissions,
    currentMenus,
    menusLoaded,
    menuLastFetchTime,
    
    // 计算属性
    menuTreeSelectData,
    
    // 方法
    saveMenus,
    saveRoles,
    saveUsers,
    hasPermission,
    initUserPermission,
    clearPermission,
    getUserMenuPaths,
    getUserMenuObjects,
    resetAllData,
    reinitMenus,
    loadUserMenusFromBackend,
    refreshUserPermission,
    refreshUserMenus,
    getUserPermissions,
    getUserMenusList,
    loadAllDataFromBackend,
  }
})

