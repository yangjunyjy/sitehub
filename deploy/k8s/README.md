Kubernetes 部署按八个应用进程建立 Deployment：rbac、blog、product 四个业务进程，以及三个 Node 前端；网关单独部署。文件服务和调度已并入 blog，不建立独立的 file-api/file-rpc/scheduler Deployment。

Ingress 将浏览器流量转发到统一 Go 网关，网关按 Host 和 `/v1/rbac`、`/v1/blog`、`/v1/product`、`/v1/file` 路由到内部 API；`/v1/file` 同样指向 blog-api。RPC 和数据库只建立集群内部 Service。将 `deploy/config` 中的地址替换为实际 Service DNS，密钥从 Secret 注入。配置 `ControlPlane` 的 etcd 端点及一致的 Prefix/Namespace/Group，RPC 使用 `discovery:///` 目标，通过注册代理发现实例。

保留 rbac、blog、product、file 四个逻辑数据库与账户，分别运行 `tools/bootstrap -service <domain>` Job。blog-api 的 `Database` 和 `FileDatabase` 指向原博客库与文件库，并挂载文件共享卷或连接对象存储。仅一个 blog-api 副本开启 `Scheduler.Enabled`，其他副本关闭该开关。API 的 `/health` 会通过 RPC 检查数据库，rbac 额外检查 Redis。

使用本地存储时，普通文件目录与其加 `-private` 的同级目录均需持久化；默认分别为 `/data/files` 和 `/data/files-private`。所有 blog-api 副本应访问相同的文件数据。迁移旧 file-rpc 前先导出这两个目录，保持原绝对路径及 `STORAGE_SECRET`，验证历史附件可读后再删除旧容器。先使用 `tools/configctl --dir deploy/etcd/config` Job 初始化远端配置，更新已有键需要显式 `--overwrite`；迁移步骤见 `deploy/etcd/README.md`。

本目录提供部署约定；当前可执行的本地编排位于 `deploy/docker-compose.yaml`。

三个前端分别构建 `deploy/Dockerfile.web` 的 `APP=product/blog/admin` 镜像，Node 监听 3000，健康探针使用 `/healthz`。网关将三个域名的页面请求转发至对应前端 Service，`/v1` 请求直接转到业务 API，Node 不代理文件或 API。HTTPS 下设置业务浏览器会话的 Secure Cookie。Node 的健康端点不检查下游，API 依赖健康需单独监控。
