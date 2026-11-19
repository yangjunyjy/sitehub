import JSONBig from 'json-bigint'

// Go uint64 fields need JSON numbers while browsers must preserve all 64 bits.
export function uint64(value = 0) {
  if (typeof value === 'number' && !Number.isSafeInteger(value)) throw new Error('ID exceeds JavaScript integer precision')
  const text = String(value || 0)
  if (!/^\d+$/.test(text) || BigInt(text) > 18446744073709551615n) throw new Error('Invalid numeric ID')
  return BigInt(text)
}
export const stringifyJSON = value => JSONBig.stringify(value)
export const queryString = values => {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(values || {})) {
    if (value == null || value === '') continue
    if (Array.isArray(value)) value.forEach(item => query.append(key, String(item)))
    else query.set(key, String(value))
  }
  return query.toString()
}

// Pagination, statuses and display counters are bounded UI numbers; resource
// IDs, versions and revisions remain strings throughout the frontend.
const numericFields = new Set(['total', 'page', 'pageSize', 'status', 'sort', 'sortOrder', 'size', 'maxSize', 'count'])
export function viewValues(value, key = '') {
  if (Array.isArray(value)) return value.map(item => viewValues(item, key))
  if (value && typeof value === 'object') return Object.fromEntries(Object.entries(value).map(([name, item]) => [name, viewValues(item, name)]))
  if (numericFields.has(key) && typeof value === 'string' && /^\d+$/.test(value) && Number.isSafeInteger(Number(value))) return Number(value)
  return value
}

