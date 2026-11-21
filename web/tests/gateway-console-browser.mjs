import assert from 'node:assert/strict'
import { chromium } from 'playwright'

const origin = process.env.GATEWAY_CONSOLE_TEST_URL
if (!origin) throw Error('Run TestConsoleBrowser with GATEWAY_CONSOLE_BROWSER_TEST=1')
const browser = await chromium.launch({ channel: process.env.PLAYWRIGHT_CHANNEL || undefined })
try {
  const page = await browser.newPage()
  page.setDefaultTimeout(6000)
  const errors = []
  page.on('pageerror', (e) => errors.push(e.message))
  await page.goto(origin)
  await page.getByRole('heading', { name: '登录网关控制台' }).waitFor()
  const shot = async (name) => {
    if (process.env.SITEHUB_PREVIEW_DIR)
      await page.screenshot({
        path: process.env.SITEHUB_PREVIEW_DIR + '/gateway-' + name + '.png',
        animations: 'disabled',
      })
  }
  await shot('login')
  await page.getByLabel('用户名', { exact: true }).fill('operator')
  await page.getByLabel('邮箱', { exact: true }).fill('ops@example.test')
  await page.getByLabel('密码', { exact: true }).fill('fixture-password')
  await page.getByRole('button', { name: '获取验证码' }).click()
  const slider = page.getByRole('slider')
  await page.locator('.captcha-content[data-state="ready"]').waitFor()
  await slider.focus()
  for (let step = 0; step < 4; step++) {
    await slider.press('ArrowRight')
    await page.waitForTimeout(115)
  }
  await slider.press('Enter')
  await page.getByText('验证码已发送，请查收邮箱').waitFor()
  await page.getByLabel('邮箱验证码', { exact: true }).fill('000000')
  await page.getByRole('button', { name: '登录', exact: true }).click()
  await page.getByRole('alert').getByText('账号、密码或邮箱验证码有误').waitFor()
  await page.getByLabel('邮箱验证码', { exact: true }).fill('654321')
  await page.getByRole('button', { name: '登录', exact: true }).click()
  const nav = page.getByRole('navigation', { name: '主导航' })
  await nav.waitFor()
  await shot('workspace')
  assert((await page.context().cookies()).find((c) => c.name === 'sitehub_admin_session')?.httpOnly)
  await page.reload()
  await nav.waitFor()
  await nav.getByRole('link', { name: '上游服务' }).click()
  await page.getByRole('button', { name: '新增上游服务' }).click()
  await page.getByLabel('上游标识').fill('catalog')
  await page.getByLabel('服务发现名称').fill('catalog-api')
  await page.getByRole('button', { name: '保存到草稿' }).click()
  await page.getByText('catalog-api', { exact: true }).waitFor()
  await nav.getByRole('link', { name: '路由', exact: true }).click()
  await page.getByRole('button', { name: '新增路由' }).click()
  await page.getByLabel('路由标识').fill('catalog-route')
  await page.getByLabel('匹配主机').fill('catalog.test')
  await page.getByLabel('路径（每行一条）').fill('/v1/*')
  await page.getByLabel('目标上游').selectOption('catalog')
  await page.getByRole('button', { name: '保存到草稿' }).click()
  await page.getByRole('button', { name: '审阅并发布' }).click()
  await page.getByText('+ route/catalog-route', { exact: true }).waitFor()
  await page.getByText('+ upstream/catalog', { exact: true }).waitFor()
  if (process.env.SITEHUB_PREVIEW_DIR) {
    await page.screenshot({ path: process.env.SITEHUB_PREVIEW_DIR + '/gateway-publish-review.png' })
  }
  await page.getByRole('button', { name: '确认发布' }).click()
  await page.getByText('请输入有效的动态验证码').waitFor()
  await page.getByLabel(/动态验证码/).fill('123456')
  await page.getByRole('button', { name: '确认发布' }).click()
  await page.getByText('配置修订 8 · 已应用 8').waitFor()
  await nav.getByRole('link', { name: '发布记录' }).click()
  await page.getByRole('button', { name: '回滚', exact: true }).click()
  await page.getByLabel(/动态验证码/).fill('123456')
  await page.getByRole('button', { name: '确认回滚' }).click()
  await page.getByText('配置修订 9 · 已应用 9').waitFor()
  await nav.getByRole('link', { name: '路由', exact: true }).click()
  assert.equal(await page.getByText('catalog-route', { exact: true }).count(), 0)
  await page.setViewportSize({ width: 390, height: 844 })
  await page.getByRole('button', { name: '切换导航' }).click()
  await nav.getByRole('link', { name: '插件策略' }).click()
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
  await page.getByRole('button', { name: '退出登录' }).click()
  await page.getByRole('button', { name: '确认退出' }).click()
  await page.getByRole('heading', { name: '登录网关控制台' }).waitFor()
  assert.equal(
    (await page.context().cookies()).some((c) => c.name === 'sitehub_admin_session'),
    false,
  )
  assert.equal((await page.request.get(origin + 'api/workspace')).status(), 401)
  assert.deepEqual(errors, [])
} finally {
  await browser.close()
}
