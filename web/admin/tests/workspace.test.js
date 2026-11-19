import test from 'node:test'
import assert from 'node:assert/strict'
import { createPinia, setActivePinia } from 'pinia'
import { useWorkspaceStore } from '../src/store/workspace.js'

function route(path, name, query = {}) {
  return {
    path,
    fullPath: path + (query.version ? '?version=' + query.version : ''),
    query,
    meta: { title: name },
    matched: [{ components: { default: { __name: name } } }],
  }
}

test('tabs and cached components stay synchronized while query state remains navigable', () => {
  setActivePinia(createPinia())
  const store = useWorkspaceStore()
  store.visit(route('/dashboard/system/users', 'UserList'))
  store.visit(route('/dashboard/product/docs', 'DocumentList', { version: 'v2' }))
  store.visit(route('/dashboard/blog/articles/editor/1', 'ArticleEditor'))
  assert.deepEqual(store.cachedNames, ['Dashboard', 'UserList'])
  assert.equal(
    store.views.find((view) => view.path === '/dashboard/product/docs').fullPath,
    '/dashboard/product/docs?version=v2',
  )
  store.remove('/dashboard/system/users')
  assert.deepEqual(store.cachedNames, ['Dashboard'])
  store.remove('/dashboard')
  assert(store.views.some((view) => view.affix))
})

test('refresh targets one cached page and never remounts an editor with unsaved data', async () => {
  setActivePinia(createPinia())
  const store = useWorkspaceStore()
  store.visit(route('/dashboard/system/users', 'UserList'))
  store.visit(route('/dashboard/blog/articles/editor/1', 'ArticleEditor'))
  await store.refresh('/dashboard/system/users')
  await store.refresh('/dashboard/blog/articles/editor/1')
  assert.deepEqual(store.revisions, { '/dashboard/system/users': 1 })
  assert(store.cachedNames.includes('UserList'))
  store.reset()
  assert.equal(store.views.length, 1)
  assert.deepEqual(store.revisions, {})
})

test('permission refresh removes revoked tabs and cached pages while preserving allowed work', async () => {
  setActivePinia(createPinia())
  const store = useWorkspaceStore()
  store.visit(route('/dashboard/system/users', 'UserList'))
  store.visit(route('/dashboard/blog/articles', 'ArticleList'))
  await store.refresh('/dashboard/system/users')
  store.retain((view) => view.path === '/dashboard/blog/articles')
  assert.deepEqual(store.cachedNames, ['Dashboard', 'ArticleList'])
  assert.equal(
    store.views.some((view) => view.path === '/dashboard/system/users'),
    false,
  )
  assert.deepEqual(store.revisions, {})
})

test('signing out during a pending refresh cannot restore the previous session workspace', async () => {
  setActivePinia(createPinia())
  const store = useWorkspaceStore()
  store.visit(route('/dashboard/system/users', 'UserList'))
  const refreshing = store.refresh('/dashboard/system/users')
  store.reset()
  await refreshing
  assert.deepEqual(store.cachedNames, ['Dashboard'])
  assert.deepEqual(store.revisions, {})
})

test('absolute database menu paths join the same workspace and trailing slashes do not duplicate tabs', () => {
  setActivePinia(createPinia())
  const store = useWorkspaceStore()
  const custom = route('/system/people', 'UserList')
  custom.matched.unshift({ name: 'dashboard', path: '/dashboard' })
  store.visit(custom)
  store.visit({ ...custom, path: '/system/people/', fullPath: '/system/people/?page=2' })
  assert.equal(store.views.length, 2)
  assert.equal(store.views[1].path, '/system/people')
  assert.equal(store.views[1].fullPath, '/system/people/?page=2')
  assert(store.cachedNames.includes('UserList'))
})

test('restored tabs resolve through current routes and permissions without trusting persisted metadata', () => {
  const previous = globalThis.sessionStorage
  const values = new Map([
    [
      'deploy-admin-tabs:1',
      JSON.stringify(['/dashboard/system/users', '/dashboard/system/roles', '//elsewhere.test', '/missing']),
    ],
  ])
  globalThis.sessionStorage = {
    getItem: (key) => values.get(key),
    setItem: (key, value) => values.set(key, value),
    removeItem: (key) => values.delete(key),
  }
  try {
    setActivePinia(createPinia())
    const store = useWorkspaceStore()
    const resolve = (path) =>
      path === '/missing'
        ? { path, meta: {}, matched: [] }
        : route(path, path.includes('users') ? 'UserList' : 'RoleList')
    store.restore(1, resolve, (item) => !item.path.endsWith('/roles'))
    assert.deepEqual(
      store.views.map((item) => item.path),
      ['/dashboard', '/dashboard/system/users'],
    )
    store.reset()
    assert.equal(values.has('deploy-admin-tabs:1'), false)
  } finally {
    if (previous === undefined) delete globalThis.sessionStorage
    else globalThis.sessionStorage = previous
  }
})
