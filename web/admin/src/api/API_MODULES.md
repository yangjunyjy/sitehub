# 管理后台原生 API

统一前缀 `/v1`，成功响应 `code: 0`。模块直接调用本项目 `service/*/api/*.api` 定义的接口。

| 模块 | 业务前缀 |
| --- | --- |
| auth、permission、sliderCaptcha、settings | `/rbac/auth`、`/rbac/captcha`、`/rbac/settings` |
| user、role、menu | `/rbac/users`、`/rbac/roles`、`/rbac/menus` |
| article | `/blog/admin/articles`、`/blog/assets/upload` |
| category、tag、statistics | `/blog/categories`、`/blog/tags`、`/blog/statistics` |
| documents | `/product/admin/docs`、`/product/docs` |
| website | `/product/content`、`/product/downloads`、`/product/admin/downloads`、`/file/settings` |
| upload | `/file/upload` |

GET 使用 query；写操作使用 POST JSON，上传使用 multipart。文章保存发送 `requestID`、`version` 和 `input`，动作统一发送 `action`；文档写入发送 `revision`。ID 和版本在 UI 保留字符串，提交使用 `web/shared/native-api.mjs` 精确编码为 JSON 数字。

登录返回 Bearer token，启用 TOTP 时敏感操作请求携带六位验证码。rbac API 签发的 HttpOnly Cookie 为图片与下载提供同一原生凭据；不读取旧 Cookie，不重写旧 API。浏览器权限名称与新 RPC 的权限契约一致。
