import test from "node:test";
import assert from "node:assert/strict";
import { publicSite } from '../../shared/navigation/public-site.mjs';

test('configured directories preserve order, links, icons and visibility', () => {
  const pages = { './views/tags/index.vue': {} };
  const menus = [{ name: '探索', path: '/explore', children: [
    { name: '话题', component: 'tags/index', path: 'topics', icon: 'custom:leaf' },
    { name: '重复', component: 'tags/index', path: 'topics' },
    { name: '友站', pageType: 'external', externalUrl: 'https://example.org/notes' },
    { name: '隐藏目录', isVisible: false, children: [{ name: '隐藏', component:'tags/index', path:'/hidden' }] },
    { name: '未知', component:'absent/index', path:'/absent' },
    { name: '无效外链', pageType: 'external', externalUrl: 'javascript:alert(1)' }
  ]}];
  const { routes, links } = publicSite(pages, {}, menus);
  assert.deepEqual(routes.map(r => r.path), ['/explore/topics']);
  assert.equal(links[0].label, '探索');
  assert.deepEqual(links[0].children.map(i => [i.label, i.path]), [
    ['话题','/explore/topics'], ['友站','https://example.org/notes']]);
  assert.equal(links[0].children[0].icon, 'custom:leaf');
});

test('menu URL changes apply together even when two pages exchange paths', () => {
  const pages = Object.fromEntries(['about','friends','login'].map(page => ['./views/' + page + '/index.vue', {}]));
  const menus = [
    { name: '友链', component: 'friends/index', path: '/about' },
    { name: '关于', component: 'about/index', path: '/friends' },
    { name: '占用登录', component: 'friends/index', path: '/login' },
  ];
  const { routes, links } = publicSite(pages, {}, menus);
  assert.deepEqual(routes.map(r => [r.name, r.path]), [['about','/friends'],['friends','/about'],['login','/login']]);
  assert.deepEqual(links.map(link => link.path), ['/about','/friends']);
});
