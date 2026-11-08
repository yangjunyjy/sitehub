# SiteHub — 基于 Go 的站点与产品运营平台

SiteHub 提供站点内容、产品文档、下载中心、身份权限、文件和定时任务管理。官网演示内容可通过后台配置，Deploy 是当前演示产品名称。

独立 Go module（`sitehub`），使用 go-zero 1.10.3，按 `rbac`、`blog`、`product`、`file` 和 `scheduler` 划分业务。四个业务域各有原生 API 与 RPC 进程，scheduler 独立运行。功能在新服务内实现，使用新的接口、认证和数据归属；原项目不参与运行。

**迁移尚未全部完成。** 主体架构与主要业务链路已接通，待补功能、实现问题和验收范围见 [后端迁移缺口与接手清单](docs/migration-gaps.md)。

## 项目标识

| 项目项 | 当前名称 |
| --- | --- |
| 项目目录 / Go module | `sitehub` |
| 前端 workspace | `@sitehub/admin`、`@sitehub/site` |
| 开发域名 | `sitehub.local`、`admin.sitehub.local` |
| 会话 Cookie | `sitehub_admin_session`、`sitehub_site_session` |
| Nacos 默认 namespace / group | `sitehub-dev` / `SITEHUB` |
| 网关 etcd 配置键 | `/sitehub/gateway/v1/config` |
| Compose 项目名 | `sitehub` |
| 默认业务数据库 | `sitehub_rbac`、`sitehub_blog`、`sitehub_product`、`sitehub_file` |

已有部署的数据库、配置中心内容和持久卷不会随工程更名自动迁移；使用新命名部署前，需要显式迁移对应数据与配置。各业务服务仍按其职责命名。

## 服务与目录

| 服务 | HTTP / RPC 默认端口 | 职责 |
| --- | --- | --- |
| gateway | 9080 / 内部管理 9180 | 自研统一入口、原子热加载、插件、发现、限流熔断 |
| rbac | 8801 / 9801 | 管理员、会员、会话、角色菜单权限、验证码、邮件设置、黑名单 |
| blog | 8802 / 9802 | 文章草稿、发布快照、版本、分类标签、资源授权、统计 |
| product | 8803 / 9803 | 站点内容、文档模式与版本、下载清单、会员下载授权 |
| file | 8804 / 9804 | 上传、存储配置、私有文件、流式读写与下载 |
| scheduler | 无 HTTP 端口 | 通过 blog RPC 定时发布文章、清理资源 |
| web-site | 4310 | 官网 Node 页面服务，内网部署 |
| web-admin | 4311 | 管理后台 Node 页面服务，内网部署 |

```text
sitehub/
├── gateway/                   # Go 网关入口、声明式路由、插件、发现与控制面
├── service/
│   ├── rbac/                   # blog、product、file 使用相同分层
│   │   ├── api/               # .api、etc、internal/{handler,logic,svc,config,types}
│   │   ├── rpc/               # .proto、etc、internal/{server,logic,svc,config}、pb、client
│   │   ├── business/          # 本域业务规则与事务用例
│   │   ├── domain/            # 本域输入、状态及业务错误
│   │   ├── model/             # 本域仓库、实体、schema 初始化
│   │   └── dto/               # 域内转换类型，不作为跨服务协议
│   ├── blog/
│   ├── product/
│   ├── file/                  # client/ 将文件能力接口实现为 RPC 调用
│   └── scheduler/cronjob/     # 独立 main、etc、internal、worker
├── web/                       # admin、site、shared、Node server；npm workspaces
├── common/                    # 配置类型、错误转换、身份元数据、流处理、基础设施
├── tools/bootstrap/           # 单个服务的新数据库初始化
├── tools/gatewayctl/          # 网关校验、差异、发布与回滚
├── scripts/                   # goctl 生成、构建和多进程联调
├── tests/                     # 架构约束、联调辅助进程
├── deploy/                    # Dockerfile、Compose、网关和部署配置
└── go.mod
```

Handler 解析请求，API Logic 调用强类型 RPC 客户端；RPC Logic 校验身份并执行本域用例。数据库只在 RPC ServiceContext 中构造，跨域使用 RPC，不直接访问其他服务的表或仓库。

## 本地启动

需要 Go 1.24+、MySQL 8、Redis 6.2+ 及 Nacos。SQLite 用于隔离测试，执行涉及 SQLite 的测试和联调脚本时需要启用 CGO 并安装 C 编译器；生产示例使用 MySQL。先按 [治理运行指南](docs/gateway-governance-operations.md) 初始化 Nacos namespace 和九份配置，并设置 `NACOS_NAMESPACE_ID`、`NACOS_USERNAME`、`NACOS_PASSWORD`、`SERVICE_ADVERTISE_IP`（网关和其他服务可访问的本机地址）。四个服务各使用独立的新数据库及账号。以下命令均从本目录执行。

1. 配置环境变量：`SERVICE_KEY`、`AUTH_SECRET`、`STORAGE_SECRET` 分别使用至少 32 字符的随机值；设置 `RBAC_DB_DSN`、`BLOG_DB_DSN`、`PRODUCT_DB_DSN`、`FILE_DB_DSN`、`REDIS_PASSWORD`。MySQL DSN 示例：`rbac:password@tcp(127.0.0.1:3306)/sitehub_rbac?charset=utf8mb4&parseTime=true&loc=UTC`。
2. 在各服务 `etc/*.yaml` 中配置地址、超时和 SMTP。运行时由 Viper 读取配置，并展开文件值中的 `${ENV_NAME}`。管理员邮件验证码和会员注册邮件需要有效 SMTP。
3. 初始化每个空数据库。初次创建管理员前设置 `BOOTSTRAP_ADMIN_USERNAME`、`BOOTSTRAP_ADMIN_PASSWORD`（12–72 字节）、`BOOTSTRAP_ADMIN_EMAIL`。

```powershell
go run ./tools/bootstrap -service rbac -admin
go run ./tools/bootstrap -service file
go run ./tools/bootstrap -service blog
go run ./tools/bootstrap -service product
```

bootstrap 校验数据库归属，拒绝已有未知业务表的数据库；不会修改既有管理员凭据。RPC 启动会检查本域 schema 标记。之后重复检查 schema 时去掉 `-admin`。新增 schema 版本需要显式编写本域迁移；运行中不会自动修改表结构。

在独立终端启动各进程，先 rbac RPC、file RPC，再 blog/product RPC，最后 API 和 scheduler：

```powershell
go run ./service/rbac/rpc -f service/rbac/rpc/etc/rbac.yaml
go run ./service/file/rpc -f service/file/rpc/etc/file.yaml
go run ./service/blog/rpc -f service/blog/rpc/etc/blog.yaml
go run ./service/product/rpc -f service/product/rpc/etc/product.yaml
go run ./service/rbac/api -f service/rbac/api/etc/rbac-api.yaml
go run ./service/blog/api -f service/blog/api/etc/blog-api.yaml
go run ./service/product/api -f service/product/api/etc/product-api.yaml
go run ./service/file/api -f service/file/api/etc/file-api.yaml
go run ./service/scheduler/cronjob -f service/scheduler/cronjob/etc/scheduler.yaml
```

九个服务进程统一通过 Cobra 解析参数：`--config/-f` 指定配置文件，`--mode` 覆盖运行模式，`--help` 查看帮助。默认配置路径相对于当前工作目录；从项目根目录启动时使用上面的显式路径。帮助和参数错误不会连接数据库、Redis 或 RPC。

Viper 管理 YAML、JSON、TOML 文件、环境变量与已绑定参数。配置优先级为 **显式命令行参数 > 服务专属环境变量 > Nacos > 本地配置文件 > go-zero 默认值**。`common/command` 负责命令，`common/configloader` 合并配置后复用 go-zero 的类型解码、默认值和校验，不再使用 `conf.MustLoad` 读取服务配置。动态 `Runtime` 配置通过 Nacos 监听、候选校验和原子快照生效；其余字段变化需要重启。错误更新保留原快照，详情见治理运行指南。

| 进程 | 环境变量前缀 | 示例 |
| --- | --- | --- |
| rbac API / RPC | `RBAC_API_` / `RBAC_RPC_` | `RBAC_RPC_CACHE_ADDR` |
| blog API / RPC | `BLOG_API_` / `BLOG_RPC_` | `BLOG_API_PORT` |
| product API / RPC | `PRODUCT_API_` / `PRODUCT_RPC_` | `PRODUCT_RPC_DATABASE_MAXOPEN` |
| file API / RPC | `FILE_API_` / `FILE_RPC_` | `FILE_RPC_STORAGEROOT` |
| scheduler | `SCHEDULER_` | `SCHEDULER_JOBTIMEOUT` |

变量名由前缀加配置路径转大写组成，层级之间使用 `_`，字段名内部不额外插入下划线。例如 `RPC.Timeout` 对应 `BLOG_API_RPC_TIMEOUT`；集合使用 JSON，例如 `BLOG_API_TRUSTEDPROXYCIDRS='["172.30.0.0/24"]'`。文件中未出现的字段也可以被覆盖；布尔值使用 `true/false`，`JobTimeout` 等时长字段使用 `30s/2m`，go-zero 的 `Timeout` 数值字段仍使用毫秒。

```powershell
$env:BLOG_API_PORT = '8812'
$env:BLOG_API_RPC_TIMEOUT = '5000'
go run ./service/blog/api --config service/blog/api/etc/blog-api.yaml --mode dev
go run ./service/scheduler/cronjob --help
```

`${SERVICE_KEY}`、`${BLOG_DB_DSN}` 等文件占位符继续读取同名环境变量。占位符在文件解析后展开，密钥中的引号、冒号和换行不会改变 YAML 结构；直接设置的服务专属环境变量按字面值读取。未设置的占位符展开为空字符串，显式设置为空的环境变量会覆盖文件值。

本地文件默认存放于 `data/files`。资源 URL 是 `/v1/...` 相对路径，浏览器访问应使用统一网关域名。API `/health` 会调用本域 RPC 并检查数据库；rbac 同时检查 Redis。下游业务依赖及 SMTP/S3 的可用性由业务请求验证。

## 新接口约定

HTTP 前缀为 `/v1/rbac`、`/v1/blog`、`/v1/product`、`/v1/file`，以各服务 `.api` 和 `.proto` 为契约。GET 参数放 query；修改操作使用 POST JSON 和命名请求体。成功响应为 `{ "code": 0, "message": "ok", "data": ... }`。protobuf 的 int64/uint64 在响应 JSON 中表示为字符串，客户端不要转成可能丢失精度的浮点数。

登录返回 Bearer token；后续请求设置 `Authorization: Bearer <token>`。管理员与会员身份分离；启用 TOTP 后敏感管理请求还需 `X-TOTP-Token`。内部 RPC 使用配置的服务密钥，最终用户不能通过 HTTP 指定服务身份或 actor。API 不记录请求正文，RPC 不记录业务载荷，SQL 日志不展开绑定参数。

文章操作带 `requestID` 和 `version`，文档修改带 `revision`；冲突返回 HTTP 409。文章公开读取来自发布快照。文件流接口及功能清单见 [功能迁移说明](docs/full-migration.md)。`web/admin` 与 `web/site` 已按新契约接入。前端源码、共享资源和 Node 启动入口均在本项目中，详见 [前端运行说明](web/README.md)。

## 前端启动

需要 Node 22.12+，在本项目 `web` 目录安装依赖并构建，然后在两个终端分别启动：

```powershell
cd web
npm ci
npm run build
npm run start:site
# 另一个终端进入 web 后执行：npm run start:admin
```

官网入口 `http://sitehub.local:8082`，后台入口 `http://admin.sitehub.local:8082`，均经过 自研 Go 网关。Node 只提供页面；开发 Vite 使用 `npm run dev:site` / `npm run dev:admin`，只代理统一网关。配置、会话和独立镜像见 [前端运行说明](web/README.md)。

## 容器部署

复制 `deploy/.env.example` 为 `deploy/.env` 并填写随机密钥、数据库密码、管理员及邮件配置。Compose 使用四个 MySQL 实例、Redis、九个业务 Go 进程、两个 Node 页面进程、自研 Go 网关、Nacos 与网关专用 etcd；仅网关发布 8082。初次启动及配置发布步骤见 [治理运行指南](docs/gateway-governance-operations.md)。

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml --profile init build
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml --profile init run --rm rbac-bootstrap -service rbac -admin
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml --profile init run --rm file-bootstrap
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml --profile init run --rm blog-bootstrap
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml --profile init run --rm product-bootstrap
docker compose --env-file deploy/.env -f deploy/docker-compose.yaml up -d
```

`TrustedProxyCIDRs` 必须与实际网关网段匹配，网关必须覆盖 `X-Real-IP`。本地直连默认不信任转发头。Compose 示例使用 Docker 常见私网段，生产应收窄至网关实际子网。RPC 和数据库保持在内部网络；对外 TLS 在网关配置。scheduler 当前部署单副本。

## 生成和验证

代码生成使用 goctl 1.10.2、protoc 31.1、protoc-gen-go 和 protoc-gen-go-grpc。`.api` / `.proto` 是生成源，`--style go_zero` 统一命名。生成文件由工具维护，已有 Logic/Svc 的自定义实现保留；重新生成后仍需审查差异并执行测试。

```powershell
./scripts/generate.ps1
go test ./...
go vet ./...
go build ./...
python scripts/smoke.py
```

`smoke.py` 自动构建并启动四组 API/RPC、测试 Redis 和 scheduler，使用临时独立 SQLite 数据库与文件目录；结束时关闭子进程。测试结果与进程日志在 `.tmp/smoke/runs/`，`--skip-build` 可复用已构建二进制。测试邮件验证码由夹具写入，不发送真实邮件。构建前端并准备 Playwright 浏览器后，增加 `--governance` 启动隔离的 自研 Go 网关/Nacos/etcd 并验证注册发现和配置；增加 `--frontends` 同时通过真实网关联调两个 Node 前端（需要 Docker，Windows 使用 WSL Ubuntu），生成浏览器报告和桌面/手机截图。

本地验证不等同于真实 MySQL、SMTP、对象存储或 Docker 部署验证；这些环境仍需上线前联调。本项目提供新服务 schema 初始化，不自动转换已有生产数据。

详细边界见 [架构说明](ARCHITECTURE.md)，参考项目与采用依据见 [架构参考](docs/architecture-reference.md)。

设计依据见 [统一网关、服务注册发现与配置中心方案](docs/gateway-governance-design.md)，已接入的能力、热更新边界及运维步骤见 [治理运行指南](docs/gateway-governance-operations.md)。

## 仓库维护

仓库保留单元测试、架构测试和可重复运行的集成测试。临时日志、截图、测试数据库、编译产物、前端 `dist`、`node_modules` 和生成的编辑器资源不提交；所需资源通过 `npm ci`、`npm run build` 或测试命令重新生成。依赖锁文件、`.api`、`.proto` 和对应生成代码均应保留。

配置只提交空凭据示例；实际环境变量文件、私钥、运行时数据和网关发布记录留在本地。开发与验证命令见 [贡献说明](CONTRIBUTING.md)。
