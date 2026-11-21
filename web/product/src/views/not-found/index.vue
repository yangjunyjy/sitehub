<template>
  <section class="not-found">
    <div class="site-shell">
      <p class="site-kicker">404 / Page not found</p>
      <h1>这个页面<br />不在交付路径上</h1>
      <p>你访问的地址不存在，可能已被迁移或移除。</p>
      <router-link class="site-button site-button--primary" :to="{ name: 'home' }"
        >返回首页</router-link
      >
    </div>
  </section>
</template>

<style lang="scss" scoped>
.not-found {
  display: grid;
  place-items: center;
  min-height: 68vh;
  padding: 90px 0;
  background: var(--deploy-paper);
  text-align: center;
}
.not-found h1 {
  margin: 20px 0 24px;
  font-size: clamp(28px, 3vw, 36px);
  line-height: 1.4;
  letter-spacing: -0.03em;
}
.not-found .site-shell > p:not(.site-kicker) {
  margin: 0 0 30px;
  color: var(--deploy-muted);
  font-size: 15px;
}
</style>
