# 自研 Go 网关

`gateway/` 是独立 Go 入口，使用 Cobra/Viper 配置、go-zero 日志、标准库 `httputil.ReverseProxy`。业务 API/RPC 仍使用 go-zero。网关只负责入口治理，不连接业务数据库。

```text
gateway/
├── main.go                 # Cobra 命令、启动与本地 readiness 检查
├── etc/                    # 本地 bootstrap、原生路由清单、gatewayctl 配置
├── manifest/               # 原生配置契约、整份校验
├── plugin/                 # 注册表、固定执行顺序、内置 Go 插件与测试
└── internal/
    ├── config/             # Viper bootstrap 强类型配置
    ├── app/                # 进程依赖、双监听、优雅停机与连接回收
    ├── control/            # 管理接口、etcd CAS/watch、原子发布
    ├── discovery/          # Nacos 订阅、空列表、过期与恢复
    └── engine/             # 不可变快照、代理、健康检查、连接池、指标
```

能力：Host/路径/方法/优先级路由；API 与 Node 统一入口；浏览器 Cookie/Origin 策略；主动健康检查和加权轮询；限流、登录频控、熔断；Range/HEAD、流式上传下载、SSE、WebSocket；整份配置热发布与版本回滚。

配置由 `tools/gatewayctl` 发布到网关私有管理端口，单次 etcd CAS 写入整份文档，各实例 watch 后校验并切换。没有 APISIX、Lua 或原项目兼容层。Go 插件代码编译进镜像，插件参数可热更新。

- [插件配置与扩展](plugin/README.md)
- [架构设计](../docs/gateway-governance-design.md)
- [启动、发布、回滚及验收](../docs/gateway-governance-operations.md)
- [开源实现参考与能力对照](../docs/gateway-open-source-reference.md)
