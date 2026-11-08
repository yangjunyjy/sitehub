// Both applications use this renderer; HTML is sanitized before entering any live DOM.
export function createArticleRenderer(Marked, DOMPurify, hljs) {
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
    if (format !== "markdown") throw new Error("Only Markdown content is supported");
    const cacheKey = source;
    const cached = cache.get(cacheKey);
    if (cached) {
      cache.delete(cacheKey);
      cache.set(cacheKey, cached);
      return copyResult(cached);
    }

    const headings = [],
      counts = new Map(),
      usedIDs = new Set();
    const escape = (s) =>
      String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;");
    const idFor = (text) => {
      const base =
        text
          .toLowerCase()
          .replace(/[^\p{L}\p{N}]+/gu, "-")
          .replace(/^-|-$/g, "") || "section";
      let count = counts.get(base) || 0;
      let id = `heading-${base}${count ? `-${count}` : ""}`;
      // A literal "foo-1" heading may already own the suffix that a repeated
      // "foo" heading would otherwise receive. Every anchor must stay unique.
      while (usedIDs.has(id)) {
        count++;
        id = `heading-${base}-${count}`;
      }
      counts.set(base, count + 1);
      usedIDs.add(id);
      return id;
    };
    const parser = new Marked({ gfm: true, breaks: true });
    parser.use({
      renderer: {
        heading(token) {
          const depth = token?.depth;
          const level = Number.isInteger(depth) && depth >= 1 && depth <= 6 ? depth : 1;
          const inline = Array.isArray(token?.tokens)
            ? this.parser.parseInline(token.tokens)
            : parser.parseInline(typeof token?.text === "string" ? token.text : "");
          // Extract visible text, including decoded entities and inline formatting, for the TOC.
          const fragment = DOMPurify.sanitize(inline, {
            RETURN_DOM_FRAGMENT: true,
          });
          const text = fragment.textContent || "";
          const id = idFor(text);
          headings.push({ level, text, slug: id });
          return `<h${level} id="${escape(id)}">${inline}</h${level}>\n`;
        },
        code(token) {
          const source = typeof token?.text === "string" ? token.text : "";
          const language = String(token?.lang || "").split(/\s/)[0];
          const code =
            language && hljs?.getLanguage(language)
              ? hljs.highlight(source, { language }).value
              : escape(source);
          return `<div class="code-block-wrapper"><div class="code-header"><span>${escape(language || "text")}</span><button class="code-copy-btn" type="button">复制</button></div><pre class="code-content"><code class="hljs language-${escape(language)}">${code}</code></pre></div>`;
        },
      },
    });
    let html = parser.parse(source);
    html = DOMPurify.sanitize(html, {
      USE_PROFILES: { html: true },
      FORBID_TAGS: [
        "style",
        "iframe",
        "form",
        "input",
        "video",
        "audio",
        "object",
        "embed",
      ],
      FORBID_ATTR: ["style"],
      ADD_ATTR: ["target"],
    });
    const result = { html, headings };
    const characters = cacheKey.length + html.length +
      headings.reduce((total, heading) => total + heading.text.length + heading.slug.length, 0);
    if (characters <= maxCharacters) {
      while (cache.size >= maxEntries || cachedCharacters + characters > maxCharacters) {
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
