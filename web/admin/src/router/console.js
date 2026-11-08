// Native console navigation is deployed with the frontend; authorization remains in RPC.
export const consoleRoutes = [
  { path: 'statistics/content', name: 'ContentStatistics', component: () => import('../views/statistics/Content.vue'), meta: { title: '内容统计', section: '内容管理', permission: 'blog:article:list' } },
  { path: 'articles', name: 'Articles', component: () => import('../views/articles/ArticleList.vue'), meta: { title: '文章管理', section: '内容管理', permission: 'blog:article:list' } },
  { path: 'categories', name: 'Categories', component: () => import('../views/categories/CategoryList.vue'), meta: { title: '分类管理', section: '内容管理', permission: 'blog:category:list' } },
  { path: 'tags', name: 'Tags', component: () => import('../views/tags/TagList.vue'), meta: { title: '标签管理', section: '内容管理', permission: 'blog:tag:list' } },
  { path: 'users', name: 'Users', component: () => import('../views/users/UserList.vue'), meta: { title: '用户管理', section: '权限管理', permission: 'system:user:list' } },
  { path: 'roles', name: 'Roles', component: () => import('../views/roles/RoleList.vue'), meta: { title: '角色管理', section: '权限管理', permission: 'system:role:list' } },
  { path: 'menus', name: 'Menus', component: () => import('../views/menus/MenuList.vue'), meta: { title: '菜单管理', section: '权限管理', permission: 'system:menu:list' } },

  {
    path: 'docs/modes',
    name: 'DocumentModes',
    component: () => import('../views/docs/DocumentModes.vue'),
    meta: { title: '文档模式', section: '产品文档' },
  },
  {
    path: 'docs',
    name: 'DocumentList',
    component: () => import('../views/docs/DocumentList.vue'),
    meta: { title: '文档列表', section: '产品文档' },
  },
  {
    path: 'docs/groups',
    name: 'DocumentGroups',
    component: () => import('../views/docs/DocumentGroups.vue'),
    meta: { title: '分组管理', section: '产品文档' },
  },
  {
    path: 'docs/editor',
    name: 'DocumentEditor',
    component: () => import('../views/docs/DocumentEditor.vue'),
    meta: { title: '文档编辑', section: '产品文档', activeMenu: '/dashboard/docs' },
  },
  {
    path: 'docs/settings',
    name: 'DocumentSettings',
    component: () => import('../views/docs/DocumentSettings.vue'),
    meta: { title: '文档站设置', section: '产品文档' },
  },
  {
    path: 'docs/versions',
    name: 'DocumentVersions',
    component: () => import('../views/docs/DocumentVersions.vue'),
    meta: { title: '版本管理', section: '产品文档' },
  },
  { path: 'settings', redirect: '/dashboard/settings/basic' },
  {
    path: 'settings/basic',
    name: 'BasicSettings',
    component: () => import('../views/settings/BasicSettings.vue'),
    meta: { title: '基本设置', section: '系统设置' },
  },
  {
    path: 'settings/storage',
    name: 'StorageSettings',
    component: () => import('../views/settings/StorageSettings.vue'),
    meta: { title: '文件存储', section: '系统设置' },
  },
  {
    path: 'settings/email',
    name: 'EmailSettings',
    component: () => import('../views/settings/EmailSettings.vue'),
    meta: { title: '邮件服务', section: '系统设置', superAdmin: true },
  },
]
export const isSuperAdmin = (user) =>
  (user?.roleIds || []).some((id) => Number(id) === 1) ||
  (user?.roles || []).some((role) => Number(role.id) === 1)
