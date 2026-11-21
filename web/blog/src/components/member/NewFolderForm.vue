<script setup>
import { ref } from "vue";
import { createFolder } from "@/api/library";
import { useMember } from "@/composables/useMember";

const emit = defineEmits(["created"]);
const { expired } = useMember();
const name = ref(""),
  busy = ref(false),
  error = ref("");
async function submit() {
  const title = name.value.trim();
  if (!title || busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    const folder = await createFolder(title);
    name.value = "";
    emit("created", folder);
  } catch (err) {
    if (err.status === 401) expired();
    error.value = err.message;
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <form class="new-folder" @submit.prevent="submit">
    <label
      >新收藏夹名称<input
        v-model="name"
        required
        pattern=".{1,50}" title="收藏夹名称最多 50 个字符"
        placeholder="例如，技术笔记"
    /></label>
    <button :disabled="busy || !name.trim()">
      {{ busy ? "创建中…" : "创建收藏夹" }}
    </button>
    <p v-if="error" role="alert">{{ error }}</p>
  </form>
</template>
<style scoped>
.new-folder {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}
.new-folder label {
  display: grid;
  gap: 4px;
  min-width: 120px;
  flex: 1;
  font-size: 12px;
  color: #516058;
}
.new-folder input,
.new-folder button {
  border: 1px solid #c6d0c5;
  border-radius: 6px;
  padding: 8px 10px;
  background: white;
  font-size: 13px;
  min-width: 0;
}
.new-folder button {
  color: #294c3c;
  font-weight: 600;
}
.new-folder p {
  width: 100%;
  color: #993f31;
  font-size: 13px;
}
</style>
