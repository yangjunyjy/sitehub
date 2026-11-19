<script setup>
import { onMounted } from "vue";
import SkyHero from "@/components/SkyHero.vue";
import ContentState from "@/components/ContentState.vue";
import AboutCard from "./components/AboutCard.vue";
import { useResource } from "@/composables/useResource";
import { about } from "@/api";
const { data, loading, error, load } = useResource(about);
onMounted(load);
</script>
<template>
  <SkyHero
    title="关于我"
    scene="forest"
  />
  <section class="content-shell">
    <ContentState
      v-if="loading || error"
      :loading="loading"
      :error="error"
      @retry="load"
    /><AboutCard v-else-if="data" :profile="data" /><ContentState
      v-else
      title="介绍正在酝酿中"
      description="先读一篇文章，或许也能了解这里。"
    />
  </section>
</template>
