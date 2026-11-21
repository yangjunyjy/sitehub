<template>
  <main class="manual-workspace" aria-label="文档工作区" v-loading="loading">
    <el-alert v-if="error" :title="error" type="error" :closable="false">
      <el-button link @click="load">重新加载</el-button>
    </el-alert>
    <template v-else-if="workspace">
      <section class="manual-panel">
        <div class="manual-context">
          <label>
            模式
            <el-select
              :model-value="mode?.id"
              aria-label="文档模式"
              filterable
              placeholder="选择模式"
              :disabled="saving"
              @change="selectMode"
            >
              <el-option
                v-for="item in workspace.modes"
                :key="item.id"
                :label="item.label"
                :value="item.id"
              />
            </el-select>
          </label>
          <el-dropdown trigger="click" @command="modeCommand">
            <el-button :icon="MoreFilled" aria-label="模式操作" :disabled="saving" />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="new">新建模式</el-dropdown-item>
                <el-dropdown-item command="edit" :disabled="!mode">编辑模式</el-dropdown-item>
                <el-dropdown-item
                  command="default"
                  :disabled="
                    !mode || mode.status !== 'published' || mode.id === workspace.defaultMode
                  "
                >
                  设为默认模式
                </el-dropdown-item>
                <el-dropdown-item
                  command="delete"
                  :disabled="
                    !mode || mode.id === workspace.defaultMode || !!workspace.versions.length
                  "
                >
                  删除模式
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <label>
            版本
            <el-select
              :model-value="version?.id"
              aria-label="文档版本"
              filterable
              placeholder="选择版本"
              :disabled="!mode || saving"
              @change="selectVersion"
            >
              <el-option
                v-for="item in workspace.versions"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </label>
          <el-dropdown trigger="click" @command="versionCommand">
            <el-button :icon="MoreFilled" aria-label="版本操作" :disabled="!mode || saving" />
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="new">新建版本</el-dropdown-item>
                <el-dropdown-item command="edit" :disabled="!version">编辑版本</el-dropdown-item>
                <el-dropdown-item command="publication" :disabled="!version">
                  {{ version?.status === 'published' ? '撤回版本' : '发布版本' }}
                </el-dropdown-item>
                <el-dropdown-item
                  command="default"
                  :disabled="
                    !version ||
                    version.status !== 'published' ||
                    version.id === mode?.defaultVersion
                  "
                >
                  设为默认版本
                </el-dropdown-item>
                <el-dropdown-item
                  command="delete"
                  :disabled="
                    !version ||
                    version.status === 'published' ||
                    version.id === mode?.defaultVersion ||
                    !!groups.length
                  "
                >
                  删除版本
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-tag
            v-if="version"
            :type="version.status === 'published' ? 'success' : 'info'"
            effect="plain"
          >
            {{ version.status === 'published' ? '公开版本' : '草稿版本' }}
          </el-tag>
          <div class="manual-actions manual-site-actions">
            <el-button :icon="Setting" @click="dialog.open('settings', workspace)">文档站设置</el-button>
            <el-button :icon="TopRight" @click="previewProductPage('docs', { mode: mode?.id, version: version?.id })">查看前台</el-button>
          </div>
        </div>
        <div v-if="version" class="manual-directory">
          <aside aria-label="文档分组" class="manual-groups">
            <div class="manual-section">
              <h2>分组</h2>
              <el-button :icon="Plus" text aria-label="新建分组" @click="dialog.open('group')" />
            </div>
            <button
              v-for="item in groups"
              :key="item.id"
              class="manual-group"
              :class="{ selected: group?.id === item.id }"
              :aria-current="group?.id === item.id ? 'true' : undefined"
              @click="selectGroup(item.id)"
            >
              <el-icon><Folder /></el-icon>
              {{ item.title }}
              <small>{{ item.items.length }}</small>
            </button>
            <p v-if="!groups.length">暂无分组</p>
          </aside>
          <section v-if="group" class="manual-documents">
            <div class="manual-section">
              <h2>{{ group.title }}</h2>
              <div class="manual-actions">
                <el-dropdown trigger="click" @command="groupCommand">
                  <el-button :icon="MoreFilled" aria-label="分组操作" :disabled="saving" />
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="edit">重命名</el-dropdown-item>
                      <el-dropdown-item command="up" :disabled="groups[0] === group">
                        上移
                      </el-dropdown-item>
                      <el-dropdown-item command="down" :disabled="groups.at(-1) === group">
                        下移
                      </el-dropdown-item>
                      <el-dropdown-item command="delete" :disabled="!!group.items.length">
                        删除空分组
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <el-button type="primary" :icon="Plus" @click="edit()">新建文档</el-button>
              </div>
            </div>
            <el-input v-model="search" clearable aria-label="搜索分组文档" placeholder="搜索标题、摘要或访问标识" class="manual-search" />
            <el-table :data="documents" :empty-text="search.trim() ? '没有匹配的文档' : '分组内暂无文档'">
              <el-table-column label="文档" min-width="180">
                <template #default="{ row }">
                  <button class="manual-doc-link" @click="edit(row)">{{ row.title }}</button>
                  <small>{{ row.slug }}</small>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="126">
                <template #default="{ row }">{{ documentStatus[row.status] }}</template>
              </el-table-column>
              <el-table-column label="修订" width="70" prop="revision" />
              <el-table-column label="操作" width="76">
                <template #default="{ row }">
                  <el-button
                    link
                    type="danger"
                    :disabled="saving || row.status !== 'draft'"
                    @click="removeDocument(row)"
                  >
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
          <el-empty v-else description="新建分组后即可编写文档">
            <el-button @click="dialog.open('group')">新建分组</el-button>
          </el-empty>
        </div>
        <el-empty v-else :description="mode ? '此模式暂无版本' : '暂无文档模式'">
          <el-button type="primary" @click="dialog.open(mode ? 'version' : 'mode')">
            {{ mode ? '新建版本' : '新建模式' }}
          </el-button>
        </el-empty>
      </section>
      <WorkspaceDialog ref="dialog" :workspace="workspace" @saved="onSaved" />
    </template>
  </main>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Folder, MoreFilled, Plus, Setting, TopRight } from '@element-plus/icons-vue'
import { useDocumentWorkspace } from '@/composables/useDocumentWorkspace'
import { previewProductPage } from '@/utils/product-preview'
import {
  defaultMode,
  deleteMode,
  defaultVersion,
  deleteVersion,
  publishVersion,
  withdrawVersion,
  moveGroup,
  deleteGroup,
  deleteDocument,
} from '@/api/documents'
import { documentStatus } from '@/utils/documents'
import WorkspaceDialog from './components/WorkspaceDialog.vue'

const router = useRouter(),
  route = useRoute()
const { workspace, mode, version, groups, location, loading, saving, error, load, save } =
  useDocumentWorkspace()
const dialog = ref()
const search = ref('')
const group = computed(
  () => groups.value.find((item) => String(item.id) === route.query.group) || groups.value[0],
)
const documents = computed(() => {
  const term = search.value.trim().toLocaleLowerCase()
  return (group.value?.items || []).filter(item =>
    !term || [item.title, item.slug, item.summary].some(field => field?.toLocaleLowerCase().includes(term)))
})
watch(() => group.value?.id, () => { search.value = '' })
watch(
  [() => route.query.mode, () => route.query.version],
  () => {
    if (route.meta.componentPath === 'product/docs/index') load()
  },
  { immediate: true },
)
const selectMode = (id) => router.replace({ path: route.path, query: { mode: id } })
const selectVersion = (id) =>
  router.replace({ path: route.path, query: { mode: mode.value.id, version: id } })
const selectGroup = (id) =>
  router.replace({ path: route.path, query: { ...route.query, group: id } })
const edit = (row) =>
  router.push(location('product/docs/editor/index', { group: group.value.id, id: row?.id }))
async function onSaved({ kind, id, created }) {
  if (created && kind === 'mode') await selectMode(id)
  else if (created && kind === 'version') await selectVersion(id)
  else await load()
}
async function commit(action, payload, label, after = load) {
  try {
    await ElMessageBox.confirm(label + '？', '确认操作', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  if (await save(() => action(payload), '操作已保存')) await after()
}
function modeCommand(command) {
  if (command === 'new' || command === 'edit')
    return dialog.value.open('mode', command === 'edit' ? mode.value : null)
  const { id, revision, label } = mode.value
  if (command === 'default')
    return commit(defaultMode, { id, revision }, '将 ' + label + ' 设为默认模式')
  return commit(deleteMode, { id, revision }, '删除模式 ' + label, () =>
    route.query.mode ? router.replace({ path: route.path }) : load(),
  )
}
function versionCommand(command) {
  if (command === 'new' || command === 'edit')
    return dialog.value.open('version', command === 'edit' ? version.value : null)
  const { id, revision, name, status } = version.value
  if (command === 'default')
    return commit(defaultVersion, { id, revision }, '将 ' + name + ' 设为默认版本')
  if (command === 'delete')
    return commit(deleteVersion, { id, revision }, '删除版本 ' + name, () =>
      route.query.version ? selectMode(mode.value.id) : load(),
    )
  return status === 'published'
    ? commit(withdrawVersion, { id, revision }, '停止公开版本 ' + name)
    : commit(publishVersion, { id, revision }, '公开版本 ' + name + ' 中已发布的文档')
}
async function groupCommand(command) {
  const { id, revision, title } = group.value
  if (command === 'edit') return dialog.value.open('group', group.value)
  const payload = { id, revision, version: version.value.id }
  if (command === 'delete') return commit(deleteGroup, payload, '删除分组 ' + title)
  if (
    await save(
      () => moveGroup({ ...payload, direction: command === 'up' ? -1 : 1 }),
      '分组顺序已更新',
    )
  )
    await load()
}
function removeDocument(row) {
  return commit(
    deleteDocument,
    { version: version.value.id, id: row.id, revision: row.revision },
    '删除文档 ' + row.title,
  )
}
</script>
<style scoped lang="scss" src="./workspace.scss"></style>
