<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import SkyHero from "@/components/SkyHero.vue";
import BlogIcon from "@/components/BlogIcon.vue";
import { about } from "@/api";
import { useResource } from "@/composables/useResource";
const settings = useResource(about);
const verse = computed(() => settings.data.value?.homeVerse || "");
const letters = computed(() => Array.from(verse.value));
const reduced = ref(true);
const playing = computed(() => settings.data.value?.homeAnimation === true && !reduced.value);
const length = ref(0);
const ink = computed(() => playing.value ? letters.value.slice(0, length.value).join("") : verse.value);
let timer, preference;

function runVerse() {
  clearTimeout(timer);
  length.value = 0;
  if (!playing.value || !letters.value.length) return;
  let direction = 1;
  function advance() {
    length.value += direction;
    let delay = direction > 0 ? 145 : 70;
    if (length.value === letters.value.length) { direction = -1; delay = 1900; }
    if (length.value === 0) { direction = 1; delay = 260; }
    timer = setTimeout(advance, delay);
  }
  advance();
}
const followSystem = () => { reduced.value = preference.matches; };
const refreshSettings = () => { if (!document.hidden) settings.load(); };
watch([playing, verse], runVerse);
onMounted(() => {
  preference = matchMedia("(prefers-reduced-motion: reduce)");
  followSystem();
  preference.addEventListener("change", followSystem);
  settings.load();
  window.addEventListener('focus', refreshSettings);
  document.addEventListener('visibilitychange', refreshSettings);
});
onBeforeUnmount(() => {
  clearTimeout(timer);
  window.removeEventListener('focus', refreshSettings);
  document.removeEventListener('visibilitychange', refreshSettings);
  preference?.removeEventListener("change", followSystem);
});
</script>

<template>
  <div class="home-hero">
    <SkyHero scene="home" :home-motion="playing" full>
      <h1 v-if="verse" class="home-verse" :aria-label="verse">
        <span class="verse-measure" aria-hidden="true">{{ verse }}</span>
        <span class="verse-ink" aria-hidden="true">{{ ink }}</span>
      </h1>
      <nav class="hero-socials" aria-label="博客互动">
        <router-link to="/about" title="关于博主" aria-label="关于博主"><BlogIcon name="feather" /></router-link>
        <router-link to="/messages" title="留言板" aria-label="留言板"><BlogIcon name="mail" /></router-link>
        <router-link to="/friends" title="友情链接" aria-label="友情链接"><BlogIcon name="link" /></router-link>
      </nav>
    </SkyHero>
    <a class="scroll-down" href="#articles">向下阅读 <BlogIcon name="arrow" :size="16" /></a>
  </div>
</template>

<style lang="scss" scoped>
.home-hero {
  position: relative;
  color: white;
}
.home-verse {
  display: grid;
  width: max-content;
  max-width: 100%;
  margin: 0 auto 24px;
  font: 600 clamp(16px, 4.3vw, 26px) / 1.6 "Songti SC", "Noto Serif CJK SC", SimSun, serif;
  letter-spacing: .12em;
  text-align: start;
  overflow-wrap: anywhere;
  text-shadow: 0 1px 4px #121e46, 0 3px 16px #171d3aab;
  > span { grid-area: 1 / 1; }
}
.verse-measure { visibility: hidden; }
.verse-ink {
  justify-self: start;
  align-self: center;
}
.hero-socials {
  display: inline-flex;
  padding: 2px 8px;
  gap: 3px;
  background: rgb(31 37 66 / 42%);
  border: 1px solid rgb(255 251 245 / 28%);
  border-radius: 30px;
  a {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    transition: background .2s, transform .2s;
    &:hover, &:focus-visible { background: #fff2; transform: translateY(-2px); }
  }
}
.scroll-down {
  position: absolute;
  bottom: max(28px, env(safe-area-inset-bottom));
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 13px;
  letter-spacing: .12em;
  text-shadow: 0 1px 5px #15192e;
}
</style>
