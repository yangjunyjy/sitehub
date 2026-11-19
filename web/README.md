# SiteHub 三个前端

产品、管理和博客分别拥有自己的 Vue 页面、SCSS、路由、请求客户端和 Node 页面服务。`web/package.json` 负责统一安装和执行命令。

| 应用 | 开发命令（在 web 目录运行） | 本机端口 | Compose 网关 Host |
| --- | --- | --- | --- |
| 产品 | `npm run dev:product` | 4310 | `product.sitehub.local`（也接受 `sitehub.local`） |
| 管理 | `npm run dev:admin` | 4311 | `admin.sitehub.local` |
| 博客 | `npm run dev:blog` | 4312 | `blog.sitehub.local` |

使用 Node 22.12 及以上版本，在 `web` 运行 `npm ci`。`npm run build` 构建全部应用，`npm test` 运行三个应用的测试。单独构建例如 `npm run build -w @sitehub/blog`，构建后用 `npm run start:blog` 启动博客页面服务。

开发模式下 Vite 将 `/v1` 发给 `GATEWAY_URL`（默认 `http://127.0.0.1:8082`）。部署时让三个域名解析到同一网关，并同步更新网关 Host 与对应允许的 Origin。各 Node 服务仅提供对应应用的页面、资源及 `/healthz`。
