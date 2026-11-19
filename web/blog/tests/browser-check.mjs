import assert from "node:assert/strict";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import { createFrontendServer } from "../server/server.mjs";

// Isolated empty API fixtures verify page behavior without touching live data.
const dist = fileURLToPath(new URL("../dist/", import.meta.url));
const artifacts = fileURLToPath(new URL("../.artifacts/", import.meta.url));
await mkdir(artifacts, { recursive: true });
const server = createFrontendServer({ dist });
await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
const origin = `http://127.0.0.1:${server.address().port}`;
const browser = await chromium.launch({
  channel: process.env.PLAYWRIGHT_CHANNEL || "chrome",
  headless: true,
});
const errors = [];
const context = await browser.newContext({
  viewport: { width: 1440, height: 1000 },
});
await context.route("**/v1/**", async (route) => {
  const path = new URL(route.request().url()).pathname;
  if (path.endsWith("/members/me"))
    return route.fulfill({
      status: 401,
      json: { code: 401, message: "请先登录" },
    });
  let data = { items: [], total: 0 };
  if (path.endsWith("/menus/public"))
    data.items = [
      "home",
      "categories",
      "tags",
      "archives",
      "friends",
      "about",
      "messages",
    ].map((name, index) => ({
      menuName: ["首页", "分类", "标签", "归档", "友链", "关于", "留言"][index],
      component: name + "/index",
    }));
  if (path.endsWith("/about")) data = {};
  if (path.endsWith("/articles/detail"))
    data = {
      id: "1",
      title: "Isolated article fixture",
      content_format: "markdown",
      content:
        "# Reader heading\n\nA paragraph.\n\n<script>window.untrustedArticle=true</script>",
      tags: [],
    };
  if (route.request().method() === "POST")
    data = { id: "1", status: "pending" };
  return route.fulfill({ json: { code: 0, message: "ok", data } });
});
const page = await context.newPage();
page.on("pageerror", (error) => errors.push(error.message));
try {
  for (const path of [
    "/",
    "/categories",
    "/tags",
    "/archives",
    "/friends",
    "/about",
    "/messages",
    "/search?q=cloud",
    "/login",
    "/article/1",
    "/missing-page",
  ]) {
    await page.goto(origin + path, { waitUntil: "networkidle" });
    const hero = page.locator(".sky-hero__image");
    if (await hero.count()) {
      await hero.evaluate(async (element) => {
        const url = getComputedStyle(element).backgroundImage.match(/url\("([^"]+)"\)/)?.[1];
        if (!url) throw new Error("Missing backdrop for " + location.pathname);
        const image = new Image();
        image.src = url;
        await image.decode();
      });
    }
    if (path === "/missing-page") await page.getByText("404 · NOT FOUND").waitFor();
    assert(await page.locator("main").innerText(), "Empty main for " + path);
    assert(
      await page
        .getByRole("navigation", { name: "主要导航", exact: true })
        .getByText("首页", { exact: true })
        .count(),
      "Missing dynamic navigation",
    );
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth > innerWidth,
      ),
      false,
      "Desktop horizontal overflow: " + path,
    );
    if (["/", "/messages", "/friends"].includes(path))
      await page.screenshot({
        path: `${artifacts}/${path === "/" ? "home" : path.slice(1)}-desktop.png`,
        fullPage: true,
      });
  }
  await page.goto(origin + "/friends", { waitUntil: "networkidle" });
  await page.getByPlaceholder("你的站点叫什么名字？").fill("Browser test");
  await page
    .getByPlaceholder("https://your-blog.com", { exact: true })
    .fill("https://example.test");
  await page
    .getByPlaceholder("用一句话介绍你的世界")
    .fill("Isolated browser form validation");
  await page.getByRole("button", { name: "提交友链申请 ↗" }).click();
  await page
    .getByText("申请已提交，正在等待博主审核。感谢你的相遇！", { exact: true })
    .waitFor();
  await page.goto(origin + "/messages", { waitUntil: "networkidle" });
  await page
    .getByLabel("留言内容", { exact: true })
    .fill("Isolated browser test");
  await page.getByRole("button", { name: "发送留言", exact: true }).click();
  await page
    .getByText("你的星光已送达，审核通过后会出现在这片星空。", { exact: true })
    .waitFor();
  await page.goto(origin + "/article/1", { waitUntil: "networkidle" });
  await page
    .getByRole("heading", { name: "Reader heading", exact: true })
    .waitFor();
  assert.equal(
    await page.evaluate(() => window.untrustedArticle),
    undefined,
    "Article HTML was not sanitized",
  );
  await context.route("**/v1/blog/messages", (route) =>
    route.request().method() === "POST"
      ? route.fulfill({
          status: 503,
          json: { code: 503, message: "unavailable" },
        })
      : route.fallback(),
  );
  await page.goto(origin + "/messages", { waitUntil: "networkidle" });
  await page
    .getByLabel("留言内容", { exact: true })
    .fill("Keep this after a failed request");
  await page.getByRole("button", { name: "发送留言", exact: true }).click();
  await page
    .getByRole("alert")
    .getByText("服务暂时不可用，请稍后重试", { exact: true })
    .waitFor();
  assert.equal(
    await page.getByLabel("留言内容", { exact: true }).inputValue(),
    "Keep this after a failed request",
  );
  assert.equal(
    await page
      .getByText("你的星光已送达，审核通过后会出现在这片星空。", {
        exact: true,
      })
      .count(),
    0,
  );
  await page.setViewportSize({ width: 390, height: 844 });
  for (const path of [
    "/",
    "/categories",
    "/friends",
    "/messages",
    "/login",
    "/missing-page",
  ]) {
    await page.goto(origin + path, { waitUntil: "networkidle" });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth > innerWidth,
      ),
      false,
      "Mobile horizontal overflow: " + path,
    );
    assert.equal(
      await page
        .locator("input:visible, textarea:visible")
        .evaluateAll((elements) =>
          elements.every((element) => {
            const rect = element.getBoundingClientRect();
            return rect.left >= 0 && rect.right <= innerWidth;
          }),
        ),
      true,
      "Mobile form controls are clipped: " + path,
    );
    await page.getByRole("button", { name: "切换导航", exact: true }).click();
    await page
      .getByRole("navigation", { name: "移动导航", exact: true })
      .getByText("分类", { exact: true })
      .waitFor();
    await page.getByRole("button", { name: "切换导航", exact: true }).click();
    if (["/", "/messages"].includes(path))
      await page.screenshot({
        path: `${artifacts}/${path === "/" ? "home" : "messages"}-mobile.png`,
        fullPage: true,
      });
  }
  assert.deepEqual(errors, []);
  console.log(
    "Blog browser checks passed: directory routes, dynamic navigation, empty content, pending forms and responsive layouts",
  );
} finally {
  await browser.close();
  server.closeAllConnections();
  await new Promise((resolve) => server.close(resolve));
}
