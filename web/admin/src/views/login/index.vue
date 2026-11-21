<template>
  <div class="login-container">
    <a class="login-brand" href="/"><AdminBrand /> <strong>SiteHub Admin</strong></a>
    <aside class="login-illustration" aria-hidden="true">
      <svg viewBox="0 0 600 420" fill="none">
        <ellipse cx="290" cy="379" rx="240" ry="19" fill="#e6effa" />
        <rect
          x="65"
          y="50"
          width="446"
          height="290"
          rx="12"
          fill="white"
          stroke="#d8e6f8"
          stroke-width="2"
        />
        <path d="M65 89h446" stroke="#e6effa" stroke-width="2" />
        <circle cx="87" cy="70" r="5" fill="#409eff" />
        <circle cx="105" cy="70" r="5" fill="#b3d8ff" />
        <circle cx="123" cy="70" r="5" fill="#d9ecff" />
        <rect x="82" y="106" width="93" height="215" rx="5" fill="#f3f8ff" />
        <rect x="95" y="125" width="67" height="23" rx="4" fill="#409eff" />
        <path
          d="M101 169h49m-49 26h42m-42 26h49m-49 26h36"
          stroke="#b3cfee"
          stroke-width="7"
          stroke-linecap="round"
        />
        <rect x="194" y="106" width="88" height="60" rx="5" fill="#ecf5ff" />
        <rect x="297" y="106" width="88" height="60" rx="5" fill="#f0f9eb" />
        <rect x="400" y="106" width="88" height="60" rx="5" fill="#fdf6ec" />
        <path
          d="M211 126h23m-23 18h47m56-18h23m-23 18h47m56-18h23m-23 18h47"
          stroke="#9bb4d3"
          stroke-width="6"
          stroke-linecap="round"
        />
        <path d="M202 285h276M202 250h276M202 215h276" stroke="#e6effa" />
        <path d="m205 272 42-24 44 11 49-49 42 20 49-42 44 12v103H205Z" fill="#ecf5ff" />
        <path
          d="m205 272 42-24 44 11 49-49 42 20 49-42 44 12"
          stroke="#409eff"
          stroke-width="4"
          stroke-linejoin="round"
        />
        <rect
          x="418"
          y="267"
          width="119"
          height="97"
          rx="10"
          fill="white"
          stroke="#b3d8ff"
          stroke-width="2"
        />
        <circle cx="477" cy="307" r="20" fill="#ecf5ff" />
        <path
          d="m466 307 8 8 15-17"
          stroke="#409eff"
          stroke-width="4"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path d="M451 343h51" stroke="#d9e8fa" stroke-width="6" stroke-linecap="round" />
        <path
          d="M241 340v31h91v-31M222 375h128"
          stroke="#b3cfee"
          stroke-width="7"
          stroke-linecap="round"
        />
      </svg>
      <h2>内容与运维，井然有序</h2>
      <p>管理内容、发布文档，让每一次交付更从容。</p>
    </aside>
    <div class="login-box">
      <div class="login-header container">
        <AdminBrand class="login-brand-mark" />
        <div class="right">
          <h1>欢迎登录</h1>
          <p class="login-subtitle">登录 SiteHub 管理后台</p>
        </div>
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form">
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            :prefix-icon="User"
            size="large"
            clearable
          />
        </el-form-item>
        <el-form-item prop="email">
          <el-input
            v-model="loginForm.email"
            placeholder="请输入邮箱"
            :prefix-icon="Message"
            size="large"
            :disabled="showCaptcha || codeloading"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item prop="code">
          <div class="code-input-group">
            <el-input
              v-model="loginForm.code"
              placeholder="请输入验证码"
              :prefix-icon="Key"
              size="large"
              maxlength="6"
              clearable
            />
            <el-button
              :disabled="codeCountdown > 0 || !loginForm.email || showCaptcha || codeloading"
              @click="sendCode"
              size="large"
              :loading="codeloading"
              class="code-btn"
            >
              {{
                codeloading
                  ? '正在发送'
                  : codeCountdown > 0
                    ? `${codeCountdown}s 后重发`
                    : '获取验证码'
              }}
            </el-button>
          </div>
          <p
            v-if="sendFeedback"
            class="send-feedback"
            :class="{ 'is-error': sendFailed }"
            role="status"
          >
            {{ sendFeedback }}
          </p>
        </el-form-item>

        <el-form-item>
          <div class="form-footer">
            <el-checkbox v-model="loginForm.remember" size="large">记住账号</el-checkbox>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-btn"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 滑块验证组件 -->
      <SliderCaptcha v-model="showCaptcha" @success="onCaptchaSuccess" />
    </div>
  </div>
</template>

<script setup>
import AdminBrand from '@/components/AdminBrand.vue'
import { ref, reactive, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, Message, Key, User } from '@element-plus/icons-vue'
import { rbacLogin, sendEmailCode } from '@/api'
import { useRbacStore } from '@/store/rbac'
import SliderCaptcha from '@/components/SliderCaptcha.vue'

const router = useRouter()
const route = useRoute()
const rbacStore = useRbacStore()
const loginFormRef = ref(null)
const loading = ref(false)
const codeloading = ref(false)
const codeCountdown = ref(0)
const showCaptcha = ref(false)
const sendFeedback = ref('')
const sendFailed = ref(false)
let pendingEmail = ''
let countdownTimer
let sendController
let disposed = false

const loginForm = reactive({
  username: '',
  email: '',
  password: '',
  code: '',
  remember: false,
})

const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 4, max: 20, message: '请输入正确的用户名', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码长度为6位', trigger: 'blur' },
  ],
}

// 组件挂载时，从localStorage读取记住的密码
onMounted(() => {
  const savedLogin = localStorage.getItem('rememberedLogin')
  if (savedLogin) {
    let parsed
    try {
      parsed = JSON.parse(savedLogin)
    } catch {
      localStorage.removeItem('rememberedLogin')
      return
    }
    if (parsed.remember) {
      loginForm.username = parsed.username
      loginForm.email = parsed.email
      loginForm.remember = true
    }
  }
})

// 滑块验证成功回调
const onCaptchaSuccess = (captchaKey) => {
  const email = pendingEmail
  pendingEmail = ''
  if (email && email === loginForm.email.trim()) doSendCode(email, captchaKey)
}

const sendCode = async () => {
  if (codeloading.value || codeCountdown.value > 0 || showCaptcha.value) return
  try {
    await loginFormRef.value.validateField('email')
    pendingEmail = loginForm.email.trim()
    sendFeedback.value = ''
    showCaptcha.value = true
  } catch {
    /* 表单在邮箱输入框下显示校验信息。 */
  }
}

const doSendCode = async (email, captchaKey) => {
  if (codeloading.value || disposed) return
  codeloading.value = true
  sendFailed.value = false
  sendFeedback.value = ''
  sendController = new AbortController()
  try {
    const data = await sendEmailCode(email, captchaKey, { signal: sendController.signal, silentError: true })
    if (disposed) return
    if (data.code !== 0) throw new Error('send failed')
    sendFeedback.value = '验证码已发送，请查收邮箱'
    codeCountdown.value = 60
    clearInterval(countdownTimer)
    const deadline = Date.now() + 60000
    countdownTimer = setInterval(() => {
      codeCountdown.value = Math.max(0, Math.ceil((deadline - Date.now()) / 1000))
      if (!codeCountdown.value) clearInterval(countdownTimer)
    }, 1000)
  } catch (error) {
    if (disposed) return
    sendFailed.value = true
    sendFeedback.value =
      error.response?.status === 429 ? '发送过于频繁，请稍后重试' : '邮件未能发送，请稍后重新验证'
  } finally {
    if (!disposed) codeloading.value = false
  }
}

watch(
  () => loginForm.email,
  () => {
    pendingEmail = ''
    showCaptcha.value = false
    clearInterval(countdownTimer)
    codeCountdown.value = 0
    loginForm.code = ''
    sendFeedback.value = ''
  },
)

onBeforeUnmount(() => {
  disposed = true
  clearInterval(countdownTimer)
  sendController?.abort()
})

const handleLogin = async () => {
  if (!loginFormRef.value || loading.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (!valid || loading.value) return

    loading.value = true

    try {
      const res = await rbacLogin({
        username: loginForm.username,
        password: loginForm.password,
        email: loginForm.email,
        code: loginForm.code,
      })

      const { token, user } = res.data
      if (typeof token !== 'string' || !token) throw new Error('登录响应缺少会话，请重试')

      // 记住账号
      if (loginForm.remember) {
        localStorage.setItem(
          'rememberedLogin',
          JSON.stringify({
            username: loginForm.username,
            email: loginForm.email,
            remember: true,
          }),
        )
      } else {
        localStorage.removeItem('rememberedLogin')
      }

      // 清理旧用户的菜单和权限数据
      rbacStore.resetAllData()

      router.resetMenus()

      sessionStorage.setItem('token', token)
      sessionStorage.setItem('userInfo', JSON.stringify(user))

      const redirect = route.query.redirect
      const navigation = await router.replace(typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//') && !redirect.startsWith('/login') ? redirect : '/dashboard')
      if (!navigation && router.currentRoute.value.matched.some(record => record.name === 'dashboard')) ElMessage.success('登录成功')
    } catch (error) {
      console.error('登录失败:', error)
      ElMessage.error(error.response?.data?.message || error.message || '登录失败，请重试')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss" src="./login.scss"></style>
