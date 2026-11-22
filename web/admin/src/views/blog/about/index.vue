<script setup>
import { computed, ref } from "vue";
import AdminPage from "@/components/AdminPage.vue";
import AdminFormSection from "@/components/admin/AdminFormSection.vue";
import {
  useAdminResource,
  useUnsavedChanges,
} from "@/composables/useAdminResource";
import request from "@/utils/request";

const { data, loading, saving, error, load, save } = useAdminResource(() =>
  request.get("/blog/about"),
);
const form = ref(null);
const baseline = ref("");
const formRef = ref(null);
const dirty = computed(
  () => form.value && JSON.stringify(form.value) !== baseline.value,
);
useUnsavedChanges(() => dirty.value);
function reset() {
  const profile = data.value;
  form.value = Object.fromEntries(
    ["name", "title", "bio", "content", "location", "email", "avatar"].map(
      (key) => [key, profile[key] || ""],
    ),
  );
  form.value.skills = [...(profile.skills || [])];
  form.value.socials = (profile.socials || []).map(({ label, url }) => ({
    label,
    url,
  }));
  baseline.value = JSON.stringify(form.value);
  formRef.value?.clearValidate();
}
async function initialize() {
  await load();
  if (!error.value && data.value) reset();
}
function validURL(value, allowPath) {
  if (/[\\\s]/.test(value)) return false;
  if (allowPath && /^\/(?!\/)/.test(value)) return true;
  try {
    const url = new URL(value);
    return (
      ["http:", "https:"].includes(url.protocol) &&
      !!url.hostname &&
      !url.username &&
      !url.password
    );
  } catch {
    return false;
  }
}
const linkRule = {
  validator: (_, value, done) =>
    done(
      validURL(value, false)
        ? undefined
        : new Error("请填写完整的 HTTP(S) 网址"),
    ),
  trigger: "blur",
};
const avatarRule = {
  validator: (_, value, done) =>
    done(
      !value || validURL(value, true)
        ? undefined
        : new Error("请填写站内路径或完整的图片网址"),
    ),
  trigger: "blur",
};
async function submit() {
  if (
    saving.value ||
    !dirty.value ||
    !(await formRef.value.validate().catch(() => false))
  )
    return;
  await save(async () => {
    const fields = {
      ...form.value,
      skills: [...new Set(form.value.skills.map((s) => s.trim()))],
      socials: form.value.socials.map(({ label, url }) => ({
        label: label.trim(),
        url: url.trim(),
      })),
    };
    await request.post("/blog/about/settings", fields, { silentError: true });
    data.value = fields;
    reset();
  }, "关于页已保存");
}
initialize();
</script>
<template>
  <AdminPage title="关于页设置" :loading="loading" :error="error">
    <template #actions>
      <el-button v-if="error" @click="initialize">重试</el-button>
      <template v-else>
        <el-button :disabled="!dirty || saving" @click="reset"
          >还原修改</el-button
        >
        <el-button
          type="primary"
          :disabled="!dirty"
          :loading="saving"
          @click="submit"
          >保存关于页</el-button
        >
      </template>
    </template>
    <el-form
      v-if="form"
      ref="formRef"
      class="about-settings"
      label-position="top"
      :model="form"
      :disabled="saving"
      @submit.prevent="submit"
    >
      <AdminFormSection title="博主资料">
        <div class="fields">
          <el-form-item label="博主名称"
            ><el-input v-model.trim="form.name" maxlength="80"
          /></el-form-item>
          <el-form-item label="身份或头衔"
            ><el-input v-model.trim="form.title" maxlength="120"
          /></el-form-item>
          <el-form-item label="所在地"
            ><el-input v-model.trim="form.location" maxlength="100"
          /></el-form-item>
          <el-form-item
            label="公开邮箱"
            prop="email"
            :rules="[
              { type: 'email', message: '请填写有效的邮箱', trigger: 'blur' },
            ]"
            ><el-input v-model.trim="form.email" maxlength="254"
          /></el-form-item>
        </div>
        <el-form-item label="头像或配图" prop="avatar" :rules="[avatarRule]"
          ><el-input
            v-model.trim="form.avatar"
            maxlength="500"
            placeholder="站内图片路径或 https://…"
        /></el-form-item>
        <el-form-item label="简介"
          ><el-input
            v-model.trim="form.bio"
            type="textarea"
            :rows="2"
            maxlength="600"
            show-word-limit
        /></el-form-item>
        <el-form-item label="自述"
          ><el-input
            v-model.trim="form.content"
            type="textarea"
            :rows="6"
            maxlength="12000"
            show-word-limit
        /></el-form-item>
      </AdminFormSection>
      <AdminFormSection title="兴趣与联系">
        <el-form-item
          label="喜欢与擅长"
          prop="skills"
          :rules="[
            {
              validator: (_, tags, done) =>
                done(
                  tags.length > 32 ||
                    tags.some((t) => !t.trim() || [...t.trim()].length > 40)
                    ? new Error('最多 32 个标签，每个 1 至 40 个字符')
                    : undefined,
                ),
              trigger: 'change',
            },
          ]"
        >
          <el-input-tag
            v-model="form.skills"
            :max="32"
            placeholder="输入标签后按回车"
          />
        </el-form-item>
        <div v-for="(link, i) in form.socials" :key="i" class="social-row">
          <el-form-item
            label="链接名称"
            :prop="'socials.' + i + '.label'"
            :rules="[
              { required: true, message: '请填写名称', trigger: 'blur' },
            ]"
            ><el-input v-model.trim="link.label" maxlength="60"
          /></el-form-item>
          <el-form-item
            label="链接网址"
            :prop="'socials.' + i + '.url'"
            :rules="[linkRule]"
            ><el-input
              v-model.trim="link.url"
              maxlength="500"
              placeholder="https://…"
          /></el-form-item>
          <el-button
            text
            type="danger"
            :aria-label="'移除第 ' + (i + 1) + ' 个链接'"
            @click="form.socials.splice(i, 1)"
            >移除</el-button
          >
        </div>
        <el-button
          :disabled="form.socials.length >= 16"
          @click="form.socials.push({ label: '', url: '' })"
          >添加社交链接</el-button
        >
      </AdminFormSection>
    </el-form>
  </AdminPage>
</template>
<style lang="scss" scoped>
.about-settings {
  display: grid;
  gap: 16px;
  max-width: 1000px;
}
.fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
}
.social-row {
  display: grid;
  grid-template-columns: minmax(110px, 1fr) minmax(180px, 2fr) auto;
  gap: 12px;
  align-items: start;
}
.social-row > button {
  margin-top: 28px;
}
@media (max-width: 700px) {
  .fields,
  .social-row {
    grid-template-columns: minmax(0, 1fr);
  }
  .social-row {
    gap: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
    margin-bottom: 16px;
  }
  .social-row > button {
    justify-self: end;
    margin: 0 0 12px;
  }
}
</style>
