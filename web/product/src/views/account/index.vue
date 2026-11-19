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
        <router-link to="/download"
          ><span>01 / GET STARTED</span>
          <h3>下载产品</h3>
          <p>选择合适的版本与部署方式。</p></router-link
        ><router-link to="/docs"
          ><span>02 / LEARN</span>
          <h3>阅读文档</h3>
          <p>从安装到配置，一步步完成部署。</p></router-link
        >
      </div>
    </template>
  </section>
</template>
<style lang="scss" scoped>
.member-account {
  padding-block: 64px 96px;
  min-height: 65vh;
}
.member-account h1 {
  font-size: 28px;
  margin: 12px 0 32px;
}
.account-profile {
  display: flex;
  gap: 20px;
  align-items: center;
  padding: 32px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
}
.account-avatar {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: 8px;
  color: var(--deploy-primary);
  background: var(--deploy-soft);
  font-size: 24px;
}
.account-profile h2 {
  font-size: 20px;
  margin: 0;
  overflow-wrap: anywhere;
}
.account-profile p {
  font-size: 13px;
  color: var(--deploy-muted);
  margin: 6px 0 0;
}
.account-profile p span {
  margin-inline: 8px;
}
.account-profile button {
  margin-left: auto;
  flex-shrink: 0;
}
.account-next {
  font-size: 23px;
  margin: 44px 0 20px;
}
.account-links {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.account-links a {
  padding: 28px;
  border: 1px solid var(--deploy-line);
  border-radius: 12px;
  background: var(--deploy-white);
}
.account-links a:hover {
  border-color: var(--deploy-line);
}
.account-links span {
  font-size: 11px;
  color: var(--deploy-muted);
  letter-spacing: 0.06em;
}
.account-links h3 {
  font-size: 18px;
  margin: 24px 0 10px;
}
.account-links p {
  font-size: 14px;
  color: var(--deploy-muted);
  margin: 0;
}
.account-error {
  color: #b54444;
  margin-bottom: 20px;
}
@media (max-width: 760px) {
  .member-account {
    padding-block: 36px 60px;
  }
  .account-profile {
    padding: 24px;
    flex-wrap: wrap;
    gap: 16px;
  }
  .account-profile > div {
    min-width: 0;
    flex: 1;
  }
  .account-profile button {
    width: 100%;
    margin-top: 8px;
  }
  .account-profile h2 {
    font-size: 16px;
  }
  .account-links {
    grid-template-columns: 1fr;
  }
}
</style>
