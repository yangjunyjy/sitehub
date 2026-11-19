// Build the admin first, then run npm run test:captcha. All API calls are mocked;
// no backend, credentials, SMTP server or live email recipient is used.
import assert from 'node:assert/strict'
import { mkdir, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { chromium } from 'playwright'
import { createFrontendServer } from '../admin/server/server.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const { values } = parseArgs({ options: {
  dist: { type: 'string' },
  baseline: { type: 'string' },
  channel: { type: 'string' },
  artifacts: { type: 'string' },
} })
const artifacts = resolve(values.artifacts || resolve(root, '.tmp/captcha-browser'))
const results = []
let browser

async function withFrontend(dist, run) {
  const server = createFrontendServer({ app: 'admin', dist, websiteURL: '' })
  try {
    await new Promise((done, reject) => {
      server.once('error', reject)
      server.listen(0, '127.0.0.1', done)
    })
    await run(`http://127.0.0.1:${server.address().port}`)
  } finally {
    server.closeAllConnections()
    await new Promise(done => server.close(done))
  }
}

async function challengeImages(page) {
  return page.evaluate(() => {
    const png = (width, height, color) => {
      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const context = canvas.getContext('2d')
      context.fillStyle = color
      context.fillRect(0, 0, width, height)
      context.fillStyle = '#409eff'
      context.fillRect(20, 20, width / 2, height / 2)
      return canvas.toDataURL('image/png')
    }
    return { backgroundImage: png(300, 150, '#f3f6fa'), sliderImage: png(52, 52, '#b3d8ff') }
  })
}

async function verifyCaptcha(origin, { name, strings, expected = 'ready', drag = false }) {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 } })
  const page = await context.newPage()
  const errors = [], unexpectedRequests = [], verificationBodies = [], emailBodies = []
  page.setDefaultTimeout(10000)
  page.on('pageerror', error => errors.push(error.message))
  try {
    const images = await challengeImages(page)
    const geometry = { y: 19, imgWidth: 300, imgHeight: 150, pieceSize: 44, piecePadding: 4 }
    const data = {
      captchaKey: 'local-captcha-fixture', ...images,
      ...Object.fromEntries(Object.entries(geometry).map(([key, value]) => [key, strings ? String(value) : value])),
    }
    let challenges = 0
    await context.route('**/*', async route => {
      const request = route.request()
      const url = new URL(request.url())
      // Restrict the page to this local server. No API is ever forwarded.
      if (url.origin !== origin) {
        unexpectedRequests.push(url.origin)
        return route.abort()
      }
      if (!url.pathname.startsWith('/v1/')) return route.continue()
      let response
      if (url.pathname === '/v1/rbac/captcha' && request.method() === 'GET') {
        challenges++
        response = data
      } else if (url.pathname === '/v1/rbac/captcha/verify' && request.method() === 'POST') {
        verificationBodies.push(JSON.parse(request.postData()))
        response = { value: true }
      } else if (url.pathname === '/v1/rbac/auth/code' && request.method() === 'POST') {
        emailBodies.push(JSON.parse(request.postData()))
        response = { value: true }
      } else {
        unexpectedRequests.push(`${request.method()} ${url.pathname}`)
        return route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ code: 500, message: 'Unexpected local test API call' }) })
      }
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, message: 'ok', data: response }) })
    })
    await page.goto(origin + '/login', { waitUntil: 'networkidle' })
    await page.getByPlaceholder('请输入邮箱', { exact: true }).fill('captcha@example.test')
    await page.getByRole('button', { name: '获取验证码', exact: true }).click()
    const captcha = page.locator('.captcha-content')
    await page.locator(`.captcha-content[data-state="${expected}"]`).waitFor({ state: 'visible' })
    assert.equal(challenges, 1)
    await page.locator('.captcha-dialog').hover()
    await page.screenshot({ path: resolve(artifacts, name + '.png'), fullPage: true, animations: 'disabled' })
    if (expected === 'error') {
      assert.equal(await page.getByRole('slider', { name: '拖动拼图滑块' }).isDisabled(), true)
      assert.equal(verificationBodies.length, 0)
      assert.equal(emailBodies.length, 0)
    } else {
      const dimensions = await captcha.locator('img').evaluateAll(images => images.map(image => ({ width: image.naturalWidth, height: image.naturalHeight, complete: image.complete })))
      assert.deepEqual(dimensions, [{ width: 300, height: 150, complete: true }, { width: 52, height: 52, complete: true }])
      const slider = page.getByRole('slider', { name: '拖动拼图滑块' })
      assert.equal(await slider.isEnabled(), true)
      const pieceWidth = await captcha.locator('.captcha-piece').evaluate(image => image.getBoundingClientRect().width)
      assert(pieceWidth > 0 && pieceWidth < 100, 'piece should use the numeric 52px geometry before scaling')
      if (drag) {
        // Hover waits until the dialog's opening animation has settled.
        await slider.hover()
        const thumb = await slider.boundingBox()
        const track = await captcha.locator('.captcha-track').boundingBox()
        assert(thumb && track && track.width > thumb.width)
        const x = thumb.x + thumb.width / 2, y = thumb.y + thumb.height / 2
        const distance = (track.width - thumb.width) * 0.55
        await page.mouse.move(x, y)
        await page.mouse.down()
        for (let step = 1; step <= 12; step++) {
          await page.mouse.move(x + distance * step / 12, y)
          await page.waitForTimeout(40)
        }
        await page.mouse.up()
        await page.getByText('验证码已发送，请查收邮箱', { exact: true }).waitFor()
        assert.equal(verificationBodies.length, 1)
        const [body] = verificationBodies
        for (const field of ['sliderX', 'duration', 'moveCount']) assert(Number.isSafeInteger(body[field]), `${field} must be sent as an integer`)
        assert(body.sliderX > 0 && body.sliderX <= 256)
        assert(body.duration >= 300 && body.duration <= 10000)
        assert(body.moveCount >= 3)
        assert.equal(emailBodies.length, 1)
        assert.deepEqual(emailBodies[0], { email: 'captcha@example.test', captchaKey: 'local-captcha-fixture' })
      }
    }
    if (drag) await page.screenshot({ path: resolve(artifacts, name + '-email.png'), fullPage: true, animations: 'disabled' })
    assert.deepEqual(errors, [], 'page must not raise browser errors')
    assert.deepEqual(unexpectedRequests, [], 'all requests must remain local and use expected mocks')
    results.push({ name, expected, challenges, verified: verificationBodies.length, mockedEmails: emailBodies.length })
    console.log(`CAPTCHA PASS ${name}: ${expected}${drag ? ', numeric drag payload and mocked email handoff' : ''}`)
  } finally {
    await context.close()
  }
}

try {
  await mkdir(artifacts, { recursive: true })
  browser = await chromium.launch({ headless: true, channel: values.channel || process.env.PLAYWRIGHT_CHANNEL || undefined, args: ['--no-proxy-server', '--disable-background-networking'] })
  if (values.baseline) {
    await withFrontend(resolve(values.baseline), async origin => {
      await verifyCaptcha(origin, { name: 'baseline-strings', strings: true, expected: 'error' })
      await verifyCaptcha(origin, { name: 'baseline-numbers', strings: false })
    })
  }
  await withFrontend(resolve(values.dist || resolve(root, 'web/admin/dist')), async origin => {
    await verifyCaptcha(origin, { name: 'fixed-strings', strings: true, drag: true })
    await verifyCaptcha(origin, { name: 'fixed-numbers', strings: false, drag: true })
  })
} finally {
  await browser?.close()
  await writeFile(resolve(artifacts, 'result.json'), JSON.stringify({ results }, null, 2) + '\n')
}
