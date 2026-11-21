import test from "node:test";
import assert from "node:assert/strict";
import { readdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { relative, join } from "node:path";
import { parse } from "vue/compiler-sfc";
import { createMemoryHistory } from "vue-router";
import {
  componentKey,
  filterMenuTree,
  navigationMenus,
} from "../src/router/pages.js";
import { createAdminRouter } from "../src/router/runtime.js";

const views = fileURLToPath(new URL("../src/views/", import.meta.url));
const modules = {};
const settings = {};
function discover(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const file = join(directory, entry.name);
    if (entry.isDirectory()) discover(file);
    else if (entry.name === "menu.json") {
      settings['../views/' + relative(views, file).replaceAll('\\','/')] = JSON.parse(readFileSync(file, 'utf8'));
    } else if (entry.name === "index.vue") {
      const path = relative(views, file).replaceAll("\\", "/");
      const source = readFileSync(file, "utf8");
      const { descriptor, errors } = parse(source, { filename: file });
      assert.deepEqual(errors, [], path);
      if (!descriptor.template) {
        const shared = descriptor.script?.content.match(/export \{ default \} from ['"](.+\.vue)['"]/);
        assert(shared, path);
        assert(parse(readFileSync(join(directory, shared[1]), "utf8")).descriptor.template, path);
      }
      modules[`../views/${path}`] = { name: path, template: "<div />" };
    }
  }
}
discover(views);

function menu(id, path, component, extra = {}) {
  return {
    id,
    path,
    component,
    menu_name: `页面 ${id}`,
    menu_type: 2,
    status: 1,
    is_visible: true,
    ...extra,
  };
}
function setup(menus = []) {
  const values = new Map([["token", "session"]]);
  const storage = {
    getItem: (key) => values.get(key),
    removeItem: (key) => values.delete(key),
  };
  const rbac = {
    menusLoaded: false,
    currentMenus: [],
    currentUser: { roleIds: [2] },
    permissions: [],
    async loadAllDataFromBackend() {
      this.currentMenus = menus;
      this.menusLoaded = true;
    },
    hasPermission(permission) {
      return this.permissions.includes(permission);
    },
    resetAllData() {
      this.currentMenus = [];
      this.menusLoaded = false;
    },
  };
  const router = createAdminRouter({
    history: createMemoryHistory(),
    modules, settings,
    layout: { template: "<router-view />" },
    store: () => rbac,
    storage,
  });
  return { router, rbac, values };
}

test("page discovery uses actual index.vue files and exact component paths", () => {
  assert(Object.keys(modules).length >= 20);
  for (const path of [
    "system/users/index",
    "blog/articles/create/index",
    "blog/articles/editor/index",
    "product/docs/groups/index",
    "product/download-settings/index",
  ])
    assert(modules[componentKey(path)], path);
  for (const old of [
    "views/users/UserList",
    "system/users/index.vue",
    "/src/views/system/users/index",
    "../login/index",
  ])
    assert.equal(componentKey(old), "");
});

test("first refresh waits for menus, installs them under layout and resolves the original URL", async () => {
  const { router, rbac } = setup([
    menu(7, "/custom/people", "system/users/index"),
  ]);
  let finish;
  rbac.loadAllDataFromBackend = () =>
    new Promise((resolve) => {
      finish = () => {
        rbac.currentMenus = [menu(7, "/custom/people", "system/users/index")];
        rbac.menusLoaded = true;
        resolve();
      };
    });
  const navigation = router.push("/custom/people?search=alice");
  while (!finish) await new Promise((resolve) => setImmediate(resolve));
  assert.equal(router.resolve("/custom/people").name, "NotFound");
  finish();
  await navigation;
  assert.equal(router.currentRoute.value.name, "menu-7");
  assert.equal(
    router.currentRoute.value.fullPath,
    "/custom/people?search=alice",
  );
  assert.equal(router.currentRoute.value.matched[0].name, "dashboard");
  assert.equal(
    router.currentRoute.value.matched.at(-1).components.default,
    modules[componentKey("system/users/index")],
  );
});

test("unknown components and unknown URLs show 404 inside layout without aliasing", async () => {
  const { router } = setup([menu(9, "/broken", "views/users/UserList")]);
  await router.push("/broken");
  assert.equal(router.currentRoute.value.meta.missingComponent, true);
  assert.equal(
    router.currentRoute.value.matched.at(-1).components.default,
    modules[componentKey("not-found/index")],
  );
  await router.push("/no-such-page");
  assert.equal(router.currentRoute.value.name, "NotFound");
  assert.equal(router.currentRoute.value.matched[0].name, "dashboard");
});

test("menu edits replace routes; revocation returns 403; logout removes all menu routes", async () => {
  const { router, rbac, values } = setup([
    menu(3, "/people", "system/users/index", {
      permission: "system:user:list",
    }),
  ]);
  await router.push("/people");
  assert.equal(router.currentRoute.value.name, "Forbidden");
  assert.equal(router.currentRoute.value.matched[0].name, "dashboard");
  rbac.permissions = ["system:user:list"];
  await router.push("/people");
  assert.equal(router.currentRoute.value.name, "menu-3");
  rbac.currentMenus = [menu(3, "/renamed", "system/users/index")];
  await router.push("/renamed");
  assert.equal(router.currentRoute.value.name, "menu-3");
  assert.equal(router.resolve("/people").name, "NotFound");
  values.delete("token");
  await router.push("/renamed?again=1");
  assert.equal(router.currentRoute.value.name, "Login");
  assert.equal(router.hasRoute("menu-3"), false);
});

test("a temporary menu failure can retry the original page in the same session", async () => {
  const { router, rbac, values } = setup();
  rbac.loadAllDataFromBackend = async () => {
    throw new Error("offline");
  };
  await router.push("/dashboard/system/users");
  assert.equal(router.currentRoute.value.name, "AccessUnavailable");
  assert.equal(values.has("token"), true);
  assert.equal(router.hasRoute("menu-3"), false);
  rbac.loadAllDataFromBackend = async () => {
    rbac.currentMenus = [menu(3, '/dashboard/system/users', 'system/users/index')];
    rbac.menusLoaded = true;
  };
  await router.replace(router.currentRoute.value.query.redirect);
  assert.equal(router.currentRoute.value.name, 'menu-3');
});

test('expired credentials return to login with the requested destination', async () => {
  const { router, rbac, values } = setup();
  rbac.loadAllDataFromBackend = async () => { throw Object.assign(new Error('expired'), { response: { status: 401 } }); };
  await router.push('/dashboard/system/users');
  assert.equal(router.currentRoute.value.name, "Login");
  assert.equal(router.currentRoute.value.query.redirect, '/dashboard/system/users');
  assert.equal(values.has("token"), false);
  assert.equal(router.hasRoute("menu-3"), false);
});

test("workspace links follow renamed home, profile and parametrized editor records", async () => {
  const { router } = setup([
    menu(1, "/studio", "dashboard/index"),
    menu(2, "/account", "profile/index"),
    menu(3, "/stories/:id/edit", "blog/articles/editor/index", { is_visible: false }),
  ]);
  await router.push("/");
  assert.equal(router.currentRoute.value.path, "/studio");
  await router.push(router.page("profile/index"));
  assert.equal(router.currentRoute.value.path, "/account");
  await router.push(router.page("blog/articles/editor/index", {
    params: { id: 7 }, query: { from: "drafts" },
  }));
  assert.equal(router.currentRoute.value.fullPath, "/stories/7/edit?from=drafts");
  await router.push(router.page("system/users/index"));
  assert.equal(router.currentRoute.value.name, "Forbidden");
});

test("navigation preserves only backend hierarchy, paths, order and visibility", () => {
  const source = [
    {
      id: 1,
      menu_name: "自定义目录",
      menu_type: 1,
      path: "/custom",
      is_visible: true,
      children: [
        menu(2, "people", "system/users/index", {
          menu_name: "自定义用户",
          sort: 2,
        }),
        menu(3, "/downloads", "product/downloads/index", { sort: 1 }),
        menu(4, "/editor", "blog/articles/create/index", { is_visible: false }),
        menu(5, "/email", "admin/settings/email/index"),
        { id: 6, menu_type: 3, permission: "system:user:create", status: 1 },
      ],
    },
  ];
  const before = structuredClone(source);
  assert.deepEqual(
    navigationMenus(source, false, settings)[0].children.map((item) => item.path),
    ["/downloads", "/custom/people"],
  );
  assert.equal(navigationMenus(source, true, settings)[0].children.length, 3);
  assert.equal(filterMenuTree(source)[0].children.length, 5);
  assert.deepEqual(source, before);
  assert.deepEqual(navigationMenus([]), []);
});
