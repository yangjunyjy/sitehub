<script setup>
import { ref } from "vue";
import BlogIcon from "@/components/BlogIcon.vue";
import { safeURL } from "@/utils/format";
defineProps({ items: { type: Array, default: () => [] }, layout: String });
const broken = ref(new Set());
</script>
<template>
  <div class="friend-grid" :class="{ 'friend-list': layout === 'list' }">
    <a
      v-for="friend in items"
      :key="friend.id || friend.url"
      :href="safeURL(friend.url) || undefined"
      target="_blank"
      rel="noopener noreferrer"
      class="friend-card"
      ><div class="friend-avatar">
        <img
          v-if="safeURL(friend.avatar) && !broken.has(friend.id || friend.url)"
          :src="safeURL(friend.avatar)"
          alt=""
          loading="lazy"
          @error="broken = new Set([...broken, friend.id || friend.url])"
        /><BlogIcon v-else name="feather" :size="25" />
      </div>
      <div>
        <h3>{{ friend.name }}</h3>
        <p>{{ friend.description }}</p>
      </div>
      <span>↗</span></a
    >
  </div>
</template>
<style lang="scss" scoped>
.friend-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}
.friend-card {
  display: flex;
  align-items: center;
  gap: 15px;
  position: relative;
  isolation: isolate;
  overflow: hidden;
  min-height: 98px;
  padding: 18px 14px;
  border: 1px solid #fff;
  border-radius: 14px;
  --wash: #e8e3f7;
  background: linear-gradient(150deg, white, #ffffffb3);
  box-shadow: 0 4px 12px #74769c1f;
  &:nth-child(4n + 2) { --wash: #dbeef7; }
  &:nth-child(4n + 3) { --wash: #f5e4dc; }
  &:nth-child(4n) { --wash: #dceddf; }
  &::before { content: ''; position: absolute; z-index: -1; inset: 65% -5% -18%; background: var(--wash); border-radius: 50% 70% 0 0; transform: rotate(-5deg); }
  transition:
    transform 0.25s,
    box-shadow 0.25s;
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 10px 25px #9583ac12;
  }
  > div:nth-child(2) {
    flex: 1;
    min-width: 0;
  }
  h3 {
    margin: 0 0 7px;
    font-size: 14px;
    font-weight: 500;
  }
  p {
    margin: 0;
    font-size: 11px;
    color: var(--muted);
    line-height: 1.8;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  > span {
    align-self: flex-start;
    color: #c0b3ca;
    font-size: 14px;
  }
}
.friend-grid.friend-list { grid-template-columns: 1fr; gap: 7px; }
.friend-list .friend-card { min-height: 65px; padding: 9px 18px; }
.friend-list .friend-card::before { inset: 0 auto 0 0; width: 5px; border-radius: 0; transform: none; }
.friend-list .friend-avatar { height: 38px; width: 38px; }
.friend-list h3 { margin-bottom: 2px; }
.friend-avatar {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: white;
  background: linear-gradient(145deg, #c6b6cf, #e5cccd);
  overflow: hidden;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}
@media (max-width: 1000px) {
  .friend-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 640px) {
  .friend-grid {
    grid-template-columns: 1fr;
  }
}
</style>
