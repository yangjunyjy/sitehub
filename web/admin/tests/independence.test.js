import test from "node:test";
import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "vue/compiler-sfc";

const app = fileURLToPath(new URL("../", import.meta.url));
function files(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const file = join(directory, entry.name);
    return entry.isDirectory() ? files(file) : [file];
  });
}
test("admin sources stay inside the application and all authored styles use Sass", () => {
  const sources = ["src", "server", "scripts"]
    .flatMap((directory) => files(join(app, directory)))
    .concat(join(app, "vite.config.js"));
  for (const file of sources) {
    if (!/\.(?:vue|js|mjs|scss|css)$/.test(file)) continue;
    assert(!file.endsWith(".css"), relative(app, file));
    const source = readFileSync(file, "utf8");
    for (const match of source.matchAll(
      /(?:\bfrom\s+|\bimport\s*\(\s*|\bimport\s+)["']([^"']+)["']/g,
    )) {
      if (!match[1].startsWith(".")) continue;
      const target = relative(app, resolve(dirname(file), match[1]));
      assert(
        !target.startsWith(".."),
        `${relative(app, file)} imports ${match[1]}`,
      );
    }
    if (file.endsWith(".vue")) {
      const { descriptor, errors } = parse(source, { filename: file });
      assert.deepEqual(errors, []);
      for (const style of descriptor.styles)
        assert.equal(style.lang, "scss", relative(app, file));
    }
    if (file.endsWith(".scss"))
      assert(!/@import\b/.test(source), relative(app, file));
  }
});
