// Absolute menu routes can live outside /dashboard while still using its layout.
export const workspacePath = (value) =>
  (typeof value === 'string' ? value : value.path).replace(/\/+$/, '') || '/'

export function isWorkspaceRoute(route) {
  return (
    Boolean(route.meta?.title) && !route.meta?.hidden && !route.meta?.missingComponent &&
    (route.matched?.some((record) => record.name === 'dashboard' || record.path === '/dashboard') ||
      route.path.startsWith('/dashboard/'))
  )
}
