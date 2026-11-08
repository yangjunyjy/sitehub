export const defaultTableAppearance = Object.freeze({
  stripe: false,
  border: false,
  headerBackground: false,
})

export function columnGroup(column) {
  if (column.locked) return 0
  if (column.fixed === true || column.fixed === 'left') return 1
  return column.fixed === 'right' ? 3 : 2
}

// Persist keys and presentation only; never persist row data, labels or permissions.
export function normalizeTablePreferences(columns, stored) {
  const saved = stored && typeof stored === 'object' && stored.version === 1 ? stored : {}
  const byKey = new Map(columns.map((column) => [column.key, column]))
  const savedOrder = Array.isArray(saved.order) ? saved.order : []
  const order = [...new Set([...savedOrder.filter((key) => byKey.has(key)), ...byKey.keys()])].sort(
    (a, b) => {
      const group = columnGroup(byKey.get(a)) - columnGroup(byKey.get(b))
      return (
        group ||
        (byKey.get(a).locked ? columns.indexOf(byKey.get(a)) - columns.indexOf(byKey.get(b)) : 0)
      )
    },
  )
  const hidden = new Set(
    Array.isArray(saved.hidden)
      ? saved.hidden.filter((key) => byKey.has(key) && !byKey.get(key).locked)
      : [],
  )
  const dataColumns = columns.filter(
    (column) => !column.utility && !['selection', 'index', 'expand'].includes(column.type),
  )
  if (dataColumns.length && dataColumns.every((column) => hidden.has(column.key)))
    hidden.delete(dataColumns[0].key)
  const appearance = Object.fromEntries(
    Object.keys(defaultTableAppearance).map((key) => [key, saved.appearance?.[key] === true]),
  )
  return {
    version: 1,
    order,
    hidden: [...hidden],
    appearance,
    density: ['comfortable', 'default', 'compact'].includes(saved.density)
      ? saved.density
      : 'default',
  }
}

export function moveTableColumn(columns, order, source, target) {
  const from = columns.find((column) => column.key === source),
    to = columns.find((column) => column.key === target)
  if (!from || !to || from.locked || to.locked || columnGroup(from) !== columnGroup(to))
    return order
  const fromIndex = order.indexOf(source),
    toIndex = order.indexOf(target)
  if (fromIndex < 0 || toIndex < 0 || fromIndex === toIndex) return order
  const moved = [...order]
  moved.splice(fromIndex, 1)
  moved.splice(toIndex, 0, source)
  return moved
}

export function tablePreferenceStorageKey(pageKey, userID) {
  return pageKey
    ? `deploy-admin:table:v1:${encodeURIComponent(String(userID || 'anonymous'))}:${encodeURIComponent(pageKey)}`
    : ''
}
