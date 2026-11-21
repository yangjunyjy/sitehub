<script setup>
import ResourceIcon from '@shared/icons/ResourceIcon.vue'
defineProps({ items: { type: Array, default: () => [] }, linked: Boolean });
</script>
<template>
  <div class="product-grid">
    <article
      v-for="item in items"
      :key="item.key"
      class="product-card"
      :id="`capability-${item.key}`"
    >
      <div class="product-card__top">
        <span class="product-card__icon"><ResourceIcon :name="item.icon || item.key" :size="24" /></span
        ><span class="product-card__index">/ {{ item.index }}</span>
      </div>
      <div class="product-card__body">
        <h3>{{ item.title }}</h3>
        <p>{{ item.description }}</p>
        <div class="tag-row">
          <span v-for="tag in item.tags" :key="tag">{{ tag }}</span>
        </div>
        <router-link v-if="linked" class="capability-link" :to="{ name: 'docs', params: { id: '' } }"
          >查看使用文档 <span aria-hidden="true">→</span></router-link
        >
      </div>
    </article>
  </div>
</template>
<style lang="scss" scoped>
.product-card {
  scroll-margin-top: 110px;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
}
.product-card:hover {
  border-color: var(--deploy-line);
  box-shadow: 0 8px 24px #00000012;
}
.capability-link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
  color: var(--deploy-primary);
  font-size: 13px;
  font-weight: 600;
  min-height: 32px;
}
.capability-link:hover {
  gap: 14px;
}
</style>
