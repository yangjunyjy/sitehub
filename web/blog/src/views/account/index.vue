<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import SkyHero from "@/components/SkyHero.vue";
import ContentState from "@/components/ContentState.vue";
import MemberCard from "./components/MemberCard.vue";
import { useMember } from "@/composables/useMember";
const router = useRouter(),
  { member, restore, logout } = useMember(),
  loading = ref(true),
  error = ref(""),
  busy = ref(false);
async function load() {
  loading.value = true;
  error.value = "";
  try {
    await restore(true);
    if (!member.user)
      await router.replace({ path: "/login", query: { redirect: "/account" } });
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
    router.replace("/");
  } catch (failure) {
    error.value = failure.message;
  } finally {
    busy.value = false;
  }
}
onMounted(load);
</script>
<template>
  <SkyHero
    title="会员中心"
  />
  <section class="content-shell">
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
  </section>
</template>
