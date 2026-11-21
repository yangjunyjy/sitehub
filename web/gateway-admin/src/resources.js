import { reactive } from 'vue'

export function createEditor(section, doc, key, global) {
  const fields = []
  function field(label, value, options = {}) {
    const model = reactive({ label, value: String(value ?? ''), ...options })
    fields.push(model)
    return model
  }
  const save =
    section === 'routes'
      ? routeFields(doc, key, field)
      : section === 'upstreams'
        ? upstreamFields(doc, key, field)
        : policyFields(doc, key, global, field)
  return { fields, save }
}

const integer = (text, min, max) => {
  const n = Number(text)
  if (!text.trim() || !Number.isSafeInteger(n) || n < min || n > max)
    throw Error('请填写范围内的整数')
  return n
}
const words = (value) => value.split(/[\s,，]+/).filter(Boolean)
const jsonObject = (control) => {
  let parsed
  try {
    parsed = JSON.parse(control.value || '{}')
  } catch {
    throw Error('插件配置的 JSON 格式有误')
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object')
    throw Error('插件配置须为 JSON 对象')
  return parsed
}
const identifier = (field, label, key) =>
  field(label, key, { required: true, readOnly: !!key, pattern: '[a-zA-Z0-9_\\-]{1,64}' })
const plugins = (field, value) =>
  field('插件覆盖（JSON）', JSON.stringify(value || {}, null, 2), { multiline: true })
const advancedNumbers = [
  ['connect_timeout_ms', '连接超时（毫秒）', 3000, 1, 60000],
  ['response_header_timeout_ms', '响应头超时（毫秒）', 10000, 1, 3600000],
  ['io_timeout_ms', '读写超时（毫秒）', 30000, 1, 3600000],
  ['health_interval_ms', '健康检查间隔（毫秒）', 5000, 100, 60000],
  ['health_timeout_ms', '健康检查超时（毫秒）', 1000, 1, 60000],
  ['health_failures', '连续失败阈值', 3, 1, 100],
]

function routeFields(doc, key, field) {
  const row = doc.routes.find((r) => r.id === key) || {}
  const id = identifier(field, '路由标识', key)
  const host = field('匹配主机', row.host, { required: true, placeholder: 'api.example.com' })
  const paths = field('路径（每行一条）', row.paths?.join('\n'), {
    required: true,
    multiline: true,
    placeholder: '/api/*',
  })
  const methods = field('请求方法（留空匹配全部）', row.methods?.join(', '))
  const priority = field('优先级', row.priority ?? 0, { type: 'number', required: true })
  const upstream = field('目标上游', row.upstream, {
    choices: ['', ...Object.keys(doc.upstreams || {})],
  })
  const profile = field('插件策略', row.profile, {
    choices: ['', ...Object.keys(doc.profiles || {})],
  })
  const config = plugins(field, row.plugins)
  return () => {
    if (!key && doc.routes.some((r) => r.id === id.value)) throw Error('路由标识已存在')
    const value = {
      id: id.value,
      host: host.value.trim().toLowerCase(),
      paths: words(paths.value),
      priority: integer(priority.value, -2147483648, 2147483647),
    }
    if (methods.value.trim()) value.methods = words(methods.value.toUpperCase())
    if (upstream.value) value.upstream = upstream.value
    if (profile.value) value.profile = profile.value
    const overrides = jsonObject(config)
    if (Object.keys(overrides).length) value.plugins = overrides
    if (key) doc.routes.splice(doc.routes.indexOf(row), 1, value)
    else doc.routes.push(value)
  }
}
function upstreamFields(doc, key, field) {
  const row = doc.upstreams?.[key] || {}
  const id = identifier(field, '上游标识', key)
  const service = field('服务发现名称', row.service)
  const nodes = field(
    '静态节点（每行填写地址和权重，与服务发现二选一）',
    row.nodes?.map((n) => n.url + ' ' + n.weight).join('\n'),
    { multiline: true, placeholder: 'http://127.0.0.1:8080 1' },
  )
  const path = field('健康检查路径', row.health_path || '/healthz', { required: true })
  const values = advancedNumbers.map(([name, label, initial]) =>
    field(label, row[name] ?? initial, { required: true, type: 'number' }),
  )
  return () => {
    if (!key && Object.hasOwn(doc.upstreams || {}, id.value)) throw Error('上游标识已存在')
    const name = service.value.trim(),
      list = nodes.value.trim()
    if (!!name === !!list) throw Error('请填写服务发现名称或静态节点，两者只能选择一项')
    const value = { health_path: path.value.trim() }
    if (name) value.service = name
    else
      value.nodes = list
        .split(/\n/)
        .filter((line) => line.trim())
        .map((line) => {
          const [url, weight = '1', extra] = line.trim().split(/\s+/)
          if (extra) throw Error('每行填写一个节点地址和权重')
          return { url, weight: integer(weight, 1, 1000) }
        })
    advancedNumbers.forEach(([name, , , min, max], i) => {
      value[name] = integer(values[i].value, min, max)
    })
    doc.upstreams = { ...doc.upstreams, [id.value]: value }
  }
}
function policyFields(doc, key, global, field) {
  const id = global ? null : identifier(field, '策略标识', key)
  const config = plugins(field, global ? doc.plugins : doc.profiles?.[key])
  return () => {
    const value = jsonObject(config)
    if (global) {
      doc.plugins = value
      return
    }
    if (!key && Object.hasOwn(doc.profiles || {}, id.value)) throw Error('策略标识已存在')
    doc.profiles = { ...doc.profiles, [id.value]: value }
  }
}
