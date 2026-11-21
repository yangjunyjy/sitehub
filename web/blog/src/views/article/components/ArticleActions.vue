<script setup>
import { onMounted, ref, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
import ArticleActionDock from "./ArticleActionDock.vue";
import NewFolderForm from "@/components/member/NewFolderForm.vue";
import {
  getInteraction,
  getFolders,
  setFavorite,
  setLike,
} from "@/api/library";
import { useMember } from "@/composables/useMember";

const props = defineProps({ articleId: { type: String, required: true } });
const router = useRouter(),
  route = useRoute(),
  { member, restore, expired } = useMember();
const state = ref(null),
  busy = ref(false),
  error = ref(""),
  dialog = ref(null),
  favoriteButton = ref(null),
  modalOpen = ref(false),
  folders = ref([]),
  selected = ref("0"),
  newFolder = ref(false);
const loadingFolders = ref(false),
  folderError = ref("");
const login = () =>
  router.push({ name: "login", query: { redirect: route.fullPath } });
async function requireMember() {
  await restore();
  if (member.user) return true;
  await login();
  return false;
}
async function handle(err) {
  if (err.status === 401) {
    expired();
    dialog.value?.close();
    await login();
  } else if (dialog.value?.open) folderError.value = err.message;
  else error.value = err.message;
}
async function load() {
  error.value = "";
  try {
    state.value = await getInteraction(props.articleId);
  } catch (err) {
    error.value = err.message;
  }
}
async function like() {
  if (busy.value || !state.value) return;
  busy.value = true;
  error.value = "";
  try {
    if (await requireMember())
      state.value = await setLike(props.articleId, !state.value.liked);
  } catch (err) {
    await handle(err);
  } finally {
    busy.value = false;
  }
}
async function choose() {
  if (!state.value || dialog.value?.open || busy.value) return;
  error.value = "";
  folderError.value = "";
  newFolder.value = false;
  folders.value = [];
  try {
    if (!(await requireMember())) return;
    selected.value = String(state.value.folderId || "0");
    modalOpen.value = true;
    await nextTick();
    dialog.value.showModal();
    await loadFolders();
  } catch (err) {
    await handle(err);
  }
}
async function loadFolders() {
  loadingFolders.value = true;
  folderError.value = "";
  try {
    folders.value = await getFolders();
  } catch (err) {
    await handle(err);
  } finally {
    loadingFolders.value = false;
  }
}
function created(folder) {
  folders.value.push(folder);
  selected.value = String(folder.id);
  newFolder.value = false;
}
async function save(enabled) {
  if (busy.value || loadingFolders.value) return;
  busy.value = true;
  folderError.value = "";
  try {
    state.value = await setFavorite(props.articleId, enabled, selected.value);
    dialog.value.close();
  } catch (err) {
    await handle(err);
  } finally {
    busy.value = false;
  }
}
async function dialogClosed() {
  modalOpen.value = false;
  await nextTick();
  if (document.activeElement === document.body && favoriteButton.value?.isConnected)
    favoriteButton.value.focus({preventScroll:true});
}
onMounted(load);
</script>
<template>
  <ArticleActionDock :modal-open="modalOpen">
    <button
      type="button"
      :aria-pressed="!!state?.liked"
      :disabled="busy || !state"
      @click="like"
    >
      <BlogIcon name="heart" :size="17" />{{ state?.liked ? "已点赞" : "点赞" }}
      <span v-if="state">{{ state.likes }}</span>
    </button>
    <button
      type="button"
      ref="favoriteButton"
      :aria-pressed="!!state?.favorited"
      :disabled="busy || !state"
      @click="choose"
    >
      <BlogIcon name="bookmark" :size="16" />{{
        state?.favorited ? "已收藏" : "收藏"
      }}
      <span v-if="state">{{ state.favorites }}</span>
    </button>
    <p v-if="error" role="alert">
      {{ error }} <button v-if="!state" @click="load">重试</button>
    </p>
  </ArticleActionDock>
  <dialog
    ref="dialog"
    aria-labelledby="folder-dialog-title"
    @close="dialogClosed"
    @cancel="busy && $event.preventDefault()"
  >
    <header>
      <h2 id="folder-dialog-title">收藏到文件夹</h2>
      <button
        class="close"
        type="button"
        aria-label="关闭"
        :disabled="busy"
        @click="dialog.close()"
      >
        <BlogIcon name="close" :size="18" />
      </button>
    </header>
    <p v-if="loadingFolders" role="status">正在读取收藏夹…</p>
    <div v-else class="choices">
      <label v-for="folder in folders" :key="folder.id">
        <input
          v-model="selected"
          type="radio"
          name="folder"
          :value="String(folder.id)"
        />
        <BlogIcon name="folder" :size="17" /><strong>{{ folder.name }}</strong
        ><small>{{ folder.count }} 篇</small>
      </label>
    </div>
    <NewFolderForm v-if="newFolder" @created="created" />
    <button
      v-else-if="folders.length"
      class="new"
      type="button"
      @click="newFolder = true"
    >
      ＋ 新建收藏夹
    </button>
    <p v-if="folderError" role="alert">
      {{ folderError }}
      <button v-if="!folders.length" @click="loadFolders">重试</button>
    </p>
    <footer>
      <button
        v-if="state?.favorited"
        type="button"
        class="remove"
        :disabled="busy"
        @click="save(false)"
      >
        取消收藏
      </button>
      <button
        type="button"
        class="save"
        :disabled="busy || !folders.some((f) => String(f.id) === selected)"
        @click="save(true)"
      >
        {{ busy ? "保存中…" : "保存收藏" }}
      </button>
    </footer>
  </dialog>
</template>
<style scoped>
button {
  background: white;
  color: #32473b;
  border: 1px solid #c9d2c8;
  border-radius: 6px;
  font-size: 14px;
  padding: 7px 13px;
}
button[aria-pressed] {
  justify-content: flex-start;
  border: 0;
  background: transparent;
  min-height: 40px;
  width: 100%;
  padding: 6px 9px;
  display: inline-flex;
  gap: 7px;
  align-items: center;
}
button[aria-pressed] span {
  margin-left: auto;
}
button[aria-pressed="true"] {
  background: #e8f1e7;
  border-color: #adc7ad;
  color: #264f30;
}
span,
small {
  font-variant-numeric: tabular-nums;
  color: #5d6c61;
}
dialog {
  color: #28352c;
  border: 1px solid #c9d2c8;
  border-radius: 13px;
  padding: 24px;
  width: min(400px, calc(100% - 32px));
  max-height: 82dvh;
}
dialog::backdrop {
  background: #14251a63;
}
header,
footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
header h2 {
  margin: 0;
  font-size: 22px;
}
.close,
.new,
.remove {
  border: 0;
  padding-inline: 3px;
}
.choices {
  margin: 18px 0;
  max-height: 40dvh;
  overflow: auto;
}
.choices label {
  padding: 12px 8px;
  display: flex;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid #e1e7df;
  cursor: pointer;
}
.choices label:has(:checked) {
  background: #f2f6ef;
}
input {
  accent-color: #31593f;
}
.choices strong {
  flex: 1;
  overflow-wrap: anywhere;
  font-size: 14px;
}
footer {
  border-top: 1px solid #dce4d9;
  margin-top: 18px;
  padding-top: 16px;
}
.save {
  margin-left: auto;
  background: #284d38;
  color: white;
  border-color: #284d38;
}
[role="alert"] {
  width: 100%;
  font-size: 13px;
  color: #983b3b;
}
button {
  transition:
    background 0.16s,
    color 0.16s,
    border-color 0.16s;
}
button[aria-pressed]:hover:enabled,
.choices label:hover {
  background: #f0f5ed;
}
dialog[open] {
  animation: folder-enter 0.17s ease-out;
}
@keyframes folder-enter {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
