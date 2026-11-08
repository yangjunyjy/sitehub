import { ref, onMounted, onBeforeUnmount } from "vue";

const key = "deploy-reading-library-v1";
const saved = ref([]),
  history = ref([]),
  persistent = ref(true);
let subscribers = 0;
function entry(article) {
  if (
    !article ||
    !/^[1-9]\d*$/.test(String(article.id)) ||
    typeof article.title !== "string"
  )
    return null;
  return {
    id: String(article.id),
    title: article.title.slice(0, 300),
    summary: String(article.summary || "").slice(0, 500),
    date: String(
      article.publishedAt ||
        article.published_at ||
        article.createdAt ||
        article.created_at ||
        article.date ||
        "",
    ).slice(0, 40),
  };
}
function valid(values, limit) {
  const seen = new Set();
  return (Array.isArray(values) ? values : [])
    .map(entry)
    .filter((item) => item && !seen.has(item.id) && seen.add(item.id))
    .slice(0, limit);
}
function sync() {
  try {
    const data = JSON.parse(localStorage.getItem(key) || "{}");
    saved.value = valid(data.saved, 50);
    history.value = valid(data.history, 12);
  } catch {
    persistent.value = false;
  }
}
function persist() {
  try {
    localStorage.setItem(
      key,
      JSON.stringify({ saved: saved.value, history: history.value }),
    );
    persistent.value = true;
  } catch {
    persistent.value = false;
  }
}
function storageChanged(event) {
  if (event.key === key || event.key === null) sync();
}
sync();
export function useReadingLibrary() {
  onMounted(() => {
    if (subscribers++ === 0) {
      sync();
      window.addEventListener("storage", storageChanged);
    }
  });
  onBeforeUnmount(() => {
    if (--subscribers === 0)
      window.removeEventListener("storage", storageChanged);
  });
  const isSaved = (id) => saved.value.some((item) => item.id === String(id));
  function toggle(article) {
    const item = entry(article);
    if (!item) return;
    saved.value = isSaved(item.id)
      ? saved.value.filter((value) => value.id !== item.id)
      : [item, ...saved.value].slice(0, 50);
    persist();
  }
  function remember(article) {
    const item = entry(article);
    if (!item) return;
    history.value = [
      item,
      ...history.value.filter((value) => value.id !== item.id),
    ].slice(0, 12);
    persist();
  }
  return {
    saved,
    history,
    persistent,
    isSaved,
    toggle,
    remember,
    remove: (id) => {
      saved.value = saved.value.filter((item) => item.id !== String(id));
      persist();
    },
    clearHistory: () => {
      history.value = [];
      persist();
    },
  };
}
