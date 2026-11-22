import {
  Connection,
  Switch,
  Operation,
  SetUp,
  Clock,
} from "@element-plus/icons-vue";

export const pages = [
  {
    id: "routes",
    name: "HTTP 路由",
    icon: Connection,
    group: "流量管理",
    description:
      "按主机、路径和请求方法匹配流量，转发到上游服务或直接返回响应。",
  },
  {
    id: "streams",
    name: "四层转发",
    icon: Switch,
    group: "流量管理",
    description: "在指定地址和端口接收 TCP 连接或 UDP 数据报，转发到目标端口。",
  },
  {
    id: "upstreams",
    name: "上游服务",
    icon: SetUp,
    group: "流量管理",
    description: "管理 HTTP 路由的转发目标，使用固定节点地址或服务发现。",
  },
  {
    id: "policies",
    name: "插件策略",
    icon: Operation,
    group: "流量管理",
    description:
      "HTTP 路由先应用全局设置，再应用策略模板，最后应用路由自己的设置。",
  },
  {
    id: "releases",
    name: "发布记录",
    icon: Clock,
    group: "变更管理",
    description:
      "查看配置修订与操作人。回滚当前发布会把上一版配置作为新修订发布。",
  },
];
export function currentPage() {
  return pages.find((p) => p.id === location.hash.slice(1)) || pages[0];
}
