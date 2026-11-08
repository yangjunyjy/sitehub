# scheduler

独立 go-zero cronjob 进程。`internal/svc` 创建 blog RPC 客户端与发布/清理能力适配器；`internal/logic` 管理触发、超时和取消；`worker` 组装任务生命周期。任务没有数据库依赖。

从项目根目录运行：

```powershell
go run ./service/scheduler/cronjob --config service/scheduler/cronjob/etc/scheduler.yaml
```

启动参数由 Cobra 解析，支持 `--config/-f`、`--mode` 和 `--help`；配置由 Viper 加载。`SCHEDULER_` 前缀的环境变量覆盖文件值，例如 `SCHEDULER_JOBTIMEOUT=30s`、`SCHEDULER_BLOGRPC_TIMEOUT=30000`。显式 `--mode` 优先于 `SCHEDULER_MODE`，未配置的字段使用 go-zero 默认值。

配置 `SERVICE_KEY`、`BlogRPC.Endpoints`、`PublishSchedule`、`CleanupSchedule`、`JobTimeout`。默认每分钟发布、每小时清理、单次两分钟。main 启动任务后等待退出信号，随后取消任务并等待停止。

同一进程防止任务重叠；当前使用单副本，没有跨进程选主或租约。发布状态与清理重试记录由 blog 持久化，scheduler 不自行判断文章状态或执行文件删除。
