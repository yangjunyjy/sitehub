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
  async function expireSession() {
    await page.context().clearCookies({ name: "sitehub_gateway_session" });
    await button("审阅并发布").click();
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
  const home = page.getByRole("row").filter({ hasText: "app.test" });
  await home.getByRole("button", { name: "编辑" }).click();
  await button("保存到草稿").click();
  assert.equal(await button("审阅并发布").count(), 0);

  // Exercise a lagging instance without making the gateway fixture depend on
  // a failed deployment or real upstream sockets.
  const statePath = "**/api/workspace?history=0";
  await page.route(statePath, async (route) => {
    const response = await route.fetch();
    const json = await response.json();
    const state = json.state;
    state.applied_manifest = structuredClone(state.manifest);
    state.applied_manifest.routes[0].host = "live-instance.test";
    state.applied_revision = state.revision - 1;
    state.last_error = "无法监听测试端口";
    state.runtime = {
      http_listen: ":8082",
      path_warnings: ["精确路径 / 优先"],
      streams: [
        {
          id: "tcp-live",
          protocol: "tcp",
          listen: ":0",
          address: "127.0.0.1:12000",
          target: "127.0.0.1:13000",
        },
      ],
      upstreams: {
        product: [{ url: "http://127.0.0.1:13000", weight: 3, healthy: false }],
      },
    };
    await route.fulfill({ response, json });
  });
  await button("重新载入").click();
  await page.getByText("无法监听测试端口").waitFor();
  await label("查看配置").selectOption({ label: "本机生效配置（只读）" });
  await page.getByText("live-instance.test", { exact: true }).waitFor();
  assert.equal(await button("编辑").count(), 0);
  await label("查看配置").selectOption({ label: "期望配置与草稿" });
  await home.waitFor();
  await page.getByText("路径匹配提示（1）").click();
  await page.getByText("精确路径 / 优先").waitFor();
  await page.getByText("试匹配 HTTP 请求", { exact: true }).click();
  await page.route("**/api/match", async (route) => {
    const input = route.request().postDataJSON();
    assert.equal(input.host, "APP.test:2357");
    assert.equal(input.method, "PROPFIND");
    assert.equal(input.scheme, "http");
    const matched = input.path === "/";
    await route.fulfill({
      json: matched
        ? {
            revision: 6,
            matched,
            route_id: "home",
            priority: 0,
            pattern: "/*",
            terminal: "static-response",
          }
        : { revision: 6, matched, reason: "主机和方法已匹配，但路径没有命中" },
    });
  });
  await label("协议").selectOption("http");
  await label("试匹配方法").fill("propfind");
  await label("请求 Host").fill("APP.test:2357");
  await button("试匹配").click();
  await page.getByText("命中 home", { exact: true }).waitFor();
  await label("试匹配路径").fill("/absent");
  await button("试匹配").click();
  await page.getByText("主机和方法已匹配，但路径没有命中").waitFor();
  await gotoSection("上游服务");
  await page.getByText("http://127.0.0.1:13000 · 不可用 · 权重 3").waitFor();
  await gotoSection("四层转发");
  await page.getByText("TCP · 127.0.0.1:12000 → 127.0.0.1:13000").waitFor();
  await page.unroute(statePath);
  await button("重新载入").click();
  await page.getByText("无法监听测试端口").waitFor({ state: "detached" });
  await gotoSection("上游服务");
  await button("新增上游服务").click();
  await label("上游标识").fill("catalog");
  await label("服务发现名称").fill("catalog-api");
  await button("保存到草稿").click();
  await page.getByText("catalog-api", { exact: true }).waitFor();
  await expireSession();
  await login("operator", "fixture-password");
  await page.getByText("已恢复当前账号的草稿。").waitFor();
  await page.getByText("catalog-api", { exact: true }).waitFor();
  await page.route(statePath, async (route) => {
    const response = await route.fetch();
    const json = await response.json();
    json.publish = json.rollback = false;
    Object.assign(json.state, {
      desired_unavailable: true,
      revision: 0,
      applied_manifest: json.state.manifest,
      last_error: "无法读取期望配置",
    });
    await route.fulfill({ response, json });
  });
  await button("重新载入").click();
  await button("合并最新配置").click();
  await page.getByText("期望配置暂不可用 · 本机生效修订 7（只读）").waitFor();
  assert.equal(await button("新增上游服务").count(), 0);
  assert.equal(await page.getByText("catalog-api", { exact: true }).count(), 0);
  await page.unroute(statePath);
  await button("重新载入").click();
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
  await page.getByText("关联路由（1）", { exact: true }).click();
  await page
    .getByRole("dialog")
    .getByText("catalog-route", { exact: true })
    .last()
    .waitFor();
  await page.getByRole("columnheader", { name: "目标值" }).first().waitFor();
  await shot("publish-review");
  await label("确认网关账号密码").fill("fixture-password");
  await button("确认发布").click();
  await page.getByText("配置修订 8 · 已应用 8").waitFor();
  await gotoSection("发布记录");
  await button("回滚").click();
  await page.getByText("删除上游服务 · catalog", { exact: true }).waitFor();
  await page.getByRole("columnheader", { name: "当前值" }).first().waitFor();
  await label("确认网关账号密码").fill("fixture-password");
  await button("确认回滚").click();
  await page.getByText("配置修订 9 · 已应用 9").waitFor();
  await page.route("**/api/history?*", async (route) => {
    const index = Number(
      new URL(route.request().url()).searchParams.get("page"),
    );
    const entry = {
      generation: "release-" + index,
      actor: "operator",
      previous_revision: 7,
      created_at: new Date().toISOString(),
      status: "rejected",
      message: "拒绝原因 " + index,
    };
    await route.fulfill({
      json: { items: [entry], total: 16, page: index, limit: 15 },
    });
  });
  await button("重新载入").click();
  await page.getByText("拒绝原因 1").waitFor();
  await button("下一页").click();
  await page.getByText("拒绝原因 2").waitFor();
  await button("查看").click();
  await page
    .getByRole("dialog")
    .getByText("拒绝原因 2", { exact: true })
    .waitFor();
  await button("关闭").click();
  await page.unroute("**/api/history?*");
  await gotoSection("四层转发");
  await page.getByText("暂无四层转发").waitFor();
  await gotoSection("HTTP 路由");
  await home.getByRole("button", { name: "编辑" }).click();
  await label("优先级").fill("11");
  await button("保存到草稿").click();
  // A second writer changes the same resource after this tab starts editing.
  const { state } = await (
    await page.request.get(origin + "api/workspace?history=0")
  ).json();
  state.manifest.routes[0].priority = 5;
  const competing = await page.request.post(origin + "api/publish", {
    headers: { Origin: new URL(origin).origin },
    data: {
      revision: state.revision,
      manifest: state.manifest,
      password: "fixture-password",
    },
  });
  assert.equal(competing.status(), 200, await competing.text());
  await button("审阅并发布").click();
  await page.getByRole("heading", { name: "处理并发修改" }).waitFor();
  await page.getByRole("row", { name: "priority 5 11" }).waitFor();
  await label("采用的版本").selectOption("local");
  await button("应用选择").click();
  await button("审阅并发布").click();
  await page
    .getByRole("dialog")
    .getByRole("row", { name: "priority 5 11" })
    .waitFor();
  await button("取消").click();
  await expireSession();
  await login("deployer", "fixture-password");
  await home.getByText("优先级 5", { exact: true }).waitFor();
  assert.equal(await button("审阅并发布").count(), 0);
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
