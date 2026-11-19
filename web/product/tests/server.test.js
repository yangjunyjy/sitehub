import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { once } from 'node:events';
import { createFrontendServer } from '../server/server.mjs';
import { gatewayProxy } from '../server/dev-proxy.mjs';

test('product serves SPA routes with its own runtime and never proxies APIs', async () => {
  const dist = await mkdtemp(join(tmpdir(), 'sitehub-product-server-'));
  await writeFile(join(dist, 'index.html'), '<main>product shell</main>');
  const server = createFrontendServer({ dist, productURL: 'https://product.example.test', blogURL: 'https://blog.example.test' });
  try {
    server.listen(0, '127.0.0.1');
    await once(server, 'listening');
    const base = `http://127.0.0.1:${server.address().port}`;
    assert.deepEqual(await (await fetch(base + '/healthz')).json(), { status: 'ok', app: 'product' });
    assert.match(await (await fetch(base + '/docs/install', { headers: { Accept: 'text/html' } })).text(), /product shell/);
    assert.match(await (await fetch(base + '/missing', { headers: { Accept: 'text/html' } })).text(), /product shell/);
    assert.match(await (await fetch(base + '/docs/release.v2', { headers: { Accept: 'text/html' } })).text(), /product shell/);
    assert.equal((await fetch(base + '/assets/removed.js')).status, 404);
    assert.equal((await fetch(base + '/v1/product/content')).status, 404);
    const runtime = await fetch(base + '/sitehub-runtime.js');
    assert.equal(runtime.headers.get('cache-control'), 'no-store');
    assert.match(await runtime.text(), /"productURL":"https:\/\/product.example.test"/);
  } finally { await new Promise(resolve => server.close(resolve)); await rm(dist, { recursive: true, force: true }); }
});

test('the local development proxy selects product independently', () => {
  assert.equal(gatewayProxy({})['/v1'].headers.host, 'product.sitehub.local');
  assert.equal(gatewayProxy({ GATEWAY_HOST: 'custom.test' })['/v1'].headers.host, 'custom.test');
  assert.throws(() => gatewayProxy({ GATEWAY_URL: 'http://user:password@localhost' }), /Invalid GATEWAY_URL/);
});
