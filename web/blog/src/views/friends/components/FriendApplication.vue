<script setup>
import { reactive, ref } from "vue";
import { applyFriend } from "@/api";
const form = reactive({ name: "", url: "", description: "", avatar: "" }),
  submitting = ref(false),
  error = ref(""),
  success = ref("");
async function submit() {
  if (submitting.value) return;
  error.value = "";
  success.value = "";
  submitting.value = true;
  try {
    const { data } = await applyFriend(
      Object.fromEntries(
        Object.entries(form).map(([key, value]) => [key, value.trim()]),
      ),
    );
    if (data?.status !== "pending")
      throw new Error("申请状态尚未确认，请稍后查看或联系博主");
    success.value = "申请已提交，正在等待博主审核。感谢你的相遇！";
    Object.assign(form, { name: "", url: "", description: "", avatar: "" });
  } catch (failure) {
    error.value = failure.message;
  } finally {
    submitting.value = false;
  }
}
</script>
<template>
  <section id="apply" class="friend-application">
    <div class="friend-application__heading">
      <span>LET'S BE FRIENDS</span>
      <h2>交换一份相遇</h2>
      <p>如果你也在认真记录生活，欢迎交换友链。提交后将由博主审核。</p>
    </div>
    <form class="form-grid" @submit.prevent="submit">
      <label class="field"
        >网站名称<input
          v-model="form.name"
          required
          maxlength="80"
          placeholder="你的站点叫什么名字？" /></label
      ><label class="field"
        >网站地址<input
          v-model="form.url"
          type="url"
          required
          maxlength="500"
          placeholder="https://your-blog.com" /></label
      ><label class="field full"
        >站点介绍<input
          v-model="form.description"
          required
          maxlength="240"
          placeholder="用一句话介绍你的世界" /></label
      ><label class="field full"
        >头像地址 <small>选填</small
        ><input
          v-model="form.avatar"
          type="url"
          maxlength="500"
          placeholder="https://your-blog.com/avatar.png"
      /></label>
      <p v-if="error" class="form-error full" role="alert">{{ error }}</p>
      <p v-if="success" class="form-success full" role="status">
        {{ success }}
      </p>
      <div class="full form-footer">
        <small>请确保网站可正常访问，内容友善且持续更新。</small
        ><button class="button" :disabled="submitting">
          {{ submitting ? "正在提交…" : "提交友链申请 ↗" }}
        </button>
      </div>
    </form>
  </section>
</template>
<style lang="scss" scoped>
.friend-application {
  padding: 40px;
  border: 1px solid var(--line);
  border-radius: 6px;
  margin-top: 55px;
  &__heading {
    margin-bottom: 30px;
    > span {
      font-size: 9px;
      letter-spacing: 2px;
      color: #b09dbd;
    }
    h2 {
      font-size: 23px;
      font-weight: 400;
      letter-spacing: 2px;
      margin: 12px 0;
    }
    p {
      font-size: 12px;
      color: var(--muted);
      line-height: 2;
    }
  }
  .field small {
    font-size: 10px;
    color: #bfb7c7;
  }
}
.form-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-top: 10px;
  > small {
    font-size: 10px;
    color: var(--muted);
  }
}
@media (max-width: 640px) {
  .friend-application {
    padding: 25px;
  }
  .form-footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
