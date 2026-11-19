<template>
  <div class="content-statistics" :aria-busy="loading">
    <el-card class="statistics-shell" shadow="never">
      <template #header>
        <div class="card-header">
          <span>内容统计</span>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            :shortcuts="dateShortcuts"
            @change="handleDateChange"
          />
        </div>
      </template>

      <el-alert
        v-if="loadError"
        :title="loadError"
        type="error"
        show-icon
        :closable="false"
        style="margin-bottom: 20px"
      >
        <el-button link type="primary" :loading="loading" @click="loadData">重新加载</el-button>
      </el-alert>

      <!-- 文章统计卡片 -->
      <el-row :gutter="20" class="stats-cards">
        <el-col :xs="12" :sm="12" :lg="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ loaded ? stats.total : '—' }}</div>
              <div class="stat-label">文章总数</div>
              <div class="stat-icon article-icon">
                <el-icon><Document /></el-icon>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="12" :lg="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ loaded ? stats.published : '—' }}</div>
              <div class="stat-label">已发布</div>
              <div class="stat-icon published-icon">
                <el-icon><CircleCheck /></el-icon>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="12" :lg="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ loaded ? stats.draft : '—' }}</div>
              <div class="stat-label">草稿</div>
              <div class="stat-icon draft-icon">
                <el-icon><EditPen /></el-icon>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="12" :lg="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-item">
              <div class="stat-value">{{ loaded ? stats.totalViews : '—' }}</div>
              <div class="stat-label">总浏览量</div>
              <div class="stat-icon views-icon">
                <el-icon><View /></el-icon>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 趋势图表 -->
      <el-row :gutter="20" class="charts-row">
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>文章发布趋势</template>
            <div ref="publishChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>浏览量趋势</template>
            <div ref="viewChartRef" style="height: 300px"></div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 热门文章 TOP 10 -->
      <el-card class="hot-articles-card">
        <template #header>热门文章 TOP 10</template>
        <el-table :data="hotArticles" style="width: 100%" v-loading="loading">
          <el-table-column type="index" label="#" width="50" align="center" />
          <el-table-column prop="title" label="标题" min-width="250" show-overflow-tooltip />
          <el-table-column label="浏览" width="90" align="center" sortable>
            <template #default="{ row }">{{ row.views ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="点赞" width="90" align="center" sortable>
            <template #default="{ row }">{{ row.likes ?? 0 }}</template>
          </el-table-column>
          <el-table-column label="评论" width="90" align="center" sortable>
            <template #default="{ row }">{{ row.commentCount ?? 0 }}</template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 分类和标签统计 -->
      <el-row :gutter="20" class="category-tag-row">
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>分类文章数</template>
            <div ref="categoryChartRef" style="height: 320px"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>标签使用率 (TOP 10)</template>
            <div ref="tagChartRef" style="height: 320px"></div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 热门标签 -->
      <el-card v-if="hotTags.length > 0" class="hot-tags-card">
        <template #header>热门标签</template>
        <div class="tag-cloud">
          <el-tag
            v-for="(tag, index) in hotTags"
            :key="tag.id"
            :type="['', 'success', 'warning', 'danger', 'info'][index % 5]"
            size="large"
            effect="dark"
            round
            class="tag-item"
            :style="{ fontSize: Math.max(13, 14 + tag.articleCount * 1.5) + 'px' }"
          >
            {{ tag.name }} ({{ tag.articleCount }})
          </el-tag>
        </div>
      </el-card>
    </el-card>
  </div>
</template>

<script setup>
defineOptions({ name: 'Content' })

import {
  ref,
  onMounted,
  onBeforeUnmount,
  onActivated,
  onDeactivated,
  reactive,
  nextTick,
} from 'vue'
import * as echarts from 'echarts/core'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
echarts.use([
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
])
import { Document, CircleCheck, EditPen, View, Star, ChatDotRound } from '@element-plus/icons-vue'
import {
  getArticleStats,
  getArticleTrend,
  getViewTrend,
  getHotArticles,
  getCategoryStats,
  getTagStats,
  getHotTags,
} from '@/api/statistics'

// 工具函数
const unwrap = (res) => (res?.data !== undefined ? res.data : null)
const isOk = (res) => res && res.code === 0

// 快捷日期选项
const dateShortcuts = [
  {
    text: '最近7天',
    value: (() => {
      const e = new Date()
      const s = new Date()
      s.setDate(s.getDate() - 7)
      return [s, e]
    })(),
  },
  {
    text: '最近30天',
    value: (() => {
      const e = new Date()
      const s = new Date()
      s.setDate(s.getDate() - 30)
      return [s, e]
    })(),
  },
  {
    text: '最近3个月',
    value: (() => {
      const e = new Date()
      const s = new Date()
      s.setMonth(s.getMonth() - 3)
      return [s, e]
    })(),
  },
]

const loading = ref(false)
const loadError = ref('')
const loaded = ref(false)
let generation = 0,
  activated = false,
  disposed = false,
  resizeFrame,
  resizeObserver
const dateRange = ref([])
const days = reactive({ value: 7 })

const stats = reactive({
  total: 0,
  published: 0,
  draft: 0,
  totalViews: 0,
  trend: [],
})

const hotArticles = ref([])
const categoryData = ref([])
const tagData = ref([])
const hotTags = ref([])

let publishChart = null
let viewChart = null
let categoryChart = null
let tagChart = null

const publishChartRef = ref(null)
const viewChartRef = ref(null)
const categoryChartRef = ref(null)
const tagChartRef = ref(null)

const calcDaysFromRange = () => {
  if (dateRange.value && dateRange.value.length === 2) {
    const diff = Math.ceil((new Date(dateRange.value[1]) - new Date(dateRange.value[0])) / 86400000)
    days.value = Math.max(diff, 1)
  } else days.value = 7
}

const loadData = async () => {
  const current = ++generation
  loading.value = true
  loadError.value = ''
  try {
    calcDaysFromRange()
    const requestOptions = { silentError: true }

    const [statsRes, hotRes, catRes, tagRes, hotTagRes, articleTrendRes, viewTrendRes] =
      await Promise.all([
        getArticleStats(undefined, requestOptions),
        getHotArticles({ limit: 10 }, requestOptions),
        getCategoryStats(requestOptions),
        getTagStats(requestOptions),
        getHotTags({ limit: 15 }, requestOptions),
        getArticleTrend({ days: days.value }, requestOptions),
        getViewTrend({ days: days.value }, requestOptions),
      ])

    if (disposed || current !== generation) return
    // 统计卡片
    if (isOk(statsRes)) {
      const d = unwrap(statsRes)
      if (d) Object.assign(stats, d)
    }

    // 热门文章
    if (isOk(hotRes)) {
      hotArticles.value = unwrap(hotRes) || []
    }

    // 分类/标签统计
    if (isOk(catRes)) {
      categoryData.value = Array.isArray(unwrap(catRes)) ? unwrap(catRes) : []
    }
    if (isOk(tagRes)) {
      tagData.value = Array.isArray(unwrap(tagRes)) ? unwrap(tagRes) : []
    }
    // 热门标签
    if (isOk(hotTagRes)) {
      hotTags.value = Array.isArray(unwrap(hotTagRes)) ? unwrap(hotTagRes) : []
    }

    loaded.value = true
    await nextTick()
    if (!disposed && current === generation) initCharts(articleTrendRes, viewTrendRes)
  } catch (error) {
    if (!disposed && current === generation)
      loadError.value = loaded.value
        ? '统计更新失败，当前显示上次成功的数据'
        : '统计加载失败，请重试'
  } finally {
    if (current === generation) loading.value = false
  }
}

const handleDateChange = () => {
  loadData()
}

const initCharts = (articleTrendRes, viewTrendRes) => {
  const trendDays = days.value || 7

  // ---- 发布趋势图 ----
  if (publishChartRef.value) {
    publishChart ||= echarts.init(publishChartRef.value)
    const at = unwrap(articleTrendRes) || {}
    const dates = at.dates || generateDates(trendDays)
    const counts = at.values || at.counts || []

    publishChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', boundaryGap: false, data: dates },
      yAxis: { type: 'value' },
      series: [
        {
          name: '发布数',
          type: 'line',
          smooth: true,
          data: counts.length > 0 ? counts : Array(trendDays).fill(0),
          itemStyle: { color: '#67c23a' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(103,194,58,0.25)' },
              { offset: 1, color: 'rgba(103,194,58,0.02)' },
            ]),
          },
        },
      ],
    })
  }

  // ---- 浏览量趋势图 ----
  if (viewChartRef.value) {
    viewChart ||= echarts.init(viewChartRef.value)
    const vt = unwrap(viewTrendRes) || {}
    const dates = vt.dates || generateDates(trendDays)
    const views = vt.values || vt.views || []

    viewChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', boundaryGap: false, data: dates },
      yAxis: { type: 'value' },
      series: [
        {
          name: '浏览量',
          type: 'bar',
          barWidth: '40%',
          data: views.length > 0 ? views : Array(trendDays).fill(0),
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: '#409eff' },
              { offset: 1, color: '#79bbff' },
            ]),
            borderRadius: [4, 4, 0, 0],
          },
        },
      ],
    })
  }

  // ---- 分类饼图 ----
  if (categoryChartRef.value) {
    categoryChart ||= echarts.init(categoryChartRef.value)
    const pieData = categoryData.value
      .filter((c) => c.articleCount > 0)
      .map((c) => ({ name: c.name, value: c.articleCount }))

    categoryChart.setOption({
      tooltip: { trigger: 'item', formatter: '{b}: {c}篇 ({d}%)' },
      legend: { orient: 'vertical', left: 'left', top: 'middle' },
      series: [
        {
          name: '分类文章数',
          type: 'pie',
          radius: ['35%', '65%'],
          center: ['58%', '52%'],
          avoidLabelOverlap: true,
          itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 2 },
          label: { show: true, formatter: '{b}\n{c}', fontSize: 11 },
          emphasis: { label: { show: true, fontWeight: 'bold', fontSize: 14 } },
          data:
            pieData.length > 0
              ? pieData
              : [{ value: 1, name: '暂无数据', itemStyle: { color: '#ddd' } }],
        },
      ],
    })
  }

  // ---- 标签柱状图 ----
  if (tagChartRef.value) {
    tagChart ||= echarts.init(tagChartRef.value)
    const top10 = tagData.value.slice(0, 10).sort((a, b) => a.articleCount - b.articleCount)

    tagChart.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '8%', bottom: '3%', containLabel: true },
      xAxis: { type: 'value' },
      yAxis: {
        type: 'category',
        data: top10.map((t) => t.name),
        axisLabel: { fontSize: 12 },
      },
      series: [
        {
          name: '使用次数',
          type: 'bar',
          data: top10.map((t) => t.articleCount || 0),
          barWidth: '60%',
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
              { offset: 0, color: '#409eff' },
              { offset: 1, color: '#79bbff' },
            ]),
            borderRadius: [0, 4, 4, 0],
          },
          label: { show: true, position: 'right', fontSize: 11, color: '#666' },
        },
      ],
    })
  }
}

const generateDates = (d) => {
  const arr = []
  for (let i = d - 1; i >= 0; i--) {
    const dt = new Date()
    dt.setDate(dt.getDate() - i)
    arr.push(
      `${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`,
    )
  }
  return arr
}

const handleResize = () => {
  if (resizeFrame) return
  resizeFrame = requestAnimationFrame(() => {
    resizeFrame = undefined
    if (!publishChartRef.value?.getClientRects().length) return
    publishChart?.resize()
    viewChart?.resize()
    categoryChart?.resize()
    tagChart?.resize()
  })
}
const disposeCharts = () => {
  publishChart?.dispose()
  viewChart?.dispose()
  categoryChart?.dispose()
  tagChart?.dispose()
  publishChart = null
  viewChart = null
  categoryChart = null
  tagChart = null
}

onMounted(() => {
  loadData()
  resizeObserver = new ResizeObserver(handleResize)
  for (const element of [
    publishChartRef.value,
    viewChartRef.value,
    categoryChartRef.value,
    tagChartRef.value,
  ])
    if (element) resizeObserver.observe(element)
  window.addEventListener('resize', handleResize)
})
onActivated(() => {
  if (activated) loadData()
  activated = true
  nextTick(handleResize)
})
onDeactivated(() => {
  generation++
  loading.value = false
})

onBeforeUnmount(() => {
  disposed = true
  generation++
  cancelAnimationFrame(resizeFrame)
  resizeObserver?.disconnect()
  window.removeEventListener('resize', handleResize)
  disposeCharts()
})
</script>

<style scoped lang="scss">
.content-statistics {
  padding: 20px;
  min-width: 0;
}
.statistics-shell {
  background: transparent;
  border: 0;
  overflow: visible;
}
.statistics-shell > :deep(.el-card__header) {
  padding: 16px 20px;
  margin-bottom: 20px;
  border: 0;
  border-radius: 4px;
  background: #fff;
}
.statistics-shell > :deep(.el-card__body) {
  padding: 0;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  font-size: 16px;
  font-weight: 500;
}
.card-header :deep(.el-date-editor) {
  flex: 0 1 340px;
  max-width: 100%;
}
.statistics-shell :deep(.el-card) {
  border: 0;
  border-radius: 4px;
  box-shadow: none;
}
.statistics-shell :deep(.el-card__header) {
  font-weight: 500;
  font-size: 14px;
  border-color: #f0f2f5;
}
.stats-cards {
  row-gap: 16px;
}
.stat-item {
  position: relative;
  padding: 8px 0;
}
.stat-value {
  font-size: 30px;
  font-weight: 500;
  color: #303133;
  line-height: 1.3;
  margin-bottom: 8px;
}
.stat-label {
  font-size: 13px;
  color: #909399;
}
.stat-icon {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 23px;
}
.article-icon {
  background: #ecf5ff;
  color: #409eff;
}
.published-icon {
  background: #f0f9eb;
  color: #67c23a;
}
.draft-icon {
  background: #fdf6ec;
  color: #e6a23c;
}
.views-icon {
  background: #f4f0ff;
  color: #9b70e8;
}
.charts-row,
.category-tag-row {
  margin-top: 20px;
  row-gap: 20px;
}
.hot-articles-card,
.hot-tags-card {
  margin-top: 20px;
}
.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 0;
}
.tag-item {
  cursor: default;
}
@media (max-width: 720px) {
  .content-statistics {
    padding: 12px;
  }
  .stat-icon {
    width: 30px;
    height: 30px;
    font-size: 18px;
    top: 12px;
    transform: none;
  }
  .stat-value {
    font-size: 24px;
  }
  .stat-item {
    padding-top: 0;
  }
  .stat-card :deep(.el-card__body) {
    padding: 14px;
  }
  .card-header :deep(.el-date-editor) {
    flex-basis: 100%;
    width: 100%;
  }
}
</style>
