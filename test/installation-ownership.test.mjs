import assert from "node:assert/strict";
import { test } from "node:test";
import { execFile as callback } from "node:child_process";
import { promisify } from "node:util";
import { mkdir, mkdtemp, readFile, readdir, rm, writeFile, cp } from "node:fs/promises";
import { createHash } from "node:crypto";
import path from "node:path";
import os from "node:os";
import { repositoryRoot } from "../scripts/skill-registry.mjs";
import { getManagedSkillProvenances } from "../scripts/media-skill-provenance.mjs";
const execFile = promisify(callback);
const installer = path.join(repositoryRoot, "runtime/puretokens-skill-install.sh");
async function fixture(t) {
  const root = await mkdtemp(path.join(os.tmpdir(), "pt-ownership-"));
  t.after(() => rm(root, { recursive: true, force: true }));
  const target = path.join(root, "skills");
  const install = source => execFile("sh", [installer, "sync", "--target", target, "--source", source ?? repositoryRoot]);
  return { root, target, install };
}
test("upgrade preserves unmanaged same-name Skill and added or modified files", async t => {
  const f = await fixture(t);
  const dir = path.join(f.target, "puretokens-image");
  await mkdir(dir, { recursive: true });
  await writeFile(path.join(dir, "SKILL.md"), "# Independent Skill");
  await writeFile(path.join(dir, "skill.json"), JSON.stringify({ name: "puretokens-image" }, null, 2));
  await assert.rejects(f.install(), /ownership/);
  assert.equal(await readFile(path.join(dir, "SKILL.md"), "utf8"), "# Independent Skill");
  await rm(dir, { recursive: true });
  await f.install();
  for (const file of ["personal.txt", "SKILL.md"]) {
    const destination = path.join(dir, file);
    const original = await readFile(destination).catch(() => null);
    await writeFile(destination, "keep my local work");
    await assert.rejects(f.install(), /ownership/);
    assert.equal(await readFile(destination, "utf8"), "keep my local work");
    if (original) await writeFile(destination, original); else await rm(destination);
  }
  await f.install();
  assert.equal((await readdir(f.target)).some(name => name.includes("stage")), false);
});
test("stock 0.17.0 installations migrate without trusting names or self-reported hashes", async t => {
  const f = await fixture(t);
  const old = path.join(f.root, "old-source");
  await mkdir(old);
  const archive = path.join(f.root, "old.tar");
  await execFile("git", ["archive", "--output", archive, "1cc3082", "skills", "runtime", "package.json", "README.md"], { cwd: repositoryRoot });
  await execFile("tar", ["-xf", archive, "-C", old]);
  await execFile("sh", [path.join(old, "runtime/puretokens-skill-install.sh"), "sync", "--target", f.target, "--source", old]);
  await f.install();
  assert.equal(JSON.parse(await readFile(path.join(f.target, "puretokens-image/.puretokens-managed.json"), "utf8")).format, "puretokens-managed-files-v1");
});
test("interrupted update recovery preserves user changes and retains its backup", async t => {
  const f = await fixture(t);
  await f.install();
  const dir = path.join(f.target, "puretokens-image");
  const stage = path.join(f.target, ".puretokens-skill-stage.user-change");
  await mkdir(path.join(stage, "backup"), { recursive: true });
  await cp(dir, path.join(stage, "backup/puretokens-image"), { recursive: true });
  await writeFile(path.join(stage, "transaction-v1"), "");
  await writeFile(path.join(stage, "plan"), "replace puretokens-image\n");
  await writeFile(path.join(dir, "personal.txt"), "keep");
  await assert.rejects(f.install(), /recovery failed/);
  assert.equal(await readFile(path.join(dir, "personal.txt"), "utf8"), "keep");
  assert.ok(await readFile(path.join(stage, "backup/puretokens-image/SKILL.md")));
});
test("all managed provenance entries derive from the registry", async () => {
  const entries = await getManagedSkillProvenances();
  assert.equal(entries.length, 6);
  for (const entry of entries) {
    assert.match(entry.name, /^puretokens-/);
    assert.match(entry.sourceSha256, /^[a-f0-9]{64}$/);
  }
});

test("every historical migration inventory accepts its exact repository bytes", async t => {
  const f = await fixture(t);
  const history = JSON.parse(await readFile(path.join(repositoryRoot, "runtime/executor/installation-history.json"), "utf8"));
  const platform = `${process.platform === "darwin" ? "darwin" : "linux"}-${process.arch === "arm64" ? "arm64" : "amd64"}`;
  const guard = path.join(repositoryRoot, "runtime/executor/bin", `puretokens-api-${platform}`);
  const trees = new Map(), blobs = new Map();
  const git = async args => (await execFile("git", args, { cwd: repositoryRoot, encoding: "buffer", maxBuffer: 64 << 20 })).stdout;
  async function tree(revision) {
    if (!trees.has(revision)) {
      const output = await git(["ls-tree", "-rz", revision, "--", "skills", "runtime", "scripts/legacy-bootstrap"]);
      trees.set(revision, new Map(output.toString().split("\0").filter(Boolean).map(line => {
        const [, hash, file] = line.match(/^\d+ blob ([a-f0-9]+)\t(.+)$/);
        return [file, hash];
      })));
    }
    return trees.get(revision);
  }
  async function read(revision, file) {
    const hash = (await tree(revision)).get(file);
    assert.ok(hash, `missing historical ${file}`);
    if (!blobs.has(hash)) blobs.set(hash, await git(["cat-file", "blob", hash]));
    return blobs.get(hash);
  }
  const hashJSON = bytes => {
    const sort = value => Array.isArray(value) ? value.map(sort) : value && typeof value === "object"
      ? Object.fromEntries(Object.keys(value).sort().map(key => [key, sort(value[key])])) : value;
    return createHash("sha256").update(JSON.stringify(sort(JSON.parse(bytes.toString().replace(/^\uFEFF/, ""))))).digest("hex");
  };
  let index = 0;
  for (const snapshot of history.snapshots) {
    const directory = path.join(f.root, `historical-${index++}`);
    await mkdir(directory);
    if (snapshot.name === ".puretokens-executor") {
      const manifest = JSON.parse(await read(snapshot.revision, "runtime/executor/manifest.json"));
      const binary = snapshot.files["puretokens-api"] ?? snapshot.files["puretokens-api.exe"];
      const [historicalPlatform, artifact] = Object.entries(manifest.artifacts).find(([, entry]) => entry.sha256 === binary);
      const windows = historicalPlatform.startsWith("windows-");
      await writeFile(path.join(directory, windows ? "puretokens-api.exe" : "puretokens-api"), await read(snapshot.revision, `runtime/executor/${artifact.path}`));
      await writeFile(path.join(directory, "runtime.json"), JSON.stringify({ schemaVersion: 1, name: "puretokens-api-executor", version: manifest.version, platform: historicalPlatform }));
      for (const file of Object.keys(snapshot.files).filter(file => file.endsWith(".sh") || file.endsWith(".ps1"))) {
        await writeFile(path.join(directory, file), await read(snapshot.revision, `runtime/${file}`));
      }
    } else if (snapshot.name === ".puretokens-runtime") {
      let prefix;
      for (const candidate of ["runtime", "scripts/legacy-bootstrap"]) {
        if (!(await tree(snapshot.revision)).has(`${candidate}/runtime.json`)) continue;
        if (hashJSON(await read(snapshot.revision, `${candidate}/runtime.json`)) === snapshot.files["runtime.json"]) {
          prefix = candidate;
          break;
        }
      }
      assert.ok(prefix, "legacy runtime source not found");
      for (const file of Object.keys(snapshot.files)) await writeFile(path.join(directory, file), await read(snapshot.revision, `${prefix}/${file}`));
    } else {
      for (const file of Object.keys(snapshot.files)) {
        const destination = path.join(directory, file);
        if (file.endsWith("/")) await mkdir(destination, { recursive: true });
        else {
          await mkdir(path.dirname(destination), { recursive: true });
          await writeFile(destination, await read(snapshot.revision, `skills/${snapshot.name}/${file}`));
        }
      }
    }
    await execFile(guard, ["install-verify", "--directory", directory, "--name", snapshot.name]);
    // Every retired version must also refuse to delete added user data.
    await writeFile(path.join(directory, "user-owned.txt"), "preserve");
    await assert.rejects(execFile(guard, ["install-verify", "--directory", directory, "--name", snapshot.name]));
    await rm(directory, { recursive: true });
  }
});
