<script setup>
import { computed, onMounted, ref } from "vue";
import { messages } from "@/api";
import DiscussionComposer from "./discussion/DiscussionComposer.vue";
import { commentParts } from "./discussion/emoji";
import { useResource } from "@/composables/useResource";
const props = defineProps({ articleId: String });
const emit = defineEmits(["latest"]);
const label = computed(() => (props.articleId ? "评论" : "留言"));
const page = ref(1),
  heading = ref(null);
const size = 10;
const feed = useResource(async () => {
  const requestedPage = page.value;
  const response = await messages({
    page: requestedPage,
    pageSize: size,
    articleId: props.articleId,
  });
  if (!props.articleId && requestedPage === 1)
    emit("latest", response.data.items || []);
  return response;
});
const items = computed(() => feed.data.value?.items || []);
const pages = computed(() =>
  Math.max(1, Math.ceil((feed.data.value?.total || 0) / size)),
);
async function turn(next) {
  page.value = next;
  await feed.load();
  heading.value?.focus({ preventScroll: true });
  heading.value?.scrollIntoView({ block: "start" });
}
onMounted(() => feed.load());
</script>
<template>
  <section class="discussion" :aria-label="label + '列表'">
    <header>
      <h2 ref="heading" tabindex="-1">
        {{ articleId ? "文章评论" : "留言板" }}
      </h2>
      <span v-if="feed.data.value"
        >{{ feed.data.value.total }} 条{{ label }}</span
      >
    </header>
    <DiscussionComposer :article-id="articleId" />
    <p v-if="feed.loading.value" role="status">正在读取{{ label }}…</p>
    <p v-else-if="feed.error.value" role="alert">
      {{ feed.error.value }} <button @click="feed.load()">重试</button>
    </p>
    <ol v-else-if="items.length">
      <li v-for="item in items" :key="item.id">
        <span class="avatar" aria-hidden="true">{{
          (item.nickname || "读者").slice(0, 1)
        }}</span>
        <div>
          <div class="byline">
            <strong>{{ item.nickname || "读者" }}</strong
            ><time v-if="item.createdAt" :datetime="item.createdAt">{{
              new Date(item.createdAt).toLocaleDateString("zh-CN")
            }}</time>
          </div>
          <p>
            <template
              v-for="(part, index) in commentParts(item.content)"
              :key="index"
              ><img
                v-if="part.src"
                :src="part.src"
                :alt="part.symbol"
                :title="part.name"
                width="24"
                height="24"
              /><template v-else>{{ part.text }}</template></template
            >
          </p>
        </div>
      </li>
    </ol>
    <p v-else class="empty">还没有{{ label }}。</p>
    <nav v-if="pages > 1" :aria-label="label + '分页'">
      <button
        :disabled="page <= 1 || feed.loading.value"
        @click="turn(page - 1)"
      >
        上一页
      </button>
      <span>{{ page }} / {{ pages }}</span
      ><button
        :disabled="page >= pages || feed.loading.value"
        @click="turn(page + 1)"
      >
        下一页
      </button>
    </nav>
  </section>
</template>
<style scoped>
.discussion {
  width: min(100%, 840px);
  margin: 0 auto;
  padding: clamp(20px, 4vw, 40px);
  background: #fff;
  color: #26332a;
  border: 1px solid #dce1da;
  border-radius: 12px;
}
header,
.byline,
nav {
  display: flex;
  align-items: center;
  gap: 14px;
}
h2 {
  font-size: 23px;
  letter-spacing: -0.04em;
  margin: 0;
  scroll-margin-top: 84px;
}
header span,
time,
nav span,
.empty {
  color: #657169;
  font-size: 13px;
}
.empty {
  margin: 22px 0 0;
  padding-top: 15px;
  border-top: 1px solid #e5e8e2;
}
ol {
  padding: 0;
  list-style: none;
}
li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  border-top: 1px solid #e5e8e2;
  padding: 18px 0;
}
.avatar {
  flex: none;
  display: grid;
  place-items: center;
  background: #e7eee5;
  border-radius: 50%;
  width: 34px;
  height: 34px;
  font-size: 13px;
  color: #365749;
}
.byline {
  justify-content: flex-start;
  flex-wrap: wrap;
  gap: 9px;
}
li strong {
  font-size: 14px;
}
li p {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  margin: 6px 0 0;
  line-height: 1.75;
}
li img {
  vertical-align: middle;
}
nav {
  justify-content: center;
}
nav button,
[role="alert"] button {
  padding: 5px 12px;
  border: 1px solid #cdd6cb;
  background: #fff;
  border-radius: 5px;
}
</style>
