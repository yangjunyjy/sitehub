<script setup>
import { reactive, ref, watch } from 'vue'
import { api } from '../api'
defineProps({ listen: String })
const form = reactive({ scheme: 'https', method: 'GET', host: '', path: '/' })
const result = ref(null),
  error = ref(''),
  busy = ref(false)
watch(form, () => {
  result.value = null
  error.value = ''
})
async function match() {
  busy.value = true
  result.value = null
  error.value = ''
  try {
    result.value = await api('match', {
      ...form,
      method: form.method.trim().toUpperCase(),
      host: form.host.trim(),
      path: form.path.trim(),
    })
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <details class="card runtime-panel">
    <summary>试匹配 HTTP 请求</summary>
    <p>
      使用本机生效规则检查网关收到的 Host、方法和路径。<span v-if="listen"
        >HTTP 监听 {{ listen }}。</span
      >
    </p>
    <form class="match-form" @submit.prevent="match">
      <fieldset :disabled="busy">
        <label
          >协议<select v-model="form.scheme" aria-label="协议">
            <option>https</option>
            <option>http</option>
          </select></label
        >
        <label
          >请求方法<input
            aria-label="试匹配方法"
            v-model="form.method"
            list="http-methods"
            required
        /></label>
        <datalist id="http-methods">
          <option
            v-for="m in ['GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS']"
            :key="m"
          >
            {{ m }}
          </option>
        </datalist>
        <label
          >请求 Host<input v-model="form.host" required placeholder="blog.sitehub.local:8082"
        /></label>
        <label
          >请求路径<input aria-label="试匹配路径" v-model="form.path" required placeholder="/"
        /></label>
        <button class="primary">{{ busy ? '匹配中…' : '试匹配' }}</button>
      </fieldset>
    </form>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div v-if="result" role="status" class="match-result">
      <strong>{{ result.matched ? '命中 ' + result.route_id : '未命中规则' }}</strong> · 生效修订
      {{ result.revision }}
      <p v-if="result.matched">
        路径 {{ result.pattern }} · 优先级 {{ result.priority }} ·
        {{ result.upstream ? '转发到 ' + result.upstream : '响应插件 ' + (result.terminal || '')
        }}<span v-if="result.profile"> · 策略 {{ result.profile }}</span>
      </p>
      <p v-if="result.reason">{{ result.reason }}</p>
    </div>
  </details>
</template>
