# 云间小记

独立博客前端，使用 Vue 3、Vite 和应用内 Node 静态服务器。路由根据 `src/views/**/index.vue` 的普通目录生成。首页、文章详情和 404 在各自目录的 `route.mjs` 声明特殊 URL。未知地址显示 `views/not-found/index.vue`，保留博客布局。

```sh
npm install
npm run dev
npm run build
npm test
npm start
```

开发与生产默认监听 `127.0.0.1:4312`。开发时 `/v1` 请求转发到 `GATEWAY_URL`（默认 `http://127.0.0.1:8082`），使用 `GATEWAY_HOST`（默认 `blog.sitehub.local`）作为 Host。生产 Node 只提供静态页面，API 必须通过网关访问。

首页、分类、标签、归档、搜索和文章详情使用真实博客接口。会员登录使用 `/v1/rbac/members/*`。关于、友链和留言使用 `/v1/blog/about`、`/v1/blog/friends`、`/v1/blog/messages`；提交成功以接口返回 `pending` 为准，不会把请求失败显示为成功。空内容保留空态，不注入演示文章。

导航从 `/v1/rbac/menus/public?system=blog` 获取，仅展示本应用已有页面的组件路径；路由生成不依赖菜单请求。会员会话使用 RBAC 签发的 HttpOnly Cookie，浏览器不保存访问令牌。

插画位于 `src/assets/skies`，Vite 会将 WebP 文件打进本应用的静态资源。首页导航和互动入口使用 Inspira UI 的 Liquid Glass，天空中的流星使用 Meteors。组件代码位于 `src/components/inspira`，许可证位于 `public/licenses`。留言页使用本地 SVG 地球和 SCSS 星空，所有页面样式、接口与启动代码均由本应用维护。
