<template>
  <el-sub-menu v-if="menu.type === 'directory'" :index="String(menu.id)">
    <template #title>
      <el-icon><component :is="icon" /></el-icon>
      <span>{{ menu.title }}</span>
    </template>
    <SidebarItem v-for="child in menu.children" :key="child.id" :menu="child" />
  </el-sub-menu>
  <li v-else-if="menu.pageType === 'external'" role="none">
    <a class="el-menu-item" :href="menu.externalUrl" target="_blank" rel="noopener noreferrer" role="menuitem" :title="menu.title">
      <el-icon><component :is="icon" /></el-icon><span>{{ menu.title }} ↗</span>
    </a>
  </li>
  <el-menu-item v-else :index="menu.path">
    <el-icon><component :is="icon" /></el-icon>
    <template #title>{{ menu.title }}</template>
  </el-menu-item>
</template>

<script setup>
import { computed } from 'vue'
import * as icons from '@element-plus/icons-vue'

const props = defineProps({ menu: { type: Object, required: true } })
const icon = computed(() => icons[props.menu.icon] || icons.Document)
</script>
