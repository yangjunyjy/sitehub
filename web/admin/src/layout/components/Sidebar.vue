<template>
  <aside class="sidebar" :class="{ collapsed: isCollapse }">
    <div class="brand">
      <BrandLogo class="brand-mark" />
      <div v-if="!isCollapse" class="brand-copy">
        <strong>SiteHub Admin</strong>
      </div>
    </div>

    <el-menu
      :default-active="activeMenu"
      :collapse="isCollapse"
      router
      unique-opened
      class="side-menu"
    >
      <SidebarItem v-for="menu in menus" :key="menu.id" :menu="menu" />
    </el-menu>
    <button
      class="sidebar-collapse"
      type="button"
      :aria-label="isCollapse ? '展开菜单' : '收起菜单'"
      @click="$emit('toggle-collapse')"
    >
      <el-icon><DArrowRight v-if="isCollapse" /><DArrowLeft v-else /></el-icon
      ><span v-if="!isCollapse">收起菜单</span>
    </button>
  </aside>
</template>

<script setup>
import BrandLogo from '@/components/AdminBrand.vue'
import SidebarItem from './SidebarItem.vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRbacStore } from '@/store/rbac'
import { DArrowLeft, DArrowRight } from '@element-plus/icons-vue'

defineProps({ isCollapse: { type: Boolean, default: false } })
defineEmits(['toggle-collapse'])
const router = useRouter()
const route = useRoute()
const rbacStore = useRbacStore()
const activeMenu = computed(() => route.meta.activeMenu || route.path)
const menus = computed(() =>
  router.navigation(rbacStore.currentMenus, rbacStore.currentUser),
)
</script>

<style lang="scss" scoped>
.sidebar {
  width: 210px;
  height: 100%;
  display: flex;
  flex-direction: column;
  flex: 0 0 auto;
  overflow: hidden;
  background: #fff;
  border-right: 1px solid #ebeef5;
  color: #606266;
  transition: width 0.2s;
}
.sidebar.collapsed {
  width: 54px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 48px;
  flex-shrink: 0;
  padding: 0 14px;
}
.brand-mark {
  width: 29px;
  height: 29px;
}
.brand-copy {
  white-space: nowrap;
}
.brand-copy strong {
  font-size: 17px;
  font-weight: 600;
  color: #141414;
}
.side-menu {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 0 8px !important;
}
.sidebar :deep(.el-menu) {
  border: 0;
  background: transparent;
}
.sidebar :deep(.el-menu-item),
.sidebar :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 5px 0;
  border-radius: 4px;
  color: #606266;
  font-size: 14px;
  transition: background 0.15s;
  padding-left: 12px !important;
}
.sidebar :deep(.el-menu-item .el-icon),
.sidebar :deep(.el-sub-menu__title .el-icon) {
  font-size: 18px;
  color: inherit;
  margin-right: 8px;
}
.sidebar :deep(.el-menu-item:hover),
.sidebar :deep(.el-sub-menu__title:hover) {
  background: #f6f6f6;
  color: #303133;
}
.sidebar :deep(.el-menu-item.is-active) {
  color: #fff;
  background: #409eff;
  font-weight: 500;
}
.sidebar :deep(.el-menu-item.is-active .el-icon) {
  color: #fff;
}
.sidebar :deep(.el-sub-menu .el-menu-item) {
  height: 40px;
  line-height: 40px;
  padding-left: 38px !important;
  font-size: 13px;
}
.sidebar :deep(.el-sub-menu__icon-arrow) {
  font-size: 11px;
  color: inherit;
}
.sidebar :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: #409eff;
}
.sidebar-collapse {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  height: 40px;
  flex-shrink: 0;
  background: #fff;
  color: #606266;
  border: 0;
  border-top: 1px solid #ebeef5;
  padding: 0 18px;
  cursor: pointer;
  font-size: 12px;
}
.sidebar-collapse:hover {
  background: #f5f7fa;
  color: #409eff;
}
.sidebar-collapse .el-icon {
  font-size: 16px;
}
.sidebar.collapsed .brand {
  padding: 0;
  justify-content: center;
}
.sidebar.collapsed .side-menu {
  padding: 0 4px !important;
}
.sidebar.collapsed :deep(.el-menu-item),
.sidebar.collapsed :deep(.el-sub-menu__title) {
  padding: 0 !important;
  justify-content: center;
}
.sidebar.collapsed :deep(.el-menu-item .el-icon),
.sidebar.collapsed :deep(.el-sub-menu__title .el-icon) {
  margin: 0;
}
.sidebar.collapsed :deep(.el-menu-item > span),
.sidebar.collapsed :deep(.el-sub-menu__title > span),
.sidebar.collapsed :deep(.el-sub-menu__icon-arrow) {
  display: none;
}
.sidebar.collapsed .sidebar-collapse {
  padding: 0;
  justify-content: center;
}
</style>
