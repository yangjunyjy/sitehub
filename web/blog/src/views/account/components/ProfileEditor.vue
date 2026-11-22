<script setup>
import { reactive, ref, onMounted } from "vue";
import MemberAvatar from "@/components/member/MemberAvatar.vue";
import { useMember } from "@/composables/useMember";

const props = defineProps({ user: { type: Object, required: true } });
const emit = defineEmits(["finish"]);
const { saveProfile } = useMember();
const draft = reactive({
  username: props.user.username || "",
  nickname: props.user.nickname || "",
  avatar: props.user.avatar || "",
});
const error = ref("");
const busy = ref(false);
const firstInput = ref(null);
onMounted(() => firstInput.value?.focus({ preventScroll: true }));
async function submit() {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    await saveProfile({ ...draft });
    emit("finish", true);
  } catch (failure) {
    error.value = failure.message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <form
    class="profile-editor"
    aria-label="编辑个人资料"
    @submit.prevent="submit"
  >
    <div class="fields">
      <label
        >用户名<input
          ref="firstInput"
          v-model.trim="draft.username"
          autocomplete="username"
          pattern="[A-Za-z0-9_]{3,40}"
          maxlength="40"
          required
          placeholder="3–40 位字母、数字或下划线"
      /></label>
      <label
        >昵称<input
          v-model.trim="draft.nickname"
          autocomplete="nickname"
          maxlength="40"
          required
      /></label>
      <label class="avatar-field"
        >头像图片地址
        <span
          ><MemberAvatar :user="draft" /><input
            v-model.trim="draft.avatar"
            type="url"
            maxlength="1024"
            pattern="https?://.+"
            placeholder="https://…"
        /></span>
      </label>
    </div>
    <p v-if="error" role="alert">{{ error }}</p>
    <footer>
      <button type="button" :disabled="busy" @click="emit('finish', false)">
        取消
      </button>
      <button type="submit" :disabled="busy">
        {{ busy ? "保存中…" : "保存资料" }}
      </button>
    </footer>
  </form>
</template>
<style scoped>
.profile-editor {
  margin-top: 18px;
  padding: clamp(16px, 3vw, 28px);
  background: linear-gradient(125deg, #ffffffb8, #f3eff4a8);
  border: 1px solid var(--journal-rule);
  border-radius: 12px;
  backdrop-filter: blur(16px);
}
.fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
label {
  display: grid;
  gap: 7px;
  font-weight: 550;
  font-size: 13px;
}
input {
  width: 100%;
  min-width: 0;
  padding: 10px 12px;
  font: inherit;
  color: var(--journal-ink);
  background: #ffffffc4;
  border: 1px solid var(--journal-rule);
  border-radius: 6px;
}
.avatar-field {
  grid-column: 1 / -1;
}
.avatar-field > span,
footer {
  display: flex;
  align-items: center;
  gap: 10px;
}
footer {
  justify-content: flex-end;
  margin-top: 18px;
}
button {
  padding: 8px 15px;
  color: var(--journal-muted);
  border: 1px solid var(--journal-rule);
  background: transparent;
  border-radius: 6px;
}
button[type="submit"] {
  background: #615368;
  color: white;
  border-color: #615368;
}
[role="alert"] {
  font-size: 13px;
  color: #963e35;
}
@media (max-width: 550px) {
  .fields {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
