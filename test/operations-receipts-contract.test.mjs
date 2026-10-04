import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { compileSchema } from "./support/schema-validator.mjs";

test("anonymous receipts remain optional and separate from inference/authentication transport", async () => {
  const contract = JSON.parse(await readFile(new URL("../references/direct-api-execution-contract.json", import.meta.url), "utf8"));
  const schema = JSON.parse(await readFile(new URL("../schemas/direct-api-execution-contract.schema.json", import.meta.url), "utf8"));
  const validate = compileSchema(schema, [schema]);
  assert.deepEqual(validate(contract), []);
  for (const change of [
    { enabledByDefault: true }, { authentication: "bearer" }, { readsHostConfiguration: true },
    { persistentIdentifier: true }, { retries: 1 }, { durableQueue: true }, { changesBusinessResult: true },
    { fields: [...contract.operationsReceipts.fields, "accountId"] },
    { url: "https://api.puretokensx.com/v1" }, { asyncSuccessMeaning: "output_delivered" }
  ]) {
    const altered = structuredClone(contract);
    Object.assign(altered.operationsReceipts, change);
    assert.ok(validate(altered).length, "unsafe receipt drift must fail the schema contract");
  }
  const missing = structuredClone(contract);
  delete missing.operationsReceipts;
  assert.ok(validate(missing).length);
});
