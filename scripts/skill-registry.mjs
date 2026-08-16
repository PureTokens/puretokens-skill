import { createHash } from "node:crypto";
import { readFile, readdir, stat } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

export const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
export const skillsRoot = path.join(repositoryRoot, "skills");

const forbiddenPattern = /(BEGIN [A-Z ]*PRIVATE|api[_-]?key|authorization:|bearer\s+|pts-router-token|127\.0\.0\.1:|\/Users\/)/i;

export async function readSkillRegistry() {
  const registryPath = path.join(skillsRoot, "index.json");
  const registry = JSON.parse(await readFile(registryPath, "utf8"));
  if (registry?.schemaVersion !== 1 || !Array.isArray(registry.skills)) {
    throw new Error("skills/index.json must contain schemaVersion=1 and a skills array");
  }
  return registry;
}

export async function collectSkillRecords() {
  const registry = await readSkillRegistry();
  const directoryEntries = await readdir(skillsRoot, { withFileTypes: true });
  const directories = directoryEntries.filter((entry) => entry.isDirectory()).map((entry) => entry.name).sort();
  const records = [];

  for (const directory of directories) {
    const skillDir = path.join(skillsRoot, directory);
    const [skillText, manifestText] = await Promise.all([
      readFile(path.join(skillDir, "SKILL.md"), "utf8"),
      readFile(path.join(skillDir, "skill.json"), "utf8")
    ]);
    const manifest = JSON.parse(manifestText);
    const frontmatter = readFrontmatter(skillText);
    records.push({ directory, skillDir, skillText, manifest, frontmatter });
  }

  return { registry, records };
}

export async function validateRepository() {
  const errors = [];
  const { registry, records } = await collectSkillRecords();
  const seen = new Set();
  const registryByName = new Map(registry.skills.map((skill) => [skill?.name, skill]));

  for (const record of records) {
    const { directory, skillText, manifest, frontmatter } = record;
    const name = manifest?.name;
    if (!/^[a-z][a-z0-9_]*$/.test(directory)) errors.push(`${directory}: directory must use snake_case`);
    if (!/^[a-z][a-z0-9_]*$/.test(name || "")) errors.push(`${directory}: manifest name must use snake_case`);
    if (name !== directory) errors.push(`${directory}: directory and manifest name must match`);
    if (frontmatter.name !== name) errors.push(`${directory}: SKILL.md frontmatter name must match manifest`);
    if (!frontmatter.description) errors.push(`${directory}: SKILL.md needs a non-empty frontmatter description`);
    if (!manifest?.version || !/^\d+\.\d+\.\d+$/.test(manifest.version)) errors.push(`${directory}: manifest version must be semver`);
    if (manifest?.sourceSha256 !== undefined) {
      const sourceSha256 = String(manifest.sourceSha256);
      const actualSourceSha256 = createHash("sha256").update(skillText).digest("hex");
      if (!/^[a-f0-9]{64}$/.test(sourceSha256)) {
        errors.push(`${directory}: sourceSha256 must be a lowercase SHA-256 digest`);
      } else if (sourceSha256 !== actualSourceSha256) {
        errors.push(`${directory}: sourceSha256 does not match SKILL.md`);
      }
    }
    if (seen.has(name)) errors.push(`${directory}: duplicate skill name ${name}`);
    seen.add(name);
    const registryEntry = registryByName.get(name);
    if (!registryEntry) errors.push(`${directory}: missing skills/index.json entry`);
    if (registryEntry && registryEntry.version !== manifest.version) {
      errors.push(`${directory}: manifest and skills/index.json versions differ`);
    }
    if (manifest?.behaviorTests) await verifyFile(errors, manifest.behaviorTests);
    if (manifest?.naturalLanguageAliases) await verifyFile(errors, manifest.naturalLanguageAliases);
    const claudeDesktop = manifest?.distribution?.claudeDesktop;
    if (claudeDesktop) {
      if (claudeDesktop.format !== "zip") errors.push(`${directory}: Claude Desktop distribution must use zip format`);
      if (claudeDesktop.archiveRoot !== name) errors.push(`${directory}: Claude Desktop archive root must match the skill name`);
      if (!Array.isArray(claudeDesktop.requiredFiles) || !claudeDesktop.requiredFiles.includes("SKILL.md")) {
        errors.push(`${directory}: Claude Desktop bundle must include SKILL.md`);
      }
      if (claudeDesktop.enableAfterImport !== true) errors.push(`${directory}: Claude Desktop bundle must require explicit enablement`);
    }
    const workbuddy = manifest?.distribution?.workbuddy;
    if (workbuddy) {
      if (workbuddy.managedByDesktop !== true) errors.push(`${directory}: WorkBuddy delivery must be managed by Desktop`);
      if (workbuddy.generatedSkillName !== "puretokens_workbuddy_router") {
        errors.push(`${directory}: WorkBuddy generated Skill name must be puretokens_workbuddy_router`);
      }
      if (workbuddy.alwaysApply !== true) errors.push(`${directory}: WorkBuddy delivery must be alwaysApply`);
    }
    if (forbiddenPattern.test(skillText) || forbiddenPattern.test(JSON.stringify(manifest))) {
      errors.push(`${directory}: skill content contains a forbidden credential or local-runtime marker`);
    }
  }

  for (const entry of registry.skills) {
    if (!entry || !seen.has(entry.name)) errors.push(`skills/index.json: ${entry?.name || "unnamed entry"} has no skill directory`);
    if (!entry?.entry || !entry?.manifest) errors.push(`skills/index.json: ${entry?.name || "unnamed entry"} needs entry and manifest`);
    if (entry?.entry) await verifyFile(errors, entry.entry);
    if (entry?.manifest) await verifyFile(errors, entry.manifest);
  }

  return errors;
}

function readFrontmatter(text) {
  const match = text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n/);
  if (!match) return {};
  const result = {};
  for (const line of match[1].split(/\r?\n/)) {
    const separator = line.indexOf(":");
    if (separator <= 0) continue;
    result[line.slice(0, separator).trim()] = line.slice(separator + 1).trim();
  }
  return result;
}

async function verifyFile(errors, relativePath) {
  const absolutePath = path.resolve(repositoryRoot, relativePath);
  if (!absolutePath.startsWith(`${repositoryRoot}${path.sep}`)) {
    errors.push(`skills/index.json: path escapes repository: ${relativePath}`);
    return;
  }
  try {
    if (!(await stat(absolutePath)).isFile()) errors.push(`skills/index.json: missing file ${relativePath}`);
  } catch {
    errors.push(`skills/index.json: missing file ${relativePath}`);
  }
}
