<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="min(460px, 94vw)"
    :close-on-click-modal="false"
    :before-close="close"
  >
    <el-form label-position="top" @submit.prevent="submit" :disabled="busy">
      <template v-if="kind === 'settings'">
        <el-form-item label="文档站标题" required>
          <el-input v-model="form.title" maxlength="100" />
        </el-form-item>
        <el-form-item label="文档站介绍">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" />
        </el-form-item>
      </template>
      <template v-if="kind === 'mode'">
        <el-form-item label="模式名称" required>
          <el-input v-model="form.label" maxlength="100" />
        </el-form-item>
        <el-form-item label="模式标识" required>
          <el-input v-model="form.id" :disabled="editing" placeholder="cli" />
        </el-form-item>
        <el-form-item label="图标"><IconPicker v-model="form.icon" /></el-form-item>
        <el-form-item label="介绍">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="展示状态">
          <el-select v-model="form.status">
            <el-option label="待推出" value="upcoming" />
            <el-option label="已上线" value="published" />
            <el-option label="隐藏" value="hidden" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editing" label="默认版本">
          <el-select v-model="form.defaultVersion" clearable>
            <el-option
              v-for="v in workspace.versions.filter((v) => v.status === 'published')"
              :key="v.id"
              :value="v.id"
              :label="v.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :max="99999" />
        </el-form-item>
      </template>
      <template v-if="kind === 'version'">
        <el-form-item label="版本名称" required>
          <el-input v-model="form.name" maxlength="100" />
        </el-form-item>
        <el-form-item label="版本标识" required>
          <el-input v-model="form.id" :disabled="editing" placeholder="v1.0" />
        </el-form-item>
        <el-form-item v-if="!editing" label="复制现有版本">
          <el-select v-model="form.source" clearable placeholder="空白版本">
            <el-option v-for="v in workspace.versions" :key="v.id" :value="v.id" :label="v.name" />
          </el-select>
        </el-form-item>
      </template>
      <el-form-item v-if="kind === 'group'" label="分组名称" required>
        <el-input v-model="form.title" maxlength="100" />
      </el-form-item>
    </el-form>
    <el-alert v-if="failure" :title="failure" type="error" :closable="false" />
    <template #footer>
      <el-button :disabled="busy" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="busy" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
<script setup>
import { computed, reactive, ref } from 'vue'
import IconPicker from '@/components/icons/IconPicker.vue'
import { saveDocSettings, saveMode, saveVersion, saveGroup } from '@/api/documents'

const props = defineProps({ workspace: { type: Object, required: true } })
const emit = defineEmits(['saved'])
const visible = ref(false),
  editing = ref(false),
  busy = ref(false),
  failure = ref(''),
  kind = ref('')
const form = reactive({})
const labels = { mode: '模式', version: '版本', group: '分组' }
const title = computed(() =>
  kind.value === 'settings' ? '文档站设置' : (editing.value ? '编辑' : '新建') + labels[kind.value],
)
function open(type, row) {
  kind.value = type
  editing.value = Boolean(row)
  failure.value = ''
  for (const key of Object.keys(form)) delete form[key]
  const defaults = {
    settings: { title: '', description: '' },
    mode: {
      id: '',
      label: '',
      description: '',
      icon: 'playbook',
      status: 'upcoming',
      defaultVersion: '',
      sortOrder: 0,
    },
    version: { id: '', name: '', source: '', modeId: props.workspace.modeId },
    group: { id: 0, title: '', version: props.workspace.versionId },
  }
  for (const [key, value] of Object.entries(defaults[type])) form[key] = row?.[key] ?? value
  form.revision = row?.revision || 0
  visible.value = true
}
function close(done) {
  if (!busy.value) done()
}
async function submit() {
  if (busy.value) return
  busy.value = true
  failure.value = ''
  try {
    const payload = { ...form }
    for (const key of ['title', 'label', 'id', 'name'])
      if (typeof payload[key] === 'string') payload[key] = payload[key].trim()
    if (!String(payload.title || payload.label || payload.name || '').trim())
      throw new Error('请填写名称')
    if (
      ['version', 'mode'].includes(kind.value) &&
      !/^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$/.test(payload.id)
    )
      throw new Error('标识需要 1 至 80 个字母、数字、点、短横线或下划线')
    if (kind.value === 'mode' && payload.status === 'published' && !payload.defaultVersion)
      throw new Error('上线模式前请先发布并选择默认版本')
    const save = {
      settings: saveDocSettings,
      mode: saveMode,
      version: saveVersion,
      group: saveGroup,
    }[kind.value]
    await save(payload)
    visible.value = false
    emit('saved', { kind: kind.value, id: form.id, created: !editing.value })
  } catch (error) {
    failure.value =
      error.response?.status === 409
        ? '数据已被更新。当前输入已保留，请核对最新版本后重试。'
        : error.message
  } finally {
    busy.value = false
  }
}
defineExpose({ open })
</script>
