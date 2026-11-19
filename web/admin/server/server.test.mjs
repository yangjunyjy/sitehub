import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, writeFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { runInNewContext } from "node:vm";
import { createFrontendServer } from "./server.mjs";
import { gatewayProxy } from "./dev-proxy.mjs";

test("production Node serves pages but never proxies or falls back for API requests", async (t) => {
  const dist = await mkdtemp(join(tmpdir(), "sitehub-web-"));
  await writeFile(join(dist, "index.html"), "<html>native frontend</html>");
  const server = createFrontendServer({ dist });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const origin = `http://127.0.0.1:${server.address().port}`;
  t.after(async () => {
    server.closeAllConnections();
    await new Promise((resolve) => server.close(resolve));
    await rm(dist, { recursive: true, force: true });
  });
  assert.match(
    await (await fetch(origin + "/blog/reading")).text(),
    /native frontend/,
  );
  for (const path of [
    "/v1",
    "/v1/rbac/auth/login",
    "/v1/unknown",
    "/api/v1/articles",
    "/uploads/a",
    "/assets/missing.js",
  ]) {
    assert.equal((await fetch(origin + path)).status, 404, path);
  }
  assert.deepEqual(await (await fetch(origin + "/healthz")).json(), {
    status: "ok",
    app: "admin",
  });
  assert.equal((await fetch(origin + "/", { method: "POST" })).status, 405);
});

test("Vite forwards only to one gateway with the admin host audience", () => {
  assert.equal(gatewayProxy()["/v1"].headers.host, "admin.sitehub.local");
  const proxy = gatewayProxy({
    GATEWAY_URL: "http://gateway.test:9080",
    RBAC_API_URL: "http://ignored.test",
  });
  assert.equal(proxy["/v1"].target, "http://gateway.test:9080");
  assert.equal(proxy["/v1"].headers.host, "admin.sitehub.local");
  assert.throws(() =>
    gatewayProxy({ GATEWAY_URL: "http://secret:password@gateway" }),
  );
  assert.throws(() => gatewayProxy({ GATEWAY_HOST: "bad\r\nhost" }));
});

async function runtimeFrontend(t, options = {}) {
  const dist = await mkdtemp(join(tmpdir(), "sitehub-runtime-"));
  await writeFile(join(dist, "index.html"), "<html>frontend</html>");
  const server = createFrontendServer({ dist, ...options });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  t.after(async () => {
    server.closeAllConnections();
    await new Promise((resolve) => server.close(resolve));
    await rm(dist, { recursive: true, force: true });
  });
  return `http://127.0.0.1:${server.address().port}`;
}

function readRuntime(script) {
  const context = { window: {} };
  runInNewContext(script, context);
  return JSON.parse(JSON.stringify(context.window.__SITEHUB_CONFIG__));
}

test("runtime configuration defaults to same-origin and is never cached", async (t) => {
  const origin = await runtimeFrontend(t, { websiteURL: "" });
  const response = await fetch(origin + "/sitehub-runtime.js");
  assert.equal(response.status, 200);
  assert.equal(
    response.headers.get("content-type"),
    "application/javascript; charset=utf-8",
  );
  assert.equal(response.headers.get("cache-control"), "no-store");
  assert.equal(response.headers.get("x-content-type-options"), "nosniff");
  assert.equal(response.headers.get("etag"), null);
  assert.deepEqual(readRuntime(await response.text()), { websiteURL: "" });
  const head = await fetch(origin + "/sitehub-runtime.js", { method: "HEAD" });
  assert.equal(head.status, 200);
  assert.equal(head.headers.get("cache-control"), "no-store");
  assert.equal(await head.text(), "");
  assert.equal(
    (await fetch(origin + "/sitehub-runtime.js", { method: "POST" })).status,
    405,
  );
});

test("runtime configuration escapes script delimiters and ignores request overrides", async (t) => {
  const websiteURL =
    'https://site.example/preview?next=</script><script>alert("x")</script>&line=\u2028\u2029&end=1';
  const origin = await runtimeFrontend(t, { websiteURL });
  const script = await (
    await fetch(
      origin + "/sitehub-runtime.js?websiteURL=https://ignored.example",
    )
  ).text();
  assert.equal(script.includes("<"), false);
  assert.equal(script.includes("\u2028"), false);
  assert.equal(script.includes("\u2029"), false);
  assert.match(script, /\\u003c\/script>/);
  assert.deepEqual(readRuntime(script), { websiteURL });
});

test("runtime website URL is read from the environment and validated at startup", async (t) => {
  const previous = process.env.WEB_SITE_URL;
  process.env.WEB_SITE_URL = "https://site.example";
  t.after(() => {
    if (previous === undefined) delete process.env.WEB_SITE_URL;
    else process.env.WEB_SITE_URL = previous;
  });
  const origin = await runtimeFrontend(t);
  assert.deepEqual(
    readRuntime(await (await fetch(origin + "/sitehub-runtime.js")).text()),
    { websiteURL: "https://site.example" },
  );
  const dist = await mkdtemp(join(tmpdir(), "sitehub-invalid-runtime-"));
  await writeFile(join(dist, "index.html"), "<html>frontend</html>");
  t.after(() => rm(dist, { recursive: true, force: true }));
  for (const websiteURL of [
    "/relative",
    "//site.example",
    "javascript:alert(1)",
    "ftp://site.example",
    "http:site.example",
    "https://user:secret@site.example",
  ]) {
    assert.throws(
      () => createFrontendServer({ dist, websiteURL }),
      /Invalid WEB_SITE_URL/,
      websiteURL,
    );
  }
});
