export const columns = [
  { key: "code", prop: "code", label: "系统代码", minWidth: 130 },
  { key: "name", prop: "name", label: "系统名称", minWidth: 150 },
  { key: "origin", prop: "origin", label: "访问地址", minWidth: 240 },
  { key: "description", prop: "description", label: "描述", minWidth: 200 },
  { key: "enabled", label: "状态", width: 90 },
  { key: "public", label: "可见性", width: 90 },
  { key: "actions", label: "操作", width: 132, fixed: "right", utility: true },
];
