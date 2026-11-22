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
curl.exe http://127.0.0.1:9080/ping
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

每个 HTTP 上游使用静态 `nodes` 或服务发现 `service`。服务发现需要启用 `Discovery.Enabled` 并配置 ControlPlane。每条七层路由由一个上游或一个响应型插件处理。

### 如何读一条七层规则

例如 `host: "192.0.2.10:8443"`、`paths: [/api, /api/*]`、`methods: [GET]`、`upstream: application` 表示访问该 IP 和端口的 GET 请求，只要路径是 `/api` 或以 `/api/` 开头，就转给 application。只填域名或 IP 可以匹配该主机的任意端口。省略 Host 或填写 `"*"` 时可按路径直接分流。`"*:8443"` 匹配任意主机的指定端口。IPv6 搭配端口时使用方括号，例如 `"[::1]:8443"`。

多个规则同时命中时，依次比较优先级（数值较大优先）、指定主机、指定端口和匹配路径长度，最后按规则 ID 排序。端口取自 HTTP Host，省略时 HTTP 使用 80，HTTPS 使用 443。前置反向代理若需保留浏览器端口，应传递原始 Host。

### 四层端口转发

在同一份清单的 `streams` 中声明独立的 TCP 或 UDP 监听。以下配置将发往本机 15432 端口的 TCP 连接转到数据库，并将 1053 端口的 UDP 数据报转到 DNS 服务。目标可填容器服务名、域名或 IP。可以只配置四层规则。

```yaml
version: 1
streams:
  - id: database
    protocol: tcp
    listen: "127.0.0.1:15432"
    target: "db:5432"
  - id: resolver
    protocol: udp
    listen: "127.0.0.1:1053"
    target: "dns:53"
```

四层转发按监听地址与协议选择目标，TCP 可直接透传 TLS。每条规则对应一个目标。四层规则可随清单预览、发布、重载和回滚。容器部署需要把规则的监听地址设为容器可达的接口，例如 `0.0.0.0:1053`，并在 Compose 的 gateway 服务 `ports` 中映射 `127.0.0.1:1053:1053/udp`。TCP 端口同理。对外开放时按实际用途选择宿主机绑定地址。

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

## 独立网关控制台

后台菜单“网关控制台”在新窗口打开独立应用。控制台提供路由、上游、插件策略和发布记录。编辑先保存在当前页面的草稿中，“审阅并发布”会校验整份配置并展示各项修改前后的值。回滚会将上一次配置发布为新修订。文件模式提供只读视图，页面发布与回滚使用 `Source: etcd`。

网关进程同时提供 Vue 3 页面、登录接口和配置管理接口。Go 构建时嵌入 `web/gateway-admin/static` 中随源码分发的页面资源。修改 `web/gateway-admin/src` 后，从仓库根目录执行 `npm --prefix web run build -w @sitehub/gateway-admin`。前端构建成功后会自动更新 `static`，提交页面修改时一并提交其中的资源。

本地可以用独立示例体验控制台。先设置上文生成的 `GATEWAY_ADMIN_KEY`，从仓库根目录执行下列命令。第一条账号命令会提示输入专属密码。

```powershell
$env:GATEWAY_CONSOLE_ENABLED = 'true'
go run ./gateway console-user set ops --role admin -f gateway/etc/example/gateway.yaml
go run ./gateway -f gateway/etc/example/gateway.yaml
```

打开 `http://127.0.0.1:9190/`，使用刚创建的网关账号登录。控制台角色 `viewer` 可查看配置，`publisher` 可查看并发布，`admin` 还可回滚。发布和回滚要再次输入当前账号密码。账号保存在 `Console.StateDir/accounts` 中，通过同一条 `set` 命令轮换密码或变更角色，用 `console-user remove ops` 删除账号。重设或删除账号会使该账号的已有会话失效。会话保存在当前网关进程，重启后需重新登录。

升级旧部署后，先创建至少一个网关管理员。后台菜单仍可用 `GATEWAY_CONSOLE_URL` 配置为普通外链，旧 RBAC 角色中的网关授权不再参与鉴权。Compose 使用 `gateway-releases` 卷持久化账号与发布记录；进入运行中的容器创建账号即可生效。

```sh
docker compose up -d gateway
docker compose exec gateway /app/gateway console-user set ops --role admin -f /app/etc/gateway.yaml
```

本地示例将 Cookie 限定在 `/api/`。Compose 默认使用 `/` 以兼容已有反向代理路径。通过后台 Nginx 的 `/gateway-console/` 对外提供控制台时，可在 Compose 的 `.env` 中一起配置公开 Origin、更精确的 Cookie 路径和菜单地址。

```dotenv
GATEWAY_CONSOLE_ORIGIN=https://203.0.113.10:8445
GATEWAY_CONSOLE_COOKIE_PATH=/gateway-console/api/
GATEWAY_CONSOLE_URL=https://203.0.113.10:8445/gateway-console/
```

```nginx
location /gateway-console/ {
    if ($server_port != 8445) { return 404; }
    proxy_pass http://127.0.0.1:9190/;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $http_host;
    client_max_body_size 5m;
}
```

浏览器只访问控制台。网关进程通过内部管理端口 9180 读取配置、发布和回滚，CLI 继续使用 `GATEWAY_ADMIN_KEY`。

## 另一台 Nginx 使用公网域名反代云机

三个站点共用云机 8082 端口。边缘 Nginx 用公网域名选中对应的 `server`，云机网关再用收到的 HTTP Host 选中站点。比如博客入口的 `server_name` 是 `blog.ops-gofer.cn`，其 `proxy_pass` 是 `http://blog.sitehub.local:8082`。在该 `location /` 里写 `proxy_set_header Host $proxy_host;`，网关就会收到 `blog.sitehub.local:8082`，命中博客路由。产品和后台使用相同指令，分别由自己的 `proxy_pass` 选择目标。三个内部名称需在 Nginx 所在机器上解析到网关云机。完整示例见 [异机 Nginx 配置](etc/example/remote-nginx.conf)。

替换现有三个 `location /` 的 Host 指令后，在边缘机运行 `sudo nginx -t && sudo nginx -s reload`。从公网访问三个站点的 `/` 验证首页。若还需启用浏览器登录，核对实际生效清单中的 `browser-session.origins`，在 `profiles.site` 加入 `https://www.ops-gofer.cn:2357` 和 `https://blog.ops-gofer.cn:2357`，在 `profiles.admin` 加入 `https://admin.ops-gofer.cn:2357`。保留已有可信来源，按配置源发布或重载。

## 线上通过同一个公网 IP 访问三个站点

以下按云机宿主机运行 Nginx、应用使用项目 Compose 部署编写。`203.0.113.10` 是示例 IPv4，请统一替换为实际公网 IP。

| 站点 | 浏览器入口 | Nginx 发给 Go 网关的 Host |
| --- | --- | --- |
| 产品 | `https://203.0.113.10:8443` | `product.sitehub.local` |
| 博客 | `https://203.0.113.10:8444` | `blog.sitehub.local` |
| 后台 | `https://203.0.113.10:8445` | `admin.sitehub.local` |

三个前端都使用根路径，且都通过本源的 `/v1` 访问 API。现有部署模板用公开端口区分站点，Nginx 设置固定 Host 后将请求转到 `127.0.0.1:8082`。这些内部名称无需 DNS。也可以把各站点的路由 Host 配成对应的公开 IP 和端口，并将 Nginx 的 `proxy_set_header Host` 配为 `$http_host`。

1. 修改云机 Compose 目录的 `.env`，保留现有密码及其他部署参数。这里的 `GATEWAY_PORT` 对应已有端口映射 `${GATEWAY_PORT:-8082}:9080`。

   ```dotenv
   GATEWAY_PORT=127.0.0.1:8082
   WEB_SITE_URL=https://203.0.113.10:8443
   WEB_PRODUCT_URL=https://203.0.113.10:8443
   WEB_BLOG_URL=https://203.0.113.10:8444
   WEB_COOKIE_SECURE=true
   ```

2. 在云机发布目录的 `gateway/routes.json` 中修改下列两个 `origins` 数组。用当前生产清单作基础，保留原有路由、上游及其他插件配置。Origin 必须与浏览器的协议、IP、端口完全一致，末尾不加斜杠。

   | JSON 字段 | 切换后的值 |
   | --- | --- |
   | `profiles.site.browser-session.origins` | `["https://203.0.113.10:8443", "https://203.0.113.10:8444"]` |
   | `profiles.admin.browser-session.origins` | `["https://203.0.113.10:8445"]` |

3. 将仓库的 [Nginx 配置模板](etc/example/public-ip.nginx.conf) 复制到云机的 `/etc/nginx/conf.d/sitehub-ip.conf`，替换 `server_name` 和证书路径。三个 HTTPS 端口可以共用一张受客户端信任、SAN 包含该 IP 地址的证书。安全组和宿主机防火墙按需开放 8443、8444、8445，后台端口可限定管理来源。8082 绑定回环，9180 为容器内部控制端口。

4. 验证清单，重建读取新环境变量的容器，然后发布。云机配置为 `Source: etcd`；已有 etcd 路由需通过 `gatewayctl publish` 更新。先看 diff，确认与本次迁移相符。

   ```bash
   sudo nginx -t
   docker compose run --rm --no-deps gateway-publish validate --manifest /app/manifests/routes.json
   docker compose up -d gateway web-product web-admin rbac-api
   docker compose run --rm gateway-publish diff --manifest /app/manifests/routes.json -f /app/etc/gatewayctl.yaml
   docker compose run --rm gateway-publish
   sudo systemctl reload nginx
   ```

5. 分别打开三个公网入口，核对首页及后台预览、产品站到博客的跳转。在博客或产品站验证会员登录，再在后台验证管理员登录。只验证首页不能覆盖登录时的 Origin 校验。后台配置的菜单外链、友链或下载外链中，指向旧站点的绝对 URL 也需更新。

此部署把 Nginx 放在 Go 网关之前。当前网关以 TCP 对端地址执行限流，并覆盖转发 IP 头，所以同一路由的访客共享代理地址的限流额度，业务访问日志也会记录代理地址。Nginx 的访问日志可查看客户端 IP。若已有基于真实客户端 IP 的限流或审计要求，需先处理这项差异。

## 扩展与运行边界

新增普通插件实现 `plugin.Factory`，在入口通过 `Registry.Register(name, priority, factory)` 注册；直接生成响应的插件使用 `RegisterTerminal`。终止插件通过注册元数据参与路由校验，新增时无需修改配置契约或路由引擎。入口组装同一个注册表，交给 `manifest.ReadWithRegistry`、命令检查和 `app.RunWithRegistry` 使用，避免“文件检查认识插件，运行引擎不认识”的情况。

插件代码随网关编译，参数、继承关系和启停通过配置重载更新。`Factory` 可能在校验和候选构建期间被多次调用，应只创建轻量、可并发使用的本地中间件；当前接口没有资源关闭协议，不应在构造时启动后台任务、连接外部服务或产生外部副作用。详细约束见 [插件说明](plugin/README.md)。

重载更新 HTTP 路由、上游、插件和四层监听。HTTP 公共入口、管理入口、TLS 证书路径、日志与 ControlPlane 连接等启动参数修改后需要重启。已建立的 HTTP 流式请求和升级连接继续使用旧快照，退出后回收资源。插件配置未变的请求链复用已有状态。限流与熔断状态按网关进程维护。

验证网关与发布工具：

```powershell
go test ./gateway/... ./tools/gatewayctl/...
go vet ./gateway/... ./tools/gatewayctl/...
go build ./gateway ./tools/gatewayctl
```
