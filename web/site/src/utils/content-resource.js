// Keep request ownership outside components so every consumer shares one snapshot.
export function createContentResource(
  state,
  fetchContent,
  normalize = (value) => value,
) {
  let pending;
  let controller;
  let revision = 0;

  return function load(force = false) {
    if (pending && !force) return pending;
    if (state.loaded && !force) return Promise.resolve(state.content);

    const current = ++revision;
    controller?.abort();
    controller = new AbortController();
    const signal = controller.signal;
    state.loading = true;
    state.error = "";

    pending = Promise.resolve()
      .then(() => fetchContent({ signal }))
      .then((response) => {
        if (current !== revision) return state.content;
        const content = normalize(response.data);
        state.content = content;
        state.loaded = true;
        return content;
      })
      .catch((error) => {
        // Superseded requests may settle even when the transport ignores abort.
        if (current !== revision) return state.content;
        state.error = error.message || "官网内容加载失败";
        throw error;
      })
      .finally(() => {
        if (current !== revision) return;
        state.loading = false;
        pending = undefined;
      });
    return pending;
  };
}

export function emptyWebsiteContent() {
  return {
    brand: {
      name: "Deploy",
      fullName: "Deploy 企业一体化 DevOps 运维平台",
      description: "",
      contactEmail: "",
      recordNumber: "",
      copyright: "",
    },
    navigation: [
      { label: "首页", path: "/" },
      { label: "下载", path: "/download" },
      { label: "文档", path: "/docs" },
      { label: "博客", path: "/blog" },
    ],
    home: { meta: [] },
    capabilities: [],
    deliverySteps: [],
    audiences: [],
    product: { architectureLayers: [], securityItems: [] },
    about: { paragraphs: [], principles: [], teamRoles: [] },
    footer: {},
  };
}

export function normalizeWebsiteContent(content) {
  if (!content || typeof content !== "object" || Array.isArray(content)) {
    throw new Error("官网内容格式异常，请重新加载");
  }
  const fallback = emptyWebsiteContent();
  const result = { ...fallback, ...content };
  for (const key of ["brand", "home", "product", "about", "footer"]) {
    result[key] = { ...fallback[key], ...content[key] };
    for (const [field, value] of Object.entries(fallback[key])) {
      if (Array.isArray(value) && !Array.isArray(result[key][field]))
        result[key][field] = [];
    }
  }
  for (const key of [
    "navigation",
    "capabilities",
    "deliverySteps",
    "audiences",
  ]) {
    if (!Array.isArray(result[key])) result[key] = fallback[key];
  }
  return result;
}
