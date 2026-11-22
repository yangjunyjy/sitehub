import hljs from "highlight.js/lib/core";
import bash from "highlight.js/lib/languages/bash";
import go from "highlight.js/lib/languages/go";
import javascript from "highlight.js/lib/languages/javascript";
import typescript from "highlight.js/lib/languages/typescript";
import json from "highlight.js/lib/languages/json";
import yaml from "highlight.js/lib/languages/yaml";
import xml from "highlight.js/lib/languages/xml";
import css from "highlight.js/lib/languages/css";
import sql from "highlight.js/lib/languages/sql";
import python from "highlight.js/lib/languages/python";
hljs.registerLanguage("bash", bash);
hljs.registerLanguage("go", go);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("json", json);
hljs.registerLanguage("yaml", yaml);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("css", css);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("python", python);

// Markup supplied by an author has already been sanitized. Highlight plain source only.
export function highlightCode(code, language) {
  if (
    !language ||
    !hljs.getLanguage(language) ||
    code.textContent.length > 50000
  )
    return;
  if (code.querySelector(":not(span, br), [id]")) return;
  for (const br of code.querySelectorAll("br")) br.replaceWith("\n");
  code.innerHTML = hljs.highlight(code.textContent, {
    language,
    ignoreIllegals: true,
  }).value;
}
