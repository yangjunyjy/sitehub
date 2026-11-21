<script setup>
import { onBeforeUnmount, reactive, ref, watch } from 'vue'
import { Connection, Key, Lock, Message, User, View, Hide } from '@element-plus/icons-vue'
import SliderCaptcha from '@shared/auth/SliderCaptcha.vue'
import { api, generate, verifyChallenge } from '../api'
const emit = defineEmits(['authenticated'])
const form = reactive({ username: '', email: '', password: '', code: '' })
const captcha = ref(false),
  sending = ref(false),
  busy = ref(false),
  showPassword = ref(false)
const error = ref(''),
  feedback = ref(''),
  seconds = ref(0),
  emailInput = ref()
let timer, recipient, controller
watch(
  () => form.email,
  () => {
    clearInterval(timer)
    seconds.value = 0
    feedback.value = ''
    form.code = ''
    captcha.value = false
    controller?.abort()
  },
)
onBeforeUnmount(() => {
  clearInterval(timer)
  controller?.abort()
})
function challenge() {
  if (!emailInput.value.reportValidity()) return
  recipient = form.email.trim()
  captcha.value = true
  error.value = feedback.value = ''
}
async function send(captchaKey) {
  if (sending.value || recipient !== form.email.trim()) return
  sending.value = true
  controller = new AbortController()
  try {
    await api('auth/code', { email: recipient, captchaKey }, { signal: controller.signal })
    feedback.value = '验证码已发送，请查收邮箱'
    const until = Date.now() + 60000
    seconds.value = 60
    timer = setInterval(() => {
      seconds.value = Math.max(0, Math.ceil((until - Date.now()) / 1000))
      if (!seconds.value) clearInterval(timer)
    }, 500)
  } catch (e) {
    if (e.name !== 'AbortError') error.value = e.message
  } finally {
    sending.value = false
  }
}
async function login() {
  if (busy.value) return
  busy.value = true
  error.value = ''
  try {
    await api('auth/login', { ...form, email: form.email.trim(), username: form.username.trim() })
    emit('authenticated')
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <main class="login-page">
    <div class="login-frame">
      <section class="login-art" aria-hidden="true">
        <div class="login-art-brand"><Connection /> Gateway</div>
        <div class="network-art">
          <div class="orbit orbit-a"></div>
          <div class="orbit orbit-b"></div>
          <div class="orbit orbit-c"></div>
          <div class="network-core"><Connection /></div>
          <i v-for="n in 4" :key="n" :class="'node node-' + n"></i>
        </div>
      </section>
      <form class="login-form" @submit.prevent="login">
        <h1>登录网关控制台</h1>
        <label
          >用户名<span class="input-icon"
            ><User /><input
              v-model="form.username"
              autocomplete="username"
              required
              placeholder="管理员账号" /></span
        ></label>
        <label
          >邮箱<span class="input-icon"
            ><Message /><input
              ref="emailInput"
              v-model="form.email"
              autocomplete="email"
              type="email"
              required
              placeholder="账号绑定的邮箱"
              :disabled="sending || captcha" /></span
        ></label>
        <div>
          <label for="password">密码</label>
          <div class="input-icon">
            <Lock /><input
              id="password"
              v-model="form.password"
              autocomplete="current-password"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="输入密码"
            /><button
              type="button"
              class="icon"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              @click="showPassword = !showPassword"
            >
              <component :is="showPassword ? Hide : View" />
            </button>
          </div>
        </div>
        <div>
          <label for="email-code">邮箱验证码</label>
          <div class="code-row">
            <span class="input-icon"
              ><Key /><input
                id="email-code"
                v-model="form.code"
                autocomplete="one-time-code"
                inputmode="numeric"
                maxlength="6"
                pattern="[0-9]{6}"
                required
                placeholder="6 位验证码" /></span
            ><button
              type="button"
              :disabled="!!seconds || sending || captcha || busy"
              @click="challenge"
            >
              {{ seconds ? seconds + ' 秒后重发' : sending ? '正在发送' : '获取验证码' }}
            </button>
          </div>
        </div>
        <p v-if="feedback" role="status">{{ feedback }}</p>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
        <button class="primary login-submit" :disabled="busy">
          {{ busy ? '正在登录…' : '登录' }}
        </button>
      </form>
    </div>
    <SliderCaptcha
      v-model="captcha"
      :generate="generate"
      :verify-challenge="verifyChallenge"
      @success="send"
    />
  </main>
</template>
