<script setup>
import BlogHeader from './components/BlogHeader.vue'
import BlogFooter from './components/BlogFooter.vue'
</script>
<template>
  <div class="blog-app">
    <a class="skip-link" href="#main-content">跳至主要内容</a>
    <BlogHeader />
    <main id="main-content"><slot /></main>
    <BlogFooter v-if="$route.meta.page !== 'messages'" />
  </div>
</template>
