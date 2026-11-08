<script setup>
import { computed } from "vue";
import ProductCapabilities from "@/components/product/ProductCapabilities.vue";
import ProductShowcase from "@/components/product/ProductShowcase.vue";
import HomeHero from "@/components/home/HomeHero.vue";
import HomeDelivery from "@/components/home/HomeDelivery.vue";
import HomeInsights from "@/components/home/HomeInsights.vue";
import { useMember } from "@/composables/useMember";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
const { website } = useWebsiteContent();
const { member } = useMember();
const home = computed(() => website.content.home);
const productCapabilities = computed(() => website.content.capabilities);
const deliverySteps = computed(() => website.content.deliverySteps);
const audiences = computed(() => website.content.audiences);
</script>

<template>
  <div class="home-page">
    <HomeHero :home="home" :signed-in="!!member.user" />

    <section id="capabilities" class="site-section capabilities-section">
      <div class="site-shell">
        <header class="section-header">
          <div>
            <span class="site-index">01</span>
            <p class="site-kicker">One control plane</p>
          </div>
          <h2 class="site-heading">{{ home.capabilityTitle }}</h2>
          <p>{{ home.capabilityIntro }}</p>
        </header>
        <ProductCapabilities :items="productCapabilities" linked />
        <div class="section-action">
          <router-link class="site-button site-button--outline" to="/docs"
            >阅读使用文档
          </router-link>
        </div>
      </div>
    </section>

    <section class="home-showcase site-section">
      <div class="site-shell">
        <header class="showcase-heading">
          <p class="site-kicker">A workspace that works for you</p>
          <h2 class="site-heading">从命令行到工作台，始终掌控全局。</h2>
          <p>基础设施、应用交付与操作记录，在同一个工作区自然协作。</p>
        </header>
        <ProductShowcase />
      </div>
    </section>
    <HomeDelivery :home="home" :delivery-steps="deliverySteps" />

    <section id="teams" class="site-section audience-section">
      <div class="site-shell">
        <header class="section-header">
          <div>
            <span class="site-index">03</span>
            <p class="site-kicker">Built for teams</p>
          </div>
          <h2 class="site-heading">{{ home.audienceTitle }}</h2>
          <p>{{ home.audienceIntro }}</p>
        </header>
        <div class="audience-grid">
          <article v-for="item in audiences" :key="item.index">
            <span>{{ item.index }} / {{ item.title }}</span>
            <h3>{{ item.headline }}</h3>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </div>
    </section>

    <HomeInsights :home="home" />

    <section class="home-cta">
      <div class="home-cta__grid"></div>
      <div class="site-shell home-cta__inner">
        <div>
          <p class="site-kicker">Ready to deploy?</p>
          <h2>{{ home.cta }}</h2>
        </div>
        <div>
          <p>选择适合你的部署方式，开始统一管理基础设施和应用交付。</p>
          <router-link
            class="site-button site-button--primary"
            :to="member.user ? '/account' : '/register'"
            >{{ member.user ? "进入我的账号" : "创建你的 Deploy 账号" }}
          </router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.capabilities-section {
  border-top: 1px solid var(--deploy-line);
}
.section-action {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}
.home-showcase {
  background: var(--deploy-paper);
}
.home-showcase > .site-shell {
  max-width: 1080px;
}
.showcase-heading {
  text-align: center;
  margin-bottom: 40px;
}
.showcase-heading > p:last-child {
  color: var(--deploy-muted);
  font-size: 15px;
  margin-top: 16px;
}
.audience-section {
  background: var(--deploy-white);
}
.audience-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 32px;
}
.audience-grid article {
  padding: 24px 0;
  border-top: 1px solid var(--deploy-line);
}
.audience-grid span {
  color: var(--deploy-faint);
  font-size: 12px;
}
.audience-grid h3 {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 12px;
}
.audience-grid p {
  font-size: 14px;
  line-height: 1.9;
  color: var(--deploy-muted);
  margin: 0;
}
.home-cta {
  padding: 72px 0;
  border-top: 1px solid var(--deploy-line);
  background: var(--deploy-paper);
}
.home-cta__grid {
  display: none;
}
.home-cta__inner {
  text-align: center;
}
.home-cta h2 {
  font-size: 30px;
  font-weight: 600;
  line-height: 1.5;
  margin: 16px auto;
  max-width: 740px;
}
.home-cta__inner > div:last-child > p {
  color: var(--deploy-muted);
  font-size: 14px;
  margin: 0 0 24px;
}
@media (max-width: 760px) {
  .audience-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .home-cta {
    padding: 48px 0;
  }
  .home-cta h2 {
    font-size: 25px;
  }
}
</style>
