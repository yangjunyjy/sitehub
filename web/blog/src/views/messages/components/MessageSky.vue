<script setup>
import { computed, ref } from 'vue';
import GalaxyField from './GalaxyField.vue';
const props = defineProps({ messages: { type: Array, required: true } });
const turns = ref([0, 0, 0]);
const flights = computed(() => {
  const lanes = Math.min(3, props.messages.length);
  return Array.from({ length: lanes }, (_, lane) => props.messages[(lane + turns.value[lane] * lanes) % props.messages.length]);
});
</script>
<template>
  <section class="message-universe" aria-labelledby="sky-title">
    <GalaxyField />
    <div class="meteor-stream" aria-hidden="true">
      <div v-for="(message, lane) in flights" :key="lane" class="message-meteor" @animationiteration.self="turns[lane]++">
        <i class="meteor-head"></i>
        <div class="meteor-note"><small>{{ message.nickname || '匿名访客' }}</small><span>{{ message.content }}</span></div>
      </div>
    </div>
    <h1 id="sky-title">留言板</h1>
    <a href="#guestbook">阅读留言 <span aria-hidden="true">↓</span></a>
  </section>
</template>
<style scoped>
.meteor-stream { position: absolute; inset: 0; overflow: hidden; pointer-events: none; }
.message-meteor { --slope: -17deg; --hue: #f5d9ae; --travel: 102vw; position: absolute; left: 99%; top: 11%; width: 1px; height: 1px; animation: message-flight 12s linear -4s infinite; }
.message-meteor:nth-child(2) { --slope: -13deg; --hue: #c6d8ff; top: 37%; animation-duration: 15s; animation-delay: -1.9s; }
.message-meteor:nth-child(3) { --slope: -10deg; --hue: #edc8df; top: 56%; animation-duration: 16.5s; animation-delay: -9s; }
.meteor-head { position: absolute; inset: -2px; border-radius: 50%; background: #fff9e7; box-shadow: 0 0 5px 2px var(--hue), 0 0 15px 3px var(--hue); }
.meteor-head::before { content: ''; position: absolute; left: 2px; top: 1px; width: clamp(48px, 10vw, 150px); height: 2px; background: linear-gradient(90deg, white, var(--hue) 20%, transparent); clip-path: polygon(0 0, 100% 50%, 0 100%); }
.meteor-note { position: absolute; top: 6px; left: 3px; width: max-content; max-width: clamp(105px, 25vw, 240px); transform-origin: top left; transform: rotate(calc(-1 * var(--slope))); text-shadow: 0 1px 4px #04040c, 0 0 12px #070812; }
.meteor-note small, .meteor-note span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.meteor-note small { font-size: 11px; color: var(--hue); }
.meteor-note span { font-size: clamp(12px, 1.2vw, 16px); color: #fff9f2; }
@keyframes message-flight {
  0% { opacity: 0; transform: rotate(var(--slope)) translateX(0); }
  9%, 84% { opacity: .98; }
  96%, 100% { opacity: 0; transform: rotate(var(--slope)) translateX(calc(-1 * var(--travel))); }
}
@media (prefers-reduced-motion: reduce) { .meteor-stream { display: none; } }
.message-universe { position: relative; min-height: 100svh; overflow: hidden; isolation: isolate; display: grid; place-items: center; color: #fff9ed; background: radial-gradient(ellipse at 58% 44%, #282344, #100f24 46%, #05060f 88%); }
h1 { z-index: 1; margin: 0; font: 500 clamp(34px, 5vw, 65px)/1.25 "Songti SC", SimSun, serif; letter-spacing: .22em; text-indent: .22em; text-shadow: 0 2px 16px #09081e, 0 0 44px #201833; }
a { position: absolute; bottom: max(30px, env(safe-area-inset-bottom)); left: 50%; transform: translateX(-50%); color: #fff; letter-spacing: .12em; font-size: 14px; }
</style>
