<script setup>
import { ref } from 'vue';
import MessageSky from "./components/MessageSky.vue";
import DiscussionFeed from '@/components/DiscussionFeed.vue';
const latest = ref([]);
</script>
<template><MessageSky :messages="latest" /><div id="guestbook" class="guestbook"><DiscussionFeed @latest="latest = $event" /></div></template>
<style scoped>.guestbook { padding: clamp(32px, 7vw, 90px) 16px; scroll-margin-top: 64px; }</style>
