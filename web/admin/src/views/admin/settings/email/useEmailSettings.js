import { reactive, ref } from "vue";
import { getEmailSettings, updateEmailSettings } from "@/api/settings";
import {
  useAdminResource,
  useUnsavedChanges,
} from "@/composables/useAdminResource";

export function useEmailSettings() {
  const { data, loading, saving, error, load, save } =
    useAdminResource(getEmailSettings);
  const form = reactive({
      smtpHost: "",
      smtpPort: 465,
      fromEmail: "",
      fromName: "",
      subject: "",
      password: "",
    }),
    baseline = ref(""),
    formRef = ref();
  const rules = {
    smtpHost: [
      { required: true, message: "请输入 SMTP 主机", trigger: "blur" },
    ],
    smtpPort: [{ required: true, message: "请输入端口", trigger: "blur" }],
    fromEmail: [
      { required: true, message: "请输入发件邮箱", trigger: "blur" },
      { type: "email", message: "请输入有效的邮箱地址", trigger: "blur" },
    ],
    fromName: [{ required: true, message: "请输入发件名称", trigger: "blur" }],
    subject: [{ required: true, message: "请输入邮件主题", trigger: "blur" }],
    password: [
      {
        validator: (_, value, callback) =>
          !value && !data.value?.hasPassword
            ? callback(new Error("请输入 SMTP 授权码"))
            : callback(),
        trigger: "blur",
      },
    ],
  };
  const confirmLeave = useUnsavedChanges(
    () => !!baseline.value && JSON.stringify(form) !== baseline.value,
  );
  function reset() {
    for (const key in form)
      form[key] = key === "password" ? "" : (data.value[key] ?? "");
    baseline.value = JSON.stringify(form);
    formRef.value?.clearValidate();
  }
  async function initialize() {
    await load();
    if (!error.value) reset();
  }
  async function reload() {
    if (await confirmLeave()) await initialize();
  }
  async function submit() {
    if (!(await formRef.value?.validate().catch(() => false))) return;
    await save(async () => {
      const payload = { ...form };
      if (!payload.password) delete payload.password;
      data.value = (await updateEmailSettings(payload)).data;
      reset();
    }, "邮件配置已保存，下次发信生效");
  }
  initialize();
  return {
    data,
    loading,
    saving,
    error,
    form,
    formRef,
    rules,
    reset,
    reload,
    submit,
  };
}
