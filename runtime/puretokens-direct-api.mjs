#!/usr/bin/env node

import { createWriteStream, openAsBlob } from "node:fs";
import { lstat, mkdir, readFile, stat, unlink } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import process from "node:process";
import { Readable, Transform } from "node:stream";
import { pipeline } from "node:stream/promises";

const apiOrigin = "https://api.puretokensx.com";
const pureTokensApiBaseUrl = `${apiOrigin}/v1`;
const managedRuntimeHosts = ["claude-code", "codex", "workbuddy", "gemini-cli", "grok-build", "opencode"];
const claudeSettingsPath = path.join(os.homedir(), ".claude", "settings.json");
const codexConfigPath = path.join(os.homedir(), ".codex", "config.toml");
const codexAuthPath = path.join(os.homedir(), ".codex", "auth.json");
const grokConfigPath = path.join(os.homedir(), ".grok", "config.toml");
const geminiEnvPath = path.join(os.homedir(), ".gemini", ".env");
const geminiSettingsPath = path.join(os.homedir(), ".gemini", "settings.json");
const openCodeConfigPath = path.join(os.homedir(), ".config", "opencode", "opencode.json");
const workBuddyModelConfigPath = path.join(os.homedir(), ".workbuddy", "models.json");
const maxMultipartFiles = 8;
const maxMultipartFileBytes = 512 * 1024 * 1024;
const maxJsonBodyBytes = 1024 * 1024;
const maxImageMediaBytes = 128 * 1024 * 1024;
const maxVideoOrAudioMediaBytes = 2 * 1024 * 1024 * 1024;
export const directRequestDeadlineMs = 90_000;
const allowedMultipartContentTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
  "image/avif",
  "video/mp4",
  "video/quicktime",
  "video/webm",
  "audio/mpeg",
  "audio/mp4",
  "audio/wav",
  "audio/x-wav",
  "audio/ogg",
  "audio/webm"
]);

export async function main(argv = process.argv.slice(2), environment = process.env) {
  const deadline = AbortSignal.timeout(directRequestDeadlineMs);
  let phase = "validation";
  try {
    const options = parseArguments(argv);
    const credential = await withinDeadline(resolveConfiguredCredential(options.host, environment), deadline);
    const target = new URL(options.path, apiOrigin);
    const headers = { authorization: `Bearer ${credential}` };
    const request = { method: options.method, headers, signal: deadline };

    if (options.jsonBase64 !== undefined) {
      headers["content-type"] = "application/json";
      request.body = JSON.stringify(readJsonBase64Argument(options.jsonBase64));
    }
    if (options.multipartBase64 !== undefined) {
      request.body = await withinDeadline(createMultipartRequestBody(readJsonBase64Argument(options.multipartBase64)), deadline);
    }

    phase = options.method === "POST" ? "submission" : "read";
    const response = await withinDeadline(fetch(target, request), deadline);
    phase = "response";
    if (options.outputFile) {
      const delivered = await writeNativeResponse(response, options.outputFile, deadline);
      if (!delivered) return;
      writeJson({
        http_status: response.status,
        content_type: response.headers.get("content-type") || null,
        retry_after: response.headers.get("retry-after") || null,
        output_file: options.outputFile
      });
      return;
    }

    const responseText = await withinDeadline(response.text(), deadline);
    let body = responseText;
    if (responseText) {
      try {
        body = JSON.parse(responseText);
      } catch {
        // A non-JSON body is still surfaced to the calling Skill as an opaque API response.
      }
    } else {
      body = null;
    }
    writeJson({
      http_status: response.status,
      content_type: response.headers.get("content-type") || null,
      retry_after: response.headers.get("retry-after") || null,
      body
    });
  } catch (error) {
    throw new DirectRuntimeFailure(phase, error);
  }
}

export async function resolveConfiguredCredential(host, environment = process.env) {
  if (host === "claude-code") return resolveClaudeCodeCredential();
  if (host === "codex") return resolveCodexCredential();
  if (host === "workbuddy") return resolveWorkBuddyCredential();
  if (host === "gemini-cli") return resolveGeminiCliCredential();
  if (host === "grok-build") return resolveGrokBuildCredential(environment);
  if (host === "opencode") return resolveOpenCodeCredential();
  if (host === "trae") throw new Error("Trae uses manual connection setup and has no approved local credential resolver for this Skill.");
  throw new Error("The requested host is not supported by the Pure Tokens direct runtime.");
}

export async function resolveGrokBuildCredential(environment = process.env, configPath = grokConfigPath) {
  const config = await readConfigurationText(configPath, "Grok Build");
  const entries = parseGrokModelEntries(config, environment);
  return selectUnambiguousCredential(entries, isPureTokensApiEndpoint, "Grok Build");
}

export async function resolveWorkBuddyCredential(configPath = workBuddyModelConfigPath) {
  const entries = parseWorkBuddyModelEntries(await readJsonConfiguration(configPath, "WorkBuddy"));
  return selectUnambiguousCredential(entries, isWorkBuddyPureTokensApiEndpoint, "WorkBuddy");
}

export async function resolveClaudeCodeCredential(settingsPath = claudeSettingsPath) {
  const settings = await readJsonConfiguration(settingsPath, "Claude Code");
  const environment = isPlainObject(settings?.env) ? settings.env : undefined;
  return selectUnambiguousCredential([{
    baseUrl: typeof environment?.ANTHROPIC_BASE_URL === "string" ? environment.ANTHROPIC_BASE_URL : undefined,
    credential: typeof environment?.ANTHROPIC_AUTH_TOKEN === "string" ? environment.ANTHROPIC_AUTH_TOKEN : undefined
  }], isPureTokensOriginEndpoint, "Claude Code");
}

export async function resolveCodexCredential(configPath = codexConfigPath, authPath = codexAuthPath) {
  const connection = parseCodexConnection(await readConfigurationText(configPath, "Codex"));
  // This is credential-source eligibility only. main() always sends the API request to apiOrigin.
  if (!isPureTokensApiEndpoint(connection.baseUrl)) {
    throw new Error("No usable Pure Tokens API credential is configured for Codex.");
  }
  if (isUsableCredential(connection.credential)) return connection.credential;
  const auth = await readJsonConfiguration(authPath, "Codex authentication");
  const credential = typeof auth?.OPENAI_API_KEY === "string" ? auth.OPENAI_API_KEY : undefined;
  if (!isUsableCredential(credential)) {
    throw new Error("No usable Pure Tokens API credential is configured for Codex.");
  }
  return credential;
}

export async function resolveGeminiCliCredential(envPath = geminiEnvPath, settingsPath = geminiSettingsPath) {
  const bindings = parseDotEnv(await readConfigurationText(envPath, "Gemini CLI"));
  const settings = await readJsonConfiguration(settingsPath, "Gemini CLI");
  const selectedType = settings?.security?.auth?.selectedType;
  if (selectedType !== "gemini-api-key") {
    throw new Error("No usable Pure Tokens API-key connection is selected for Gemini CLI.");
  }
  return selectUnambiguousCredential([{
    baseUrl: bindings.get("GOOGLE_GEMINI_BASE_URL"),
    credential: bindings.get("GEMINI_API_KEY")
  }], isPureTokensOriginEndpoint, "Gemini CLI");
}

export async function resolveOpenCodeCredential(configPath = openCodeConfigPath) {
  const configuration = await readJsonConfiguration(configPath, "OpenCode");
  const options = configuration?.provider?.puretokens?.options;
  return selectUnambiguousCredential([{
    baseUrl: typeof options?.baseURL === "string" ? options.baseURL : undefined,
    credential: typeof options?.apiKey === "string" ? options.apiKey : undefined
  }], isPureTokensApiEndpoint, "OpenCode");
}

export function parseGrokModelEntries(config, environment = process.env) {
  const entries = [];
  let entry;
  for (const line of config.split(/\r?\n/)) {
    if (/^\s*\[model\./.test(line)) {
      entry = { baseUrl: undefined, credential: undefined };
      entries.push(entry);
      continue;
    }
    if (/^\s*\[/.test(line)) {
      entry = undefined;
      continue;
    }
    if (!entry) continue;
    const assignment = line.match(/^\s*([A-Za-z0-9_-]+)\s*=\s*(.+?)\s*(?:#.*)?$/);
    if (!assignment) continue;
    const [, key, rawValue] = assignment;
    if (key === "base_url") entry.baseUrl = parseTomlString(rawValue);
    if (key === "api_key") entry.credential = parseTomlString(rawValue);
    if (key === "env_key" && !entry.credential) entry.credential = resolveEnvironmentCredential(rawValue, environment);
  }
  return entries;
}

export function parseWorkBuddyModelEntries(value) {
  if (!Array.isArray(value)) return [];
  return value.map((model) => ({
    baseUrl: typeof model?.url === "string" ? model.url : undefined,
    credential: typeof model?.apiKey === "string" ? model.apiKey : undefined
  }));
}

export function parseCodexConnection(config) {
  let activeProvider;
  let currentProvider;
  const providers = new Map();
  for (const line of config.split(/\r?\n/)) {
    const header = line.match(/^\s*\[model_providers\.([A-Za-z0-9_-]+|"(?:[^"\\]|\\.)*"|'[^']*')\]\s*(?:#.*)?$/);
    if (header) {
      currentProvider = parseTomlString(header[1]) ?? header[1];
      if (!providers.has(currentProvider)) providers.set(currentProvider, { baseUrl: undefined, credential: undefined });
      continue;
    }
    if (/^\s*\[/.test(line)) {
      currentProvider = undefined;
      continue;
    }
    const assignment = parseTomlAssignment(line);
    if (!assignment) continue;
    const { key, value } = assignment;
    if (!currentProvider && key === "model_provider") activeProvider = value;
    if (currentProvider && key === "base_url") providers.get(currentProvider).baseUrl = value;
    if (currentProvider && key === "experimental_bearer_token") providers.get(currentProvider).credential = value;
  }
  const active = typeof activeProvider === "string" ? providers.get(activeProvider) : undefined;
  return active ? { baseUrl: active.baseUrl, credential: active.credential } : { baseUrl: undefined, credential: undefined };
}

export function parseDotEnv(source) {
  const bindings = new Map();
  for (const rawLine of source.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;
    const assignment = line.match(/^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
    if (!assignment) continue;
    const [, key, rawValue] = assignment;
    const value = parseEnvironmentValue(rawValue);
    if (value !== undefined) bindings.set(key, value);
  }
  return bindings;
}

export function parseArguments(argv) {
  const args = [...argv];
  if (args.shift() !== "request") throw new Error(`Usage: puretokens-direct-api request --host <${managedRuntimeHosts.join("|")}> --method <GET|POST> --path <fixed-api-path> [--json-base64 <base64>|--multipart-base64 <base64>] [--output-file <absolute-path>]`);
  const options = {
    host: undefined,
    method: undefined,
    path: undefined,
    jsonBase64: undefined,
    multipartBase64: undefined,
    outputFile: undefined
  };
  while (args.length) {
    const flag = args.shift();
    if (flag === "--json-base64") {
      options.jsonBase64 = requiredArgumentValue(args, flag);
      continue;
    }
    if (flag === "--multipart-base64") {
      options.multipartBase64 = requiredArgumentValue(args, flag);
      continue;
    }
    if (flag === "--json-stdin" || flag === "--multipart-stdin") {
      throw new Error("Standard-input request bodies are not supported; use --json-base64 or --multipart-base64.");
    }
    if (!["--host", "--method", "--path", "--output-file"].includes(flag)) throw new Error(`Unknown option: ${flag}`);
    const value = requiredArgumentValue(args, flag);
    if (flag === "--host") options.host = value;
    if (flag === "--method") options.method = value.toUpperCase();
    if (flag === "--path") options.path = value;
    if (flag === "--output-file") options.outputFile = value;
  }
  if (!managedRuntimeHosts.includes(options.host)) throw new Error("The requested host does not have a managed Pure Tokens credential runtime.");
  if (!["GET", "POST"].includes(options.method)) throw new Error("--method must be GET or POST");
  if (!isAllowedPath(options.method, options.path)) throw new Error("The requested API path is not allowed by the Pure Tokens direct API contract.");
  const requestBodyModeCount = Number(options.jsonBase64 !== undefined) + Number(options.multipartBase64 !== undefined);
  if (requestBodyModeCount !== (options.method === "POST" ? 1 : 0)) {
    throw new Error(options.method === "POST"
      ? "POST requests require exactly one JSON or multipart request-body mode."
      : "GET requests cannot use a request body.");
  }
  if (options.outputFile && !path.isAbsolute(options.outputFile)) throw new Error("--output-file must be an absolute path.");
  if (options.outputFile && (options.method !== "GET" || !isContentPath(options.path))) {
    throw new Error("--output-file is allowed only for a fixed completed media content path.");
  }
  return options;
}

function requiredArgumentValue(args, flag) {
  const value = args.shift();
  if (!value || value.startsWith("--")) throw new Error(`${flag} requires a value`);
  return value;
}

function isAllowedPath(method, value) {
  if (typeof value !== "string" || !value.startsWith("/") || value.includes("://") || value.includes("#")) return false;
  if (method === "POST") return ["/v1/images/generations", "/v1/images/edits", "/v1/videos", "/v1/videos/edits"].includes(value);
  if (["/v1", "/v1/media/models", "/api/product/desktop/account/balance"].includes(value)) return true;
  const taskSegment = "(?:[A-Za-z0-9._~-]|%[0-9A-Fa-f]{2})+";
  return new RegExp(`^/v1/images/${taskSegment}$`).test(value) ||
    new RegExp(`^/v1/images/${taskSegment}/content\\?index=[0-9]+$`).test(value) ||
    new RegExp(`^/v1/videos/${taskSegment}$`).test(value) ||
    new RegExp(`^/v1/videos/${taskSegment}/content$`).test(value);
}

function isContentPath(value) {
  const taskSegment = "(?:[A-Za-z0-9._~-]|%[0-9A-Fa-f]{2})+";
  return new RegExp(`^/v1/images/${taskSegment}/content\\?index=[0-9]+$`).test(value) ||
    new RegExp(`^/v1/videos/${taskSegment}/content$`).test(value);
}

function isPureTokensApiEndpoint(baseUrl) {
  try {
    const parsed = new URL(baseUrl);
    return parsed.href === pureTokensApiBaseUrl || parsed.href === `${pureTokensApiBaseUrl}/`;
  } catch {
    return false;
  }
}

function isWorkBuddyPureTokensApiEndpoint(baseUrl) {
  try {
    const parsed = new URL(baseUrl);
    return parsed.origin === apiOrigin && !parsed.search && !parsed.hash &&
      (parsed.pathname === "/v1" || parsed.pathname === "/v1/" || parsed.pathname.startsWith("/v1/"));
  } catch {
    return false;
  }
}

function isPureTokensOriginEndpoint(baseUrl) {
  try {
    const parsed = new URL(baseUrl);
    return parsed.href === `${apiOrigin}/`;
  } catch {
    return false;
  }
}

function selectUnambiguousCredential(entries, endpointMatcher, host) {
  const credentials = [...new Set(entries
    .filter((entry) => endpointMatcher(entry?.baseUrl))
    .map((entry) => entry?.credential)
    .filter(isUsableCredential))];
  if (credentials.length === 0) {
    throw new Error(`No usable Pure Tokens API credential is configured for ${host}.`);
  }
  if (credentials.length > 1) {
    throw new Error(`Multiple different Pure Tokens API credentials are configured for ${host}; select one connection before using this Skill.`);
  }
  return credentials[0];
}

function isUsableCredential(value) {
  return typeof value === "string" && value.trim().length > 0;
}

async function readConfigurationText(configPath, label) {
  try {
    return await readFile(configPath, "utf8");
  } catch {
    throw new Error(`${label} connection configuration is unavailable.`);
  }
}

async function readJsonConfiguration(configPath, label) {
  const source = await readConfigurationText(configPath, label);
  try {
    return JSON.parse(source);
  } catch {
    throw new Error(`${label} connection configuration is invalid.`);
  }
}

function parseTomlString(rawValue) {
  const value = rawValue.trim();
  if (value.startsWith('"') && value.endsWith('"')) {
    try {
      return JSON.parse(value);
    } catch {
      return undefined;
    }
  }
  if (value.startsWith("'") && value.endsWith("'")) return value.slice(1, -1);
  return undefined;
}

function parseTomlAssignment(line) {
  const assignment = line.match(/^\s*([A-Za-z0-9_-]+)\s*=\s*(.+?)\s*(?:#.*)?$/);
  if (!assignment) return undefined;
  const value = parseTomlString(assignment[2]);
  return value === undefined ? undefined : { key: assignment[1], value };
}

function parseEnvironmentValue(rawValue) {
  const value = rawValue.trim();
  if (!value) return "";
  if (value.startsWith('"') && value.endsWith('"')) {
    try {
      return JSON.parse(value);
    } catch {
      return undefined;
    }
  }
  if (value.startsWith("'") && value.endsWith("'")) return value.slice(1, -1);
  return value.replace(/\s+#.*$/, "").trim();
}

function resolveEnvironmentCredential(rawValue, environment) {
  const value = rawValue.trim();
  const names = value.startsWith("[") && value.endsWith("]")
    ? [...value.matchAll(/(?:"((?:[^"\\]|\\.)*)"|'([^']*)')/g)].map((match) => match[1] ? JSON.parse(`"${match[1]}"`) : match[2])
    : [parseTomlString(value)];
  return names.map((name) => environment[name]).find((candidate) => typeof candidate === "string" && candidate.length > 0);
}

export function readJsonBase64Argument(value) {
  if (typeof value !== "string" || !value || value.length > encodedBodyLengthLimit() || value.length % 4 !== 0 || !/^[A-Za-z0-9+/]*={0,2}$/.test(value)) {
    throw new Error("Base64 request body must be one bounded canonical UTF-8 JSON value.");
  }
  const bytes = Buffer.from(value, "base64");
  if (!bytes.length || bytes.length > maxJsonBodyBytes || bytes.toString("base64") !== value) {
    throw new Error("Base64 request body must be one bounded canonical UTF-8 JSON value.");
  }
  let json;
  try {
    json = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  } catch {
    throw new Error("Base64 request body must be valid UTF-8 JSON.");
  }
  try {
    return JSON.parse(json);
  } catch {
    throw new Error("Base64 request body must contain one valid JSON value.");
  }
}

function encodedBodyLengthLimit() {
  return Math.ceil(maxJsonBodyBytes / 3) * 4;
}

async function createMultipartRequestBody(specification) {
  if (!specification || typeof specification !== "object" || Array.isArray(specification)) {
    throw new Error("Multipart request body must contain one object.");
  }
  const fields = specification.fields ?? {};
  const files = specification.files ?? [];
  if (!isPlainObject(fields) || !Array.isArray(files) || files.length === 0 || files.length > maxMultipartFiles) {
    throw new Error("Multipart request body must contain up to eight explicit files and an optional fields object.");
  }
  const form = new FormData();
  for (const [field, value] of Object.entries(fields)) {
    if (!isSafeFormField(field) || !isScalarFormValue(value)) throw new Error("Multipart fields must have safe names and scalar values.");
    form.append(field, String(value));
  }
  for (const file of files) {
    if (!isPlainObject(file) || !isSafeFormField(file.field) || typeof file.path !== "string" || !path.isAbsolute(file.path)) {
      throw new Error("Each multipart file requires a safe field name and an absolute explicit attachment path.");
    }
    const metadata = await readExplicitMediaFile(file.path, file.contentType);
    const filename = typeof file.filename === "string" && isSafeFilename(file.filename)
      ? file.filename
      : path.basename(file.path);
    form.append(file.field, metadata.blob, filename);
  }
  return form;
}

async function readExplicitMediaFile(filePath, contentType) {
  const fileStatus = await lstat(filePath);
  if (!fileStatus.isFile() || fileStatus.isSymbolicLink()) {
    throw new Error("Multipart input must reference a current explicit regular media attachment, not a directory or symbolic link.");
  }
  if (fileStatus.size <= 0 || fileStatus.size > maxMultipartFileBytes) {
    throw new Error("An explicit media attachment must be larger than zero and no more than 512 MiB.");
  }
  const resolvedContentType = resolveMultipartContentType(filePath, contentType);
  return { blob: await openAsBlob(filePath, { type: resolvedContentType }), contentType: resolvedContentType };
}

function resolveMultipartContentType(filePath, value) {
  if (typeof value === "string" && allowedMultipartContentTypes.has(value)) return value;
  const extension = path.extname(filePath).toLowerCase();
  const inferred = {
    ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".webp": "image/webp", ".gif": "image/gif", ".avif": "image/avif",
    ".mp4": "video/mp4", ".mov": "video/quicktime", ".webm": "video/webm",
    ".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".wav": "audio/wav", ".ogg": "audio/ogg"
  }[extension];
  if (!inferred) throw new Error("Multipart attachments must use a supported image, video, or audio media type.");
  return inferred;
}

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function isSafeFormField(value) {
  return typeof value === "string" && /^[A-Za-z][A-Za-z0-9_-]{0,127}$/.test(value);
}

function isScalarFormValue(value) {
  return typeof value === "string" || typeof value === "number" || typeof value === "boolean";
}

function isSafeFilename(value) {
  return value.length > 0 && value.length <= 255 && value === path.basename(value) && !/[\0\r\n]/.test(value);
}

async function writeNativeResponse(response, outputFile, deadline) {
  if (!response.ok) {
    const responseText = await withinDeadline(response.text(), deadline);
    let body = responseText;
    try {
      body = JSON.parse(responseText);
    } catch {
      // Preserve the opaque response only in structured tool output.
    }
    writeJson({
      http_status: response.status,
      content_type: response.headers.get("content-type") || null,
      retry_after: response.headers.get("retry-after") || null,
      body
    });
    return false;
  }
  const contentType = response.headers.get("content-type") || "";
  const mediaLimit = nativeMediaLimit(contentType);
  if (!mediaLimit) {
    throw new Error("The completed media response did not contain a supported native media content type.");
  }
  await mkdir(path.dirname(outputFile), { recursive: true });
  try {
    await stat(outputFile);
    throw new Error("Refusing to overwrite an existing media file.");
  } catch (error) {
    if (error?.code !== "ENOENT") throw error;
  }
  if (!response.body) throw new Error("The completed media response did not contain native bytes.");
  try {
    await pipeline(
      Readable.fromWeb(response.body),
      new BoundedNativeMediaStream(mediaLimit),
      createWriteStream(outputFile, { flags: "wx", mode: 0o600 }),
      { signal: deadline }
    );
  } catch (error) {
    await unlink(outputFile).catch(() => undefined);
    throw error;
  }
  return true;
}

function nativeMediaLimit(contentType) {
  const normalized = contentType.toLowerCase().split(";", 1)[0].trim();
  if (normalized.startsWith("image/")) return maxImageMediaBytes;
  if (normalized.startsWith("video/") || normalized.startsWith("audio/") || normalized === "application/octet-stream") {
    return maxVideoOrAudioMediaBytes;
  }
  return undefined;
}

class BoundedNativeMediaStream extends Transform {
  constructor(limit) {
    super();
    this.limit = limit;
    this.bytes = 0;
  }

  _transform(chunk, encoding, callback) {
    this.bytes += chunk.length;
    if (this.bytes > this.limit) {
      callback(new Error("The completed media response exceeds the managed runtime delivery limit."));
      return;
    }
    callback(null, chunk);
  }
}

function writeJson(value) {
  process.stdout.write(`${JSON.stringify(value)}\n`);
}

async function withinDeadline(promise, deadline) {
  if (deadline.aborted) throw new Error("The direct API runtime exceeded its 90-second total deadline.");
  return new Promise((resolve, reject) => {
    const onAbort = () => {
      cleanup();
      reject(new Error("The direct API runtime exceeded its 90-second total deadline."));
    };
    const cleanup = () => deadline.removeEventListener("abort", onAbort);
    deadline.addEventListener("abort", onAbort, { once: true });
    Promise.resolve(promise).then((value) => {
      cleanup();
      resolve(value);
    }, (error) => {
      cleanup();
      reject(error);
    });
  });
}

class DirectRuntimeFailure extends Error {
  constructor(phase, cause) {
    super("Pure Tokens direct API runtime failed.");
    this.phase = phase;
    this.cause = cause;
  }
}

export function runtimeFailureEnvelope(error) {
  const failure = error instanceof DirectRuntimeFailure ? error : new DirectRuntimeFailure("validation", error);
  return {
    runtime_error: {
      phase: failure.phase,
      message: safeRuntimeFailureMessage(failure.cause)
    }
  };
}

function safeRuntimeFailureMessage(error) {
  for (const message of runtimeErrorMessages(error)) {
    if (message === "The direct API runtime exceeded its 90-second total deadline.") return message;
    if (message === "Standard-input request bodies are not supported; use --json-base64 or --multipart-base64.") return message;
    if (message.startsWith("No usable Pure Tokens API credential is configured for") ||
      message === "No usable Pure Tokens API-key connection is selected for Gemini CLI." ||
      message === "Trae uses manual connection setup and has no approved local credential resolver for this Skill." ||
      message === "The requested host does not have a managed Pure Tokens credential runtime.") return message;
    if (message.startsWith("--") || message.startsWith("Usage:") || message.startsWith("POST requests require") || message.startsWith("GET requests cannot") ||
      message.startsWith("Unknown option:") || message.startsWith("The requested API path") || message.startsWith("Base64 request body") ||
      message.startsWith("Multipart request body") || message.startsWith("Each multipart file") || message.startsWith("An explicit media attachment") ||
      message.startsWith("Multipart attachments") || message.startsWith("The completed media response") || message.startsWith("Refusing to overwrite")) return message;
  }
  return "The direct API runtime did not complete the request.";
}

function runtimeErrorMessages(error) {
  const messages = [];
  const visited = new Set();
  let current = error;
  while (current instanceof Error && !visited.has(current)) {
    visited.add(current);
    if (typeof current.message === "string") messages.push(current.message);
    current = current.cause;
  }
  return messages;
}

function isMainModule() {
  // The managed installer copies this standalone .mjs file without a package root.
  // Match its fixed executable filename instead of relying on package-context URL resolution.
  return Boolean(process.argv[1]) && path.basename(process.argv[1]) === "puretokens-direct-api.mjs";
}

if (isMainModule()) {
  main().catch((error) => {
    writeJson(runtimeFailureEnvelope(error));
    process.exitCode = 1;
  });
}
