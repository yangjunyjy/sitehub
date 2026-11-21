import assert from 'node:assert/strict'
import { once } from 'node:events'
import { mkdir } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { join } from 'node:path'
import { chromium } from 'playwright'
import { createFrontendServer } from '../server/server.mjs'

const baseline = process.argv.includes('--baseline')
const artifacts = process.env.FILTER_SCREENSHOTS
if (artifacts) await mkdir(artifacts, { recursive: true })
const menus = [
  ['1', '菜单管理', '/menus', 'system/menus/index'],
  ['2', '用户管理', '/users', 'system/users/index'],
  ['3', '工作台', '/dashboard/overview', 'dashboard/index'],
].map(([id, menu_name, path, component]) => ({
  id, menu_name, menu_code: component, path, component, menu_type: 2, status: 1, is_visible: true, parent_id: '0',
}))
const rows = [...menus.slice(0, 2), { ...menus[0], id: '4', menu_name: '菜单编辑', menu_type: 3 }]
const requests = []
const server = createFrontendServer({ dist: fileURLToPath(new URL('../dist/', import.meta.url)) })
server.listen(0, '127.0.0.1')
await once(server, 'listening')
const origin = 'http://127.0.0.1:' + server.address().port
let browser
try {
  browser = await chromium.launch({ channel: process.env.PLAYWRIGHT_CHANNEL || undefined })
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.setDefaultTimeout(8000)
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  await page.addInitScript(() => sessionStorage.setItem('token', 'filter-test-session'))
  await page.route('**/v1/**', async route => {
    const url = new URL(route.request().url())
    let data = { items: [], total: 0 }
    if (url.pathname.endsWith('/auth/me')) data = { id: '1', username: 'admin', roleIds: ['1'] }
    if (url.pathname.endsWith('/auth/menus')) data = { items: menus }
    if (url.pathname.endsWith('/auth/permissions')) data = { items: ['*'] }
    if (url.pathname.endsWith('/rbac/systems')) data = { items: [{ name: '后台管理', code: 'admin' }, { name: '产品站', code: 'product' }] }
    if (url.pathname.endsWith('/rbac/users')) data = { items: [{ id: '21', username: 'editor', roles: [{ id: '4', roleName: '编辑' }] }], total: 1 }
    if (url.pathname.endsWith('/rbac/roles')) {
      const roles = [{ id: '4', roleName: '编辑', roleCode: 'writer', status: 1 }, { id: '5', roleName: '审核员', roleCode: 'reviewer', status: 1 }]
      const items = roles.filter(role => role.roleCode.includes(url.searchParams.get('roleCode') || ''))
      data = { items, total: items.length }
    }
    if (url.pathname.endsWith('/menus/tree')) {
      requests.push(url.searchParams.get('system'))
      data = { items: rows }
    }
    await route.fulfill({ json: { code: 0, data } })
  })
  const screenshot = async name => {
    if (artifacts) await page.screenshot({ path: join(artifacts, name + '.png'), animations: 'disabled' })
  }
  await page.goto(origin + '/menus')
  await page.locator('.el-table__body').getByText('菜单编辑', { exact: true }).first().waitFor()
  const filters = page.getByRole('main', { name: '菜单管理' }).getByRole('region', { name: '查询条件' })
  const table = page.locator('.el-table__body').first()
  const fields = filters.locator('.el-form-item__content > :is(.el-input,.el-select)')
  const more = filters.getByRole('button', { name: '展开', exact: true })
  const less = filters.getByRole('button', { name: '收起', exact: true })
  for (const width of [1440, 1920]) {
    await page.setViewportSize({ width, height: 900 })
    if (!baseline) {
      assert.equal(await more.getAttribute('aria-expanded'), 'false')
      await screenshot('desktop-' + width + '-collapsed')
      await more.click()
    }
    const sizes = await fields.evaluateAll(nodes => nodes.map(el => +el.getBoundingClientRect().width.toFixed(2)))
    const tabHeight = await page.locator('.tags-view').evaluate(el => el.getBoundingClientRect().height)
    console.log(JSON.stringify({ width, sizes, tabHeight }))
    await screenshot((baseline ? 'before-' : 'desktop-') + width + (baseline ? '' : '-expanded'))
    if (!baseline) {
      const before = width === 1440 ? 404 : 398.67
      assert(sizes.every(size => Math.abs(size - before / 3) < 2))
      assert(tabHeight <= 32)
      await less.click()
    }
  }
  if (!baseline) {
    await page.setViewportSize({ width: 1440, height: 900 })
    await more.click()
    const type = filters.getByRole('combobox', { name: '菜单类型' })
    await filters.getByText('全部类型', { exact: true }).click()
    await page.getByRole('option', { name: '按钮', exact: true }).click()
    const name = filters.getByRole('textbox', { name: '菜单名称' })
    await name.fill('菜单')
    await less.click()
    assert.equal(await type.isVisible(), false)
    assert.equal(requests.length, 1)
    await name.press('Enter')
    await table.getByText('菜单编辑', { exact: true }).waitFor()
    await page.waitForFunction(() => document.querySelectorAll('.el-table__body-wrapper tbody tr').length === 1)
    assert.equal(requests.length, 2)
    await more.click()
    assert.equal(await filters.getByText('按钮', { exact: true }).count(), 1)
    await less.click()
    await filters.getByRole('button', { name: '重置' }).click()
    await table.getByText('用户管理', { exact: true }).waitFor()
    assert.equal(requests.length, 3)
    await page.getByRole('menuitem', { name: '用户管理' }).click()
    await page.waitForURL('**/users')
    await page.getByRole('button', { name: '分配角色', exact: true }).click()
    const roleDialog = page.getByRole('dialog', { name: '分配角色' })
    await roleDialog.locator('.selected-item').getByText('编辑', { exact: true }).waitFor()
    await roleDialog.getByRole('button', { name: '展开', exact: true }).click()
    const code = roleDialog.getByRole('textbox', { name: '角色编码' })
    await code.fill('reviewer')
    const roleQuery = page.waitForResponse(res => new URL(res.url()).searchParams.get('roleCode') === 'reviewer')
    await code.press('Enter')
    await roleQuery
    await roleDialog.locator('tr').filter({ hasText: '审核员' }).locator('.el-checkbox__inner').click()
    assert.deepEqual(await roleDialog.locator('.selected-item .el-tag').allTextContents(), ['编辑', '审核员'])
    await roleDialog.getByRole('button', { name: '重置', exact: true }).click()
    await roleDialog.locator('tr').filter({ hasText: '编辑' }).locator('.el-checkbox__inner').waitFor()
    assert.deepEqual(await roleDialog.locator('.selected-item .el-tag').allTextContents(), ['编辑', '审核员'])
    await roleDialog.getByRole('button', { name: '取消', exact: true }).click()
    const tabs = page.getByRole('navigation', { name: '已打开的页面' })
    const menuTab = tabs.getByRole('button', { name: '菜单管理', exact: true })
    await menuTab.click()
    await page.waitForURL('**/menus')
    await menuTab.press('ArrowRight')
    await page.waitForURL('**/users')
    await tabs.getByRole('button', { name: '用户管理', exact: true }).click({ button: 'right' })
    await page.getByRole('menuitem', { name: '关闭当前' }).click()
    await page.waitForURL('**/menus')
    assert.equal(await tabs.getByRole('button', { name: '关闭 用户管理' }).count(), 0)
    await page.setViewportSize({ width: 390, height: 844 })
    await more.click()
    await screenshot('mobile-expanded')
    assert(await filters.getByRole('button', { name: '查询' }).isVisible())
    assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
    assert((await fields.first().boundingBox()).width >= 190)
    await less.click()
    assert.deepEqual(errors, [])
    console.log('PASS Enter query, hidden criteria, reset, keyboard tabs, close menu and mobile form')
  }
} finally {
  await browser?.close()
  server.closeAllConnections()
  await new Promise(resolve => server.close(resolve))
}
