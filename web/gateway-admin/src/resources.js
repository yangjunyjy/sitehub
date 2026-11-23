import { reactive } from "vue";

export function createEditor(section, doc, key, global) {
  const fields = [];
  function field(label, value, options = {}) {
    const model = reactive({ label, value: String(value ?? ""), ...options });
    fields.push(model);
    return model;
  }
  const editors = {
    routes: routeFields,
    streams: streamFields,
    upstreams: upstreamFields,
  };
  const save = editors[section]
    ? editors[section](doc, key, field)
    : policyFields(doc, key, global, field);
  return { fields, save };
}

const integer = (text, min, max) => {
  const n = Number(text);
  if (!text.trim() || !Number.isSafeInteger(n) || n < min || n > max)
    throw Error("请填写范围内的整数");
  return n;
};
const words = (value) => value.split(/[\s,，]+/).filter(Boolean);
const jsonObject = (control) => {
  let parsed;
  try {
    parsed = JSON.parse(control.value || "{}");
  } catch {
    throw Error("插件配置的 JSON 格式有误");
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== "object")
    throw Error("插件配置须为 JSON 对象");
  return parsed;
};
const identifier = (field, label, key) =>
  field(label, key, {
    required: true,
    readOnly: !!key,
    pattern: "[a-zA-Z0-9_\\-]{1,64}",
  });
const plugins = (field, value) =>
  field("插件配置（JSON）", JSON.stringify(value || {}, null, 2), {
    multiline: true,
    group: "高级设置",
  });
const advancedNumbers = [
  ["connect_timeout_ms", "连接超时（毫秒）", 3000, 1, 60000],
  ["response_header_timeout_ms", "响应头超时（毫秒）", 10000, 1, 3600000],
  ["io_timeout_ms", "读写超时（毫秒）", 30000, 1, 3600000],
  ["health_interval_ms", "健康检查间隔（毫秒）", 5000, 100, 60000],
  ["health_timeout_ms", "健康检查超时（毫秒）", 1000, 1, 60000],
  ["health_failures", "连续失败阈值", 3, 1, 100],
];

function routeFields(doc, key, field) {
  const row = doc.routes?.find((r) => r.id === key) || {};
  const id = identifier(field, "路由标识", key);
  const host = field("匹配主机", row.host, {
    placeholder: "全部主机",
    group: "匹配条件",
    hint: "可留空，或填写域名、IP、主机:端口。*:8443 匹配请求主机中的端口。",
  });
  const paths = field("请求路径", row.paths?.join("\n") || "/*", {
    required: true,
    multiline: true,
    placeholder: "/api/*",
    hint: "每行一条。/api 精确匹配，/api/* 匹配目录下的请求，/* 匹配全部路径。",
    group: "匹配条件",
  });
  const methods = field("请求方法", row.methods?.join(", "), {
    group: "匹配条件",
    placeholder: "全部方法",
    hint: "可填写 GET, POST 等多个方法。",
  });
  const priority = field("优先级", row.priority ?? 0, {
    type: "number",
    required: true,
    group: "匹配条件",
    hint: "数值较大的规则先匹配。",
  });
  const inherited = { ...doc.plugins, ...doc.profiles?.[row.profile] };
  const effective = { ...inherited, ...row.plugins };
  const direct = effective["static-response"];
  const action = field(
    "匹配后执行",
    row.upstream || !key ? "转发到服务" : direct ? "直接响应" : "响应插件",
    {
      choices: ["转发到服务", "直接响应", "响应插件"],
      group: "处理方式",
    },
  );
  const upstream = field("目标上游", row.upstream, {
    choices: ["", ...Object.keys(doc.upstreams || {})],
    emptyLabel: "选择上游服务",
    required: true,
    group: "处理方式",
    show: () => action.value === "转发到服务",
  });
  const status = field("响应状态码", direct?.status ?? 200, {
    required: true,
    type: "number",
    group: "处理方式",
    show: () => action.value === "直接响应",
  });
  const body = field("响应正文", direct?.body, {
    multiline: true,
    group: "处理方式",
    show: () => action.value === "直接响应",
  });
  const headers = field(
    "响应头（JSON）",
    JSON.stringify(direct?.headers || {}, null, 2),
    {
      multiline: true,
      group: "高级设置",
      show: () => action.value === "直接响应",
    },
  );
  const profile = field("插件策略", row.profile, {
    choices: ["", ...Object.keys(doc.profiles || {})],
    emptyLabel: "仅应用全局策略",
    group: "处理方式",
  });
  const otherPlugins = { ...row.plugins };
  delete otherPlugins["static-response"];
  const config = plugins(field, otherPlugins);
  config.hint = "自定义插件按插件名填写。固定响应使用「处理方式」中的字段。";
  return () => {
    if (!key && doc.routes?.some((r) => r.id === id.value))
      throw Error("路由标识已存在");
    const value = {
      id: id.value,
      host: host.value.trim().toLowerCase(),
      paths: words(paths.value),
      priority: integer(priority.value, -2147483648, 2147483647),
    };
    if (methods.value.trim())
      value.methods = words(methods.value.toUpperCase());
    if (action.value === "转发到服务") {
      if (!upstream.value)
        throw Error("请选择目标上游。可先到「上游服务」新建服务。");
      value.upstream = upstream.value;
    }
    if (profile.value) value.profile = profile.value;
    const overrides = jsonObject(config);
    if ("static-response" in overrides)
      throw Error("请在「处理方式」设置固定响应");
    const inheritedResponse =
      doc.profiles?.[value.profile] &&
      Object.hasOwn(doc.profiles[value.profile], "static-response")
        ? doc.profiles[value.profile]["static-response"]
        : doc.plugins?.["static-response"];
    if (action.value !== "直接响应") {
      if (inheritedResponse || row.plugins?.["static-response"] === null)
        overrides["static-response"] = null;
    } else {
      const response = {
        status: integer(status.value, 200, 599),
        body: body.value,
      };
      const responseHeaders = jsonObject(headers);
      if (Object.keys(responseHeaders).length)
        response.headers = responseHeaders;
      if ([204, 304].includes(response.status) && response.body)
        throw Error("204 和 304 响应须留空正文");
      if (
        row.plugins?.["static-response"] ||
        !inheritedResponse ||
        response.status !== inheritedResponse.status ||
        response.body !== (inheritedResponse.body || "") ||
        JSON.stringify(responseHeaders) !==
          JSON.stringify(inheritedResponse.headers || {})
      ) {
        const original = row.plugins?.["static-response"];
        // Preserve omitted optional fields when reopening an existing response.
        const unchanged =
          original &&
          original.status === response.status &&
          (original.body ?? "") === response.body &&
          JSON.stringify(original.headers || {}) ===
            JSON.stringify(responseHeaders);
        overrides["static-response"] = unchanged ? original : response;
      }
    }
    if (Object.keys(overrides).length) value.plugins = overrides;
    if (key) doc.routes.splice(doc.routes.indexOf(row), 1, value);
    else (doc.routes ||= []).push(value);
  };
}
function streamFields(doc, key, field) {
  const row = doc.streams?.find((r) => r.id === key) || {};
  const id = identifier(field, "规则标识", key);
  const protocol = field("传输协议", (row.protocol || "tcp").toUpperCase(), {
    choices: ["TCP", "UDP"],
  });
  const listen = field("监听地址", row.listen, {
    required: true,
    placeholder: "127.0.0.1:15432",
    group: "端口转发",
    hint: "填写网关上的地址与端口。0.0.0.0 表示所有 IPv4 网卡。",
  });
  const target = field("目标地址", row.target, {
    required: true,
    placeholder: "db.internal:5432",
    group: "端口转发",
    hint: "填写后端的主机与端口，IPv6 地址用方括号包围。",
  });
  return () => {
    const value = {
      id: id.value,
      protocol: protocol.value.toLowerCase(),
      listen: listen.value.trim(),
      target: target.value.trim(),
    };
    if (!key && doc.streams?.some((r) => r.id === value.id))
      throw Error("规则标识已存在");
    if (
      doc.streams?.some(
        (r) =>
          r !== row &&
          r.protocol === value.protocol &&
          r.listen === value.listen,
      )
    )
      throw Error("同协议的监听地址已被其他规则使用");
    if (key) doc.streams.splice(doc.streams.indexOf(row), 1, value);
    else (doc.streams ||= []).push(value);
  };
}
function upstreamFields(doc, key, field) {
  const row = doc.upstreams?.[key] || {};
  const id = identifier(field, "上游标识", key);
  const service = field("服务发现名称", row.service);
  const nodes = field(
    "静态节点（每行填写地址和权重，与服务发现二选一）",
    row.nodes?.map((n) => n.url + " " + n.weight).join("\n"),
    { multiline: true, placeholder: "http://127.0.0.1:8080 1" },
  );
  const path = field("健康检查路径", row.health_path || "/healthz", {
    required: true,
  });
  const values = advancedNumbers.map(([name, label, initial]) =>
    field(label, row[name] ?? initial, {
      required: true,
      type: "number",
      group: "高级设置",
    }),
  );
  return () => {
    if (!key && Object.hasOwn(doc.upstreams || {}, id.value))
      throw Error("上游标识已存在");
    const name = service.value.trim(),
      list = nodes.value.trim();
    if (!!name === !!list)
      throw Error("请填写服务发现名称或静态节点，两者只能选择一项");
    const value = { health_path: path.value.trim() };
    if (name) value.service = name;
    else
      value.nodes = list
        .split(/\n/)
        .filter((line) => line.trim())
        .map((line) => {
          const [url, weight = "1", extra] = line.trim().split(/\s+/);
          if (extra) throw Error("每行填写一个节点地址和权重");
          return { url, weight: integer(weight, 1, 1000) };
        });
    advancedNumbers.forEach(([name, , , min, max], i) => {
      value[name] = integer(values[i].value, min, max);
    });
    doc.upstreams = { ...doc.upstreams, [id.value]: value };
  };
}
function policyFields(doc, key, global, field) {
  const id = global ? null : identifier(field, "策略标识", key);
  const config = { ...(global ? doc.plugins : doc.profiles?.[key]) };
  const mode = (label, name) =>
    field(
      label,
      config[name] === null ? "停用" : config[name] ? "启用" : "继承",
      {
        choices: global ? ["停用", "启用", "继承"] : ["继承", "启用", "停用"],
        group: "请求处理",
        hint: global
          ? "继承表示使用网关默认行为。"
          : "继承沿用全局设置，停用会覆盖全局设置。",
      },
    );
  const tracing = mode("请求追踪 ID", "request-id");
  const limiting = mode("请求限速", "rate-limit");
  const originalLimits = config["rate-limit"];
  const rate = field("每秒请求数", originalLimits?.rate ?? 100, {
    required: true,
    type: "number",
    step: "any",
    group: "请求处理",
    show: () => limiting.value === "启用",
  });
  const burst = field("突发请求数", config["rate-limit"]?.burst ?? 100, {
    required: true,
    type: "number",
    group: "请求处理",
    show: () => limiting.value === "启用",
  });
  delete config["request-id"];
  delete config["rate-limit"];
  const other = plugins(field, config);
  return () => {
    const value = jsonObject(other);
    if ("rate-limit" in value || "request-id" in value)
      throw Error("请在「请求处理」中配置请求限速和追踪");
    if (tracing.value !== "继承")
      value["request-id"] = tracing.value === "启用" ? {} : null;
    if (limiting.value === "停用") value["rate-limit"] = null;
    if (limiting.value === "启用") {
      const requests = Number(rate.value);
      if (!Number.isFinite(requests) || requests <= 0)
        throw Error("每秒请求数必须大于零");
      value["rate-limit"] = {
        ...originalLimits,
        rate: requests,
        burst: integer(burst.value, 1, 1000000),
      };
    }
    if (global) {
      doc.plugins = value;
      return;
    }
    if (!key && Object.hasOwn(doc.profiles || {}, id.value))
      throw Error("策略标识已存在");
    doc.profiles = { ...doc.profiles, [id.value]: value };
  };
}
