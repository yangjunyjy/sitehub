export const clone = (value) => JSON.parse(JSON.stringify(value))
export function websiteURL(path = '/') {
  const base = window.__SITEHUB_CONFIG__?.websiteURL || window.location.origin
  return new URL(path, base).href
}
export const previewWebsite = (path) =>
  window.open(websiteURL(path), '_blank', 'noopener,noreferrer')
