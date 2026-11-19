<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArticleFeed from "@/components/ArticleFeed.vue";
import SearchForm from "./components/SearchForm.vue";
</script>
<template>
  <SkyHero
    title="搜索文章"
    ><SearchForm /></SkyHero
  ><ArticleFeed search heading="搜索结果" />
</template>
