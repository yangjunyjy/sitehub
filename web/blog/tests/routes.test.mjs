import test from "node:test";
import assert from "node:assert/strict";
import { pageRoutes } from "../src/router/routes.mjs";
import articlePath from "../src/views/article/route.mjs";
import notFoundPath from "../src/views/not-found/route.mjs";

test("ordinary directories create routes and local metadata declares dynamic URLs", () => {
  const component = () => Promise.resolve({});
  const routes = pageRoutes({
    "../views/categories/index.vue": component,
    "../views/article/index.vue": component,
    "../views/not-found/index.vue": component,
  }, {
    "../views/article/route.mjs": articlePath,
    "../views/not-found/route.mjs": notFoundPath,
  });
  assert.deepEqual(routes.map(({ path }) => path), ["/categories", "/article/:id", "/:pathMatch(.*)*"]);
  assert.deepEqual(
    pageRoutes({ "../views/home/components/card/index.vue": component }),
    [],
  );
});
