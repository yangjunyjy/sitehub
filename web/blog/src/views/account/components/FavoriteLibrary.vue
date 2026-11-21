<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
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
const router = useRouter(),
  route = useRoute(),
  { expired } = useMember();
const pageSize = 20,
  pages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));
const selected = computed(() =>
  folders.value.find((f) => String(f.id) === active.value),
);
let revision = 0;
function failed(err) {
  if (err.status === 401) {
    expired();
    router.replace({ name: "login", query: { redirect: route.fullPath } });
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
          ＋
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
      <p v-if="busy" role="status">正在读取收藏…</p>
      <p v-else-if="error" role="alert">
        {{ error }} <button @click="init">重试</button>
      </p>
      <ul v-else-if="articles.length">
        <li v-for="article in articles" :key="article.id">
          <RouterLink :to="{ name: 'article', params: { id: article.id } }">
            <strong>{{ article.title }}</strong
            ><small v-if="article.savedAt"
              >收藏于
              {{ new Date(article.savedAt).toLocaleDateString("zh-CN") }}</small
            ></RouterLink
          >
        </li>
      </ul>
      <div v-else class="empty">
        <BlogIcon name="bookmark" :size="24" />
        <p>这个收藏夹还没有文章</p>
        <RouterLink :to="{ name: 'home' }">浏览文章 →</RouterLink>
      </div>
      <nav v-if="!busy && pages > 1" class="pager" aria-label="收藏分页">
        <button :disabled="page <= 1" @click="load(active, page - 1)">
          上一页
        </button>
        <span>{{ page }} / {{ pages }}</span
        ><button :disabled="page >= pages" @click="load(active, page + 1)">
          下一页
        </button>
      </nav>
    </div>
  </section>
</template>
<style scoped>
.library {
  display: grid;
  grid-template-columns: minmax(170px, 235px) minmax(0, 1fr);
  gap: clamp(20px, 5vw, 50px);
  margin-top: 30px;
  color: #27362b;
}
.aside-title,
.saved header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
h2 {
  margin: 0;
  font-size: 19px;
}
.aside-title h2 {
  font-size: 14px;
}
button {
  border: 1px solid #c7d2c7;
  border-radius: 6px;
  background: white;
  color: #35533c;
}
.aside-title button {
  font-size: 19px;
  line-height: 1;
  padding: 3px 6px;
}
aside nav {
  display: grid;
  gap: 4px;
  margin-top: 12px;
}
aside form {
  margin-top: 12px;
}
aside nav button {
  border-color: transparent;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
  padding: 9px;
  font-size: 14px;
}
aside nav button[aria-pressed="true"] {
  background: #e6eee3;
  color: #284b32;
}
aside nav button span {
  flex: 1;
  overflow-wrap: anywhere;
}
.saved {
  background: #fff;
  border: 1px solid #dce3da;
  border-radius: 9px;
  padding: clamp(18px, 3vw, 28px);
}
small {
  color: #647168;
  font-size: 12px;
}
ul {
  list-style: none;
  padding: 0;
  margin-bottom: 0;
}
li {
  padding: 14px 0;
  border-top: 1px solid #dfe6dc;
}
li a {
  display: grid;
  gap: 4px;
}
li a:hover,
.empty a:hover {
  text-decoration: underline;
}
.empty {
  text-align: center;
  padding: 40px 0;
  color: #526257;
}
.empty svg {
  color: #8b9d8c;
}
.empty p {
  margin: 8px 0;
}
.empty a {
  font-size: 13px;
  color: #30563c;
}
.pager {
  display: flex;
  justify-content: center;
  gap: 12px;
  font-size: 13px;
}
@media (max-width: 680px) {
  .library {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  aside nav {
    display: flex;
    overflow: auto;
  }
  aside nav button {
    white-space: nowrap;
  }
  .saved header {
    flex-wrap: wrap;
  }
}
</style>
