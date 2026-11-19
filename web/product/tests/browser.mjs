import assert from 'node:assert/strict';
import { existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { once } from 'node:events';
import { chromium } from 'playwright';
import { createFrontendServer } from '../server/server.mjs';

const server = createFrontendServer({ dist: fileURLToPath(new URL('../dist', import.meta.url)) });
server.listen(0, '127.0.0.1');
await once(server, 'listening');
const browserOptions = { headless: true };
const chrome = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
if (existsSync(chrome)) browserOptions.executablePath = chrome;
const browser = await chromium.launch(browserOptions);
const errors = [], requests = [];
let authenticated = false;
try {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
  page.on('pageerror', error => errors.push(error.message));
  await page.route('**/v1/**', async route => {
    const url = new URL(route.request().url());
    requests.push(url.pathname);
    let data, status = 200;
    if (url.pathname === '/v1/product/content') data = {
      brand: { name: 'Deploy', fullName: 'Deploy 产品平台' },
      navigation: [{ label: '首页', path: '/' }, { label: '下载', path: '/download' }, { label: '文档', path: '/docs' }],
      home: { description: '独立产品首页', meta: ['快速部署'] }, footer: {},
    };
    else if (url.pathname === '/v1/rbac/menus/public') {
      assert.equal(url.searchParams.get('system'), 'product');
      data = { items: [{ name: '首页', path: '/', isVisible: true }, { name: '下载', path: '/download' }, { name: '文档', path: '/docs' }, { name: '旧版手册', path: '/missing-page/deeper' }] };
    }
    else if (url.pathname === '/v1/rbac/members/me') { status = authenticated ? 200 : 401; data = authenticated ? { id: '1', email: 'reader@example.test' } : {}; }
    else if (url.pathname === '/v1/rbac/members/login') { authenticated = true; data = { token: 'isolated-test-token', user: { id: '1', email: 'reader@example.test' } }; }
    else if (url.pathname === '/v1/product/downloads') data = { title: '下载中心', packages: [], groups: [], installOptions: [], requirements: [] };
    else if (url.pathname === '/v1/product/docs/modes') data = { title: '产品文档', description: '产品文档', defaultMode: 'server', modes: [{ id: 'server', label: '服务器', available: true, versionIds: ['v1'] }] };
    else if (url.pathname === '/v1/product/docs/navigation') data = { title: '产品文档', versionId: 'v1', modeId: 'server', versions: [{ id: 'v1', label: 'v1' }], groups: [{ title: '快速开始', items: [{ id: '1', slug: 'install', title: '安装部署' }] }] };
    else if (url.pathname === '/v1/product/docs/document') data = { id: '1', title: '安装部署', content: '# 安装部署\n\n这是独立产品应用的文档。' };
    else throw new Error(`Unexpected product API: ${url.pathname}`);
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify({ code: status === 200 ? 0 : status, data }) });
  });
  const base = `http://127.0.0.1:${server.address().port}`;
  await page.goto(base);
  await page.getByText('独立产品首页').waitFor();
  assert.equal(await page.locator('a[href="/blog"]').count(), 0);
  await page.goto(base + '/download');
  await page.getByRole('heading', { name: '下载中心', exact: true }).waitFor();
  await page.goto(base + '/docs/install');
  await page.getByText('这是独立产品应用的文档。').waitFor();
  await page.goto(base + '/register');
  await page.getByRole('heading', { name: '创建你的账号' }).waitFor();
  assert.equal(await page.locator('#member-code').count(), 1);
  await page.goto(base + '/login');
  await page.locator('#member-email').fill('reader@example.test');
  await page.locator('#member-password').fill('TestPassword123');
  await page.locator('button[type="submit"]').click();
  await page.getByRole('heading', { name: 'reader@example.test' }).waitFor();
  await page.getByRole('navigation', { name: '主导航' }).getByRole('link', { name: '旧版手册' }).click();
  await page.waitForURL(base + '/missing-page/deeper');
  await page.reload();
  await page.getByRole('heading', { name: /这个页面.*不在交付路径上/ }).waitFor();
  assert.equal(await page.locator('.product-header').count(), 1);
  assert.equal(await page.locator('.product-footer').count(), 1);
  assert.equal(requests.some(path => path.startsWith('/v1/blog/')), false);
  assert.deepEqual(errors, []);
  console.log('Product browser smoke passed: home, downloads, docs, member routes and layout 404; zero blog requests or page errors.');
} finally { await browser.close(); await new Promise(resolve => server.close(resolve)); }
