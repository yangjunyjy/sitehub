Kubernetes 部署按十一个应用进程建立 Deployment：四个 API、四个 RPC、一个 scheduler、两个 Node 前端。

API 的 Service 由 Ingress 按 `/v1/rbac`、`/v1/blog`、`/v1/product`、`/v1/file` 转发；RPC 和数据库只建立集群内部 Service。将 `deploy/config` 中的地址替换为实际 Service DNS，密钥从 Secret 注入。RPC 配置沿用 go-zero 的 `Endpoints` 或 `Etcd` 服务发现配置。

每个 RPC 使用自己的数据库与账户，初始化单独运行 `tools/bootstrap -service <domain>` Job。file 使用共享持久卷或对象存储；scheduler 初始部署一个副本。API 的 `/health` 会通过 RPC 检查数据库，rbac 额外检查 Redis。

本目录提供部署约定；当前可执行的本地编排位于 `deploy/docker-compose.yaml`。

官网和后台分别构建 `deploy/Dockerfile.web` 的 `APP=site/admin` 镜像，Node 监听 3000，健康探针使用 `/healthz`。两个域名的 Ingress 分别转发到各自前端 Service，包含 `/v1` 在内的浏览器流量经 Node 转发至 API。通过环境变量设置四个 `*_API_URL` 为集群 Service 地址，HTTPS 设置 `COOKIE_SECURE=true`。Node 的健康端点不检查下游，API 依赖健康需单独监控。
