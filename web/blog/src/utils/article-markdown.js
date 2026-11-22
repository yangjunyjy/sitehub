import { highlightCode } from "./code-highlight";

// HTML is sanitized before entering the blog's live DOM.
function indexHeadings(fragment) {
  const headings = [];
  const used = new Set();
  const aliases = new Map();
  const nextSuffix = new Map();
  const nodes = fragment.querySelectorAll("h1,h2,h3,h4,h5,h6,[id]");
  for (const node of nodes) {
    const original = node.id;
    const heading = /^H[1-6]$/.test(node.tagName);
    const text = node.textContent.trim();
    const label =
      (heading ? text : original)
        .toLowerCase()
        .replace(/[^\p{L}\p{N}]+/gu, "-")
        .replace(/^-+|-+$/g, "") || "section";
    const base = (heading ? "heading-" : "anchor-") + label;
    let suffix = nextSuffix.get(base) || 0;
    let id = suffix ? base + "-" + suffix : base;
    while (used.has(id)) id = base + "-" + ++suffix;
    nextSuffix.set(base, suffix + 1);
    used.add(id);
    node.id = id;
    if (original && !aliases.has(original)) aliases.set(original, id);
    if (heading && text)
      headings.push({ level: Number(node.tagName[1]), text, slug: id });
  }
  for (const link of fragment.querySelectorAll('a[href^="#"]')) {
    const original = link.getAttribute("href").slice(1);
    let decoded;
    try {
      decoded = decodeURIComponent(original);
    } catch {
      continue;
    }
    if (aliases.has(decoded))
      link.setAttribute("href", "#" + aliases.get(decoded));
  }
  return headings;
}

export function createArticleRenderer(Marked, DOMPurify) {
  const parser = new Marked({ gfm: true, breaks: true });
  // Switching articles or toggling the editor preview should not parse/highlight
  // identical content again. Bound both entry count and retained text size.
  const cache = new Map();
  const maxEntries = 4;
  const maxCharacters = 400_000;
  let cachedCharacters = 0;
  const copyResult = ({ html, headings }) => ({
    html,
    headings: headings.map((heading) => ({ ...heading })),
  });

  return (source = "", format = "markdown") => {
    if (source == null) source = "";
    if (typeof source !== "string")
      throw new TypeError("Markdown content must be a string");
    if (format !== "markdown" && format !== "html")
      throw new Error("Unsupported article format");
    const cacheKey = format + "\0" + source;
    const cached = cache.get(cacheKey);
    if (cached) {
      cache.delete(cacheKey);
      cache.set(cacheKey, cached);
      return copyResult(cached);
    }

    const markup = format === "html" ? source : parser.parse(source);
    const fragment = DOMPurify.sanitize(markup, {
      RETURN_DOM_FRAGMENT: true,
      ALLOWED_TAGS:
        "a abbr article aside b blockquote br caption cite code col colgroup dd del details div dl dt em figcaption figure h1 h2 h3 h4 h5 h6 hr i img kbd li mark ol p pre s section small span strong sub summary sup table tbody td th thead tfoot tr u ul".split(
          " ",
        ),
      ALLOWED_ATTR:
        "align alt class colspan dir height href id lang open rowspan scope src start title width".split(
          " ",
        ),
      ALLOW_ARIA_ATTR: false,
      ALLOW_DATA_ATTR: false,
    });
    // Only the language on code blocks affects presentation. Article markup
    // must never acquire controls or layout styles from the surrounding site.
    for (const node of fragment.querySelectorAll("[class]")) {
      const language =
        node.matches("pre > code") &&
        [...node.classList].find((value) => /^language-[\w.+#-]+$/.test(value));
      node.removeAttribute("class");
      if (language) node.className = language;
    }
    // Marked emits GFM column alignment as an HTML attribute. Only keep the three cell values.
    for (const node of fragment.querySelectorAll("[align]")) {
      const alignment = node.getAttribute("align").toLowerCase();
      node.removeAttribute("align");
      if (
        node.matches("th,td") &&
        ["left", "center", "right"].includes(alignment)
      )
        node.className = "align-" + alignment;
    }
    const headings = indexHeadings(fragment);
    for (const table of fragment.querySelectorAll("table")) {
      const scroll = document.createElement("div");
      scroll.className = "article-table";
      scroll.tabIndex = 0;
      scroll.setAttribute("role", "region");
      scroll.setAttribute(
        "aria-label",
        table.caption?.textContent || "文章表格",
      );
      table.replaceWith(scroll);
      scroll.append(table);
    }
    for (const pre of fragment.querySelectorAll("pre")) {
      const code = pre.querySelector(":scope > code");
      const languageName = code?.className.replace(/^language-/, "") || "";
      if (code) highlightCode(code, languageName.toLowerCase());
      const wrapper = document.createElement("figure");
      wrapper.className = "code-block-wrapper";
      const bar = document.createElement("figcaption");
      bar.className = "code-header";
      const language = document.createElement("span");
      language.textContent = languageName || "纯文本";
      const actions = document.createElement("div");
      actions.className = "code-actions";
      const wrap = document.createElement("button");
      wrap.type = "button";
      wrap.className = "code-wrap-btn";
      wrap.setAttribute("aria-pressed", "false");
      wrap.textContent = "自动换行";
      const button = document.createElement("button");
      button.type = "button";
      button.className = "code-copy-btn";
      button.textContent = "复制代码";
      actions.append(wrap, button);
      bar.append(language, actions);
      pre.tabIndex = 0;
      pre.replaceWith(wrapper);
      wrapper.append(bar, pre);
    }
    for (const image of fragment.querySelectorAll("img")) {
      image.loading = "lazy";
      image.decoding = "async";
    }
    const container = document.createElement("div");
    container.append(fragment);
    const html = container.innerHTML;
    const result = { html, headings };
    const characters =
      cacheKey.length +
      html.length +
      headings.reduce(
        (total, heading) => total + heading.text.length + heading.slug.length,
        0,
      );
    if (characters <= maxCharacters) {
      while (
        cache.size >= maxEntries ||
        cachedCharacters + characters > maxCharacters
      ) {
        const oldestKey = cache.keys().next().value;
        cachedCharacters -= cache.get(oldestKey).characters;
        cache.delete(oldestKey);
      }
      cache.set(cacheKey, { ...result, characters });
      cachedCharacters += characters;
      return copyResult(result);
    }
    return result;
  };
}
