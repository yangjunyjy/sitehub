export const columns = [
  { key: "label", label: "模式名称", minWidth: 190 },
  { key: "id", prop: "id", label: "模式标识", minWidth: 170 },
  { key: "description", prop: "description", label: "说明", minWidth: 220 },
  { key: "status", label: "展示状态", width: 110 },
  { key: "count", label: "所属内容", width: 110 },
  { key: "sortOrder", prop: "sortOrder", label: "顺序", width: 80 },
  { key: "actions", label: "操作", width: 212, fixed: "right", utility: true },
];
