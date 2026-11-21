// 通用构建菜单树函数（用于表格展示）
export const buildMenuTree = (parentId, list) => {
  if (!list || !Array.isArray(list)) {
    return [];
  }
  return list
    .filter((m) => String(m.parent_id) === String(parentId))
    .sort((a, b) => a.sort - b.sort)
    .map((m) => {
      const children = buildMenuTree(m.id, list);
      return {
        ...m,
        hasChildren: children.length > 0,
        children,
      };
    });
};

// 递归搜索树：检查节点及其子节点是否匹配搜索条件
export const matchTree = (node, searchTitle, searchType) => {
  const matchTitle =
    !searchTitle || (node.menu_name && node.menu_name.includes(searchTitle));
  const matchType = !searchType || node.menu_type === searchType;
  const isMatch = matchTitle && matchType;

  if (node.children && node.children.length > 0) {
    // 递归过滤子节点
    const matchedChildren = node.children
      .map((child) => matchTree(child, searchTitle, searchType))
      .filter(Boolean);
    if (matchedChildren.length > 0) {
      return { ...node, children: matchedChildren, hasChildren: true };
    }
  }
  return isMatch
    ? {
        ...node,
        children: node.children || [],
        hasChildren: node.children && node.children.length > 0,
      }
    : null;
};

export const getTypeLabel = (type) => {
  const map = {
    directory: "目录",
    menu: "菜单",
    button: "按钮",
    1: "目录",
    2: "菜单",
    3: "按钮",
  };
  return map[type] || type;
};

export const getTypeTagType = (type) => {
  const map = {
    directory: "warning",
    menu: "primary",
    button: "success",
    1: "warning",
    2: "primary",
    3: "success",
  };
  return map[type] || "info";
};
