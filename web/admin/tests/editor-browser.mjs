import assert from 'node:assert/strict'
import { once } from 'node:events'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'
import { createFrontendServer } from '../server/server.mjs'

function gate() {
  let resolve
  const promise = new Promise(r => { resolve = r })
  return { resolve, promise }
}

const menus = [
  ['1', '新建文章', '/write', 'blog/articles/create/index'],
  ['2', '编辑文章', '/draft/:id', 'blog/articles/editor/index'],
  ['3', '文章列表', '/articles', 'blog/articles/index'],
].map(([id, menu_name, path, component]) => ({
  id, menu_name, path, component, menu_type: 2, status: 1, is_visible: id !== '2',
}))

const server = createFrontendServer({ dist: fileURLToPath(new URL('../dist/', import.meta.url)) })
server.listen(0, '127.0.0.1')
await once(server, 'listening')
const origin = 'http://127.0.0.1:' + server.address().port
let browser
const errors = []

async function openEditor(totp = false) {
  const page = await browser.newPage({ viewport: { width: 1440, height: 960 } })
  page.setDefaultTimeout(10000)
  page.on('pageerror', error => errors.push(error.message))
  await page.addInitScript(() => sessionStorage.setItem('token', 'editor-test-session'))
  const requests = []
  let article
  const gates = [gate(), gate()]
  const listGate = gate()
  let listRows = [
    { id: '48', title: 'MODULES', status: 1, version: '2', views: 9, updatedAt: '2026-11-19T17:21:00' },
    { id: '49', title: '博客启程', status: 1, version: '1', views: 5, updatedAt: '2026-11-19T11:19:00' },
  ]
  await page.route('**/v1/**', async route => {
    const request = route.request()
    const path = new URL(request.url()).pathname
    let data = { items: [], total: 0 }
    if (path.endsWith('/auth/me')) data = { id: '1', roleIds: ['1'] }
    if (path.endsWith('/auth/menus')) data = { items: menus }
    if (path.endsWith('/auth/permissions')) data = { items: ['*'] }
    if (path.endsWith('/auth/totp')) data = { enabled: totp }
    if (path.endsWith('/articles/options')) data = { categories: [], tags: [] }
    if (path.endsWith('/articles/detail')) throw new Error('First draft promotion reloaded the article')
    if (path.endsWith('/admin/articles')) {
      await listGate.promise
      data = { items: listRows, total: listRows.length }
    }
    if (path.endsWith('/articles/action') && request.postDataJSON().action === 'delete') {
      const body = request.postDataJSON()
      requests.push(body)
      listRows = listRows.filter(row => row.id !== String(body.id))
    } else if (path.endsWith('/articles/draft') || path.endsWith('/articles/action')) {
      const body = request.postDataJSON()
      requests.push({ ...body, token: request.headers()['x-totp-token'] })
      const index = requests.length - 1
      if (!totp && index < gates.length) await gates[index].promise
      article = {
        ...article, ...body.input, id: '37', version: String(index + 1),
        slug: 'server-slug', status: body.action === 'publish' ? 1 : 0,
      }
      data = article
    }
    await route.fulfill({ json: { code: 0, data } })
  })
  await page.goto(origin + '/write')
  return { page, requests, gates, listGate }
}

try {
  browser = await chromium.launch({ channel: process.env.PLAYWRIGHT_CHANNEL || undefined })
  const { page, requests, gates, listGate } = await openEditor()
  const title = page.getByRole('textbox', { name: '文章标题' })
  const body = page.locator('.vditor-ir [contenteditable=true]').first()
  await body.fill('草稿原文')
  await title.fill('第一版')
  await page.getByText('正在保存草稿…', { exact: true }).waitFor()
  await page.getByRole('button', { name: '发布预览', exact: true }).click()
  await page.getByRole('heading', { name: '发布预览' }).waitFor()
  await page.getByRole('button', { name: '关闭发布预览' }).click()
  await body.fill('慢请求期间继续写的正文')
  await title.fill('慢请求期间继续写的标题')
  await title.evaluate(el => { window.focusedTitle = el })
  const nextDraft = page.waitForRequest(req => req.url().endsWith('/articles/draft'))
  await page.keyboard.press('Control+s')
  gates[0].resolve()
  await page.waitForURL('**/draft/37')
  await nextDraft
  assert.equal(await title.inputValue(), '慢请求期间继续写的标题')
  assert(await title.evaluate(el => el === window.focusedTitle && el === document.activeElement))
  assert.deepEqual(await page.evaluate(() => JSON.parse(sessionStorage.getItem('deploy-admin-tabs:1'))), ['/dashboard', '/draft/37'])
  assert.equal(requests[0].id, 0)
  assert.equal(requests[1].id, 37)
  assert.equal(requests[1].version, 1)
  assert.equal(requests[1].input.slug, 'server-slug')
  assert.match(requests[1].input.content, /慢请求期间继续写的正文/)
  await title.fill('发布版本')
  const published = page.waitForResponse(res => res.url().endsWith('/articles/action'))
  await page.getByRole('button', { name: '发布文章', exact: true }).click()
  gates[1].resolve()
  await published
  assert.deepEqual(requests.map(({ version }) => version), [0, 1, 2, 3])
  assert.equal(requests[2].input.title, '发布版本')
  assert.equal(requests[3].action, 'publish')
  console.log('PASS slow saves preserve typing, preview and focus, then publish the latest version')

  await page.goto(origin + '/articles', { waitUntil: 'domcontentloaded' })
  const ring = page.locator('.el-loading-spinner .circular').first()
  await ring.waitFor()
  assert.equal(await ring.evaluate(el => getComputedStyle(el).animationDuration), '2s')
  await page.emulateMedia({ reducedMotion: 'reduce' })
  assert.equal(await ring.evaluate(el => getComputedStyle(el).animationName), 'none')
  assert.equal(await ring.locator('.path').evaluate(el => getComputedStyle(el).animationName), 'none')
  listGate.resolve()
  console.log('PASS delayed article list uses a 2s spinner and honors reduced motion')

  const row = page.locator('.el-table__row').filter({ hasText: 'MODULES' })
  const trash = row.getByRole('button', { name: '移入回收站' })
  await trash.hover()
  await page.getByRole('tooltip', { name: '移入回收站' }).waitFor()
  const contrast = await row.locator('.article-copy p').evaluate(summary => {
    const luminance = color => {
      const [r, g, b] = color.match(/\d+/g).slice(0, 3).map(Number).map(value => {
        const s = value / 255
        return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4
      })
      return 0.2126 * r + 0.7152 * g + 0.0722 * b
    }
    return (luminance(getComputedStyle(summary.closest('td')).backgroundColor) + 0.05) /
      (luminance(getComputedStyle(summary).color) + 0.05)
  })
  assert.ok(contrast >= 4.5, 'hovered article summaries must remain legible')
  const countBeforeTrash = requests.length
  await trash.click()
  const confirmTrash = page.getByRole('dialog', { name: '移入回收站' })
  const panel = confirmTrash.locator('.el-message-box')
  await confirmTrash.getByText(/MODULES/).waitFor()
  const desktopBox = await panel.boundingBox()
  assert.ok(desktopBox.width < 500 && Math.abs(desktopBox.x + desktopBox.width / 2 - 720) < 2)
  await confirmTrash.getByRole('button', { name: '取消' }).click()
  await confirmTrash.waitFor({ state: 'hidden' })
  assert.equal(requests.length, countBeforeTrash)
  await trash.click()
  await page.setViewportSize({ width: 390, height: 844 })
  const mobileBox = await panel.boundingBox()
  assert.ok(mobileBox.x >= 16 && mobileBox.x + mobileBox.width <= 374)
  await confirmTrash.getByRole('button', { name: '移入回收站', exact: true }).click()
  await row.waitFor({ state: 'detached' })
  assert.equal(requests.length, countBeforeTrash + 1)
  assert.equal(requests.at(-1).id, 48)
  assert.equal(requests.at(-1).version, 2)
  await page.close()
  console.log('PASS readable row hover and compact trash confirmation on desktop and mobile')

  const protectedEditor = await openEditor(true)
  const securePage = protectedEditor.page
  const secureTitle = securePage.getByRole('textbox', { name: '文章标题' })
  await secureTitle.fill('本地自动保存')
  await securePage.getByText('草稿已备份到本机，保存到服务器时需验证身份').waitFor()
  assert.equal(await securePage.getByRole('dialog').count(), 0)
  assert.equal(protectedEditor.requests.length, 0)
  assert.equal(await securePage.evaluate(() => JSON.parse(localStorage.getItem('article-draft:1:new')).fields.title), '本地自动保存')
  await securePage.getByRole('button', { name: '保存草稿' }).click()
  const dialog = securePage.getByRole('dialog', { name: '确认操作' })
  await dialog.getByRole('textbox').fill('123456')
  await dialog.locator('button.el-button--primary').click()
  await securePage.waitForURL('**/draft/37')
  assert.equal(protectedEditor.requests[0].token, '123456')
  assert.deepEqual(errors, [])
  console.log('PASS protected accounts back up locally and explicitly confirm server writes')
} finally {
  await browser?.close()
  server.closeAllConnections()
  await new Promise(resolve => server.close(resolve))
}
