// Maintainer-only, offline template preparation. Never installs a client,
// inspects connection files, makes API calls or marks an acceptance case passed.
import { readFile, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { acceptanceCases, acceptanceExecutionModes, acceptancePlatforms } from "./validate-host-acceptance.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

export function createAcceptanceEvidence({host, platform, architecture, executionMode = "local"}, {document, manifest, support}) {
  const target = document.hosts.find(item => item.host === host);
  const adapter = support.supported.find(item => item.id === host);
  if (!target || !adapter || !acceptancePlatforms.includes(platform) ||
      !["amd64", "arm64"].includes(architecture) || !acceptanceExecutionModes.includes(executionMode) ||
      (executionMode === "wsl" && platform !== "linux")) {
    throw new Error("Choose a registered host, OS, architecture and execution mode.");
  }
  const executorPlatform = `${platform === "macOS" ? "darwin" : platform}-${architecture}`;
  const executor = manifest.artifacts[executorPlatform];
  if (manifest.version !== document.version || !/^[a-f0-9]{64}$/.test(executor?.sha256)) {
    throw new Error("Acceptance release and executor manifest must match.");
  }
  const noTarget = target[platform]?.installation === "unavailable";
  return {
    method: "real-host", host, os: platform,
    hostVersion: "", osVersion: "", osArchitecture: architecture,
    executionMode, shellVersion: "", executorPlatform, executorSha256: executor.sha256,
    version: document.version, checkedAt: "", artifact: "",
    cases: Object.fromEntries(acceptanceCases.map(id => [id,
      noTarget || (adapter.credentialAdapter !== "fixture-tested" && !["installation", "help-no-api"].includes(id))
        ? "unavailable" : "pending"])),
    measurements: [],
  };
}

async function main() {
  const args = process.argv.slice(2), options = {};
  const allowed = new Set(["--host", "--os", "--arch", "--mode", "--output"]);
  for (let i = 0; i < args.length; i += 2) {
    if (!allowed.has(args[i]) || !args[i + 1] || args[i + 1].startsWith("--") || options[args[i]]) {
      throw new Error("Use --host ID --output FILE, optionally --os macOS|windows|linux --arch amd64|arm64 --mode local|wsl|remote|sandbox.");
    }
    options[args[i]] = args[i + 1];
  }
  if (!options["--output"]) throw new Error("Choose an explicit output file; existing files are never overwritten.");
  const read = async file => JSON.parse(await readFile(path.join(root, file), "utf8"));
  const [document, manifest, support] = await Promise.all([
    read("references/host-acceptance.json"), read("runtime/executor/manifest.json"), read("references/host-support.json"),
  ]);
  const evidence = createAcceptanceEvidence({
    host: options["--host"],
    platform: options["--os"] ?? ({darwin: "macOS", win32: "windows", linux: "linux"})[os.platform()],
    architecture: options["--arch"] ?? ({x64: "amd64", arm64: "arm64"})[os.arch()],
    executionMode: options["--mode"] ?? "local",
  }, {document, manifest, support});
  await writeFile(path.resolve(options["--output"]), JSON.stringify(evidence, null, 2) + "\n", {flag: "wx", mode: 0o600});
  console.log("Prepared an unverified evidence template. Fill actual identity, date and evidence location after testing; no client or API was invoked.");
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { await main(); }
  catch (error) {
    console.error(error.code === "EEXIST" ? "Output already exists; it was preserved." :
      error.code ? "Could not prepare the evidence file; no client or API was invoked." : error.message);
    process.exitCode = 1;
  }
}
