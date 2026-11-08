export const normalizeDownloadURL = (value = '') => value.trim()
export const isLocalReleaseURL = value => normalizeDownloadURL(value).startsWith('/v1/file/local/releases/')
export function validDownloadName(value = '') {
  return !value || (value === value.trim() && value !== '.' && value !== '..' && new TextEncoder().encode(value).length <= 255 && !/[\\/:*?"<>|\u0000-\u001f\u007f]/.test(value))
}
export const downloadHref = item => '/v1/file/download?' + new URLSearchParams({ id: item.id })
