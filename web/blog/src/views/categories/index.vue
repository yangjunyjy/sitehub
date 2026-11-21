<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArticleFeed from "@/components/ArticleFeed.vue";
import TopicSelector from "@/components/TopicSelector.vue";
</script>
<template>
  <SkyHero
    title="分类"
  /><TopicSelector kind="categories" /><ArticleFeed heading="分类文章" />
</template>
