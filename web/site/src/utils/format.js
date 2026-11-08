import dayjs from 'dayjs'

export const formatDate = (date, format = 'YYYY-MM-DD') => date ? dayjs(date).format(format) : ''

export const truncateText = (text, length = 150) => {
  if (!text || text.length <= length) return text || ''
  return `${text.slice(0, length)}...`
}
