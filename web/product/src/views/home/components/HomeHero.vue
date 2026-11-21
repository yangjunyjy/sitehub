<script setup>
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import ParticleScene from "./ParticleScene.vue";
defineProps({
  home: { type: Object, required: true },
  brand: { type: String, required: true },
});
</script>
<template>
  <section class="home-hero" aria-labelledby="product-name">
    <div class="exhibit">
      <header class="hero-copy">
        <h1 id="product-name">{{ brand }}</h1>
        <p v-if="home.title || home.highlight" class="category">
          {{ home.title }}{{ home.highlight }}
        </p>
        <p v-if="home.description" class="description">
          {{ home.description }}
        </p>
        <div class="hero-actions">
          <RouterLink class="download" :to="{ name: 'download' }"
            ><ResourceIcon name="operations/download" :size="17" />{{
              home.primaryAction || "下载"
            }}</RouterLink
          >
          <RouterLink :to="{ name: 'docs', params: { id: '' } }"
            ><ResourceIcon name="data/book" :size="16" />{{
              home.secondaryAction || "文档"
            }}</RouterLink
          >
        </div>
      </header>
      <ParticleScene
        :key="home.particleShape"
        :shape="home.particleShape"
        :autoplay="home.particleAutoplay"
        :interactive="home.particleInteractive"
      />
    </div>
  </section>
</template>
<style scoped lang="scss">
.home-hero {
  --exhibit-paper: #0b1428;
  --exhibit-ink: #eef6ff;
  --exhibit-muted: #a5b6d1;
  --exhibit-line: #536481;
  --exhibit-accent: #70e2ff;
  color: var(--exhibit-ink);
  background-color: var(--exhibit-paper);
  background-image: radial-gradient(ellipse at 74% 48%, #25416d4d, transparent 58%);
  border-bottom: 1px solid #6988ae42;
}
:global(.dark .home-hero) {
  --exhibit-paper: #040914;
  --exhibit-ink: #e7f0ff;
  --exhibit-muted: #a1b3ce;
  --exhibit-line: #445877;
  --exhibit-accent: #80d4ff;
}
.exhibit {
  width: min(1480px, calc(100% - clamp(64px, 10vw, 180px)));
  margin-inline: auto;
  display: grid;
  grid-template-columns: minmax(0, 0.82fr) minmax(0, 1.18fr);
  align-items: center;
  min-height: clamp(480px, 75svh, 820px);
  padding: clamp(28px, 4.4vw, 70px) 0;
}
.hero-copy {
  padding-block: 28px;
}
h1 {
  margin: 0;
  color: var(--exhibit-ink);
  font-size: clamp(68px, 8.65vw, 146px);
  font-weight: 620;
  letter-spacing: -0.078em;
  line-height: 0.98;
  overflow-wrap: anywhere;
}
.category {
  margin: 20px 0 0 4px;
  font-size: clamp(18px, 1.7vw, 26px);
  line-height: 1.5;
  letter-spacing: 0.025em;
}
.description {
  max-width: 32em;
  margin: 14px 0 0 4px;
  color: var(--exhibit-muted);
  overflow-wrap: anywhere;
}
.hero-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin: 36px 0 0 4px;
}
.hero-actions a {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  padding: 10px 19px;
  border-radius: 5px;
  border: 1px solid var(--exhibit-line);
  font-size: 14px;
  font-weight: 550;
  transition:
    transform 0.22s,
    border-color 0.22s;
}
.hero-actions a:hover {
  transform: translateY(-2px);
  border-color: var(--exhibit-accent);
}
.hero-actions .download {
  background: #c1efff;
  border-color: #c1efff;
  color: var(--exhibit-paper);
}
.hero-actions a:focus-visible {
  outline: 2px solid var(--exhibit-accent);
}
@media (max-width: 720px) {
  .exhibit {
    width: calc(100% - 40px);
    grid-template-columns: 1fr;
    min-height: auto;
    gap: 2px;
    padding: 40px 0 24px;
  }
  .hero-copy {
    padding: 0;
  }
  h1 {
    font-size: clamp(64px, 16vw, 100px);
  }
  .category {
    margin-top: 12px;
  }
  .hero-actions {
    margin-top: 19px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .hero-actions a {
    transition: none;
  }
  .hero-actions a:hover {
    transform: none;
  }
}
</style>
