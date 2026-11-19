import test from "node:test";
import assert from "node:assert/strict";
import { systemPayload } from "../src/utils/systems.js";
test("system edits cannot change persisted codes or submit credential URLs", () => {
  const form = {
    id: "18446744073709551615",
    code: "blog",
    name: " 博客 ",
    origin: "https://blog.example.com/",
    enabled: true,
    public: true,
  };
  assert.deepEqual(systemPayload(form, "blog"), {
    ...form,
    name: "博客",
    origin: "https://blog.example.com",
    description: "",
  });
  assert.throws(
    () => systemPayload({ ...form, code: "product" }, "blog"),
    /不能修改/,
  );
  for (const origin of [
    "javascript:alert(1)",
    "https://user:secret@example.com",
    "https://blog.example.com/path",
  ])
    assert.throws(() => systemPayload({ ...form, origin }));
});
