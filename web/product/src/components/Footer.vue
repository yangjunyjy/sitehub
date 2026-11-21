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
    <div class="site-shell product-footer__top">
      <div class="footer-brand">
        <router-link class="footer-logo" :to="{ name: 'home' }"
          ><BrandMark /><strong>{{
            website.content.brand.name
          }}</strong></router-link
        >
        <p>
          {{ website.content.brand.fullName }}<br />{{
            website.content.footer.slogan
          }}
        </p>
      </div>
      <div class="footer-links">
        <nav aria-label="页脚导航">
          <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" mobile />
        </nav>
        <a v-if="website.content.brand.contactEmail" :href="'mailto:' + website.content.brand.contactEmail">联系我们</a>
      </div>
    </div>
    <div class="site-shell product-footer__bottom">
      <span
        >© {{ year }}
        {{
          website.content.brand.copyright || "Deploy. All rights reserved."
        }}</span
      ><a
        v-if="website.content.brand.recordNumber"
        href="https://beian.miit.gov.cn/"
        target="_blank"
        rel="noopener"
        >{{ website.content.brand.recordNumber }}</a
      ><span v-else>{{ website.content.footer.statement }}</span>
    </div>
  </footer>
</template>

<style lang="scss" scoped>
.product-footer {
  border-top: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  padding: 48px 0 24px;
}
.product-footer__top {
  display: flex;
  justify-content: space-between;
  gap: 48px;
  padding-bottom: 32px;
}
.footer-logo {
  display: flex;
  align-items: center;
  gap: 9px;
}
.footer-logo :deep(svg) {
  width: 28px;
  height: 28px;
}
.footer-logo strong {
  font-size: 18px;
  font-weight: 600;
}
.footer-brand p {
  font-size: 12px;
  line-height: 1.9;
  max-width: 320px;
  margin: 16px 0;
  color: var(--deploy-faint);
}
.footer-links {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 60px;
}
.footer-links > div {
  display: grid;
  gap: 12px;
  align-content: start;
  font-size: 12px;
}
.footer-links strong {
  font-weight: 600;
  margin-bottom: 4px;
}
.footer-links a {
  color: var(--deploy-muted);
}
.footer-links a:hover {
  color: var(--deploy-primary);
}
.product-footer__bottom {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  border-top: 1px solid var(--deploy-line);
  padding-top: 20px;
  color: var(--deploy-faint);
  font-size: 11px;
}
@media (max-width: 760px) {
  .product-footer__top {
    flex-direction: column;
    gap: 24px;
  }
  .footer-links {
    gap: 24px;
  }
  .product-footer__bottom {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
