import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, readdir, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, relative } from 'node:path';
import { createRouter, createMemoryHistory } from 'vue-router';
import { publicSite } from '../../shared/navigation/public-site.mjs';

async function discover(root) {
  const files = await readdir(root, { recursive: true, withFileTypes: true });
  return Object.fromEntries(files.filter(file => file.isFile() && file.name === 'index.vue').map(file => {
    const filename = relative(root, join(file.parentPath, file.name)).replaceAll('\\', '/');
    return [`./views/${filename}`, { template: '<div>page</div>' }];
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
    const options = { './views/home/menu.json': {path:'/'}, './views/not-found/menu.json': {path:'/:pathMatch(.*)*'}, './views/tools/detail/menu.json': {path:'/tools/:id'}, './views/docs/menu.json': {path:'/docs/:id?'} };
    const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/', component: layout, children: publicSite(await discover(root), options, []).routes }] });
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
    const afterRemoval = createRouter({ history: createMemoryHistory(), routes: [{ path: '/', component: layout, children: publicSite(await discover(root), options, []).routes }] });
    const missing = afterRemoval.resolve('/tools/42');
    assert.equal(missing.meta.notFound, true);
    assert.equal(missing.matched[0].components.default, layout);
  } finally { await rm(root, { recursive: true, force: true }); }
});

test('renamed document route keeps its optional slug and named links follow the same menu', () => {
  const page = { template: '<main />' };
  const { routes, links } = publicSite({ './views/docs/index.vue': page, './views/components/button/index.vue': page },
    { './views/docs/menu.json': { params: '/:id?' } },
    [{ name: '支持', path: '/support', children: [{ name: '新手册', component: 'docs/index', path: 'manual' }] }]);
  const router = createRouter({history: createMemoryHistory(), routes});
  assert.equal(routes.length, 1);
  assert.equal(links[0].children[0].path, '/support/manual');
  assert.equal(router.resolve('/support/manual/install').params.id, 'install');
  assert.equal(router.resolve({name: 'docs', params: {id: 'intro'}}).path, '/support/manual/intro');
});
