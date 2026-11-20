<script setup>
import Meteors from "./inspira/Meteors.vue";
defineProps({
  title: String,
  subtitle: String,
  eyebrow: String,
  scene: { type: String, default: "twilight" },
  full: Boolean,
  homeMotion: Boolean,
});
</script>
<template>
  <section
    class="sky-hero"
    :class="['scene-' + scene, { 'sky-hero--full': full }]"
  >
    <div class="sky-hero__image"></div>
    <div class="sky-hero__veil"></div>
    <Meteors v-if="scene === 'home'" v-show="homeMotion" :count="12" shower />
    <Meteors v-else :count="full ? 10 : 5" />
    <div class="sky-hero__content">
      <span v-if="eyebrow" class="eyebrow">{{ eyebrow }}</span>
      <h1 v-if="title">{{ title }}</h1>
      <p v-if="subtitle">{{ subtitle }}</p>
      <slot />
    </div>
    <div v-if="scene !== 'home'" class="sky-hero__line" aria-hidden="true"></div>
  </section>
</template>
<style lang="scss" scoped>
.sky-hero {
  position: relative;
  min-height: 385px;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  place-items: center;
  overflow: hidden;
  color: #fff;
  background: linear-gradient(140deg, #8a9bca, #cec1e7 55%, #f6cbd2);
  isolation: isolate;
  &--full {
    min-height: 100svh;
  }
  &__image,
  &__veil {
    position: absolute;
    inset: 0;
    z-index: -1;
  }
  &__image {
    background: url("../assets/skies/twilight.webp") center 55% / cover;
    transform: scale(1.025);
  }
  &__veil {
    background: linear-gradient(180deg, #15233b26, #25254037 45%, #161d322e);
  }
  &__content {
    position: relative;
    text-align: center;
    width:100%;
    min-width:0;
    padding: 100px 24px 52px;
    max-width: 900px;
    text-shadow: 0 2px 22px #55548844;
  }
  h1 {
    font-family: "Songti SC", SimSun, serif;
    font-size: clamp(42px, 5.4vw, 76px);
    font-weight: 600;
    text-shadow: 0 2px 22px #1721398c;
    letter-spacing: 0.15em;
    margin: 16px 0 20px;
    line-height: 1.4;
  }
  p {
    font-size: 15px;
    letter-spacing: 0.15em;
    opacity: 0.9;
    line-height: 2;
  }
  .eyebrow {
    font: 11px var(--font-sans);
    letter-spacing: 0.4em;
    opacity: 0.85;
  }
  &__line {
    position: absolute;
    bottom: 0;
    width: 1px;
    height: 35px;
    background: #fff8;
  }
}
.scene-home {
  background: #444f83;
  --meteor-distance: clamp(155px, 43vmin, 440px);
  --meteor-tail: clamp(70px, 11vw, 156px);
  .sky-hero__image {
    background-image: url("../assets/skies/home.webp");
    transform: none;
  }
  .sky-hero__veil {
    background: linear-gradient(180deg, #161c52a8, #24294752 46%, #4a344c14 73%, #1d27445c);
  }
  .sky-hero__content {
    padding: 64px 16px 20px;
  }
}
.scene-mountain {
  background: #b4a0bb;
  .sky-hero__image {
    background-image: url("../assets/skies/mountain.webp");
  }
}
.scene-pink .sky-hero__image {
  background-image: url("../assets/skies/home.webp");
  background-position: center 70%;
}
.scene-purple .sky-hero__image {
  background-image: url("../assets/skies/archive.webp");
}
.scene-harbor .sky-hero__image {
  background-image: url("../assets/skies/harbor.webp");
  background-position: center 65%;
}
.scene-forest .sky-hero__image {
  background-image: url("../assets/skies/forest.webp");
}
@media (max-width: 640px) {
  .sky-hero {
    min-height: 325px;
    &--full {
      min-height: 100svh;
    }
    h1 {
      font-size: 32px;
      letter-spacing: 0.08em;
    }
    p {
      font-size: 13px;
    }
  }
}
</style>
