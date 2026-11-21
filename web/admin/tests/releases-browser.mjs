import assert from "node:assert/strict";
import { once } from "node:events";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import { createFrontendServer } from "../server/server.mjs";

const server = createFrontendServer({
  dist: fileURLToPath(new URL("../dist/", import.meta.url)),
});
server.listen(0, "127.0.0.1");
await once(server, "listening");
const browser = await chromium.launch({
  channel: process.env.PLAYWRIGHT_CHANNEL || "chrome",
});
try {
  const page = await browser.newPage({
    viewport: { width: 1440, height: 900 },
  });
  page.setDefaultTimeout(8000);
  await page.context().route("**/packages", route => route.fulfill({ contentType: "text/html", body: "下载中心" }));
  await page.addInitScript(() =>
    sessionStorage.setItem("token", "test-release"),
  );
  const groups = [],
    files = [
      {
        id: "linux",
        revision: "1",
        product: "安装工具",
        version: "v2",
        platform: "Linux",
        status: "draft",
        url: "https://files.example.test/cli.zip",
        sortOrder: 0,
      },
    ];
  let settings = {
    title: "下载中心",
    revision: "1",
    requirements: [],
    installOptions: [],
  };
  const writes = [],
    errors = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.route("**/v1/**", async (route) => {
    const req = route.request(),
      path = new URL(req.url()).pathname,
      action = path.split("/admin/releases/")[1];
    const post = req.method() === "POST",
      payload = post ? req.postDataJSON() : null;
    let data = { items: [] };
    if (path.endsWith("/menus/public"))
      data = { items: [{ name: "下载", path: "/packages", component: "download/index" }] };
    if (path.endsWith("/auth/me")) data = { id: "1", roleIds: ["1"] };
    if (path.endsWith("/auth/menus"))
      data = {
        items: [
          {
            id: "1",
            menu_name: "下载管理",
            path: "/releases",
            component: "product/downloads/index",
            status: 1,
            menu_type: 2,
            is_visible: true,
          },
        ],
      };
    if (path.endsWith("/auth/permissions")) data = { items: ["*"] };
    if (path.endsWith("/auth/totp")) data = { enabled: false };
    if (action === "groups") data = { items: groups };
    if (action === "files") data = { items: files };
    if (action === "settings") data = settings;
    if (post && action) {
      writes.push({ action, payload });
      if (action === "groups/create") {
        data = { ...payload, revision: "1" };
        groups.push(data);
      }
      if (action === "settings/update")
        settings = data = { ...payload, revision: "2" };
      if (
        action === "files/update" ||
        action === "files/publish" ||
        action === "files/withdraw"
      ) {
        assert.equal(payload.revision, Number(files[0].revision));
        files[0] = {
          ...files[0],
          ...payload,
          revision: String(payload.revision + 1),
        };
        if (action === "files/publish") files[0].status = "published";
        if (action === "files/withdraw") files[0].status = "archived";
        data = files[0];
      }
    }
    await route.fulfill({ json: { code: 0, message: "ok", data } });
  });
  await page.goto("http://127.0.0.1:" + server.address().port + "/releases");
  await page.getByRole("main").getByText("安装工具", { exact: true }).waitFor();
  if (process.env.SITEHUB_PREVIEW_DIR)
    await page.screenshot({ path: process.env.SITEHUB_PREVIEW_DIR + "/release-workspace.png" });
  const opened = page.waitForEvent("popup");
  await page.getByRole("button", { name: "下载页预览" }).click();
  const preview = await opened;
  await preview.waitForURL(url => url.pathname === "/packages");
  await preview.close();
  const statusFilters = page.getByRole("group", { name: "按上架状态筛选" });
  await statusFilters.getByRole("button", { name: "已发布 0" }).click();
  await page.getByText("暂无匹配的下载文件").waitFor();
  await statusFilters.getByRole("button", { name: "草稿 1" }).click();
  await page.getByRole("row").filter({ hasText: "安装工具" }).waitFor();
  await statusFilters.getByRole("button", { name: "全部 1" }).click();
  await page.getByRole("button", { name: "管理分组" }).click();
  await page.getByRole("button", { name: "新增分组" }).click();
  await page.getByLabel("分组名称").fill("工具");
  await page.getByRole("button", { name: "保存分组" }).click();
  await page.getByRole("dialog").getByText("工具", { exact: true }).waitFor();
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "下载页设置" }).click();
  await page.getByPlaceholder("下载中心").fill("下载与工具");
  await page.getByRole("button", { name: "保存设置" }).click();
  await page.getByText("下载页设置已保存", { exact: true }).waitFor();
  const row = page.getByRole("row").filter({ hasText: "安装工具" });
  await row.getByRole("button", { name: "编辑" }).click();
  await page.getByPlaceholder("v1.0.0").fill("v2.1");
  await page.getByRole("button", { name: "保存文件" }).click();
  await row.getByText("v2.1", { exact: true }).waitFor();
  await row.getByRole("button", { name: /发布|上架/ }).click();
  await page
    .locator(".el-message-box")
    .getByRole("button", { name: /确定|确认/ })
    .click();
  await row.getByText("已发布", { exact: true }).waitFor();
  await row.getByRole("button", { name: /下架|撤回/ }).click();
  await page
    .locator(".el-message-box")
    .getByRole("button", { name: /确定|确认/ })
    .click();
  await row.getByText("已下架", { exact: true }).waitFor();
  assert.deepEqual(errors, []);
  assert.deepEqual(
    writes.map((w) => w.action),
    [
      "groups/create",
      "settings/update",
      "files/update",
      "files/publish",
      "files/withdraw",
    ],
  );
  console.log(
    "PASS release groups, page settings, file editing, publication and withdrawal",
  );
} finally {
  await browser.close();
  server.closeAllConnections();
  await new Promise((resolve) => server.close(resolve));
}
