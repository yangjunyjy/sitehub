<script setup>
import { computed } from "vue";
import ProductCapabilities from "./components/ProductCapabilities.vue";
import ProductShowcase from "./components/ProductShowcase.vue";
import ContentState from '@/components/product/ContentState.vue';
import HomeHero from "./components/HomeHero.vue";
import HomeDelivery from "./components/HomeDelivery.vue";
import { useMember } from "@/composables/useMember";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
const { website, loadWebsiteContent } = useWebsiteContent();
const { member } = useMember();
const home = computed(() => website.content.home);
const productCapabilities = computed(() => website.content.capabilities);
const deliverySteps = computed(() => website.content.deliverySteps);
const audiences = computed(() => website.content.audiences);
</script>

<template>
  <ContentState v-if="website.error && !website.loaded" class="site-shell home-state" error title="暂时无法加载产品内容"
    message="连接未能完成，请稍后重试。" action-label="重新加载" :loading="website.loading" @retry="loadWebsiteContent(true).catch(() => {})" />
  <div v-else-if="!website.loaded" class="site-shell home-state" role="status">正在加载产品信息…</div>
  <div v-else class="home-page">
    <HomeHero :home="home" :brand="website.content.brand.name" />

    <section id="capabilities" class="site-section capabilities-section">
      <div class="site-shell">
        <header class="section-header">
          <h2 class="site-heading">{{ home.capabilityTitle }}</h2>
          <p>{{ home.capabilityIntro }}</p>
        </header>
        <ProductCapabilities :items="productCapabilities" linked />
        <div class="section-action">
          <router-link class="site-button site-button--outline" :to="{ name: 'docs', params: { id: '' } }"
            >阅读使用文档
          </router-link>
        </div>
      </div>
    </section>

    <section class="home-showcase site-section">
      <div class="site-shell">
        <header class="showcase-heading">
          <h2 class="site-heading">产品工作台</h2>
          <p>基础设施、应用交付与操作记录，在同一个工作区自然协作。</p>
        </header>
        <ProductShowcase />
      </div>
    </section>
    <HomeDelivery :home="home" :delivery-steps="deliverySteps" />

    <section id="teams" class="site-section audience-section">
      <div class="site-shell">
        <header class="section-header">
          <h2 class="site-heading">{{ home.audienceTitle }}</h2>
          <p>{{ home.audienceIntro }}</p>
        </header>
        <div class="audience-grid">
          <article v-for="item in audiences" :key="item.index">
            <span>{{ item.title }}</span>
            <h3>{{ item.headline }}</h3>
            <p>{{ item.description }}</p>
          </article>
        </div>
      </div>
    </section>

    <section class="home-cta">
      <div class="home-cta__grid"></div>
      <div class="site-shell home-cta__inner">
        <div>
          <h2>{{ home.cta }}</h2>
        </div>
        <div>
          <p>选择适合你的部署方式，开始统一管理基础设施和应用交付。</p>
          <router-link
            class="site-button site-button--primary"
            :to="{ name: member.user ? 'account' : 'register' }"
            >{{ member.user ? "进入我的账号" : "创建你的 Deploy 账号" }}
          </router-link>
        </div>
      </div>
    </section>
  </div>
</template>

<style lang="scss" scoped src="./home.scss"></style>
