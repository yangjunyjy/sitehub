export const releaseStates = { published: '已发布', draft: '草稿', archived: '已下架' }

export function normalizePackage(item, defaults = {}) {
  return {
    id: item.id || '',
    groupId: item.groupId || '',
    edition: (item.edition || '').trim(),
    sortOrder: Number(item.sortOrder) || 0,
    product: item.product || defaults.product || 'Deploy',
    platform: item.platform || '',
    arch: item.arch || '',
    version: item.version || defaults.version || '',
    format: item.format || '',
    url: normalizeDownloadURL(item.url || ''),
    checksum: item.checksum || '',
    icon: item.icon || 'PKG',
    description: item.description || '',
    fileName: item.fileName || '',
    size: Number(item.size) || 0,
    releaseDate: item.releaseDate || defaults.releaseDate || '',
    status: item.status || 'draft',
  }
}

export function filterPackages(packages, filters) {
  const keyword = (filters.keyword || '').trim().toLowerCase()
  return packages.filter(
    (item) =>
      (!filters.product || item.product === filters.product) &&
      (!filters.groupId || item.groupId === filters.groupId) &&
      (!filters.edition || item.edition === filters.edition) &&
      (!filters.platform || item.platform === filters.platform) &&
      (!filters.status || item.status === filters.status) &&
      (!keyword ||
        [item.product, item.edition, item.platform, item.version, item.arch, item.fileName, item.description]
          .join(' ')
          .toLowerCase()
          .includes(keyword)),
  )
}

export function formatFileSize(bytes) {
  if (!bytes) return '—'
  const unit = Math.min(3, Math.floor(Math.log(bytes) / Math.log(1024)))
  return `${(bytes / 1024 ** unit).toLocaleString('zh-CN', { maximumFractionDigits: 1 })} ${['B', 'KB', 'MB', 'GB'][unit]}`
}

export function isDownloadURL(value) {
  if (!value || /[\\\s]/.test(value)) return false
  if (value.startsWith('/')) return !value.startsWith('//')
  try {
    const url = new URL(value)
    return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password
  } catch {
    return false
  }
}

export function packageError(item) {
  if (item.edition.length > 80 || !Number.isInteger(item.sortOrder) || item.sortOrder < 0 || item.sortOrder > 99999)
    return '版本类型不能超过 80 字，排序应为 0 到 99999 的整数'
  if (!item.product.trim() || !item.platform.trim() || !item.version.trim())
    return '请填写产品名称、版本号和平台'
  if (!releaseStates[item.status]) return '请选择有效的发布状态'
  if (!validDownloadName(item.fileName)) return '下载文件名不能包含路径、特殊符号或超过 255 字节'
  if (item.url && !isDownloadURL(item.url)) return '请输入 HTTP、HTTPS 下载地址或站内绝对路径'
  if (item.status === 'published' && !item.url) return '请上传安装包或填写下载地址后再发布'
  if (item.checksum && !/^[a-f\d]{64}$/i.test(item.checksum))
    return 'SHA256 必须为 64 位十六进制字符'
  return ''
}

export { websiteURL } from './website.js'
import { normalizeDownloadURL, validDownloadName } from './download-links.mjs'
export { isLocalReleaseURL, normalizeDownloadURL } from './download-links.mjs'
