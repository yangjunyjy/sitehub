<script setup>
import { computed, onBeforeUnmount, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  registerMember,
  sendMemberCode,
  sendResetCode,
  resetMemberPassword,
} from "@/api/member";
import { memberDestination, useMember } from "@/composables/useMember";
import BlogIcon from "@/components/BlogIcon.vue";

const props = defineProps({ registering: Boolean });
const router = useRouter(),
  route = useRoute(),
  { login } = useMember();
const form = reactive({
  email: "",
  account: "",
  username: "",
  nickname: "",
  password: "",
  code: "",
});
const visible = ref(false),
  sending = ref(false),
  busy = ref(false),
  registered = ref(false);
const notice = ref(""),
  error = ref(""),
  seconds = ref(0),
  emailInput = ref(null);
const recovering = ref(false);
const creating = computed(() => props.registering && !registered.value);
const needsCode = computed(() => creating.value || recovering.value);
let timer;
let disposed = false;

function setRecovery(active) {
  recovering.value = active;
  if (active && form.account.includes("@")) form.email = form.account.trim();
  form.code = form.password = notice.value = error.value = "";
  visible.value = false;
}

async function send() {
  if (sending.value || seconds.value || !emailInput.value.reportValidity())
    return;
  sending.value = true;
  error.value = notice.value = "";
  try {
    const result = await (recovering.value ? sendResetCode : sendMemberCode)(
      form.email.trim(),
    );
    if (disposed) return;
    seconds.value = Number(result?.retryAfter) || 60;
    notice.value = recovering.value
      ? "如该邮箱已注册，请查看密码重置邮件。"
      : "验证码已发送，请查看邮箱。";
    timer = setInterval(() => {
      if (--seconds.value <= 0) clearInterval(timer);
    }, 1000);
  } catch (err) {
    error.value = err.message;
  } finally {
    sending.value = false;
  }
}
async function submit() {
  if (busy.value) return;
  error.value = notice.value = "";
  const values = { email: form.email.trim(), password: form.password };
  if (needsCode.value) {
    const length = new TextEncoder().encode(form.password).length;
    if (
      length < 8 ||
      length > 72 ||
      !/\p{L}/u.test(form.password) ||
      !/\p{Nd}/u.test(form.password)
    ) {
      error.value = "密码需为 8–72 字节，并包含字母和数字。";
      return;
    }
  }
  busy.value = true;
  try {
    if (recovering.value) {
      await resetMemberPassword({ ...values, code: form.code.trim() });
      form.account = values.email;
      setRecovery(false);
      notice.value = "密码已重置，请使用新密码登录。";
      return;
    }
    if (creating.value) {
      await registerMember({
        ...values,
        code: form.code.trim(),
        username: form.username.trim(),
        nickname: form.nickname.trim(),
      });
      form.account = values.email;
      registered.value = true;
    }
    await login({ account: form.account.trim(), password: form.password });
    await router.replace(memberDestination(route.query.redirect, router));
  } catch (err) {
    error.value =
      (registered.value ? "账号已创建，请重新登录。" : "") + err.message;
  } finally {
    busy.value = false;
  }
}
onBeforeUnmount(() => {
  disposed = true;
  clearInterval(timer);
});
</script>
<template>
  <section class="member-entry">
    <div class="entry-panel">
      <div class="entry-photo" aria-hidden="true"></div>
      <div class="entry-form">
        <router-link :to="{ name: 'home' }" class="wordmark"
          ><BlogIcon name="feather" :size="20" />云间小记</router-link
        >
        <h1>
          {{
            recovering ? "找回密码" : creating ? "创建读者账号" : "登录读者账号"
          }}
        </h1>
        <p>
          {{
            recovering
              ? "验证注册邮箱后，设置新的登录密码。"
              : creating
                ? "用邮箱注册，将喜欢的文章收入自己的收藏夹。"
                : "继续阅读你的收藏，参与文章讨论。"
          }}
        </p>
        <form @submit.prevent="submit">
          <template v-if="creating">
            <label
              >用户名<input
                v-model.trim="form.username"
                autocomplete="username"
                required
                pattern="[A-Za-z0-9_]{3,40}"
                maxlength="40"
                placeholder="3–40 位字母、数字或下划线"
            /></label>
            <label
              >昵称<input
                v-model.trim="form.nickname"
                autocomplete="nickname"
                required
                maxlength="40"
                placeholder="希望大家怎么称呼你"
            /></label>
          </template>
          <label v-if="!needsCode"
            >用户名或邮箱<input
              v-model.trim="form.account"
              autocomplete="username"
              maxlength="254"
              required
          /></label>
          <label v-if="needsCode"
            >邮箱<input
              ref="emailInput"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              maxlength="254"
              placeholder="you@example.com"
          /></label>
          <label v-if="needsCode"
            >邮箱验证码<span class="input-action"
              ><input
                v-model="form.code"
                inputmode="numeric"
                autocomplete="one-time-code"
                required
                pattern="[0-9]{6}"
                maxlength="6"
                placeholder="6 位验证码"
              />
              <button
                type="button"
                :disabled="sending || !!seconds"
                @click="send"
              >
                {{
                  seconds ? seconds + " 秒" : sending ? "发送中…" : "获取验证码"
                }}
              </button></span
            ></label
          >
          <label
            >{{ recovering ? "新密码" : "密码"
            }}<span class="input-action"
              ><input
                v-model="form.password"
                :type="visible ? 'text' : 'password'"
                :autocomplete="needsCode ? 'new-password' : 'current-password'"
                required
                maxlength="200"
              />
              <button
                type="button"
                :aria-label="visible ? '隐藏密码' : '显示密码'"
                @click="visible = !visible"
              >
                {{ visible ? "隐藏" : "显示" }}
              </button></span
            ></label
          >
          <small v-if="needsCode">8–72 字节，同时包含字母和数字。</small>
          <button
            v-if="!registering && !recovering"
            type="button"
            class="recovery-link"
            @click="setRecovery(true)"
          >
            忘记密码？
          </button>
          <div v-if="error" role="alert">{{ error }}</div>
          <div v-if="notice" role="status">{{ notice }}</div>
          <button class="submit" :disabled="busy || sending">
            {{
              busy
                ? "提交中…"
                : recovering
                  ? "重置密码"
                  : creating
                    ? "注册并登录"
                    : "登录"
            }}
          </button>
        </form>
        <p v-if="recovering" class="switch">
          <button
            class="recovery-link"
            type="button"
            :disabled="busy || sending"
            @click="setRecovery(false)"
          >
            返回登录
          </button>
        </p>
        <p v-else class="switch">
          {{ registering ? "已有账号？" : "还没有账号？" }}
          <RouterLink
            :to="{
              name: registering ? 'login' : 'register',
              query: route.query,
            }"
            >{{ registering ? "登录" : "注册账号" }} →</RouterLink
          >
        </p>
      </div>
    </div>
  </section>
</template>
<style scoped lang="scss" src="./member-auth.scss"></style>
