<template>
  <main class="dashboard">
    <header class="workspace-heading">
      <div>
        <h1>工作台</h1>
        <p>{{ displayName }} · {{ today }}</p>
      </div>
      <el-button :loading="loading" :icon="Refresh" @click="load">刷新数据</el-button>
    </header>
    <el-alert
      v-if="failed.length"
      :title="`${failed.join('、')}暂时无法加载，请刷新重试。`"
      type="warning"
      :closable="false"
      show-icon
      class="load-alert"
    />
    <section class="metrics-grid" aria-label="工作台核心统计">
      <article
        v-for="card in metrics"
        :key="card.label"
        class="metric-card"
        :class="'metric-card--' + card.tone"
      >
        <span
          >{{ card.label
          }}<el-icon class="metric-icon"><component :is="card.icon" /></el-icon></span
        ><strong>{{ formatNumber(card.value) }}</strong
        ><small>{{ card.note }}</small>
      </article>
    </section>
    <div class="workspace-grid">
      <section class="workspace-main">
        <article class="panel trend-panel">
          <header>
            <h2>内容与访问趋势</h2>
            <el-radio-group v-model="days" size="small" @change="loadTrend"
              ><el-radio-button :value="7">近7天</el-radio-button
              ><el-radio-button :value="30">近30天</el-radio-button
              ><el-radio-button :value="90">近90天</el-radio-button></el-radio-group
            >
          </header>
          <div class="chart-wrap" v-loading="trendLoading">
            <div ref="chartRef" class="chart" role="img" aria-label="文章访问与新增文章趋势图" />
            <p v-if="trendError" class="chart-message">
              趋势加载失败 <el-button link @click="loadTrend">重试</el-button>
            </p>
            <p v-else-if="!trend && !trendLoading" class="chart-message">暂无趋势数据</p>
          </div>
        </article>
        <article class="panel latest-panel">
          <header>
            <h2>最近文章</h2>
            <el-button link type="primary" @click="go(router.page('blog/articles/index'))">全部文章</el-button>
          </header>
          <el-table
            :data="articles"
            empty-text="暂无文章，点击发布文章开始创作"
            class="article-table"
            ><el-table-column label="文章标题" min-width="230"
              ><template #default="{ row }"
                ><div class="article-title">
                  <el-icon><Document /></el-icon
                  ><button @click="go(router.page('blog/articles/editor/index', { params: { id: row.id } }))">{{ row.title }}</button>
                </div></template
              ></el-table-column
            ><el-table-column label="状态" width="95"
              ><template #default="{ row }"
                ><el-tag
                  :type="row.status === 1 ? 'success' : row.status === 2 ? 'warning' : 'info'"
                  size="small"
                  effect="light"
                  >{{
                    row.status === 1 ? '已发布' : row.status === 2 ? '定时发布' : '草稿'
                  }}</el-tag
                ></template
              ></el-table-column
            ><el-table-column label="更新时间" width="115"
              ><template #default="{ row }">{{
                formatDate(row.updated_at || row.created_at)
              }}</template></el-table-column
            ><el-table-column label="操作" width="65"
              ><template #default="{ row }"
                ><el-button link type="primary" @click="go(router.page('blog/articles/editor/index', { params: { id: row.id } }))"
                  >编辑</el-button
                ></template
              ></el-table-column
            ></el-table
          >
        </article>
      </section>
      <aside class="workspace-aside">
        <article class="panel">
          <header><h2>常用操作</h2></header>
          <div class="quick-actions">
            <button v-for="item in workspaceActions" :key="item.label" @click="go(item.path)">
              <el-icon><component :is="item.icon" /></el-icon><span>{{ item.label }}</span>
            </button>
          </div>
        </article>
        <article class="panel">
          <header>
            <h2>待处理</h2>
            <span class="count-badge">{{ pendingTasks.length }} 项</span>
          </header>
          <div v-if="pendingTasks.length" class="pending-tasks">
            <button v-for="task in pendingTasks" :key="task.label" @click="go(task.path)">
              <span
                >{{ task.label }}<small>{{ task.note }}</small></span
              ><b>{{ task.count }}</b>
            </button>
          </div>
          <p v-else class="quiet-state">{{ loading ? '正在读取…' : '暂无待处理内容' }}</p>
        </article>
        <article class="panel">
          <header><h2>内容概况</h2></header>
          <dl class="content-overview">
            <div v-for="item in articleStates" :key="item.name">
              <dt>{{ item.name }}</dt>
              <dd>{{ formatNumber(item.count) }}</dd>
            </div>
            <div>
              <dt>内容分类</dt>
              <dd>{{ formatNumber(stats?.content?.categories) }}</dd>
            </div>
            <div>
              <dt>内容标签</dt>
              <dd>{{ formatNumber(stats?.content?.tags) }}</dd>
            </div>
            <div>
              <dt>产品文档</dt>
              <dd>{{ documents ? documentCount : '—' }}</dd>
            </div>
            <div>
              <dt>已发布文件</dt>
              <dd>{{ website ? publishedPackages : '—' }}</dd>
            </div>
          </dl>
        </article>
      </aside>
    </div>
  </main>
</template>
<script setup>
import {
  computed,
  ref,
  onMounted,
  onBeforeUnmount,
  onActivated,
  onDeactivated,
  nextTick,
} from 'vue'
import { useRouter } from 'vue-router'
import { useRbacStore } from '@/store/rbac'
import { navigationMenus, isSuperAdmin } from '@/router/pages'
import { getDashboardStats, getArticleStats, getViewTrend, getArticleTrend } from '@/api/statistics'
import { getArticleList } from '@/api/article'
import { getDocumentWorkspace } from '@/api/documents'
import { getWebsiteDownloads } from '@/api/website'
import { normalizePackage } from '@/utils/downloads'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import {
  Document,
  View,
  EditPen,
  Download,
  Collection,
  FolderOpened,
  PriceTag,
  Refresh,
  Setting,
  DataAnalysis,
} from '@element-plus/icons-vue'
echarts.use([LineChart, GridComponent, TooltipComponent, LegendComponent, CanvasRenderer])
defineOptions({ name: 'Dashboard' })
const router = useRouter(),
  rbac = useRbacStore()
const stats = ref(null),
  articleStats = ref(null),
  website = ref(null),
  documents = ref(null),
  articles = ref([])
const loading = ref(false),
  failed = ref([]),
  chartRef = ref(null),
  days = ref(7),
  trendLoading = ref(false),
  trendError = ref(false),
  trend = ref(null)
let chart,
  observer,
  trendRequest = 0,
  disposed = false,
  activated = false
const displayName = computed(
  () => rbac.currentUser?.nickname || rbac.currentUser?.username || '管理员',
)
const today = new Date().toLocaleDateString('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'long',
})
const packages = computed(() =>
  (website.value?.downloads?.packages || []).map((item) =>
    normalizePackage(item, website.value.downloads),
  ),
)
const platforms = computed(() => [...new Set(packages.value.map((item) => item.platform))])
const publishedPackages = computed(
  () => packages.value.filter((item) => item.status === 'published').length,
)
const draftPackages = computed(
  () => packages.value.filter((item) => item.status === 'draft').length,
)
const documentCount = computed(() =>
  (documents.value?.versions?.flatMap((version) => version.groups || []) || []).reduce(
    (sum, group) => sum + (group.items?.length || 0),
    0,
  ),
)
const metrics = computed(() => [
  {
    label: '文章总数',
    value: articleStats.value?.total,
    note: `${articleStats.value?.published ?? '—'} 篇已发布`,
    icon: Document,
    tone: 'blue',
  },
  {
    label: '累计访问',
    value: stats.value?.overview?.views,
    note: '文章详情页访问次数',
    icon: View,
    tone: 'purple',
  },
  {
    label: '待发布草稿',
    value: articleStats.value?.draft,
    note: '未发布的文章',
    icon: EditPen,
    tone: 'amber',
  },
  {
    label: '下载版本',
    value: website.value ? packages.value.length : null,
    note: `${website.value ? publishedPackages.value : '—'} 个版本已发布`,
    icon: Download,
    tone: 'green',
  },
])
const articleStates = computed(() => [
  { name: '已发布', count: articleStats.value?.published, color: '#2681eb' },
  { name: '草稿', count: articleStats.value?.draft, color: '#00c49a' },
  {
    name: '其他待发布',
    count: articleStats.value
      ? Math.max(
          0,
          articleStats.value.total - articleStats.value.published - articleStats.value.draft,
        )
      : null,
    color: '#b2a0f9',
  },
])
const go = (path) => router.push(path)
const formatNumber = (value) => (value == null ? '—' : Number(value).toLocaleString('zh-CN'))
const formatDate = (value) => (value ? String(value).slice(0, 10) : '—')
const renderChart = () => {
  if (disposed || !chartRef.value || !trend.value) return
  chart ||= echarts.init(chartRef.value)
  chart.setOption(
    {
      animationDuration: 300,
      color: ['#2878ff', '#00c49a'],
      tooltip: { trigger: 'axis' },
      legend: {
        top: 4,
        icon: 'roundRect',
        itemWidth: 12,
        itemHeight: 5,
        textStyle: { color: '#8796aa', fontSize: 11 },
      },
      grid: { top: 39, left: 42, right: 27, bottom: 24 },
      xAxis: {
        type: 'category',
        data: trend.value.dates,
        boundaryGap: true,
        axisLine: { lineStyle: { color: '#e8eef6' } },
        axisTick: { show: false },
        axisLabel: { color: '#96a4b5', fontSize: 10 },
      },
      yAxis: [
        {
          type: 'value',
          minInterval: 1,
          axisLabel: { color: '#96a4b5', fontSize: 10 },
          splitLine: { lineStyle: { color: '#f0f3f8', type: 'dashed' } },
        },
        {
          type: 'value',
          minInterval: 1,
          splitLine: { show: false },
          axisLabel: { color: '#96a4b5', fontSize: 10 },
        },
      ],
      series: [
        {
          name: '文章访问',
          type: 'line',
          smooth: true,
          symbolSize: 5,
          data: trend.value.views,
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(38,129,235,.14)' },
              { offset: 1, color: 'rgba(38,129,235,0)' },
            ]),
          },
        },
        {
          name: '新增文章',
          type: 'line',
          smooth: true,
          symbolSize: 5,
          yAxisIndex: 1,
          barMaxWidth: 12,
          itemStyle: { borderRadius: [3, 3, 0, 0] },
          data: trend.value.articles,
        },
      ],
    },
    true,
  )
  chart.resize()
}
const loadTrend = async () => {
  const requestID = ++trendRequest
  trendLoading.value = true
  trendError.value = false
  try {
    const [views, counts] = await Promise.all([
      getViewTrend({ days: days.value }),
      getArticleTrend({ days: days.value }),
    ])
    if (requestID !== trendRequest || disposed) return
    trend.value = {
      dates: views.data.dates || [],
      views: views.data.views || [],
      articles: counts.data.counts || [],
    }
    await nextTick()
    renderChart()
  } catch {
    if (requestID === trendRequest && !disposed) {
      trendError.value = true
      trend.value = null
      chart?.clear()
    }
  } finally {
    if (requestID === trendRequest) trendLoading.value = false
  }
}
const load = async () => {
  if (loading.value) return
  loading.value = true
  failed.value = []
  const tasks = [
    [
      '工作台统计',
      getDashboardStats,
      (value) => {
        stats.value = value
      },
    ],
    [
      '发布统计',
      getArticleStats,
      (value) => {
        articleStats.value = value
      },
    ],
    [
      '下载清单',
      getWebsiteDownloads,
      (value) => {
        website.value = value
      },
    ],
    ['产品文档', getDocumentWorkspace, value => { documents.value = value }],
    [
      '最近文章',
      () => getArticleList({ page: 1, pageSize: 5 }),
      (value) => {
        articles.value = value.items || []
      },
    ],
  ]
  const results = await Promise.allSettled(
    tasks.map(async ([, fetch, apply]) => {
      const result = await fetch()
      if (!disposed) apply(result.data)
    }),
  )
  if (disposed) return
  failed.value = results.flatMap((result, index) =>
    result.status === 'rejected' ? [tasks[index][0]] : [],
  )
  loading.value = false
  await loadTrend()
}
onMounted(() => {
  load()
  observer = new ResizeObserver(() => chart?.resize())
  if (chartRef.value) observer.observe(chartRef.value)
})
onActivated(() => {
  if (activated) load()
  activated = true
  nextTick(() => chart?.resize())
})
onDeactivated(() => {
  trendRequest++
})
onBeforeUnmount(() => {
  disposed = true
  trendRequest++
  observer?.disconnect()
  chart?.dispose()
})

const workspaceActions = computed(() => {
  const actions = []
  const visit = (menus) => menus.forEach((menu) => {
    if (menu.type === 'menu' && menu.component !== 'dashboard/index' && (!menu.permission || rbac.hasPermission(menu.permission)))
      actions.push({ label: menu.title, path: menu.path, icon: menu.icon || 'Document' })
    visit(menu.children || [])
  })
  visit(navigationMenus(rbac.currentMenus, isSuperAdmin(rbac.currentUser)))
  return actions.slice(0, 6)
})
const pendingTasks = computed(() =>
  [
    {
      label: '完善文章草稿',
      note: '继续创作，准备发布',
      count: articleStats.value?.draft || 0,
      path: router.page('blog/articles/index', { query: { status: 0 } }),
    },
    {
      label: '发布下载版本',
      note: '检查文件后发布到官网',
      count: website.value ? draftPackages.value : 0,
      path: router.page('product/downloads/index'),
    },
    {
      label: '补充文档内容',
      note: '让产品文档更加完整',
      count: (documents.value?.versions?.flatMap((version) => version.groups || []) || [])
        .flatMap((group) => group.items || [])
        .filter((item) => !item.content?.trim()).length,
      path: router.page('product/docs/index'),
    },
  ].filter((item) => {
    const target = router.resolve(item.path)
    return item.count > 0 && String(target.name).startsWith('menu-') && (!target.meta.permission || rbac.hasPermission(target.meta.permission))
  }),
)
</script>
<style lang="scss" scoped>
.dashboard {
  padding: 16px 20px 24px;
  color: #303133;
}
.workspace-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
.workspace-heading h1 {
  font-size: 17px;
  font-weight: 500;
  margin: 0;
}
.workspace-heading p {
  font-size: 12px;
  color: #909399;
  margin: 5px 0 0;
}
.workspace-heading .el-button {
  height: 30px;
  font-size: 12px;
}
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}
.metric-card {
  padding: 20px;
  background: white;
  border-radius: 4px;
  min-width: 0;
  --metric-color: #409eff;
  --metric-soft: #ecf5ff;
}
.metric-card--purple {
  --metric-color: #a07bf2;
  --metric-soft: #f4f0ff;
}
.metric-card--cyan {
  --metric-color: #36cfc9;
  --metric-soft: #e6fffb;
}
.metric-card--green {
  --metric-color: #30c88b;
  --metric-soft: #eafaf4;
}
.metric-card > span {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  color: #606266;
}
.metric-icon {
  font-size: 20px;
  background: var(--metric-soft);
  color: var(--metric-color);
  width: 32px;
  height: 32px;
  border-radius: 5px;
}
.metric-card strong {
  display: block;
  margin: 16px 0 6px;
  font-size: 28px;
  font-weight: 500;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
}
.metric-card small {
  font-size: 12px;
  color: #909399;
}
.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 20px;
  align-items: start;
}
.workspace-main,
.workspace-aside {
  display: grid;
  gap: 20px;
  min-width: 0;
}
.panel {
  min-width: 0;
  padding: 20px;
  background: white;
  border-radius: 4px;
}
.panel > header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 22px;
}
.panel h2 {
  font-size: 15px;
  font-weight: 500;
  margin: 0;
}
.chart-wrap {
  position: relative;
}
.chart {
  height: 320px;
  width: 100%;
}
.chart-message {
  position: absolute;
  top: 40%;
  left: 0;
  width: 100%;
  text-align: center;
  font-size: 13px;
  color: #909399;
}
.article-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.article-title .el-icon {
  color: #909399;
  flex-shrink: 0;
}
.article-title button {
  border: 0;
  background: none;
  color: #606266;
  text-align: left;
  font: inherit;
  cursor: pointer;
}
.article-title button:hover {
  color: #409eff;
}
.article-table {
  --el-table-header-bg-color: #fafafa;
  --el-table-header-text-color: #909399;
  --el-table-border-color: #ebeef5;
  --el-table-text-color: #606266;
}
.article-table :deep(.el-table__cell) {
  padding: 12px 0;
}
.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px 10px;
}
.quick-actions button {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 12px 4px;
  border: 0;
  background: white;
  color: #606266;
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}
.quick-actions .el-icon {
  font-size: 22px;
  color: #409eff;
  background: #ecf5ff;
  width: 42px;
  height: 42px;
  border-radius: 6px;
}
.quick-actions button:hover {
  background: #f5f7fa;
  color: #409eff;
}
.count-badge {
  font-size: 12px;
  color: #909399;
}
.pending-tasks button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
  padding: 14px 0;
  border: 0;
  border-top: 1px solid #f0f0f0;
  background: transparent;
  text-align: left;
  font: inherit;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
}
.pending-tasks button:first-child {
  border-top: 0;
  padding-top: 0;
}
.pending-tasks small {
  display: block;
  margin-top: 6px;
  color: #909399;
  font-size: 12px;
}
.pending-tasks b {
  color: #409eff;
  font-size: 18px;
  font-weight: 500;
}
.content-overview {
  margin: 0;
}
.content-overview > div {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 9px 0;
  font-size: 13px;
}
.content-overview dt {
  color: #909399;
}
.content-overview dd {
  margin: 0;
}
.quiet-state {
  font-size: 13px;
  color: #909399;
}
.load-alert {
  margin-bottom: 16px;
}
@media (max-width: 1100px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }
  .workspace-aside {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .quick-actions {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 760px) {
  .dashboard {
    padding: 12px;
  }
  .metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
    margin-bottom: 12px;
  }
  .metric-card {
    padding: 16px;
  }
  .metric-card strong {
    font-size: 25px;
  }
  .metric-card > span {
    font-size: 12px;
  }
  .workspace-grid,
  .workspace-main,
  .workspace-aside {
    gap: 12px;
  }
  .workspace-aside {
    grid-template-columns: 1fr;
  }
  .quick-actions {
    grid-template-columns: repeat(3, 1fr);
  }
  .panel {
    padding: 16px;
  }
  .panel > header {
    flex-wrap: wrap;
  }
  .chart {
    height: 260px;
  }
}
</style>
