import test from "node:test";
import assert from "node:assert/strict";
import { navigationLinks } from "../src/utils/navigation.mjs";

test('external navigation requires a public web address and no local component', () => {
  const node = { name: '友站', pageType: 'external', externalUrl: 'https://example.org/notes' };
  assert.deepEqual(navigationLinks([node, { ...node, externalUrl: 'javascript:alert(1)' }], []), [
    { label: node.name, path: node.externalUrl, external: true },
  ]);
});

test("public navigation uses real local page components and removes unavailable or hidden pages", () => {
  const pages = [
    { name: "home/index", path: "/" },
    { name: "categories/index", path: "/categories" },
    { name: "article/index", path: "/article/:id" },
  ];
  assert.deepEqual(
    navigationLinks(
      [
        { menuName: "首页", component: "home/index" },
        {
          menuName: "目录",
          children: [{ menuName: "分类", component: "categories/index" }],
        },
        { menuName: "重复", component: "home/index" },
        { menuName: "隐藏", component: "categories/index", isVisible: false },
        { menuName: "未知", component: "missing/index" },
        { menuName: "文章", component: "article/index" },
      ],
      pages,
    ),
    [
      { path: "/", label: "首页" },
      { path: "/categories", label: "分类" },
    ],
  );
});
