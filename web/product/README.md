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

应用启动时读取一次公开菜单接口 /v1/rbac/menus/public?system=product。页头、移动导航和页脚使用这份菜单树，内部菜单按实际 views 组件注册到配置的 URL。页面同目录的 menu.json 定义默认地址和参数，文档阅读器与搜索结果会沿用后台设置的路径。页面专属组件放在 components 子目录。未知地址显示布局内 404。
