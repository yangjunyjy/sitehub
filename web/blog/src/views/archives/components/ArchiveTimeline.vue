<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { articles } from "@/api";
import { archiveGroups, collectBlogIndex } from "@/utils/blog-index";
import ContentState from "@/components/ContentState.vue";
const items = ref([]),
  loading = ref(true),
  error = ref(""),
  progress = ref(0),
  expected = ref(0);
const groups = computed(() => archiveGroups(items.value));
let generation = 0,
  controller;
async function load() {
  const id = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = "";
  try {
    const result = await collectBlogIndex(articles, {
      signal: controller.signal,
      onProgress(count, total) {
        if (id === generation) {
          progress.value = count;
          expected.value = total;
        }
      },
    });
    if (id === generation) items.value = result;
  } catch (failure) {
    if (id === generation) error.value = failure.message;
  } finally {
    if (id === generation) loading.value = false;
  }
}
onMounted(load);
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
});
</script>
<template>
  <div class="archive-timeline glass-panel">
    <ContentState
      v-if="loading || error"
      :loading="loading"
      :error="error"
      @retry="load"
    />
    <p v-if="loading && expected" class="archive-progress">
      {{ progress }} / {{ expected }}
    </p>
    <ContentState
      v-else-if="!loading && !items.length && !error"
      title="时间轴正等待第一篇故事"
    /><template v-else-if="!loading && !error"
      ><div class="archive-heading">
        <h2>文章时间轴</h2>
        <p>共 {{ items.length }} 篇文章</p>
      </div>
      <section v-for="group in groups" :key="group.year" class="archive-year">
        <h3>
          {{ group.year === "undated" ? "未标注日期" : group.year
          }}<small>{{ group.count }} 篇</small>
        </h3>
        <div
          v-for="month in group.months"
          :key="month.month"
          class="archive-month"
        >
          <div class="archive-month__label">
            {{ month.month === "undated" ? "—" : month.month + " 月" }}
          </div>
          <router-link
            v-for="article in month.items"
            :key="article.id"
            :to="{ name: 'article', params: { id: article.id } }"
            class="timeline-entry"
            ><time>{{
              article.date ? article.date.slice(5).replace("-", " / ") : "—"
            }}</time
            ><span>{{ article.title }}</span
            ><i>↗</i></router-link
          >
        </div>
      </section></template
    >
  </div>
</template>
<style lang="scss" scoped>
.archive-timeline {
  max-width: 760px;
  margin: 32px auto;
  width: calc(100% - 40px);
  padding: clamp(24px, 5vw, 65px);
}
.archive-heading {
  text-align: center;
  margin-bottom: 55px;
  > span {
    font-size: 26px;
    color: #b1a3c6;
  }
  h2 {
    font-size: 23px;
    letter-spacing: 3px;
    font-weight: 400;
  }
  p {
    font-size: 12px;
    color: var(--muted);
  }
}
.archive-year {
  padding-left: 35px;
  border-left: 1px solid #e4ddeb;
  position: relative;
  padding-bottom: 34px;
  &:before {
    content: "";
    position: absolute;
    width: 11px;
    height: 11px;
    border: 3px solid white;
    border-radius: 50%;
    background: #ac9abb;
    top: 13px;
    left: -6px;
    box-shadow: 0 0 0 1px #d7cadf;
  }
  h3 {
    margin: 0 0 25px;
    font-size: 26px;
    font-weight: 400;
    color: #8c7b9f;
    small {
      font-size: 10px;
      color: var(--muted);
      margin-left: 15px;
    }
  }
}
.archive-month__label {
  font-size: 11px;
  color: #b1a6bd;
  margin: 20px 0 8px;
}
.timeline-entry {
  display: flex;
  align-items: baseline;
  gap: 25px;
  font-size: 13px;
  padding: 14px 0;
  border-bottom: 1px dashed var(--line);
  &:hover {
    color: var(--accent);
  }
  time {
    color: var(--muted);
    font-size: 10px;
    white-space: nowrap;
  }
  > span {
    flex: 1;
  }
  i {
    font-style: normal;
    color: #bcafc7;
  }
}
.archive-progress {
  text-align: center;
  font-size: 12px;
  color: var(--muted);
}
@media (max-width: 640px) {
  .archive-timeline {
    padding: 45px 20px;
  }
  .archive-year {
    padding-left: 24px;
  }
  .timeline-entry {
    gap: 15px;
    font-size: 12px;
  }
}
</style>
