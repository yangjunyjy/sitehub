<script setup>
import { reactive, ref } from "vue";
import { Connection, Hide, Lock, User, View } from "@element-plus/icons-vue";
import { api } from "../api";
const emit = defineEmits(["authenticated"]);
const form = reactive({ username: "", password: "" });
const busy = ref(false),
  error = ref(""),
  showPassword = ref(false);
async function login() {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    await api("auth/login", {
      username: form.username.trim(),
      password: form.password,
    });
    form.password = "";
    emit("authenticated");
  } catch (e) {
    error.value = e.message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <main class="login-page">
    <div class="login-frame">
      <section class="login-art" aria-hidden="true">
        <div class="login-art-brand"><Connection /> SiteHub Gateway</div>
        <div class="network-art">
          <div class="orbit orbit-a"></div>
          <div class="orbit orbit-b"></div>
          <div class="orbit orbit-c"></div>
          <div class="network-core"><Connection /></div>
          <i v-for="n in 4" :key="n" :class="'node node-' + n"></i>
        </div>
      </section>
      <form class="login-form" @submit.prevent="login">
        <div>
          <h1>登录网关控制台</h1>
          <p class="muted">使用运维人员为你配置的网关账号。</p>
        </div>
        <label
          >用户名<span class="input-icon"
            ><User />
            <input
              v-model="form.username"
              autocomplete="username"
              required
              placeholder="网关账号"
            /> </span
        ></label>
        <div>
          <label for="password">密码</label>
          <div class="input-icon">
            <Lock />
            <input
              id="password"
              v-model="form.password"
              autocomplete="current-password"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="账号密码"
            />
            <button
              type="button"
              class="icon"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              @click="showPassword = !showPassword"
            >
              <component :is="showPassword ? Hide : View" />
            </button>
          </div>
        </div>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <button class="primary login-submit" :disabled="busy">
          {{ busy ? "正在登录…" : "登录" }}
        </button>
      </form>
    </div>
  </main>
</template>
