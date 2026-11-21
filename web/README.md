# SiteHub 三个前端

产品、管理和博客分别拥有自己的 Vue 页面、SCSS、路由、请求客户端和 Node 页面服务。`web/package.json` 负责统一安装和执行命令。

| 应用 | 开发命令（在 web 目录运行） | 本机端口 | Compose 网关 Host |
| --- | --- | --- | --- |
| 产品 | `npm run dev:product` | 4310 | `product.sitehub.local`（也接受 `sitehub.local`） |
| 管理 | `npm run dev:admin` | 4311 | `admin.sitehub.local` |
| 博客 | `npm run dev:blog` | 4312 | `blog.sitehub.local` |

使用 Node 22.12 及以上版本，在 `web` 运行 `npm ci`。`npm run build` 构建全部应用，`npm test` 运行三个应用的测试。单独构建例如 `npm run build -w @sitehub/blog`，构建后用 `npm run start:blog` 启动博客页面服务。

开发模式下 Vite 将 `/v1` 发给 `GATEWAY_URL`（默认 `http://127.0.0.1:8082`）。部署时让三个域名解析到同一网关，并同步更新网关 Host 与对应允许的 Origin。各 Node 服务仅提供对应应用的页面、资源及 `/healthz`。

## 页面与图标

路由按各应用的 `src/views/<业务>/<页面>/index.vue` 发现。复杂页面的表格列、领域编辑器和 SCSS 与页面同目录。后台的通用表单弹窗、列表和审核队列分别在 `src/components/admin`、`src/components/community` 中，图标选择器与创建表单在 `src/components/icons`。

RBAC 初始化会扫描三个应用的 views 目录。带 index.vue 的目录是页面，没有页面的上级目录可生成菜单分组，components 子目录跳过。页面或分组可以在同目录的 menu.json 设置名称、排序和默认路径。menu 设为 false 可将登录、详情等页面排除在初始菜单之外。新增页面随构建发布，运行 RBAC bootstrap 后补入数据库。菜单管理中的改名、移动、图标和 URL 调整在重新加载应用时生效。

三个应用在构建时共用 `web/shared/icons` 的 SVG 目录与渲染组件。菜单、分类、官网能力、文档模式和下载包保存图标标识。内置资源以目录和文件名引用，自定义资源使用 `custom:<name>`。RBAC 的 `/v1/rbac/system/icons` 提供列表、内容、校验、创建、更新和停用。公开页面通过内容接口读取经过校验的 SVG。停用后，新选择列表隐藏该资源，历史引用仍能显示。

升级时使用项目现有数据库连接分别运行 `go run ./tools/bootstrap -service rbac` 和 `go run ./tools/bootstrap -service blog`，使用 SQLite 时追加 `-driver sqlite`。这会更新图标表、分类图标长度和管理菜单。原友链页面的角色授权会随新菜单迁移。图标管理使用 `system:icon:list/create/update/delete` 四项独立权限。
