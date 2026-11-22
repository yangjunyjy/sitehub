<script setup>
import { computed, ref } from "vue";
const props = defineProps({
  headings: { type: Array, required: true },
  active: String,
});
const baseline = computed(() =>
  props.headings.reduce((min, heading) => Math.min(min, heading.level), 6),
);
const mobile = ref(null);
function follow(event, heading) {
  if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
  event.preventDefault();
  mobile.value.open = false;
  const target = document.getElementById(heading.slug);
  if (!target) return;
  history.replaceState(
    history.state,
    "",
    "#" + encodeURIComponent(heading.slug),
  );
  target.scrollIntoView({ block: "start" });
  target.tabIndex = -1;
  target.focus({ preventScroll: true });
}
</script>
<template>
  <aside class="article-outline">
    <details ref="mobile">
      <summary>文章目录 <span aria-hidden="true">⌄</span></summary>
    </details>
    <nav aria-label="文章目录">
      <strong>文章目录</strong>
      <a
        v-for="heading in headings"
        :key="heading.slug"
        :href="'#' + heading.slug"
        :aria-current="active === heading.slug ? 'location' : undefined"
        :style="{ '--depth': Math.min(heading.level - baseline, 3) }"
        @click="follow($event, heading)"
        >{{ heading.text }}</a
      >
    </nav>
  </aside>
</template>
<style scoped>
.article-outline {
  position: sticky;
  top: 90px;
  align-self: start;
  grid-area: outline;
  max-height: calc(100vh - 130px);
  overflow: auto;
  padding: 4px 4px 12px;
  scrollbar-width: thin;
}
details {
  display: none;
}
nav {
  display: grid;
}
nav strong,
summary {
  color: var(--journal-ink);
  font-size: 13px;
  font-weight: 650;
}
nav strong {
  margin: 0 0 12px 11px;
}
nav a {
  border-left: 1px solid var(--journal-rule);
  padding: 5px 9px 5px calc(10px + 10px * var(--depth));
  font-size: 13px;
  line-height: 1.55;
  color: var(--journal-muted);
  overflow-wrap: anywhere;
}
a[aria-current] {
  border-left: 2px solid var(--journal-accent);
  color: var(--journal-accent);
  background: var(--journal-wash);
}
a:hover {
  color: var(--journal-accent);
}
@media (max-width: 1179px) {
  .article-outline {
    position: static;
    max-height: none;
    padding: 0;
    margin: 0 0 26px;
    border-block: 1px solid var(--journal-rule);
  }
  details {
    display: block;
  }
  summary {
    padding: 12px 0;
    cursor: pointer;
    list-style: none;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  summary::-webkit-details-marker {
    display: none;
  }
  summary span {
    font-size: 19px;
    line-height: 1;
  }
  details[open] summary span {
    transform: rotate(180deg);
  }
  nav {
    display: none;
    padding-bottom: 12px;
    max-height: 45vh;
    overflow: auto;
  }
  details[open] + nav {
    display: grid;
  }
  nav strong {
    display: none;
  }
}
</style>
