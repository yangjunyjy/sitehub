import {
  computed,
  nextTick,
  onMounted,
  onBeforeUnmount,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { readCatalog, readDocument } from "@/api/documents";
import { renderMarkdown } from "@/utils/markdown";

// Own the complete request/navigation lifecycle; the page renders this state only.
export function useDocumentReader() {
  const route = useRoute(),
    router = useRouter();
  const documentModes = ref([]);
  const selectedMode = ref("");
  const modesReady = ref(false), modesError = ref("");
  const currentMode = computed(() => documentModes.value.find(mode => mode.id === selectedMode.value)
    || { id: "", label: "", available: false });
  const readableMode = computed(() => documentModes.value.find(mode => mode.available));
  const modeLocation = (mode) => ({ name: "docs", params: { id: "" }, query: { mode: mode.id } });
  const navigation = ref({
    title: "Deploy 文档",
    description: "",
    version: "",
    groups: [],
  });
  const document = ref(null),
    html = ref(""),
    toc = ref([]),
    titleHeading = ref(null);
  const search = ref(""),
    loading = ref(true),
    error = ref(""),
    errorKind = ref(""),
    sidebarOpen = ref(false),
    menuButton = ref(null),
    copyStatus = ref("");
  let requestID = 0,
    navigationID = 0,
    loadingID = "",
    navigationLoaded = false,
    disposed = false,
    copyTimer;
  let switchTarget = null;
  let documentController,
    navigationController;
  const versionNotice = ref("");
  const documentAddress = (item) => item.slug || item.id;
  const docLocation = (id) => ({
    name: "docs", params: { id: id || "" },
    query: {
      mode: currentMode.value.id,
      version: navigation.value.versionId || route.query.version,
    },
  });
  const switchVersion = (version) => {
    switchTarget = {
      version,
      id: document.value?.id,
      slug: document.value?.slug,
    };
    versionNotice.value = "";
    router.push({
      path: route.path,
      query: { mode: currentMode.value.id, version },
    });
  };
  const flatDocuments = computed(() =>
    navigation.value.groups.flatMap((group) => group.items || []),
  );
  const filteredGroups = computed(() => {
    const keyword = search.value.trim().toLowerCase();
    return navigation.value.groups
      .map((group) => ({
        ...group,
        items: (group.items || []).filter(
          (item) =>
            !keyword ||
            `${item.title} ${item.summary || ""}`
              .toLowerCase()
              .includes(keyword),
        ),
      }))
      .filter((group) => group.items.length);
  });
  const currentIndex = computed(() =>
    flatDocuments.value.findIndex(
      (item) => String(item.id) === String(document.value?.id),
    ),
  );
  const currentGroup = computed(
    () =>
      navigation.value.groups.find((group) =>
        group.items?.some(
          (item) => String(item.id) === String(document.value?.id),
        ),
      )?.title || "",
  );
  const previousDocument = computed(() =>
    currentIndex.value > 0 ? flatDocuments.value[currentIndex.value - 1] : null,
  );
  const nextDocument = computed(() =>
    currentIndex.value >= 0
      ? flatDocuments.value[currentIndex.value + 1]
      : null,
  );
  const hasToc = computed(
    () => !loading.value && !error.value && toc.value.length > 0,
  );
  const closeMenu = (focus = false) => {
    sidebarOpen.value = false;
    if (focus) menuButton.value?.focus();
  };
  const keydown = (event) => {
    if (event.key === "Escape" && sidebarOpen.value) closeMenu(true);
  };
  const desktop = window.matchMedia("(min-width:761px)");
  const resize = () => {
    if (desktop.matches) closeMenu();
  };
  const clearContent = () => {
    document.value = null;
    html.value = "";
    toc.value = [];
    titleHeading.value = null;
    copyStatus.value = "";
    clearTimeout(copyTimer);
  };

  const loadDocument = async (id) => {
    const current = ++requestID;
    documentController?.abort();
    documentController = new AbortController();
    loadingID = String(id);
    loading.value = true;
    error.value = "";
    errorKind.value = "";
    closeMenu();
    clearContent();
    let phase = "request";
    try {
      const data = await readDocument({
        id, version: navigation.value.versionId, mode: currentMode.value.id, signal: documentController.signal,
      });
      if (disposed || current !== requestID) return;
      phase = "render";
      const rendered = renderMarkdown(data.content);
      const leadingTitle = /^\s*<h1\b[^>]*>[\s\S]*?<\/h1>\s*/.exec(
        rendered.html,
      );
      titleHeading.value = leadingTitle
        ? rendered.headings.find((item) => item.level === 1)
        : null;
      html.value = leadingTitle
        ? rendered.html.slice(leadingTitle[0].length)
        : rendered.html;
      toc.value = rendered.headings.filter((item) => item.level <= 3);
      document.value = { ...data, id: data.id ?? id };
      globalThis.document.title = `${document.value.title}｜${currentMode.value.label}｜${navigation.value.title}`;
      loading.value = false;
      await nextTick();
      if (disposed || current !== requestID) return;
      let anchorID = route.hash.slice(1);
      try {
        anchorID = decodeURIComponent(anchorID);
      } catch {
        // A malformed fragment must not turn a successfully loaded document into an error.
      }
      const anchor = anchorID
        ? globalThis.document.getElementById(anchorID)
        : null;
      if (anchor) anchor.scrollIntoView({ block: "start" });
      else window.scrollTo({ top: 0 });
    } catch (failure) {
      if (disposed || current !== requestID) return;
      clearContent();
      errorKind.value =
        phase === "render"
          ? "render"
          : failure.response?.status === 404
            ? "missing"
            : "document";
      error.value =
        errorKind.value === "render"
          ? "文档暂时无法呈现，请重新加载，或选择其他文档继续阅读。"
          : errorKind.value === "missing"
            ? "这篇文档不存在或已下线，请从目录选择其他文档。"
            : "文档加载失败，请稍后重试。";
    } finally {
      if (current === requestID) {
        loadingID = "";
        loading.value = false;
      }
    }
  };
  const syncDocument = async () => {
    if (!currentMode.value.available || !navigationLoaded || disposed) return;
    const requested = route.params.id;
    const target = requested
      ? flatDocuments.value.find((item) => item.slug === requested) ||
        flatDocuments.value.find(
          (item) => String(item.id) === String(requested),
        )
      : flatDocuments.value[0];
    if (!target) {
      requestID++;
      documentController?.abort();
      loadingID = "";
      loading.value = false;
      clearContent();
      errorKind.value = requested ? "missing" : "empty";
      error.value = requested
        ? "这篇文档不存在，请从左侧目录选择其他文档。"
        : "文档发布后会展示在这里，你可以先了解产品能力。";
      return;
    }
    if (loadingID === String(target.id) || (!error.value && String(document.value?.id) === String(target.id))) return;
    const pending = loadDocument(target.id);
    if (
      String(requested || "") !== String(documentAddress(target)) ||
      route.query.mode !== currentMode.value.id ||
      (navigation.value.versionId &&
        route.query.version !== navigation.value.versionId)
    )
      await router.replace({
        ...docLocation(documentAddress(target)),
        hash: route.hash,
      });
    await pending;
  };
  const loadNavigation = async () => {
    const current = ++navigationID;
    navigationController?.abort();
    documentController?.abort();
    navigationController = new AbortController();
    requestID++;
    loadingID = "";
    loading.value = true;
    error.value = "";
    errorKind.value = "";
    clearContent();
    navigationLoaded = false;
    modesError.value = "";
    try {
      const catalog = await readCatalog({ version: route.query.version, mode: route.query.mode, signal: navigationController.signal });
      if (disposed || current !== navigationID) return;
      documentModes.value = catalog.modes;
      selectedMode.value = catalog.modeId;
      modesReady.value = true;
      navigation.value = catalog.navigation || {title: catalog.title, description: catalog.description, versions: [], groups: []};
      if (!catalog.navigation) {
        loading.value = false;
        return;
      }
      navigationLoaded = true;
      if (switchTarget && switchTarget.version === route.query.version) {
        const target =
          flatDocuments.value.find((item) => item.id === switchTarget.id) ||
          flatDocuments.value.find((item) => item.slug === switchTarget.slug);
        switchTarget = null;
        if (!target && flatDocuments.value.length) {
          versionNotice.value =
            "此版本没有同名文档，已打开该版本的第一篇文档。";
          await router.replace(
            docLocation(documentAddress(flatDocuments.value[0])),
          );
        } else if (
          target &&
          String(route.params.id) !== String(documentAddress(target))
        )
          await router.replace(docLocation(documentAddress(target)));
      }
      await syncDocument();
    } catch (failure) {
      if (disposed || current !== navigationID) return;
      navigation.value.groups = [];
      errorKind.value = "navigation";
      error.value =
        failure.response?.status === 404
          ? "此文档版本不存在或尚未发布，请返回默认版本。"
          : "暂时无法获取文档目录，请检查连接后重试。";
      modesReady.value = false;
      modesError.value = error.value;
      loading.value = false;
    }
  };
  const retry = () =>
    ["navigation", "missing", "empty"].includes(errorKind.value)
      ? loadNavigation()
      : syncDocument();
  const copyMarkdown = async () => {
    const current = requestID;
    try {
      await navigator.clipboard.writeText(document.value?.content || "");
      if (current === requestID) copyStatus.value = "Markdown 已复制";
    } catch {
      if (current === requestID)
        copyStatus.value = "无法访问剪贴板，请手动选择正文复制";
    }
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copyStatus.value = "";
    }, 2400);
  };
  watch(
    () => [route.params.id, route.query.version, route.query.mode],
    (current, previous) => {
      if (!modesReady.value) { loadNavigation(); return; }
      if (current[2] !== previous[2] && current[2] !== selectedMode.value) {
        switchTarget = null;
        versionNotice.value = "";
        search.value = "";
        closeMenu();
        loadNavigation();
        return;
      }
      if (
        current[1] !== previous[1] &&
        (!navigationLoaded || current[1] !== navigation.value.versionId)
      )
        loadNavigation();
      else syncDocument();
    },
  );
  onMounted(() => {
    loadNavigation();
    window.addEventListener("keydown", keydown);
    desktop.addEventListener("change", resize);
  });
  onBeforeUnmount(() => {
    disposed = true;
    requestID++;
    navigationID++;
    documentController?.abort();
    navigationController?.abort();
    clearTimeout(copyTimer);
    window.removeEventListener("keydown", keydown);
    desktop.removeEventListener("change", resize);
  });

  return {
    route,
    documentModes,
    modesReady,
    modesError,
    currentMode,
    readableMode,
    modeLocation,
    navigation,
    document,
    html,
    toc,
    titleHeading,
    search,
    loading,
    error,
    errorKind,
    sidebarOpen,
    menuButton,
    copyStatus,
    versionNotice,
    docLocation,
    switchVersion,
    filteredGroups,
    currentGroup,
    previousDocument,
    nextDocument,
    hasToc,
    closeMenu,
    loadNavigation,
    retry,
    copyMarkdown,
  };
}
