import { isSuperAdmin } from './console'

const pages = [
  ['index', '工作台', 'Dashboard', '', 'House'],
  ['statistics/content', '内容统计', 'statistics/Content', 'blog:article:list', 'DataAnalysis'],
  ['articles', '文章管理', 'articles/ArticleList', 'blog:article:list', 'Document'],
  ['categories', '分类管理', 'categories/CategoryList', 'blog:category:list', 'Folder'],
  ['tags', '标签管理', 'tags/TagList', 'blog:tag:list', 'PriceTag'],
  ['docs', '产品文档', 'docs/DocumentList', 'website:docs:edit', 'Reading'],
  ['docs/modes', '文档模式', 'docs/DocumentModes', 'website:docs:edit', 'Collection'],
  ['docs/versions', '文档版本', 'docs/DocumentVersions', 'website:docs:edit', 'Files'],
  ['docs/groups', '文档分组', 'docs/DocumentGroups', 'website:docs:edit', 'FolderOpened'],
  ['docs/settings', '文档设置', 'docs/DocumentSettings', 'website:docs:edit', 'Setting'],
  ['website', '官网内容', 'website/ContentCenter', 'website:content:edit', 'Monitor'],
  ['downloads', '下载管理', 'website/DownloadCenter', 'website:downloads:edit', 'Download'],
  ['users', '用户管理', 'users/UserList', 'system:user:list', 'User'],
  ['roles', '角色管理', 'roles/RoleList', 'system:role:list', 'UserFilled'],
  ['menus', '菜单管理', 'menus/MenuList', 'system:menu:list', 'Menu'],
  ['settings/storage', '文件存储', 'settings/StorageSettings', 'website:storage:edit', 'Folder'],
  ['settings/email', '邮件服务', 'settings/EmailSettings', 'super-admin', 'Message'],
  ['profile', '个人中心', 'profile/Profile', '', 'User'],
]

// Built-in navigation is a frontend capability catalog. Every operation is still
// authorized by RPC; database menus can add or rename routes without seed aliases.
export function nativeNavigation(user, permissions = [], assigned = []) {
  const admin = isSuperAdmin(user)
  const knownPaths = new Set()
  const visit = items => items.forEach(item => { knownPaths.add(item.path); visit(item.children || []) })
  visit(assigned)
  const builtins = pages.filter(([path, , , permission]) =>
    !knownPaths.has('/dashboard/' + path) && (admin || !permission || permissions.includes(permission) || permissions.includes('*')),
  ).map(([path, title, component, permission, icon], index) => ({
    id: 'navigation-' + path, title, path: '/dashboard/' + path,
    component: `views/${component}.vue`, permission, icon,
    type: 'menu', status: 1, isVisible: true, sort: index, children: [],
  }))
  return [...builtins, ...assigned]
}
