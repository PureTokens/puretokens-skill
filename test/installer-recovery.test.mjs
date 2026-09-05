import assert from "node:assert/strict";
import { test } from "node:test";
import { execFile as execFileCallback } from "node:child_process";
import { promisify } from "node:util";
import { mkdtemp, mkdir, readFile, readdir, rm, writeFile, chmod, rename, cp } from "node:fs/promises";
import path from "node:path";
import os from "node:os";
import { repositoryRoot } from "../scripts/skill-registry.mjs";
const execFile = promisify(execFileCallback);
const names = ["puretokens-balance", "puretokens-connection", "puretokens-models", "puretokens-image", "puretokens-video", "puretokens-update"];
const installer = path.join(repositoryRoot, "runtime/puretokens-skill-install.sh");
async function fixture(t) {
 const root = await mkdtemp(path.join(os.tmpdir(), "pt-installer-test-"));
 t.after(() => rm(root, { recursive: true, force: true }));
 const target = path.join(root, "target");
 const home = path.join(root, "home"); await mkdir(home);
 return { root, target, env: { ...process.env, HOME: home, CODEX_HOME: path.join(home, ".codex") } };
}
async function install(f, env = f.env, more = []) {
 return execFile("sh", [installer, "sync", "--target", f.target, ...more], { cwd: repositoryRoot, env });
}
async function waitForFile(file) {
 const deadline = Date.now() + 8000;
 while (Date.now() < deadline) {
  try { await readFile(file); return; } catch (error) { if (error.code !== "ENOENT") throw error; }
  await new Promise(resolve => setTimeout(resolve, 10));
 }
 assert.fail(`Timed out waiting for fixture marker ${path.basename(file)}`);
}
test("legacy CLI sync installs the native executor", async t => {
 const f = await fixture(t);
 await execFile(process.execPath, ["bin/puretokens-skill.js", "sync", "--target", f.target], { cwd: repositoryRoot, env: f.env });
 const executor = path.join(f.target, ".puretokens-executor/puretokens-api");
 const { stdout } = await execFile(executor, ["--version"]);
 assert.equal(stdout.trim(), "0.16.0");
});
test("a failed update restores every existing Skill and releases its lock", async t => {
 const f = await fixture(t); await install(f);
 for (const name of names) {
  const file = path.join(f.target, name, "SKILL.md");
  await writeFile(file, `${await readFile(file, "utf8")}\nOLD_REVISION_FIXTURE\n`);
 }
 const tools = path.join(f.root, "tools"); await mkdir(tools);
 const wrapper = path.join(tools, "mv");
 await writeFile(wrapper, '#!/bin/sh\ncase "$2" in */backup/puretokens-connection) exit 73 ;; esac\nexec /bin/mv "$@"\n'); await chmod(wrapper, 0o700);
 await assert.rejects(install(f, { ...f.env, PATH: `${tools}${path.delimiter}${f.env.PATH}` }));
 for (const name of names) assert.match(await readFile(path.join(f.target, name, "SKILL.md"), "utf8"), /OLD_REVISION_FIXTURE/);
 assert.equal((await readdir(f.target)).some(x => x.includes("stage") || x.includes("lock")), false);
 await install(f);
});
test("next sync restores a recoverable interrupted transaction", async t => {
 const f = await fixture(t); await install(f);
 const stage = path.join(f.target, ".puretokens-skill-stage.fixture");
 await mkdir(path.join(stage, "backup"), { recursive: true });
 await writeFile(path.join(stage, "transaction-v1"), "");
 await writeFile(path.join(stage, "plan"), "replace puretokens-image\n");
 await rename(path.join(f.target, "puretokens-image"), path.join(stage, "backup/puretokens-image"));
 await cp(path.join(repositoryRoot, "skills/puretokens-image"), path.join(f.target, "puretokens-image"), { recursive: true });
 await install(f);
 assert.equal((await readdir(f.target)).includes(".puretokens-skill-stage.fixture"), false);
 assert.match(await readFile(path.join(f.target, "puretokens-image/SKILL.md"), "utf8"), /submit/);
});
test("live update lock prevents all installation writes", async t => {
 const f = await fixture(t); await mkdir(path.join(f.target, ".puretokens-install-lock"), { recursive: true });
 await writeFile(path.join(f.target, ".puretokens-install-lock/pid"), `${process.pid}\n`);
 await assert.rejects(install(f), /another installation is in progress/);
 assert.deepEqual(await readdir(f.target), [".puretokens-install-lock"]);
});
test("a filesystem without hard links supports installation and preserves incomplete locks", async t => {
 const f = await fixture(t);
 const tools = path.join(f.root, "tools"); await mkdir(tools);
 await writeFile(path.join(tools, "ln"), "#!/bin/sh\nexit 1\n");
 await chmod(path.join(tools, "ln"), 0o700);
 const env = { ...f.env, PATH: `${tools}${path.delimiter}${f.env.PATH}` };
 await install(f, env);
 assert.equal((await readdir(f.target)).some(name => name.includes("stage") || name.includes("lock")), false);
 const lock = path.join(f.target, ".puretokens-install-lock");
 await mkdir(lock);
 await writeFile(path.join(lock, "pid"), "");
 await assert.rejects(install(f, env), /update lock needs inspection/);
 assert.equal(await readFile(path.join(lock, "pid"), "utf8"), "");
 assert.deepEqual(await readdir(lock), ["pid"]);
});
for (const hardLinks of [true, false]) test(`concurrent stale-lock recovery cannot replace the winning live updater (${hardLinks ? "hard links" : "exclusive creation"})`, async t => {
 const f = await fixture(t); await install(f);
 const lock = path.join(f.target, ".puretokens-install-lock");
 await mkdir(lock);
 await writeFile(path.join(lock, "pid"), "2147483647\n");
 const tools = path.join(f.root, "tools"); await mkdir(tools);
 if (!hardLinks) {
  await writeFile(path.join(tools, "ln"), "#!/bin/sh\nexit 1\n");
  await chmod(path.join(tools, "ln"), 0o700);
 }
 const coordination = path.join(f.root, "coordination"); await mkdir(coordination);
 const waitForRelease = [
  'attempt=0',
  'while [ ! -f "$release" ]; do',
  ' attempt=$((attempt + 1)); [ "$attempt" -lt 1500 ] || exit 91',
  ' sleep 0.01',
  'done'
 ].join("\n");
 await writeFile(path.join(tools, "cat"), [
  "#!/bin/sh",
  'case "$1" in */.puretokens-install-lock/pid) is_lock_pid=true ;; *) is_lock_pid=false ;; esac',
  'if [ "$PT_LOCK_PEER" = second ] && $is_lock_pid && [ ! -f "$PT_LOCK_COORDINATION/stale-read" ]; then',
  ' /bin/cat "$@" || exit',
  ' : > "$PT_LOCK_COORDINATION/stale-read"',
  ' release="$PT_LOCK_COORDINATION/release-second"',
  waitForRelease,
  ' exit 0',
  'fi',
  'exec /bin/cat "$@"'
 ].join("\n"));
 await writeFile(path.join(tools, "cp"), [
  "#!/bin/sh",
  'if [ "$PT_LOCK_PEER" = first ] && [ "$1" = -R ] && [ ! -f "$PT_LOCK_COORDINATION/first-copy" ]; then',
  ' : > "$PT_LOCK_COORDINATION/first-copy"',
  ' release="$PT_LOCK_COORDINATION/release-first"',
  waitForRelease,
  'fi',
  'exec /bin/cp "$@"'
 ].join("\n"));
 await chmod(path.join(tools, "cat"), 0o700);
 await chmod(path.join(tools, "cp"), 0o700);
 const env = { ...f.env, PATH: `${tools}${path.delimiter}${f.env.PATH}`, PT_LOCK_COORDINATION: coordination };
 const settle = promise => promise.then(value => ({ ok: true, value }), error => ({ ok: false, error }));
 const second = settle(install(f, { ...env, PT_LOCK_PEER: "second" }));
 let first;
 try {
  await waitForFile(path.join(coordination, "stale-read"));
  first = settle(install(f, { ...env, PT_LOCK_PEER: "first" }));
  await waitForFile(path.join(coordination, "first-copy"));
  await writeFile(path.join(coordination, "release-second"), "");
  const result = await second;
  assert.equal(result.ok, false, "the losing recovery must not enter the live updater's transaction");
  assert.match(result.error.stderr, /another installation is in progress/);
  assert.equal((await readdir(f.target)).filter(name => name.startsWith(".puretokens-skill-stage.")).length, 1);
 } finally {
  await writeFile(path.join(coordination, "release-second"), "");
  await writeFile(path.join(coordination, "release-first"), "");
  await second;
  if (first) {
   const result = await first;
   assert.equal(result.ok, true, result.error?.stderr);
  }
 }
 assert.equal((await readdir(f.target)).some(name => name.includes("stage") || name.includes("lock")), false);
});
test("a killed stale-lock takeover remains recoverable by the next sync", async t => {
 const f = await fixture(t); await install(f);
 const lock = path.join(f.target, ".puretokens-install-lock");
 await mkdir(lock);
 await writeFile(path.join(lock, "pid"), "2147483647\n");
 await writeFile(path.join(lock, "next.legacy-2147483647"), "2147483646 owner.interrupted\n");
 const stage = path.join(f.target, ".puretokens-skill-stage.interrupted");
 await mkdir(path.join(stage, "backup"), { recursive: true });
 await writeFile(path.join(stage, "transaction-v1"), "");
 await writeFile(path.join(stage, "plan"), "replace puretokens-image\n");
 await rename(path.join(f.target, "puretokens-image"), path.join(stage, "backup/puretokens-image"));
 await install(f);
 assert.equal((await readdir(f.target)).some(name => name.includes("stage") || name.includes("lock")), false);
 assert.match(await readFile(path.join(f.target, "puretokens-image/SKILL.md"), "utf8"), /submit/);
});
test("Codex plugin inspection failure does not block a clean installation", async t => {
 const f = await fixture(t);f.target=path.join(f.env.HOME,".agents/skills");
 const tools=path.join(f.root,"tools");await mkdir(tools);
 const command=path.join(tools,"codex");await writeFile(command,"#!/bin/sh\nexit 127\n");await chmod(command,0o700);
 const {stdout}=await install(f,{...f.env,PATH:`${tools}${path.delimiter}${f.env.PATH}`},["--host","codex"]);
 assert.match(stdout,/synchronized with the native API executor/);
 assert.match(stdout,/plugin inspection unavailable/);
});
