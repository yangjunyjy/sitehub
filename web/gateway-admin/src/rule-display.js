const pluginNames = {
  "browser-session": "浏览器会话",
  "circuit-breaker": "故障熔断",
  "deny-paths": "路径拦截",
  "limit-count": "窗口限流",
  "proxy-rewrite": "请求改写",
  "rate-limit": "请求限速",
  "request-id": "请求追踪",
  "static-response": "固定响应",
  "strip-credentials": "移除身份凭据",
};

import { fieldChanges } from "./manifest-changes";

function pluginList(plugins = {}) {
  return (
    Object.entries(plugins || {})
      .map(
        ([key, value]) =>
          (value === null ? "停用" : "") + (pluginNames[key] || key),
      )
      .join("、") || "未配置插件"
  );
}

export function hostLabel(host) {
  if (!host || host === "*") return "全部主机";
  if (host.startsWith("*:")) return "全部主机 · 端口 " + host.slice(2);
  return host;
}

export function pathLabel(paths) {
  return (paths || [])
    .map((path) => {
      if (path === "/*") return "全部路径";
      return path.endsWith("/*") ? "前缀 " + path.slice(0, -1) : "精确 " + path;
    })
    .join(" 或 ");
}

export function describe(section, row, id, doc) {
  const routes = doc.routes || [];
  switch (section) {
    case "routes": {
      const effective = {
        ...doc.plugins,
        ...doc.profiles?.[row.profile],
        ...row.plugins,
      };
      const result = effective["static-response"];
      const pool = doc.upstreams?.[row.upstream];
      const destination = row.upstream
        ? row.upstream
        : result
          ? "返回 HTTP " + result.status
          : "由插件响应";
      const target =
        pool?.service ||
        pool?.nodes?.map((n) => n.url).join("、") ||
        result?.body;
      return [
        [row.id, "优先级 " + row.priority],
        [hostLabel(row.host), pathLabel(row.paths)],
        [row.methods?.join("、") || "全部方法"],
        [destination, target],
        [
          row.profile || "全局设置",
          pluginList(
            Object.fromEntries(
              Object.entries(effective).filter(([, v]) => v !== null),
            ),
          ),
        ],
      ];
    }
    case "streams":
      return [
        [row.id],
        [row.protocol.toUpperCase()],
        [row.listen],
        [row.target],
      ];
    case "upstreams":
      return [
        [id],
        [
          row.service ? "服务发现" : "固定节点",
          row.service || row.nodes?.map((n) => n.url).join("、"),
        ],
        [row.health_path, row.health_interval_ms / 1000 + " 秒检查一次"],
        [routes.filter((r) => r.upstream === id).length + " 条路由"],
      ];
    case "policies":
      return [
        [id || "全局设置"],
        [pluginList(row)],
        [
          id
            ? routes.filter((r) => r.profile === id).length + " 条路由"
            : "全部 HTTP 路由",
        ],
      ];
    case "releases": {
      const unknown = row.status === "corrupt";
      let revision = "—";
      if (row.previous_revision) {
        revision = String(row.previous_revision);
        if (row.applied_revision) revision += " → " + row.applied_revision;
        else if (["pending", "unconfirmed"].includes(row.status))
          revision += " → 待确认";
      }
      return [
        [new Date(row.created_at).toLocaleString()],
        [unknown ? "—" : row.actor || "gatewayctl"],
        [revision],
        [unknown ? "—" : row.rollback_of ? "回滚" : "发布"],
        [
          {
            applied: "已应用",
            committed: "已提交",
            pending: "待确认",
            unconfirmed: "未确认",
            rejected: "已拒绝",
            conflict: "版本冲突",
            corrupt: "记录损坏",
          }[row.status] || row.status,
          row.message,
        ],
      ];
    }
  }
}

export const columns = {
  routes: ["规则", "请求主机与路径", "请求方法", "处理结果", "生效策略"],
  streams: ["规则", "传输协议", "网关监听", "转发到"],
  upstreams: ["上游服务", "目标地址", "健康检查", "使用情况"],
  policies: ["策略", "处理内容", "作用范围"],
  releases: ["操作时间", "操作人", "修订", "操作", "结果"],
};

export function describeChange(change, previous, next) {
  const [kind, id] = change.slice(2).split("/");
  const section = {
    route: "routes",
    stream: "streams",
    upstream: "upstreams",
    profile: "policies",
    plugins: "policies",
  }[kind];
  function snapshot(doc) {
    if (kind === "route" || kind === "stream")
      return doc[section]?.find((r) => r.id === id);
    if (kind === "upstream") return doc.upstreams?.[id];
    if (kind === "profile") return doc.profiles?.[id];
    return doc.plugins;
  }
  const type = { "+": "新增", "~": "修改", "-": "删除" }[change[0]];
  const resource = {
    route: "HTTP 路由",
    stream: "四层转发",
    upstream: "上游服务",
    profile: "策略模板",
    plugins: "全局策略",
  }[kind];
  const before = snapshot(previous),
    after = snapshot(next);
  const affected = new Set();
  for (const doc of [previous, next]) {
    for (const route of doc.routes || []) {
      if (
        (kind === "upstream" && route.upstream === id) ||
        (kind === "profile" && route.profile === id) ||
        kind === "plugins"
      )
        affected.add(route.id);
    }
  }
  return {
    key: change,
    title: type + resource + " · " + id,
    fields: fieldChanges(before, after),
    affected: [...affected],
  };
}
