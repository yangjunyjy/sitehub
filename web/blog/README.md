# 云间小记

独立博客前端，使用 Vue 3、Vite 和应用内 Node 静态服务器。公开菜单的路径和层级由 RBAC 管理，组件按 views 目录发现。首页、文章详情和 404 的默认地址及参数在页面同目录的 menu.json 声明。未知地址显示 `views/not-found/index.vue`，保留博客布局。

```sh
npm install
npm run dev
npm run build
npm test
npm start
```

开发与生产默认监听 `127.0.0.1:4312`。开发时 `/v1` 请求转发到 `GATEWAY_URL`（默认 `http://127.0.0.1:8082`），使用 `GATEWAY_HOST`（默认 `blog.sitehub.local`）作为 Host。生产 Node 只提供静态页面，API 必须通过网关访问。

首页、分类、标签、归档、搜索和文章详情使用真实博客接口。会员注册、登录和退出使用 `/v1/rbac/members/*`。关于、友链和留言使用 `/v1/blog/about`、`/v1/blog/friends`、`/v1/blog/messages`；提交成功以接口返回 `pending` 为准，不会把请求失败显示为成功。评论、留言、友链申请、点赞和收藏均需会员登录。个人中心按收藏夹展示已发布文章，首次收藏默认放入「默认收藏夹」。文章互动和收藏夹接口在 blog.api 的 reader 分组中定义。

导航从 `/v1/rbac/menus/public?system=blog` 获取。内部页面按本应用组件路径解析，外链页面按菜单中的 HTTP 或 HTTPS 地址打开。会员会话使用 RBAC 签发的 HttpOnly Cookie。

后台「基础设置 → 博客首页」保存动画开关和短句。博客从 `/v1/blog/about` 读取设置，关闭动画或系统启用减少动态效果时显示完整短句。友链页使用已审核的友链，卡片和列表共享数据并记住访客的展示偏好。

博客数据库结构版本为 6。升级已有数据库时，使用现有数据库连接运行 `go run ./tools/bootstrap -service blog`，SQLite 环境追加 `-driver sqlite`。在部署目录也可使用现有的 `bash init.sh database` 流程。完成迁移后再启动新版博客服务。
