<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BrandMark from "@/components/product/BrandMark.vue";
import { registerMember, sendMemberCode } from "@/api/member";
import { useMember, memberDestination } from "@/composables/useMember";

const route = useRoute(),
  router = useRouter();
const { login } = useMember();
const registering = computed(() => route.name === "register");
const email = ref(""),
  password = ref(""),
  confirm = ref(""),
  code = ref("");
const visible = ref(false),
  busy = ref(false),
  sending = ref(false),
  error = ref(""),
  notice = ref(""),
  seconds = ref(0),
  completed = ref(false);
let cooldown;
const emailValid = () => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim());
const alternate = computed(() => ({
  name: registering.value ? "login" : "register",
  query: route.query.redirect
    ? { redirect: router.resolve(memberDestination(route.query.redirect, router)).fullPath }
    : {},
}));
watch(
  () => route.name,
  () => {
    error.value = "";
    notice.value = "";
    password.value = "";
    confirm.value = "";
    code.value = "";
    completed.value = false;
    visible.value = false;
  },
);
onBeforeUnmount(() => clearInterval(cooldown));
async function sendCode() {
  if (sending.value || seconds.value) return;
  error.value = "";
  notice.value = "";
  if (!emailValid()) {
    error.value = "请先填写有效的邮箱地址";
    return;
  }
  sending.value = true;
  try {
    const result = await sendMemberCode(email.value.trim());
    notice.value = "验证码已发送，10 分钟内有效。未收到时请检查垃圾邮件。";
    seconds.value = result.retryAfter || 60;
    clearInterval(cooldown);
    cooldown = setInterval(() => {
      if (--seconds.value <= 0) clearInterval(cooldown);
    }, 1000);
  } catch (e) {
    error.value = e.message;
  } finally {
    sending.value = false;
  }
}
async function submit() {
  if (busy.value) return;
  error.value = "";
  notice.value = "";
  if (!emailValid()) {
    error.value = "请输入有效的邮箱地址";
    return;
  }
  if (registering.value) {
    const length = new TextEncoder().encode(password.value).length;
    if (
      length < 8 ||
      length > 72 ||
      !/\p{L}/u.test(password.value) ||
      !/\p{Nd}/u.test(password.value)
    ) {
      error.value = "密码需为 8–72 字节，并包含字母和数字";
      return;
    }
    if (password.value !== confirm.value) {
      error.value = "两次输入的密码不一致";
      return;
    }
    if (!/^\d{6}$/.test(code.value)) {
      error.value = "请输入 6 位邮箱验证码";
      return;
    }
  }
  busy.value = true;
  try {
    const values = { email: email.value.trim(), password: password.value };
    if (registering.value) {
      await registerMember({ ...values, code: code.value });
      completed.value = true;
      password.value = "";
      confirm.value = "";
      code.value = "";
    } else {
      await login(values);
      await router.replace(memberDestination(route.query.redirect, router));
    }
  } catch (e) {
    error.value = e.message;
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="member-auth">
    <aside class="auth-story">
      <router-link :to="{ name: 'home' }" class="auth-brand">
        <BrandMark /><strong>Deploy</strong>
      </router-link>
      <div class="auth-story__copy">
        <p class="auth-eyebrow">BUILD. SHIP. OPERATE.</p>
        <h1>让每一次交付，<br />都更从容。</h1>
        <p>连接基础设施与应用交付，<br />从第一次部署，到团队的每一次协作。</p>
        <div class="auth-flow" aria-hidden="true">
          <span>代码</span><i aria-hidden="true">·</i><span>构建</span
          ><i aria-hidden="true">·</i><span>部署</span
          ><i aria-hidden="true">·</i><span>运行</span>
        </div>
        <div class="auth-benefits">
          <span>统一资产管理</span><span>自动化应用交付</span
          ><span>全程操作审计</span>
        </div>
      </div>
      <p class="auth-story__footer">从代码到生产，每一步都清晰可见。</p>
    </aside>
    <section class="auth-content">
      <router-link class="auth-back" :to="{ name: 'home' }"> 返回官网</router-link>
      <div v-if="completed" class="auth-form auth-success" role="status">
        <span class="success-mark">✓</span>
        <p class="auth-eyebrow">YOU'RE ALL SET</p>
        <h2>欢迎加入 Deploy</h2>
        <p>
          <strong>{{ email }}</strong>
          已完成邮箱验证。<br />现在可以使用邮箱和密码登录。
        </p>
        <router-link class="site-button site-button--primary" :to="alternate"
          >前往登录
        </router-link>
      </div>
      <form v-else class="auth-form" @submit.prevent="submit">
        <p class="auth-eyebrow">YOUR DEPLOY ACCOUNT</p>
        <h2>{{ registering ? "创建你的账号" : "欢迎回来" }}</h2>
        <p class="auth-intro">
          {{
            registering
              ? "使用邮箱注册，开启你的 Deploy 之旅。"
              : "登录 Deploy，继续探索更简单的交付方式。"
          }}
        </p>
        <div class="auth-switch">
          <router-link
            :to="{ path: '/login', query: route.query }"
            :class="{ active: !registering }"
            :aria-current="!registering ? 'page' : undefined"
            >登录</router-link
          ><router-link
            :to="{ path: '/register', query: route.query }"
            :class="{ active: registering }"
            :aria-current="registering ? 'page' : undefined"
            >邮箱注册</router-link
          >
        </div>
        <label for="member-email">邮箱地址</label>
        <input
          id="member-email"
          v-model="email"
          type="email"
          autocomplete="email"
          placeholder="you@example.com"
          required
          maxlength="254"
          :disabled="busy || sending"
        />
        <template v-if="registering">
          <label for="member-code">邮箱验证码</label>
          <div class="code-field">
            <input
              id="member-code"
              v-model="code"
              inputmode="numeric"
              autocomplete="one-time-code"
              placeholder="6 位验证码"
              maxlength="6"
              pattern="[0-9]{6}"
              required
              :disabled="busy"
            /><button
              type="button"
              :disabled="sending || busy || seconds > 0"
              @click="sendCode"
            >
              {{
                sending
                  ? "发送中…"
                  : seconds
                    ? `${seconds}s 后重发`
                    : "获取验证码"
              }}
            </button>
          </div>
        </template>
        <label for="member-password">{{
          registering ? "设置密码" : "密码"
        }}</label>
        <div class="password-field">
          <input
            id="member-password"
            v-model="password"
            :type="visible ? 'text' : 'password'"
            :autocomplete="registering ? 'new-password' : 'current-password'"
            :placeholder="
              registering ? '至少 8 位，包含字母和数字' : '请输入密码'
            "
            required
            :disabled="busy"
            :aria-describedby="registering ? 'password-hint' : undefined"
          /><button
            type="button"
            :aria-label="visible ? '隐藏密码' : '显示密码'"
            :aria-pressed="visible"
            @click="visible = !visible"
          >
            {{ visible ? "隐藏" : "显示" }}
          </button>
        </div>
        <template v-if="registering"
          ><small id="password-hint"
            >8–72 字节，建议使用字母、数字和符号组合。</small
          ><label for="member-confirm">确认密码</label
          ><input
            id="member-confirm"
            v-model="confirm"
            :type="visible ? 'text' : 'password'"
            autocomplete="new-password"
            placeholder="再次输入密码"
            required
            :disabled="busy"
        /></template>
        <p v-if="error" class="auth-message auth-message--error" role="alert">
          {{ error }}
        </p>
        <p v-if="notice" class="auth-message" role="status">{{ notice }}</p>
        <button
          class="site-button site-button--primary auth-submit"
          type="submit"
          :disabled="busy || sending"
        >
          {{ busy ? "正在处理…" : registering ? "创建账号" : "登录" }}
        </button>
        <p class="auth-alternate">
          {{ registering ? "已经有账号？" : "还没有账号？" }}
          <router-link :to="alternate">{{
            registering ? "立即登录" : "邮箱注册"
          }}</router-link>
        </p>
        <p class="auth-note">
          这是 Deploy 官网账号，管理后台账号请在管理端使用。
        </p>
      </form>
      <footer class="auth-content__footer">
        <router-link :to="{ name: 'docs', params: { id: '' } }">使用文档 </router-link
        ><span>Deploy · 让交付更简单</span>
      </footer>
    </section>
  </div>
</template>

<style lang="scss" scoped>
.member-auth {
  min-height: 100dvh;
  display: grid;
  grid-template-columns: 46% 54%;
  background: var(--deploy-white);
}

.auth-story {
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  padding: 48px clamp(32px, 5vw, 88px);
  background: var(--deploy-paper);
  border-right: 1px solid var(--deploy-line);
}

.auth-brand {
  display: inline-flex;
  gap: 12px;
  align-items: center;
  align-self: flex-start;
}

.auth-brand strong {
  font-size: 28px;
  letter-spacing: -1px;
}

.auth-story__copy {
  margin: auto 0;
  padding: 80px 0;
}

.auth-eyebrow {
  font-size: 11px;
  letter-spacing: 0.16em;
  color: var(--deploy-primary);
  font-weight: 650;
}

.auth-story h1 {
  font-size: clamp(28px, 3vw, 38px);
  font-weight: 650;
  line-height: 1.5;
  letter-spacing: -0.035em;
  margin: 22px 0;
}

.auth-story__copy > p:not(.auth-eyebrow) {
  color: var(--deploy-muted);
  font-size: 16px;
  line-height: 2;
}

.auth-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 48px 0 28px;
}

.auth-flow span {
  display: grid;
  place-items: center;
  flex: 1;
  max-width: 80px;
  aspect-ratio: 1;
  border: 1px solid var(--deploy-white);
  border-radius: 8px;
  background: var(--deploy-soft);
  box-shadow: 0 10px 24px #00000012;
  color: var(--deploy-primary);
  font-size: 14px;
}

.auth-flow i {
  font-style: normal;
  color: var(--deploy-faint);
}

.auth-benefits {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--deploy-muted);
}

.auth-benefits span::before {
  content: "✓";
  color: var(--deploy-primary);
  margin-right: 6px;
}

.auth-story__footer {
  font-size: 12px;
  color: var(--deploy-muted);
}

.auth-content {
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 36px clamp(28px, 5vw, 80px);
}

.auth-back {
  font-size: 13px;
  color: var(--deploy-muted);
  align-self: flex-end;
}

.auth-form {
  width: 100%;
  max-width: 400px;
  margin: auto;
  padding: 40px 0;
}

.auth-form h2 {
  font-size: 30px;
  letter-spacing: -0.03em;
  margin: 8px 0 10px;
}

.auth-intro,
.auth-success > p:not(.auth-eyebrow) {
  font-size: 14px;
  color: var(--deploy-muted);
  margin: 0 0 26px;
}

.auth-switch {
  display: flex;
  padding: 4px;
  background: var(--deploy-paper);
  border-radius: 9px;
  margin: 28px 0;
}

.auth-switch a {
  flex: 1;
  text-align: center;
  padding: 9px;
  border-radius: 6px;
  font-size: 14px;
  color: var(--deploy-muted);
}

.auth-switch a.active {
  background: var(--deploy-white);
  color: var(--deploy-primary);
  box-shadow: 0 2px 8px #00000012;
  font-weight: 600;
}

.auth-form label {
  display: block;
  font-size: 13px;
  font-weight: 550;
  margin: 19px 0 8px;
}

.auth-form input {
  width: 100%;
  min-width: 0;
  height: 48px;
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  border-radius: 8px;
  padding: 0 14px;
  color: var(--deploy-ink);
  font-size: 14px;
}

.auth-form input:focus {
  border-color: var(--deploy-primary);
  outline: 2px solid var(--deploy-soft);
}

.auth-form input::placeholder {
  color: var(--deploy-faint);
}

.password-field,
.code-field {
  display: flex;
  position: relative;
}

.password-field input {
  padding-right: 64px;
}

.password-field button {
  position: absolute;
  right: 8px;
  top: 6px;
  height: 36px;
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 12px;
}

.code-field {
  gap: 10px;
}

.code-field input {
  flex: 1;
}

.code-field button {
  flex-shrink: 0;
  width: 116px;
  border: 1px solid var(--deploy-line);
  background: var(--deploy-paper);
  color: var(--deploy-primary);
  border-radius: 8px;
  font-size: 12px;
}

.auth-form small {
  display: block;
  margin-top: 7px;
  font-size: 11px;
  color: var(--deploy-muted);
}

.auth-submit {
  width: 100%;
  margin-top: 28px;
  justify-content: space-between;
  min-height: 48px;
}

.auth-alternate {
  font-size: 13px;
  text-align: center;
  color: var(--deploy-muted);
  margin: 22px 0;
}

.auth-alternate a {
  color: var(--deploy-primary);
  font-weight: 600;
}

.auth-note {
  border-top: 1px solid var(--deploy-line);
  padding-top: 20px;
  font-size: 11px;
  color: var(--deploy-muted);
  line-height: 1.8;
}

.auth-message {
  padding: 10px 12px;
  border-radius: 7px;
  background: var(--deploy-paper);
  color: #35785b;
  font-size: 12px;
}

.auth-message--error {
  background: var(--deploy-paper);
  color: #b54444;
}

.auth-content__footer {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  color: var(--deploy-muted);
  font-size: 11px;
}

.auth-success .site-button {
  margin-top: 16px;
  width: 100%;
}

.auth-success strong {
  overflow-wrap: anywhere;
}

.success-mark {
  display: grid;
  place-items: center;
  background: var(--deploy-paper);
  color: #4b9e75;
  width: 56px;
  height: 56px;
  border-radius: 8px;
  font-size: 28px;
  margin-bottom: 28px;
}

@media (max-width: 900px) {
  .member-auth {
    grid-template-columns: 38% 62%;
  }

  .auth-story {
    padding: 36px 28px;
  }

  .auth-story h1 {
    font-size: 32px;
  }

  .auth-flow {
    gap: 6px;
  }

  .auth-flow span {
    border-radius: 10px;
    font-size: 12px;
  }

  .auth-benefits {
    gap: 8px;
    flex-direction: column;
  }

  .auth-story__copy > p:not(.auth-eyebrow) {
    font-size: 13px;
  }

  .auth-content {
    padding: 28px 36px;
  }
}

@media (max-width: 640px) {
  .member-auth {
    display: block;
  }

  .auth-story {
    padding: 22px 24px;
    border-right: 0;
    border-bottom: 1px solid var(--deploy-line);
  }

  .auth-brand strong {
    font-size: 23px;
  }

  .auth-story__copy,
  .auth-story__footer {
    display: none;
  }

  .auth-content {
    min-height: calc(100dvh - 89px);
    padding: 22px 24px;
  }

  .auth-form {
    padding: 28px 0;
  }

  .auth-form h2 {
    font-size: 27px;
  }

  .auth-content__footer {
    padding-top: 16px;
  }

  .auth-form label {
    margin-top: 16px;
  }
}
</style>
