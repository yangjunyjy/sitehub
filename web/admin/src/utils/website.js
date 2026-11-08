export const clone = (value) => JSON.parse(JSON.stringify(value))
export function websiteURL(path = '/') {
  const base =
    import.meta.env.VITE_WEBSITE_URL ||
    `${window.location.protocol}//${window.location.hostname}:4310`
  return `${base.replace(/\/$/, '')}${path}`
}
export const previewWebsite = (path) =>
  window.open(websiteURL(path), '_blank', 'noopener,noreferrer')
