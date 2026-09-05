import assert from "node:assert/strict";
import { execFile as execFileCallback } from "node:child_process";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test, { after, before } from "node:test";
import { promisify } from "node:util";

import { repositoryRoot } from "../scripts/skill-registry.mjs";
import { compileSchema } from "./support/schema-validator.mjs";

const schemaNames = [
  "executor-request.schema.json", "executor-receipt.schema.json", "task-record.schema.json",
  "model-query.schema.json", "model-query-receipt.schema.json", "balance-snapshot.schema.json",
  "balance-receipt.schema.json", "init-receipt.schema.json", "doctor-receipt.schema.json"
];
const documents = await Promise.all(schemaNames.map(async (name) =>
  JSON.parse(await readFile(path.join(repositoryRoot, "schemas", name), "utf8"))));
const validators = new Map(schemaNames.map((name, index) => [name, compileSchema(documents[index], documents)]));
const execFile = promisify(execFileCallback);
let temporary;
let examples;

before(async () => {
  temporary = await mkdtemp(path.join(os.tmpdir(), "puretokens-schema-fixtures-"));
  const output = path.join(temporary, "examples.json");
  try {
    await execFile("go", ["test", "-run", "^TestContractExamples$", "-count=1", "."], {
      cwd: path.join(repositoryRoot, "runtime", "executor"),
      env: { ...process.env, PURETOKENS_SCHEMA_EXAMPLES_OUT: output },
      timeout: 120_000,
      maxBuffer: 2 * 1024 * 1024
    });
  } catch (error) {
    throw new Error(`Native contract examples failed:\n${error.stdout ?? ""}\n${error.stderr ?? ""}`, { cause: error });
  }
  examples = JSON.parse(await readFile(output, "utf8"));
});
after(async () => {
  if (temporary) await rm(temporary, { recursive: true, force: true });
});

function example(name) {
  const found = examples.find((entry) => entry.name === name);
  assert.ok(found, `native fixture missing: ${name}`);
  return structuredClone(found.document);
}

function rejects(schema, document, keyword) {
  const errors = validators.get(schema)(document);
  assert.ok(errors.length > 0, `${schema} accepted an invalid document`);
  if (keyword) assert.ok(errors.some((error) => error.includes(keyword)), errors.join("\n"));
}

test("actual executor JSON and executable documentation conform to their schemas", () => {
  const coveredSchemas = new Set();
  const names = new Set();
  for (const entry of examples) {
    assert.ok(!names.has(entry.name), `duplicate example name: ${entry.name}`);
    names.add(entry.name);
    const validator = validators.get(entry.schema);
    assert.ok(validator, `unexpected schema: ${entry.schema}`);
    assert.deepEqual(validator(entry.document), [], entry.name);
    coveredSchemas.add(entry.schema);
  }
  assert.deepEqual([...coveredSchemas].sort(), [...schemaNames].sort());
  assert.equal(examples.filter((entry) => /^docs-(image|video)-\d$/.test(entry.name)).length, 8);
});

test("media receipt schemas reject lost task context and unsafe metadata", () => {
  const receipt = example("record-0-submit-receipt");
  for (const field of ["kind", "operation", "task_id", "status"]) {
    const changed = structuredClone(receipt);
    delete changed[field];
    rejects("executor-receipt.schema.json", changed, "required");
  }
  for (const addition of [
    { raw_response_body: "private upstream data" }, { prompt: "private prompt" },
    { parameters: { image_urls: ["https://reference.example.test/private.png"] } },
    { requested_count: 7 }, { http_status: 700 }, { submission_outcome: "assumed_success" }
  ]) rejects("executor-receipt.schema.json", { ...receipt, ...addition });
  const failure = example("submission-400");
  assert.equal(failure.submission_outcome, "rejected");
  assert.equal(example("submission-503").submission_outcome, "unknown");
  delete failure.next_action;
  rejects("executor-receipt.schema.json", failure, "required next_action");
  const retry = example("status-retry");
  rejects("executor-receipt.schema.json", { ...retry, retry_not_before: "2026-02-30T12:00:00Z" }, "format");
  rejects("executor-receipt.schema.json", { ...retry, retry_not_before: "tomorrow" }, "format");
});

test("task records reject credentials, raw inputs and invalid progress", () => {
  const record = example("record-5-delivered-artifact");
  assert.deepEqual(record.delivered, [0, 1]);
  assert.equal(record.original_operation, "generate");
  for (const addition of [
    { prompt: "private prompt" }, { api_key: "synthetic-private-value" }, { reference_url: "https://reference.example.test" },
    { parameters: { generate_audio: "true" } }, { parameters: { n: 0 } },
    { downloaded: { 6: "/output/out.png" } }, { delivered: [0, 0] }, { delivered: [6] },
    { retry_not_before: "later" }, { format: "puretokens-task-v2" }
  ]) rejects("task-record.schema.json", { ...record, ...addition });
  for (const field of ["format", "kind"]) {
    const changed = structuredClone(record);
    delete changed[field];
    rejects("task-record.schema.json", changed, "required");
  }
});

test("raw quotas and legacy billing responses cannot masquerade as a balance projection", () => {
  const balance = example("balance-projection");
  assert.equal(balance.used, 19.672244);
  assert.equal(balance.remaining, 10);
  assert.equal(balance.unit, "USD");
  assert.equal(balance.scope, "account_wallet");
  for (const raw of [{}, { hard_limit_usd: 100000000 }, { total_usage: 275 }, { total_available: 5000000, total_used: 9836122, total_granted: 14836122, unlimited_quota: true }]) {
    rejects("balance-snapshot.schema.json", raw, "required");
  }
  for (const field of Object.keys(balance)) {
    const changed = structuredClone(balance);
    delete changed[field];
    rejects("balance-snapshot.schema.json", changed, `required ${field}`);
  }
  for (const addition of [{ unit: "CNY" }, { scope: "unlimited" }, { currency: "USD" }, { used: "2.75" }, { includes_subscription_quota: true }, { remaining: "infinity" }]) {
    rejects("balance-snapshot.schema.json", { ...balance, ...addition });
  }
  const envelope = example("balance");
  rejects("balance-receipt.schema.json", { ...envelope, result: { hard_limit_usd: 12.5 } });
});

test("console-origin balance exception does not change media transport or allow legacy billing", async () => {
  const read = async (relative) => JSON.parse(await readFile(path.join(repositoryRoot, relative), "utf8"));
  const schema = await read("schemas/media-execution-contract.schema.json");
  const validate = compileSchema(schema, [schema]);
  const balance = await read("skills/puretokens-balance/references/execution-contract.json");
  assert.deepEqual(validate(balance), []);
  for (const patch of [
    { url: "https://api.puretokensx.com/v1/dashboard/billing/subscription" },
    { url: "https://another-provider.invalid/api/usage/token/" },
    { requiresBrowserSession: true }, { unitMetadataRequiresConfiguredApiKey: true },
    { unitMetadataUrl: "https://another-provider.invalid/status" },
    { maxRequests: 3 }
  ]) {
    const changed = structuredClone(balance);
    Object.assign(changed.operations.read, patch);
    assert.ok(validate(changed).length > 0, JSON.stringify(patch));
  }
  for (const kind of ["image", "video", "models", "connection"]) {
    const contract = await read(`skills/puretokens-${kind}/references/execution-contract.json`);
    assert.deepEqual(validate(contract), [], kind);
    contract.transport.fixedApiOrigin = "https://console.puretokensx.com";
    assert.ok(validate(contract).length > 0, `${kind} accepted the balance-only origin`);
  }
  const directSchema = await read("schemas/direct-api-execution-contract.schema.json");
  const validateDirect = compileSchema(directSchema, [directSchema]);
  const direct = await read("references/direct-api-execution-contract.json");
  assert.deepEqual(validateDirect(direct), []);
  direct.balance.unit = "CNY";
  assert.ok(validateDirect(direct).length > 0);
});

test("model query receipts keep declared matching separate from raw catalog fields", () => {
  const receipt = example("models-true");
  assert.equal(receipt.result.matched_count, receipt.result.data.length);
  assert.deepEqual(receipt.result.filter.parameter_names, ["resolution"]);
  assert.deepEqual(receipt.result.data.map((entry) => entry.id), ["video-audio"]);
  for (const field of ["filter", "matched_count", "matching_scope", "note"]) {
    const changed = structuredClone(receipt);
    delete changed.result[field];
    rejects("model-query-receipt.schema.json", changed, `required ${field}`);
  }
  const leaked = structuredClone(receipt);
  leaked.result.data[0].provider = "private-upstream";
  rejects("model-query-receipt.schema.json", leaked, "false schema");
  const echoed = structuredClone(receipt);
  echoed.result.filter.parameters = { resolution: "720p" };
  rejects("model-query-receipt.schema.json", echoed, "false schema");
  for (const query of [{ kind: "chat" }, { parameters: null }, { model: "model/alias" }, { prompt: "generate this" }]) {
    rejects("model-query.schema.json", query);
  }
});

test("doctor schema requires separate local checks and structured connection evidence", () => {
  const doctor = example("doctor-success");
  assert.equal(doctor.local.checksum_status, "unverified");
  assert.equal(doctor.local.attachment_delivery_status, "unverified");
  assert.equal(doctor.connection.credential_verified, true);
  assert.equal(example("doctor-no-credential").connection.api_request_executed, false);
  for (const field of ["installations", "duplicate_skills", "checksum_status", "attachment_delivery_status"]) {
    const changed = structuredClone(doctor);
    delete changed.local[field];
    rejects("doctor-receipt.schema.json", changed, `required ${field}`);
  }
  const missingConnectionEvidence = structuredClone(doctor);
  delete missingConnectionEvidence.connection.credential_verified;
  rejects("doctor-receipt.schema.json", missingConnectionEvidence, "required credential_verified");
  const disclosed = structuredClone(doctor);
  disclosed.connection.base_url = "https://private.example.test";
  rejects("doctor-receipt.schema.json", disclosed, "false schema");
  const wrongType = structuredClone(doctor);
  wrongType.local.installations[0].loaded = "true";
  rejects("doctor-receipt.schema.json", wrongType, "type");
});

test("request schemas reject missing submission or continuation identity", () => {
  for (const request of [
    { kind: "image" },
    { kind: "image", operation: "generate", model: "gpt-image-2" },
    { kind: "image", operation: "generate", model: "gpt-image-2", prompt: "  " },
    { kind: "video", task_id: "../wrong-task" },
    { kind: "image", task_id: "fixture", poll: { max_status_reads: 100 } },
    { kind: "image", task_id: "fixture", lifecycle: { status_url: "/other" } }
  ]) rejects("executor-request.schema.json", request);
  const submission = example("docs-image-0");
  for (const addition of [
    { task_id: "existing" }, { task_status: "completed" }, { original_operation: "edit" },
    { retry_not_before: "2026-09-05T00:00:00Z" }, { reconciliation_required: false }, { index: 1 }
  ]) rejects("executor-request.schema.json", { ...submission, ...addition });
  assert.deepEqual(validators.get("executor-request.schema.json")({ kind: "image", task_id: "fixture", requested_count: 0 }), []);
});

test("schema evaluator fails closed and exercises branches, references and JSON equality", () => {
  const id = "https://fixture.example.test/schema.json";
  const compile = (schema) => {
    const document = { $id: id, ...schema };
    return compileSchema(document, [document]);
  };
  assert.throws(() => compile({ anyOf: [{ type: "string" }, { unsupportedKeyword: true }] }), /Unsupported schema keyword/);
  assert.throws(() => compile({ format: "invented-format" }), /Unsupported schema format/);
  assert.throws(() => compile({ $ref: "missing.json" }), /Unregistered schema reference/);
  assert.throws(() => compile({ $ref: "#/$defs/missing" }), /Unresolved schema reference/);
  const nonempty = compile({ type: "object", minProperties: 1 });
  assert.deepEqual(nonempty({ value: 0 }), []);
  assert.ok(nonempty({}).some((error) => error.includes("minProperties")));
  const branches = compile({
    $defs: { count: { type: "integer", minimum: 1, maximum: 6 } },
    type: "object",
    properties: { count: { $ref: "#/$defs/count" }, values: { type: "array", uniqueItems: true, items: { type: "string" } } },
    additionalProperties: false,
    oneOf: [{ required: ["count"] }, { required: ["values"] }]
  });
  assert.deepEqual(branches({ count: 2 }), []);
  for (const value of [{}, { count: 0 }, { count: "2" }, { count: 2, values: [] }, { values: ["a", "a"] }, { values: [1] }, { count: 2, extra: true }]) {
    assert.ok(branches(value).length > 0);
  }
  const equality = compile({ enum: [{ a: 1, b: 2 }] });
  assert.deepEqual(equality({ b: 2, a: 1 }), []);
  const conditional = compile({ if: { type: "integer" }, then: { minimum: 2 }, else: { const: "unknown" }, not: { const: 3 } });
  assert.deepEqual(conditional(2), []);
  assert.deepEqual(conditional("unknown"), []);
  for (const value of [1, 3, "other"]) assert.ok(conditional(value).length > 0);
});
