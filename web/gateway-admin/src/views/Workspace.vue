<script setup>
import { computed, ref, shallowRef, watch } from "vue";
import { Plus, Refresh, Search, Upload } from "@element-plus/icons-vue";
import { api } from "../api";
import { createEditor } from "../resources";
import { pages } from "../layout/pages";
import { columns, describe, describeChange } from "../rule-display";
import Dialog from "../components/Dialog.vue";
import EditorFields from "../components/EditorFields.vue";

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
const records = computed(() => {
  const doc = props.draft,
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
      entries = props.workspace.history.map((r) => [r.generation, r]);
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
const pagesCount = computed(() =>
  Math.max(1, Math.ceil(filtered.value.length / limit)),
);
const visible = computed(() =>
  filtered.value.slice((page.value - 1) * limit, page.value * limit),
);
watch(pagesCount, (n) => {
  page.value = Math.min(n, page.value);
});
watch(search, () => {
  page.value = 1;
});
watch(
  () => props.section,
  () => {
    search.value = error.value = "";
    page.value = 1;
    close();
  },
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
      confirm: "放弃草稿并载入",
      text: "尚未发布的修改会被丢弃。",
      submit: () => emit("reload"),
    });
  else emit("reload");
}
async function preview() {
  if (busy.value) return;
  busy.value = true;
  const payload = {
    revision: props.workspace.state.revision,
    manifest: JSON.parse(JSON.stringify(props.draft)),
  };
  try {
    const { changes } = await api("preview", payload);
    if (!changes?.length) {
      error.value = "没有待发布的变更";
      return;
    }
    open({
      title: "审阅配置变更",
      confirm: "确认发布",
      secure: true,
      changes: changes.map((key) =>
        describeChange(key, props.workspace.state.manifest, payload.manifest),
      ),
      submit: async () => {
        await api("publish", { ...payload, password: password.value });
        emit("reload");
      },
    });
  } catch (e) {
    error.value = e.message;
  } finally {
    busy.value = false;
  }
}
function rollback(record) {
  open({
    title: "回滚当前发布",
    confirm: "确认回滚",
    secure: true,
    danger: true,
    text:
      "将当前修订 " +
      record.applied_revision +
      " 的全部配置恢复为修订 " +
      record.previous_revision +
      "。这次回滚会生成新的发布记录。",
    submit: async () => {
      await api("rollback", {
        revision: props.workspace.state.revision,
        generation: record.generation,
        password: password.value,
      });
      emit("reload");
    },
  });
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
        @click="preview"
      >
        <Upload />审阅并发布
      </button>
      <button
        v-if="section !== 'releases' && workspace.publish"
        class="primary"
        :disabled="busy || loading"
        @click="edit()"
      >
        <Plus />新增{{ current.name }}
      </button>
    </div>
  </header>
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
            :placeholder="'搜索' + current.name"
          /> </span
      ></label>
      <span class="revision"
        ><i
          :class="{
            pending:
              workspace.state.revision !== workspace.state.applied_revision,
          }"
        ></i>
        配置修订 {{ workspace.state.revision }} · 已应用
        {{ workspace.state.applied_revision }}</span
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
                    row.value.status === 'applied' &&
                    row.value.applied_revision === workspace.state.revision
                  "
                  @click="rollback(row.value)"
                >
                  回滚
                </button>
                <span v-else>—</span>
              </template>
              <template v-else-if="workspace.publish">
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
              {{ search ? "没有匹配的记录" : "暂无" + current.name }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <footer class="pagination">
      <span>共 {{ filtered.length }} 条</span>
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
      <p>{{ change.summary }}</p>
      <details>
        <summary>配置差异</summary>
        <template v-if="change.before !== undefined"
          ><small>当前配置</small>
          <pre>{{ JSON.stringify(change.before, null, 2) }}</pre>
        </template>
        <template v-if="change.after !== undefined"
          ><small>待发布</small>
          <pre>{{ JSON.stringify(change.after, null, 2) }}</pre>
        </template>
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
