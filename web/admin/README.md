# SiteHub Admin

管理后台独立构建和运行，源码、协议工具、Sass 样式及 Node 服务均位于本目录。需要 Node 20.19+ 或 22.12+。

```sh
npm install
npm run dev
npm test
npm run build
npm start
```

开发服务默认监听 `127.0.0.1:4311`，`GATEWAY_URL` 指定开发代理目标，`GATEWAY_HOST` 默认 `admin.sitehub.local`。生产服务由 `HOST`、`PORT` 控制监听，仅提供 `dist`、SPA 回退、`/healthz` 和运行时配置；API 请求由 Go 网关转发。`WEB_SITE_URL` 设置访问产品站的地址。

页面统一采用 `src/views/<页面>/index.vue`，页面专用组件放在对应 `components/`，通用组件放在 `src/components/`。路由通过 `import.meta.glob('../views/**/index.vue')` 自动发现；后端 RBAC 菜单是导航与页面授权的来源。菜单 `component` 填写相对 `views` 的精确路径，例如 `system/users/index`，不包含 `views/` 或 `.vue`。菜单 `path` 决定 URL；隐藏菜单仍可提供编辑页路由，按钮不生成路由。

首次刷新会等待菜单和权限加载后重新匹配原始 URL。菜单组件不存在或 URL 未匹配时在管理布局中显示 404，权限不足显示 403。退出后撤销动态路由。工作台、个人中心和编辑按钮按对应组件定位已授权的菜单，修改菜单地址后，站内链接也随之更新。

系统管理提供系统 CRUD，已有系统代码不可修改。菜单管理按系统筛选并维护 `systemCode`，父菜单选项也限制在同一系统中。`code` 是菜单编码，`permission` 是权限标识，两者独立保存；按钮必须填写权限标识。
