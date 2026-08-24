import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

import { repositoryRoot } from "./skill-registry.mjs";

const catalogPath = path.join(
  repositoryRoot,
  "references",
  "media-model-catalog.json"
);

const supportedCapabilities = new Set(["image", "video"]);
const imageEndpointTypes = new Set(["image-generation", "image_generation"]);
const videoEndpointTypes = new Set(["openai-video", "openai_video"]);

function parseArgs(argv) {
  const args = {
    input: "",
    sourceUrl: process.env.PURETOKENS_BASE_MODEL_CATALOG_URL || "",
    sourcePath: "/api/product/docs/model-catalog",
    write: false,
    check: false
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--write") args.write = true;
    else if (arg === "--check") args.check = true;
    else if (arg === "--input") args.input = argv[++index] || "";
    else if (arg === "--source-url") args.sourceUrl = argv[++index] || "";
    else if (arg === "--source-path") args.sourcePath = argv[++index] || "";
    else throw new Error(`unknown argument: ${arg}`);
  }
  if (args.write === args.check) throw new Error("use exactly one of --write or --check");
  if (args.input && args.sourceUrl) throw new Error("use either --input or --source-url, not both");
  if (!args.input && !args.sourceUrl) {
    throw new Error("set PURETOKENS_BASE_MODEL_CATALOG_URL or pass --source-url/--input");
  }
  return args;
}

async function readCatalogSource(args) {
  if (args.input) {
    const source = args.input === "-"
      ? await readStdin()
      : await readFile(path.resolve(args.input), "utf8");
    return JSON.parse(source);
  }

  const headers = { Accept: "application/json" };
  const token = process.env.PURETOKENS_BASE_MODEL_CATALOG_TOKEN;
  if (token) headers.Authorization = `Bearer ${token}`;
  const response = await fetch(args.sourceUrl, { headers });
  if (!response.ok) throw new Error(`base model catalog request failed with HTTP ${response.status}`);
  return response.json();
}

function readStdin() {
  return new Promise((resolve, reject) => {
    let source = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => { source += chunk; });
    process.stdin.on("end", () => resolve(source));
    process.stdin.on("error", reject);
  });
}

function text(value) {
  return typeof value === "string" ? value.trim() : String(value ?? "").trim();
}

function normalizedEndpointTypes(model) {
  const values = model?.supported_endpoint_types
    ?? model?.supportedEndpointTypes
    ?? model?.endpoint_types
    ?? model?.endpointTypes
    ?? [];
  const types = Array.isArray(values) ? values : values ? [values] : [];
  return new Set(types.map((value) => text(value).toLowerCase()).filter(Boolean));
}

function explicitCapabilities(model) {
  const capabilities = new Set();
  const declared = Array.isArray(model?.capabilities) ? model.capabilities : [];
  for (const value of declared) {
    const capability = text(value).toLowerCase();
    if (supportedCapabilities.has(capability)) capabilities.add(capability);
  }
  const endpointTypes = normalizedEndpointTypes(model);
  if ([...endpointTypes].some((type) => imageEndpointTypes.has(type))) capabilities.add("image");
  if ([...endpointTypes].some((type) => videoEndpointTypes.has(type))) capabilities.add("video");
  return [...capabilities].sort();
}

function sourceRows(payload) {
  if (!Array.isArray(payload?.data)) throw new Error("base model catalog response is invalid");
  const vendors = new Map(
    (Array.isArray(payload?.vendors) ? payload.vendors : [])
      .map((vendor) => [text(vendor?.id), text(vendor?.name)])
      .filter(([id, name]) => id && name)
  );
  return payload.data.map((row) => {
    const id = text(row?.model_name ?? row?.id);
    const vendorId = Number(row?.vendor_id ?? row?.vendorId);
    const provider = text(row?.provider ?? row?.owned_by ?? row?.ownedBy) || vendors.get(text(vendorId)) || "";
    return { id, provider, vendorId, capabilities: explicitCapabilities(row) };
  }).filter((model) => model.id && model.capabilities.length);
}

function genericCopy(id, capabilities) {
  const isVideo = capabilities.includes("video") && !capabilities.includes("image");
  return {
    aliases: [],
    goodFor: {
      en: isVideo ? "Video generation" : "Image generation",
      zh: isVideo ? "视频生成" : "图片生成"
    },
    example: {
      en: isVideo ? `Use ${id} to generate a short video.` : `Use ${id} to generate an image.`,
      zh: isVideo ? `用 ${id} 生成一条短视频。` : `用 ${id} 生成一张图片。`
    }
  };
}

function buildPublishedCatalog(previous, models, args) {
  const previousById = new Map((Array.isArray(previous?.models) ? previous.models : []).map((model) => [model.id, model]));
  const normalized = models.map((model) => {
    if (!model.provider || !Number.isInteger(model.vendorId)) {
      throw new Error(`${model.id}: base model catalog must provide provider and vendor ID`);
    }
    const editorial = previousById.get(model.id) || genericCopy(model.id, model.capabilities);
    return {
      id: model.id,
      name: model.id,
      provider: model.provider,
      vendorId: model.vendorId,
      capabilities: model.capabilities,
      aliases: editorial.aliases || [],
      goodFor: editorial.goodFor || genericCopy(model.id, model.capabilities).goodFor,
      example: editorial.example || genericCopy(model.id, model.capabilities).example
    };
  }).sort((left, right) => left.id.localeCompare(right.id, "en", { numeric: true }));
  const capturedAt = args.capturedAt || new Date().toISOString();
  return {
    schemaVersion: 2,
    updatedAt: capturedAt.slice(0, 10),
    serviceCatalog: {
      source: "Pure Tokens base model catalog",
      path: args.sourcePath || args.sourceUrl,
      capturedAt
    },
    availabilityNotice: {
      en: "This list is generated from Pure Tokens' base model catalog using explicit image/video capabilities. At execution time, the exact model and required capability must still appear in the current authenticated GET /v1/media/models response.",
      zh: "这份清单由 Pure Tokens 基础模型目录中的明确图片/视频能力生成。实际执行时，精确模型和所需能力仍必须出现在当前认证后的 GET /v1/media/models 响应中。"
    },
    models: normalized
  };
}

function serialized(value) {
  return JSON.stringify(value, null, 2) + "\n";
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const [previous, payload] = await Promise.all([
    readFile(catalogPath, "utf8").then(JSON.parse),
    readCatalogSource(args)
  ]);
  const next = buildPublishedCatalog(previous, sourceRows(payload), {
    ...args,
    capturedAt: previous?.serviceCatalog?.capturedAt
  });
  const currentText = serialized(previous);
  const nextText = serialized(next);
  if (currentText === nextText) return;
  if (args.check) {
    throw new Error("published media model catalog is out of sync with the base model catalog; run npm run docs:sync-media-models-from-service");
  }
  const refreshed = buildPublishedCatalog(previous, sourceRows(payload), {
    ...args,
    capturedAt: new Date().toISOString()
  });
  await writeFile(catalogPath, serialized(refreshed));
  process.stdout.write(`Published ${refreshed.models.length} base media models.\n`);
}

await main();
