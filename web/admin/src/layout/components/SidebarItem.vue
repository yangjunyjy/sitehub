<template>
  <el-sub-menu v-if="menu.children?.length" :index="String(menu.id)">
    <template #title>
      <el-icon><ResourceIcon :name="menu.icon || 'Document'" /></el-icon>
      <span>{{ menu.menu_name }}</span>
    </template>
    <el-menu-item v-if="menu.menu_type === 2" :index="menu.path">{{ menu.menu_name }}</el-menu-item>
    <SidebarItem v-for="child in menu.children" :key="child.id" :menu="child" />
  </el-sub-menu>
  <li v-else-if="menu.pageType === 'external'" role="none">
    <a class="el-menu-item" :href="menu.externalUrl" target="_blank" rel="noopener noreferrer" role="menuitem" :title="menu.menu_name">
      <el-icon><ResourceIcon :name="menu.icon || 'Document'" /></el-icon><span>{{ menu.menu_name }} ↗</span>
    </a>
  </li>
  <el-menu-item v-else :index="menu.path">
    <el-icon><ResourceIcon :name="menu.icon || 'Document'" /></el-icon>
    <template #title>{{ menu.menu_name }}</template>
  </el-menu-item>
</template>

<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
defineProps({ menu: { type: Object, required: true } })
</script>
