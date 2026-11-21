<script setup>
import { ref, onMounted } from 'vue'
import { Close } from '@element-plus/icons-vue'
const props = defineProps({
  title: String,
  confirm: String,
  busy: Boolean,
  error: String,
  danger: Boolean,
  wide: Boolean,
})
const emit = defineEmits(['close', 'submit'])
const panel = ref()
onMounted(() => panel.value.showModal())
function dismiss() {
  if (!props.busy) emit('close')
}
</script>
<template>
  <dialog ref="panel" :aria-label="title" :class="{ wide }" @cancel.prevent="dismiss">
    <form @submit.prevent="!busy && $emit('submit')">
      <header>
        <h2>{{ title }}</h2>
        <button class="icon" aria-label="关闭弹窗" type="button" :disabled="busy" @click="dismiss">
          <Close />
        </button>
      </header>
      <div class="dialog-body">
        <slot />
        <p v-if="error" role="alert" class="error">{{ error }}</p>
      </div>
      <footer>
        <button v-if="confirm" type="button" :disabled="busy" @click="dismiss">取消</button
        ><button :class="['primary', { danger }]" :disabled="busy">
          {{ busy ? '处理中…' : confirm || '关闭' }}
        </button>
      </footer>
    </form>
  </dialog>
</template>
