import { ElMessage } from 'element-plus'
import { createMemoryHistory, createRouter } from 'vue-router'
import { publicSite } from '@shared/navigation/public-site.mjs'
import request from '@/utils/request'
import { websiteURL } from './website'

const descriptors = import.meta.glob('../../../product/src/views/**/menu.json', { eager: true, import: 'default' })
const metadata = {}
const pages = {}
for (const [source, settings] of Object.entries(descriptors)) {
  const file = source.replace('../../../product/src/', './')
  metadata[file] = settings
  pages[file.replace(/menu\.json$/, 'index.vue')] = {}
}

// Resolve previews with the same page metadata and public menu tree as the product app.
export async function previewProductPage(page, query = {}) {
  const tab = window.open('about:blank', '_blank')
  if (!tab) {
    ElMessage.warning('请允许浏览器打开预览窗口')
    return
  }
  tab.opener = null
  try {
    const { data } = await request.get('/rbac/menus/public', { params: { system: 'product' } })
    const { routes } = publicSite(pages, metadata, data.items)
    const router = createRouter({ history: createMemoryHistory(), routes })
    tab.location.replace(websiteURL(router.resolve({ name: page, query }).href))
  } catch (error) {
    tab.close()
    ElMessage.error(error.message || '无法打开产品页面')
  }
}
