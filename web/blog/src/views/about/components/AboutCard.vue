<script setup>
import { safeURL } from "@/utils/format";
defineProps({ profile: { type: Object, required: true } });
</script>
<template>
  <div class="about-grid">
    <figure class="about-photo glass-panel">
      <img v-if="safeURL(profile.avatar)" :src="safeURL(profile.avatar)" :alt="profile.name || '博主照片'" />
      <img v-else src="../../../assets/skies/forest.webp" alt="松林上空的云霞" />
      <figcaption>{{ profile.name || "云间小记" }}</figcaption>
    </figure>
    <div class="about-details">
    <article class="about-story glass-panel">
      <h2>{{ profile.title || "关于我，和这个小小的世界" }}</h2>
      <p class="about-bio">
        {{ profile.bio || "博主还没有留下自我介绍，欢迎先看看文章。" }}
      </p>
      <div v-if="profile.content" class="about-content">
        {{ profile.content }}
      </div>
    </article>
    <aside v-if="profile.location || profile.email || profile.skills?.length || profile.socials?.length" class="about-profile glass-panel">
      <p v-if="profile.location" class="profile-location">
        ⌖ {{ profile.location }}
      </p>
      <p v-if="profile.email" class="profile-email">
        <a :href="'mailto:' + profile.email">{{ profile.email }}</a>
      </p>
      <div v-if="profile.skills?.length" class="about-skills">
        <h4>喜欢与擅长</h4>
        <span
          v-for="skill in profile.skills"
          :key="typeof skill === 'string' ? skill : skill.name"
          >{{ typeof skill === "string" ? skill : skill.name }}</span
        >
      </div>
      <div v-if="profile.socials?.length" class="about-socials">
        <a
          v-for="social in profile.socials"
          :key="social.url"
          :href="safeURL(social.url) || undefined"
          target="_blank"
          rel="noopener noreferrer"
          >{{ social.label }} ↗</a
        >
      </div>
    </aside>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.about-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr);
  gap: 24px;
  align-items: start;
  padding: 70px 0 90px;
}
.about-details {
  display: grid;
  gap: 16px;
  overflow-wrap: anywhere;
}
.about-story, .about-profile { padding: clamp(20px, 3vw, 36px); }
.about-story {
  h2 {
    font-size: 27px;
    font-weight: 400;
    line-height: 1.8;
    margin: 19px 0 27px;
  }
  .about-bio,
  .about-content {
    font-size: 14px;
    line-height: 2.5;
    color: #8c8291;
    white-space: pre-wrap;
  }
  .about-content {
    margin-top: 30px;
  }
}
.about-profile {
  text-align: center;
  .profile-location,
  .profile-email {
    font-size: 11px;
    color: var(--muted);
    margin: 7px 0;
  }
}
.about-photo {
  margin: 0;
  padding: 12px;
  img {
    display: block;
    width: 100%;
    aspect-ratio: 4/3;
    object-fit: cover;
    border-radius: 9px;
  }
  figcaption { margin: 12px 4px 2px; color: var(--muted); }
}
.about-skills {
  text-align: left;
  margin-top: 32px;
  border-top: 1px solid var(--line);
  padding-top: 24px;
  h4 {
    font-size: 12px;
    font-weight: 400;
    margin: 0 0 17px;
    span {
      float: right;
      font-size: 8px;
      letter-spacing: 2px;
      color: var(--muted);
    }
  }
  > span {
    display: inline-block;
    font-size: 10px;
    padding: 5px 9px;
    background: #f7f0f5;
    border-radius: 3px;
    margin: 0 7px 9px 0;
    color: #ab8eaa;
  }
}
.about-socials {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
  font-size: 11px;
  margin-top: 20px;
  color: var(--accent);
}
@media (max-width: 850px) {
  .about-grid {
    gap: 35px;
    grid-template-columns: 1fr 230px;
  }
}
@media (max-width: 640px) {
  .about-grid {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    padding: 40px 0 55px;
  }
  .about-story h2 {
    font-size: 23px;
  }
}
</style>
