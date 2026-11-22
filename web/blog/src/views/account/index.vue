<script setup>
import { onMounted, ref, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import FavoriteLibrary from "./components/FavoriteLibrary.vue";
import ContentState from "@/components/ContentState.vue";
import MemberCard from "./components/MemberCard.vue";
import { useMember } from "@/composables/useMember";
const router = useRouter(),
  route = useRoute(),
  { member, restore } = useMember(),
  loading = ref(true),
  error = ref("");
async function load() {
  loading.value = true;
  error.value = "";
  try {
    await restore();
    if (!member.user) await toLogin();
  } catch (failure) {
    error.value = failure.message;
  } finally {
    loading.value = false;
  }
}
function toLogin() {
  return router.replace({ name: "login", query: { redirect: route.fullPath } });
}
watch(
  () => member.user,
  (current, previous) => {
    if (previous && !current) toLogin();
  },
);
onMounted(load);
</script>
<template>
  <div class="account-page">
    <section class="account-shell">
      <ContentState
        v-if="loading || error"
        :loading="loading"
        :error="error"
        @retry="load"
      /><MemberCard v-else-if="member.user" :user="member.user" />
      <FavoriteLibrary v-if="!loading && !error && member.user" />
    </section>
  </div>
</template>
<style scoped>
.account-page {
  min-height: 80vh;
  padding: clamp(92px, 8vw, 122px) 20px 68px;
  background: var(--journal-paper);
  color: var(--journal-ink);
}
.account-shell {
  width: min(100%, 1000px);
  margin: auto;
}
</style>
