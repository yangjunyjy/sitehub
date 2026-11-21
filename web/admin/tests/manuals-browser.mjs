import assert from 'node:assert/strict'
import { once } from 'node:events'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'
import { createFrontendServer } from '../server/server.mjs'

const menus = [
  ['1', '文档工作区', '/manuals', 'product/docs/index'],
  ['2', '文档编辑', '/manuals/edit', 'product/docs/editor/index'],
].map(([id, menu_name, path, component]) => ({
  id,
  menu_name,
  path,
  component,
  status: 1,
  menu_type: 2,
  is_visible: id === '1',
}))
let doc = {
  id: 'welcome',
  title: '安装指南',
  slug: 'install',
  summary: '安装工具',
  content: '# 安装工具',
  revision: '2',
  publishedRevision: '1',
  groupId: '9',
  status: 'modified',
}
const removed = new Set()
const calls = [],
  errors = []
let conflict = false,
  readError = true
const server = process.env.MANUAL_TEST_ORIGIN
  ? null
  : createFrontendServer({ dist: fileURLToPath(new URL('../dist/', import.meta.url)) })
if (server) {
  server.listen(0, '127.0.0.1')
  await once(server, 'listening')
}
const origin = process.env.MANUAL_TEST_ORIGIN || 'http://127.0.0.1:' + server.address().port
let browser
try {
  browser = await chromium.launch({ channel: process.env.PLAYWRIGHT_CHANNEL || undefined })
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 } })
  page.setDefaultTimeout(10000)
  await page.context().route('**/help/guide?**', route => route.fulfill({ contentType: 'text/html', body: '产品文档' }))
  page.on('pageerror', (e) => errors.push(e.message))
  await page.addInitScript(() => sessionStorage.setItem('token', 'manual-workspace-test'))
  await page.route('**/v1/**', async (route) => {
    const req = route.request(),
      url = new URL(req.url()),
      path = url.pathname
    const action = path.split('/manuals/')[1]
    const payload = req.method() === 'POST' ? req.postDataJSON() : null
    let data = { items: [] }
    if (path.endsWith('/menus/public')) data = {items:[{ name:'帮助',path:'/help',children:[{name:'指南',path:'guide',component:'docs/index'}]}]}
    if (path.endsWith('/auth/me')) data = { id: '1', roleIds: ['1'] }
    if (path.endsWith('/auth/menus')) data = { items: menus }
    if (path.endsWith('/auth/permissions')) data = { items: ['*'] }
    if (path.endsWith('/auth/totp')) data = { enabled: false }
    if (action) calls.push({ action, query: url.search, payload })
    if (action?.endsWith('/delete')) removed.add(action + '/' + payload.id)
    if (action === 'workspace') {
      if (readError) {
        readError = false
        return route.fulfill({ status: 500, json: { code: 500, message: '读取失败' } })
      }
      const { content, ...summary } = doc
      const mode = url.searchParams.get('mode') || 'cli'
      const version = url.searchParams.get('version') || (mode === 'ui' ? (removed.has('versions/delete/next') ? '' : 'next') : 'v1')
      data = {
        title: '产品手册',
        description: '安装和操作指南',
        revision: '9',
        defaultMode: 'cli',
        defaultVersion: 'v1',
        modeId: mode,
        versionId: version,
        modes: [
          { id: 'cli', label: '命令行', status: 'published', defaultVersion: 'v1' },
          { id: 'ui', label: '控制台', status: 'upcoming', revision:'5' },
        ].filter(m => !removed.has('modes/delete/'+m.id)),
        versions: mode === 'ui' ? (version ? [{id:'next', name:'next',status:'draft',revision:'4'}] : []) : ['v1', 'v2'].map((id) => ({ id, name: id, status: 'published', revision: '1' })),
        groups: mode === 'ui' ? [] : [
          { id: '9', title: '入门', revision: '2', items: removed.has('documents/delete/'+doc.id) ? [] : [summary] },
          { id: '10', title: '运维', revision: '1', items: [] },
        ].filter(g => !removed.has('groups/delete/'+g.id)),
      }
    } else if (action === 'document') data = doc
    else if (action === 'history')
      data = [{ id: '31', revision: '1', title: '旧版指南', createdAt: '2025-01-01T00:00:00Z' }]
    else if (action === 'history/detail')
      data = { id: '31', revision: '1', title: '旧版指南', contentMarkdown: '早期安装步骤' }
    else if (action?.startsWith('documents/')) {
      if (conflict) {
        conflict = false
        doc = { ...doc, revision: '20', title: '其他编辑者', content: '服务器修改' }
        return route.fulfill({ status: 409, json: { code: 409, message: '修订已更新' } })
      }
      if (Number(doc.revision) !== payload.revision) throw new Error('Unexpected document revision')
      const revision = String(Number(doc.revision) + 1)
      if (action === 'documents/save')
        doc = {
          ...doc,
          title: payload.title,
          content: payload.content,
          revision,
          status: doc.status === 'draft' ? 'draft' : 'modified',
        }
      if (action === 'documents/publish')
        doc = { ...doc, revision, publishedRevision: revision, status: 'published' }
      if (action === 'documents/withdraw')
        doc = { ...doc, revision, publishedRevision: '0', status: 'draft' }
      data = doc
    }
    return route.fulfill({ json: { code: 0, data } })
  })
  await page.goto(origin + '/manuals')
  await page.getByRole('button', { name: '重新加载' }).click()
  await page.getByRole('combobox', { name: '文档模式' }).waitFor()
  assert.equal(await page.getByRole('heading', { name: '产品手册' }).count(), 0)
  assert.equal(await page.getByText('安装和操作指南', { exact: true }).count(), 0)
  if (process.env.SITEHUB_PREVIEW_DIR) {
    await page.locator('.el-message--error').waitFor({ state: 'hidden' })
    await page.screenshot({ path: process.env.SITEHUB_PREVIEW_DIR + '/manual-workspace.png' })
  }
  const opened = page.context().waitForEvent('page')
  await page.getByRole('button', {name:'查看前台'}).click()
  const preview = await opened
  await preview.waitForURL(url => url.pathname === '/help/guide')
  assert.equal(new URL(preview.url()).searchParams.get('version'), 'v1')
  await preview.close()
  await page.getByRole('button', { name: /^运维/ }).click()
  await page.getByText('分组内暂无文档').waitFor()
  assert.equal(calls.filter((c) => c.action === 'workspace').length, 2)
  await page.getByRole('button', { name: /^入门/ }).click()
  assert.equal(calls.filter((c) => c.action === 'document').length, 0)
  assert.equal(await page.getByRole('button', {name:'删除模式',exact:true}).getAttribute('aria-disabled'),'true')
  assert.equal(await page.getByRole('button', {name:'删除文档「安装指南」'}).getAttribute('aria-disabled'),'true')
  const removeGroup = page.getByRole('button', {name:'删除分组「运维」'})
  await removeGroup.click()
  const confirmation = page.getByRole('dialog', {name:'删除分组'})
  await confirmation.getByText('永久删除分组「运维」？此操作无法撤销。').waitFor()
  const box = await confirmation.locator('.el-message-box').boundingBox()
  assert.ok(box.width <= 480, 'delete confirmation must stay compact on desktop')
  if (process.env.SITEHUB_PREVIEW_DIR) {
    await page.screenshot({path:process.env.SITEHUB_PREVIEW_DIR+'/manual-delete-confirmation.png',animations:'disabled'})
  }
  await confirmation.getByRole('button',{name:'取消'}).click()
  assert.equal(calls.filter(c => c.action === 'groups/delete').length,0)
  await removeGroup.click()
  const deleting = page.waitForRequest(req => req.url().endsWith('/groups/delete'))
  await confirmation.getByRole('button',{name:'确认删除'}).click()
  assert.deepEqual((await deleting).postDataJSON(),{version:'v1',id:10,revision:1})
  await confirmation.waitFor({state:'hidden'})
  const combo = page.getByRole('combobox', { name: '文档版本' })
  await combo.press('Enter')
  await page.getByRole('option', { name: 'v2', exact: true }).click()
  await page.waitForURL('**/manuals?mode=cli&version=v2')
  const search = page.getByRole('textbox', { name: '搜索分组文档' })
  await search.fill('找不到的标题')
  await page.getByText('没有匹配的文档').waitFor()
  await search.fill('install')
  await page.getByRole('button', { name: '文档站设置' }).click()
  const settings = page.getByRole('dialog', { name: '文档站设置' })
  await settings.getByRole('textbox', { name: '文档站标题' }).fill('部署手册')
  const settingsSave = page.waitForRequest(req => req.url().endsWith('/settings/save'))
  await settings.getByRole('button', { name: '保存' }).click()
  assert.deepEqual((await settingsSave).postDataJSON(), { title:'部署手册', description:'安装和操作指南', revision:9 })
  await settings.waitFor({ state: 'hidden' })
  await page.getByRole('button', { name: '新建分组', exact:true }).click()
  const groupForm = page.getByRole('dialog', { name:'新建分组' })
  await groupForm.getByRole('textbox', { name:'分组名称' }).fill('监控部署')
  const groupSave = page.waitForRequest(req => req.url().endsWith('/groups/save'))
  await groupForm.getByRole('button', { name:'保存' }).click()
  assert.deepEqual((await groupSave).postDataJSON(), { id:0, title:'监控部署', version:'v2', revision:0 })
  await groupForm.waitFor({ state:'hidden' })
  await page.getByRole('button', { name: '安装指南', exact: true }).click()
  const title = page.getByRole('textbox', { name: '文档标题' })
  await title.waitFor()
  assert.equal(await title.inputValue(), '安装指南')
  assert.equal(calls.filter((c) => c.action === 'history').length, 0)
  await page.getByRole('button', { name: '修订记录', exact: true }).click()
  await page.getByRole('button', { name: /#1/ }).click()
  await page.getByText('早期安装步骤').waitFor()
  await page.keyboard.press('Escape')
  await title.fill('保留我的编辑')
  conflict = true
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await page.getByRole('button', { name: '查看服务器版本' }).click()
  await page.getByText('服务器修改', { exact: true }).waitFor()
  await page.getByRole('button', { name: '保留本地稿并采用此修订号' }).click()
  assert.equal(await title.inputValue(), '保留我的编辑')
  await page.getByRole('button', { name: '保存草稿', exact: true }).click()
  await page.getByText('草稿已保存', { exact: true }).waitFor()
  assert.equal(calls.at(-1).payload.revision, 20)
  assert.equal(doc.status, 'modified')
  await page.getByRole('button', { name: '发布文档', exact: true }).click()
  await page
    .getByRole('dialog', { name: '发布文档' })
    .getByRole('button', { name: '确认', exact: true })
    .click()
  await page.getByText('文档已发布', { exact: true }).waitFor()
  await page.getByRole('button', { name: '撤回文档', exact: true }).click()
  await page
    .getByRole('dialog', { name: '撤回文档' })
    .getByRole('button', { name: '确认', exact: true })
    .click()
  await page.getByText('文档已撤回', { exact: true }).waitFor()
  assert.equal(doc.status, 'draft')
  async function confirmDelete(label,kind,action,payload) {
    await page.getByRole('button',{name:label,exact:true}).click()
    const pending = page.waitForRequest(r => r.url().endsWith('/'+action))
    await page.getByRole('dialog',{name:'删除'+kind}).getByRole('button',{name:'确认删除'}).click()
    assert.deepEqual((await pending).postDataJSON(),payload)
  }
  await page.goto(origin+'/manuals?mode=cli&version=v2')
  await confirmDelete('删除文档「'+doc.title+'」','文档','documents/delete',{id:doc.id,version:'v2',revision:Number(doc.revision)})
  await page.getByText('分组内暂无文档').waitFor()
  await page.getByRole('combobox',{name:'文档模式'}).press('Enter')
  await page.getByRole('option',{name:'控制台',exact:true}).click()
  await confirmDelete('删除版本','版本','versions/delete',{id:'next',revision:4})
  await page.getByText('此模式暂无版本').waitFor()
  await confirmDelete('删除模式','模式','modes/delete',{id:'ui',revision:5})
  await page.waitForURL('**/manuals')
  assert.deepEqual(errors, [])
  console.log(
    'PASS workspace retry, version/group navigation, on-demand document and history, conflict recovery, draft, publish, withdraw and all four deletion actions',
  )
} finally {
  await browser?.close()
  if (server) {
    server.closeAllConnections()
    await new Promise((resolve) => server.close(resolve))
  }
}
