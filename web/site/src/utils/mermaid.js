import DOMPurify from "dompurify";

let loader;
let diagramID = 0;

function loadMermaid() {
  // 普通文章不加载图表依赖；资源加载失败后，重新打开页面可以重试。
  loader ||= import("mermaid")
    .then(({ default: mermaid }) => {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: "strict",
        suppressErrorRendering: true,
        htmlLabels: false,
        flowchart: { htmlLabels: false },
        fontFamily: "Arial, 'Microsoft YaHei', sans-serif",
        maxTextSize: 50000,
        maxEdges: 500,
        // 正文中的 init 指令不能放宽安全配置或注入页面样式。
        secure: [
          "secure", "securityLevel", "startOnLoad", "suppressErrorRendering",
          "maxTextSize", "maxEdges", "htmlLabels", "flowchart",
          "themeCSS", "themeVariables", "fontFamily",
        ],
      });
      return mermaid;
    })
    .catch((error) => {
      loader = undefined;
      throw error;
    });
  return loader;
}

// root 中的 HTML 已由 Markdown 渲染器净化。保留原 code 文本，避免实体转义两次
// 或把生成的 SVG 当成图表源码；页面切换后的异步结果不得写入新正文。
export async function renderMermaidBlocks(root, isCurrent = () => true) {
  const blocks = [...root.querySelectorAll("pre > code")].filter((code) =>
    [...code.classList].some((name) => name.toLowerCase() === "language-mermaid"),
  );
  for (const code of blocks) {
    if (!isCurrent() || !root.isConnected) return;
    const wrapper = code.closest(".code-block-wrapper");
    if (!wrapper || wrapper.dataset.diagramState) continue;
    wrapper.dataset.diagramState = "loading";
    wrapper.classList.add("mermaid-block");
    const status = document.createElement("p");
    status.className = "mermaid-status";
    status.setAttribute("role", "status");
    status.textContent = "正在生成图表…";
    wrapper.append(status);
    const preview = document.createElement("div");
    preview.className = "mermaid-preview";
    preview.setAttribute("role", "img");
    preview.setAttribute("aria-label", "Mermaid 图表");
    try {
      const mermaid = await loadMermaid();
      if (!isCurrent() || !wrapper.isConnected) return;
      wrapper.append(preview);
      const { svg } = await mermaid.render(
        `reading-diagram-${++diagramID}`,
        code.textContent,
        preview,
      );
      if (!isCurrent() || !wrapper.isConnected) return;
      preview.innerHTML = DOMPurify.sanitize(svg, {
        USE_PROFILES: { svg: true, svgFilters: true },
        FORBID_TAGS: ["foreignObject", "script"],
      });
      const source = document.createElement("details");
      source.className = "mermaid-source";
      const summary = document.createElement("summary");
      summary.textContent = "查看图表源码";
      source.append(summary, code.parentElement);
      wrapper.append(source);
      status.remove();
      wrapper.dataset.diagramState = "ready";
    } catch {
      preview.remove();
      if (!isCurrent() || !wrapper.isConnected) return;
      wrapper.dataset.diagramState = "error";
      status.textContent = "图表未能显示，可查看或复制上方源码。";
    }
  }
}
