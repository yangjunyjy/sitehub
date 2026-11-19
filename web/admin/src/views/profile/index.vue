<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getUserInfo, changePassword } from '@/api/auth'

const user = ref(null)
const loading = ref(true)
const error = ref('')
const activeTab = ref('info')
const formRef = ref()
const submitting = ref(false)
const form = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const rules = {
  oldPassword: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 12, message: '新密码至少 12 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) =>
        callback(value === form.newPassword ? undefined : new Error('两次输入的密码不一致')),
      trigger: 'blur',
    },
  ],
}

async function loadUser() {
  loading.value = true
  error.value = ''
  try {
    const response = await getUserInfo()
    user.value = response.data
  } catch {
    error.value = '账号信息加载失败，请重试。'
  } finally {
    loading.value = false
  }
}

async function submitPassword() {
  if (submitting.value || !(await formRef.value.validate().catch(() => false))) return
  submitting.value = true
  try {
    await changePassword({
      email: user.value.email,
      oldPassword: form.oldPassword,
      newPassword: form.newPassword,
    })
    ElMessage.success('密码修改成功')
    formRef.value.resetFields()
  } catch {
    ElMessage.error('修改失败，请检查原密码后重试')
  } finally {
    submitting.value = false
  }
}

onMounted(loadUser)
</script>

<template>
  <el-card class="profile-page" shadow="never" v-loading="loading">
    <template #header><h2>个人中心</h2></template>
    <el-alert v-if="error" type="error" :closable="false" :title="error">
      <el-button link type="primary" @click="loadUser">重新加载</el-button>
    </el-alert>
    <el-tabs v-else-if="user" v-model="activeTab">
      <el-tab-pane label="账号信息" name="info">
        <div class="identity">
          <el-avatar :size="64" :src="user.avatar">{{
            (user.nickname || user.username || '').slice(0, 1)
          }}</el-avatar>
          <div>
            <h3>{{ user.nickname || user.username }}</h3>
            <p>{{ user.email }}</p>
          </div>
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="用户名">{{ user.username }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ user.nickname || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="登录邮箱">{{ user.email || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="角色">{{
            (user.roles || [])
              .map((role) => role.role_name || role.roleName || role.name)
              .filter(Boolean)
              .join('、') || '未分配'
          }}</el-descriptions-item>
        </el-descriptions>
        <p class="profile-help">账号资料由管理员在“用户管理”中维护。</p>
      </el-tab-pane>
      <el-tab-pane label="修改密码" name="password">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="password-form"
          @submit.prevent="submitPassword"
        >
          <el-form-item label="原密码" prop="oldPassword"
            ><el-input
              v-model="form.oldPassword"
              type="password"
              show-password
              autocomplete="current-password"
          /></el-form-item>
          <el-form-item label="新密码" prop="newPassword"
            ><el-input
              v-model="form.newPassword"
              type="password"
              show-password
              autocomplete="new-password"
          /></el-form-item>
          <el-form-item label="确认新密码" prop="confirmPassword"
            ><el-input
              v-model="form.confirmPassword"
              type="password"
              show-password
              autocomplete="new-password"
          /></el-form-item>
          <el-button
            type="primary"
            native-type="submit"
            :loading="submitting"
            :disabled="!user.email"
            >更新密码</el-button
          >
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<style lang="scss" scoped>
.profile-page {
  max-width: 1000px;
}
.profile-page h2 {
  margin: 0;
  font-size: 20px;
}
.identity {
  display: flex;
  align-items: center;
  gap: 18px;
  margin: 20px 0 28px;
}
.identity h3 {
  margin: 0 0 7px;
}
.identity p,
.profile-help {
  color: #606266;
  font-size: 14px;
}
.identity p {
  margin: 0;
}
.password-form {
  max-width: 480px;
  padding: 16px 0;
}
.code-field {
  display: flex;
  gap: 12px;
  width: 100%;
}
.code-field .el-input {
  min-width: 0;
}
.profile-help {
  margin-top: 18px;
}
</style>
