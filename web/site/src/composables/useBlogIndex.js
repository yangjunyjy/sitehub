import { ref, shallowRef, readonly, onMounted, onBeforeUnmount } from "vue";
import { getArticles } from "@/api";
import { collectBlogIndex } from "@/utils/blog-index";

const items = shallowRef([]),
  loading = ref(false),
  error = ref(""),
  loaded = ref(false);
const progress = ref(0),
  total = ref(0);
let job,
  controller,
  generation = 0,
  consumers = 0,
  loadedAt = 0;
function load(force = false) {
  if (!force && job) return job;
  if (!force && loaded.value && Date.now() - loadedAt < 300000)
    return Promise.resolve();
  const current = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = "";
  progress.value = 0;
  total.value = 0;
  job = collectBlogIndex(getArticles, {
    signal: controller.signal,
    onProgress(count, size) {
      if (current === generation) {
        progress.value = count;
        total.value = size;
      }
    },
  })
    .then((value) => {
      if (current !== generation) return;
      items.value = value;
      loaded.value = true;
      loadedAt = Date.now();
    })
    .catch(() => {
      if (current === generation)
        error.value = "文章目录暂时无法完整加载，请重试。";
    })
    .finally(() => {
      if (current === generation) {
        loading.value = false;
        job = null;
      }
    });
  return job;
}

export function useBlogIndex() {
  onMounted(() => {
    consumers++;
    load();
  });
  onBeforeUnmount(() => {
    consumers--;
    queueMicrotask(() => {
      if (consumers || !job) return;
      generation++;
      controller?.abort();
      loading.value = false;
      job = null;
    });
  });
  return {
    items: readonly(items),
    loading: readonly(loading),
    error: readonly(error),
    loaded: readonly(loaded),
    progress: readonly(progress),
    total: readonly(total),
    reload: () => load(true),
  };
}
