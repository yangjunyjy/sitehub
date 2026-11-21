import { Connection, Operation, SetUp, Clock } from '@element-plus/icons-vue'
export const pages = [
  { id: 'routes', name: '路由', icon: Connection, group: '流量管理' },
  { id: 'upstreams', name: '上游服务', icon: SetUp, group: '流量管理' },
  { id: 'policies', name: '插件策略', icon: Operation, group: '流量管理' },
  { id: 'releases', name: '发布记录', icon: Clock, group: '变更管理' },
]
export function currentPage() {
  return pages.find((p) => p.id === location.hash.slice(1)) || pages[0]
}
