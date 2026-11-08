<script setup>
defineProps({ eyebrow: String, title: String, description: String });
</script>
<template>
  <header class="journal-page-heading">
    <div>
      <p class="site-kicker">{{ eyebrow }}</p>
      <h1>{{ title }}</h1>
      <p class="journal-description">{{ description }}</p>
    </div>
    <slot />
  </header>
</template>
<style scoped>
.journal-page-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  padding: 48px 0 32px;
  border-bottom: 1px solid var(--deploy-line);
  margin-bottom: 32px;
}
h1 {
  font-size: 38px;
  letter-spacing: -1px;
  margin: 10px 0 12px;
  line-height: 1.4;
}
.journal-description {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 14px;
}
@media (max-width: 600px) {
  .journal-page-heading {
    padding: 32px 0 24px;
    flex-wrap: wrap;
    margin-bottom: 24px;
  }
  h1 {
    font-size: 30px;
  }
  .journal-description {
    font-size: 13px;
  }
}
</style>
