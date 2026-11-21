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
import { useStatisticsCharts } from './useStatisticsCharts'
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

const { publishChartRef, viewChartRef, categoryChartRef, tagChartRef, initCharts, handleResize, disposeCharts } =
  useStatisticsCharts(days, categoryData, tagData)

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
  resizeObserver?.disconnect()
  window.removeEventListener('resize', handleResize)
  disposeCharts()
})
</script>

<style scoped lang="scss" src="./statistics.scss"></style>
