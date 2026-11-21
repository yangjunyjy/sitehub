<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArticleFeed from "@/components/ArticleFeed.vue";
import TopicSelector from "@/components/TopicSelector.vue";
</script>
<template>
  <SkyHero
    title="标签"
    scene="harbor"
  /><TopicSelector kind="tags" /><ArticleFeed heading="标签文章" />
</template>
