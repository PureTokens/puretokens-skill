import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

import { repositoryRoot } from "./skill-registry.mjs";

const catalogPath = path.join(
  repositoryRoot,
  "references",
  "media-model-catalog.json"
);

const readmes = [
  { file: "README.md", locale: "en" },
  { file: "README.zh-CN.md", locale: "zh" }
];

const startMarker = "<!-- media-model-catalog:start -->";
const endMarker = "<!-- media-model-catalog:end -->";
const supportedCapabilities = new Set(["image", "video"]);

function markdownCell(value) {
  return String(value).replaceAll("|", "\\|").replaceAll("\n", "<br>");
}

function inlineCode(value) {
  return `\`${value}\``;
}

function modelAliases(model, locale) {
  if (!model.aliases.length) return locale === "zh" ? "仅精确 ID" : "Exact ID only";
  return model.aliases.map(inlineCode).join(", ");
}

function modelRows(models, locale) {
  return models.map((model) => [
    inlineCode(model.id),
    markdownCell(model.provider),
    modelAliases(model, locale),
    markdownCell(model.goodFor[locale]),
    markdownCell(inlineCode(model.example[locale]))
  ].join(" | "));
}

function renderCatalog(catalog, locale) {
  const isChinese = locale === "zh";
  const images = catalog.models.filter((model) => model.capabilities.includes("image"));
  const videos = catalog.models.filter((model) => model.capabilities.includes("video"));
  const heading = isChinese ? "## 媒体模型清单" : "## Media model catalog";
  const published = isChinese
    ? `已与基础模型目录同步：${catalog.serviceCatalog.capturedAt}。`
    : `Synchronized with the base model catalog: ${catalog.serviceCatalog.capturedAt}.`;
  const liveNotice = catalog.availabilityNotice[locale];
  const newModelNotice = isChinese
    ? "README 只从基础目录中带有明确图片/视频能力的模型生成，不通过模型名称推断。已安装快照用于普通生成的模型选择和已知参数；实时目录只在明确查询、安装资料缺口或提交被拒后的诊断时按需读取。发布前从受控基础目录刷新，并运行 `npm run release:validate`；当快照超过七天时发布校验会失败。"
    : "README is generated only from base-catalog models with explicit image/video capabilities; it never infers capability from a model name. The installed snapshot resolves normal generation models and known parameters; the live catalog is read on demand only for explicit discovery, an installed-profile gap, or post-rejection diagnosis. Before release, refresh from the controlled base catalog and run `npm run release:validate`; the release gate fails when the snapshot is over seven days old.";
  const imageHeading = isChinese ? "### 图片模型" : "### Image models";
  const videoHeading = isChinese ? "### 视频模型" : "### Video models";
  const idHeader = isChinese ? "模型 ID" : "Model ID";
  const providerHeader = isChinese ? "提供方" : "Provider";
  const phraseHeader = isChinese ? "也可以这样说" : "You can also say";
  const goodForHeader = isChinese ? "适合" : "Good for";
  const exampleHeader = isChinese ? "示例" : "Example";
  const tableHeader = `| ${idHeader} | ${providerHeader} | ${phraseHeader} | ${goodForHeader} | ${exampleHeader} |`;
  const tableDivider = "| --- | --- | --- | --- | --- |";

  return [
    startMarker,
    heading,
    "",
    published,
    "",
    liveNotice,
    "",
    newModelNotice,
    "",
    imageHeading,
    "",
    tableHeader,
    tableDivider,
    ...modelRows(images, locale).map((row) => `| ${row} |`),
    "",
    videoHeading,
    "",
    tableHeader,
    tableDivider,
    ...modelRows(videos, locale).map((row) => `| ${row} |`),
    "",
    endMarker
  ].join("\n");
}

function validateCatalog(catalog) {
  if (catalog?.schemaVersion !== 2) throw new Error("published model catalog must use schemaVersion=2");
  if (!/^\d{4}-\d{2}-\d{2}$/.test(catalog.updatedAt || "")) {
    throw new Error("published model catalog must define updatedAt as YYYY-MM-DD");
  }
  if (!catalog.serviceCatalog?.source || !catalog.serviceCatalog?.path || !/^\d{4}-\d{2}-\d{2}T/.test(catalog.serviceCatalog.capturedAt || "")) {
    throw new Error("published model catalog must identify the supported media catalog snapshot");
  }
  if (!catalog.availabilityNotice?.en || !catalog.availabilityNotice?.zh) {
    throw new Error("published model catalog must include bilingual availability notices");
  }
  if (!Array.isArray(catalog.models) || !catalog.models.length) {
    throw new Error("published model catalog must include at least one model");
  }

  const ids = new Set();
  for (const model of catalog.models) {
    if (typeof model.id !== "string" || !model.id) throw new Error("published model catalog has a model without an ID");
    if (ids.has(model.id)) throw new Error(`published model catalog duplicates ${model.id}`);
    ids.add(model.id);
    if (!Array.isArray(model.capabilities) || !model.capabilities.length || !model.capabilities.every((capability) => supportedCapabilities.has(capability))) {
      throw new Error(`${model.id}: capabilities must contain image and/or video`);
    }
    if (model.name !== model.id || typeof model.provider !== "string" || !model.provider || !Number.isInteger(model.vendorId)) {
      throw new Error(`${model.id}: base catalog name, provider, and vendor ID are required`);
    }
    if (!Array.isArray(model.aliases) || !model.aliases.every((alias) => typeof alias === "string" && alias)) {
      throw new Error(`${model.id}: aliases must be an array of non-empty strings`);
    }
    for (const field of ["goodFor", "example"]) {
      if (!model[field]?.en || !model[field]?.zh) {
        throw new Error(`${model.id}: ${field} must contain English and Chinese copy`);
      }
    }
  }
  return ids;
}

function replaceCatalogBlock(readme, rendered, file) {
  const start = readme.indexOf(startMarker);
  const end = readme.indexOf(endMarker);
  if (start === -1 || end === -1 || end < start) {
    throw new Error(`${file}: missing media model catalog markers`);
  }
  if (readme.indexOf(startMarker, start + startMarker.length) !== -1 || readme.indexOf(endMarker, end + endMarker.length) !== -1) {
    throw new Error(`${file}: media model catalog markers must appear exactly once`);
  }
  return `${readme.slice(0, start)}${rendered}${readme.slice(end + endMarker.length)}`;
}

async function main() {
  const write = process.argv.includes("--write");
  const check = process.argv.includes("--check");
  if (write === check) throw new Error("use exactly one of --write or --check");

  const catalog = JSON.parse(await readFile(catalogPath, "utf8"));
  validateCatalog(catalog);
  let changed = false;
  for (const { file, locale } of readmes) {
    const readmePath = path.join(repositoryRoot, file);
    const original = await readFile(readmePath, "utf8");
    const updated = replaceCatalogBlock(original, renderCatalog(catalog, locale), file);
    if (original === updated) continue;
    changed = true;
    if (write) await writeFile(readmePath, updated);
    else process.stderr.write(`${file}: media model catalog is out of sync; run npm run docs:sync-media-models\n`);
  }
  if (check && changed) process.exitCode = 1;
}

await main();
