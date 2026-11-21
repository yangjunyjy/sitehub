import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";

const lead =
  "春末的林道上，天色比手表先告诉我们时间。沿途的三次停留，留下了山风、云影和溪流各自的节奏。";
const picture = "/reader-test/forest.webp";
const columns = Array.from({ length: 10 }, (_, i) => "采样点 " + (i + 1));
const html = [
  '<h2 id="trail">林间的第一束光</h2>',
  "<p>" + lead + "</p>",
  "<blockquote><p>“抵达一个地方之前，先听听它的声音。”</p></blockquote>",
  "<h3>沿着溪声走</h3><ol><li>记录溪水的位置<ul><li>看清水面的反光</li></ul></li><li>标记观察时间</li></ol>",
  '<h3>沿着溪声走</h3><p><a href="#trail">回到林间</a></p>',
  '<figure><img src="' +
    picture +
    '" alt="晨光里的森林"><figcaption>林场以东，天刚放亮。</figcaption></figure>',
  "<h2>路线记录</h2><table><caption>沿途观测</caption><thead><tr>" +
    columns.map((c) => "<th>" + c + "</th>").join("") +
    "</tr></thead><tbody><tr>" +
    columns.map(() => "<td>树影</td>").join("") +
    "</tr></tbody></table>",
  '<pre><code class="language-go">trail := "' +
    "forest/".repeat(19) +
    '"</code></pre>',
  "<h2>山脊上的风</h2>" + ("<p>" + lead + "</p>").repeat(40),
].join("");
const markdown = [
  "## 林间的第一束光",
  lead,
  "> 抵达一个地方之前，先听听它的声音。",
  "### 沿着溪声走",
  "1. 记录溪水的位置\n   - 看清水面的反光\n2. 标记观察时间",
  "### 沿着溪声走",
  "![晨光里的森林](" + picture + ")",
  "## 路线记录",
  "| " +
    columns.join(" | ") +
    " |\n| " +
    columns.map(() => "---").join(" | ") +
    " |\n| " +
    columns.map(() => "树影").join(" | ") +
    " |",
  '~~~go\ntrail := "' + "forest/".repeat(19) + '"\n~~~',
  "## 山脊上的风",
  Array.from({ length: 40 }, () => lead).join("\n\n"),
].join("\n\n");

export function readerArticle(format) {
  return {
    id: format === "html" ? "3" : "2",
    title: "林间行记：在天亮以前出发",
    summary: "沿着旧林道走到山脊，用一份清晨的观察手记，记下这段缓慢的徒步。",
    author: { name: "林舟" },
    category: { id: "1", name: "自然手记" },
    published_at: "2026-05-16",
    content_format: format,
    content: format === "html" ? html : markdown,
    tags: [],
  };
}

export async function checkReaderLayout({
  browser,
  page,
  apiFixture,
  artifacts,
  origin,
}) {
  const photo = (route) =>
    route.fulfill({
      path: fileURLToPath(
        new URL("../src/assets/skies/forest.webp", import.meta.url),
      ),
    });
  await page.route("**" + picture, photo);
  await page.setViewportSize({ width: 1440, height: 940 });
  await page.goto(origin + "/article/2", { waitUntil: "networkidle" });
  const dock = page.locator("#reader-action-panel");
  const trigger = page.getByRole("button", { name: "文章互动工具" });
  const body = page.locator(".markdown-body");
  const unique = await body
    .locator("h2,h3")
    .evaluateAll((nodes) => nodes.map((n) => n.id));
  assert.equal(unique.length, 5);
  assert.equal(new Set(unique).size, 5);
  assert.equal(await body.locator("li ul li").count(), 1);
  assert.equal(await body.locator("blockquote").count(), 1);
  assert.equal(await body.locator(".code-header span").textContent(), "go");
  assert(
    await body.locator("img").evaluate(async (img) => {
      await img.decode();
      return img.naturalWidth > 0;
    }),
  );
  await page.mouse.move(700, 420);
  await dock.waitFor({ state: "hidden" });
  await page.screenshot({
    path: artifacts + "/reader-layout-desktop.png",
    animations: "disabled",
  });
  await page.mouse.move(1439, 270);
  await dock.waitFor();
  await page.getByRole("button", { name: /^已?点赞|^点赞/ }).hover();
  const rect = await dock.boundingBox(),
    main = await body.boundingBox();
  assert(rect.x >= main.x + main.width);
  await page.screenshot({
    path: artifacts + "/reader-dock-hover.png",
    animations: "disabled",
  });
  await page.mouse.move(700, 420);
  await dock.waitFor({ state: "hidden" });
  await trigger.focus();
  await page.keyboard.press("Enter");
  await page.keyboard.press("Tab");
  assert.match(await page.locator(":focus").textContent(), /点赞/);
  await page.keyboard.press("Tab");
  assert.match(await page.locator(":focus").textContent(), /收藏/);
  await page.keyboard.press("Enter");
  const dialog = page.getByRole("dialog", { name: "收藏到文件夹" });
  await dialog.waitFor();
  await page.mouse.move(370, 280);
  assert(await dialog.isVisible());
  await page.screenshot({
    path: artifacts + "/reader-folder-dialog.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await dialog.waitFor({ state: "hidden" });
  await page.keyboard.press("Escape");
  await dock.waitFor({ state: "hidden" });
  assert(await trigger.evaluate((el) => el === document.activeElement));
  await page
    .getByRole("navigation", { name: "文章目录" })
    .getByText("路线记录")
    .click();
  await page.waitForFunction(
    () => document.activeElement.id === "heading-路线记录",
  );
  assert.equal(
    await body.locator("pre").evaluate((el) => el.scrollWidth > el.clientWidth),
    true,
  );
  await page.screenshot({
    path: artifacts + "/reader-table-code.png",
    animations: "disabled",
  });
  const touch = await browser.newContext({
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true,
    reducedMotion: "reduce",
  });
  try {
    await touch.route("**/v1/**", apiFixture);
    await touch.route("**" + picture, photo);
    const phone = await touch.newPage();
    await phone.goto(origin + "/article/3", { waitUntil: "networkidle" });
    const phoneBody = phone.locator(".markdown-body");
    assert.equal(
      await phoneBody.locator("figure figcaption").first().textContent(),
      "林场以东，天刚放亮。",
    );
    assert.equal(
      await phoneBody.getByText("回到林间").getAttribute("href"),
      "#heading-林间的第一束光",
    );
    await phone.locator(".article-outline summary").tap();
    await phone
      .getByRole("navigation", { name: "文章目录" })
      .getByText("路线记录")
      .tap();
    const table = phoneBody.locator(".article-table");
    assert(await table.evaluate((el) => el.scrollWidth > el.clientWidth));
    assert(
      await phoneBody
        .locator("pre")
        .evaluate((el) => el.scrollWidth > el.clientWidth),
    );
    assert.equal(
      await phone.evaluate(
        () => document.documentElement.scrollWidth > innerWidth,
      ),
      false,
    );
    const mobileDock = phone.locator("#reader-action-panel");
    await phone.getByRole("button", { name: "文章互动工具" }).tap();
    await mobileDock.waitFor();
    await phone.screenshot({
      path: artifacts + "/reader-mobile-touch.png",
      animations: "disabled",
    });
    await mobileDock.getByRole("button", { name: /收藏/ }).tap();
    await phone.getByRole("dialog").waitFor();
    await phone.getByRole("button", { name: "关闭", exact: true }).tap();
    await phone.locator(".markdown-body h2").last().tap();
    await mobileDock.waitFor({ state: "hidden" });
    await phone.evaluate(() => window.scrollTo(0, 0));
    await phone.screenshot({
      path: artifacts + "/reader-layout-mobile.png",
      animations: "disabled",
    });
  } finally {
    await touch.close();
  }
}
