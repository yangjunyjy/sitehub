<script setup>
import { computed, nextTick, ref } from "vue";
import { useRoute } from "vue-router";
import { sendMessage } from "@/api";
import { uint64 } from "@/utils/native-api.mjs";
import { useMember } from "@/composables/useMember";
import BlogIcon from "@/components/BlogIcon.vue";
import { emoji } from "./emoji";

const props = defineProps({ articleId: String });
const label = computed(() => (props.articleId ? "评论" : "留言"));
const route = useRoute(),
  { member, expired } = useMember();
const editing = ref(false),
  palette = ref(false),
  text = ref(""),
  posting = ref(false),
  error = ref(""),
  notice = ref("");
const input = ref(null),
  trigger = ref(null),
  emojiButton = ref(null);
async function open() {
  editing.value = true;
  notice.value = error.value = "";
  await nextTick();
  input.value?.focus();
}
async function close() {
  editing.value = palette.value = false;
  await nextTick();
  trigger.value?.focus();
}
async function insert(symbol) {
  const start = input.value.selectionStart,
    end = input.value.selectionEnd;
  const value = text.value.slice(0, start) + symbol + text.value.slice(end);
  if (value.length > 1000) return;
  text.value = value;
  palette.value = false;
  await nextTick();
  input.value.focus();
  input.value.setSelectionRange(start + symbol.length, start + symbol.length);
}
async function submit() {
  if (posting.value || !text.value.trim() || !member.user) return;
  posting.value = true;
  error.value = notice.value = "";
  try {
    const result = await sendMessage({
      content: text.value.trim(),
      ...(props.articleId && { articleId: uint64(props.articleId) }),
    });
    if (result.data?.status !== "pending")
      throw new Error("提交状态未确认，请重试");
    notice.value = label.value + "已提交，审核通过后公开显示。";
    text.value = "";
    await close();
  } catch (err) {
    if (err.status === 401) expired();
    error.value = err.message;
  } finally {
    posting.value = false;
  }
}
</script>
<template>
  <div class="composer">
    <p v-if="notice" class="success" role="status">{{ notice }}</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div v-if="!member.user" class="signin">
      <BlogIcon name="message" :size="20" />
      <RouterLink :to="{ name: 'login', query: { redirect: route.fullPath } }"
        >登录后写{{ label }} →</RouterLink
      >
    </div>
    <button
      v-else-if="!editing"
      ref="trigger"
      type="button"
      class="start"
      @click="open"
    >
      <BlogIcon name="message" :size="17" />写{{ label }}
    </button>
    <form v-else @submit.prevent="submit">
      <label :for="'comment-' + (articleId || 'board')">{{
        member.user.nickname || member.user.username || "我"
      }}</label>
      <textarea
        :id="'comment-' + (articleId || 'board')"
        ref="input"
        v-model="text"
        :aria-label="label + '内容'"
        rows="4"
        maxlength="1000"
        required
        placeholder="写下你的想法…"
        :readonly="posting"
      />
      <div class="tools">
        <button
          ref="emojiButton"
          class="emoji-button"
          type="button"
          aria-label="选择表情"
          :aria-expanded="palette"
          @click="palette = !palette"
        >
          <BlogIcon name="smile" :size="19" />表情
        </button>
        <small>{{ text.length }}/1000</small>
        <button class="cancel" type="button" :disabled="posting" @click="close">
          取消
        </button>
        <button class="send" :disabled="posting || !text.trim()">
          {{ posting ? "发送中…" : "发送" + label }}
        </button>
      </div>
      <div
        v-if="palette"
        class="palette"
        role="group"
        aria-label="表情包"
        @keydown.esc.stop="
          palette = false;
          emojiButton.focus();
        "
      >
        <button
          v-for="item in emoji"
          :key="item.symbol"
          type="button"
          :title="item.name"
          :aria-label="item.name"
          @click="insert(item.symbol)"
        >
          <img :src="item.src" :alt="item.symbol" width="26" height="26" />
        </button>
        <a href="/licenses/twemoji.txt" target="_blank" rel="noopener"
          >Twemoji · CC BY 4.0</a
        >
      </div>
    </form>
  </div>
</template>
<style scoped>
.composer {
  margin: 20px 0 24px;
}
form {
  border: 1px solid #bdc9c1;
  border-radius: 9px;
  background: #fff;
  padding: 14px;
}
label {
  font-weight: 600;
  font-size: 13px;
  color: #324239;
}
textarea {
  width: 100%;
  display: block;
  min-height: 84px;
  margin: 8px 0;
  border: 0;
  color: #25342c;
  resize: vertical;
  line-height: 1.7;
}
.tools,
.signin,
.start,
.emoji-button {
  display: flex;
  align-items: center;
  gap: 8px;
}
.tools {
  border-top: 1px solid #e2e6df;
  padding-top: 10px;
}
small {
  margin-right: auto;
  color: #69776f;
  font-size: 11px;
}
button {
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  padding: 6px 9px;
  font-size: 13px;
}
.send,
.start {
  background: #294c3c;
  color: white;
}
.start {
  padding: 8px 15px;
}
.signin {
  padding: 14px;
  background: #f4f6f2;
  border: 1px solid #e0e5de;
  border-radius: 7px;
  color: #395b48;
  font-size: 14px;
}
.palette {
  border-top: 1px solid #e1e5df;
  padding: 12px 0 0;
  display: grid;
  grid-template-columns: repeat(6, 36px);
  gap: 3px;
}
.palette button {
  padding: 5px;
  line-height: 0;
}
.palette button:hover,
.emoji-button:hover,
.cancel:hover {
  background: #edf3ec;
}
.palette a {
  grid-column: 1/-1;
  font-size: 10px;
  color: #68776e;
  margin-top: 5px;
}
.success,
.error {
  font-size: 13px;
  color: #376646;
}
.error {
  color: #993f31;
}
button {
  transition:
    background 0.16s,
    color 0.16s;
}
form,
.palette {
  animation: composer-enter 0.15s ease-out;
}
@keyframes composer-enter {
  from {
    opacity: 0;
    transform: translateY(3px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
