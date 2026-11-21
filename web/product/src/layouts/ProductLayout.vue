<script setup>
import { onMounted } from 'vue';
import { useRoute } from 'vue-router';
import Header from '@/components/Header.vue';
import Footer from '@/components/Footer.vue';
import SiteNavigationStatus from '@/components/product/SiteNavigationStatus.vue';
import { useWebsiteContent } from '@/composables/useWebsiteContent';

const route = useRoute();
const { loadWebsiteContent } = useWebsiteContent();
onMounted(() => loadWebsiteContent().catch(() => {}));
</script>

<template>
  <div class="product-app" :class="{ 'site-docs': route.meta.page === 'docs' }">
    <SiteNavigationStatus />
    <a class="skip-link" href="#main-content">跳至主要内容</a>
    <Header />
    <main id="main-content" class="app-main" tabindex="-1"><router-view /></main>
    <Footer v-if="route.meta.page !== 'docs' || route.meta.notFound" />
  </div>
</template>

<style lang="scss">
.product-app { display: flex; flex-direction: column; min-height: 100vh;
  > .app-main { flex: 1; }
}
</style>
