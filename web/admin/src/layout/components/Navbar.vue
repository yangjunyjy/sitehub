<template>
  <header class="navbar">
    <div class="navbar-left">
      <button
        class="nav-icon collapse-btn"
        type="button"
        :aria-label="isCollapse ? '展开侧边栏' : '收起侧边栏'"
        @click="$emit('toggle-collapse')"
      >
        <el-icon><Fold v-if="!isCollapse" /><Expand v-else /></el-icon>
      </button>
      <el-breadcrumb class="workspace-breadcrumb" separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-if="route.meta.componentPath !== 'dashboard/index'">{{
          route.meta.title
        }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="navbar-right">
      <el-popover
        v-model:visible="searchOpen"
        placement="bottom-end"
        :width="360"
        trigger="click"
        popper-class="admin-global-search"
        @show="searchInput?.focus()"
        ><template #reference
          ><button class="nav-icon" type="button" aria-label="搜索管理功能">
            <el-icon><Search /></el-icon></button
        ></template>
        <div class="console-search">
          <el-autocomplete
            ref="searchInput"
            v-model="search"
            :fetch-suggestions="suggestions"
            :trigger-on-focus="true"
            :prefix-icon="Search"
            placeholder="搜索文章、文档、功能页面…"
            aria-label="搜索管理功能"
            clearable
            @select="navigate"
            @keyup.enter="searchArticles"
            ><template #default="{ item }"
              ><div class="search-suggestion">
                <span>{{ item.value }}</span
                ><small>{{ item.section }}</small>
              </div></template
            >
          </el-autocomplete>
        </div></el-popover
      >
      <el-popover placement="bottom-end" :width="280" trigger="click" @show="loadPending"
        ><template #reference
          ><button class="nav-icon" type="button" aria-label="查看内容待办">
            <el-icon><Bell /></el-icon><i v-if="draftCount > 0" class="notification-dot" /></button
        ></template>
        <div class="notification-panel">
          <h3>内容待办</h3>
          <p v-if="pendingError">读取失败，请稍后重试</p>
          <p v-else-if="pendingLoading">正在读取待办…</p>
          <template v-else
            ><p>
              {{
                draftCount > 0
                  ? `还有 ${draftCount} 篇文章草稿等待完善`
                  : '目前没有待发布的文章草稿'
              }}
            </p>
            <el-button link type="primary" @click="router.push(router.page('blog/articles/index', { query: { status: 0 } }))"
              >查看文章草稿 →</el-button
            ></template
          >
        </div></el-popover
      >
      <button
        class="nav-icon fullscreen-btn"
        type="button"
        :aria-label="fullscreen ? '退出全屏' : '进入全屏'"
        @click="toggleFullscreen"
      >
        <el-icon><FullScreen /></el-icon>
      </button>
      <el-dropdown @command="handleCommand" trigger="click"
        ><button class="user-info" type="button" aria-label="用户菜单">
          <el-avatar :size="26" :src="userInfo.avatar" class="user-avatar">{{
            (userInfo.nickname || userInfo.username || 'A').slice(0, 1).toUpperCase()
          }}</el-avatar
          ><span class="user-copy"
            ><strong>{{ userInfo.nickname || userInfo.username || '管理员' }}</strong></span
          ><el-icon class="dropdown-icon"><ArrowDown /></el-icon></button
        ><template #dropdown
          ><el-dropdown-menu
            ><el-dropdown-item command="profile"
              ><el-icon><User /></el-icon>个人中心</el-dropdown-item
            ><el-dropdown-item command="website"
              ><el-icon><Monitor /></el-icon>访问官网</el-dropdown-item
            ><el-dropdown-item divided command="logout"
              ><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item
            ></el-dropdown-menu
          ></template
        ></el-dropdown
      >
      <button
        class="nav-icon"
        type="button"
        aria-label="个人设置"
        @click="router.push(router.page('profile/index'))"
      >
        <el-icon><Setting /></el-icon>
      </button>
    </div>
  </header>
</template>
<script setup>
import { computed, ref, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  Fold,
  Expand,
  Search,
  Bell,
  FullScreen,
  User,
  Monitor,
  SwitchButton,
  ArrowDown,
  Setting,
} from '@element-plus/icons-vue'
import { logout } from '@/api/auth'
import { getArticleStats } from '@/api/statistics'
import { useRbacStore } from '@/store/rbac'
import { navigationMenus } from '@/router/pages'
import { isSuperAdmin } from '@/router/pages'
import { previewWebsite } from '@/utils/website'
defineProps({ isCollapse: Boolean })
defineEmits(['toggle-collapse'])
const route = useRoute()
const searchOpen = ref(false),
  searchInput = ref(null)
const router = useRouter(),
  rbac = useRbacStore(),
  search = ref(''),
  fullscreen = ref(false)
const userInfo = computed(() => rbac.currentUser || {})
const draftCount = ref(0),
  pendingLoading = ref(false),
  pendingError = ref(false)
const pages = computed(() => {
  const result = []
  const visit = (menus) =>
    menus.forEach((menu) => {
      if (menu.type === 'menu')
        result.push({ value: menu.title, path: menu.path, section: '管理页面' })
      if (menu.children) visit(menu.children)
    })
  visit(navigationMenus(rbac.currentMenus, isSuperAdmin(userInfo.value)))
  return [...new Map(result.map((item) => [item.path, item])).values()]
})
function suggestions(query, callback) {
  const value = query.trim().toLowerCase()
  const list = pages.value.filter((item) => item.value.toLowerCase().includes(value))
  if (value)
    list.push({ value: `搜索文章：${query.trim()}`, keyword: query.trim(), section: '文章检索' })
  callback(list.slice(0, 10))
}
function navigate(item) {
  if (item.keyword) router.push(router.page('blog/articles/index', { query: { keyword: item.keyword } }))
  else router.push(item.path)
  search.value = ''
  searchOpen.value = false
}
function searchArticles() {
  if (search.value.trim()) navigate({ keyword: search.value.trim() })
}
async function loadPending() {
  pendingLoading.value = true
  pendingError.value = false
  try {
    const result = await getArticleStats()
    draftCount.value = Number(result.data?.draft || 0)
  } catch {
    pendingError.value = true
  } finally {
    pendingLoading.value = false
  }
}
function updateFullscreen() {
  fullscreen.value = !!document.fullscreenElement
}
document.addEventListener('fullscreenchange', updateFullscreen)
onBeforeUnmount(() => document.removeEventListener('fullscreenchange', updateFullscreen))
async function toggleFullscreen() {
  try {
    if (document.fullscreenElement) await document.exitFullscreen()
    else await document.documentElement.requestFullscreen()
  } catch {
    ElMessage.info('当前浏览器不支持全屏模式')
  }
}
async function handleCommand(command) {
  if (command === 'profile') {
    router.push(router.page('profile/index'))
    return
  }
  if (command === 'website') {
    previewWebsite('/')
    return
  }
  if (command !== 'logout') return
  try {
    await ElMessageBox.confirm('确定退出当前账号吗？', '退出登录', {
      confirmButtonText: '退出登录',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    const articleDraftPrefix = `article-draft:${rbac.currentUser?.id || 'unknown'}:`
    await logout()
    for (const key of Object.keys(localStorage))
      if (key.startsWith(articleDraftPrefix)) localStorage.removeItem(key)
    rbac.resetAllData()
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('userInfo')
    const { resetRoutes } = await import('@/router/index')
    resetRoutes()
    await router.push('/login')
    ElMessage.success('已退出登录')
  } catch (error) {
    ElMessage.error(error.message || '退出失败，请重试')
  }
}
</script>
<style lang="scss" scoped>
.navbar {
  height: 48px;
  flex: 0 0 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 12px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  color: #606266;
  z-index: 10;
}
.navbar-left,
.navbar-right {
  display: flex;
  align-items: center;
  min-width: 0;
}
.navbar-left {
  gap: 14px;
  flex: 1;
}
.navbar-right {
  gap: 0;
}
.nav-icon {
  width: 40px;
  height: 46px;
  display: grid;
  place-items: center;
  position: relative;
  background: transparent;
  color: #606266;
  border: 0;
  cursor: pointer;
  font-size: 17px;
}
.nav-icon:hover {
  background: #f6f6f6;
  color: #409eff;
}
.workspace-breadcrumb {
  min-width: 0;
  font-size: 13px;
}
.workspace-breadcrumb :deep(.el-breadcrumb__inner) {
  font-weight: 400;
  color: #909399;
}
.workspace-breadcrumb :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #606266;
}
.console-search {
  width: 100%;
}
.console-search :deep(.el-autocomplete) {
  width: 100%;
}
.console-search :deep(.el-input__wrapper) {
  height: 36px;
  box-shadow: 0 0 0 1px #dcdfe6 inset;
  border-radius: 4px;
  background: #fff;
}
.search-suggestion {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  font-size: 13px;
}
.search-suggestion small {
  color: #909399;
  font-size: 11px;
}
.notification-dot {
  position: absolute;
  right: 11px;
  top: 11px;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #f56c6c;
}
.notification-panel h3 {
  font-size: 14px;
  margin: 0 0 12px;
}
.notification-panel p {
  font-size: 13px;
  color: #909399;
  line-height: 1.8;
  margin-bottom: 12px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 46px;
  padding: 0 12px;
  background: transparent;
  border: 0;
  color: #606266;
  cursor: pointer;
}
.user-info:hover {
  background: #f6f6f6;
}
.user-avatar {
  background: #ecf5ff;
  color: #409eff;
  font-size: 12px;
}
.user-copy strong {
  font-size: 13px;
  font-weight: 400;
  max-width: 120px;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dropdown-icon {
  display: none;
}
.nav-icon:focus-visible,
.user-info:focus-visible {
  outline: 2px solid #409eff;
  outline-offset: -3px;
}
@media (max-width: 720px) {
  .navbar {
    padding: 0 8px;
  }
  .navbar-left {
    gap: 6px;
  }
  .navbar-right {
    gap: 0;
  }
  .nav-icon {
    width: 32px;
  }
  .user-info {
    padding: 0 8px;
  }
  .user-copy {
    display: none;
  }
  .workspace-breadcrumb {
    font-size: 12px;
  }
  .navbar-right > .fullscreen-btn {
    display: none;
  }
}
</style>
