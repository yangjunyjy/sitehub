<script setup>
import BlogIcon from "@/components/BlogIcon.vue";
defineProps({ user: { type: Object, required: true }, busy: Boolean });
defineEmits(["logout"]);
</script>
<template>
  <section class="member-card">
    <div class="member-avatar"><BlogIcon name="user" :size="35" /></div>
    <span>YOUR LITTLE SPACE</span>
    <h2>{{ user.nickname || user.username || user.email || "云间旅人" }}</h2>
    <p v-if="user.email">{{ user.email }}</p>
    <div class="member-links">
      <router-link to="/">阅读新的故事 ↗</router-link
      ><router-link to="/messages">留下一点星光 ↗</router-link>
    </div>
    <button class="text-button" :disabled="busy" @click="$emit('logout')">
      {{ busy ? "正在退出…" : "退出登录" }}
    </button>
  </section>
</template>
<style lang="scss" scoped>
.member-card {
  max-width: 530px;
  margin: 60px auto 80px;
  text-align: center;
  padding: 45px;
  border: 1px solid var(--line);
  border-radius: 7px;
  > span {
    font-size: 8px;
    letter-spacing: 2px;
    color: #b39cc2;
  }
  h2 {
    font-size: 23px;
    font-weight: 400;
    margin: 16px 0 7px;
  }
  p {
    color: var(--muted);
    font-size: 12px;
  }
  .member-avatar {
    width: 80px;
    height: 80px;
    margin: 0 auto 25px;
    border-radius: 50%;
    background: linear-gradient(135deg, #b8c5df, #e2c8dc);
    color: white;
    display: grid;
    place-items: center;
  }
  .member-links {
    margin: 30px 0;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 20px;
    border-top: 1px solid var(--line);
    padding-top: 25px;
    font-size: 11px;
    color: var(--accent);
  }
}
</style>
