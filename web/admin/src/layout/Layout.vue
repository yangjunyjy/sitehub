<template>
  <div class="layout-container">
    <a class="admin-skip-link" href="#admin-main">跳至页面内容</a>
    <Sidebar v-if="!isNarrow" :is-collapse="isCollapse" @toggle-collapse="toggleCollapse" />
    <el-drawer
      v-else
      v-model="mobileMenu"
      class="admin-navigation-drawer"
      direction="ltr"
      size="260px"
      :with-header="false"
      title="管理菜单"
      append-to-body
    >
      <Sidebar :is-collapse="false" @toggle-collapse="mobileMenu = false" />
    </el-drawer>
    <div class="main-container">
      <Navbar
        :is-collapse="isNarrow ? !mobileMenu : isCollapse"
        @toggle-collapse="toggleCollapse"
      />
      <TagsView />
      <div id="admin-main" ref="content" class="app-main" tabindex="-1">
        <router-view v-slot="{ Component, route: currentRoute }">
          <Transition
            name="workspace-slide"
            @before-leave="disableLeaving"
            @before-enter="enableEntering"
          >
            <keep-alive :include="workspace.cachedNames" :max="12">
              <component
                :is="Component"
                :key="
                  workspacePath(currentRoute) +
                  ':' +
                  (workspace.revisions[workspacePath(currentRoute)] || 0)
                "
              />
            </keep-alive>
          </Transition>
        </router-view>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import TagsView from './components/TagsView.vue'
import { useWorkspaceStore } from '@/store/workspace'
import { useRbacStore } from '@/store/rbac'
import { isSuperAdmin } from '@/router/pages'
import { workspacePath } from '@/utils/workspace-route'

const route = useRoute()
const workspace = useWorkspaceStore()
const rbac = useRbacStore()
const router = useRouter()
function allowedRoute(target) {
  const { meta } = target
  return (
    (!meta.superAdmin || isSuperAdmin(rbac.currentUser)) &&
    (!meta.permission || rbac.hasPermission(meta.permission)) &&
    (!meta.permissionsAny || meta.permissionsAny.some(rbac.hasPermission))
  )
}
function canOpen(target) {
  return Boolean(target.meta.title) && !target.meta.hidden && !target.meta.missingComponent && allowedRoute(target)
}
watch(
  () => rbac.currentUser?.id,
  (id) => {
    workspace.restore(id, router.resolve, canOpen)
    workspace.visit(route)
  },
  { immediate: true },
)
watch(
  () => [rbac.currentPermissions, rbac.currentUser],
  () => {
    workspace.retain((view) => canOpen(router.resolve(view.fullPath)))
    // Reuse the normal router guard, including editor unsaved-change prompts.
    if (!allowedRoute(route)) router.replace('/403')
  },
  { deep: true },
)
const content = ref()
const isCollapse = ref(false)
const mobileMenu = ref(false)
const media = window.matchMedia('(max-width: 720px)')
const isNarrow = ref(media.matches)
function handleResize(event) {
  isNarrow.value = event.matches
  if (!event.matches) mobileMenu.value = false
}
media.addEventListener('change', handleResize)
onBeforeUnmount(() => {
  media.removeEventListener('change', handleResize)
})
function disableLeaving(element) {
  element.inert = true
}
function enableEntering(element) {
  element.inert = false
}
function toggleCollapse() {
  if (isNarrow.value) mobileMenu.value = !mobileMenu.value
  else isCollapse.value = !isCollapse.value
}
watch(
  () => route.fullPath,
  () => {
    workspace.visit(route)
    mobileMenu.value = false
    content.value?.scrollTo({ top: 0 })
  },
  { immediate: true },
)
</script>

<style lang="scss" scoped>
.layout-container {
  display: flex;
  height: 100dvh;
  overflow: hidden;
  background: var(--bg-page);
}
.main-container {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.app-main {
  flex: 1;
  min-height: 0;
  overflow: auto;
  position: relative;
  scrollbar-gutter: stable;
  outline: none;
}
.app-main::-webkit-scrollbar {
  width: 6px;
}
.app-main::-webkit-scrollbar-thumb {
  background: #ecf5ff;
  border-radius: 4px;
}
.workspace-slide-enter-active {
  transition:
    opacity 180ms ease,
    transform 220ms cubic-bezier(0.2, 0.7, 0.2, 1);
}
.workspace-slide-leave-active {
  position: absolute;
  inset: 0 0 auto;
  pointer-events: none;
  transition:
    opacity 120ms ease,
    transform 160ms ease;
}
.workspace-slide-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.workspace-slide-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
.admin-skip-link {
  position: fixed;
  top: 8px;
  left: 8px;
  z-index: 3000;
  padding: 12px 20px;
  border-radius: 4px;
  color: white;
  background: #409eff;
  transform: translateY(-150%);
}
.admin-skip-link:focus {
  transform: translateY(0);
}
@media (prefers-reduced-motion: reduce) {
  .workspace-slide-enter-active,
  .workspace-slide-leave-active {
    transition: none;
  }
}
</style>

<style lang="scss">
.admin-navigation-drawer .el-drawer__body {
  padding: 0;
  overflow: hidden;
}
.admin-navigation-drawer .sidebar {
  width: 100%;
}
</style>
