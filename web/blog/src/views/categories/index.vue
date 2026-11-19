<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArticleFeed from "@/components/ArticleFeed.vue";
import CategorySelector from "./components/CategorySelector.vue";
</script>
<template>
  <SkyHero
    title="分类"
  /><CategorySelector /><ArticleFeed heading="分类文章" />
</template>
