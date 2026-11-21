<script setup>
import HomeDemo from './HomeDemo.vue'
defineProps({ home: { type: Object, required: true }, brand: { type: String, required: true } })
</script>
<template>
  <section class="home-hero">
    <div class="site-shell home-hero__inner">
      <header class="home-hero__copy">
        <p class="hero-brand">{{ brand }}</p>
        <h1>{{ home.title || '部署与运维工作台' }}<span v-if="home.highlight">{{ home.highlight }}</span></h1>
        <p class="hero-description">{{ home.description }}</p>
        <div class="hero-actions">
          <router-link class="site-button site-button--primary" :to="{ name: 'download' }">{{ home.primaryAction || '下载产品' }}</router-link>
          <router-link class="site-button site-button--outline" :to="{ name: 'docs', params: { id: '' } }">阅读文档 <span aria-hidden="true">→</span></router-link>
        </div>
        <ul v-if="home.meta?.length" class="hero-meta"><li v-for="item in home.meta" :key="item">{{ item }}</li></ul>
      </header>
      <HomeDemo />
    </div>
  </section>
</template>
<style lang="scss" scoped>
.home-hero {
  padding-block: clamp(44px, 8vw, 112px);
  border-bottom: 1px solid var(--deploy-line);
}
.home-hero__inner {
  display: grid;
  grid-template-columns: minmax(0, 0.84fr) minmax(0, 1.16fr);
  gap: clamp(32px, 5vw, 72px);
  align-items: center;
}
.home-hero__copy { min-width: 0; }
.hero-brand { margin: 0 0 16px; color: var(--deploy-primary); font-weight: 650; font-size: 20px; letter-spacing: -.035em; }
h1 { margin: 0; font-size: clamp(36px, 4.1vw, 64px); line-height: 1.18; font-weight: 680; letter-spacing: -.045em; text-wrap: balance; }
h1 span { display: block; color: var(--deploy-primary); }
.hero-description { max-width: 34em; margin: 22px 0 0; color: var(--deploy-muted); line-height: 1.9; font-size: 16px; }
.hero-actions, .hero-meta { display: flex; flex-wrap: wrap; }
.hero-actions { gap: 10px; margin-top: 28px; }
.hero-meta { padding: 0; margin: 28px 0 0; list-style: none; gap: 7px 0; color: var(--deploy-muted); font: 12px/1.6 var(--site-mono); }
.hero-meta li + li { border-left: 1px solid var(--deploy-line); margin-left: 12px; padding-left: 12px; }
@media (max-width: 840px) {
  .home-hero__inner { grid-template-columns: minmax(0, 1fr); }
  .hero-description { margin-top: 14px; }
  .hero-actions, .hero-meta { margin-top: 18px; }
}
</style>
