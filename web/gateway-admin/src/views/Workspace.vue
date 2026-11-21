<script setup>
import { computed, ref, shallowRef, watch } from 'vue'
import { Plus, Refresh, Search, Upload } from '@element-plus/icons-vue'
import { api } from '../api'
import { createEditor } from '../resources'
import { pages } from '../layout/pages'
import Dialog from '../components/Dialog.vue'
const props = defineProps({
  section: String,
  workspace: Object,
  draft: Object,
  dirty: Boolean,
  loading: Boolean,
})
const emit = defineEmits(['reload'])
const keyword = ref(''),
  search = ref(''),
  page = ref(1),
  limit = 15
const busy = ref(false),
  error = ref(''),
  dialogError = ref(''),
  dialog = shallowRef(null),
  totp = ref('')
const name = computed(() => pages.find((p) => p.id === props.section).name)
const cols = {
  routes: ['路由标识', '匹配主机', '路径', '目标上游', '优先级'],
  upstreams: ['上游标识', '服务或节点', '健康检查', '引用路由'],
  policies: ['策略名称', '插件', '作用范围'],
  releases: ['操作时间', '操作人', '修订', '操作类型', '结果'],
}
const records = computed(() => {
  const d = props.draft
  switch (props.section) {
    case 'routes':
      return d.routes.map((r) => ({
        id: r.id,
        value: r,
        cells: [r.id, r.host, r.paths.join('、'), r.upstream || '终端插件', r.priority],
      }))
    case 'upstreams':
      return Object.entries(d.upstreams || {}).map(([id, u]) => ({
        id,
        value: u,
        cells: [
          id,
          u.service || u.nodes?.map((n) => n.url).join('、'),
          u.health_path,
          d.routes.filter((r) => r.upstream === id).length,
        ],
      }))
    case 'policies':
      return [
        {
          id: null,
          value: d.plugins || {},
          cells: ['全局插件', Object.keys(d.plugins || {}).join('、') || '—', '全部路由'],
        },
        ...Object.entries(d.profiles || {}).map(([id, p]) => ({
          id,
          value: p,
          cells: [
            id,
            Object.keys(p || {}).join('、') || '—',
            d.routes.filter((r) => r.profile === id).length + ' 条路由',
          ],
        })),
      ]
    default:
      return props.workspace.history.map((r) => ({
        id: r.generation,
        value: r,
        cells: [
          new Date(r.created_at).toLocaleString(),
          r.actor || 'gatewayctl',
          r.previous_revision + ' → ' + (r.applied_revision || '—'),
          r.rollback_of ? '回滚' : '发布',
          { applied: '已应用', pending: '待确认', unconfirmed: '未确认' }[r.status] || r.status,
        ],
      }))
  }
})
const filtered = computed(() =>
  records.value.filter((r) => r.cells.some((v) => String(v).toLowerCase().includes(search.value))),
)
const pagesCount = computed(() => Math.max(1, Math.ceil(filtered.value.length / limit)))
const visible = computed(() => filtered.value.slice((page.value - 1) * limit, page.value * limit))
watch(pagesCount, (n) => {
  page.value = Math.min(n, page.value)
})
watch(
  () => props.section,
  () => {
    keyword.value = search.value = ''
    page.value = 1
    error.value = ''
  },
)
function resetQuery() {
 keyword.value=''
 query()
}
function query() {
  search.value = keyword.value.trim().toLowerCase()
  page.value = 1
}
function open(config) {
  dialogError.value = ''
  error.value = ''
  totp.value = ''
  dialog.value = config
}
function edit(row) {
  const editing = createEditor(props.section, props.draft, row?.id, row?.id === null)
  open({
    title: (row ? '编辑' : '新增') + name.value,
    confirm: '保存到草稿',
    fields: editing.fields,
    submit: editing.save,
  })
}
function remove(row) {
  const s = props.section,
    d = props.draft
  const deps =
    s === 'routes'
      ? []
      : d.routes.filter((r) => (s === 'upstreams' ? r.upstream : r.profile) === row.id)
  if (deps.length) {
    error.value = '请先修改引用此项的路由：' + deps.map((r) => r.id).join('、')
    return
  }
  open({
    title: '删除' + name.value,
    confirm: '确认删除',
    danger: true,
    text: '从草稿移除「' + row.id + '」？发布后生效。',
    submit() {
      if (s === 'routes') d.routes = d.routes.filter((r) => r.id !== row.id)
      else delete d[s === 'upstreams' ? 'upstreams' : 'profiles'][row.id]
    },
  })
}
function reload() {
  if (props.dirty)
    open({
      title: '重新载入配置',
      confirm: '放弃草稿并载入',
      text: '尚未发布的修改会被丢弃。',
      submit: () => emit('reload'),
    })
  else emit('reload')
}
function objectAt(doc, change) {
  const [kind, id] = change.slice(2).split('/')
  if (kind === 'route') return doc.routes.find((r) => r.id === id)
  if (kind === 'upstream') return doc.upstreams?.[id]
  if (kind === 'profile') return doc.profiles?.[id]
  return doc.plugins
}
async function preview() {
  if (busy.value) return
  busy.value = true
  const payload = {
    revision: props.workspace.state.revision,
    manifest: JSON.parse(JSON.stringify(props.draft)),
  }
  try {
    const { changes } = await api('preview', payload)
    if (!changes?.length) {
      error.value = '没有待发布的变更'
      return
    }
    open({
      title: '审阅配置变更',
      confirm: '确认发布',
      secure: true,
      changes: changes.map((key) => ({
        key,
        before: objectAt(props.workspace.state.manifest, key),
        after: objectAt(payload.manifest, key),
      })),
      submit: async () => {
        await api('publish', payload, { totp: totp.value })
        emit('reload')
      },
    })
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
function rollback(record) {
  open({
    title: '回滚当前发布',
    text: '将修订 ' + record.previous_revision + ' 的配置作为新修订发布。',
    confirm: '确认回滚',
    secure: true,
    danger: true,
    submit: async () => {
      await api(
        'rollback',
        { revision: props.workspace.state.revision, generation: record.generation },
        { totp: totp.value },
      )
      emit('reload')
    },
  })
}
async function submit() {
  if (busy.value) return
  busy.value = true
  dialogError.value = ''
  try {
    await dialog.value.submit?.()
    dialog.value = null
  } catch (e) {
    dialogError.value = e.message
  } finally {
    busy.value = false
  }
}
</script>
<template>
  <form class="card filters" @submit.prevent="query">
    <label for="keyword">{{ section === 'releases' ? '记录搜索' : '关键词' }}</label
    ><span class="input-icon"
      ><Search /><input id="keyword" v-model="keyword" type="search" :placeholder="'搜索' + name"
    /></span>
    <div class="filter-actions">
      <button
        type="button"
        @click="resetQuery"
      >
        重置</button
      ><button class="primary">查询</button>
    </div>
  </form>
  <section
    class="card listing"
    :aria-label="name"
    :aria-busy="loading || busy"
    :inert="loading || busy"
  >
    <div class="tools">
      <button v-if="section !== 'releases' && workspace.publish" class="primary" @click="edit()">
        <Plus />新增{{ name }}</button
      ><button v-if="dirty" class="publish" @click="preview"><Upload />审阅并发布</button>
      <span class="revision"
        ><i></i>配置修订 {{ workspace.state.revision }} · 已应用
        {{ workspace.state.applied_revision }}</span
      ><span v-if="workspace.state.file_managed" class="tag">文件管理 · 只读</span
      ><span v-if="dirty" class="tag">未发布修改</span
      ><button class="icon refresh" title="重新载入" aria-label="重新载入" @click="reload">
        <Refresh />
      </button>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="table-scroll">
      <table>
        <thead>
          <tr>
            <th v-for="label in cols[section]" :key="label">{{ label }}</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in visible" :key="row.id">
            <td v-for="(value, index) in row.cells" :key="index">
              <span :class="{ mono: index === 0 && section !== 'releases' }">{{ value }}</span>
            </td>
            <td class="row-actions">
              <template v-if="section === 'releases'"
                ><button
                  v-if="
                    workspace.rollback &&
                    !dirty &&
                    row.value.status === 'applied' &&
                    row.value.applied_revision === workspace.state.revision
                  "
                  @click="rollback(row.value)"
                >
                  回滚</button
                ><span v-else>—</span></template
              >
              <template v-else-if="workspace.publish"
                ><button @click="edit(row)">编辑</button
                ><button v-if="row.id !== null" class="remove" @click="remove(row)">
                  删除
                </button></template
              ><button
                v-else
                @click="open({ title: '查看' + name, detail: JSON.stringify(row.value, null, 2) })"
              >
                查看
              </button>
            </td>
          </tr>
          <tr v-if="!visible.length">
            <td class="empty" :colspan="cols[section].length + 1">
              {{ search ? '没有匹配的记录' : '暂无' + name }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <footer class="pagination">
      <span>共 {{ filtered.length }} 条</span
      ><button :disabled="page <= 1" aria-label="上一页" @click="page--">‹</button
      ><span>{{ page }} / {{ pagesCount }}</span
      ><button :disabled="page >= pagesCount" aria-label="下一页" @click="page++">›</button>
    </footer>
  </section>
  <Dialog
    v-if="dialog"
    :title="dialog.title"
    :confirm="dialog.confirm"
    :danger="dialog.danger"
    :wide="!!dialog.fields || !!dialog.changes"
    :error="dialogError"
    :busy="busy"
    @submit="submit"
    @close="dialog = null"
  >
    <p v-if="dialog.text">{{ dialog.text }}</p>
    <pre v-if="dialog.detail">{{ dialog.detail }}</pre>
    <div v-if="dialog.fields" class="editor-grid">
      <label v-for="f in dialog.fields" :key="f.label" :class="{ full: f.multiline }"
        >{{ f.label }}
        <select v-if="f.choices" v-model="f.value">
          <option v-for="v in f.choices" :value="v" :key="v">{{ v || '无' }}</option>
        </select>
        <textarea
          v-else-if="f.multiline"
          v-model="f.value"
          rows="3"
          :required="f.required"
          :placeholder="f.placeholder"
        ></textarea>
        <input
          v-else
          :value="f.value"
          @input="f.value = $event.target.value"
          :type="f.type || 'text'"
          :required="f.required"
          :readonly="f.readOnly"
          :pattern="f.pattern"
          :placeholder="f.placeholder"
        />
      </label>
    </div>
    <details v-for="change in dialog.changes" :key="change.key">
      <summary>{{ change.key }}</summary>
      <template v-if="change.before !== undefined"
        ><small>当前配置</small>
        <pre>{{ JSON.stringify(change.before, null, 2) }}</pre></template
      ><template v-if="change.after !== undefined"
        ><small>待发布</small>
        <pre>{{ JSON.stringify(change.after, null, 2) }}</pre>
      </template>
    </details>
    <label v-if="dialog.secure"
      >动态验证码（已启用时填写）<input
        v-model="totp"
        pattern="[0-9]{6}"
        maxlength="6"
        inputmode="numeric"
        autocomplete="one-time-code"
    /></label>
  </Dialog>
</template>
