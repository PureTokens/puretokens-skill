import assert from "node:assert/strict";
import { test } from "node:test";
import { execFile as execFileCallback } from "node:child_process";
import { promisify } from "node:util";
import { cp, mkdir, mkdtemp, readFile, readdir, rm, writeFile, chmod } from "node:fs/promises";
import { createHash } from "node:crypto";
import os from "node:os";
import path from "node:path";
import { repositoryRoot } from "../scripts/skill-registry.mjs";
const execFile = promisify(execFileCallback);
const commit = "a".repeat(40);
const version = "0.17.0";
const scripts = ["puretokens-skill-install.sh", "puretokens-skill-install.ps1", "puretokens-skill-fetch.sh", "puretokens-skill-fetch.ps1"];

async function fixture(t) {
  const root = await mkdtemp(path.join(os.tmpdir(), "pt-distribution-test-"));
  t.after(() => rm(root, { recursive: true, force: true }));
  const source = path.join(root, "puretokens-skill");
  await mkdir(path.join(source, "runtime/executor/bin"), { recursive: true });
  await mkdir(path.join(source, "scripts"));
  await mkdir(path.join(root, "home"));
  await mkdir(path.join(root, "tmp"));
  await writeFile(path.join(source, "README.md"), "fixture\n");
  await writeFile(path.join(source, "package.json"), JSON.stringify({ version }, null, 2));
  await cp(path.join(repositoryRoot, "skills"), path.join(source, "skills"), { recursive: true });
  for (const script of scripts) await cp(path.join(repositoryRoot, "runtime", script), path.join(source, "runtime", script));
  const platform = `${process.platform === "darwin" ? "darwin" : "linux"}-${process.arch === "arm64" ? "arm64" : "amd64"}`;
  const binary = Buffer.from(`#!/bin/sh\nprintf '%s\\n' '${version}'\n`);
  const binaryPath = `bin/puretokens-api-${platform}`;
  await writeFile(path.join(source, "runtime/executor", binaryPath), binary);
  await writeFile(path.join(source, "runtime/executor/manifest.json"), JSON.stringify({
    schemaVersion: 1, name: "puretokens-api-executor", version,
    artifacts: { [platform]: { path: binaryPath, sha256: createHash("sha256").update(binary).digest("hex") } }
  }, null, 2));
  const env = { ...process.env, HOME: path.join(root, "home"), CODEX_HOME: path.join(root, "home/.codex"), TMPDIR: path.join(root, "tmp"), GITHUB_SHA: commit };
  return { root, source, platform, env, target: path.join(root, "target") };
}

test("repackaging replaces the archive and removes deleted members", async t => {
  const f = await fixture(t);
  const builder = path.join(f.source, "scripts/package-platform-releases.mjs");
  await cp(path.join(repositoryRoot, "scripts/package-platform-releases.mjs"), builder);
  const obsolete = path.join(f.source, "skills/puretokens-image/obsolete.json");
  await writeFile(obsolete, "{}");
  await execFile(process.execPath, [builder], { env: f.env });
  await rm(obsolete);
  await execFile(process.execPath, [builder], { env: f.env });
  const releases = path.join(f.source, "dist/releases");
  const { stdout } = await execFile("unzip", ["-Z1", path.join(releases, `puretokens-skill-${version}-${f.platform}.zip`)]);
  assert.doesNotMatch(stdout, /obsolete.json/);
  assert.equal(stdout.split("\n").filter(p => /runtime\/executor\/bin\/.+/.test(p)).length, 1);
  assert.equal(JSON.parse(await readFile(path.join(releases, "release-manifest.json"), "utf8")).sourceCommit, null);
  assert.equal((await readdir(releases)).some(p => p.endsWith(".tmp")), false);
});

async function mockDownloads(f, mode = "source") {
  const tools = path.join(f.root, "tools");
  await mkdir(tools);
  await writeFile(path.join(f.root, "commit.json"), JSON.stringify({ sha: commit }, null, 2));
  const sourceCopy = path.join(f.root, `puretokens-skill-${commit}`);
  await cp(f.source, sourceCopy, { recursive: true });
  await execFile("zip", ["-qr", path.join(f.root, "source.zip"), path.basename(sourceCopy)], { cwd: f.root });
  await execFile("zip", ["-qr", path.join(f.root, "platform.zip"), "puretokens-skill"], { cwd: f.root });
  const checksum = createHash("sha256").update(await readFile(path.join(f.root, "platform.zip"))).digest("hex");
  await writeFile(path.join(f.root, "release.json"), JSON.stringify({ version, sourceCommit: commit, files: {
    [f.platform]: { filename: `puretokens-skill-${version}-${f.platform}.zip`, sha256: mode === "bad-checksum" ? "0".repeat(64) : checksum }
  } }, null, 2));
  const curl = path.join(tools, "curl");
  await writeFile(curl, `#!/bin/sh
while [ "$#" -gt 0 ]; do
  case "$1" in --output) destination=$2; shift 2 ;; https:*) request_url=$1; shift ;; *) shift ;; esac
done
printf '%s\\n' "$request_url" >> "$PT_FETCH_FIXTURE/requests"
case "$request_url" in
  https://api.github.com/repos/PureTokens/puretokens-skill/commits/main) file=commit.json ;;
  https://raw.githubusercontent.com/PureTokens/puretokens-skill/${commit}/package.json) file=puretokens-skill/package.json ;;
  https://raw.githubusercontent.com/PureTokens/puretokens-skill/${commit}/runtime/puretokens-skill-install.sh) file=puretokens-skill/runtime/puretokens-skill-install.sh ;;
  https://github.com/PureTokens/puretokens-skill/releases/download/v${version}/release-manifest.json)
    if [ "$PT_FETCH_MODE" = source ]; then printf 404; exit 0; fi
    file=release.json ;;
  https://github.com/PureTokens/puretokens-skill/releases/download/v${version}/puretokens-skill-${version}-${f.platform}.zip) file=platform.zip ;;
  https://codeload.github.com/PureTokens/puretokens-skill/zip/${commit})
    if [ -n "\${PT_INSTALL_DURING_DOWNLOAD:-}" ]; then
      sh "$PT_INSTALL_DURING_DOWNLOAD/runtime/puretokens-skill-install.sh" sync --target "$PT_FIXTURE_TARGET" --source "$PT_INSTALL_DURING_DOWNLOAD" > "$PT_FETCH_FIXTURE/concurrent-update"
    fi
    file=source.zip ;;
  *) exit 99 ;;
esac
cp "$PT_FETCH_FIXTURE/$file" "$destination"
printf 200
`);
  await chmod(curl, 0o700);
  return { ...f.env, PATH: `${tools}${path.delimiter}${f.env.PATH}`, PT_FETCH_FIXTURE: f.root, PT_FETCH_MODE: mode };
}

test("check-update reads only pinned metadata and does not install or initialize", async t => {
  const f = await fixture(t);
  const env = await mockDownloads(f);
  const wrapper = path.join(f.source, "runtime/puretokens-skill-fetch.sh");
  const { stdout } = await execFile("sh", [wrapper, "check-update", "--target", f.target], { env });
  assert.match(stdout, /installed=not_installed available=0.17.0/);
  await assert.rejects(readFile(path.join(f.target, ".puretokens-executor/runtime.json")));
  const requests = (await readFile(path.join(f.root, "requests"), "utf8")).trim().split("\n");
  assert.equal(requests.length, 2);
  assert.ok(requests[1].includes(commit));
  assert.deepEqual(await readdir(path.join(f.root, "tmp")), []);
});

for (const mode of ["source", "platform"]) test(`download wrapper installs verified ${mode} through native sync`, async t => {
  const f = await fixture(t);
  const env = await mockDownloads(f, mode);
  const standalone = path.join(f.root, "puretokens-skill-fetch.sh");
  await cp(path.join(f.source, "runtime/puretokens-skill-fetch.sh"), standalone);
  const { stdout } = await execFile("sh", [standalone, "install", "--target", f.target], { env });
  assert.match(stdout, /synchronized with the native API executor/);
  const installed = JSON.parse(await readFile(path.join(f.target, ".puretokens-executor/runtime.json"), "utf8"));
  assert.equal(installed.version, version);
  const requests = await readFile(path.join(f.root, "requests"), "utf8");
  if (mode === "platform") assert.doesNotMatch(requests, /codeload/);
  else assert.ok(requests.includes(`/zip/${commit}`));
  assert.deepEqual(await readdir(path.join(f.root, "tmp")), []);
});

test("published checksum mismatch stops before target mutation", async t => {
  const f = await fixture(t);
  const env = await mockDownloads(f, "bad-checksum");
  await assert.rejects(execFile("sh", [path.join(f.source, "runtime/puretokens-skill-fetch.sh"), "update", "--target", f.target], { env }), /checksum mismatch/);
  await assert.rejects(readFile(path.join(f.target, ".puretokens-executor/runtime.json")));
  assert.deepEqual(await readdir(path.join(f.root, "tmp")), []);
});

test("a release for a different source revision falls back to the pinned source", async t => {
  const f = await fixture(t);
  const env = await mockDownloads(f, "platform");
  const releaseFile = path.join(f.root, "release.json");
  const release = JSON.parse(await readFile(releaseFile, "utf8"));
  release.sourceCommit = "b".repeat(40);
  await writeFile(releaseFile, JSON.stringify(release, null, 2));
  await execFile("sh", [path.join(f.source, "runtime/puretokens-skill-fetch.sh"), "install", "--target", f.target], { env });
  assert.match(await readFile(path.join(f.root, "requests"), "utf8"), /codeload/);
});

test("legacy sibling installer is replaced by the pinned directory selector", async t => {
  const f = await fixture(t);
  const env = await mockDownloads(f);
  const bootstrap = path.join(f.root, "bootstrap");
  await mkdir(bootstrap);
  await cp(path.join(f.source, "runtime/puretokens-skill-fetch.sh"), path.join(bootstrap, "puretokens-skill-fetch.sh"));
  await writeFile(path.join(bootstrap, "puretokens-skill-install.sh"), "#!/bin/sh\nexit 79\n");
  const { stdout } = await execFile("sh", [path.join(bootstrap, "puretokens-skill-fetch.sh"), "check-update", "--target", f.target], { env });
  assert.match(stdout, /available=0.17.0/);
  assert.match(await readFile(path.join(f.root, "requests"), "utf8"), /runtime\/puretokens-skill-install\.sh/);
  assert.deepEqual(await readdir(path.join(f.root, "tmp")), []);
});

test("sync rejects an older download after a newer version wins the installation lock", async t => {
  const f = await fixture(t);
  await execFile("sh", [path.join(f.source, "runtime/puretokens-skill-install.sh"), "sync", "--target", f.target], { env: f.env });
  const newer = path.join(f.root, "newer-source");
  await cp(f.source, newer, { recursive: true });
  await writeFile(path.join(newer, "package.json"), JSON.stringify({ version: "0.18.0" }, null, 2));
  await writeFile(path.join(newer, "skills/puretokens-image/newer-marker"), "preserve newer installation");
  const env = { ...await mockDownloads(f), PT_INSTALL_DURING_DOWNLOAD: newer, PT_FIXTURE_TARGET: f.target };
  await assert.rejects(execFile("sh", [path.join(f.source, "runtime/puretokens-skill-fetch.sh"), "update", "--target", f.target], { env }), /downgrade was stopped under the update lock/);
  assert.equal(JSON.parse(await readFile(path.join(f.target, ".puretokens-executor/runtime.json"), "utf8")).version, "0.18.0");
  assert.equal(await readFile(path.join(f.target, "puretokens-image/newer-marker"), "utf8"), "preserve newer installation");
  assert.equal((await readdir(f.target)).some(file => file.includes("stage") || file.includes("lock")), false);
});

test("release provenance requires committed package and executor build inputs", async t => {
  const f = await fixture(t);
  const builder = path.join(f.source, "scripts/package-platform-releases.mjs");
  await cp(path.join(repositoryRoot, "scripts/package-platform-releases.mjs"), builder);
  await writeFile(path.join(f.source, "scripts/build-executor.mjs"), "// fixture build input\n");
  await writeFile(path.join(f.source, "runtime/executor/main.go"), "package main\n");
  const git = args => execFile("git", args, { cwd: f.source, env: f.env });
  await git(["init"]);
  await git(["add", "."]);
  await git(["-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-m", "fixture"]);
  const { stdout: committed } = await git(["rev-parse", "HEAD"]);
  const env = { ...f.env, GITHUB_SHA: committed.trim() };
  const manifestFile = path.join(f.source, "dist/releases/release-manifest.json");
  await execFile(process.execPath, [builder], { env });
  assert.equal(JSON.parse(await readFile(manifestFile, "utf8")).sourceCommit, committed.trim());
  for (const file of ["runtime/executor/main.go", "scripts/build-executor.mjs", "skills/puretokens-image/SKILL.md"]) {
    const before = await readFile(path.join(f.source, file), "utf8");
    await writeFile(path.join(f.source, file), `${before}\nmodified fixture\n`);
    await execFile(process.execPath, [builder], { env });
    assert.equal(JSON.parse(await readFile(manifestFile, "utf8")).sourceCommit, null, file);
    await writeFile(path.join(f.source, file), before);
  }
  await writeFile(path.join(f.source, "runtime/executor/untracked.go"), "package main\n");
  await execFile(process.execPath, [builder], { env });
  assert.equal(JSON.parse(await readFile(manifestFile, "utf8")).sourceCommit, null);
});
