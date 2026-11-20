<script setup>
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { messages, sendMessage } from "@/api";
import { uint64 } from "@/utils/native-api.mjs";
import { useResource } from "@/composables/useResource";
const props = defineProps({ articleId: String });
const emit = defineEmits(['latest']);
const label = computed(() => (props.articleId ? "评论" : "留言"));
const page = ref(1),
  editing = ref(false),
  posting = ref(false),
  success = ref(""),
  failure = ref("");
const heading = ref(null),
  input = ref(null),
  opener = ref(null);
const draft = reactive({ nickname: "", content: "" });
const size = 10;
const feed = useResource(async () => {
  const requestedPage = page.value;
  const response = await messages({ page: requestedPage, pageSize: size, articleId: props.articleId });
  if (!props.articleId && requestedPage === 1) emit('latest', response.data.items || []);
  return response;
});
const items = computed(() => feed.data.value?.items || []);
const pages = computed(() =>
  Math.max(1, Math.ceil((feed.data.value?.total || 0) / size)),
);
async function turn(next) {
  page.value = next;
  await feed.load();
  heading.value?.focus({ preventScroll: true });
  heading.value?.scrollIntoView({ block: "start" });
}
async function compose() {
  editing.value = !editing.value;
  await nextTick();
  if (editing.value) input.value?.focus();
  else opener.value?.focus();
}
async function submit() {
  if (posting.value || !draft.content.trim()) return;
  posting.value = true;
  failure.value = "";
  success.value = "";
  try {
    const result = await sendMessage({
      nickname: draft.nickname.trim(),
      content: draft.content.trim(),
      ...(props.articleId && { articleId: uint64(props.articleId) }),
    });
    if (result.data?.status !== "pending")
      throw new Error("提交状态未确认，请重试");
    success.value = props.articleId
      ? "评论已提交，审核通过后会显示在本文下方。"
      : "你的星光已送达，审核通过后会出现在这片星空。";
    draft.content = "";
    editing.value = false;
    await nextTick();
    opener.value?.focus();
  } catch (err) {
    failure.value = err.message;
  } finally {
    posting.value = false;
  }
}
onMounted(() => feed.load());
</script>
<template>
  <section class="discussion" :aria-label="label + '列表'">
    <header>
      <h2 ref="heading" tabindex="-1">
        {{ articleId ? "文章评论" : "访客留言" }}
      </h2>
      <span>最新在前</span>
    </header>
    <p v-if="feed.loading.value" role="status">正在读取{{ label }}…</p>
    <p v-else-if="feed.error.value" role="alert">
      {{ feed.error.value }} <button @click="feed.load()">重试</button>
    </p>
    <ol v-else-if="items.length">
      <li v-for="item in items" :key="item.id">
        <strong>{{ item.nickname || "匿名访客" }}</strong
        ><time v-if="item.createdAt" :datetime="item.createdAt">{{
          new Date(item.createdAt).toLocaleDateString("zh-CN")
        }}</time>
        <p>{{ item.content }}</p>
      </li>
    </ol>
    <p v-else>还没有{{ label }}，欢迎写下第一条。</p>
    <nav v-if="pages > 1" :aria-label="label + '分页'">
      <button
        :disabled="page <= 1 || feed.loading.value"
        @click="turn(page - 1)"
      >
        上一页</button
      ><span>{{ page }} / {{ pages }}</span
      ><button
        :disabled="page >= pages || feed.loading.value"
        @click="turn(page + 1)"
      >
        下一页
      </button>
    </nav>
    <p v-if="success" role="status">{{ success }}</p>
    <form v-if="editing" @submit.prevent="submit">
      <label
        >昵称（选填）<input
          v-model="draft.nickname"
          maxlength="40"
          autocomplete="nickname" /></label
      ><label
        >{{ label }}内容<textarea
          ref="input"
          v-model="draft.content"
          :aria-label="label + '内容'"
          required
          maxlength="1000"
          rows="4"
        />
      </label>
      <p v-if="failure" role="alert">{{ failure }}</p>
      <div>
        <button type="button" :disabled="posting" @click="compose">取消</button
        ><button class="primary" :disabled="posting || !draft.content.trim()">
          {{ posting ? "提交中…" : "发送" + label }}
        </button>
      </div>
    </form>
    <button v-else ref="opener" class="primary compose" @click="compose">
      写{{ label }}
    </button>
  </section>
</template>
<style scoped>
.discussion {
  margin: 0 auto;
  width: min(100%, 860px);
  padding: clamp(24px, 4vw, 50px);
  border: 1px solid #fff;
  border-radius: 18px;
  background: #fffc;
  box-shadow: 0 12px 46px #61547b10;
  color: var(--ink);
}
header,
nav,
form div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
h2 {
  font-size: 23px;
  margin: 0;
  scroll-margin-top: 82px;
}
header span,
time,
nav span {
  color: var(--muted);
  font-size: 13px;
}
ol {
  padding: 0;
  list-style: none;
}
li {
  padding: 18px 0 8px;
  border-top: 1px solid var(--line);
  overflow-wrap: anywhere;
}
li p {
  white-space: pre-wrap;
}
time {
  margin-left: 12px;
}
button {
  border: 1px solid var(--line);
  padding: 7px 14px;
  border-radius: 9px;
  background: white;
}
button:disabled {
  opacity: 0.5;
}
nav,
form div {
  justify-content: center;
}
.primary {
  background: #62587c;
  border-color: #62587c;
  color: white;
}
.compose {
  display: block;
  margin: 24px auto 0;
}
form {
  margin-top: 24px;
}
label {
  display: grid;
  gap: 5px;
  margin: 12px 0;
  font-size: 14px;
}
input,
textarea {
  width: 100%;
  padding: 9px;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: #fff;
  color: inherit;
  font: inherit;
}
textarea {
  resize: vertical;
}
</style>
