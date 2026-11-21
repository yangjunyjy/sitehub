<template>
  <nav class="tags-view" aria-label="已打开的页面">
    <button
      v-if="overflow"
      class="tags-scroll-button"
      aria-label="向左滚动标签"
      :disabled="!canLeft"
      @click="scroll(-240)"
    >
      <el-icon><ArrowLeft /></el-icon>
    </button>
    <div ref="viewport" class="tags-viewport" @scroll="measure" @wheel="onWheel">
      <div ref="list" class="tags-list">
        <div
          v-for="tag in workspace.views"
          :key="tag.path"
          class="tag-item"
          :class="{ active: tag.path === activePath }"
          :data-tab-path="tag.path"
          @contextmenu.prevent="openMenu($event, tag)"
          @keydown="tabKeydown($event, tag)"
        >
          <button
            type="button"
            class="tag-link"
            :title="tag.title"
            :aria-current="tag.path === activePath ? 'page' : undefined"
            :tabindex="tag.path === activePath ? 0 : -1"
            @click="router.push(tag.fullPath)"
          >
            {{ tag.title }}
          </button>
          <button
            v-if="!tag.affix"
            type="button"
            class="tag-close"
            :aria-label="'关闭 ' + tag.title"
            :disabled="busy"
            @click="perform('close', tag.path)"
          >
            <el-icon><Close /></el-icon>
          </button>
        </div>
      </div>
    </div>
    <button
      v-if="overflow"
      class="tags-scroll-button"
      aria-label="向右滚动标签"
      :disabled="!canRight"
      @click="scroll(240)"
    >
      <el-icon><ArrowRight /></el-icon>
    </button>
    <el-dropdown trigger="click" @command="(command) => perform(command, activePath)">
      <button type="button" class="tags-action-btn" aria-label="页面标签操作">
        <el-icon><Operation /></el-icon>
      </button>
      <template #dropdown
        ><el-dropdown-menu>
          <el-dropdown-item
            v-for="item in actions(activePath)"
            :key="item.command"
            :command="item.command"
            :disabled="item.disabled || busy"
            >{{ item.label }}</el-dropdown-item
          >
        </el-dropdown-menu></template
      >
    </el-dropdown>
    <Teleport to="body">
      <div
        v-if="menu"
        ref="contextMenu"
        class="workspace-tab-menu"
        role="menu"
        :aria-label="menu.title + ' 标签操作'"
        :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
        @keydown="menuKeydown"
        @contextmenu.prevent
      >
        <button
          v-for="item in actions(menu.path)"
          :key="item.command"
          type="button"
          role="menuitem"
          :disabled="item.disabled || busy"
          @click="perform(item.command, menu.path)"
        >
          {{ item.label }}
        </button>
      </div>
    </Teleport>
  </nav>
</template>

<script setup>
import { computed, ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter, isNavigationFailure } from 'vue-router'
import { Close, Operation, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { useWorkspaceStore } from '@/store/workspace'
import { workspacePath } from '@/utils/workspace-route'

const route = useRoute(),
  router = useRouter(),
  workspace = useWorkspaceStore()
const activePath = computed(() => workspacePath(route))
const viewport = ref(),
  list = ref(),
  contextMenu = ref(),
  menu = ref(null),
  busy = ref(false)
const overflow = ref(false),
  canLeft = ref(false),
  canRight = ref(false)
let observer, menuTrigger
const tagElement = (path) =>
  [...(list.value?.children || [])].find((element) => element.dataset.tabPath === path)
function measure() {
  const element = viewport.value
  if (!element) return
  overflow.value = element.scrollWidth > element.clientWidth + 2
  canLeft.value = element.scrollLeft > 1
  canRight.value = element.scrollLeft + element.clientWidth < element.scrollWidth - 1
}
function scroll(left) {
  viewport.value?.scrollBy({
    left,
    behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth',
  })
}
function onWheel(event) {
  if (!overflow.value) return
  event.preventDefault()
  viewport.value.scrollLeft +=
    Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY
}
async function revealActive() {
  await nextTick()
  const element = tagElement(activePath.value),
    wrap = viewport.value
  if (!element || !wrap) return
  const tag = element.getBoundingClientRect(),
    bounds = wrap.getBoundingClientRect()
  if (tag.left < bounds.left) wrap.scrollLeft += tag.left - bounds.left
  else if (tag.right > bounds.right) wrap.scrollLeft += tag.right - bounds.right
  measure()
}
function closingPaths(command, path) {
  const index = workspace.views.findIndex((view) => view.path === path)
  if (index < 0) return []
  return workspace.views
    .filter(
      (view, i) =>
        !view.affix &&
        (command === 'closeAll' ||
          (command === 'close' && view.path === path) ||
          (command === 'closeOthers' && view.path !== path) ||
          (command === 'closeLeft' && i < index) ||
          (command === 'closeRight' && i > index)),
    )
    .map((view) => view.path)
}
function actions(path) {
  return [
    { command: 'refresh', label: '刷新当前', disabled: !workspace.canRefresh(path) },
    ...[
      ['close', '关闭当前'],
      ['closeOthers', '关闭其他'],
      ['closeLeft', '关闭左侧标签'],
      ['closeRight', '关闭右侧标签'],
      ['closeAll', '关闭所有'],
    ].map(([command, label]) => ({
      command,
      label,
      disabled: !closingPaths(command, path).length,
    })),
  ]
}
async function perform(command, path) {
  if (busy.value) return
  closeMenu()
  busy.value = true
  try {
    const target = workspace.views.find((view) => view.path === path)
    if (!target) return
    if (command === 'refresh') {
      if (!workspace.canRefresh(path)) return
      if (activePath.value !== path && isNavigationFailure(await router.push(target.fullPath)))
        return
      await workspace.refresh(path)
      return
    }
    const removed = closingPaths(command, path)
    if (removed.includes(activePath.value)) {
      const remaining = workspace.views.filter((view) => !removed.includes(view.path))
      const currentIndex = workspace.views.findIndex((view) => view.path === activePath.value)
      const destination =
        remaining.find((view) => view.path === path) ||
        [...workspace.views.slice(0, currentIndex)]
          .reverse()
          .find((view) => !removed.includes(view.path)) ||
        remaining[0]
      const failure = await router.push(destination?.fullPath || '/dashboard')
      // A rejected dirty-form guard must leave the page and all tabs intact.
      if (isNavigationFailure(failure)) return
    }
    for (const key of removed) workspace.remove(key)
    await revealActive()
  } finally {
    busy.value = false
  }
}
function closeMenu(restoreFocus = false) {
  menu.value = null
  if (restoreFocus) menuTrigger?.focus()
}
async function openMenu(event, tag) {
  menuTrigger = tagElement(tag.path)?.querySelector('.tag-link')
  const rect = menuTrigger?.getBoundingClientRect()
  menu.value = {
    path: tag.path,
    title: tag.title,
    x: event.clientX || rect?.left || 8,
    y: event.clientY || rect?.bottom || 80,
  }
  await nextTick()
  const bounds = contextMenu.value?.getBoundingClientRect()
  if (!bounds || !menu.value) return
  menu.value.x = Math.max(8, Math.min(menu.value.x, innerWidth - bounds.width - 8))
  menu.value.y = Math.max(8, Math.min(menu.value.y, innerHeight - bounds.height - 8))
  contextMenu.value.querySelector('button:not(:disabled)')?.focus()
}
function tabKeydown(event, tag) {
  if (event.key === 'ContextMenu' || (event.key === 'F10' && event.shiftKey)) {
    event.preventDefault()
    openMenu(event, tag)
    return
  }
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End', 'Delete'].includes(event.key)) return
  event.preventDefault()
  if (event.key === 'Delete') {
    perform('close', tag.path)
    return
  }
  const index = workspace.views.findIndex((view) => view.path === tag.path)
  const next =
    event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? workspace.views.length - 1
        : (index + (event.key === 'ArrowRight' ? 1 : -1) + workspace.views.length) %
          workspace.views.length
  const target = workspace.views[next]
  router.push(target.fullPath).then((failure) => {
    if (!isNavigationFailure(failure)) tagElement(target.path)?.querySelector('.tag-link')?.focus()
  })
}
function menuKeydown(event) {
  if (event.key === 'Escape' || event.key === 'Tab') {
    closeMenu(event.key === 'Escape')
    return
  }
  if (!['ArrowUp', 'ArrowDown', 'Home', 'End'].includes(event.key)) return
  event.preventDefault()
  const items = [...contextMenu.value.querySelectorAll('button:not(:disabled)')]
  const index = items.indexOf(document.activeElement)
  const next =
    event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? items.length - 1
        : (index + (event.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length
  items[next]?.focus()
}
function outside(event) {
  if (menu.value && !contextMenu.value?.contains(event.target)) closeMenu()
}
function resized() {
  closeMenu()
  revealActive()
}
watch(
  () => [route.fullPath, workspace.views.map((view) => view.path).join('|')],
  () => {
    closeMenu()
    revealActive()
  },
  { immediate: true, flush: 'post' },
)
onMounted(() => {
  observer = new ResizeObserver(() => {
    measure()
    revealActive()
  })
  observer.observe(viewport.value)
  observer.observe(list.value)
  document.addEventListener('pointerdown', outside)
  window.addEventListener('resize', resized)
})
onBeforeUnmount(() => {
  observer?.disconnect()
  document.removeEventListener('pointerdown', outside)
  window.removeEventListener('resize', resized)
})
</script>

<style lang="scss" scoped>
.tags-view {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 30px;
  flex: 0 0 30px;
  min-width: 0;
  padding: 0 10px;
  background: #fff;
  border-bottom: 1px solid #d8dce5;
}
.tags-viewport {
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}
.tags-viewport::-webkit-scrollbar {
  display: none;
}
.tags-list {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 100%;
  width: max-content;
}
.tag-item {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  position: relative;
  height: 28px;
  color: #606266;
  transition:
    background 0.15s,
    color 0.15s;
}
.tag-item:hover {
  background: #f5f7fa;
  color: #409eff;
}
.tag-item.active {
  color: var(--el-color-primary-dark-2);
}
.tag-item.active::after {
  position: absolute;
  right: 6px;
  bottom: -1px;
  left: 6px;
  height: 2px;
  background: currentColor;
  content: '';
}
.tag-link,
.tag-close,
.tags-action-btn,
.tags-scroll-button {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
}
.tag-link {
  display: flex;
  align-items: center;
  max-width: 210px;
  height: 100%;
  padding: 0 7px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12px;
  text-align: left;
}
.tag-close {
  display: grid;
  place-items: center;
  width: 20px;
  height: 24px;
  padding: 0;
  margin-right: 3px;
  border-radius: 3px;
  font-size: 11px;
}
@media (hover: hover) and (pointer: fine) {
  .tag-item:not(.active, :hover, :focus-within) .tag-close { opacity: 0; }
}
.tag-close:hover {
  color: var(--el-text-color-primary);
  background: var(--el-fill-color);
}
.tags-action-btn,
.tags-scroll-button {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  font-size: 13px;
  color: #606266;
  flex-shrink: 0;
}
.tags-action-btn:hover,
.tags-scroll-button:hover {
  background: #f5f7fa;
}
.tags-scroll-button:disabled {
  opacity: 0.3;
  cursor: default;
}
.tag-link:focus-visible,
.tag-close:focus-visible,
.tags-action-btn:focus-visible,
.tags-scroll-button:focus-visible {
  outline: 2px solid #409eff;
  outline-offset: -2px;
}
.workspace-tab-menu {
  position: fixed;
  z-index: 3100;
  width: 184px;
  padding: 6px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 5px;
  box-shadow: 0 6px 24px #0002;
}
.workspace-tab-menu button {
  display: block;
  width: 100%;
  text-align: left;
  border: 0;
  padding: 9px 12px;
  color: #606266;
  font-size: 13px;
  background: transparent;
  cursor: pointer;
  border-radius: 3px;
}
.workspace-tab-menu button:hover:not(:disabled),
.workspace-tab-menu button:focus-visible {
  background: #ecf5ff;
  color: #409eff;
  outline: none;
}
.workspace-tab-menu button:disabled {
  color: #c0c4cc;
  cursor: default;
}
@media (max-width: 720px) {
  .tags-view {
    padding-inline: 6px;
  }
  .tag-link {
    max-width: 150px;
    font-size: 12px;
    padding-inline: 10px;
  }
}
</style>
