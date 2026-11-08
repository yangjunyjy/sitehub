# Go 网关插件

插件注册在 `DefaultRegistry()`，运行顺序由注册优先级决定，与 JSON 字段顺序无关。未知插件、未知参数和无效参数会使整份候选配置被拒绝。

| 插件 | 配置示例 | 作用 |
| --- | --- | --- |
| `request-id` | `{}` | 校验或生成请求 ID，写入上游请求和下游响应 |
| `browser-session` | `{"audience":"admin","origins":["https://admin.example"]}` | admin/site Cookie 隔离、显式 Authorization 优先、Cookie 写请求检查 Origin |
| `strip-credentials` | `{}` | 发给 Node 前端的请求移除会话和内部身份凭据 |
| `rate-limit` | `{"rate":100,"burst":100}` | 每条路由、每个直接连接 IP 的令牌桶限流 |
| `limit-count` | `{"count":10,"window_seconds":60}` | 每条路由、每个直接连接 IP 的固定窗口计数 |
| `circuit-breaker` | `{"failures":5,"successes":2,"initial_open_seconds":1,"max_open_seconds":30}` | 连续失败触发熔断、指数退避、单并发半开探测恢复 |
| `static-response` | `{"status":404,"body":"missing","headers":{"Content-Type":"text/plain"}}` | 无上游的显式静态响应 |

执行顺序与上表一致。`circuit-breaker` 的四个参数可以全部省略，默认值如表。`rate-limit` 可配置 `idle_ttl_seconds`（默认 300）和 `max_keys`（默认 10000）；`limit-count` 可配置 `max_keys`（默认 10000）。限流状态仅在当前进程内共享，不是集群总配额。达到 key 容量后拒绝新 key，旧 key 到期后释放，避免活跃计数器被逐出后反复重置。

限流使用 `RemoteAddr`，不信任客户端 `X-Forwarded-For`。限流/熔断对象由 `Build()` 构造，网关复用配置未变的 Middleware，热加载不会重置这些对象。熔断观察状态码，不缓存响应体；支持 `Flush`、`ResponseController` 和连接升级，不阻塞 SSE、下载或 WebSocket 转发。

新增插件实现 `Factory`，用 `Registry.Register(name, priority, factory)` 注册，并添加参数校验和行为测试。编译后新增插件代码才生效；已注册插件的配置可以热更新。不加载任意本机 `.so`、Lua 或网络下载代码。

验证：

```sh
go test -race ./gateway/plugin
```
