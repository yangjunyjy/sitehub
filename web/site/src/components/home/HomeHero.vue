<script setup>
import BrandMark from "@/components/product/BrandMark.vue";
import HomeDemo from "./HomeDemo.vue";
defineProps({ home: { type: Object, required: true }, signedIn: Boolean });
</script>
<template>
  <section class="home-hero">
    <div class="home-hero__inner">
      <div class="home-hero__copy">
        <BrandMark class="hero-mark" />
        <h1><strong>De</strong>ploy</h1>
        <p class="hero-tagline">为开发者而生的部署与运维平台</p>
        <p class="hero-description">{{ home.description }}</p>
        <div class="hero-actions">
          <router-link class="site-button site-button--primary" to="/docs"
            >快速开始 <span aria-hidden="true">→</span></router-link
          ><router-link class="site-button site-button--outline" to="/download"
            >下载 Deploy</router-link
          >
        </div>
        <div class="hero-meta">
          <span v-for="item in home.meta" :key="item">{{ item }}</span>
        </div>
      </div>
      <HomeDemo />
    </div>
    <nav class="home-hero__rail" aria-label="首页章节">
      <a href="#capabilities">探索产品能力 <span aria-hidden="true">↓</span></a
      ><a href="#delivery">了解交付流程</a
      ><router-link to="/download">下载 Deploy</router-link>
    </nav>
  </section>
</template>
<style scoped>
.home-hero {
  background: var(--deploy-white);
  padding: 60px 40px 0;
}
.home-hero__inner {
  max-width: 1440px;
  margin: auto;
  display: grid;
  grid-template-columns: minmax(0, 3fr) minmax(0, 4fr);
  gap: 52px;
  align-items: center;
  min-height: 650px;
  padding: 0 24px 40px;
}
.home-hero__copy {
  text-align: center;
  padding-bottom: 40px;
}
.hero-mark {
  width: 174px;
  height: 174px;
  margin: 0 auto 6px;
  filter: drop-shadow(0 14px 16px #00000010);
}
h1 {
  font-size: 76px;
  font-weight: 300;
  letter-spacing: -4px;
  line-height: 1.1;
  margin: 0;
  color: #238da9;
}
h1 strong {
  font-weight: 600;
}
.hero-tagline {
  font-size: 24px;
  color: var(--deploy-muted);
  margin: 26px 0 0;
  line-height: 1.5;
}
.hero-description {
  font-size: 14px;
  line-height: 1.9;
  color: var(--deploy-faint);
  max-width: 420px;
  margin: 18px auto 0;
}
.hero-actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 12px;
  margin: 28px 0 0;
}
.hero-actions .site-button {
  font-size: 15px;
  padding: 11px 19px;
  border-radius: 10px;
}
.hero-meta {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 7px 16px;
  margin-top: 24px;
  color: var(--deploy-faint);
  font: 11px var(--site-mono);
}
.home-hero__rail {
  max-width: 1120px;
  margin: 0 auto;
  display: flex;
  justify-content: center;
  gap: 36px;
  padding: 20px 0 30px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.home-hero__rail a:hover {
  color: var(--deploy-primary);
}
.home-hero__rail span {
  margin-left: 8px;
}
@media (min-width: 1600px) {
  .home-hero__inner {
    min-height: 760px;
    gap: 80px;
  }
  .hero-mark {
    width: 205px;
    height: 205px;
  }
}
@media (max-width: 1080px) {
  .home-hero {
    padding: 48px 28px 0;
  }
  .home-hero__inner {
    gap: 28px;
    padding: 0 0 40px;
  }
  .hero-mark {
    width: 132px;
    height: 132px;
  }
  h1 {
    font-size: 64px;
  }
  .hero-tagline {
    font-size: 21px;
  }
  .hero-description {
    font-size: 13px;
  }
  .hero-actions .site-button {
    font-size: 13px;
    padding-inline: 13px;
  }
}
@media (max-width: 760px) {
  .home-hero {
    padding: 40px 20px 0;
  }
  .home-hero__inner {
    grid-template-columns: 1fr;
    gap: 20px;
    min-height: 0;
  }
  .home-hero__copy {
    padding-bottom: 20px;
  }
  .hero-mark {
    width: 116px;
    height: 116px;
  }
  h1 {
    font-size: 60px;
  }
  .hero-tagline {
    font-size: 21px;
    margin-top: 20px;
  }
  .hero-description {
    max-width: 400px;
  }
  .home-hero__rail {
    gap: 20px;
    flex-wrap: wrap;
    font-size: 11px;
    padding-block: 24px 28px;
  }
  .hero-meta {
    margin-top: 20px;
  }
}
</style>
