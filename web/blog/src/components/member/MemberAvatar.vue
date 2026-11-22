<script setup>
import { computed, ref, watch } from "vue";
const props = defineProps({ user: { type: Object, required: true } });
const failed = ref(false);
watch(
  () => props.user.avatar,
  () => {
    failed.value = false;
  },
);
const initial = computed(() =>
  Array.from(
    props.user.nickname || props.user.username || props.user.email || "读",
  )[0].toLocaleUpperCase(),
);
</script>
<template>
  <span class="member-avatar" aria-hidden="true">
    <img
      v-if="user.avatar && !failed"
      :src="user.avatar"
      alt=""
      referrerpolicy="no-referrer"
      @error="failed = true"
    />
    <template v-else>{{ initial }}</template>
  </span>
</template>
<style scoped>
.member-avatar {
  width: var(--avatar-size, 30px);
  height: var(--avatar-size, 30px);
  border-radius: 50%;
  display: inline-grid;
  place-items: center;
  flex: none;
  overflow: hidden;
  background: linear-gradient(135deg, #f2e5df, #dbd5ea 65%, #c7cfe1);
  color: #50435c;
  box-shadow: inset 0 0 0 1px #665b7422;
  font: 600 calc(var(--avatar-size, 30px) * 0.43)/1 var(--journal-serif);
}
img {
  height: 100%;
  width: 100%;
  object-fit: cover;
}
</style>
