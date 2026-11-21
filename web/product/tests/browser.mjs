import assert from 'node:assert/strict';
import { existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { once } from 'node:events';
import http from 'node:http';
import { chromium } from 'playwright';
import { createFrontendServer } from '../server/server.mjs';

const frontend = createFrontendServer({ dist: fileURLToPath(new URL('../dist', import.meta.url)) });
let nativeTransfers = 0;
const server = http.createServer((req, res) => {
  if (!req.url.startsWith('/v1/file/download?')) return frontend.emit('request', req, res);
  assert.equal(new URL(req.url, 'http://localhost').searchParams.get('id'), 'release');
  assert.match(req.headers.cookie || '', /sitehub_site_session=browser-test/);
  nativeTransfers++;
  res.writeHead(200, { 'Content-Disposition':'attachment; filename="deploy.zip"', 'Content-Type':'application/octet-stream', 'Content-Length':7 });
  res.end('fixture');
});
server.listen(0, '127.0.0.1');
await once(server, 'listening');
const browserOptions = { headless: true };
const chrome = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE || 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
if (existsSync(chrome)) browserOptions.executablePath = chrome;
const browser = await chromium.launch(browserOptions);
const errors = [], requests = [];
let authenticated = false;
try {
  const page = await browser.newPage({ acceptDownloads: true, viewport: { width: 1440, height: 1000 } });
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
      data = { items: [
        { name: '首页', path: '/', component: 'home/index' },
        { name: '下载', path: '/artifacts', component: 'download/index' },
        { name: '支持', path: '/support', children: [{ name: '文档', path: 'manual', component: 'docs/index' }] },
      ] };
    }
    else if (url.pathname === '/v1/rbac/members/me') { status = authenticated ? 200 : 401; data = authenticated ? { id: '1', email: 'reader@example.test' } : {}; }
    else if (url.pathname === '/v1/rbac/members/login') { authenticated = true; await page.context().addCookies([{ name:'sitehub_site_session', value:'browser-test', url: new URL(page.url()).origin, httpOnly:true }]); data = { token: 'isolated-test-token', user: { id: '1', email: 'reader@example.test' } }; }
    else if (url.pathname === '/v1/product/downloads') data = {
      title: '下载中心', version: '1.0', groups: [{id:'tools', name:'发行文件'}],
      packages: [{ id:'release', groupId:'tools', product:'Deploy CLI', fileName:'deploy.zip', platform:'Linux', version:'1.0', size:8000, url:'/v1/file/download?id=release' }],
    };
    else if (url.pathname === '/v1/file/download') return route.continue();
    else if (url.pathname === '/v1/product/docs/catalog' && url.searchParams.get('mode') === 'retired') { status = 404; data = {}; }
    else if (url.pathname === '/v1/product/docs/catalog') data = {
      title: '产品文档', defaultMode:'server', modeId:'server',
      modes: [{ id:'server', label:'服务器', available:true, versionIds:['v1'] }],
      navigation: { title:'产品文档', versionId:'v1', defaultVersion:'v1', versions:[{id:'v1', name:'版本 1'}],
        groups:[{title:'快速开始', items:[{id:'1',slug:'install',title:'安装部署'},{id:'2',slug:'configure',title:'配置项目'}]}] }
    };
    else if (url.pathname === '/v1/product/docs/document') data = url.searchParams.get('id') === '2'
      ? { id:'2', slug:'configure', title:'配置项目', content:'# 配置项目\n\n配置目录与环境变量。' }
      : { id:'1', slug:'install', title:'安装部署', content:'# 安装部署\n\n这是独立产品应用的文档。' };
    else throw new Error(`Unexpected product API: ${url.pathname}`);
    await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify({ code: status === 200 ? 0 : status, data }) });
  });
  const base = `http://127.0.0.1:${server.address().port}`;
  await page.goto(base);
  await page.getByText('独立产品首页').waitFor();
  assert.equal(requests.filter(path => path === '/v1/rbac/menus/public').length, 1);
  assert.equal(await page.locator('a[href="/blog"]').count(), 0);
  await page.goto(base + '/support/manual/install');
  await page.getByText('这是独立产品应用的文档。').waitFor();
  const sidebar = page.getByRole('navigation', { name: '文档目录' });
  await sidebar.locator('summary').click();
  assert.equal(await sidebar.getByRole('link', { name: '配置项目' }).isVisible(), false);
  await sidebar.locator('summary').click();
  await sidebar.getByRole('link', { name: '配置项目' }).click();
  await page.getByText('配置目录与环境变量。').waitFor();
  assert.equal(requests.filter(path => path.endsWith('/docs/catalog')).length, 1);
  assert.equal(requests.filter(path => path.endsWith('/docs/document')).length, 2);
  await page.setViewportSize({width:390, height:844});
  await page.getByRole('button', { name:'文档目录', exact:true }).click();
  await sidebar.getByRole('link', {name:'安装部署'}).click();
  await page.getByText('这是独立产品应用的文档。').waitFor();
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  await page.goto(base + '/support/manual?mode=retired');
  await page.getByRole('link', {name:'返回默认文档'}).click();
  await page.getByText('这是独立产品应用的文档。').waitFor();
  await page.goto(base + '/register');
  await page.getByRole('heading', { name: '创建你的账号' }).waitFor();
  assert.equal(await page.locator('#member-code').count(), 1);
  await page.goto(base + '/artifacts');
  await page.getByRole('heading', {name:'Deploy CLI'}).waitFor();
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  await page.getByRole('searchbox', {name:/查找文件/}).fill('missing');
  await page.locator('.catalog-search button').click();
  await page.getByRole('heading', {name:'没有匹配的文件'}).waitFor();
  await page.locator('.catalog-count button').click();
  await page.getByRole('button', { name: '登录后下载' }).click();
  await page.locator('#member-email').fill('reader@example.test');
  assert.equal(new URL(page.url()).searchParams.get('redirect'), '/artifacts?package=release');
  await page.locator('#member-password').fill('TestPassword123');
  const downloadReady = page.waitForEvent('download');
  await page.locator('button[type="submit"]').click();
  const download = await downloadReady;
  assert.equal(download.suggestedFilename(), 'deploy.zip');
  assert.equal(await download.failure(), null);
  assert.equal(new URL(page.url()).searchParams.has('package'), false);
  assert.equal(nativeTransfers, 1);
  await page.reload();
  await page.getByRole('button', {name:'下载文件'}).waitFor();
  assert.equal(nativeTransfers, 1);
  await page.setViewportSize({width:1440,height:1000});
  await page.goto(base + '/account');
  await page.getByRole('heading', { name: 'reader@example.test' }).waitFor();
  const navigation = page.getByRole('navigation', { name: '主导航' });
  await navigation.getByRole('button', {name: '支持'}).click();
  assert.equal(await navigation.getByRole('link', {name: '文档'}).getAttribute('href'), '/support/manual');
  await page.goto(base + '/missing-page/deeper');
  await page.reload();
  await page.getByRole('heading', { name: /这个页面.*不在交付路径上/ }).waitFor();
  assert.equal(await page.locator('.product-header').count(), 1);
  assert.equal(await page.locator('.product-footer').count(), 1);
  assert.equal(requests.some(path => path.startsWith('/v1/blog/')), false);
  assert.deepEqual(errors, []);
  console.log('Product browser smoke passed: home, downloads, docs, member routes and layout 404; zero blog requests or page errors.');
} finally { await browser.close(); await new Promise(resolve => server.close(resolve)); }
