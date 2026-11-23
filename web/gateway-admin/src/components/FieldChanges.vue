<script setup>
defineProps({
  changes: Array,
  left: { type: String, default: '当前值' },
  right: { type: String, default: '目标值' },
})
function display(value) {
  if (value === undefined) return '未设置'
  return typeof value === 'string' && value !== '' ? value : JSON.stringify(value)
}
</script>
<template>
  <table class="field-diff">
    <thead>
      <tr>
        <th>字段</th>
        <th>{{ left }}</th>
        <th>{{ right }}</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="f in changes" :key="f.path">
        <th scope="row">{{ f.path || '配置' }}</th>
        <td>{{ display(f.before) }}</td>
        <td>{{ display(f.after) }}</td>
      </tr>
    </tbody>
  </table>
</template>
