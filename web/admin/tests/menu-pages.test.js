import test from 'node:test'
import assert from 'node:assert/strict'
import { filterMenuTree, navigationMenus, normalizeComponent } from '../src/router/pages.js'

test('legacy system directory retains supported RBAC pages and removes retired pages', () => {
  const source = [
    {
      id: 1,
      type: 'directory',
      path: '/system',
      status: 1,
      isVisible: true,
      children: [
        { id: 2, type: 'menu', component: 'views/users/UserList', isVisible: true, status: 1 },
        { id: 3, type: 'menu', component: 'views/system/index', isVisible: true, status: 1 },
        { id: 4, type: 'menu', component: 'views/audit/ArticleAudit', isVisible: true, status: 1 },
      ],
    },
  ]
  const before = structuredClone(source)
  const result = filterMenuTree(source, true)
  assert.deepEqual(
    result[0].children.map((menu) => menu.id),
    [2],
  )
  assert.deepEqual(source, before)
})

test('persisted navigation keeps edited names, paths, hierarchy and visibility', () => {
  const source = [
    {
      id: 55,
      title: '站点管理',
      type: 'directory',
      path: '/site',
      isVisible: true,
      children: [
        {
          id: 43,
          title: '自定义官网',
          type: 'menu',
          component: 'views/website/ContentCenter',
          path: 'website',
          sort: 2,
          isVisible: true,
        },
        {
          id: 60,
          title: '下载管理',
          type: 'menu',
          component: 'views/website/DownloadCenter',
          path: '/site/downloads',
          sort: 1,
          isVisible: true,
        },
        {
          id: 61,
          title: '隐藏下载',
          type: 'menu',
          component: 'views/website/DownloadCenter',
          path: '/hidden',
          isVisible: false,
        },
        {
          id: 62,
          title: '邮件服务',
          type: 'menu',
          component: 'views/settings/EmailSettings',
          path: '/custom/email',
          isVisible: true,
        },
      ],
    },
  ]
  const before = structuredClone(source)
  const result = navigationMenus(source)
  assert.deepEqual(
    result[0].children.map((menu) => [menu.title, menu.path]),
    [
      ['下载管理', '/site/downloads'],
      ['自定义官网', '/site/website'],
    ],
  )
  assert.equal(navigationMenus(source, true)[0].children.length, 3)
  assert.deepEqual(source, before)
  assert.equal(normalizeComponent('views/docs/DocumentVersions'), 'views/docs/DocumentVersions.vue')
})

test('hidden editors and buttons remain routable/authorized but stay out of sidebar', () => {
  const menus = [
    { id: 1, type: 'menu', component: 'articles/ArticleEditor.vue', isVisible: false, status: 1 },
    {
      id: 2,
      type: 'menu',
      component: 'views/articles/ArticleList',
      isVisible: true,
      status: 1,
      children: [
        {
          id: 3,
          type: 'button',
          permission: 'content:article:create',
          status: 1,
          isVisible: false,
        },
      ],
    },
    { id: 4, type: 'menu', component: 'views/tags/TagList', status: 0, isVisible: true },
  ]
  assert.equal(filterMenuTree(menus).length, 2)
  assert.equal(filterMenuTree(menus)[1].children[0].permission, 'content:article:create')
  assert.deepEqual(
    filterMenuTree(menus, true).map((menu) => menu.id),
    [2],
  )
  assert.deepEqual(filterMenuTree(menus, true)[0].children, [])
})

test('unknown components and empty directories do not create broken navigation', () => {
  assert.deepEqual(
    filterMenuTree([
      { type: 'directory', children: [] },
      { type: 'menu', component: null },
      { type: 'menu', component: 'views/host/Host' },
      { type: 'menu', component: 'views/statistics/Overview' },
    ]),
    [],
  )
  assert.equal(
    normalizeComponent('/src/views/articles/ArticleList'),
    'views/articles/ArticleList.vue',
  )
})
