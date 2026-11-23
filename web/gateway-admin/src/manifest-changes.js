export function sameValue(a, b) {
  if (a === b) return true
  if (!a || !b || typeof a !== 'object' || typeof b !== 'object') return false
  if (Array.isArray(a) !== Array.isArray(b)) return false
  const keys = Object.keys(a)
  return (
    keys.length === Object.keys(b).length &&
    keys.every((key) => Object.hasOwn(b, key) && sameValue(a[key], b[key]))
  )
}

function resources(doc) {
  const entries = []
  for (const kind of ['routes', 'streams']) {
    for (const row of doc[kind] || []) entries.push([kind + '/' + row.id, row])
  }
  for (const kind of ['upstreams', 'profiles']) {
    for (const [id, row] of Object.entries(doc[kind] || {})) entries.push([kind + '/' + id, row])
  }
  if (Object.keys(doc.plugins || {}).length) entries.push(['plugins/global', doc.plugins])
  return new Map(entries)
}

export function changedResources(base, draft) {
  const old = resources(base),
    local = resources(draft)
  const result = []
  for (const key of new Set([...old.keys(), ...local.keys()])) {
    if (!sameValue(old.get(key), local.get(key))) {
      result.push({ key, base: old.get(key), local: local.get(key) })
    }
  }
  return result
}

export function replaceResource(doc, key, value) {
  const [kind, id] = key.split('/')
  if (kind === 'plugins') {
    if (value === undefined) delete doc.plugins
    else doc.plugins = value
  } else if (kind === 'routes' || kind === 'streams') {
    const rows = doc[kind] || []
    const index = rows.findIndex((row) => row.id === id)
    if (value === undefined) {
      if (index >= 0) rows.splice(index, 1)
    } else if (index < 0) (doc[kind] ||= []).push(value)
    else rows[index] = value
  } else if (value === undefined) {
    if (doc[kind]) delete doc[kind][id]
  } else doc[kind] = { ...doc[kind], [id]: value }
}

// Reapply independently edited resources. Both writers touching the same
// resource requires an explicit decision before publication.
export function rebaseDraft(base, local, latest) {
  const merged = structuredClone(latest),
    remote = resources(latest),
    conflicts = []
  for (const change of changedResources(base, local)) {
    const current = remote.get(change.key)
    if (sameValue(current, change.local)) continue
    if (sameValue(current, change.base)) replaceResource(merged, change.key, change.local)
    else conflicts.push({ ...change, current, choice: '' })
  }
  return { merged, conflicts }
}

export function fieldChanges(before, after, path = '') {
  if (sameValue(before, after)) return []
  const record = (v) => v && typeof v === 'object' && !Array.isArray(v)
  if ((record(before) || before === undefined) && (record(after) || after === undefined)) {
    const keys = [...new Set([...Object.keys(before || {}), ...Object.keys(after || {})])].sort()
    if (keys.length)
      return keys.flatMap((key) =>
        fieldChanges(before?.[key], after?.[key], path ? path + '.' + key : key),
      )
  }
  return [{ path, before, after }]
}
