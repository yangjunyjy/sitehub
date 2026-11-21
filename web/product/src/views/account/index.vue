<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useMember } from "@/composables/useMember";
const { member, restoreMember, logout } = useMember();
const router = useRouter(),
  loading = ref(true),
  error = ref(""),
  leaving = ref(false);
async function load() {
  error.value = "";
  loading.value = true;
  try {
    if (!(await restoreMember()))
      await router.replace("/login?redirect=/account");
  } catch (e) {
    error.value = e.message;
  } finally {
    loading.value = false;
  }
}
async function signOut() {
  leaving.value = true;
  error.value = "";
  try {
    await logout();
    await router.replace("/login");
  } catch (e) {
    error.value = e.message;
  } finally {
    leaving.value = false;
  }
}
onMounted(load);
</script>
<template>
  <section class="site-shell member-account">
    <p class="site-kicker">MY DEPLOY</p>
    <h1>我的账号</h1>
    <p v-if="loading" role="status">正在加载账号信息…</p>
    <div v-if="error" role="alert" class="account-error">
      {{ error }}
      <button
        v-if="!member.user"
        class="site-button site-button--outline"
        @click="load"
      >
        重新加载
      </button>
    </div>
    <template v-if="!loading && member.user">
      <div class="account-profile">
        <span class="account-avatar">{{
          member.user.email[0].toUpperCase()
        }}</span>
        <div>
          <h2>{{ member.user.email }}</h2>
          <p>邮箱已验证 <span>·</span> Deploy 官网账号</p>
        </div>
        <button
          class="site-button site-button--outline"
          :disabled="leaving"
          @click="signOut"
        >
          {{ leaving ? "正在退出…" : "退出登录" }}
        </button>
      </div>
      <h2 class="account-next">开始使用 Deploy</h2>
      <div class="account-links">
        <router-link :to="{ name: 'download' }"
          ><span>01 / GET STARTED</span>
          <h3>下载产品</h3>
          <p>选择合适的版本与部署方式。</p></router-link
        ><router-link :to="{ name: 'docs', params: { id: '' } }"
          ><span>02 / LEARN</span>
          <h3>阅读文档</h3>
          <p>从安装到配置，一步步完成部署。</p></router-link
        >
      </div>
    </template>
  </section>
</template>
<style lang="scss" scoped src="./account.scss"></style>
