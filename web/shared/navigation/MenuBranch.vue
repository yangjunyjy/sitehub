<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import ResourceIcon from '../icons/ResourceIcon.vue';

const props = defineProps({ item: {type: Object, required: true}, mobile: Boolean });
const route = useRoute();
const root = ref(null), trigger = ref(null), open = ref(false);
const close = () => { open.value = false; };
function outside(event) { if (!root.value?.contains(event.target)) close(); }
function leave(event) {
  if (!props.mobile && event.pointerType === 'mouse' && !root.value?.querySelector(':focus-visible')) close();
}
function escape() { close(); trigger.value?.focus(); }
watch(() => route.fullPath, close);
onMounted(() => { if (props.item.children.length) document.addEventListener('pointerdown', outside); });
onBeforeUnmount(() => document.removeEventListener('pointerdown', outside));
</script>

<template>
  <div ref="root" class="menu-branch" :class="{ mobile }"
    @pointerenter="!mobile && $event.pointerType === 'mouse' && (open = true)"
    @pointerleave="leave"
    @focusout="!root.contains($event.relatedTarget) && close()">
    <component v-if="item.path" :is="item.external ? 'a' : 'router-link'"
      v-bind="item.external ? { href: item.path, target: '_blank', rel: 'noopener noreferrer' } : { to: item.path }">
      <ResourceIcon v-if="item.icon" :name="item.icon" />{{ item.label }}
    </component>
    <button v-if="item.children.length" ref="trigger" type="button" :aria-expanded="open"
      :aria-label="item.path ? item.label + ' 子菜单' : undefined"
      @click="open = !mobile && $event.pointerType === 'mouse' ? true : !open"
      @keydown.esc.stop="escape">
      <ResourceIcon v-if="item.icon && !item.path" :name="item.icon" />
      {{ item.path ? '' : item.label }} <span aria-hidden="true">▾</span>
    </button>
    <div v-if="open && item.children.length" class="menu-children" @keydown.esc.stop="escape">
      <MenuBranch v-for="child in item.children" :key="child.id" :item="child" :mobile="mobile" />
    </div>
  </div>
</template>

<style scoped>
.menu-branch { position: relative; display: flex; align-items: center; }
a, button { color: inherit; background: transparent; border: 0; font: inherit; cursor: pointer; display: inline-flex; align-items: center; gap: .4em; padding: .55em .25em; white-space: nowrap; }
a:hover, button:hover, .router-link-exact-active { text-decoration: underline; text-underline-offset: .3em; }
.menu-children { position: absolute; z-index: 1; top: 100%; inset-inline-start: 0; min-width: max-content; padding: .45rem .75rem; border-radius: .6rem; background: var(--menu-surface, var(--deploy-white, white)); color: var(--menu-ink, var(--deploy-ink)); box-shadow: 0 8px 24px #19213933; }
.menu-children .menu-children, .mobile .menu-children { position: static; box-shadow: none; background: transparent; border-inline-start: 1px solid currentColor; margin-inline-start: .65rem; }
.menu-children .menu-branch, .mobile { display: block; }
</style>
