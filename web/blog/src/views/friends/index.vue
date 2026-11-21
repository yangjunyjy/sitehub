<script setup>
import { computed, onMounted, ref } from "vue";
import SkyHero from "@/components/SkyHero.vue";
import BlogIcon from "@/components/BlogIcon.vue";
import ContentState from "@/components/ContentState.vue";
import FriendCards from "./components/FriendCards.vue";
import FriendApplication from "./components/FriendApplication.vue";
import { about, friends } from "@/api";
import { useResource } from "@/composables/useResource";
import { safeURL } from "@/utils/format";
const { data, loading, error, load } = useResource(friends);
const profile = useResource(about);
const layout = ref('cards');
const copied = ref('');
const applying = ref(location.hash === '#apply');
const manualCopy = ref('');
const items = computed(() => data.value?.items || []);
const site = computed(() => ({
  name: profile.data.value?.name || '云间小记',
  url: new URL(import.meta.env.BASE_URL, location.origin).href,
  description: profile.data.value?.bio || profile.data.value?.title || '',
  avatar: safeURL(profile.data.value?.avatar),
}));
function chooseLayout(value) {
  layout.value = value;
  try { localStorage.setItem('blog:friends:layout', value); } catch {}
}
async function copySite() {
  const text = [`名称：${site.value.name}`, `链接：${site.value.url}`, site.value.description && `描述：${site.value.description}`, site.value.avatar && `图标：${site.value.avatar}`].filter(Boolean).join('\n');
  try { await navigator.clipboard.writeText(text); copied.value = '已复制'; }
  catch { manualCopy.value = text; copied.value = '请选中下方文字复制'; }
}
onMounted(() => {
  try { if (localStorage.getItem('blog:friends:layout') === 'list') layout.value = 'list'; } catch {}
  load();
  profile.load();
});
</script>
<template>
  <SkyHero class="friends-hero" title="友情链接" scene="mountain" />
  <section id="friends" class="content-shell friends-content">
    <section class="site-sheet">
      <h2>关于本站的友链信息</h2>
      <div class="site-details">
        <div class="site-brand"><img v-if="site.avatar" :src="site.avatar" alt="" /><BlogIcon v-else name="feather" :size="36" /><div><strong>{{ site.name }}</strong><p v-if="site.description">{{ site.description }}</p></div></div>
        <dl><dt>链接</dt><dd><a :href="site.url">{{ site.url }}</a></dd><template v-if="site.description"><dt>描述</dt><dd>{{ site.description }}</dd></template><template v-if="site.avatar"><dt>图标</dt><dd>{{ site.avatar }}</dd></template></dl>
        <button type="button" @click="copySite">复制友链信息</button>
      </div>
      <small role="status">{{ copied }}</small><textarea v-if="manualCopy" readonly :value="manualCopy" aria-label="本站友链信息" @focus="$event.target.select()" />
    </section>
    <header class="friends-heading">
      <h2><BlogIcon name="user" /> 友链们</h2>
      <span v-if="!loading && !error">共 {{ items.length }} 个朋友</span>
      <div class="view-switch" role="group" aria-label="友链展示方式"><button :aria-pressed="layout === 'cards'" @click="chooseLayout('cards')">卡片</button><button :aria-pressed="layout === 'list'" @click="chooseLayout('list')">列表</button></div>
    </header>
    <ContentState v-if="loading || error" :loading="loading" :error="error" @retry="load" />
    <FriendCards v-else-if="items.length" :items="items" :layout="layout" />
    <p v-else class="friends-empty">还没有已发布的友链，欢迎成为第一位朋友。</p>
    <div class="apply-banner"><div><h2>申请友链</h2><p>如果你也在认真记录生活，欢迎交换友链。</p></div><button aria-controls="apply" :aria-expanded="applying" @click="applying = !applying">{{ applying ? '收起申请' : '前往申请 →' }}</button></div>
    <FriendApplication v-show="applying" />
  </section>
</template>
<style lang="scss" scoped src="./friends.scss"></style>
