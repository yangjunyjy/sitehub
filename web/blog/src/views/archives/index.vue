<script setup>
import SkyHero from "@/components/SkyHero.vue";
import ArchiveTimeline from "./components/ArchiveTimeline.vue";
</script>
<template>
  <SkyHero
    title="归档"
    scene="purple"
    full
    ><a class="button button--glass" href="#timeline"
      >查看时间轴 ↓</a
    ></SkyHero
  >
  <div id="timeline"><ArchiveTimeline /></div>
</template>
