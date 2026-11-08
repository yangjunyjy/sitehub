import test from 'node:test'
import assert from 'node:assert/strict'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createFrontendServer } from './server.mjs'
import { gatewayProxy } from './dev-proxy.mjs'

test('production Node serves pages but never proxies or falls back for API requests', async t => {
  const dist = await mkdtemp(join(tmpdir(), 'sitehub-web-'))
  await writeFile(join(dist, 'index.html'), '<html>native frontend</html>')
  const server = createFrontendServer({ app: 'site', dist })
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve))
  const origin = `http://127.0.0.1:${server.address().port}`
  t.after(async () => { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)); await rm(dist, { recursive: true, force: true }) })
  assert.match(await (await fetch(origin + '/blog/reading')).text(), /native frontend/)
  for (const path of ['/v1', '/v1/rbac/auth/login', '/v1/unknown', '/api/v1/articles', '/uploads/a', '/assets/missing.js']) {
    assert.equal((await fetch(origin + path)).status, 404, path)
  }
  assert.equal((await fetch(origin + '/healthz')).status, 200)
  assert.equal((await fetch(origin + '/', { method: 'POST' })).status, 405)
})

test('Vite forwards only to one gateway and preserves separate host audiences', () => {
  assert.equal(gatewayProxy('site')['/v1'].headers.host, 'sitehub.local')
  const proxy = gatewayProxy('admin', { GATEWAY_URL: 'http://gateway.test:9080', RBAC_API_URL: 'http://ignored.test' })
  assert.equal(proxy['/v1'].target, 'http://gateway.test:9080')
  assert.equal(proxy['/v1'].headers.host, 'admin.sitehub.local')
  assert.throws(() => gatewayProxy('site', { GATEWAY_URL: 'http://secret:password@gateway' }))
  assert.throws(() => gatewayProxy('site', { GATEWAY_HOST: 'bad\r\nhost' }))
})
