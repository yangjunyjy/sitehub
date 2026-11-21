<script setup>
import BlogIcon from "@/components/BlogIcon.vue";
defineProps({ user: { type: Object, required: true }, busy: Boolean });
defineEmits(["logout"]);
</script>
<template>
  <header class="member-card">
    <div class="identity">
      <div class="avatar"><BlogIcon name="user" :size="28" /></div>
      <div>
        <h1>{{ user.nickname || user.username || "个人中心" }}</h1>
        <p>{{ user.email }}</p>
      </div>
    </div>
    <button type="button" :disabled="busy" @click="$emit('logout')">
      {{ busy ? "退出中…" : "退出登录" }}
    </button>
  </header>
</template>
<style scoped>
.member-card,
.identity {
  display: flex;
  align-items: center;
  gap: 16px;
}
.member-card {
  justify-content: space-between;
  padding-bottom: 26px;
  border-bottom: 1px solid #d3dbd2;
  flex-wrap: wrap;
}
.avatar {
  border-radius: 50%;
  background: #dae8da;
  color: #355540;
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  flex: none;
}
h1 {
  font:
    600 28px/1.3 Georgia,
    serif;
  margin: 0;
}
p {
  margin: 3px 0 0;
  color: #5c695f;
  font-size: 14px;
}
.identity {
  min-width: 0;
}
button {
  font-size: 13px;
  border: 1px solid #c4cfc3;
  border-radius: 6px;
  background: white;
  color: #405548;
  padding: 7px 14px;
}
</style>
