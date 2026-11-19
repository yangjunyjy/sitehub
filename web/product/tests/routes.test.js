import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, readdir, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, relative } from 'node:path';
import { createRouter, createMemoryHistory } from 'vue-router';
import { createPageRoutes, pagePath } from '../src/router/pages.js';

async function discover(root) {
  const files = await readdir(root, { recursive: true, withFileTypes: true });
  return Object.fromEntries(files.filter(file => file.isFile() && file.name === 'index.vue').map(file => {
    const filename = relative(root, join(file.parentPath, file.name)).replaceAll('\\', '/');
    return [`../views/${filename}`, { template: '<div>page</div>' }];
  }));
}

test('new page directories generate routes without a route map, and deleted pages retain the 404 layout', async () => {
  const root = await mkdtemp(join(tmpdir(), 'sitehub-product-pages-'));
  try {
    for (const dir of ['home', 'not-found', 'tools/detail', 'docs', 'tools/components/example']) {
      await mkdir(join(root, dir), { recursive: true });
      await writeFile(join(root, dir, 'index.vue'), '<template>page</template>');
    }
    const layout = { name: 'ProductLayout', template: '<header/><router-view/><footer/>' };
    const options = { '../views/not-found/route.js': {path:'/:pathMatch(.*)*'}, '../views/tools/detail/route.js': {path:'/tools/:id'}, '../views/docs/route.js': {path:'/docs/:id?'} };
    const router = createRouter({ history: createMemoryHistory(), routes: createPageRoutes(await discover(root), layout, [], options) });
    assert.equal(router.resolve('/').name, 'home');
    assert.equal(router.resolve('/tools/42').params.id, '42');
    assert.equal(router.resolve('/docs').name, 'docs');
    assert.equal(router.resolve('/docs/install').params.id, 'install');
    assert.equal(router.resolve('/tools/components/example').meta.notFound, true);
    await router.push('/missing/deeper');
    assert.equal(router.currentRoute.value.meta.notFound, true);
    assert.equal(router.currentRoute.value.matched[0].components.default, layout);
    assert.equal(router.currentRoute.value.matched.length, 2);
    await rm(join(root, 'tools'), { recursive: true });
    const afterRemoval = createRouter({ history: createMemoryHistory(), routes: createPageRoutes(await discover(root), layout, [], options) });
    const missing = afterRemoval.resolve('/tools/42');
    assert.equal(missing.meta.notFound, true);
    assert.equal(missing.matched[0].components.default, layout);
  } finally { await rm(root, { recursive: true, force: true }); }
});

test('only page index.vue files and valid dynamic segments create routes', () => {
  assert.equal(pagePath('../views/features/index.vue'), '/features');
  assert.equal(pagePath('../views/docs/index.vue', {path: '/docs/:id?'}), '/docs/:id?');
  assert.equal(pagePath('../views/components/button/index.vue'), null);
  assert.equal(pagePath('../views/Home.vue'), null);
});
