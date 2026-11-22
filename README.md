# SiteHub — 基于 Go 的站点与产品运营平台

SiteHub 提供站点内容、产品文档、下载中心、身份权限、文件和定时任务管理。官网演示内容可通过后台配置，Deploy 是当前演示产品名称。

项目由 RBAC、博客、产品三个业务域组成。博客 API 直接执行文章、文件和调度用例，产品 API 直接执行文档与下载用例。只有 RBAC 提供独立 RPC 服务。

**迁移尚未全部完成。** 主体架构与主要业务链路已接通，待补功能、实现问题和验收范围见 [后端迁移缺口与接手清单](docs/migration-gaps.md)。

## 项目标识

| 项目项 | 当前名称 |
| --- | --- |
| 项目目录 / Go module | `sitehub` |
| 前端 workspace | `@sitehub/admin`、`@sitehub/product`、`@sitehub/blog` |
| 开发域名 | `product.sitehub.local`、`blog.sitehub.local`、`admin.sitehub.local` |
| 会话 Cookie | `sitehub_admin_session`、`sitehub_site_session` |
| etcd 默认 namespace / group | `sitehub-dev` / `SITEHUB` |
| 网关 etcd 配置键 | `/sitehub/gateway/v1/config` |
| Compose 项目名 | `sitehub` |
| 默认业务数据库 | `sitehub_rbac`、`sitehub_blog`、`sitehub_product`、`sitehub_file` |

已有部署的数据库、配置中心内容和持久卷不会随工程更名自动迁移；使用新命名部署前，需要显式迁移对应数据与配置。各业务服务仍按其职责命名。

## 服务与目录

| 服务 | HTTP / RPC 默认端口 | 职责 |
| --- | --- | --- |
| gateway | HTTP 9080 / 管理 API 9180 / 控制台 9190 | HTTP 域名、IP 和路径路由，TCP/UDP 端口转发，配置热加载，独立运维账号 |
| rbac | 8801 / 9801 | 管理员、会员、会话、角色菜单权限、验证码、邮件设置、黑名单 |
| blog | 8802 | 文章、分类标签、资源授权与统计；文件上传、存储、下载；定时发布与清理 |
| product | 8803 | 站点内容、文档模式与版本、下载清单、会员下载授权 |
| web-product | 4310 | 产品页面服务 |
| web-blog | 4312 | 博客页面服务 |
| web-admin | 4311 | 管理后台 Node 页面服务，内网部署 |

```text
sitehub/
├── gateway/                   # Go 网关入口、声明式路由、插件、发现与控制面
├── service/
│   ├── rbac/                   # 身份服务提供 API 与 RPC
│   │   ├── api/               # .api、etc、internal/{handler,logic,svc,config,types}
│   │   ├── rpc/               # .proto、etc、internal/{server,logic,svc,config}、pb、client
│   │   ├── business/          # 本域业务规则与事务用例
│   │   ├── domain/            # 本域输入、状态及业务错误
│   │   ├── model/             # 本域仓库、实体、schema 初始化
│   │   └── dto/               # 域内转换类型，不作为跨服务协议
│   ├── blog/                  # 博客、文件与调度的完整服务边界
│   │   ├── api/              # blog.api、internal/{handler,logic,svc,config,types}
│   │   ├── business/         # 文章、分类标签、统计与文件存储用例
│   │   ├── model/            # 仓库、实体与 schema（包含存储库初始化）
│   │   ├── contract/         # 对外公开的存储能力接口与类型
│   │   ├── client/           # 通过内部 HTTP 访问文件设置
│   │   └── jobs/             # 进程内定时发布与资源清理
│   └── product/
├── web/                       # admin、product、blog 各自维护 Node server；npm workspaces
├── common/                    # 配置类型、错误转换、身份元数据、流处理、基础设施
├── tools/bootstrap/           # 单个服务的新数据库初始化
├── tools/gatewayctl/          # 网关校验、差异、发布与回滚
├── scripts/                   # goctl 生成、构建和多进程联调
├── tests/                     # 架构约束、联调辅助进程
├── deploy/                    # Dockerfile、Compose、网关和部署配置
└── go.mod
```

博客和产品的 HTTP Handler 解析请求，Logic 校验权限并调用本域用例。各自的 ServiceContext 管理数据库和 RBAC 客户端。

博客维护文件实体，产品维护发布清单。产品通过内部 HTTP 校验发布文件，博客在下载前通过内部 HTTP 向产品核对会员身份及发布状态。

## 本地启动

需要 Go 1.24+、MySQL 8、Redis 6.2+ 和 etcd。SQLite 用于隔离测试，涉及 SQLite 的测试需启用 CGO 并安装 C 编译器；生产使用 MySQL。设置 `ETCD_ENDPOINT`、`CONTROL_NAMESPACE` 和 `SERVICE_ADVERTISE_IP`（调用方可达的本机地址），可选 `ETCD_USERNAME`、`ETCD_PASSWORD`。先按 [etcd 配置说明](deploy/etcd/README.md) 用 `go run ./tools/configctl --dir deploy/etcd/config` 导入四份配置。本机开发使用独立 namespace 并调整存储路径。保留 rbac、blog、product、file 四个逻辑数据库及账号，blog-api 同时连接 blog/file 两库，不自动合库。以下命令均从项目根目录执行。

1. 配置环境变量：`SERVICE_KEY`、`AUTH_SECRET`、`STORAGE_SECRET` 分别使用至少 32 字符的随机值；设置 `RBAC_DB_DSN`、`BLOG_DB_DSN`、`PRODUCT_DB_DSN`、`FILE_DB_DSN`、`REDIS_PASSWORD`。MySQL DSN 示例：`rbac:password@tcp(127.0.0.1:3306)/sitehub_rbac?charset=utf8mb4&parseTime=true&loc=UTC`。
2. 在各服务 `etc/*.yaml` 中配置地址、超时和 SMTP。运行时由 Viper 读取配置，并展开文件值中的 `${ENV_NAME}`。管理员邮件验证码和会员注册邮件需要有效 SMTP。
3. 初始化每个空数据库。初次创建管理员前设置 `BOOTSTRAP_ADMIN_USERNAME`、`BOOTSTRAP_ADMIN_PASSWORD`（12–72 字节）、`BOOTSTRAP_ADMIN_EMAIL`。

```powershell
go run ./tools/bootstrap -service rbac -admin
go run ./tools/bootstrap -service file
go run ./tools/bootstrap -service blog
go run ./tools/bootstrap -service product
```

bootstrap 校验数据库归属，拒绝已有未知业务表的数据库；不会修改既有管理员凭据。业务服务启动会检查本域 schema 标记。之后重复检查 schema 时去掉 `-admin`。新增 schema 版本需要显式编写本域迁移；运行中不会自动修改表结构。

升级已有部署前，保留原数据库凭据、存储密钥和文件卷，把原 RPC 配置迁入各自的 API 配置。核对后执行 `bash init.sh config --overwrite` 发布合并后的配置；停止旧 blog-rpc、product-rpc 容器再启动新版本，防止重复运行调度。

本地私有文章资源存放在存储目录加 `-private` 的同级目录中。默认分别为 `/data/files` 与 `/data/files-private`，新部署已分别挂载 `files`、`private-files` 两个持久卷。旧部署可能只有前一个目录挂卷：删除旧文件容器之前，必须检查并导出其 `/data/files-private`，将原文件复制到新私有卷中，在新容器内保持原绝对路径和 `STORAGE_SECRET`，再验证历史附件可以读取。后台若配置过其他本地存储路径，其普通文件目录和对应的 `-private` 目录也需要分别持久化；仅保留数据库和公开文件卷不能完整迁移私有附件。

在独立终端启动 RBAC RPC 和三个 API。博客与产品各自在 API 进程内启动 RPC。

```powershell
go run ./service/rbac/rpc -f service/rbac/rpc/etc/rbac.yaml
go run ./service/rbac/api -f service/rbac/api/etc/rbac-api.yaml
go run ./service/blog/api -f service/blog/api/etc/blog-api.yaml
go run ./service/product/api -f service/product/api/etc/product-api.yaml
```

四个业务进程统一通过 Cobra 解析参数：`--config/-f` 指定配置文件，`--mode` 覆盖运行模式，`--help` 查看帮助。默认配置路径相对于当前工作目录；从项目根目录启动时使用上面的显式路径。帮助和参数错误不会连接数据库、Redis 或 RPC。

Viper 管理 YAML、JSON、TOML 文件、环境变量与已绑定参数。配置优先级为 **显式命令行参数 > 服务专属环境变量 > etcd > 本地配置文件 > go-zero 默认值**。`common/command` 负责命令，`common/configloader` 合并配置后复用 go-zero 的类型解码、默认值和校验，不再使用 `conf.MustLoad` 读取服务配置。动态 `Runtime` 配置通过 etcd 监听、候选校验和原子快照生效；其余字段变化需要重启。错误更新保留原快照，详情见治理运行指南。

| 进程 | 环境变量前缀 | 示例 |
| --- | --- | --- |
| rbac API / RPC | `RBAC_API_` / `RBAC_RPC_` | `RBAC_RPC_CACHE_ADDR` |
| blog API | `BLOG_API_` | `BLOG_API_PORT`、`BLOG_API_SCHEDULER_ENABLED`、`BLOG_API_STORAGEROOT` |
| product API | `PRODUCT_API_` | `PRODUCT_API_DATABASE_MAXOPEN` |

变量名由前缀加配置路径转大写组成，层级之间使用 `_`，字段名内部不额外插入下划线。例如 `RPC.Timeout` 对应 `BLOG_API_RPC_TIMEOUT`；集合使用 JSON，例如 `BLOG_API_TRUSTEDPROXYCIDRS='["172.30.0.0/24"]'`。文件中未出现的字段也可以被覆盖；布尔值使用 `true/false`，`Scheduler.JobTimeout` 等时长字段使用 `30s/2m`，go-zero 的 `Timeout` 数值字段仍使用毫秒。

```powershell
$env:BLOG_API_PORT = '8812'
$env:BLOG_API_RPC_TIMEOUT = '5000'
go run ./service/blog/api --config service/blog/api/etc/blog-api.yaml --mode dev
```

`${SERVICE_KEY}`、`${BLOG_DB_DSN}` 等文件占位符继续读取同名环境变量。占位符在文件解析后展开，密钥中的引号、冒号和换行不会改变 YAML 结构；直接设置的服务专属环境变量按字面值读取。未设置的占位符展开为空字符串，显式设置为空的环境变量会覆盖文件值。

本地文件默认存放于 `data/files`。资源 URL 是 `/v1/...` 相对路径，浏览器访问应使用统一网关域名。API `/health` 会调用本域 RPC 并检查数据库；blog 同时检查博客库与文件库，rbac 同时检查 Redis。下游业务依赖及 SMTP/S3 的可用性由业务请求验证。

## 新接口约定

HTTP 前缀为 `/v1/rbac`、`/v1/blog`、`/v1/product`、`/v1/file`，文件接口由 blog-api 提供，前缀继续保留 `/v1/file`；以各服务 `.api` 和 `.proto` 为契约。GET 参数放 query；修改操作使用 POST JSON 和命名请求体。成功响应为 `{ "code": 0, "message": "ok", "data": ... }`。protobuf 的 int64/uint64 在响应 JSON 中表示为字符串，客户端不要转成可能丢失精度的浮点数。

登录返回 Bearer token；后续请求设置 `Authorization: Bearer <token>`。管理员与会员身份分离；启用 TOTP 后敏感管理请求还需 `X-TOTP-Token`。内部 RPC 使用配置的服务密钥，最终用户不能通过 HTTP 指定服务身份或 actor。API 不记录请求正文，RPC 不记录业务载荷，SQL 日志不展开绑定参数。

文章操作带 `requestID` 和 `version`，文档修改带 `revision`；冲突返回 HTTP 409。文章公开读取来自发布快照。文件流接口及功能清单见 [功能迁移说明](docs/full-migration.md)。三个前端 已按新契约接入。前端源码、共享资源和 Node 启动入口均在本项目中，详见 [前端运行说明](web/README.md)。

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

CentOS 7.9 虚拟机使用 `docker run` 部署中间件、连接 Windows FlClash 代理的步骤见 [虚拟机中间件部署](deploy/vm/README.md)。业务与前端在本地编译，只把编译产物和配置传到云机；云机用这些产物构建运行镜像，再通过 Compose 连接本地组件。完整操作见 [编译产物发布与初始化脚本](deploy/cloud/README.md)。

在有 Go、Node.js/npm 和 PowerShell 的本地电脑执行，不要求本地 Docker：

```powershell
./scripts/package-cloud.ps1 -Version 2026.11.17.1 -Platform linux/amd64
```

云机无法连接 Docker Hub 时，先在本地虚拟机执行发布包中的 `export-base-images.sh`，经代理准备 Go 运行基础镜像及 Node 镜像，再将镜像归档传到云机并 `docker load`。具体命令见部署指南；云机构建不再拉镜像或执行 apt 下载。

将 `output/cloud/sitehub-2026.11.17.1.tar.gz` 上传云机，解压进入 `sitehub-2026.11.17.1`，基础镜像导入后执行：

```bash
cp -n .env.example .env
chmod 600 .env
vi .env
bash build-images.sh
bash init.sh
bash start.sh
```

发布包包含八个 Linux 二进制、三套前端产物及 Node 页面服务 bundle、运行镜像定义、Compose、配置和脚本。云机不需要项目源码、Go 或 npm，镜像构建只复制已有编译产物。`.env` 复用虚拟机中已生成的业务库、Redis 密码及可选 etcd 凭据，填写业务密钥、管理员及邮件配置。

前端默认使用当前浏览器访问地址，跨站链接可在云机 `.env` 中配置。使用同一公网 IP 访问三个站点时，可让宿主机 Nginx 按端口转发，网关保留三个内部 Host，同时将浏览器登录 Origin 切换到实际 IP 地址。见[公网 IP 部署步骤](gateway/README.md#线上通过同一个公网-ip-访问三个站点)。

`init.sh` 显式初始化四个 MySQL 逻辑库的业务表、超级管理员和 etcd 中缺失的四份配置键；可用 `init.sh database`、`init.sh config` 分别执行。已有管理员密码和用户配置保持原值，显式更新远端配置使用 `init.sh config --overwrite`。`start.sh` 仅启动应用，不建表或构建。`-ConfigOnly` 仅生成缺少编译产物的模板预览包，不能用于完整部署。

`TrustedProxyCIDRs` 必须与实际网关网段匹配，网关覆盖 `X-Real-IP`。RPC 保持在云机内部网络，仅发布网关业务端口。blog-api 的 Scheduler 目前只允许一个实例启用，额外副本应设置 `Scheduler.Enabled: false`。

## 生成和验证

运行 `scripts/generate.ps1` 从各域的 `.api` 生成 HTTP 适配层。只有 RBAC 还需要 protoc 和 gRPC 插件。

```powershell
./scripts/generate.ps1
go test ./...
go vet ./...
go build ./...
python scripts/smoke.py
```

`smoke.py` 自动构建并启动四个业务进程、测试 Redis 和 blog 内嵌调度，使用临时独立 SQLite 数据库与文件目录；结束时关闭子进程。测试结果与进程日志在 `.tmp/smoke/runs/`，`--skip-build` 可复用已构建二进制。测试邮件验证码由夹具写入，不发送真实邮件。构建前端并准备 Playwright 浏览器后，增加 `--governance` 启动隔离的 自研 Go 网关/etcd 并验证注册发现和配置；增加 `--frontends` 同时通过真实网关联调三个 Node 前端（需要 Docker，Windows 使用 WSL Ubuntu），生成浏览器报告和桌面/手机截图。

本地验证不等同于真实 MySQL、SMTP、对象存储或 Docker 部署验证；这些环境仍需上线前联调。本项目提供新服务 schema 初始化，不自动转换已有生产数据。

详细边界见 [架构说明](ARCHITECTURE.md)，参考项目与采用依据见 [架构参考](docs/architecture-reference.md)。

设计依据见 [统一网关、服务注册发现与配置中心方案](docs/gateway-governance-design.md)，已接入的能力、热更新边界及运维步骤见 [etcd 配置指南](deploy/etcd/README.md)。

## 仓库维护

仓库保留单元测试、架构测试和可重复运行的集成测试。临时日志、截图、测试数据库、编译产物、前端 `dist`、`node_modules` 和生成的编辑器资源不提交；所需资源通过 `npm ci`、`npm run build` 或测试命令重新生成。依赖锁文件、`.api`、`.proto` 和对应生成代码均应保留。

配置只提交空凭据示例；实际环境变量文件、私钥、运行时数据和网关发布记录留在本地。开发与验证命令见 [贡献说明](CONTRIBUTING.md)。
