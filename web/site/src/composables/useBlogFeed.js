import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getArticles, getCategories, getTags } from "@/api";

// Query state, cancellation, pagination and filter metadata are owned together.
export function useBlogFeed() {
  const route = useRoute(),
    router = useRouter();
  const articles = ref([]),
    categories = ref([]),
    loading = ref(true),
    error = ref("");
  const categoryError = ref(false),
    categoryLoading = ref(false),
    total = ref(0),
    keywordInput = ref("");
  const pageSize = 10;
  const view = ref("list");
  try {
    if (localStorage.getItem("deploy-blog-view") === "grid")
      view.value = "grid";
  } catch {}
  watch(view, (value) => {
    try {
      localStorage.setItem("deploy-blog-view", value);
    } catch {}
  });
  const tags = ref([]),
    tagError = ref(false),
    tagLoading = ref(false),
    initialized = ref(false);
  const cache = new Map();
  let controller;
  const selectedTags = computed(() =>
    filters.value.tags.split(",").filter(Boolean),
  );
  const tagName = (id) =>
    tags.value.find((item) => String(item.id) === id)?.name || `标签 ${id}`;
  const toggleTag = (id) => {
    const key = String(id);
    const next = selectedTags.value.includes(key)
      ? selectedTags.value.filter((item) => item !== key)
      : [...selectedTags.value, key];
    updateFilters({ tags: next.join(","), page: 1 });
  };
  const loadTags = async () => {
    tagLoading.value = true;
    tagError.value = false;
    try {
      const response = await getTags();
      tags.value = Array.isArray(response.data)
        ? response.data
        : response.data?.items || [];
    } catch {
      tagError.value = true;
    } finally {
      tagLoading.value = false;
    }
  };
  let requestId = 0,
    categoryRequestId = 0;
  const filters = computed(() => ({
    tags: [
      ...new Set(
        String(route.query.tags || "")
          .split(",")
          .filter((id) => /^[1-9]\d*$/.test(id)),
      ),
    ]
      .slice(0, 100)
      .sort()
      .join(","),
    category: /^\d+$/.test(String(route.query.category || ""))
      ? String(route.query.category)
      : "",
    keyword:
      typeof route.query.q === "string"
        ? route.query.q.trim().slice(0, 200)
        : "",
    sort: route.query.sort === "hot" ? "hot" : "latest",
    page: /^\d+$/.test(String(route.query.page))
      ? Math.max(1, Math.min(100000, Number(route.query.page)))
      : 1,
  }));
  const totalPages = computed(() =>
    Math.max(1, Math.ceil(total.value / pageSize)),
  );
  const pageNumbers = computed(() => {
    const count = Math.min(5, totalPages.value);
    const start = Math.max(
      1,
      Math.min(filters.value.page - 2, totalPages.value - count + 1),
    );
    return Array.from({ length: count }, (_, index) => start + index);
  });
  const categoryName = computed(
    () =>
      categories.value.find(
        (item) => String(item.id) === filters.value.category,
      )?.name || "分类筛选",
  );
  const updateFilters = (patch, replace = false) => {
    const next = { ...filters.value, ...patch };
    const query = {};
    if (next.keyword) query.q = next.keyword;
    if (next.category) query.category = next.category;
    if (next.tags) query.tags = next.tags;
    if (next.sort !== "latest") query.sort = next.sort;
    if (next.page > 1) query.page = String(next.page);
    return router[replace ? "replace" : "push"]({ path: "/blog", query });
  };
  const load = async () => {
    const id = ++requestId;
    controller?.abort();
    controller = new AbortController();
    const key = JSON.stringify(filters.value);
    const cached = cache.get(key);
    if (cached) {
      articles.value = cached.items;
      total.value = cached.total;
      initialized.value = true;
    }
    loading.value = true;
    error.value = "";
    try {
      const current = filters.value;
      const params = { page: current.page, pageSize, sort: current.sort };
      if (current.category) params.categoryId = current.category;
      if (current.keyword) params.keyword = current.keyword;
      if (current.tags) params.tagIds = current.tags;
      const response = await getArticles(params, { signal: controller.signal });
      if (id !== requestId) return;
      articles.value = Array.isArray(response.data?.items)
        ? response.data.items
        : [];
      total.value = Number(response.data?.total ?? articles.value.length);
      cache.set(key, { items: articles.value, total: total.value });
      if (cache.size > 20) cache.delete(cache.keys().next().value);
      if (current.page > totalPages.value)
        await updateFilters({ page: totalPages.value }, true);
    } catch {
      if (id !== requestId) return;
      articles.value = [];
      total.value = 0;
      error.value = "暂时无法获取文章，请稍后重试。你的筛选条件已保留。";
    } finally {
      if (id === requestId) {
        loading.value = false;
        initialized.value = true;
      }
    }
  };
  const loadCategories = async () => {
    const id = ++categoryRequestId;
    categoryLoading.value = true;
    categoryError.value = false;
    try {
      const response = await getCategories();
      if (id === categoryRequestId)
        categories.value = response.data?.items || [];
    } catch {
      if (id === categoryRequestId) categoryError.value = true;
    } finally {
      if (id === categoryRequestId) categoryLoading.value = false;
    }
  };
  const search = () =>
    updateFilters({
      keyword: keywordInput.value.trim().slice(0, 200),
      page: 1,
    });
  const reset = () =>
    updateFilters({
      keyword: "",
      category: "",
      tags: "",
      sort: "latest",
      page: 1,
    });
  const changePage = (value) => {
    if (!loading.value && value >= 1 && value <= totalPages.value)
      updateFilters({ page: value });
  };
  watch(
    () => route.fullPath,
    () => {
      keywordInput.value = filters.value.keyword;
      load();
    },
    { immediate: true },
  );
  onMounted(() => {
    loadCategories();
    loadTags();
  });
  onBeforeUnmount(() => {
    controller?.abort();
    requestId++;
    categoryRequestId++;
  });

  return {
    articles,
    categories,
    loading,
    error,
    categoryError,
    categoryLoading,
    total,
    keywordInput,
    tags,
    tagError,
    tagLoading,
    initialized,
    filters,
    totalPages,
    categoryName,
    selectedTags,
    tagName,
    toggleTag,
    loadTags,
    updateFilters,
    load,
    loadCategories,
    search,
    reset,
    changePage,
    view,
    pageNumbers,
  };
}
