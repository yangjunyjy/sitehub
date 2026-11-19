<script setup>
import SkyHero from "@/components/SkyHero.vue";
import BlogIcon from "@/components/BlogIcon.vue";
import LiquidGlass from "@/components/inspira/LiquidGlass.vue";
</script>

<template>
  <div class="home-hero">
    <SkyHero title="首页" scene="home" full>
      <nav class="hero-socials" aria-label="博客互动">
        <LiquidGlass :radius="32" />
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
.hero-socials {
  display: inline-flex;
  position: relative;
  isolation: isolate;
  padding: 4px 8px;
  gap: 4px;
  --glass-tint: rgb(22 38 80 / 35%);
  a {
    display: grid;
    place-items: center;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    transition: background .2s, transform .2s;
    &:hover, &:focus-visible { background: #303451af; transform: translateY(-2px); }
  }
}
.scroll-down {
  position: absolute;
  bottom: 25px;
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
