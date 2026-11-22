<script setup>
import { computed, onMounted, ref } from "vue";
import BlogPagination from "@/components/BlogPagination.vue";
import BlogIcon from "@/components/BlogIcon.vue";
import NewFolderForm from "@/components/member/NewFolderForm.vue";
import { getFolders, getFavorites } from "@/api/library";
import { useMember } from "@/composables/useMember";

const folders = ref([]),
  active = ref("0"),
  articles = ref([]),
  total = ref(0),
  page = ref(1);
const creating = ref(false),
  busy = ref(true),
  error = ref("");
const { expired } = useMember();
const pageSize = 20,
  pages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));
const selected = computed(() =>
  folders.value.find((f) => String(f.id) === active.value),
);
let revision = 0;
function failed(err) {
  if (err.status === 401) {
    expired();
  } else error.value = err.message;
}
async function load(folderId = active.value, current = 1) {
  const id = ++revision;
  active.value = String(folderId);
  page.value = current;
  busy.value = true;
  error.value = "";
  articles.value = [];
  try {
    const data = await getFavorites(active.value, current, pageSize);
    if (id === revision) {
      articles.value = data.items;
      total.value = Number(data.total);
    }
  } catch (err) {
    if (id === revision) failed(err);
  } finally {
    if (id === revision) busy.value = false;
  }
}
async function init() {
  busy.value = true;
  error.value = "";
  try {
    folders.value = await getFolders();
    await load();
  } catch (err) {
    failed(err);
    busy.value = false;
  }
}
function created(folder) {
  folders.value.push(folder);
  creating.value = false;
  load(folder.id);
}
onMounted(init);
</script>
<template>
  <section class="library">
    <aside>
      <div class="aside-title">
        <h2>我的收藏夹</h2>
        <button
          title="新建收藏夹"
          aria-label="新建收藏夹"
          :aria-expanded="creating"
          @click="creating = !creating"
        >
          <span aria-hidden="true">＋</span> 新建
        </button>
      </div>
      <NewFolderForm v-if="creating" @created="created" />
      <nav aria-label="收藏夹">
        <button
          v-for="folder in folders"
          :key="folder.id"
          :aria-pressed="active === String(folder.id)"
          @click="load(folder.id)"
        >
          <BlogIcon name="folder" :size="17" /><span>{{ folder.name }}</span
          ><small>{{ folder.count }}</small>
        </button>
      </nav>
    </aside>
    <div class="saved">
      <header>
        <h2>{{ selected?.name || "收藏的文章" }}</h2>
        <small v-if="selected">{{ selected.count }} 篇文章</small>
      </header>
      <p v-if="busy" class="library-state" role="status">正在读取收藏…</p>
      <p v-else-if="error" class="library-state" role="alert">
        {{ error }} <button @click="init">重试</button>
      </p>
      <ul v-else-if="articles.length">
        <li v-for="article in articles" :key="article.id">
          <RouterLink :to="{ name: 'article', params: { id: article.id } }">
            <span class="book-mark" aria-hidden="true"
              ><BlogIcon name="bookmark" :size="17"
            /></span>
            <span class="saved-entry"
              ><strong>{{ article.title }}</strong>
              <small v-if="article.savedAt"
                >收藏于
                <time :datetime="article.savedAt">{{
                  new Date(article.savedAt).toLocaleDateString("zh-CN")
                }}</time></small
              >
            </span>
            <BlogIcon class="read-arrow" name="chevron" :size="15" />
          </RouterLink>
        </li>
      </ul>
      <div v-else class="empty">
        <BlogIcon name="bookmark" :size="24" />
        <p>这个收藏夹还没有文章</p>
        <RouterLink :to="{ name: 'home', hash: '#articles' }"
          >浏览文章 →</RouterLink
        >
      </div>
      <BlogPagination
        v-if="!error && pages > 1"
        label="收藏分页"
        :page="page"
        :pages="pages"
        :busy="busy"
        @change="load(active, $event)"
      />
    </div>
  </section>
</template>
<style lang="scss" scoped>
.library {
  display: grid;
  grid-template-columns: minmax(180px, 225px) minmax(0, 1fr);
  align-items: start;
  gap: clamp(24px, 4vw, 58px);
  padding-top: 28px;
}
.aside-title,
.saved header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 12px;
}
h2,
ul,
p {
  margin: 0;
}
.aside-title h2 {
  font-size: 14px;
}
.aside-title button,
.library-state button {
  border: none;
  background: transparent;
  padding: 4px;
  color: var(--journal-accent);
  font-size: 12px;
}
aside form,
aside nav {
  margin-top: 12px;
}
aside nav {
  display: grid;
  gap: 4px;
}
aside nav button {
  display: flex;
  align-items: center;
  gap: 9px;
  text-align: left;
  padding: 9px 12px;
  border: 0;
  border-left: 2px solid transparent;
  background: transparent;
  color: var(--journal-muted);
}
aside nav button[aria-pressed="true"] {
  background: var(--journal-wash);
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}
aside nav button span {
  flex: 1;
  font-size: 14px;
  overflow-wrap: anywhere;
}
.saved {
  position: relative;
  background: #fffefd;
  padding: clamp(20px, 3vw, 36px);
  border: 1px solid var(--journal-rule);
  box-shadow: 3px 4px 0 #ede9e2;
}
.saved header {
  padding-bottom: 14px;
  border-bottom: 1px solid var(--journal-rule);
}
.saved header h2 {
  font: 600 clamp(21px, 2.2vw, 28px)/1.35 var(--journal-serif);
}
small,
.library-state {
  color: var(--journal-muted);
  font-size: 12px;
}
.saved header small {
  flex: none;
}
.library-state {
  padding: 20px 0;
}
ul {
  list-style: none;
  padding: 0;
}
li + li {
  border-top: 1px solid var(--journal-rule);
}
li a {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 0;
}
.book-mark {
  color: var(--journal-accent);
}
.saved-entry {
  display: grid;
  gap: 3px;
  min-width: 0;
  overflow-wrap: anywhere;
}
.saved-entry strong {
  font: 600 18px/1.45 var(--journal-serif);
}
.read-arrow {
  margin-left: auto;
  flex: none;
  color: var(--journal-muted);
}
li a:hover,
.empty a {
  color: var(--journal-accent);
}
.empty {
  display: grid;
  justify-items: center;
  gap: 8px;
  text-align: center;
  padding: 32px 0 12px;
  color: var(--journal-muted);
  font-size: 14px;
}
.empty a {
  font-size: 13px;
  text-decoration: underline;
  text-underline-offset: 3px;
}
@media (max-width: 680px) {
  .library {
    grid-template-columns: 1fr;
    gap: 18px;
  }
  aside nav {
    display: flex;
    overflow-x: auto;
  }
  aside nav button {
    flex: none;
    max-width: 90%;
  }
}
</style>
