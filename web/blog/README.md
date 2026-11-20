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

导航从 `/v1/rbac/menus/public?system=blog` 获取。内部页面按本应用组件路径解析，外链页面按菜单中的 HTTP 或 HTTPS 地址打开。会员会话使用 RBAC 签发的 HttpOnly Cookie。

后台「基础设置 → 博客首页」保存动画开关和短句。博客从 `/v1/blog/about` 读取设置，关闭动画或系统启用减少动态效果时显示完整短句。友链页使用已审核的友链，卡片和列表共享数据并记住访客的展示偏好。

本次升级将 RBAC 和博客数据库的结构版本从 2 提升到 3。云机使用新版发布包构建镜像后，在发布目录运行 `bash init.sh database`，随后启动新版服务。旧菜单默认使用内部页面类型。
