export const columns = [
  {
    key: "selection",
    type: "selection",
    width: "48",
    align: "center",
    label: "选择文章",
    locked: true,
    utility: true,
  },
  {
    key: "content",
    label: "文章内容",
    minWidth: "310",
  },
  {
    key: "taxonomy",
    label: "分类 / 标签",
    minWidth: "150",
  },
  {
    key: "status",
    label: "状态",
    width: "105",
  },
  {
    key: "views",
    label: "访问",
    width: "80",
    align: "center",
  },
  {
    key: "updatedAt",
    label: "更新时间",
    width: "130",
  },
  {
    key: "isTop",
    label: "置顶",
    width: "86",
    align: "center",
  },
  {
    key: "actions",
    label: "操作",
    width: "184",
    fixed: "right",
    utility: true,
  },
];
