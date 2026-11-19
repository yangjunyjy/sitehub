<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArticleFeed from "@/components/ArticleFeed.vue";
import TagSelector from "./components/TagSelector.vue";
</script>
<template>
  <SkyHero
    title="标签"
    scene="harbor"
  /><TagSelector /><ArticleFeed heading="标签文章" />
</template>
