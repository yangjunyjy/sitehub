import { Marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import go from 'highlight.js/lib/languages/go'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import sql from 'highlight.js/lib/languages/sql'
import python from 'highlight.js/lib/languages/python'
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('go', go)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('python', python)
import { createArticleRenderer } from '../../../shared/article-markdown'

export const renderArticle = createArticleRenderer(Marked, DOMPurify, hljs)
export const MAX_ARTICLE_LENGTH = 500000
export function validateMarkdownFile(file) {
  if (!/\.(md|markdown)$/i.test(file.name))
    throw new Error('只支持 Markdown 文件（.md / .markdown）')
  if (!file.size || file.size > 2_000_000) throw new Error('文件须在 1 字节到 2 MB 之间')
}
export function draftFields(article) {
  return {
    title: article.title || '',
    slug: article.slug || '',
    summary: article.summary || '',
    content: article.content || '',
    content_format: 'markdown',
    cover: article.cover || '',
    category_id: String(article.category_id || '0'),
    tag_ids: [...(article.tag_ids || [])],
  }
}
export const draftFingerprint = (article) => JSON.stringify(draftFields(article))
export { createUUID as newRequestID } from './uuid.js'
