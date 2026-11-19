<script setup>
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useMember, memberDestination } from "@/composables/useMember";
const router = useRouter(),
  route = useRoute(),
  { login } = useMember();
const form = reactive({ email: "", password: "" }),
  error = ref(""),
  submitting = ref(false),
  visible = ref(false);
async function submit() {
  if (submitting.value) return;
  error.value = "";
  submitting.value = true;
  try {
    await login({ email: form.email.trim(), password: form.password });
    await router.replace(memberDestination(route.query.redirect));
  } catch (failure) {
    error.value = failure.message;
  } finally {
    submitting.value = false;
  }
}
</script>
<template>
  <div class="login-card">
    <span class="login-eyebrow">WELCOME BACK</span>
    <h2>很高兴，又见到你</h2>
    <p class="login-description">登录你的会员账号，继续这段旅程。</p>
    <form @submit.prevent="submit">
      <label class="field"
        >邮箱<input
          v-model="form.email"
          required
          type="email"
          maxlength="254"
          autocomplete="username"
          placeholder="you@example.com" /></label
      ><label class="field password-field"
        >密码
        <div>
          <input
            v-model="form.password"
            required
            :type="visible ? 'text' : 'password'"
            maxlength="200"
            autocomplete="current-password"
            placeholder="请输入密码"
          /><button
            type="button"
            :aria-label="visible ? '隐藏密码' : '显示密码'"
            @click="visible = !visible"
          >
            {{ visible ? "隐藏" : "显示" }}
          </button>
        </div></label
      >
      <p v-if="error" class="form-error" role="alert">{{ error }}</p>
      <button class="button login-submit" :disabled="submitting">
        {{ submitting ? "正在登录…" : "登录账号 ↗" }}
      </button>
    </form>
    <router-link class="login-back" to="/">暂不登录，随意逛逛 →</router-link>
  </div>
</template>
<style lang="scss" scoped>
.login-card {
  margin-inline:auto;
  width: 420px;
  max-width: 100%;
  padding: 42px 38px;
  background: #fffffff7;
  box-shadow: 0 20px 70px #4b36541c;
  border-radius: 8px;
  text-align: left;
  color: var(--ink);
  h2 {
    font-size: 23px;
    font-weight: 400;
    letter-spacing: 2px;
    margin: 13px 0;
  }
  .login-description {
    font-size: 11px !important;
    color: var(--muted);
    text-shadow: none;
    letter-spacing: 0 !important;
    margin: 0 0 30px;
  }
  .field {
    margin-bottom: 20px;
  }
  .login-eyebrow {
    font-size: 8px;
    letter-spacing: 3px;
    color: #b7a3bc;
  }
  .login-submit {
    width: 100%;
    margin-top: 8px;
  }
  .login-back {
    font-size: 10px;
    color: var(--muted);
    display: block;
    text-align: center;
    margin-top: 24px;
  }
}
.password-field > div {
  display: flex;
  border: 1px solid #e7e3ec;
  border-radius: 4px;
  background: white;
  input {
    flex: 1;
    min-width: 0;
    border: 0;
  }
  button {
    border: 0;
    background: none;
    color: #a89ab2;
    font-size: 10px;
    padding: 0 12px;
  }
}
@media (max-width: 640px) {
  .login-card {
    padding: 32px 25px;
  }
}
</style>
