import { createRouter, createWebHistory } from 'vue-router'
import { useRbacStore } from '../store/rbac'
import { storeToRefs } from 'pinia'
import { filterMenuTree, normalizeComponent, pageLoaders } from './pages'
import { consoleRoutes, isSuperAdmin } from './console'

const fixedPaths = new Set([
  '/dashboard/index',
  '/dashboard/website',
  '/dashboard/downloads',
  '/dashboard/profile',
  '/dashboard/articles/create',
  '/dashboard/articles/editor/:id',
  '/create',
  '/editor/:id',
])
consoleRoutes.forEach((route) => fixedPaths.add(`/dashboard/${route.path}`))

function generateRoutesFromMenus(menus) {
  const routes = []
  const visit = (items, parentPath = '') => {
    for (const menu of items) {
      if (menu.type === 'button') continue
      const path = menu.path?.startsWith('/')
        ? menu.path
        : [parentPath, menu.path].filter(Boolean).join('/')
      if (menu.type === 'directory') {
        visit(menu.children || [], path)
        continue
      }
      const componentName = normalizeComponent(menu.component)
      const component = pageLoaders[componentName]
      if (component && path?.startsWith('/') && !fixedPaths.has(path)) {
        routes.push({
          path,
          name: `menu-${menu.id}`,
          component,
          meta: {
            title: menu.title,
            icon: menu.icon,
            permission: menu.permission,
            menuCode: menu.menuCode,
            superAdmin: componentName === 'views/settings/EmailSettings.vue',
          },
        })
      }
      visit(menu.children || [], path)
    }
  }
  visit(filterMenuTree(menus))
  return routes
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    redirect: '/dashboard/index', // 重定向到子路由
    component: () => import('../layout/Layout.vue'),
    meta: { hidden: true }, // 隐藏，不显示在菜单/标签中
    children: [
      {
        path: 'articles/create',
        name: 'ArticleCreate',
        component: pageLoaders['views/articles/ArticleEditor.vue'],
        meta: {
          activeMenu: '/dashboard/articles',
          title: '新建文章',
          permission: 'blog:article:create',
        },
      },
      {
        path: 'articles/editor/:id',
        name: 'ArticleEdit',
        component: pageLoaders['views/articles/ArticleEditor.vue'],
        meta: {
          activeMenu: '/dashboard/articles',
          title: '编辑文章',
          permissionsAny: [
            'blog:article:update',
            'blog:article:select',
            'blog:article:list',
          ],
        },
      },

      ...consoleRoutes,
      {
        path: 'index',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '工作台' },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/profile/Profile.vue'),
        meta: { title: '个人中心' },
      },
      {
        path: 'website',
        name: 'WebsiteContentCenter',
        beforeEnter: (to) =>
          ({
            docs: '/dashboard/docs',
            storage: '/dashboard/settings/storage',
            downloads: '/dashboard/downloads',
          })[to.query.tab] || true,
        component: () => import('../views/website/ContentCenter.vue'),
        meta: { title: '官网内容中心' },
      },
      {
        path: 'downloads',
        name: 'DownloadCenter',
        component: () => import('../views/website/DownloadCenter.vue'),
        meta: { title: '下载管理' },
      },
    ],
  },
  { path: '/create', redirect: '/dashboard/articles/create' },
  { path: '/editor/:id', redirect: (to) => `/dashboard/articles/editor/${to.params.id}` },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('../views/404.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})
// Vue Router returns removal callbacks; custom fields are not retained in
// normalized route records, so keep these callbacks for logout/menu refresh.
let dynamicRouteRemovers = []
export function addRoutes(menus) {
  for (const route of generateRoutesFromMenus(menus || [])) {
    dynamicRouteRemovers.push(router.addRoute('dashboard', route))
  }
}

export function resetRoutes() {
  for (const remove of dynamicRouteRemovers) remove()
  dynamicRouteRemovers = []
}

let lastMenus = []
let routesInitialized = false

// 检查是否需要重新初始化路由
const shouldReinitRoutes = () => {
  const flag = sessionStorage.getItem('routeInitialized')
  return flag === 'false' || flag === null
}

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  document.title = to.meta.title ? `${to.meta.title} · SiteHub 内容中心` : 'SiteHub 内容中心'
  const token = sessionStorage.getItem('token')
  const whiteList = ['/login']

  if (whiteList.includes(to.path)) {
    if (token) {
      next()
    } else {
      next()
    }
    return
  }

  if (!token) {
    next('/login')
    return
  }

  // 页面刷新时初始化TOTP状态（仅首次）
  if (!window.__totpInitialized) {
    window.__totpInitialized = true
    import('@/utils/request').then(({ initRequestInterceptors }) => {
      initRequestInterceptors().catch(() => {})
    })
  }

  try {
    const userInfo = JSON.parse(sessionStorage.getItem('userInfo') || '{}')
    // 验证用户信息是否有email
    if (!userInfo.email) {
      console.error('用户信息不完整，跳转到登录页')
      sessionStorage.clear()
      next('/login')
      return
    }

    const rbacStore = useRbacStore()
    const { menus, menusLoaded } = storeToRefs(rbacStore)

    // 检查是否需要重新初始化路由
    if (shouldReinitRoutes()) {
      routesInitialized = false
      sessionStorage.setItem('routeInitialized', 'true')
    }

    // 首次加载或菜单未加载时，从后端获取菜单并初始化路由
    if (!routesInitialized || !menusLoaded.value) {

      try {
        // 如果菜单未加载，先加载菜单
        if (!menusLoaded.value) {
          await rbacStore.loadAllDataFromBackend()
        }

        // 清除旧的动态路由
        resetRoutes()

        // 根据菜单生成并添加新路由
        addRoutes(menus.value)

        // 记录当前菜单数据
        lastMenus = JSON.parse(JSON.stringify(menus.value))
        routesInitialized = true

        // 首次加载时，总是重新导航以确保动态路由生效
        next({ ...to, replace: true })
        return
      } catch (error) {
        console.error('路由初始化失败:', error)
        sessionStorage.clear()
        next('/login')
        return
      }
    }

    // 检查菜单是否发生变化
    const menusChanged = JSON.stringify(menus.value) !== JSON.stringify(lastMenus)

    // 菜单变化时，重新加载路由
    if (menusChanged) {

      // 清除旧的动态路由
      resetRoutes()

      // 根据菜单生成并添加新路由
      addRoutes(menus.value)

      // 记录当前菜单数据
      lastMenus = JSON.parse(JSON.stringify(menus.value))
      // Resolve the destination again after edited paths have been registered.
      next({ ...to, replace: true })
      return
    }
  } catch (error) {
    console.error('路由初始化失败:', error)
  }

  if (to.meta.superAdmin && !isSuperAdmin(useRbacStore().currentUser)) {
    next('/dashboard/index')
    return
  }
  if (to.meta.permission && !useRbacStore().hasPermission(to.meta.permission)) {
    next('/dashboard/index')
    return
  }
  if (
    to.meta.permissionsAny &&
    !to.meta.permissionsAny.some((permission) => useRbacStore().hasPermission(permission))
  ) {
    next('/dashboard/index')
    return
  }
  next()
})

export default router
