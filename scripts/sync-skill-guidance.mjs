import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { repositoryRoot } from "./skill-registry.mjs";

const bindingPattern = /^宿主绑定与本地失败停止：[^\n]+$/gm;

export async function syncSkillGuidance(root, { write = false } = {}) {
  const read = file => readFile(path.join(root, file), "utf8");
  const registry = JSON.parse(await read("skills/index.json"));
  const support = JSON.parse(await read("references/host-support.json"));
  const binding = (await read("references/skill-fragments/host-binding.md")).trim();
  if (!/^宿主绑定与本地失败停止：[^\n]+$/.test(binding)) throw new Error("Invalid host-binding fragment");
  const source = await read("references/desktop-hosts.md");
  if (source.split("<!-- host-binding -->").length !== 2) throw new Error("Desktop guidance needs one host-binding marker");
  const desktop = source.replace("<!-- host-binding -->", binding);
  const hostList = support.supported.map(host => host.id).join("、");
  const changed = [];
  async function output(file, next) {
    let current = "";
    try { current = await read(file); } catch (error) { if (error.code !== "ENOENT") throw error; }
    if (current === next) return;
    changed.push(file);
    if (write) await writeFile(path.join(root, file), next);
  }
  for (const skill of registry.skills) {
    const text = await read(skill.entry);
    if ([...text.matchAll(bindingPattern)].length !== 1) throw new Error(`${skill.name}: expected one host-binding paragraph`);
    let next = text.replace(bindingPattern, () => binding);
    if (["puretokens-image", "puretokens-video"].includes(skill.name)) {
      if (!/当前宿主 ID 为 [^。\n]+。/.test(next)) throw new Error(`${skill.name}: missing host list`);
      next = next.replace(/当前宿主 ID 为 [^。\n]+。/, `当前宿主 ID 为 ${hostList}。`);
    }
    if (skill.name === "puretokens-update") {
      if (!/^宿主 ID 为 [^；\n]+；/m.test(next)) throw new Error("Update: missing host list");
      next = next.replace(/^宿主 ID 为 [^；\n]+；/m, `宿主 ID 为 ${hostList}；`);
    }
    await output(skill.entry, next);
    await output(`skills/${skill.name}/references/desktop-hosts.md`, desktop);
    const manifest = JSON.parse(await read(skill.manifest));
    manifest.supportedClients = support.supported.map(host => host.id);
    for (const host of support.supported) {
      const key = host.id.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
      const delivery = manifest.distribution[key] ??= { manualInstallationSupported: true };
      if (host.globalSkillDirectory) delivery.globalSkillDirectory = host.globalSkillDirectory;
      else delete delivery.globalSkillDirectory;
      if (host.workspaceSkillDirectory) delivery.workspaceSkillDirectory = host.workspaceSkillDirectory;
      else delete delivery.workspaceSkillDirectory;
    }
    if (manifest.sourceSha256 !== undefined) manifest.sourceSha256 = createHash("sha256").update(next).digest("hex");
    await output(skill.manifest, JSON.stringify(manifest, null, 2) + "\n");
  }
  return changed;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const write = process.argv.includes("--write"), check = process.argv.includes("--check");
  if (write === check) throw new Error("Use exactly one of --write or --check");
  const changed = await syncSkillGuidance(repositoryRoot, { write });
  if (check && changed.length) {
    console.error(`Shared Skill guidance is out of sync; run npm run docs:sync-guidance:\n${changed.join("\n")}`);
    process.exitCode = 1;
  }
}
