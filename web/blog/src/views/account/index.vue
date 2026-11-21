<script setup>
import { onMounted, ref, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import FavoriteLibrary from "./components/FavoriteLibrary.vue";
import ContentState from "@/components/ContentState.vue";
import MemberCard from "./components/MemberCard.vue";
import { useMember } from "@/composables/useMember";
const router = useRouter(),
  route = useRoute(),
  { member, restore, logout } = useMember(),
  loading = ref(true),
  error = ref(""),
  busy = ref(false);
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
async function signOut() {
  busy.value = true;
  error.value = "";
  try {
    await logout();
    await router.replace({ name: "home" });
  } catch (failure) {
    error.value = failure.message;
  } finally {
    busy.value = false;
  }
}
function toLogin() {
  return router.replace({ name: "login", query: { redirect: route.fullPath } });
}
watch(
  () => member.user,
  (current, previous) => {
    if (previous && !current && !busy.value) toLogin();
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
      /><MemberCard
        v-else-if="member.user"
        :user="member.user"
        :busy="busy"
        @logout="signOut"
      />
      <FavoriteLibrary v-if="!loading && !error && member.user" />
    </section>
  </div>
</template>
<style scoped>
.account-page {
  min-height: 84vh;
  padding: 106px 18px 64px;
  background: #f8f9f5;
  color: #233027;
}
.account-shell {
  width: min(100%, 960px);
  margin: auto;
}
</style>
