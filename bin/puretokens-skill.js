#!/usr/bin/env node

import { randomUUID } from "node:crypto";
import { cp, mkdir, readFile, rename, rm, stat, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { collectSkillRecords, repositoryRoot, skillsRoot, validateRepository } from "../scripts/skill-registry.mjs";
import { getMediaSkillProvenance, mediaSkillSourceFiles } from "../scripts/media-skill-provenance.mjs";

const [command, ...argumentsList] = process.argv.slice(2);

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
    case "uninstall":
      await uninstallSkill(argumentsList);
      break;
    case "bundle":
      await bundleSkill(argumentsList);
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

  const staging = path.join(root, `.${options.name}.upgrade-${randomUUID()}`);
  const backup = path.join(root, `.${options.name}.backup-${randomUUID()}`);
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
  process.stdout.write(`Upgraded ${options.name} at ${destination}\n`);
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

async function bundleSkill(args) {
  const options = parseSkillOptions(args, "bundle", { format: true, out: true, force: true });
  if (options.format !== "claude-desktop") {
    throw new Error("Only --format claude-desktop is supported");
  }
  const source = await resolveSkillSource(options.name);
  const manifest = JSON.parse(await readFile(path.join(source, "skill.json"), "utf8"));
  const files = mediaSkillSourceFiles.filter((relativePath) => (
    relativePath !== "agents/openai.yaml" && !relativePath.startsWith("adapters/")
  ));
  const entries = [];
  for (const relativePath of files) {
    const absolutePath = path.join(source, relativePath);
    if (!(await exists(absolutePath))) throw new Error(`Skill bundle file is missing: ${relativePath}`);
    entries.push({
      path: relativePath,
      data: relativePath === "skill.json"
        ? `${JSON.stringify(createClaudeDesktopManifest(manifest), null, 2)}\n`
        : await readFile(absolutePath)
    });
  }
  entries.push({
    path: "source-delivery.json",
    data: `${JSON.stringify({ delivery: "claude-desktop", derivedFrom: await getMediaSkillProvenance() }, null, 2)}\n`
  });
  const output = path.resolve(options.out || `${options.name}-${manifest.version}-claude-desktop.zip`);
  if (await exists(output) && !options.force) {
    throw new Error(`Refusing to overwrite existing bundle: ${output} (pass --force to replace it)`);
  }
  await mkdir(path.dirname(output), { recursive: true });
  await writeFile(output, createZip(entries));
  process.stdout.write(`Created Claude Desktop Skill bundle at ${output}\n`);
}

function createClaudeDesktopManifest(sourceManifest) {
  const { workBuddyAdapter, distribution, supportedClients, ...sharedManifest } = sourceManifest;
  const { claudeDesktop } = distribution;
  return {
    ...sharedManifest,
    distribution: { claudeDesktop },
    supportedClients: supportedClients.filter((client) => client !== "codex" && client !== "workbuddy")
  };
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
  const options = { name, target: undefined, out: undefined, format: "claude-desktop", yes: false, force: false };
  for (let index = 1; index < args.length; index += 1) {
    const flag = args[index];
    if (flag === "--target" && allowed.target) {
      options.target = requireOptionValue(args, index++, flag);
    } else if (flag === "--out" && allowed.out) {
      options.out = requireOptionValue(args, index++, flag);
    } else if (flag === "--format" && allowed.format) {
      options.format = requireOptionValue(args, index++, flag);
    } else if (flag === "--yes" && allowed.yes) {
      options.yes = true;
    } else if (flag === "--force" && allowed.force) {
      options.force = true;
    } else {
      throw new Error(`Unknown option for ${commandName}: ${flag}`);
    }
  }
  return options;
}

function requireOptionValue(args, index, flag) {
  const value = args[index + 1];
  if (!value || value.startsWith("--")) throw new Error(`${flag} requires a value`);
  return value;
}

function usageSuffix(allowed) {
  const flags = [];
  if (allowed.target) flags.push("[--target <directory>]");
  if (allowed.format) flags.push("[--format claude-desktop]");
  if (allowed.out) flags.push("[--out <zip-file>]");
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

function createZip(entries) {
  const localParts = [];
  const centralParts = [];
  let offset = 0;
  for (const entry of entries) {
    const name = Buffer.from(`puretokens_media/${entry.path}`, "utf8");
    const data = Buffer.isBuffer(entry.data) ? entry.data : Buffer.from(entry.data);
    const crc = crc32(data);
    const localHeader = Buffer.alloc(30);
    localHeader.writeUInt32LE(0x04034b50, 0);
    localHeader.writeUInt16LE(20, 4);
    localHeader.writeUInt16LE(0, 6);
    localHeader.writeUInt16LE(0, 8);
    localHeader.writeUInt16LE(0, 10);
    localHeader.writeUInt16LE(33, 12);
    localHeader.writeUInt32LE(crc, 14);
    localHeader.writeUInt32LE(data.length, 18);
    localHeader.writeUInt32LE(data.length, 22);
    localHeader.writeUInt16LE(name.length, 26);
    localHeader.writeUInt16LE(0, 28);
    localParts.push(localHeader, name, data);

    const centralHeader = Buffer.alloc(46);
    centralHeader.writeUInt32LE(0x02014b50, 0);
    centralHeader.writeUInt16LE(20, 4);
    centralHeader.writeUInt16LE(20, 6);
    centralHeader.writeUInt16LE(0, 8);
    centralHeader.writeUInt16LE(0, 10);
    centralHeader.writeUInt16LE(0, 12);
    centralHeader.writeUInt16LE(33, 14);
    centralHeader.writeUInt32LE(crc, 16);
    centralHeader.writeUInt32LE(data.length, 20);
    centralHeader.writeUInt32LE(data.length, 24);
    centralHeader.writeUInt16LE(name.length, 28);
    centralHeader.writeUInt16LE(0, 30);
    centralHeader.writeUInt16LE(0, 32);
    centralHeader.writeUInt16LE(0, 34);
    centralHeader.writeUInt16LE(0, 36);
    centralHeader.writeUInt32LE(0, 38);
    centralHeader.writeUInt32LE(offset, 42);
    centralParts.push(centralHeader, name);
    offset += localHeader.length + name.length + data.length;
  }
  const centralDirectory = Buffer.concat(centralParts);
  const end = Buffer.alloc(22);
  end.writeUInt32LE(0x06054b50, 0);
  end.writeUInt16LE(0, 4);
  end.writeUInt16LE(0, 6);
  end.writeUInt16LE(entries.length, 8);
  end.writeUInt16LE(entries.length, 10);
  end.writeUInt32LE(centralDirectory.length, 12);
  end.writeUInt32LE(offset, 16);
  end.writeUInt16LE(0, 20);
  return Buffer.concat([...localParts, centralDirectory, end]);
}

function crc32(buffer) {
  let crc = 0xffffffff;
  for (const byte of buffer) {
    crc ^= byte;
    for (let bit = 0; bit < 8; bit += 1) crc = (crc >>> 1) ^ (crc & 1 ? 0xedb88320 : 0);
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function printHelp() {
  process.stdout.write(`Pure Tokens Skill Manager\n\nUsage:\n  puretokens-skill list\n  puretokens-skill validate\n  puretokens-skill install <skill-name> --target <directory>\n  puretokens-skill upgrade <skill-name> --target <directory>\n  puretokens-skill uninstall <skill-name> --target <directory> --yes\n  puretokens-skill bundle <skill-name> --format claude-desktop [--out <zip-file>] [--force]\n\nEvery install target is explicit. Pure Tokens Desktop can also apply managed Codex and WorkBuddy deliveries; Claude Desktop requires the generated ZIP to be uploaded and enabled in its Skills settings.\n\nRepository: ${repositoryRoot}\n`);
}
