import assert from "node:assert/strict";

export async function checkJournalLayout({ page, origin, artifacts }) {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto(origin, { waitUntil: "networkidle" });
  const rows = page.locator("#articles .article-row");
  const pager = page.getByRole("navigation", { name: "文章分页" });
  assert.equal(await rows.count(), 8);
  assert(
    !(await rows.first().evaluate((row) => row.classList.contains("arrived"))),
  );
  await page.locator("#articles").scrollIntoViewIfNeeded();
  await page.waitForFunction(() =>
    document.querySelector("#articles .article-row.arrived"),
  );
  assert.equal(await rows.locator(".thumbnail").count(), 0);
  const heights = await rows.evaluateAll((elements) =>
    elements.map((row) => Math.round(row.offsetHeight)),
  );
  assert(
    Math.max(...heights) <= 115,
    "Feed rows exceed compact height: " + heights,
  );
  await page.screenshot({
    path: artifacts + "/feed-desktop.png",
    animations: "disabled",
  });
  await pager.getByRole("button", { name: "下一页" }).click();
  await rows.first().getByText("部署手记").waitFor();
  assert.match(page.url(), /page=2/);
  await page.reload({ waitUntil: "networkidle" });
  assert.equal(await rows.count(), 2);
  assert(await pager.getByRole("button", { name: "下一页" }).isDisabled());
  await page.goBack();
  await rows.first().getByText("deploy 工具简介").waitFor();
  await page.goto(origin + "/categories?category=1&page=20", {
    waitUntil: "networkidle",
  });
  await pager.getByText("第 1 / 1 页").waitFor();
  assert(!new URL(page.url()).searchParams.has("page"));
  await page.unroute("**/v1/blog/about");
  await page.goto(origin + "/about", { waitUntil: "networkidle" });
  await page.getByRole("heading", { name: "关于云间小记" }).waitFor();
  await page.screenshot({
    path: artifacts + "/about-desktop.png",
    animations: "disabled",
  });
  const account = page.getByRole("button", { name: "账号菜单" });
  await account.press("ArrowDown");
  const profileLink = page
    .locator("#member-options")
    .getByRole("link", { name: "个人中心" });
  assert(await profileLink.evaluate((el) => el === document.activeElement));
  await profileLink.press("Escape");
  assert.equal(await account.getAttribute("aria-expanded"), "false");
  await account.click();
  await page.getByRole("heading", { name: "关于云间小记" }).click();
  assert.equal(await account.getAttribute("aria-expanded"), "false");
  await page.setViewportSize({ width: 390, height: 844 });
  await account.click();
  await profileLink.click();
  await page.waitForURL("**/library");
  await account.click();
  await page.screenshot({
    path: artifacts + "/account-menu-mobile.png",
    animations: "disabled",
  });
  await page.getByRole("button", { name: "退出登录" }).click();
  await page.waitForURL("**/signin?**");
  await page.getByRole("link", { name: "会员登录" }).waitFor();
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto(origin + "/?category=1#articles", {
    waitUntil: "networkidle",
  });
  await pager.getByText("第 1 / 1 页").waitFor();
  assert(
    await rows
      .first()
      .evaluate(
        (row) =>
          getComputedStyle(row).opacity === "1" &&
          getComputedStyle(row).animationName === "none",
      ),
  );
  assert.equal(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
    false,
  );
  await page.screenshot({
    path: artifacts + "/feed-mobile.png",
    animations: "disabled",
  });
  console.log("Measured desktop feed row heights:", heights);
}
