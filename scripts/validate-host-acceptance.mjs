import { readFile, stat } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
export const acceptanceCases = ["installation", "authenticatedAPI", "image-generation", "uploaded-image-edit", "generated-image-edit", "multiple-images", "video-generation", "sameTaskResume", "nativeAttachmentDelivery"];
export const acceptancePlatforms = ["macOS", "windows", "linux"];
export const acceptanceExecutionModes = ["local", "wsl", "remote", "sandbox"];
const summaryCases = ["installation", "authenticatedAPI", "sameTaskResume", "nativeAttachmentDelivery"];
const states = new Set(["passed", "failed", "pending", "unavailable"]);

export function hostAcceptanceLevel(evidence) {
  if (evidence.cases?.installation !== "passed") return "unverified";
  if (evidence.cases?.authenticatedAPI !== "passed") return "installation-verified";
  if (acceptanceCases.every(id => evidence.cases?.[id] === "passed")) return "media-verified";
  return "api-verified";
}

export function validateHostAcceptance(document, version, manifest, hostIds) {
  const errors = [];
  if (document?.schemaVersion !== 1 || document.version !== version || !Array.isArray(document.hosts) ||
      !Array.isArray(document.evidence) || !/^\d{4}-\d{2}-\d{2}$/.test(document.checkedAt ?? "")) {
    return ["Host acceptance must identify the release/date, host matrix and real-host evidence array."];
  }
  const ids = document.hosts.map(host => host.host);
  if (new Set(ids).size !== ids.length || ids.length !== hostIds.length || hostIds.some(id => !ids.includes(id))) {
    errors.push("Host acceptance must cover each supported host exactly once.");
  }
  for (const evidence of document.evidence) {
    const artifact = manifest.artifacts?.[evidence.executorPlatform];
    if (!ids.includes(evidence.host) || !acceptancePlatforms.includes(evidence.os) ||
        evidence.method !== "real-host" || evidence.version !== version ||
        !artifact || evidence.executorSha256 !== artifact.sha256 ||
        !["hostVersion", "osVersion", "shellVersion", "artifact"].every(field => typeof evidence[field] === "string" && evidence[field].trim()) ||
        !["amd64", "arm64"].includes(evidence.osArchitecture) ||
        !acceptanceExecutionModes.includes(evidence.executionMode) ||
        !/^\d{4}-\d{2}-\d{2}$/.test(evidence.checkedAt ?? "") ||
        !evidence.cases || Object.keys(evidence.cases).some(id => !acceptanceCases.includes(id)) ||
        acceptanceCases.some(id => !states.has(evidence.cases[id]))) {
      errors.push("Real-host evidence requires release bytes, host/OS/shell versions, OS architecture, execution mode, date, artifact and every case outcome.");
    }
    if ((evidence.os === "macOS" && !evidence.executorPlatform?.startsWith("darwin-")) ||
        (evidence.os === "windows" && !evidence.executorPlatform?.startsWith("windows-")) ||
        (evidence.os === "linux" && !evidence.executorPlatform?.startsWith("linux-"))) {
      errors.push("Real-host evidence executor platform must match its OS.");
    }
    if (evidence.executionMode === "wsl" && evidence.os !== "linux") {
      errors.push("WSL evidence must identify its Linux execution environment.");
    }
  }
  for (const host of document.hosts) for (const os of acceptancePlatforms) for (const check of summaryCases) {
    const state = host[os]?.[check];
    if (state === "unavailable" && !host[os]?.reason?.trim()) errors.push(`${host.host}/${os} must explain why this release cannot accept that target.`);
    if (!states.has(state)) errors.push(`Invalid ${host.host}/${os}/${check} acceptance state.`);
    if (state === "passed" && !document.evidence.some(item => item.method === "real-host" && item.host === host.host && item.os === os && item.executionMode === "local" && item.cases?.[check] === "passed")) {
      errors.push(`${host.host}/${os}/${check} has no matching local real-host evidence.`);
    }
  }
  return errors;
}

export function validateStableHostAcceptance(document) {
  const errors = [];
  let mediaEnvironments = 0;
  for (const host of document.hosts ?? []) for (const os of acceptancePlatforms) {
    const target = host[os] ?? {};
    if (summaryCases.some(id => !["passed", "unavailable"].includes(target[id]))) {
      errors.push(`${host.host}/${os} has pending or failed real-host acceptance.`);
    }
    if (target.authenticatedAPI === "passed") {
      const complete = document.evidence?.some(item => item.method === "real-host" && item.host === host.host && item.os === os &&
        item.executionMode === "local" && hostAcceptanceLevel(item) === "media-verified");
      if (complete) mediaEnvironments++;
      else errors.push(`${host.host}/${os} has no complete local media acceptance.`);
    }
  }
  if (mediaEnvironments === 0) errors.push("Stable publication requires at least one complete real-host media environment.");
  return errors;
}

export async function checkHostAcceptance({stable = false} = {}) {
  const read = async file => JSON.parse(await readFile(path.join(root, file), "utf8"));
  const [document, pkg, manifest, support] = await Promise.all([
    read("references/host-acceptance.json"), read("package.json"),
    read("runtime/executor/manifest.json"), read("references/host-support.json"),
  ]);
  const errors = validateHostAcceptance(document, pkg.version, manifest, support.supported.map(host => host.id));
  if (stable) errors.push(...validateStableHostAcceptance(document));
  for (const evidence of document.evidence ?? []) {
    if (typeof evidence.artifact !== "string") continue;
    const artifact = path.resolve(root, evidence.artifact);
    if (!artifact.startsWith(root + path.sep) || !(await stat(artifact).catch(() => null))?.isFile()) {
      errors.push("Acceptance evidence must reference an existing file inside this repository.");
    }
  }
  if (errors.length) throw new Error(errors.join("\n"));
  const pending = document.hosts.reduce((count, host) => count + acceptancePlatforms.reduce((n, os) => n + summaryCases.filter(check => host[os][check] === "pending").length, 0), 0);
  const details = (document.evidence ?? []).flatMap(item => Object.values(item.cases ?? {}));
  console.log(`Host acceptance record is consistent; ${pending} host/OS summary checks remain pending. Detailed real-host case results recorded: ${details.length} (${details.filter(x => x === "passed").length} passed); ${acceptanceCases.length} distinct cases are required per accepted host/OS.`);
  const levels = ["installation-verified", "api-verified", "media-verified"];
  console.log("Evidence levels (each applies only to its recorded host/OS/architecture/mode): " +
    levels.map(level => `${level}=${document.evidence.filter(item => hostAcceptanceLevel(item) === level).length}`).join(", ") + ".");
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) await checkHostAcceptance({stable:process.argv.includes("--stable")});
