import assert from "node:assert/strict";
import test from "node:test";
import { compileSchema } from "./support/schema-validator.mjs";

test("bounded evaluation object counts are enforced by the schema evaluator", () => {
  const schema = {$id: "https://example.test/bounded", type: "object", maxProperties: 2};
  const validate = compileSchema(schema, [schema]);
  assert.deepEqual(validate({a: 1, b: 2}), []);
  assert.ok(validate({a: 1, b: 2, c: 3}).some(error => error.includes("maxProperties")));
});
