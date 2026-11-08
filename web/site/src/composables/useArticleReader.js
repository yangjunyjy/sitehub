import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { getArticleDetail, incrementArticleView } from "@/api";

import { renderMarkdown } from "@/utils/markdown";

// Own the complete request/navigation lifecycle; the page renders this state only.
export function useArticleReader() {
  const route = useRoute();
  const article = ref(null);
  const loading = ref(true);
  const error = ref("");
  const notFound = ref(false);
  const copyStatus = ref("");
  const toc = ref([]);

  const renderedContent = ref("");
  let requestId = 0;
  let controller;
  let copyTimer;

  const loadArticle = async () => {
    const currentRequest = ++requestId;
    controller?.abort();
    controller = new AbortController();
    const articleId = route.params.id;
    loading.value = true;
    error.value = "";
    notFound.value = false;
    copyStatus.value = "";
    clearTimeout(copyTimer);
    article.value = null;
    toc.value = [];
    renderedContent.value = "";

    try {
      const response = await getArticleDetail(articleId, {
        signal: controller.signal,
      });
      if (currentRequest !== requestId) return;
      if (!response.data || typeof response.data.content !== "string")
        throw new Error("Invalid article response");
      article.value = response.data;
      const rendered = renderMarkdown(
        article.value?.content || "",
        article.value.content_format,
      );
      renderedContent.value = rendered.html;
      toc.value = rendered.headings;
      document.title = `${article.value?.title || "文章"}｜Deploy`;
      incrementArticleView(articleId).catch(() => {});
    } catch (failure) {
      if (currentRequest !== requestId) return;
      article.value = null;
      notFound.value = failure.response?.status === 404;
      error.value = notFound.value
        ? "这篇文章不存在或已下线，可以返回博客看看其他内容。"
        : "文章暂时无法加载，请稍后重试。";
    } finally {
      if (currentRequest === requestId) loading.value = false;
    }
  };

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      copyStatus.value = "链接已复制";
    } catch {
      copyStatus.value = "请复制浏览器地址栏中的链接";
    }
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copyStatus.value = "";
    }, 2400);
  };

  onMounted(loadArticle);
  watch(() => route.params.id, loadArticle);
  onBeforeUnmount(() => {
    requestId++;
    controller?.abort();
    clearTimeout(copyTimer);
  });

  return {
    article,
    loading,
    error,
    notFound,
    copyStatus,
    toc,
    renderedContent,
    loadArticle,
    copyLink,
  };
}
