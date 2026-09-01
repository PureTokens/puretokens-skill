#!/usr/bin/env node

import { randomUUID } from "node:crypto";
import { cp, mkdir, readFile, rename, rm, stat } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { collectSkillRecords, repositoryRoot, skillsRoot, validateRepository } from "../scripts/skill-registry.mjs";

const [command, ...argumentsList] = process.argv.slice(2);
const retiredSkillNames = [
  "puretokens_media",
  "puretokens_balance",
  "puretokens_connection",
  "puretokens_models",
  "puretokens_image",
  "puretokens_video",
  "puretokens_update",
  "puretokens_get_balance",
  "puretokens_get_model_price",
  "puretokens_workbuddy_router"
];

try {
  switch (command) {
    case "list":
      await listSkills();
      break;
    case "install":
      await installSkill(argumentsList);
      break;
    case "upgrade":
      await upgradeSkill(argumentsList);
      break;
    case "sync":
      await syncSkills(argumentsList);
      break;
    case "uninstall":
      await uninstallSkill(argumentsList);
      break;
    case "validate":
      await validate();
      break;
    case "help":
    case "--help":
    case "-h":
    case undefined:
      printHelp();
      break;
    default:
      throw new Error(`Unknown command: ${command}`);
  }
} catch (error) {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 1;
}

async function listSkills() {
  const { registry, records } = await collectSkillRecords();
  const byName = new Map(records.map((record) => [record.manifest.name, record]));
  for (const entry of registry.skills) {
    const record = byName.get(entry.name);
    if (!record) continue;
    process.stdout.write(`${entry.name}\t${entry.version}\t${entry.summary || record.manifest.description || ""}\n`);
  }
}

async function installSkill(args) {
  const options = parseSkillOptions(args, "install", { target: true });
  const source = await resolveSkillSource(options.name);
  const root = resolveInstallRoot(options.target);
  const destination = path.join(root, options.name);
  ensureChildPath(root, destination);
  if (await exists(destination)) {
    throw new Error(`Refusing to overwrite existing skill: ${destination}`);
  }
  await mkdir(root, { recursive: true });
  await cp(source, destination, { recursive: true, errorOnExist: true, force: false });
  process.stdout.write(`Installed ${options.name} to ${destination}\n`);
}

async function upgradeSkill(args) {
  const options = parseSkillOptions(args, "upgrade", { target: true });
  const source = await resolveSkillSource(options.name);
  const root = resolveInstallRoot(options.target);
  const destination = path.join(root, options.name);
  ensureChildPath(root, destination);
  if (!(await isManagedSkill(destination, options.name))) {
    throw new Error(`Cannot upgrade missing or unmanaged skill: ${destination}`);
  }

  await replaceManagedSkill(source, root, destination, options.name);
  process.stdout.write(`Upgraded ${options.name} at ${destination}\n`);
}

async function syncSkills(args) {
  const options = parseSyncOptions(args);
  const root = resolveInstallRoot(options.target);
  const { registry } = await collectSkillRecords();
  const errors = await validateRepository();
  if (errors.length) throw new Error(`Repository validation failed:\n${errors.map((error) => `- ${error}`).join("\n")}`);

  const plan = [];
  for (const retiredName of retiredSkillNames) {
    const destination = path.join(root, retiredName);
    ensureChildPath(root, destination);
    if (!(await exists(destination))) continue;
    if (!(await isManagedSkill(destination, retiredName))) {
      throw new Error(`Refusing to sync because an unmanaged retired Skill conflicts: ${destination}`);
    }
    plan.push({ action: "retire", name: retiredName, destination });
  }
  for (const entry of registry.skills) {
    const source = path.join(skillsRoot, entry.name);
    const destination = path.join(root, entry.name);
    ensureChildPath(root, destination);
    if (!(await exists(destination))) {
      plan.push({ action: "install", name: entry.name, source, destination });
    } else if (await isManagedSkill(destination, entry.name)) {
      plan.push({ action: "upgrade", name: entry.name, source, destination });
    } else {
      throw new Error(`Refusing to sync because an unmanaged Skill conflicts: ${destination}`);
    }
  }

  await mkdir(root, { recursive: true });
  for (const item of plan) {
    if (item.action === "retire") {
      const backup = await archiveRetiredSkill(root, item.destination, item.name);
      process.stdout.write(`Archived retired ${item.name} to ${backup}\n`);
    } else if (item.action === "install") {
      await cp(item.source, item.destination, { recursive: true, errorOnExist: true, force: false });
      process.stdout.write(`Installed ${item.name} to ${item.destination}\n`);
    } else {
      await replaceManagedSkill(item.source, root, item.destination, item.name);
      process.stdout.write(`Upgraded ${item.name} at ${item.destination}\n`);
    }
  }
}

async function archiveRetiredSkill(root, destination, name) {
  const backup = path.join(root, `.${name}.retired-${randomUUID()}`);
  ensureChildPath(root, backup);
  await rename(destination, backup);
  return backup;
}

async function uninstallSkill(args) {
  const options = parseSkillOptions(args, "uninstall", { target: true, yes: true });
  if (!options.yes) {
    throw new Error("Uninstall is destructive; pass --yes after checking the exact skill directory");
  }
  const root = resolveInstallRoot(options.target);
  const destination = path.join(root, options.name);
  ensureChildPath(root, destination);
  if (!(await isManagedSkill(destination, options.name))) {
    throw new Error(`Refusing to remove missing or unmanaged skill: ${destination}`);
  }
  await rm(destination, { recursive: true, force: false });
  process.stdout.write(`Uninstalled ${options.name} from ${destination}\n`);
}

async function validate() {
  const errors = await validateRepository();
  if (errors.length) throw new Error(errors.map((error) => `- ${error}`).join("\n"));
  process.stdout.write("Skill repository validation passed.\n");
}

async function resolveSkillSource(name) {
  if (!name || name.startsWith("-")) throw new Error("A skill name is required");
  const source = path.join(skillsRoot, name);
  ensureChildPath(skillsRoot, source);
  if (!(await existsDirectory(source))) throw new Error(`Unknown skill: ${name}`);
  const errors = await validateRepository();
  if (errors.length) throw new Error(`Repository validation failed:\n${errors.map((error) => `- ${error}`).join("\n")}`);
  return source;
}

function resolveInstallRoot(target) {
  if (!target) {
    throw new Error("A --target directory is required; this CLI never selects a client install location automatically");
  }
  return path.resolve(target);
}

function parseSkillOptions(args, commandName, allowed) {
  const name = args[0];
  if (!name || name.startsWith("-")) {
    throw new Error(`Usage: puretokens-skill ${commandName} <skill-name>${usageSuffix(allowed)}`);
  }
  const options = { name, target: undefined, yes: false };
  for (let index = 1; index < args.length; index += 1) {
    const flag = args[index];
    if (flag === "--target" && allowed.target) {
      options.target = requireOptionValue(args, index++, flag);
    } else if (flag === "--yes" && allowed.yes) {
      options.yes = true;
    } else {
      throw new Error(`Unknown option for ${commandName}: ${flag}`);
    }
  }
  return options;
}

function parseSyncOptions(args) {
  if (args.length !== 2 || args[0] !== "--target") {
    throw new Error("Usage: puretokens-skill sync --target <directory>");
  }
  if (!args[1] || args[1].startsWith("--")) throw new Error("--target requires a value");
  return { target: args[1] };
}

function requireOptionValue(args, index, flag) {
  const value = args[index + 1];
  if (!value || value.startsWith("--")) throw new Error(`${flag} requires a value`);
  return value;
}

function usageSuffix(allowed) {
  const flags = [];
  if (allowed.target) flags.push("[--target <directory>]");
  if (allowed.yes) flags.push("[--yes]");
  if (allowed.force) flags.push("[--force]");
  return flags.length ? ` ${flags.join(" ")}` : "";
}

async function isManagedSkill(directory, name) {
  if (!(await existsDirectory(directory))) return false;
  try {
    const manifest = JSON.parse(await readFile(path.join(directory, "skill.json"), "utf8"));
    return manifest?.name === name && (await exists(path.join(directory, "SKILL.md")));
  } catch {
    return false;
  }
}

async function replaceManagedSkill(source, root, destination, name) {
  const staging = path.join(root, `.${name}.upgrade-${randomUUID()}`);
  const backup = path.join(root, `.${name}.backup-${randomUUID()}`);
  await cp(source, staging, { recursive: true, errorOnExist: true, force: false });
  try {
    await rename(destination, backup);
    try {
      await rename(staging, destination);
    } catch (error) {
      await rename(backup, destination).catch(() => undefined);
      throw error;
    }
    await rm(backup, { recursive: true, force: false });
  } catch (error) {
    await rm(staging, { recursive: true, force: true }).catch(() => undefined);
    throw error;
  }
}

function ensureChildPath(root, candidate) {
  const resolvedRoot = path.resolve(root);
  const resolvedCandidate = path.resolve(candidate);
  if (!resolvedCandidate.startsWith(`${resolvedRoot}${path.sep}`)) throw new Error("Invalid target path");
}

async function exists(value) {
  try {
    await stat(value);
    return true;
  } catch {
    return false;
  }
}

async function existsDirectory(value) {
  try {
    return (await stat(value)).isDirectory();
  } catch {
    return false;
  }
}

function printHelp() {
  process.stdout.write(`Pure Tokens Skill Manager\n\nUsage:\n  puretokens-skill list\n  puretokens-skill validate\n  puretokens-skill install <skill-name> --target <directory>\n  puretokens-skill upgrade <skill-name> --target <directory>\n  puretokens-skill sync --target <directory>\n  puretokens-skill uninstall <skill-name> --target <directory> --yes\n\nSync archives verified retired managed Skills, installs missing official Skills, and upgrades only managed matching Skills. It refuses before changing anything when an unmanaged directory conflicts. Every install target is explicit. Use the seven supported host directories in README.md.\n\nRepository: ${repositoryRoot}\n`);
}
