<script setup>
import { onBeforeUnmount, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { registerMember, sendMemberCode } from "@/api/member";
import { memberDestination, useMember } from "@/composables/useMember";
import BlogIcon from "@/components/BlogIcon.vue";

const props = defineProps({ registering: Boolean });
const router = useRouter(),
  route = useRoute(),
  { login } = useMember();
const form = reactive({ email: "", password: "", code: "" });
const visible = ref(false),
  sending = ref(false),
  busy = ref(false),
  registered = ref(false);
const notice = ref(""),
  error = ref(""),
  seconds = ref(0),
  emailInput = ref(null);
let timer;
async function send() {
  if (sending.value || seconds.value || !emailInput.value.reportValidity())
    return;
  sending.value = true;
  error.value = notice.value = "";
  try {
    const result = await sendMemberCode(form.email.trim());
    seconds.value = Number(result?.retryAfter) || 60;
    notice.value = "验证码已发送，请查看邮箱。";
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
  if (props.registering && !registered.value) {
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
    if (props.registering && !registered.value) {
      await registerMember({ ...values, code: form.code.trim() });
      registered.value = true;
    }
    await login(values);
    await router.replace(memberDestination(route.query.redirect, router));
  } catch (err) {
    error.value =
      (registered.value ? "账号已创建，请重新登录。" : "") + err.message;
  } finally {
    busy.value = false;
  }
}
onBeforeUnmount(() => clearInterval(timer));
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
          {{ registering && !registered ? "创建读者账号" : "登录读者账号" }}
        </h1>
        <p>
          {{
            registering && !registered
              ? "用邮箱注册，将喜欢的文章收入自己的收藏夹。"
              : "继续阅读你的收藏，参与文章讨论。"
          }}
        </p>
        <form @submit.prevent="submit">
          <label
            >邮箱<input
              ref="emailInput"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              maxlength="254"
              placeholder="you@example.com"
          /></label>
          <label v-if="registering && !registered"
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
            >密码<span class="input-action"
              ><input
                v-model="form.password"
                :type="visible ? 'text' : 'password'"
                :autocomplete="
                  registering && !registered
                    ? 'new-password'
                    : 'current-password'
                "
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
          <small v-if="registering && !registered"
            >8–72 字节，同时包含字母和数字。</small
          >
          <div v-if="error" role="alert">{{ error }}</div>
          <div v-if="notice" role="status">{{ notice }}</div>
          <button class="submit" :disabled="busy">
            {{
              busy
                ? "提交中…"
                : registering && !registered
                  ? "注册并登录"
                  : "登录"
            }}
          </button>
        </form>
        <p class="switch">
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
