import assert from "node:assert/strict";
import { chromium } from "playwright";

const origin = process.env.GATEWAY_CONSOLE_TEST_URL;
if (!origin)
  throw Error("Run TestConsoleBrowser with GATEWAY_CONSOLE_BROWSER_TEST=1");
const browser = await chromium.launch({
  channel: process.env.PLAYWRIGHT_CHANNEL || undefined,
});
try {
  const page = await browser.newPage();
  page.setDefaultTimeout(8000);
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  const nav = page.getByRole("navigation", { name: "主导航" });
  const gotoSection = (name) =>
    nav.getByRole("link", { name, exact: true }).click();
  const button = (name) => page.getByRole("button", { name, exact: true });
  const label = (name) => page.getByLabel(name, { exact: true });
  const shot = async (name) => {
    if (process.env.SITEHUB_PREVIEW_DIR)
      await page.screenshot({
        path: process.env.SITEHUB_PREVIEW_DIR + "/gateway-" + name + ".png",
        animations: "disabled",
      });
  };
  async function login(username, password) {
    await label("用户名").fill(username);
    await label("密码").fill(password);
    await button("登录").click();
  }
  async function logout() {
    await button("退出登录").click();
    await button("确认退出").click();
    await page.getByRole("heading", { name: "登录网关控制台" }).waitFor();
  }

  await page
    .context()
    .addCookies([
      { name: "sitehub_admin_session", value: "unrelated", url: origin },
    ]);
  await page.goto(origin);
  await page.getByRole("heading", { name: "登录网关控制台" }).waitFor();
  await shot("login");
  await login("operator", "wrong-password");
  await page.getByRole("alert").waitFor();
  await login("operator", "fixture-password");
  await nav.waitFor();
  await shot("workspace");
  assert(
    (await page.context().cookies()).find(
      (c) => c.name === "sitehub_gateway_session",
    )?.httpOnly,
  );
  await page.reload();
  await gotoSection("上游服务");
  await button("新增上游服务").click();
  await label("上游标识").fill("catalog");
  await label("服务发现名称").fill("catalog-api");
  await button("保存到草稿").click();
  await page.getByText("catalog-api", { exact: true }).waitFor();

  await gotoSection("HTTP 路由");
  await button("新增HTTP 路由").click();
  await label("路由标识").fill("catalog-route");
  await label("请求路径").fill("/v1/*");
  await label("目标上游").selectOption("catalog");
  await button("保存到草稿").click();
  const route = page.getByRole("row").filter({ hasText: "catalog-route" });
  await route.getByText("全部主机", { exact: true }).waitFor();
  await route.getByRole("button", { name: "编辑" }).click();
  await label("匹配主机").fill("127.0.0.1:9080");
  await button("保存到草稿").click();
  await route.getByText("127.0.0.1:9080", { exact: true }).waitFor();
  await gotoSection("四层转发");
  for (const protocol of ["TCP", "UDP"]) {
    await button("新增四层转发").click();
    await label("规则标识").fill("port-" + protocol.toLowerCase());
    await label("传输协议").selectOption(protocol);
    await label("监听地址").fill("127.0.0.1:15353");
    await label("目标地址").fill("127.0.0.1:5353");
    await button("保存到草稿").click();
  }
  await label("搜索四层转发").fill("udp");
  assert.equal(
    await page.getByRole("cell", { name: "TCP", exact: true }).count(),
    0,
  );
  await label("搜索四层转发").fill("");
  await page.getByRole("cell", { name: "TCP", exact: true }).waitFor();
  await shot("streams");
  await button("审阅并发布").click();
  for (const title of [
    "新增HTTP 路由 · catalog-route",
    "新增上游服务 · catalog",
    "新增四层转发 · port-tcp",
    "新增四层转发 · port-udp",
  ])
    await page.getByText(title, { exact: true }).waitFor();
  await shot("publish-review");
  await label("确认网关账号密码").fill("fixture-password");
  await button("确认发布").click();
  await page.getByText("配置修订 8 · 已应用 8").waitFor();
  await gotoSection("发布记录");
  await button("回滚").click();
  await label("确认网关账号密码").fill("fixture-password");
  await button("确认回滚").click();
  await page.getByText("配置修订 9 · 已应用 9").waitFor();
  await gotoSection("四层转发");
  await page.getByText("暂无四层转发").waitFor();
  await page.setViewportSize({ width: 390, height: 844 });
  await button("切换导航").click();
  await gotoSection("插件策略");
  assert(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  );
  await logout();
  const cookies = await page.context().cookies();
  assert(cookies.some((c) => c.name === "sitehub_admin_session"));
  assert(!cookies.some((c) => c.name === "sitehub_gateway_session"));
  assert.equal(
    (await page.request.get(origin + "api/workspace")).status(),
    401,
  );
  await login("reader", "fixture-password");
  await page.getByText("只读账号 · 可查看规则和发布历史。").waitFor();
  assert.equal(await button("新增HTTP 路由").count(), 0);
  await logout();
  assert.deepEqual(errors, []);
} finally {
  await browser.close();
}
