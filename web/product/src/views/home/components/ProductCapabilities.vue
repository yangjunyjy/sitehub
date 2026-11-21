<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
defineProps({ items: { type: Array, required: true } })
</script>
<template>
  <ul v-if="items.length" class="capability-list">
    <li v-for="item in items" :key="item.key" :id="'capability-' + item.key">
      <ResourceIcon class="capability-icon" :name="item.icon || item.key" :size="17" />
      <div>
        <h3>{{ item.title }}</h3>
        <p v-if="item.description">{{ item.description }}</p>
        <p v-if="item.tags?.length" class="capability-tags">{{ item.tags.join(' · ') }}</p>
      </div>
    </li>
  </ul>
</template>
<style scoped>
.capability-list { display: grid; gap: 14px; margin: 0; padding: 0 0 18px; list-style: none; }
.capability-list li { display: grid; grid-template-columns: 18px minmax(0, 1fr); gap: 9px; }
.capability-icon { color: var(--workbench-accent); margin-top: 3px; }
h3 { margin: 0; font-size: 14px; font-weight: 600; }
p { margin: 3px 0 0; color: var(--workbench-muted); font-size: 13px; line-height: 1.7; }
.capability-tags { margin-top: 3px; font-size: 11px; }
</style>
