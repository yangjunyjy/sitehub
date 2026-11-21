<script setup>
defineProps({ steps: { type: Array, required: true } })
</script>
<template>
  <ol v-if="steps.length" class="delivery-list">
    <li v-for="step in steps" :key="step.key">
      <h3>{{ step.title }}</h3>
      <p v-if="step.description">{{ step.description }}</p>
    </li>
  </ol>
</template>
<style scoped>
.delivery-list { list-style: none; margin: 0; padding: 0 0 17px 3px; }
.delivery-list li { position: relative; padding: 0 0 12px 14px; border-left: 1px solid var(--workbench-line); }
.delivery-list li:last-child { padding-bottom: 0; border-left-color: transparent; }
.delivery-list li::before { content: ''; position: absolute; inset: 7px auto auto -3px; width: 5px; height: 5px; border-radius: 50%; background: var(--workbench-accent); }
h3 { margin: 0; font-size: 14px; font-weight: 600; }
p { margin: 3px 0 0; color: var(--workbench-muted); font-size: 13px; line-height: 1.7; }
</style>
