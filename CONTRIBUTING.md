# 参与 SiteHub 开发

从本项目根目录执行后端命令。Go 版本以 `go.mod` 为准；前端需要 Node 22.12+。本地运行需要的数据库、Redis、etcd 与网关配置见 [README](README.md) 和 [etcd 配置指南](deploy/etcd/README.md)。

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

注册代理与配置代理可单独连接临时 etcd 验证，无需数据库、Redis 或 Docker。测试使用随机前缀并清理自己的数据。Windows 示例：

```powershell
$env:ETCD_TEST_ENDPOINT = 'http://127.0.0.1:2379'
go test ./common/controlplane/... -run 'TestEtcdIntegration|TestRegistryProxy' -v
go build -o .tmp/controlplane-smoke-bin/gateway.exe ./gateway
go build -o .tmp/controlplane-smoke-bin/lifecycle-api.exe ./tests/e2e/lifecycle
go build -o .tmp/controlplane-smoke-bin/configctl.exe ./tools/configctl
$env:CONTROLPLANE_SMOKE_BIN = '.tmp/controlplane-smoke-bin'
python scripts/controlplane_smoke.py
```

端到端检查覆盖配置导入/保留/覆盖、服务注册、网关发现、维护开关、错误与删除配置保留旧快照、空实例移除、恢复和进程崩溃后的租约过期。Linux/macOS 构建时去掉上述 `.exe` 后缀即可。

也可将 `SMOKE_ETCD_BINARY` 设为本机 etcd 可执行文件的绝对路径，由脚本创建独立临时实例，额外验证停止 etcd 后的地址过期及重启恢复；这种模式无需填写 `ETCD_TEST_ENDPOINT`。

修改依赖后同步提交 `go.mod` / `go.sum` 或 `web/package.json` / `web/package-lock.json`，避免无关的依赖升级。

## 提交内容

- 保留覆盖业务规则、鉴权、配置和架构边界的自动化测试；测试夹具使用虚构账号和隔离数据。
- 不提交 `.tmp`、运行日志、浏览器截图、测试数据库、编译文件、`node_modules`、`dist` 和生成的编辑器资源。
- 实际配置保留在本地；只提交不含真实凭据的 `.env.example` 和使用环境变量占位符的部署配置。
- PR 说明具体行为变化、验证命令及未验证的环境。功能待办见 [迁移与功能缺口](docs/migration-gaps.md)。
