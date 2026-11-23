<script setup>
import { computed, ref, shallowRef, watch } from "vue";
import { Plus, Refresh, Search, Upload } from "@element-plus/icons-vue";
import { api } from "../api";
import { createEditor } from "../resources";
import { pages } from "../layout/pages";
import { columns, describe, describeChange } from "../rule-display";
import Dialog from "../components/Dialog.vue";
import EditorFields from "../components/EditorFields.vue";
import RouteMatch from "../components/RouteMatch.vue";
import FieldChanges from "../components/FieldChanges.vue";

const props = defineProps({
  section: String,
  workspace: Object,
  draft: Object,
  dirty: Boolean,
  loading: Boolean,
});
const emit = defineEmits(["reload"]);
const search = ref(""),
  page = ref(1),
  limit = 15;
const busy = ref(false),
  error = ref(""),
  dialogError = ref(""),
  dialog = shallowRef(null),
  password = ref("");
const current = computed(() => pages.find((p) => p.id === props.section));
const history = ref({ items: [], total: 0 }),
  historyBusy = ref(false);
const showingActive = ref(false);
const diverged = computed(
  () =>
    !props.workspace.state.desired_unavailable &&
    props.workspace.state.revision !== props.workspace.state.applied_revision,
);
const activeView = computed(
  () =>
    showingActive.value &&
    diverged.value &&
    props.workspace.state.applied_manifest &&
    props.section !== "releases",
);
const editable = computed(() => props.workspace.publish && !activeView.value);
const records = computed(() => {
  const doc = activeView.value
      ? props.workspace.state.applied_manifest || {}
      : props.draft,
    section = props.section;
  let entries;
  switch (section) {
    case "routes":
    case "streams":
      entries = (doc[section] || []).map((rule) => [rule.id, rule]);
      break;
    case "upstreams":
      entries = Object.entries(doc.upstreams || {});
      break;
    case "policies":
      entries = [
        [null, doc.plugins || {}],
        ...Object.entries(doc.profiles || {}),
      ];
      break;
    case "releases":
      entries = (history.value.items || []).map((r) => [r.generation, r]);
      break;
  }
  return entries.map(([id, value]) => ({
    id,
    value,
    cells: describe(section, value, id, doc),
  }));
});
const filtered = computed(() => {
  const query = search.value.trim().toLowerCase();
  return records.value.filter((r) =>
    r.cells.flat().some((v) =>
      String(v || "")
        .toLowerCase()
        .includes(query),
    ),
  );
});
const total = computed(() =>
  props.section === "releases" ? history.value.total : filtered.value.length,
);
const pagesCount = computed(() => Math.max(1, Math.ceil(total.value / limit)));
const visible = computed(() =>
  props.section === "releases"
    ? filtered.value
    : filtered.value.slice((page.value - 1) * limit, page.value * limit),
);
watch(pagesCount, (n) => {
  page.value = Math.min(n, page.value);
});
watch(search, () => {
  if (props.section !== "releases") page.value = 1;
});
watch(
  () => props.section,
  () => {
    search.value = error.value = "";
    page.value = 1;
    close();
  },
);
watch(
  [() => props.section, page, () => props.workspace],
  async ([section, index], _, onCleanup) => {
    if (section !== "releases") {
      historyBusy.value = false;
      return;
    }
    let active = true;
    onCleanup(() => {
      active = false;
    });
    historyBusy.value = true;
    error.value = "";
    try {
      const data = await api("history?page=" + index + "&limit=" + limit);
      if (active) history.value = data;
    } catch (e) {
      if (active) {
        error.value = e.message;
        history.value = { items: [], total: 0 };
      }
    } finally {
      if (active) historyBusy.value = false;
    }
  },
  { immediate: true },
);

function close() {
  dialog.value = null;
  password.value = "";
}
function open(config) {
  dialogError.value = error.value = password.value = "";
  dialog.value = config;
}
function edit(row) {
  const editor = createEditor(
    props.section,
    props.draft,
    row?.id,
    row?.id === null,
  );
  open({
    title: (row ? "编辑" : "新增") + current.value.name,
    confirm: "保存到草稿",
    fields: editor.fields,
    submit: editor.save,
  });
}
function remove(row) {
  const s = props.section,
    d = props.draft;
  if (s === "upstreams" || s === "policies") {
    const field = s === "upstreams" ? "upstream" : "profile";
    const refs = (d.routes || []).filter((r) => r[field] === row.id);
    if (refs.length) {
      error.value =
        "请先修改引用此项的路由：" + refs.map((r) => r.id).join("、");
      return;
    }
  }
  open({
    title: "删除" + current.value.name,
    confirm: "从草稿移除",
    danger: true,
    text: "移除「" + row.id + "」？发布草稿后生效。",
    submit() {
      if (s === "routes" || s === "streams")
        d[s] = d[s].filter((r) => r.id !== row.id);
      else delete d[s === "upstreams" ? "upstreams" : "profiles"][row.id];
    },
  });
}
function view(row) {
  open({
    title: row.id || "全局设置",
    labels: columns[props.section],
    values: row.cells,
    detail: row.value,
  });
}
function reload() {
  if (props.dirty)
    open({
      title: "重新载入配置",
      confirm: "合并最新配置",
      text: "保留草稿并核对最新修订。双方同时修改的资源需要逐项确认。",
      discard: true,
      submit: () => emit("reload", true),
    });
  else emit("reload");
}
async function preview(record) {
  if (busy.value) return;
  busy.value = true;
  const payload = {
    revision: props.workspace.state.revision,
    ...(record
      ? { generation: record.generation }
      : { manifest: JSON.parse(JSON.stringify(props.draft)) }),
  };
  try {
    const { changes, manifest } = await api("preview", payload);
    if (!changes?.length) {
      error.value = "没有待发布的变更";
      return;
    }
    open({
      title: record ? "审阅回滚变更" : "审阅配置变更",
      confirm: record ? "确认回滚" : "确认发布",
      danger: !!record,
      text: record
        ? "将修订 " +
          record.applied_revision +
          " 恢复为修订 " +
          record.previous_revision +
          " 的配置，确认后生成新修订。"
        : "",
      secure: true,
      changes: changes.map((key) =>
        describeChange(
          key,
          props.workspace.state.manifest,
          payload.manifest || manifest,
        ),
      ),
      submit: async () => {
        await api(record ? "rollback" : "publish", {
          ...payload,
          password: password.value,
        });
        emit("reload");
      },
    });
  } catch (e) {
    error.value = e.message;
    if (e.status === 409) emit("reload", true);
  } finally {
    busy.value = false;
  }
}
async function submit() {
  if (busy.value) return;
  busy.value = true;
  dialogError.value = "";
  try {
    await dialog.value.submit?.();
    close();
  } catch (e) {
    dialogError.value = e.message;
    password.value = "";
    if (e.status === 409 && dialog.value?.secure) {
      close();
      error.value = e.message;
      emit("reload", true);
    }
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <header class="page-intro">
    <div>
      <h1>{{ current.name }}</h1>
      <p>{{ current.description }}</p>
    </div>
    <div class="page-actions">
      <button
        v-if="dirty && workspace.publish"
        class="publish"
        :disabled="busy || loading"
        @click="preview()"
      >
        <Upload />审阅并发布
      </button>
      <button
        v-if="section !== 'releases' && editable"
        class="primary"
        :disabled="busy || loading"
        @click="edit()"
      >
        <Plus />新增{{ current.name }}
      </button>
    </div>
  </header>
  <RouteMatch
    v-if="section === 'routes'"
    :listen="workspace.state.runtime?.http_listen"
  />
  <details
    v-if="
      section === 'routes' && workspace.state.runtime?.path_warnings?.length
    "
    class="card runtime-panel"
  >
    <summary>
      路径匹配提示（{{ workspace.state.runtime.path_warnings.length }}）
    </summary>
    <ul>
      <li
        v-for="warning in workspace.state.runtime.path_warnings"
        :key="warning"
      >
        {{ warning }}
      </li>
    </ul>
  </details>
  <section
    v-if="
      diverged ||
      workspace.state.desired_unavailable ||
      workspace.state.last_error
    "
    class="card runtime-alert"
    role="status"
  >
    <strong v-if="workspace.state.desired_unavailable"
      >期望配置暂不可用 · 本机生效修订
      {{ workspace.state.applied_revision }}（只读）</strong
    >
    <strong v-else
      >配置修订 {{ workspace.state.revision }} · 本机生效修订
      {{ workspace.state.applied_revision }}</strong
    >
    <p v-if="workspace.state.last_error">{{ workspace.state.last_error }}</p>
    <p v-if="diverged && !workspace.state.applied_manifest">
      本机生效配置暂不可用。
    </p>
    <label
      v-if="
        diverged && section !== 'releases' && workspace.state.applied_manifest
      "
    >
      查看配置<select aria-label="查看配置" v-model="showingActive">
        <option :value="false">期望配置与草稿</option>
        <option :value="true">本机生效配置（只读）</option>
      </select>
    </label>
  </section>
  <details
    v-if="section === 'upstreams' && workspace.state.runtime"
    open
    class="card runtime-panel"
  >
    <summary>本机上游健康状态</summary>
    <div
      v-for="(nodes, id) in workspace.state.runtime.upstreams"
      :key="id"
      class="runtime-row"
    >
      <strong>{{ id }}</strong>
      <span v-for="n in nodes" :key="n.url" :class="{ unhealthy: !n.healthy }"
        >{{ n.url }} · {{ n.healthy ? "健康" : "不可用" }} · 权重
        {{ n.weight }}</span
      >
      <span v-if="!nodes?.length">暂无节点</span>
    </div>
    <p v-if="!Object.keys(workspace.state.runtime.upstreams || {}).length">
      本机暂无 HTTP 上游。
    </p>
  </details>
  <details
    v-if="section === 'streams' && workspace.state.runtime"
    open
    class="card runtime-panel"
  >
    <summary>本机正在监听的端口</summary>
    <div
      class="runtime-row"
      v-for="s in workspace.state.runtime.streams"
      :key="s.id"
    >
      <strong>{{ s.id }}</strong
      ><span
        >{{ s.protocol.toUpperCase() }} · {{ s.address || s.listen }} →
        {{ s.target }}</span
      >
    </div>
    <p v-if="!workspace.state.runtime.streams?.length">本机暂无四层监听。</p>
  </details>
  <section
    class="card listing"
    :aria-label="current.name"
    :aria-busy="loading || busy"
    :inert="loading || busy"
  >
    <div class="tools">
      <label class="search-control"
        ><span class="input-icon"
          ><Search />
          <input
            v-model="search"
            type="search"
            :aria-label="'搜索' + current.name"
            :placeholder="
              section === 'releases' ? '搜索本页记录' : '搜索' + current.name
            "
          /> </span
      ></label>
      <span class="revision"
        ><i
          :class="{
            pending:
              workspace.state.revision !== workspace.state.applied_revision,
          }"
        ></i>
        <template v-if="workspace.state.desired_unavailable"
          >本机生效修订 {{ workspace.state.applied_revision }}</template
        >
        <template v-else
          >配置修订 {{ workspace.state.revision }} · 已应用
          {{ workspace.state.applied_revision }}</template
        ></span
      >
      <span v-if="dirty" class="tag">有草稿</span>
      <button
        class="icon refresh"
        title="重新载入"
        aria-label="重新载入"
        @click="reload"
      >
        <Refresh />
      </button>
    </div>
    <p v-if="workspace.state.file_managed" class="mode-note">
      文件管理 · 只读。修改网关配置文件后，执行 gateway reload 应用变更。
    </p>
    <p v-else-if="workspace.role === 'viewer'" class="mode-note">
      只读账号 · 可查看规则和发布历史。
    </p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div class="table-scroll">
      <table>
        <thead>
          <tr>
            <th v-for="label in columns[section]" :key="label">{{ label }}</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in visible" :key="row.id">
            <td v-for="([primary, secondary], i) in row.cells" :key="i">
              <span
                :class="{
                  'rule-id': i === 0 && section !== 'releases',
                  protocol: section === 'streams' && i === 1,
                }"
                >{{ primary }}</span
              >
              <small
                v-if="secondary"
                class="cell-note"
                :title="String(secondary)"
                >{{ secondary }}</small
              >
            </td>
            <td class="row-actions">
              <template v-if="section === 'releases'">
                <button
                  v-if="
                    workspace.rollback &&
                    !dirty &&
                    ['applied', 'committed'].includes(row.value.status) &&
                    row.value.applied_revision === workspace.state.revision
                  "
                  @click="preview(row.value)"
                >
                  回滚
                </button>
                <button @click="view(row)">查看</button>
              </template>
              <template v-else-if="editable">
                <button @click="edit(row)">编辑</button
                ><button
                  v-if="row.id !== null"
                  class="remove"
                  @click="remove(row)"
                >
                  删除
                </button>
              </template>
              <button v-else @click="view(row)">查看</button>
            </td>
          </tr>
          <tr v-if="!visible.length">
            <td class="empty" :colspan="columns[section].length + 1">
              {{
                historyBusy
                  ? "正在载入发布记录…"
                  : search
                    ? "没有匹配的记录"
                    : "暂无" + current.name
              }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <footer class="pagination">
      <span>共 {{ total }} 条</span>
      <button :disabled="page <= 1" aria-label="上一页" @click="page--">
        ‹
      </button>
      <span>{{ page }} / {{ pagesCount }}</span
      ><button
        :disabled="page >= pagesCount"
        aria-label="下一页"
        @click="page++"
      >
        ›
      </button>
    </footer>
  </section>
  <Dialog
    v-if="dialog"
    :title="dialog.title"
    :confirm="dialog.confirm"
    :danger="dialog.danger"
    :wide="!!dialog.fields || !!dialog.changes || !!dialog.values"
    :error="dialogError"
    :busy="busy"
    @submit="submit"
    @close="close"
  >
    <p v-if="dialog.text">{{ dialog.text }}</p>
    <button
      v-if="dialog.discard"
      type="button"
      class="remove"
      @click="
        open({
          title: '放弃草稿',
          text: '确认丢弃所有未发布修改并载入最新配置？',
          confirm: '放弃草稿',
          danger: true,
          submit: () => emit('reload'),
        })
      "
    >
      放弃草稿…
    </button>
    <dl v-if="dialog.values" class="field-list">
      <template v-for="([value, note], i) in dialog.values" :key="i"
        ><dt>{{ dialog.labels[i] }}</dt>
        <dd>
          {{ value }}<small v-if="note">{{ note }}</small>
        </dd></template
      >
    </dl>
    <details v-if="dialog.detail">
      <summary>查看配置原文</summary>
      <pre>{{ JSON.stringify(dialog.detail, null, 2) }}</pre>
    </details>
    <EditorFields v-if="dialog.fields" :fields="dialog.fields" />
    <div v-for="change in dialog.changes" :key="change.key" class="change">
      <strong>{{ change.title }}</strong>
      <FieldChanges :changes="change.fields" />
      <details v-if="change.affected.length">
        <summary>关联路由（{{ change.affected.length }}）</summary>
        <p>{{ change.affected.join("、") }}</p>
      </details>
    </div>
    <label v-if="dialog.secure" class="confirm-password"
      >确认网关账号密码<input
        v-model="password"
        required
        type="password"
        autocomplete="current-password"
    /></label>
  </Dialog>
</template>
