import assert from "node:assert/strict";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import { checkJournalLayout } from "./journal-layout.mjs";
import { readerArticle, checkReaderLayout } from "./reader-layout.mjs";
import { createFrontendServer } from "../server/server.mjs";

// Fixtures isolate full reader workflows from real member accounts and published content.
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
let signedIn = false;
let registeredEmail = "";
let customPaths = false;
let codeAttempts = 0;
let liked = false;
let savedIn = null;
const reader = {id:"99", email:"reader@example.test"};
const folders = [{id:"0", name:"默认收藏夹"}];
const interaction = () => ({liked, likes:Number(liked), favorited:savedIn!==null,
  favorites:Number(savedIn!==null), folderId:savedIn ?? "0"});
const apiFixture = async (route) => {
  const path = new URL(route.request().url()).pathname;
  if (path.endsWith("/members/me") && !signedIn)
    return route.fulfill({
      status: 401,
      json: { code: 401, message: "请先登录" },
    });
  let data = { items: [], total: 0 };
  if (path.endsWith("/menus/public"))
    data.items = [
      { name: '首页', component: 'home/index', path: '/' },
      { name: '内容', path: '/explore', children: [
        {name: '分类', component: 'categories/index', path: '/categories'},
        {name: '标签', component: 'tags/index', path: '/tags'},
        {name: '归档', component: 'archives/index', path: '/archives'}]},
      { name: '友链', component: 'friends/index', path: '/friends' },
      { name: '关于', component: 'about/index', path: '/about' },
      { name: '留言', component: 'messages/index', path: '/messages' },
    ];
  if (path.endsWith("/articles")) {
    const params = new URL(route.request().url()).searchParams;
    const pageSize = Number(params.get("pageSize")), page = Number(params.get("page"));
    const titles = ["deploy 工具简介", "Go 模块的版本管理", "博客启程", "秋天的山路", "终端里的日常", "一次慢查询的排查", "写在雨过天晴之后", "为小站搭一间书房", "部署手记", "在命令行里听雨"];
    const records = titles.map((title, index) => ({
      id: String(index + 1), title, publishedAt: "2026-11-" + (21 - index), views: 12 - index,
      isTop: index === 0, category: {id:"1", name:index % 2 ? "随笔" : "开发手记"},
      tags: index % 2 ? [] : [{id:"2", name:"golang"}],
      summary: index === 1 ? "从一个 go.mod 文件开始，整理项目的依赖和版本。" : "",
    }));
    const matching = params.has("categoryID") ? records.slice(0, 3) : records;
    data = { items: matching.slice((page - 1) * pageSize, page * pageSize), total: matching.length };
  }
  if (path.endsWith("/members/logout") && route.request().method() === "POST") signedIn = false;
  if (path.endsWith("/about")) data = { name: '云间小记', bio: '这里放着开发笔记，也放着周末沿河散步时记下的小事。', skills: ['Go', '运维', '摄影'], location: '杭州', homeVerse: '在文字的星河里，记录生活的光与热', homeAnimation: true };
  if (path.endsWith('/friends')) data.items = [
    ['夏末的猫', '记录开发日常和窗边的小事。'], ['工具小记', '分享每天用到的小工具。'],
    ['空白之地', '在代码和文字之间寻找平衡。'], ['山间木屋', '写点技术，也记四季的风景。'],
    ['阿树的笔记', '学习笔记和项目实践。'], ['晚风小栈', '关于读书和散步的随笔。'],
    ['猫鸣的博客', '前端、设计和生活碎片。'], ['旅人手记', '旅行途中的山河与见闻。'],
  ].map(([name, description], index) => ({ id: String(index + 1), name, description, url: `https://example.org/friend/${index + 1}` }));
  if (path.endsWith('/menus/public')) data.items.push({ name: '站外手记', pageType: 'external', externalUrl: 'https://example.org/notes' });
  if (path.endsWith('/messages') && !new URL(route.request().url()).searchParams.has('articleId')) {
    const second = new URL(route.request().url()).searchParams.get('page') === '2';
    data = { total: 11, items: [{ id: second ? '1' : '11', nickname: '星旅客', content: second ? '前一页的星光' : '新到的星光' }] };
  }
  if (path.endsWith("/articles/detail"))
    data = {
      id: "1",
      title: "Isolated article fixture",
      content_format: "markdown",
      content:
        "# Reader heading\n\n" + 'A paragraph.\n\n'.repeat(65) + "<script>window.untrustedArticle=true</script>",
      tags: [],
    };
  if (path.endsWith("/articles/detail") && new URL(route.request().url()).searchParams.get("id") === "2")
    data=readerArticle("markdown");
  if (path.endsWith("/articles/detail") && new URL(route.request().url()).searchParams.get("id") === "3")
    data=readerArticle("html");
  const post = route.request().method() === "POST";
  const values = post ? route.request().postDataJSON() : {};
  if(path.endsWith("/members/me")) data=reader;
  if(path.endsWith("/members/code")) {
    if (++codeAttempts === 1) return route.fulfill({status:503,json:{code:503}});
    data={};
  }
  if (path.endsWith("/menus/public") && customPaths) {
    for (const [component,path] of [["login","/signin"],["register","/join"],["account","/library"],["article","/read"]])
      data.items.push({name:component,path,component:component+"/index"});
  }
  if(path.endsWith("/members/register")) registeredEmail=values.email;
  if(path.endsWith("/members/login")){
    assert.equal(registeredEmail,values.email);
    signedIn=true;
    data={user:reader};
  }
  if(path.endsWith("/articles/like")) {assert(signedIn); liked=values.liked;}
  if(path.endsWith("/articles/favorite")) {assert(signedIn); savedIn=values.favorited?String(values.folderId):null;}
  if(/\/articles\/(interaction|favorite|like)$/.test(path)) data=interaction();
  if(path.endsWith("/favorites/folders")) {
    assert(signedIn);
    if(post) {
      data={id:String(folders.length),name:values.name,count:0};
      folders.push(data);
    } else data={items:folders.map(folder=>({...folder,count:Number(savedIn===folder.id)}))};
  }
  if(path.endsWith("/favorites/articles")){
    assert(signedIn);
    const selected=new URL(route.request().url()).searchParams.get("folderId");
    const items=selected===savedIn?[{id:"1",title:"Isolated article fixture",savedAt:"2026-01-01T12:00:00Z"}]:[];
    data={items,total:items.length};
  }
  if(path.endsWith("/messages") && !post && new URL(route.request().url()).searchParams.has("articleId"))
    data={items:[{id:"24",nickname:"晴川",content:"谢谢分享👏"}],total:1};
  if(post && (path.endsWith("/messages") || path.endsWith("/friends"))) {
    assert(signedIn);
    data={id:"1",status:"pending"};
  }
  return route.fulfill({ json: { code: 0, message: "ok", data } });
};
await context.route("**/v1/**",apiFixture);
const page = await context.newPage();
page.on("pageerror", (error) => errors.push(error.message));
async function openReaderDock(page) {
  const trigger=page.getByRole("button",{name:"文章互动工具"});
  await trigger.waitFor();
  const width=await page.evaluate(()=>document.documentElement.clientWidth);
  await page.mouse.move(width-2, Math.round(page.viewportSize().height*.5));
  await page.locator("#reader-action-panel").waitFor();
}
async function scrollHeaderTo(position, shown) {
  await page.evaluate((top) => window.scrollTo(0, top), position);
  await page.waitForFunction(([top, visible]) => {
    const rect = document.querySelector(".blog-header").getBoundingClientRect();
    return window.scrollY === top && (visible ? Math.abs(rect.top) < .5 : rect.bottom <= 0);
  }, [position, shown]);
}

async function checkHomeMotion() {
  const verse = page.getByRole("heading", { name: "在文字的星河里，记录生活的光与热", exact: true });
  assert.equal(await verse.count(), 1);
  assert.equal(await page.getByRole("heading", { name: "首页", exact: true }).count(), 0);
  await page.emulateMedia({ reducedMotion: "reduce" });
  const shower = page.locator(".meteor-shower");
  await shower.waitFor({ state: "hidden" });
  await waitForFullVerse();
  assert.equal(await page.locator('.motion-control').count(), 0);
  await page.emulateMedia({ reducedMotion: "no-preference" });
  await checkMeteorFlight();
  await waitForFullVerse();
  await page.waitForFunction(() => !document.querySelector(".verse-ink").textContent);
  await page.waitForFunction(() => document.querySelector(".verse-ink").textContent.length >= 2);
  await scrollHeaderTo(42, true);
  await scrollHeaderTo(240, false);
  await scrollHeaderTo(235, false);
  await scrollHeaderTo(210, true);
  assert.equal(await page.locator('.blog-header.is-solid').count(), 0);
  await scrollHeaderTo(0, true);
}
async function waitForFullVerse() {
  await page.waitForFunction(() => {
    const verse = document.querySelector(".verse-ink");
    return verse.textContent === verse.parentElement.getAttribute("aria-label");
  });
}
async function checkMeteorFlight() {
  const showerLayer = page.locator(".meteor-shower");
  await showerLayer.waitFor({ state: 'visible' });
  const visibleFlight = await showerLayer.evaluate(async (shower) => {
    function sample() {
      return Array.from(shower.children, meteor => {
        const { left, top } = meteor.getBoundingClientRect();
        const visible = left > 0 && left < innerWidth && top > 60 && top < innerHeight && getComputedStyle(meteor).opacity > .55;
        return { left, top, visible };
      });
    }
    let previous = sample();
    for (let attempt = 0; attempt < 12; attempt++) {
      await new Promise(done => setTimeout(done, 160));
      const current = sample();
      if (current.some((end, index) => previous[index].visible && end.visible && end.left < previous[index].left - 2 && end.top > previous[index].top + 1)) return true;
      previous = current;
    }
    return false;
  });
  assert(visibleFlight, "At least one clearly visible meteor should travel southwest in real time");
}
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
    "/register",
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
    if (path === "/") {
      await checkHomeMotion();
      assert.equal(await page.getByRole('navigation', { name: '主要导航', exact: true }).getByRole('link', { name: '站外手记' }).getAttribute('href'), 'https://example.org/notes');
      const content = page.getByRole("navigation", {name: "主要导航", exact: true}).getByRole("button", { name: "内容" });
      const tag = page.getByRole("navigation", { name: "主要导航", exact: true }).getByRole("link", { name: "标签", exact: true });
      await content.hover();
      await tag.waitFor();
      await tag.hover();
      await page.screenshot({ path: `${artifacts}/home-content-menu.png` });
      await page.mouse.move(10, 260);
      await tag.waitFor({ state: "hidden" });
      await content.click();
      await tag.waitFor();
      await scrollHeaderTo(200, true);
      await content.press("Escape");
      assert.equal(await content.getAttribute("aria-expanded"), "false");
      await content.press("Enter");
      await tag.waitFor();
      await content.press("Escape");
      await scrollHeaderTo(0, true);
      await page.locator(".home-hero").click({ position: { x: 10, y: 90 } });
      await waitForFullVerse();
    }
    if (["/", "/messages", "/friends"].includes(path))
      await page.screenshot({
        path: `${artifacts}/${path === "/" ? "home" : path.slice(1)}-desktop.png`,
        fullPage: true,
      });
  }
  await page.goto(origin + "/friends#apply", { waitUntil: "networkidle" });
  await page.getByRole("link", {name:/登录后申请友链/}).waitFor();
  assert.equal(await page.getByPlaceholder("你的站点叫什么名字？").count(), 0);
  await page.goto(origin + "/messages", { waitUntil: "networkidle" });
  await page.getByRole("link", {name:/登录后写留言/}).waitFor();
  await page.goto(origin + "/article/1", { waitUntil: "networkidle" });
  await page.getByRole("link", {name:/登录后写评论/}).waitFor();
  await page.getByRole("img", {name:"👏"}).waitFor();
  await openReaderDock(page);
  await page.getByRole("button", {name:/^收藏/}).click();
  await page.getByRole("heading", {name:"登录读者账号"}).waitFor();
  await page.getByRole("link", {name:/注册账号/}).click();
  await page.locator('input[type="email"]').fill(reader.email);
  await page.getByRole("button", {name:"获取验证码"}).click();
  await page.getByRole("alert").getByText("服务暂时不可用，请稍后重试").waitFor();
  assert.equal(await page.locator('input[type="email"]').inputValue(),reader.email);
  await page.getByRole("button", {name:"获取验证码"}).click();
  await page.getByRole("status").getByText("验证码已发送，请查看邮箱。").waitFor();
  await page.getByPlaceholder("6 位验证码").fill("123456");
  await page.locator('[autocomplete="new-password"]').fill("Reader123!");
  await page.screenshot({path:artifacts+"/register-desktop.png"});
  await page.getByRole("button", {name:"注册并登录"}).click();
  await page.waitForURL("**/article/1");
  await openReaderDock(page);
  await page.getByRole("button", {name:/^点赞/}).click();
  await page.getByRole("button", {name:"已点赞 1"}).waitFor();
  await openReaderDock(page);
  await page.getByRole("button", {name:/^收藏/}).click();
  const dialog=page.getByRole("dialog");
  assert(await dialog.getByRole("radio", {name:/默认收藏夹/}).isChecked());
  await dialog.getByRole("button", {name:"保存收藏"}).click();
  await dialog.waitFor({state:"hidden"});
  await openReaderDock(page);
  await page.getByRole("button", {name:/^已收藏/}).click();
  await dialog.getByRole("button", {name:/新建收藏夹/}).click();
  await dialog.getByLabel("新收藏夹名称").fill("技术笔记");
  await dialog.getByRole("button", {name:"创建收藏夹"}).click();
  await dialog.getByRole("radio", {name:/技术笔记/}).waitFor();
  assert(await dialog.getByRole("radio", {name:/技术笔记/}).isChecked());
  await dialog.getByRole("button", {name:"保存收藏"}).click();
  await dialog.waitFor({state:"hidden"});
  assert.equal(savedIn,"1");
  await page.goto(origin+"/account", {waitUntil:"networkidle"});
  await page.getByText("这个收藏夹还没有文章").waitFor();
  await page.getByRole("button", {name:/技术笔记/}).click();
  await page.getByRole("link", {name:/Isolated article fixture/}).waitFor();
  await page.screenshot({path:artifacts+"/account-desktop.png"});
  await page.getByRole("link", {name:/Isolated article fixture/}).click();
  await openReaderDock(page);
  await page.getByRole("button", {name:/^已收藏/}).click();
  await dialog.getByRole("button", {name:"取消收藏"}).click();
  await dialog.waitFor({state:"hidden"});
  assert.equal(savedIn,null);
  await page.goto(origin + "/friends", { waitUntil: "networkidle" });
  await page.getByRole('button', { name: '列表', exact: true }).click();
  await page.locator('.friend-list').getByText('夏末的猫').waitFor();
  await page.reload({ waitUntil: 'networkidle' });
  assert.equal(await page.getByRole('button', { name: '列表', exact: true }).getAttribute('aria-pressed'), 'true');
  await page.getByRole('button', { name: '卡片', exact: true }).click();
  await page.getByRole('button', { name: '前往申请' }).click();
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
  assert.equal(await page.getByLabel('留言内容').count(), 0);
  assert(await page.locator('.message-universe canvas').isVisible());
  const streak = page.locator('.message-meteor').first();
  await streak.getByText('新到的星光').waitFor();
  const start = await streak.boundingBox();
  const firstFrame = await page.locator('.message-universe canvas').evaluate(canvas => canvas.toDataURL());
  await page.waitForFunction(([left, top]) => {
    const current = document.querySelector('.message-meteor').getBoundingClientRect();
    return current.left < left - 8 && current.top > top + 2;
  }, [start.x, start.y]);
  assert.notEqual(await page.locator('.message-universe canvas').evaluate(canvas => canvas.toDataURL()), firstFrame);
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await streak.waitFor({ state: 'hidden' });
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await page.getByRole('button', { name: '下一页' }).click();
  await page.getByText('前一页的星光').waitFor();
  assert.equal(await streak.locator('span').textContent(), '新到的星光');
  await page.getByRole('button', { name: '写留言' }).click();
  await page
    .getByLabel("留言内容", { exact: true })
    .fill("Isolated browser test");
  await page.getByRole("button", { name: "发送留言", exact: true }).click();
  await page
    .getByText("留言已提交，审核通过后公开显示。", { exact: true })
    .waitFor();
  await page.goto(origin + "/article/1", { waitUntil: "networkidle" });
  const progress = page.getByRole('progressbar', { name: '文章阅读进度' });
  assert.equal(await progress.getAttribute('aria-valuenow'), '0');
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await page.waitForFunction(() => document.querySelector('[role="progressbar"]').getAttribute('aria-valuenow') === '100');
  await page.getByRole('button', { name: '写评论' }).click();
  await page.getByLabel('评论内容').fill('文章下的评论');
  await page.getByRole("button", {name:"选择表情"}).click();
  const palette=page.getByRole("group",{name:"表情包"});
  assert.equal(await palette.locator("img").count(),12);
  await page.screenshot({path:artifacts+"/article-comments-desktop.png"});
  await palette.getByRole("button", {name:"鼓掌"}).click();
  assert.equal(await page.getByLabel("评论内容").inputValue(),"文章下的评论👏");
  const comment = page.waitForRequest(req => req.method() === 'POST' && req.url().endsWith('/blog/messages'));
  await page.getByRole('button', { name: '发送评论' }).click();
  assert.equal((await comment).postDataJSON().articleId, 1);
  await page.getByText('评论已提交，审核通过后公开显示。').waitFor();
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
  await page.getByRole('button', { name: '写留言' }).click();
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
      .getByText("留言已提交，审核通过后公开显示。", {
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
    "/register",
    "/account",
    "/article/1",
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
    await page.getByRole("navigation", {name: "移动导航", exact:true}).getByRole("button", {name:"内容"}).click();
    await page
      .getByRole("navigation", { name: "移动导航", exact: true })
      .getByText("分类", { exact: true })
      .waitFor();
    if (path === "/") {
      await scrollHeaderTo(240, true);
      await page.setViewportSize({ width: 844, height: 390 });
      await page.waitForFunction(() => !document.querySelector("#mobile-navigation"));
      await scrollHeaderTo(360, false);
      await page.setViewportSize({ width: 390, height: 844 });
      await scrollHeaderTo(0, true);
      await checkMeteorFlight();
      await waitForFullVerse();
    } else {
      await page.getByRole("button", { name: "切换导航", exact: true }).click();
    }
    if (["/register","/article/1","/account"].includes(path)) await page.screenshot({path:artifacts+"/"+path.slice(1).replace("/","-")+"-mobile.png"});
    if (["/", "/messages"].includes(path))
      await page.screenshot({
        path: `${artifacts}/${path === "/" ? "home" : "messages"}-mobile.png`,
        fullPage: true,
      });
  }
  assert.deepEqual(errors, []);
  await page.route('**/v1/blog/about', route => route.fulfill({ json: { code: 0, data: { homeVerse: '后端设置的短句', homeAnimation: false } } }));
  await page.goto(origin, { waitUntil: 'networkidle' });
  await page.getByRole('heading', { name: '后端设置的短句' }).waitFor();
  await waitForFullVerse();
  assert.equal(await page.locator('.sky-rain').isVisible(), false);
  await page.route('**/v1/blog/about', route => route.fulfill({ json: { code: 0, data: { homeVerse: '切回博客后继续播放', homeAnimation: true } } }));
  await page.evaluate(() => window.dispatchEvent(new Event('focus')));
  await page.getByRole('heading', { name: '切回博客后继续播放' }).waitFor();
  await checkMeteorFlight();
  await checkReaderLayout({browser,page,apiFixture,artifacts,origin});
  customPaths=true;
  signedIn=false;
  await page.goto(origin+"/read/1",{waitUntil:"networkidle"});
  await openReaderDock(page);
  await page.getByRole("button",{name:/^收藏/}).click();
  await page.waitForURL("**/signin?**");
  await page.getByRole("link",{name:/注册账号/}).click();
  await page.waitForURL("**/join?**");
  await page.getByRole("link",{name:/登录 →/}).click();
  await page.locator('input[type="email"]').fill(reader.email);
  await page.locator('[autocomplete="current-password"]').fill("Reader123!");
  await page.getByRole("button",{name:"登录",exact:true}).click();
  await page.waitForURL("**/read/1");
  signedIn=false;
  await page.goto(origin+"/join",{waitUntil:"networkidle"});
  await page.locator('input[type="email"]').fill(reader.email);
  await page.getByPlaceholder("6 位验证码").fill("123456");
  await page.locator('[autocomplete="new-password"]').fill("Reader123!");
  await page.getByRole("button",{name:"注册并登录"}).click();
  await page.waitForURL("**/library");
  savedIn="0";
  await page.reload({waitUntil:"networkidle"});
  await page.getByRole("link",{name:/Isolated article fixture/}).click();
  await page.waitForURL("**/read/1");
  await checkJournalLayout({page, origin, artifacts});
  assert.equal(signedIn, false);
  assert.deepEqual(errors, []);
  console.log(
    "Blog browser checks passed: dynamic page paths, registration and login, folder creation and saved articles, emoji, moderated submissions and responsive layouts",
  );
} finally {
  await browser.close();
  server.closeAllConnections();
  await new Promise((resolve) => server.close(resolve));
}
