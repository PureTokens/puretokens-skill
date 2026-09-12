import assert from "node:assert/strict";
import { test } from "node:test";
import { acceptanceCases, hostAcceptanceLevel, validateHostAcceptance, validateStableHostAcceptance } from "../scripts/validate-host-acceptance.mjs";
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
  const evidence = { method: "real-host", host: "codex", os: "macOS", hostVersion: "fixture-host-version", osVersion: "fixture-os", osArchitecture: "arm64", executionMode: "local", shellVersion: "fixture-shell", executorPlatform: "darwin-arm64", executorSha256: sha, version, checkedAt: "2026-09-06", artifact: "references/fixture.md", cases: Object.fromEntries(acceptanceCases.map(id => [id, id === "installation" ? "passed" : "pending"])) };
  record.evidence.push(evidence);
  assert.deepEqual(validateHostAcceptance(record, version, manifest, ["codex"]), []);
  for (const [key, value] of [["method", "fixture"], ["version", "0.16.0"], ["executorSha256", "b".repeat(64)], ["hostVersion", ""], ["osArchitecture", ""], ["shellVersion", ""], ["executionMode", "unknown"], ["executionMode", "remote"]]) {
    const changed = structuredClone(record);
    changed.evidence[0][key] = value;
    assert.ok(validateHostAcceptance(changed, version, manifest, ["codex"]).length, key);
  }
});

test("Linux evidence requires Linux bytes and cannot masquerade as a different OS", () => {
 const record = fixture(); const linuxManifest = { artifacts: { "linux-arm64": { sha256: sha } } };
 record.hosts[0].linux.installation = "passed";
 record.evidence.push({ method:"real-host",host:"codex",os:"linux",hostVersion:"fixture",osVersion:"fixture",osArchitecture:"arm64",executionMode:"local",shellVersion:"fixture",executorPlatform:"linux-arm64",executorSha256:sha,version,checkedAt:"2026-09-08",artifact:"references/fixture.md",cases:Object.fromEntries(acceptanceCases.map(id => [id,id==="installation"?"passed":"pending"])) });
 assert.deepEqual(validateHostAcceptance(record,version,linuxManifest,["codex"]),[]);
 record.evidence[0].os="macOS";
 assert.ok(validateHostAcceptance(record,version,linuxManifest,["codex"]).length);
});

test("installation and API success never imply complete media delivery", () => {
  const evidence = {cases: Object.fromEntries(acceptanceCases.map(id => [id, "pending"]))};
  assert.equal(hostAcceptanceLevel(evidence), "unverified");
  evidence.cases.installation = "passed";
  assert.equal(hostAcceptanceLevel(evidence), "installation-verified");
  evidence.cases.authenticatedAPI = "passed";
  assert.equal(hostAcceptanceLevel(evidence), "api-verified");
  for (const id of acceptanceCases) evidence.cases[id] = "passed";
  assert.equal(hostAcceptanceLevel(evidence), "media-verified");
  evidence.cases["video-generation"] = "unavailable";
  assert.equal(hostAcceptanceLevel(evidence), "api-verified");
});

test("WSL evidence belongs to Linux and cannot accept a local host summary", () => {
  const record = fixture();
  const evidence = {
    method: "real-host", host: "codex", os: "linux", executionMode: "wsl",
    hostVersion: "fixture", osVersion: "fixture", osArchitecture: "amd64", shellVersion: "fixture",
    executorPlatform: "linux-amd64", executorSha256: sha, version, checkedAt: "2026-09-12",
    artifact: "references/fixture.md", cases: Object.fromEntries(acceptanceCases.map(id => [id, "passed"])),
  };
  record.evidence.push(evidence);
  const artifacts = Object.fromEntries(["linux-amd64", "windows-amd64", "darwin-amd64"].map(id => [id, {sha256: sha}]));
  assert.deepEqual(validateHostAcceptance(record, version, {artifacts}, ["codex"]), []);
  for (const [os, platform] of [["windows", "windows-amd64"], ["macOS", "darwin-amd64"]]) {
    evidence.os = os;
    evidence.executorPlatform = platform;
    assert.ok(validateHostAcceptance(record, version, {artifacts}, ["codex"]).some(error => error.includes("WSL")));
  }
  evidence.os = "linux";
  evidence.executorPlatform = "linux-amd64";
  record.hosts[0].linux.installation = "passed";
  assert.ok(validateHostAcceptance(record, version, {artifacts}, ["codex"]).some(error => error.includes("local real-host")));
});
test("unavailable host OS combinations require an explanation", () => {
 const record=fixture(); record.hosts[0].linux.installation="unavailable";
 assert.ok(validateHostAcceptance(record,version,manifest,["codex"]).length);
 record.hosts[0].linux.reason="No verified client target for this release";
 assert.deepEqual(validateHostAcceptance(record,version,manifest,["codex"]),[]);
});

test("stable publication cannot promote pending, empty or partially delivered evidence", () => {
  const record = fixture();
  assert.ok(validateStableHostAcceptance(record).length);
  for (const os of ["macOS","windows","linux"]) for (const id of ["installation","authenticatedAPI","sameTaskResume","nativeAttachmentDelivery"]) record.hosts[0][os][id] = "unavailable";
  assert.ok(validateStableHostAcceptance(record).length, "all unavailable is not acceptance");
  for (const id of ["installation","authenticatedAPI","sameTaskResume","nativeAttachmentDelivery"]) record.hosts[0].macOS[id] = "passed";
  record.evidence.push({method:"real-host",host:"codex",os:"macOS",executionMode:"local",cases:Object.fromEntries(acceptanceCases.map(id=>[id,"passed"]))});
  assert.deepEqual(validateStableHostAcceptance(record),[]);
  record.evidence[0].cases["video-generation"] = "pending";
  assert.ok(validateStableHostAcceptance(record).length);
  record.evidence[0].cases["video-generation"] = "passed";
  record.evidence[0].executionMode = "remote";
  assert.ok(validateStableHostAcceptance(record).length);
});
