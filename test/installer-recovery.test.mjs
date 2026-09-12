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
test("desktop hosts locate and install into isolated local skill roots", async t => {
 const f = await fixture(t);
 for (const [host, key] of [["claude-desktop", "CLAUDE_CONFIG_DIR"], ["dsh-desktop", "DSH_HOME"]]) {
  const directory = path.join(f.root, `${host} with spaces`);
  const env = { ...f.env, [key]: directory };
  const { stdout: located } = await execFile("sh", [installer, "locate", "--host", host], { env });
  assert.equal(located.trim(), path.join(directory, "skills"));
  await execFile("sh", [installer, "sync", "--host", host], { env });
  for (const name of names) {
   const manifest = JSON.parse(await readFile(path.join(directory, "skills", name, "skill.json"), "utf8"));
   assert.ok(manifest.supportedClients.includes(host));
  }
  assert.equal((await readdir(path.join(directory, "skills"))).filter(x => x === ".puretokens-executor").length, 1);
  await assert.rejects(execFile("sh", [installer, "locate", "--host", host], { env: { ...env, [key]: "relative" } }));
 }
 // Default paths are platform-owned; no synthetic connection is opened.
 const env = { ...f.env, CLAUDE_CONFIG_DIR: "", DSH_HOME: "" };
 const { stdout } = await execFile("sh", [installer, "locate", "--host", "claude-desktop"], { env });
 assert.equal(stdout.trim(), path.join(env.HOME, ".claude", "skills"));
 if (process.platform === "darwin") {
  const { stdout } = await execFile("sh", [installer, "locate", "--host", "dsh-desktop"], { env });
  assert.equal(stdout.trim(), path.join(env.HOME, "Library/Application Support/dsh-desktop/harness/skills"));
 } else {
  await assert.rejects(execFile("sh", [installer, "locate", "--host", "dsh-desktop"], { env }), /local macOS or Windows/);
 }
});
test("ZCode honors the data base and installs all managed components", async t => {
 const f = await fixture(t);
 const base = path.join(f.root, "zcode data with spaces");
 const env = { ...f.env, ZCODE_DATA_BASE_DIR: base };
 const target = path.join(base, ".zcode", "skills");
 const { stdout } = await execFile("sh", [installer, "locate", "--host", "zcode"], { env });
 assert.equal(stdout.trim(), target);
 await execFile("sh", [installer, "sync", "--host", "zcode"], { env });
 for (const name of names) assert.ok(JSON.parse(await readFile(path.join(target, name, "skill.json"), "utf8")).supportedClients.includes("zcode"));
 const executor = path.join(target, ".puretokens-executor", "puretokens-api");
 await execFile(executor, ["--version"]);
 await execFile("sh", [installer, "sync", "--host", "zcode"], { env });
 await assert.rejects(execFile("sh", [installer, "locate", "--host", "zcode"], { env: { ...env, ZCODE_DATA_BASE_DIR: "relative" } }));
 const result = await execFile("sh", [installer, "locate", "--host", "zcode"], { env: { ...env, ZCODE_DATA_BASE_DIR: "" } });
 assert.equal(result.stdout.trim(), path.join(env.HOME, ".zcode", "skills"));
});
test("a failed update restores every existing Skill and releases its lock", async t => {
 const f = await fixture(t); await install(f);
 const originals = new Map();
 for (const name of names) {
  const file = path.join(f.target, name, "SKILL.md");
  originals.set(name, await readFile(file, "utf8"));
 }
 const tools = path.join(f.root, "tools"); await mkdir(tools);
 const wrapper = path.join(tools, "mv");
 await writeFile(wrapper, '#!/bin/sh\ncase "$2" in */backup/puretokens-connection) exit 73 ;; esac\nexec /bin/mv "$@"\n'); await chmod(wrapper, 0o700);
 await assert.rejects(install(f, { ...f.env, PATH: `${tools}${path.delimiter}${f.env.PATH}` }));
 for (const name of names) assert.equal(await readFile(path.join(f.target, name, "SKILL.md"), "utf8"), originals.get(name));
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
test("Gemini selects an existing shared Skill and detects its lower-priority duplicate", async t => {
 const f = await fixture(t);
 const shared = path.join(f.env.HOME, ".agents/skills");
 const legacy = path.join(f.env.HOME, ".gemini/skills");
 await mkdir(shared, { recursive: true });
 await mkdir(legacy, { recursive: true });
 await cp(path.join(repositoryRoot, "skills/puretokens-image"), path.join(shared, "puretokens-image"), { recursive: true });
 await cp(path.join(repositoryRoot, "skills/puretokens-image"), path.join(legacy, "puretokens-image"), { recursive: true });
 const { stdout: selected } = await execFile("sh", [installer, "locate", "--host", "gemini-cli"], { env: f.env });
 assert.equal(selected.trim(), shared);
 f.target = shared;
 const { stdout } = await install(f, f.env, ["--host", "gemini-cli"]);
 assert.match(stdout, /Managed duplicate detected/);
 assert.match(await readFile(path.join(legacy, "puretokens-image/SKILL.md"), "utf8"), /Pure Tokens/);
 assert.match(await readFile(path.join(shared, ".puretokens-executor/puretokens-skill-fetch.sh"), "utf8"), /check-update/);
 f.target = legacy;
 await assert.rejects(install(f, f.env, ["--host", "gemini-cli"]), /higher-priority/);
});

test("new hosts install and update in their declared isolated roots", async t => {
 const f = await fixture(t);
 for (const [host, variable] of [["kimi-code", "KIMI_CODE_HOME"], ["qoder", "QODER_CONFIG_DIR"], ["pi", "PI_CODING_AGENT_DIR"]]) {
  const root = path.join(f.root, `${host} \u7528\u6237 with spaces`);
  const env = {...f.env, [variable]: root};
  const target = path.join(root, "skills");
  const located = await execFile("sh", [installer, "locate", "--host", host], {env});
  assert.equal(located.stdout.trim(), target);
  for (let i=0;i<2;i++) await execFile("sh", [installer, "sync", "--host", host], {env});
  assert.ok(await readFile(path.join(target, "puretokens-image", "SKILL.md")));
  assert.ok(await readFile(path.join(target, ".puretokens-executor", "puretokens-api")));
  await assert.rejects(execFile("sh", [installer, "sync", "--host", host], {env:{...env,[variable]:"relative"}}));
 }
 const piDefault = await execFile("sh", [installer, "locate", "--host", "pi"], {env:{...f.env,PI_CODING_AGENT_DIR:""}});
 assert.equal(piDefault.stdout.trim(), path.join(f.env.HOME, ".pi", "agent", "skills"));
 const invalidPi = `${f.root}/invalid/../unexpected`;
 for (const command of ["locate", "sync"]) {
  await assert.rejects(execFile("sh", [installer, command, "--host", "pi"], {env:{...f.env,PI_CODING_AGENT_DIR:invalidPi}}));
 }
 await assert.rejects(readdir(path.join(f.root, "invalid")), {code:"ENOENT"});
 await assert.rejects(readdir(path.join(f.root, "unexpected")), {code:"ENOENT"});
 const env={...f.env,QODER_CONFIG_DIR:"",QODER_CLI_HOME:path.join(f.root,"qoder parent"),QODER_CONFIG_DIR_NAME:"custom"};
 const result=await execFile("sh",[installer,"locate","--host","qoder"],{env});
 assert.equal(result.stdout.trim(),path.join(env.QODER_CLI_HOME,"custom","skills"));
 await assert.rejects(execFile("sh",[installer,"locate","--host","qoder"],{env:{...env,QODER_CONFIG_DIR_NAME:"../bad"}}));
});

test("installation never invokes Codex plugins and still runs host init", async t => {
 const f=await fixture(t);
 const mockBin=path.join(f.root,"mock-bin");await mkdir(mockBin);
 const marker=path.join(f.root,"codex-called");
 await writeFile(path.join(mockBin,"codex"),'#!/bin/sh\nprintf called >> "$PLUGIN_MARKER"\nprintf \'{"installed":[]}\\n\'\n');
 await chmod(path.join(mockBin,"codex"),0o755);
 const env={...f.env,PATH:mockBin+path.delimiter+f.env.PATH,PLUGIN_MARKER:marker,QODER_CONFIG_DIR:path.join(f.root,"missing-qoder")};
 const shared=path.join(f.home??f.env.HOME,".agents","skills");
 const {stdout}=await execFile("sh",[installer,"sync","--host","qoder","--target",shared],{env});
 await assert.rejects(readFile(marker),{code:"ENOENT"});
 assert.match(stdout,/Pure Tokens Skill init:/);
 assert.doesNotMatch(stdout,/connection check was deferred/);
 assert.match(stdout,/Installation host: qoder/);
 assert.match(stdout,/Start a new qoder conversation/);
 await execFile("sh",[installer,"sync","--host","codex","--target",shared],{env});
 await assert.rejects(readFile(marker),{code:"ENOENT"});
 await execFile(path.join(mockBin,"codex"),[],{env});
 assert.equal(await readFile(marker,"utf8"),"called");
});
