// Run via: python scripts/smoke.py --skip-build --frontends
// Uses only the isolated native-service fixtures created by smoke.py.
import assert from 'node:assert/strict'
import { chromium } from 'playwright'
import { mkdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'

if (!process.env.FRONTEND_ARTICLE_ID || !process.env.FRONTEND_ARTIFACT_DIR) {
  throw new Error('Run python scripts/smoke.py --frontends to create isolated test services and fixtures')
}
const artifacts = process.env.FRONTEND_ARTIFACT_DIR
await mkdir(artifacts, { recursive: true })
const errors = [], checks = [], responses = []
let browser
const check = name => { checks.push(name); console.log('WEB PASS', name) }
function observe(page, app) {
  page.on('pageerror', error => errors.push(`${app}: ${error.message}`))
  page.on('response', response => {
    if (response.url().includes('/v1/')) {
      responses.push({ app, path: new URL(response.url()).pathname, status: response.status() })
      if (response.status() >= 400 && response.status() !== 401) errors.push(`${app}: ${response.status()} ${response.url()}`)
    }
  })
}
async function visit(page, origin, path) {
  await page.goto(origin + path, { waitUntil: 'networkidle' })
  assert(!page.url().includes('/login'), `Unexpected login redirect: ${path}`)
  assert(await page.locator('body').innerText(), `Empty page: ${path}`)
  assert.equal(await page.getByRole('heading', { name: /^(403|404)$/ }).count(), 0, `Missing or inaccessible page: ${path}`)
}
try {
  const adminURL = process.env.WEB_ADMIN_URL, siteURL = process.env.WEB_SITE_URL
  const blogURL = process.env.WEB_BLOG_URL
  if (!adminURL || !siteURL || !blogURL) throw new Error('Browser tests require the unified gateway URLs')
  browser = await chromium.launch({ headless: true, channel: process.env.PLAYWRIGHT_CHANNEL || undefined, args: ['--host-resolver-rules=MAP sitehub.local 127.0.0.1, MAP blog.sitehub.local 127.0.0.1, MAP admin.sitehub.local 127.0.0.1', '--no-proxy-server'] })
  const adminContext = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
  const siteContext = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
  const admin = await adminContext.newPage(), site = await siteContext.newPage()
  observe(admin, 'admin'); observe(site, 'site')
  await admin.goto(adminURL + '/login')
  await admin.getByPlaceholder('请输入用户名', { exact: true }).fill('nativeadmin')
  await admin.getByPlaceholder('请输入邮箱', { exact: true }).fill('admin@example.test')
  await admin.getByPlaceholder('请输入密码', { exact: true }).fill('NativePassword123!')
  await admin.getByPlaceholder('请输入验证码', { exact: true }).fill('654321')
  await admin.getByRole('button', { name: '登录', exact: true }).click()
  await admin.waitForURL('**/dashboard/index')
  await admin.waitForLoadState('networkidle')
  assert((await adminContext.cookies()).some(c => c.name === 'sitehub_admin_session' && c.httpOnly))
  check('administrator UI login and independent HttpOnly session')
  await admin.screenshot({ path: resolve(artifacts, 'admin-desktop.png'), fullPage: true })
  for (const path of ['blog/articles', 'blog/statistics', 'blog/categories', 'blog/tags', 'system/systems', 'system/users', 'system/roles', 'system/menus', 'product/docs', 'product/content', 'product/downloads', 'admin/settings/basic', 'admin/settings/storage', 'admin/settings/email', 'profile']) {
    await visit(admin, adminURL, '/dashboard/' + path)
  }
  check('all native administration pages load without API errors')
  await visit(admin, adminURL, '/dashboard/product/docs')
  await admin.getByRole('button', {name:'文档站设置'}).click()
  await admin.getByRole('dialog').getByLabel('文档站标题').fill('Browser documentation')
  const savedSettings = admin.waitForResponse(r => r.url().endsWith('/v1/product/admin/manuals/settings/save') && r.request().method() === 'POST')
  await admin.getByRole('dialog').getByRole('button', { name: '保存', exact: true }).click()
  assert.equal((await savedSettings).status(), 200)
  await admin.waitForLoadState('networkidle')
  check('document settings preserve the native revision contract')
  await visit(admin, adminURL, '/dashboard/blog/categories')
  await admin.getByRole('button', { name: '新增分类', exact: true }).click()
  const dialog = admin.getByRole('dialog')
  await dialog.getByPlaceholder('请输入分类名称').fill('Browser category')
  await dialog.getByPlaceholder('请输入分类标识（英文）').fill('browser-category')
  const savedCategory = admin.waitForResponse(r => r.url().endsWith('/v1/blog/categories/create') && r.request().method() === 'POST')
  await dialog.getByRole('button', { name: '确定', exact: true }).click()
  assert.equal((await savedCategory).status(), 200)
  await dialog.waitFor({ state: 'hidden' })
  await admin.getByText('Browser category', { exact: true }).first().waitFor()
  check('category creation through the native API from the actual form')
  await visit(admin, adminURL, '/dashboard/blog/articles/editor/' + process.env.FRONTEND_ARTICLE_ID)
  await admin.getByPlaceholder('请输入文章标题').fill('Browser unpublished edit')
  const savedArticle = admin.waitForResponse(r => r.url().endsWith('/v1/blog/admin/articles/draft') && r.request().method() === 'POST')
  await admin.getByRole('button', { name: '保存草稿', exact: true }).click()
  assert.equal((await savedArticle).status(), 200)
  check('article editor saves native uint64 IDs and draft version')

  for (const path of ['/', '/docs/native-doc?version=v1', '/download']) {
    await visit(site, siteURL, path)
    if (path.startsWith('/docs/')) await site.getByRole('heading', { name: 'Native documentation', exact: true }).waitFor()
  }
  for (const path of ['/', '/categories', '/tags', '/archives', '/friends', '/messages', '/about']) await visit(site, blogURL, path)
  await visit(site, blogURL, '/article/' + process.env.FRONTEND_ARTICLE_ID)
  await site.getByRole('heading', { name: 'Native migration', exact: true }).waitFor()
  assert(!(await site.locator('body').innerText()).includes('Browser unpublished edit'))
  check('public site reads published articles, native documents and downloads')
  await site.screenshot({ path: resolve(artifacts, 'site-desktop.png'), fullPage: true })
  await site.goto(siteURL + '/login?redirect=/download')
  await site.locator('input[type=email]').fill('member@example.test')
  await site.locator('input[autocomplete=current-password]').fill('MemberPassword123')
  await site.getByRole('button', { name: '登录', exact: true }).click()
  await site.waitForURL('**/download')
  await site.waitForLoadState('networkidle')
  assert((await siteContext.cookies()).some(c => c.name === 'sitehub_site_session' && c.httpOnly))
  const ranged = await site.evaluate(async () => { const r = await fetch('/v1/file/download?id=native', { headers: { Range: 'bytes=2-5' } }); return { status: r.status, body: await r.text() } })
  assert.equal(ranged.status, 206)
  assert.equal(ranged.body, '2345')
  check('member UI login and credentialed streamed Range download')
  const memberLogout = await site.evaluate(async () => (await fetch('/v1/rbac/members/logout', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' })).status)
  assert.equal(memberLogout, 200)
  assert(!(await siteContext.cookies()).some(c => c.name === 'sitehub_site_session'))
  assert.equal(await site.evaluate(async () => (await fetch('/v1/file/download?id=native')).status), 401)
  check('member logout clears the cookie and revokes download access')
  await site.setViewportSize({ width: 390, height: 844 })
  await visit(site, blogURL, '/')
  await site.screenshot({ path: resolve(artifacts, 'site-mobile.png'), fullPage: true })
  await admin.setViewportSize({ width: 390, height: 844 })
  await visit(admin, adminURL, '/dashboard/blog/categories')
  await admin.screenshot({ path: resolve(artifacts, 'admin-mobile.png'), fullPage: true })
  await admin.setViewportSize({ width: 1440, height: 1000 })
  await visit(admin, adminURL, '/dashboard/profile')
  await admin.getByRole('tab', { name: '修改密码' }).click()
  await admin.locator('input[autocomplete=current-password]').fill('NativePassword123!')
  await admin.locator('input[autocomplete=new-password]').nth(0).fill('BrowserNewPassword123!')
  await admin.locator('input[autocomplete=new-password]').nth(1).fill('BrowserNewPassword123!')
  const changedPassword = admin.waitForResponse(r => r.url().endsWith('/v1/rbac/auth/password') && r.request().method() === 'POST')
  await admin.getByRole('button', { name: '更新密码', exact: true }).click()
  assert.equal((await changedPassword).status(), 200)
  check('password form uses the native current-session contract')
  assert.deepEqual(errors, [], 'Browser errors or failing native API calls')
  check('desktop and mobile pages render without browser errors')
} catch (error) {
  errors.push(error.stack)
  throw error
} finally {
  await writeFile(resolve(artifacts, 'result.json'), JSON.stringify({ checks, errors, responses }, null, 2))
  await browser?.close()
  console.log('Web artifacts:', artifacts)
}
