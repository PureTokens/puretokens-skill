# Jev evaluation integration

Reviewed on 2026-09-27. This repository owns the client-side adapter and Skill;
it does not configure gateway channels, mappings, prices or provider credentials.

## Authority and protocol

- TypeSafe HTTP API: https://docs.typesafe.ai/api
- TypeSafe models: https://docs.typesafe.ai/models
- Confidence semantics: https://docs.typesafe.ai/confidence
- Original authoring guide: https://github.com/typesafe-ai/skills
- Gateway native adapter:
  https://github.com/QuantumNous/new-api-plugins/blob/main/plugins/tasks/typesafe/1.0.0/plugin.js

The reviewed `typesafe` 1.0.0 native route is
`POST https://api.puretokensx.com/typesafe/v1/systemone`. It uses the existing
Pure Tokens bearer credential, not a TypeSafe account or SDK. The synchronous
renderer returns `{model, answers, usage}` directly. It does not return a media
task receipt. The plugin marks the internal task immediate and `retainResult:
false`; its query hook explicitly has no retrieval endpoint. Do not use
`/v1/tasks/typesafe` or invent a status/content path as a client fallback.

Reviewed request model IDs: `jev-latest`, `jev-1.13.0`, `jev-preview`.
`jev` is not an API model ID in this adapter. Default is `jev-latest`; callers
may explicitly pin the reviewed version. Version aliases can return a newer
well-formed Jev semantic version without changing the returned identity.

## Client contract

The managed one-shot executor exposes `evaluate --host <host> --request
<absolute-file>`. Request validation happens before credential resolution.
It makes one synchronous JSON POST with a 90-second total deadline and no
redirect, retries, preflight, media polling or fallback transport.

Text/structured text `state`; named independent questions of type choice,
score or noul; no media or streaming. Required instructions and the stricter
documented subset follow the HTTP API, not all permissive SDK/plugin inputs.
Local resource limits: 1 MiB request, JSON depth 32, 64 questions, IDs up to 64
ASCII identifier characters, option labels up to 128 Unicode characters,
1–255 choice options and 2–10 score levels. The 1 MiB limit is not a tokenizer
estimate; upstream context admission remains authoritative.

Typed projection validates each answer against its requested question,
including probability distributions and score/choice consistency. Arbitrary
upstream descriptions, errors and legends are never copied into receipts.
Results are user-requested data; they are never included in the safe support
summary. Unknown submission stops without claiming a failed evaluation, no
charge, refund, or recoverability.

Explicit model discovery uses one authenticated `GET /v1/models`, intersects
the returned IDs with the reviewed request-model set, and labels the result
`reviewed_evaluation_models_listed_by_api`. This proves directory visibility
only, not native-route availability, price, authorization to submit or output
quality. Existing image/video discovery remains `/v1/media/models`.

## Gateway and release boundary

The gateway must deploy the reviewed native plugin route, bind an eligible
channel and model mapping, preserve request authentication, and return the
documented synchronous envelope. A listing or offline fixture cannot prove
the deployed route or account billing. No gateway code change was required
for the already implemented dynamic plugin route mechanism; any future
protocol change must update this contract and both repositories' tests.

No live paid evaluation has been accepted as part of this implementation.
Offline protocol, credential-fixture, packaging and scenario checks must be
reported separately from real-host/production acceptance. The pending host
media acceptance record does not establish Jev acceptance.
