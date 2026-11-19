import test from "node:test";
import assert from "node:assert/strict";
import {
  collectBlogIndex,
  archiveGroups,
  archiveDate,
  blogTopics,
} from "../src/utils/blog-index.js";

test("archive includes every page, ignores pin order, and keeps only published list metadata", async () => {
  const records = Array.from({ length: 56 }, (_, i) => ({
    id: i + 1,
    title: `Post ${i}`,
    published_at: i === 0 ? "2024-12-02T00:00:00Z" : "2026-11-03T00:00:00Z",
    category: { id: 1, name: "Go" },
    tags: [
      { id: 2, name: "Vue" },
      { id: 2, name: "Vue" },
    ],
    content: "must not be retained",
  }));
  const requests = [];
  const items = await collectBlogIndex(async ({ page, pageSize }) => {
    requests.push(page);
    return {
      data: {
        total: records.length,
        items: records.slice((page - 1) * pageSize, page * pageSize),
      },
    };
  });
  assert.deepEqual(requests, [1, 2]);
  assert.equal(items.length, 56);
  assert.equal(items.at(-1).id, "1");
  assert(!("content" in items[0]));
  const groups = archiveGroups(items);
  assert.deepEqual(
    groups.map((item) => [item.year, item.count]),
    [
      ["2026", 55],
      ["2024", 1],
    ],
  );
  assert.equal(blogTopics(items, "tags")[0].count, 56);
  assert.equal(blogTopics(items, "categories")[0].count, 56);
});

test("unknown and invalid dates are not assigned to the current year", () => {
  assert.equal(archiveDate("2025-02-29"), "");
  assert.equal(archiveDate("not-a-date"), "");
  assert.equal(archiveDate("2024-02-29T23:00:00-08:00"), "2024-02-29");
  assert.equal(archiveDate(null), "");
  assert.equal(archiveGroups([{ date: "", id: "1" }])[0].year, "undated");
});

test("incomplete, changing and repeated pages reject instead of showing false archive totals", async () => {
  await assert.rejects(
    collectBlogIndex(async () => ({ data: { total: 1, items: [] } })),
  );
  await assert.rejects(
    collectBlogIndex(async () => ({
      data: { total: 2, items: [{ id: 1, title: "Duplicate" }] },
    })),
  );
  await assert.rejects(
    collectBlogIndex(async ({ page }) => ({
      data: {
        total: page === 1 ? 2 : 3,
        items: [{ id: page, title: "Changed" }],
      },
    })),
  );
});

test("cancelled archive request never starts another page", async () => {
  const controller = new AbortController();
  let requests = 0;
  await assert.rejects(
    collectBlogIndex(
      async () => {
        requests++;
        controller.abort();
        return { data: { total: 2, items: [{ id: 1, title: "Cancelled" }] } };
      },
      { signal: controller.signal },
    ),
    { name: "AbortError" },
  );
  assert.equal(requests, 1);
});
