import request from '@/utils/request'
const records = response => ({ ...response, data: response.data.items || [] })

export const getDashboardStats = () => request.get('/blog/statistics/dashboard')
export const getLatestArticles = (params) => request.get('/blog/articles/recent', { params })
export const getArticleStats = (params, config = {}) =>
  request.get('/blog/statistics/articlestats', { ...config, params })
export const getArticleTrend = (params, config = {}) =>
  request.get('/blog/statistics/articletrend', { ...config, params })
export const getViewTrend = (params, config = {}) =>
  request.get('/blog/statistics/viewtrend', { ...config, params })
export const getHotArticles = (params, config = {}) =>
  request.get('/blog/statistics/hotarticlestats', { ...config, params }).then(records)
export const getCategoryStats = (config = {}) => request.get('/blog/statistics/categorystats', config).then(records)
export const getTagStats = (config = {}) => request.get('/blog/statistics/tagstats', config).then(records)
export const getHotTags = (params, config = {}) =>
  request.get('/blog/statistics/hottags', { ...config, params }).then(records)
