// Maintainer-only migration inventory, generated from reviewed repository
// history. Never run during installation and never inspect a user installation.
import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { writeFile } from "node:fs/promises";
import { repositoryRoot } from "./skill-registry.mjs";

const git = args => execFileSync("git", args, { cwd: repositoryRoot, maxBuffer: 64 << 20 });
const sortedObject = value => Array.isArray(value) ? value.map(sortedObject) : value && typeof value === "object"
  ? Object.fromEntries(Object.keys(value).sort().map(key => [key, sortedObject(value[key])])) : value;
const canonical = value => JSON.stringify(sortedObject(value));
const digest = (name, bytes) => createHash("sha256").update(name === "runtime.json" ? canonical(JSON.parse(bytes.toString("utf8").replace(/^\uFEFF/, ""))) : bytes).digest("hex");
const commits = git(["log", "--format=%H", "--", "skills", "runtime", "scripts/legacy-bootstrap"]).toString().trim().split("\n");
const blobs = new Map();
const snapshots = [];
const seen = new Set();
function add(name, revision, files) {
  const expanded = { ...files };
  for (const file of Object.keys(files)) {
    let at = file.lastIndexOf("/");
    while (at >= 0) {
      expanded[file.slice(0, at + 1)] = "directory";
      at = file.lastIndexOf("/", at - 1);
    }
  }
  const sorted = Object.fromEntries(Object.entries(expanded).sort(([a], [b]) => a.localeCompare(b)));
  const identity = name + JSON.stringify(sorted);
  if (!seen.has(identity)) {
    snapshots.push({ name, revision, files: sorted });
    seen.add(identity);
  }
}
for (const revision of commits) {
  const entries = git(["ls-tree", "-rz", revision, "--", "skills", "runtime", "scripts/legacy-bootstrap"]).toString().split("\0").filter(Boolean).map(line => {
    const [, mode, hash, file] = line.match(/^(\d+) blob ([a-f0-9]+)\t(.+)$/) ?? [];
    return { mode, hash, file };
  }).filter(entry => entry.file && entry.mode !== "120000");
  const read = entry => {
    if (!blobs.has(entry.hash)) blobs.set(entry.hash, git(["cat-file", "blob", entry.hash]));
    return blobs.get(entry.hash);
  };
  const byPath = new Map(entries.map(entry => [entry.file, entry]));
  const skills = new Map();
  for (const entry of entries) {
    const match = entry.file.match(/^skills\/([^/]+)\/(.+)$/);
    if (!match) continue;
    const [, name, file] = match;
    if (!skills.has(name)) skills.set(name, {});
    skills.get(name)[file] = digest(file, read(entry));
  }
  for (const [name, files] of skills) if (files["SKILL.md"] && files["skill.json"]) add(name, revision, files);
  for (const prefix of ["runtime", "scripts/legacy-bootstrap"]) {
    if (byPath.has(`${prefix}/runtime.json`) && byPath.has(`${prefix}/puretokens-direct-api.mjs`)) {
      add(".puretokens-runtime", revision, Object.fromEntries(["runtime.json", "puretokens-direct-api.mjs"].map(file => [file, digest(file, read(byPath.get(`${prefix}/${file}`)))])));
    }
  }
  const manifestEntry = byPath.get("runtime/executor/manifest.json");
  if (!manifestEntry) continue;
  const manifest = JSON.parse(read(manifestEntry));
  for (const [platform, artifact] of Object.entries(manifest.artifacts)) {
    const windows = platform.startsWith("windows-");
    const files = { [windows ? "puretokens-api.exe" : "puretokens-api"]: artifact.sha256 };
    const extension = windows ? "ps1" : "sh";
    for (const script of ["puretokens-skill-install", "puretokens-skill-fetch"]) {
      const entry = byPath.get(`runtime/${script}.${extension}`);
      if (entry) files[`${script}.${extension}`] = digest(entry.file, read(entry));
    }
    const runtime = { schemaVersion: 1, name: "puretokens-api-executor", version: manifest.version, platform };
    files["runtime.json"] = digest("runtime.json", Buffer.from(JSON.stringify(runtime)));
    add(".puretokens-executor", revision, files);
  }
}
await writeFile(`${repositoryRoot}/runtime/executor/installation-history.json`, JSON.stringify({ schemaVersion: 1, sourceCommit: commits[0], snapshots }, null, 2) + "\n");
console.log(`Recorded ${snapshots.length} exact migration inventories from ${commits.length} repository revisions.`);
