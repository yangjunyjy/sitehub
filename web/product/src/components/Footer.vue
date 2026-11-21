<script setup>
import { inject } from 'vue';
import MenuBranch from '@shared/navigation/MenuBranch.vue';
const navigation = inject('navigation');
import BrandMark from "@/components/product/BrandMark.vue";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
const year = new Date().getFullYear();
const { website } = useWebsiteContent();
</script>

<template>
  <footer class="product-footer">
    <div class="site-shell">
      <div class="footer-main">
        <router-link class="footer-brand" :to="{ name: 'home' }"><BrandMark />{{ website.content.brand.name }}</router-link>
        <nav aria-label="页脚导航">
          <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" mobile />
          <a v-if="website.content.brand.contactEmail" :href="'mailto:' + website.content.brand.contactEmail">联系我们</a>
        </nav>
      </div>
      <p v-if="website.content.footer.slogan" class="footer-copy">{{ website.content.footer.slogan }}</p>
      <div class="footer-legal">
        <span>© {{ year }} {{ website.content.brand.copyright || website.content.brand.name }}</span>
        <span v-if="website.content.brand.fullName">{{ website.content.brand.fullName }}</span>
        <a v-if="website.content.brand.recordNumber" href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">{{ website.content.brand.recordNumber }}</a>
        <span v-if="website.content.footer.statement">{{ website.content.footer.statement }}</span>
      </div>
    </div>
  </footer>
</template>
<style scoped>
.product-footer {
  border-top: 1px solid var(--deploy-line);
  padding-block: 26px;
  background: var(--deploy-white);
  color: var(--deploy-ink);
}
.footer-main, .footer-brand, .footer-main nav, .footer-legal { display: flex; align-items: center; }
.footer-main { justify-content: space-between; gap: 18px; }
.footer-brand { gap: 8px; font-weight: 650; font-size: 16px; }
.footer-brand :deep(svg) { width: 22px; height: 22px; }
.footer-main nav { flex-wrap: wrap; align-items: flex-start; gap: 8px 18px; font-size: 13px; }
.footer-copy, .footer-legal { color: var(--deploy-muted); font-size: 12px; }
.footer-copy { margin: 14px 0 0; }
.footer-legal { flex-wrap: wrap; gap: 4px 18px; margin-top: 13px; }
@media (max-width: 600px) {
  .footer-main { align-items: flex-start; flex-wrap: wrap; }
  .footer-legal { align-items: flex-start; flex-direction: column; }
}
</style>
