<script setup>
import { computed, defineAsyncComponent } from "vue";
import { iconURL } from "./catalog";

const props = defineProps({
  name: String,
  size: { type: Number, default: 18 },
});
const url = computed(() => iconURL(props.name));
const legacyElement = defineAsyncComponent(
  () => import("./LegacyElementIcon.vue"),
);
</script>

<template>
  <span
    v-if="url && !name.startsWith('custom:')"
    class="resource-icon builtin"
    aria-hidden="true"
    :style="{
      width: size + 'px',
      height: size + 'px',
      maskImage: 'url(' + JSON.stringify(url) + ')',
    }"
  />
  <img
    v-else-if="url"
    class="resource-icon"
    :src="url"
    :width="size"
    :height="size"
    alt=""
  />
  <component
    v-else-if="/^[A-Z][a-zA-Z0-9]*$/.test(name || '')"
    :is="legacyElement"
    :name="name"
    :size="size"
  />
  <span v-else-if="name" class="resource-icon" aria-hidden="true">{{
    name
  }}</span>
</template>

<style scoped lang="scss">
.resource-icon {
  display: inline-block;
  vertical-align: middle;
  flex: none;
  object-fit: contain;
}
.builtin {
  background: currentColor;
  mask-size: contain;
  mask-repeat: no-repeat;
  mask-position: center;
}
</style>
