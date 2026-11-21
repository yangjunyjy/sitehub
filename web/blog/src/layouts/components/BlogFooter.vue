<script setup>
import { inject } from "vue";
import MenuBranch from "@shared/navigation/MenuBranch.vue";
const navigation = inject("navigation");
import BlogIcon from "@/components/BlogIcon.vue";
</script>
<template>
  <footer class="blog-footer">
    <BlogIcon name="feather" :size="25" />
    <nav aria-label="页脚导航">
      <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" mobile />
    </nav>
    <small
      >© {{ new Date().getFullYear() }} 云间小记 · Built with SiteHub</small
    >
  </footer>
</template>
<style scoped lang="scss">
.blog-footer {
  background: #f8f8fa;
  padding: 46px 20px 32px;
  text-align: center;
  color: #aaa3b4;
  > p {
    font-size: 11px;
    letter-spacing: 2px;
    margin: 16px 0 23px;
  }
  > nav {
    flex-wrap: wrap;
    margin-top: 16px;
    display: flex;
    justify-content: center;
    gap: 20px;
    font-size: 10px;
  }
  small {
    display: block;
    font-size: 9px;
    margin-top: 24px;
    letter-spacing: 0.5px;
  }
}
</style>
