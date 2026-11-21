import { onBeforeUnmount, onMounted, ref } from 'vue'
import { readDownloads } from '@/api/downloads'
import { useMember } from '@/composables/useMember'

export function useReleaseCatalog(route, router) {
  const { member, restoreMember } = useMember()
  const downloads = ref({ packages: [], groups: [] })
  const loading = ref(true), error = ref('')
  const downloadError = ref(''), downloadNotice = ref(''), checkingPackage = ref('')
  let controller
  let disposed = false

  async function downloadPackage(item) {
    if (checkingPackage.value || disposed) return
    checkingPackage.value = item.id
    downloadError.value = ''
    downloadNotice.value = ''
    try {
      if (!member.ready || member.user) await restoreMember()
      if (disposed) return
      if (!member.user) {
        const redirect = router.resolve({ path: route.path, query: { ...route.query, package: item.id } }).fullPath
        await router.push({ name: 'login', query: { redirect } })
        return
      }

      // The manifest supplies the member-protected URL. Let the browser stream large assets.
      const target = new URL(item.url, window.location.origin)
      if (target.origin !== window.location.origin || target.pathname !== '/v1/file/download' || target.searchParams.get('id') !== String(item.id))
        throw new Error('安装包下载地址无效，请重新加载发行目录。')
      const link = document.createElement('a')
      link.href = target.href
      link.download = item.fileName || ''
      document.body.appendChild(link)
      link.click()
      link.remove()
      downloadNotice.value = '已请求下载，可在浏览器下载列表中查看进度。'
    } catch (failure) {
      if (!disposed) downloadError.value = failure.message || '暂时无法下载，请稍后重试。'
    } finally {
      checkingPackage.value = ''
    }
  }

  async function load() {
    controller?.abort()
    const request = new AbortController()
    controller = request
    loading.value = true
    error.value = ''
    try {
      const catalog = await readDownloads({ signal: request.signal })
      if (request.signal.aborted) return
      downloads.value = catalog
      const packageID = route.query.package
      if (typeof packageID !== 'string') return
      const query = { ...route.query }
      delete query.package
      await router.replace({ path: route.path, query, hash: '#packages' })
      if (disposed || request.signal.aborted) return
      const selected = catalog.packages.find(item => item.id === packageID && item.url)
      if (selected) await downloadPackage(selected)
      else downloadError.value = '所选安装包不存在或已下架，请选择其他版本。'
    } catch (failure) {
      if (!request.signal.aborted) error.value = failure.message || '发行目录加载失败'
    } finally {
      if (!request.signal.aborted) loading.value = false
    }
  }

  onMounted(load)
  onBeforeUnmount(() => { disposed = true; controller?.abort() })
  return { member, downloads, loading, error, downloadError, downloadNotice, checkingPackage, load, downloadPackage }
}
