# Go 网关：文件代理配置与插件

`config.Config` 只描述进程启动参数：公开监听、管理监听、日志、TLS、配置源，以及可选的 ControlPlane 服务发现连接和 etcd 路由存储连接。业务服务、上游地址、路由和插件参数都写在配置文件中，新增代理无需增加 Go 结构体字段。

配置分为两层：

| 文件 | 职责 |
| --- | --- |
| `etc/gateway.yaml` | 网关进程参数；默认 `Source: file`、`Manifest: routes.yaml` |
| `etc/routes.yaml` | 项目代理配置入口，声明全局插件并通过 `include` 加载其他文件 |
| `etc/conf.d/00-upstreams.yaml` | 上游节点或发现服务名、健康检查及超时 |
| `etc/conf.d/10-profiles.yaml` | 可复用的插件配置模板 |
| `etc/conf.d/20-site.yaml`、`30-admin.yaml` | 官网和管理后台路由 |
| `etc/example/` | 使用静态上游、无需控制平面 的独立示例 |

`Manifest` 相对启动配置文件解析；每个 `include` 相对声明它的文件解析。HTTP 代理使用标准库 `httputil.ReverseProxy`，支持 Range/HEAD、流式上传下载、SSE 和 WebSocket；上游支持主动健康检查和加权轮询。

## 从独立示例开始

从项目根目录执行。先在 PowerShell 生成至少 32 字符的管理密钥，不要输出或提交实际密钥：

```powershell
$gatewayKeyBytes = New-Object byte[] 32
$gatewayRng = [Security.Cryptography.RandomNumberGenerator]::Create()
$gatewayRng.GetBytes($gatewayKeyBytes)
$gatewayRng.Dispose()
$env:GATEWAY_ADMIN_KEY = -join ($gatewayKeyBytes | ForEach-Object { $_.ToString('x2') })

# 只检查配置，不连接 etcd 或上游服务。
go run ./gateway check -f gateway/etc/example/gateway.yaml

# 前台启动，公开端口 9080，管理端口 127.0.0.1:9180。
go run ./gateway -f gateway/etc/example/gateway.yaml
```

另一个 PowerShell 终端可检查示例的静态响应：

```powershell
curl.exe -H 'Host: example.test' http://127.0.0.1:9080/ping
```

正常返回 `pong`，这条路由不需要上游。示例的 `/api` 和 `/public` 路由连接 `http://127.0.0.1:8802`，可在 `etc/example/conf.d/application.yaml` 修改节点地址和健康检查路径。上游不健康时，代理请求返回 503。

修改代理或插件配置后，在另一个终端设置与启动时**相同**的 `GATEWAY_ADMIN_KEY`，再检查并重载：

```powershell
# 将占位值替换为启动网关时使用的同一个密钥，不要重新生成。
$env:GATEWAY_ADMIN_KEY = 'replace-with-the-same-key-used-when-starting-gateway'
go run ./gateway check -f gateway/etc/example/gateway.yaml
go run ./gateway reload -f gateway/etc/example/gateway.yaml
```

`reload` 调用私有管理接口 `POST /v1/reload`。运行中的网关重新读取自己的配置文件，整份校验成功后原子切换；读取或构建失败时保留原配置。文件模式不会自动监听文件变化，也不接受 `PUT /v1/config`，更改文件后需显式重载。重载入口保留管理网络与密钥校验。

默认项目配置 `etc/gateway.yaml` 同样使用文件模式，但 `etc/conf.d` 中的业务 API 使用 `service` 发现，因此仍需启用 ControlPlane 服务发现。要独立代理普通 HTTP 服务，使用 `nodes` 并设置 `Discovery.Enabled: false` 即可。

## 维护代理与插件配置

配置使用 JSON 或 YAML。以下是完整的单文件例子，也可将上游和路由分别放进 `conf.d`：

```yaml
version: 1
plugins:
  request-id: {}
  rate-limit: {rate: 100, burst: 100}
  circuit-breaker: {}

upstreams:
  application:
    nodes:
      - url: http://127.0.0.1:8802
    health_path: /health

profiles:
  api:
    proxy-rewrite:
      strip_prefix: /api
      host: application.internal

routes:
  - id: application
    host: example.test
    paths: [/api, /api/*]
    upstream: application
    profile: api
    plugins:
      rate-limit: null
```

插件按全局 `plugins` → `profiles.<名称>` → 路由 `plugins` 逐级覆盖。同名插件替换整个配置对象，不合并内部字段；`null` 关闭继承插件。上例路由保留请求 ID、熔断与路径改写，关闭全局限流。插件执行顺序由注册优先级确定，与 YAML/JSON 字段顺序无关。

常用内置插件包括 `rate-limit`、`limit-count`、`circuit-breaker`、`proxy-rewrite`、`deny-paths`、`request-id` 和 `static-response`。例如全局配置 `deny-paths` 可限制管理路径，路由可按需要覆盖或关闭；内核不固定占用这些公共路径。`browser-session` 可通过 `cookie_name`、`origins` 和 `origin_check_paths` 配置其他系统的 Cookie 与 Origin 策略。详细参数、限制与示例见 [插件说明](plugin/README.md)。

每个上游必须选择 `nodes` 或 `service` 之一。`service` 是配置中的通用服务名，不要求 `-api` 后缀；运行时需要 `Discovery.Enabled: true` 和有效的 ControlPlane 参数。路由按优先级、路径匹配选择，并要求显式的 DNS Host；路径支持精确匹配和结尾 `/*` 的前缀匹配。每条路由只能有一个上游或一个响应型插件。

文件加载会补齐省略的上游参数：节点权重 1，连接超时 3000ms，响应头超时 10000ms，I/O 空闲超时 30000ms，健康路径 `/healthz`、检查间隔 2000ms、检查超时 1000ms、失败阈值 2。显式填写的无效值仍会被拒绝；需要长时间下载等行为时按上游覆盖。

`include: [conf.d/*.yaml]` 支持按字典序展开通配符和递归包含。根文件必须声明 `version: 1`，被包含文件可省略版本。重复的上游、模板、路由 ID 或全局插件定义，重复包含、循环包含、未匹配的通配符、未知字段和未知插件都会报错；拼错的插件名即使值为 `null` 也会报错。总计最多 128 个文件、4 MiB 输入，解析结果也受 4 MiB 限制。YAML 不接受重复键、别名或合并键，复用配置通过模板和包含文件表达。

## 使用 etcd 发布到多个网关

`Source: etcd` 适用于共享配置发布，需要填写 `Etcd.Endpoints` 和 `Etcd.Key`。配置源与服务发现相互独立：路由存储保存代理配置，注册代理通过 ControlPlane 提供 `service` 上游发现，两者可使用同一 etcd 的独立键空间。已有配置未填写 `Source` 时，存在 `Etcd.Endpoints` 则沿用 etcd，否则使用文件模式。

etcd 模式仅在键为空时用 `Manifest` 初始化。键已有内容时，以 etcd 中的配置为准，修改磁盘文件或重启进程不会覆盖在线版本。候选文件同样支持 YAML 和 `include`，发布时先合并为一份完整文档，再通过 CAS 写入 etcd，各网关校验并原子应用。

设置与网关一致的 `GATEWAY_ADMIN_KEY`，并在 `etc/gatewayctl.yaml` 填写实际 `AdminURL` 后，从项目根目录执行：

```powershell
go run ./tools/gatewayctl validate --manifest gateway/etc/routes.yaml
go run ./tools/gatewayctl diff --manifest gateway/etc/routes.yaml -f gateway/etc/gatewayctl.yaml
go run ./tools/gatewayctl publish --manifest gateway/etc/routes.yaml -f gateway/etc/gatewayctl.yaml
```

`gatewayctl` 未显式传入 `--manifest` 时仍读取原有 `gateway/etc/routes.json`，兼容已有发布流程。发布记录写入配置的 `StateDir`，已确认的记录可用 `rollback --release <记录文件>` 恢复其前一版本；并发版本变化时会拒绝覆盖。文件模式使用 `gateway reload`，etcd 模式使用 `gatewayctl publish`。

## 扩展与运行边界

新增普通插件实现 `plugin.Factory`，在入口通过 `Registry.Register(name, priority, factory)` 注册；直接生成响应的插件使用 `RegisterTerminal`。终止插件通过注册元数据参与路由校验，新增时无需修改配置契约或路由引擎。入口组装同一个注册表，交给 `manifest.ReadWithRegistry`、命令检查和 `app.RunWithRegistry` 使用，避免“文件检查认识插件，运行引擎不认识”的情况。

插件代码随网关编译，参数、继承关系和启停通过配置重载更新。`Factory` 可能在校验和候选构建期间被多次调用，应只创建轻量、可并发使用的本地中间件；当前接口没有资源关闭协议，不应在构造时启动后台任务、连接外部服务或产生外部副作用。详细约束见 [插件说明](plugin/README.md)。

重载只替换代理、上游与插件配置。监听地址、TLS 证书路径、日志以及 ControlPlane/路由存储连接等进程参数修改后需要重启。已建立的流式请求和升级连接继续使用旧快照，相关资源在请求退出后回收；插件配置未变的请求链复用已有状态。限流与熔断状态仅在单个网关进程内维护，不代表集群总配额。

验证网关与发布工具：

```powershell
go test ./gateway/... ./tools/gatewayctl/...
go vet ./gateway/... ./tools/gatewayctl/...
go build ./gateway ./tools/gatewayctl
```
