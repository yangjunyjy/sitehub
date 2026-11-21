import test from "node:test";
import assert from "node:assert/strict";
import { publicSite } from '../../shared/navigation/public-site.mjs';
import { readFileSync } from 'node:fs';

test('page-local metadata declares article parameters and the 404 route', () => {
  const pages = {}, settings = {};
  for (const page of ['categories','article','not-found']) {
    pages['./views/' + page + '/index.vue'] = () => Promise.resolve({});
    settings['./views/' + page + '/menu.json'] = JSON.parse(readFileSync(new URL('../src/views/' + page + '/menu.json', import.meta.url)));
  }
  const { routes } = publicSite(pages, settings, []);
  assert.deepEqual(routes.map(({path}) => path), ['/categories','/article/:id','/:pathMatch(.*)*']);
});
