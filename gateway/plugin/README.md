# Go 网关插件

限流、熔断等内置能力通过 `plugins` 配置启用，无需修改 `config.Config` 或路由引擎。插件注册在 `DefaultRegistry()`，运行顺序由注册优先级决定，与 YAML/JSON 字段顺序无关。未知插件、未知参数和无效参数会使整份候选配置被拒绝。

配置按全局 `plugins` → `profiles.<名称>` → 路由 `plugins` 逐级覆盖。同名插件整体替换配置，不逐字段合并；写 `null` 可以关闭继承的插件。拼错的插件名即使值为 `null` 也会报错。最终生效的配置由注册表校验后构建。

```yaml
version: 1
plugins:
  request-id: {}
  circuit-breaker: {}
profiles:
  public-api:
    rate-limit: {rate: 100, burst: 100}
routes:
  - id: public-downloads
    host: downloads.example.com
    paths: ["/api/*"]
    upstream: downloads
    profile: public-api
    plugins:
      rate-limit: null
      proxy-rewrite:
        strip_prefix: /api
        host: files.internal
        headers: {X-Upstream-Region: east}
# upstreams.downloads 在此文件或 include 引入的配置中声明。
```

| 插件 | 配置示例 | 作用 |
| --- | --- | --- |
| `request-id` | `{}` | 校验或生成请求 ID，写入上游请求和下游响应 |
| `deny-paths` | `{"paths":["/healthz","/internal","/internal/*"],"status":404}` | 按配置拒绝精确路径或路径段前缀，默认返回 404 |
| `browser-session` | `{"cookie_name":"project_session","origins":["https://portal.example"],"origin_check_paths":["/session/login"]}` | 将指定 Cookie 转为 Bearer，显式 Authorization 优先，检查 Cookie 写请求与指定入口的 Origin |
| `strip-credentials` | `{}` | 发给 Node 前端的请求移除会话和内部身份凭据 |
| `rate-limit` | `{"rate":100,"burst":100}` | 每条路由、每个直接连接 IP 的令牌桶限流 |
| `limit-count` | `{"count":10,"window_seconds":60}` | 每条路由、每个直接连接 IP 的固定窗口计数 |
| `proxy-rewrite` | `{"strip_prefix":"/api","host":"backend.internal","headers":{"X-Region":"east"}}` | 路由匹配后，改写发送给上游的路径前缀、Host 与普通请求头 |
| `circuit-breaker` | `{"failures":5,"successes":2,"initial_open_seconds":1,"max_open_seconds":30}` | 连续失败触发熔断、指数退避、单并发半开探测恢复 |
| `static-response` | `{"status":404,"body":"missing","headers":{"Content-Type":"text/plain"}}` | 无上游的显式静态响应 |

执行顺序与上表一致。`circuit-breaker` 的四个参数可以全部省略，默认值如表。`rate-limit` 可配置 `idle_ttl_seconds`（默认 300）和 `max_keys`（默认 10000）；`limit-count` 可配置 `max_keys`（默认 10000）。限流状态仅在当前进程内共享，不是集群总配额。达到 key 容量后拒绝新 key，旧 key 到期后释放，避免活跃计数器被逐出后反复重置。

`deny-paths` 在请求 ID 插件之后、会话插件之前执行。`paths` 必须包含 1..64 个不同的规范路径；`/internal/*` 匹配 `/internal/` 及子路径，不匹配 `/internal` 或 `/internalized`，需要同时阻止 `/internal` 时显式列出它。`status` 可设为 400..599，省略时为 404；HEAD 请求不返回正文。项目示例把管理路径限制声明为全局插件配置，网关内核不固定占用这些公共路径。其他站点可调整列表或在路由层写 `deny-paths: null` 取消继承。

`browser-session` 可以只设置 `cookie_name`，不设置 `audience`，便于代理其他系统。兼容原配置：只设置 `audience: admin` 或 `audience: site` 时，分别使用 `sitehub_admin_session`、`sitehub_site_session`；同时设置两者时，以 `cookie_name` 为准，`audience` 若填写仍只接受 `admin` 或 `site`。Cookie 名称按 HTTP Cookie 规则校验，`origins` 必须显式列出允许的 HTTP(S) Origin，不包含路径。

`origin_check_paths` 是额外检查浏览器登录等入口的精确路径列表：非 GET/HEAD/OPTIONS 请求命中列表且携带 Origin 时，Origin 必须在允许列表中；不携带 Origin 的 CLI 请求仍可通过。为保留旧配置的保护，省略该字段（包括使用自定义 `cookie_name` 时）沿用 `/v1/rbac/auth/login`、`/v1/rbac/auth/register`、`/v1/rbac/members/login`、`/v1/rbac/members/register` 四个默认路径。其他系统应显式提供自己的路径列表，或设置 `origin_check_paths: []` 关闭这项额外的路径检查。无论该列表为何值，只要请求通过 Cookie 提供身份且没有显式 Authorization，所有写请求仍须提供允许的 Origin。

`proxy-rewrite` 的三个字段均可省略。`strip_prefix: /api` 将 `/api/files` 改为 `/files`，将 `/api` 改为 `/`，但不会匹配 `/apiv2`。改写保留查询字符串和路径的原始转义；`host` 只改变上游 HTTP Host，不改变上游连接目标。请求头不允许覆盖连接管理、消息长度、会话身份、转发来源和请求 ID 等受保护字段；Host 使用单独的 `host` 参数设置。该插件在会话检查和限流之后执行，不改变原始路由选择。

限流使用 `RemoteAddr`，不信任客户端 `X-Forwarded-For`。限流/熔断对象由 `Build()` 构造，网关复用配置未变的 Middleware，热加载不会重置这些对象。熔断观察状态码，不缓存响应体；支持 `Flush`、`ResponseController` 和连接升级，不阻塞 SSE、下载或 WebSocket 转发。

新增插件实现 `Factory`，用 `Registry.Register(name, priority, factory)` 注册，并添加参数校验和行为测试。直接生成响应的插件使用 `Registry.RegisterTerminal(name, priority, factory)`；路由根据注册元数据检查“一个上游或一个终止插件”，不依赖 `static-response` 等具体名称。注册后的终止插件不能继续转发到下一处理器。同一路由不能启用两个终止插件。

`Factory` 可能在配置检查、继承层校验和候选路由构建时被多次调用，必须只构造轻量、可并发使用的本地中间件。构造阶段不应启动后台 goroutine、建立外部连接或产生外部副作用；当前接口没有 `Close` 或其他资源回收协议，不能用它承诺支持持有这类资源的插件。需要网络连接或后台任务的扩展，应先设计明确的构造、提交失败清理和快照退役回收接口，再接入请求链。

注册表在应用入口组装后传入配置读取和请求引擎，扩展插件无需为业务服务增加配置结构字段。编译后新增插件代码才生效；已注册插件的参数、继承和启停可以通过配置重载更新。不加载任意本机 `.so`、Lua 或网络下载代码。

验证：

```sh
go test ./gateway/plugin
# 在具备 CGO 和 C 编译器的环境中检查并发行为：
go test -race ./gateway/plugin
```
