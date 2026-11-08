# 参与 SiteHub 开发

从本项目根目录执行后端命令。Go 版本以 `go.mod` 为准；前端需要 Node 22.12+。本地运行需要的数据库、Redis、Nacos 与网关配置见 [README](README.md) 和 [治理运行指南](docs/gateway-governance-operations.md)。

## 修改与验证

业务实现按 `service/<domain>` 划分；跨域调用使用 RPC。修改 HTTP/RPC 契约后，运行 `scripts/generate.ps1`，审查生成差异，并提交契约及生成代码。

```powershell
go test ./...
go vet ./...
go build ./...
npm ci --prefix web
npm test --prefix web
npm run build --prefix web
```

涉及 SQLite 的测试需要 CGO 和 C 编译器。业务进程联调使用 `python scripts/smoke.py`；涉及网关、注册发现或配置热加载时，按治理运行指南增加 `--governance`；前端业务联调增加 `--frontends`。Linux 停机注销与排空验证使用 `python scripts/lifecycle_smoke.py`。这些测试会自行创建隔离资源。

修改依赖后同步提交 `go.mod` / `go.sum` 或 `web/package.json` / `web/package-lock.json`，避免无关的依赖升级。

## 提交内容

- 保留覆盖业务规则、鉴权、配置和架构边界的自动化测试；测试夹具使用虚构账号和隔离数据。
- 不提交 `.tmp`、运行日志、浏览器截图、测试数据库、编译文件、`node_modules`、`dist` 和生成的编辑器资源。
- 实际配置保留在本地；只提交不含真实凭据的 `.env.example` 和使用环境变量占位符的部署配置。
- PR 说明具体行为变化、验证命令及未验证的环境。功能待办见 [迁移与功能缺口](docs/migration-gaps.md)。
