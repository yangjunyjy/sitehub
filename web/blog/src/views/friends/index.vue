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
<style lang="scss" scoped>
.friends-hero { min-height: clamp(180px, 26vh, 260px); }
.friends-content { position: relative; margin-top: -22px; padding-bottom: 60px; color: #373950; }
.site-sheet, .apply-banner, .friends-empty { border: 1px solid #fffc; border-radius: 16px; box-shadow: 0 8px 26px #7980aa1c; }
.site-sheet { background: linear-gradient(115deg, #fdfcff, #e9edffed); padding: clamp(16px, 2.5vw, 28px); }
.friends-content h2 { margin: 0; font-size: 19px; }
.site-sheet h2 { border-inline-start: 4px solid #8883ba; padding-left: 10px; }
.site-details { display: flex; align-items: center; gap: 20px; margin-top: 18px; }
.site-brand { flex: 1; display: flex; align-items: center; gap: 14px; min-width: 0; }
.site-brand > :first-child { flex-shrink: 0; width: 54px; height: 54px; border-radius: 50%; object-fit: cover; background: #dbddf5; }
.site-brand > svg { padding: 13px; color: #656396; }
.site-details p, .apply-banner p { margin: 5px 0 0; font-size: 13px; }
.site-details dl { flex: 1.6; min-width: 0; display: grid; grid-template-columns: auto 1fr; gap: 3px 10px; font-size: 13px; margin: 0; padding-left: 22px; border-left: 1px solid #d2d6eb; }
.site-details dd { margin: 0; overflow-wrap: anywhere; }
.site-details a { color: #565797; }
.site-details button, .apply-banner button, .view-switch button { border: 1px solid #d2cfe3; border-radius: 9px; padding: 7px 12px; color: #555178; background: #fffc; font-size: 13px; }
.site-sheet small { display: block; color: var(--muted); }
.site-sheet textarea { width: 100%; margin-top: 10px; }
.friends-heading, .friends-heading h2 { display: flex; align-items: center; gap: 9px; }
.friends-heading { flex-wrap: wrap; margin: 25px 0 14px; }
.friends-heading span { margin-left: auto; font-size: 12px; color: var(--muted); }
.view-switch { display: flex; gap: 2px; padding: 3px; background: #e7e6f0; border-radius: 10px; }
.view-switch button { background: transparent; border: 0; padding: 4px 10px; }
.view-switch [aria-pressed='true'] { color: #494066; background: white; box-shadow: 0 1px 4px #35305721; }
.friends-empty { padding: 24px; text-align: center; background: #ffffffa8; color: var(--muted); }
.apply-banner { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 22px clamp(16px, 3vw, 34px); margin-top: 25px; background: linear-gradient(100deg, #e2f4f2 30%, #e2f3f4cf 70%, #deeff3a6), url('../../assets/skies/mountain.webp') center 65% / cover; }
.apply-banner button { flex-shrink: 0; color: white; background: #6865aa; border-color: transparent; }
@media (max-width: 720px) { .site-details { flex-wrap: wrap; } .site-brand, .site-details dl { flex-basis: 100%; } .site-details dl { padding-left: 0; border: 0; } }
</style>
