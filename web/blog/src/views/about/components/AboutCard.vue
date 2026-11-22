<script setup>
import { computed, ref } from "vue";
import BlogIcon from "@/components/BlogIcon.vue";
import { safeURL } from "@/utils/format";
import forest from "@/assets/skies/forest.webp";
const props = defineProps({ profile: { type: Object, required: true } });
const broken = ref("");
const portrait = computed(() => safeURL(props.profile.avatar));
const socials = computed(() =>
  (props.profile.socials || []).filter((social) => safeURL(social.url)),
);
</script>
<template>
  <article class="author">
    <div class="author-intro">
      <div class="intro-copy">
        <h1>关于{{ profile.name || "云间小记" }}</h1>
        <p v-if="profile.title" class="byline">{{ profile.title }}</p>
        <p v-if="profile.bio" class="bio">
          {{ profile.bio }}
        </p>
        <div
          v-if="profile.location || profile.email || socials.length"
          class="contact"
        >
          <span v-if="profile.location">{{ profile.location }}</span>
          <a v-if="profile.email" :href="'mailto:' + profile.email"
            ><BlogIcon name="mail" :size="15" />{{ profile.email }}</a
          >
          <a
            v-for="(social, index) in socials"
            :key="index"
            :href="safeURL(social.url)"
            target="_blank"
            rel="noopener noreferrer"
            >{{ social.label || social.url }} ↗</a
          >
        </div>
      </div>
      <div class="author-photo">
        <img
          v-if="portrait && broken !== portrait"
          :src="portrait"
          :alt="(profile.name || '博主') + '的头像'"
          @error="broken = portrait"
        />
        <img v-else :src="forest" alt="" />
      </div>
    </div>
    <div v-if="profile.content" class="chapter">
      <h2>自述</h2>
      <p>{{ profile.content }}</p>
    </div>
    <div v-if="profile.skills?.length" class="chapter">
      <h2>喜欢与擅长</h2>
      <ul>
        <li v-for="skill in profile.skills" :key="skill.name || skill">
          {{ skill.name || skill }}
        </li>
      </ul>
    </div>
  </article>
</template>
<style scoped>
.author-intro {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(0, 1fr);
  gap: clamp(24px, 5vw, 74px);
  align-items: center;
  padding: 8px 4px 30px 0;
}
.intro-copy {
  min-width: 0;
}
h1 {
  margin: 0;
  max-width: 15em;
  font: 600 clamp(32px, 4.4vw, 56px)/1.22 var(--journal-serif);
  letter-spacing: -0.03em;
}
.byline {
  font-size: 14px;
  color: var(--journal-accent);
}
.bio {
  white-space: pre-line;
  margin: 16px 0 0;
  color: var(--journal-muted);
  font: 17px/1.9 var(--journal-serif);
}
.contact,
ul {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
}
.contact {
  margin-top: 18px;
  color: var(--journal-muted);
  font-size: 13px;
  overflow-wrap: anywhere;
}
.contact a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--journal-accent);
}
.author-photo {
  padding: 7px;
  background: white;
  border: 1px solid var(--journal-rule);
  box-shadow: 3px 4px 0 #e8e3db;
  transform: rotate(1.5deg);
}
.author-photo img {
  display: block;
  width: 100%;
  aspect-ratio: 3 / 2;
  object-fit: cover;
}
.chapter {
  display: grid;
  grid-template-columns: 115px minmax(0, 1fr);
  gap: 18px;
  border-top: 1px solid var(--journal-rule);
  padding: 18px 0;
}
h2 {
  margin: 0;
  font: 600 17px/1.8 var(--journal-serif);
}
.chapter p {
  margin: 0;
  white-space: pre-wrap;
  font-size: 15px;
  line-height: 1.9;
}
ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
li {
  color: var(--journal-accent);
  font-size: 13px;
}
@media (max-width: 600px) {
  .author-intro {
    grid-template-columns: 1fr;
    gap: 20px;
    padding-bottom: 20px;
  }
  .author-photo {
    max-width: 96%;
  }
  .author-photo img {
    aspect-ratio: 2.7;
  }
  .chapter {
    grid-template-columns: 1fr;
    gap: 5px;
  }
}
</style>
