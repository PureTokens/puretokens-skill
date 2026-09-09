import assert from "node:assert/strict";
import { test } from "node:test";
import { acceptanceCases, validateHostAcceptance } from "../scripts/validate-host-acceptance.mjs";
const version = "0.17.1";
const sha = "a".repeat(64);
const manifest = { artifacts: { "darwin-arm64": { sha256: sha } } };
function fixture() {
  const checks = () => ({ installation: "pending", authenticatedAPI: "pending", sameTaskResume: "pending", nativeAttachmentDelivery: "pending" });
  return { schemaVersion: 1, version, checkedAt: "2026-09-06", hosts: [{ host: "codex", macOS: checks(), windows: checks(), linux: checks() }], evidence: [] };
}
test("fixture success cannot mark real host checks passed", () => {
  const record = fixture();
  assert.deepEqual(validateHostAcceptance(record, version, manifest, ["codex"]), []);
  record.hosts[0].macOS.installation = "passed";
  assert.ok(validateHostAcceptance(record, version, manifest, ["codex"]).length);
});
test("host evidence must identify actual versions and release bytes", () => {
  const record = fixture();
  record.hosts[0].macOS.installation = "passed";
  const evidence = { method: "real-host", host: "codex", os: "macOS", hostVersion: "fixture-host-version", osVersion: "fixture-os", executorPlatform: "darwin-arm64", executorSha256: sha, version, checkedAt: "2026-09-06", artifact: "references/fixture.md", cases: Object.fromEntries(acceptanceCases.map(id => [id, id === "installation" ? "passed" : "pending"])) };
  record.evidence.push(evidence);
  assert.deepEqual(validateHostAcceptance(record, version, manifest, ["codex"]), []);
  for (const [key, value] of [["method", "fixture"], ["version", "0.16.0"], ["executorSha256", "b".repeat(64)], ["hostVersion", ""]]) {
    const changed = structuredClone(record);
    changed.evidence[0][key] = value;
    assert.ok(validateHostAcceptance(changed, version, manifest, ["codex"]).length, key);
  }
});

test("Linux evidence requires Linux bytes and cannot masquerade as a different OS", () => {
 const record = fixture(); const linuxManifest = { artifacts: { "linux-arm64": { sha256: sha } } };
 record.hosts[0].linux.installation = "passed";
 record.evidence.push({ method:"real-host",host:"codex",os:"linux",hostVersion:"fixture",osVersion:"fixture",executorPlatform:"linux-arm64",executorSha256:sha,version,checkedAt:"2026-09-08",artifact:"references/fixture.md",cases:Object.fromEntries(acceptanceCases.map(id => [id,id==="installation"?"passed":"pending"])) });
 assert.deepEqual(validateHostAcceptance(record,version,linuxManifest,["codex"]),[]);
 record.evidence[0].os="macOS";
 assert.ok(validateHostAcceptance(record,version,linuxManifest,["codex"]).length);
});
test("unavailable host OS combinations require an explanation", () => {
 const record=fixture(); record.hosts[0].linux.installation="unavailable";
 assert.ok(validateHostAcceptance(record,version,manifest,["codex"]).length);
 record.hosts[0].linux.reason="No verified client target for this release";
 assert.deepEqual(validateHostAcceptance(record,version,manifest,["codex"]),[]);
});
