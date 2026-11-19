<script setup>
import { onMounted } from "vue";
import SkyHero from "@/components/SkyHero.vue";
import ContentState from "@/components/ContentState.vue";
import FriendCards from "./components/FriendCards.vue";
import FriendApplication from "./components/FriendApplication.vue";
import { friends } from "@/api";
import { useResource } from "@/composables/useResource";
const { data, loading, error, load } = useResource(friends);
onMounted(load);
</script>
<template>
  <SkyHero
    title="友链"
    scene="mountain"
    full
    ><a class="button button--glass" href="#friends">认识新朋友 ↓</a></SkyHero
  >
  <section id="friends" class="content-shell friends-content">
    <div class="friends-heading">
      <h2>友情链接</h2>
      <a class="text-button" href="#apply">申请交换友链 ↓</a>
    </div>
    <ContentState
      v-if="loading || error"
      :loading="loading"
      :error="error"
      @retry="load"
    /><FriendCards
      v-else-if="data?.items?.length"
      :items="data.items"
    /><ContentState
      v-else
      title="期待第一位朋友的到来"
      description="每一个独立的小站，都有自己独特的光。"
    />
    <div v-if="data?.site" class="my-site">
      <h3>本站信息</h3>
      <dl>
        <dt>名称</dt>
        <dd>{{ data.site.name }}</dd>
        <dt>链接</dt>
        <dd>{{ data.site.url }}</dd>
        <dt>简介</dt>
        <dd>{{ data.site.description }}</dd>
      </dl>
    </div>
    <FriendApplication />
  </section>
</template>
<style lang="scss" scoped>
.friends-content {
  padding-top: 65px;
  padding-bottom: 80px;
}
.friends-heading {
  text-align: center;
  margin-bottom: 40px;
  > span {
    font-size: 9px;
    letter-spacing: 3px;
    color: #b29bbb;
  }
  h2 {
    font-size: 26px;
    font-weight: 400;
    letter-spacing: 3px;
    margin: 13px 0;
  }
  p {
    font-size: 12px;
    color: var(--muted);
  }
}
.my-site {
  margin-top: 40px;
  background: #f9f7fb;
  padding: 25px 30px;
  border-radius: 5px;
  h3 {
    font-size: 14px;
    font-weight: 500;
    margin: 0 0 15px;
  }
  dl {
    display: grid;
    grid-template-columns: 45px 1fr;
    gap: 9px;
    font-size: 12px;
    margin: 0;
  }
  dt {
    color: var(--muted);
  }
  dd {
    margin: 0;
    overflow-wrap: anywhere;
  }
}
</style>
