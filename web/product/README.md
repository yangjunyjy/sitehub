# 产品前端

此应用独立提供产品首页、下载中心、产品文档及会员登录、注册和账号页。代码、静态资源、样式、请求工具、依赖和生产页面服务器均在本目录内，不引用其他前端应用。

```bash
npm install
npm run dev
npm run build
npm start
npm test
```

开发和默认生产监听端口为 `4310`。生产服务器只提供静态页面；`/v1` 请求由网关转发，开发时通过本应用的 Vite 代理访问 `GATEWAY_URL`，默认请求 Host 为 `product.sitehub.local`，可用 `GATEWAY_HOST` 修改。

路由直接来自 `src/views/**/index.vue`，无页面映射表：`home/index.vue` 对应 `/`，`download/index.vue` 对应 `/download`，文档和 404 页在各自目录的 `route.js` 声明可选参数或兜底路径。页面专属组件放在页面目录的 `components/`，该目录不参与路由生成。未匹配路径显示布局内 404，保留导航和页脚；导航菜单接口失败不会阻止文件路由建立。

导航读取 `/v1/rbac/menus/public?system=product`；页面仍由本地文件约定决定，不根据菜单维护第二套路由。样式使用 SCSS 和 `sass-embedded`。
