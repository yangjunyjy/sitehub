export function systemPayload(form, originalCode = "") {
  const code = String(form.code || "").trim();
  const name = String(form.name || "").trim();
  if (!/^[a-z][a-z0-9-]{0,31}$/.test(code))
    throw new Error("系统代码须为小写字母开头的 1–32 位字母、数字或连字符");
  if (originalCode && code !== originalCode)
    throw new Error("已有系统的代码不能修改");
  if (!name) throw new Error("请填写系统名称");
  const origin = String(form.origin || "")
    .trim()
    .replace(/\/$/, "");
  if (origin) {
    let url;
    try {
      url = new URL(origin);
    } catch {
      throw new Error("请输入有效的系统访问地址");
    }
    if (
      !["http:", "https:"].includes(url.protocol) ||
      url.username ||
      url.password ||
      url.origin !== origin
    )
      throw new Error("系统访问地址须为 HTTP/HTTPS 源地址，不包含路径或账号");
  }
  return {
    ...(form.id ? { id: form.id } : {}),
    code,
    name,
    description: String(form.description || "").trim(),
    origin,
    enabled: !!form.enabled,
    public: !!form.public,
  };
}
