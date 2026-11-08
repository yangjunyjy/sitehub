import test from "node:test";
import assert from "node:assert/strict";
import {
  createContentResource,
  emptyWebsiteContent,
  normalizeWebsiteContent,
} from "../src/utils/content-resource.js";

function resource() {
  const requests = [];
  const state = {
    loading: false,
    loaded: false,
    error: "",
    content: emptyWebsiteContent(),
  };
  const load = createContentResource(
    state,
    ({ signal }) =>
      new Promise((resolve, reject) => {
        requests.push({ signal, resolve, reject });
      }),
    normalizeWebsiteContent,
  );
  return { state, load, requests };
}

test("consumers share the request and then use the accepted content snapshot", async () => {
  const { state, load, requests } = resource();
  const first = load();
  assert.equal(load(), first);
  await Promise.resolve();
  assert.equal(requests.length, 1);
  requests[0].resolve({ data: { brand: { name: "Test" } } });
  await first;
  assert.equal(state.content.brand.name, "Test");
  assert.equal((await load()).brand.name, "Test");
  assert.equal(requests.length, 1);
});

test("a superseded success cannot replace new content or finish its loading state", async () => {
  const { state, load, requests } = resource();
  const old = load();
  await Promise.resolve();
  const latest = load(true);
  await Promise.resolve();
  assert.equal(requests[0].signal.aborted, true);
  requests[0].resolve({ data: { brand: { name: "Old" } } });
  await old;
  assert.equal(state.loading, true);
  assert.equal(load(), latest);
  requests[1].resolve({ data: { brand: { name: "New" } } });
  await latest;
  assert.equal(state.content.brand.name, "New");
  assert.equal(state.loading, false);
});

test("a stale failure does not erase a newer success", async () => {
  const { state, load, requests } = resource();
  const old = load();
  await Promise.resolve();
  const latest = load(true);
  await Promise.resolve();
  requests[1].resolve({ data: { brand: { name: "New" } } });
  await latest;
  requests[0].reject(new Error("old failure"));
  await old;
  assert.equal(state.content.brand.name, "New");
  assert.equal(state.error, "");
});

test("an invalid refresh retains the last good snapshot and remains retryable", async () => {
  const { state, load, requests } = resource();
  const first = load();
  await Promise.resolve();
  requests[0].resolve({ data: { home: { title: "Saved content" } } });
  await first;
  const broken = load(true);
  await Promise.resolve();
  requests[1].resolve({ data: null });
  await assert.rejects(broken, /格式异常/);
  assert.equal(state.loaded, true);
  assert.equal(state.content.home.title, "Saved content");
  const retry = load(true);
  await Promise.resolve();
  requests[2].resolve({ data: { home: { title: "Recovered" } } });
  await retry;
  assert.equal(state.content.home.title, "Recovered");
  assert.equal(state.error, "");
});
