# SiteHub 官网与管理后台

`site`、`admin` 是两个独立 Vue 应用，使用 npm workspaces。各自构建后由 `server/start.mjs` 启动 Node 静态页面服务。Node 不代理 API，也不签发或转换 Cookie。

## 启动与开发

需要 Node 22.12+。从 `web` 执行：

```powershell
npm ci
npm run build
npm run start:site
# 另一个终端
npm run start:admin
```

Node 默认端口为官网 4310、后台 4311，直接访问可检查页面与 `/healthz`。完整业务入口必须经过网关；Compose 中两个 Node 均不发布宿主端口。官网 `http://sitehub.local:8082`，后台 `http://admin.sitehub.local:8082`。

开发使用 `npm run dev:site` / `npm run dev:admin`，保留 Vite 热更新。Vite 只把 `/v1` 转发给 `GATEWAY_URL`，分别携带对应的 `GATEWAY_HOST`。网关必须允许开发 Origin，不能同时运行同端口的 Node 和 Vite。

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| HOST | 127.0.0.1 | Node 监听地址，容器为 0.0.0.0 |
| PORT | 官网 4310、后台 4311 | Node 监听端口 |
| GATEWAY_URL | http://127.0.0.1:8082 | 仅用于 Vite 开发代理 |
| GATEWAY_HOST | sitehub.local / admin.sitehub.local | 仅用于 Vite，选择身份域 |
| VITE_WEBSITE_URL | 开发环境的官网地址 | 管理后台打开官网按钮，生产构建前显式指定 |

`web/.env.example` 提供变量示例。Node 使用环境变量，可用 `node --env-file=.env server/start.mjs --app site`。Vite 读取应用目录的 `.env` 文件。

## 会话和数据契约

浏览器请求同源 `/v1/rbac`、`/v1/blog`、`/v1/product`、`/v1/file`。rbac API 签发 HttpOnly、SameSite=Lax Cookie；自研 Go 网关 按 Host 和身份域提取 Cookie，检查 Cookie 写请求的 Origin。显式 Bearer 优先。生产配置 `WEB_COOKIE_SECURE=true` 并启用 HTTPS。后台与会员仍由 RPC 实时验证权限、撤销及 TOTP。

图片和普通下载链接通过 Cookie 认证；自研 Go 网关 保留 Range/HEAD、Content-Disposition 和流式传输。上传、下载不经过 Node。后端地址、RPC 和内部服务密钥不下发到浏览器。

API 成功响应仅接受 `code: 0`。uint64 ID、版本号与 revision 在 UI 内保留字符串，提交时精确序列化为 JSON 数字。接口完全采用新项目契约。

## 验证

```powershell
npm test
npm run build
```

浏览器业务验收应连接真实网关，并配合隔离后端测试账号；不能使用 Node 直连地址代替统一入口。架构、部署和集成测试说明见 [治理运行指南](../docs/gateway-governance-operations.md)。业务缺口仍见 [迁移清单](../docs/migration-gaps.md)。
