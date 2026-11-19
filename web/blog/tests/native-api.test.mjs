import test from "node:test";
import assert from "node:assert/strict";
import {
  uint64,
  stringifyJSON,
  queryString,
  viewValues,
  allPages,
} from "../src/utils/native-api.mjs";

test("native requests preserve uint64 precision and collection query fields", () => {
  assert.equal(
    stringifyJSON({ id: uint64("18446744073709551615"), enabled: false }),
    '{"id":18446744073709551615,"enabled":false}',
  );
  assert.throws(() => uint64(Number.MAX_SAFE_INTEGER + 1));
  assert.throws(() => uint64("-1"));
  assert.throws(() => uint64("18446744073709551616"));
  assert.equal(
    queryString({ tagIDs: ["1", "2"], status: 0, trash: false }),
    "tagIDs=1&tagIDs=2&status=0&trash=false",
  );
  assert.deepEqual(
    viewValues({
      id: "18446744073709551615",
      version: "3",
      total: "4",
      status: "0",
    }),
    { id: "18446744073709551615", version: "3", total: 4, status: 0 },
  );
});

test("protobuf captcha geometry becomes bounded UI numbers without changing identifiers or images", () => {
  const response = {
    captchaKey: "captcha-key",
    backgroundImage: "data:image/png;base64,background",
    sliderImage: "data:image/png;base64,piece",
    y: "19",
    imgWidth: "300",
    imgHeight: "150",
    pieceSize: "44",
    piecePadding: "4",
    id: "18446744073709551615",
    version: "3",
    revision: "7",
  };
  const data = viewValues(response);
  for (const [field, expected] of Object.entries({
    y: 19,
    imgWidth: 300,
    imgHeight: 150,
    pieceSize: 44,
    piecePadding: 4,
  })) {
    assert.equal(data[field], expected, field);
  }
  assert.equal(data.y + data.pieceSize, 63);
  assert.ok(data.y + data.pieceSize <= data.imgHeight);
  assert.equal(data.pieceSize + data.piecePadding * 2, 52);
  for (const field of [
    "captchaKey",
    "backgroundImage",
    "sliderImage",
    "id",
    "version",
    "revision",
  ]) {
    assert.equal(data[field], response[field], field);
  }
  assert.equal(
    response.pieceSize,
    "44",
    "normalization must not mutate the API response",
  );
});

test("captcha geometry conversion rejects values outside safe unsigned integer strings", () => {
  for (const value of ["9007199254740992", "-1", "1.5", "NaN", ""]) {
    const geometry = Object.fromEntries(
      ["imgWidth", "imgHeight", "pieceSize", "piecePadding"].map((field) => [
        field,
        value,
      ]),
    );
    assert.deepEqual(viewValues(geometry), geometry);
  }
});

test("catalog aggregation requests all pages and propagates failures", async () => {
  const result = await allPages(async ({ page }) => ({
    code: 0,
    data: { items: [page], total: "3" },
  }));
  assert.deepEqual(result.data, [1, 2, 3]);
  await assert.rejects(
    allPages(async () => {
      throw new Error("offline");
    }),
    /offline/,
  );
});
