# Media Parameter Review

Reviewed 2026-09-18. Public source: the production console's
`/api/product/docs/model-catalog`. This is selection metadata, not authenticated
model access or paid generation proof.

- GPT Image: exact IDs are `gpt-image-2`, `gpt-image-2.5` and
  `gpt-image-2(Sub)`. Retire Flare/Sunburst selection profiles without adding
  replacement aliases or rewriting existing task identities.
- All three declare only `image_size=1K`, `n=1`, at most ten URL references
  or multipart image-edit files, `quality=auto|low|medium|high`,
  `output_format=png|jpeg|webp`, and `response_format=url`.
  Ratios are `1:1`, `2:3`, `3:2`, `3:4`, `4:3`, `4:5`, `5:4`, `9:16`,
  `16:9`, `21:9`; do not retain the retired variants' `3:1`/`1:3`.
  Pricing rows are not authority to enable 2K/4K.
- Enforce `aspect_ratio_by_image_size` in both query filters and submission,
  using declared defaults for omitted fields. Model IDs permit one bounded
  parenthesized suffix while still rejecting path separators and traversal.
  Submission, query projection, safe receipts and continuation retain the
  exact ID. Supplier routing stays inside the gateway; public task lifecycle
  remains asynchronous.
- The catalog still contains 12 image and 11 video models. All other public
  parameter and operation declarations match the previous snapshot. Backend
  adapter additions alone do not expand the installed public API contract.
- Nano Banana 2, Lite and Pro: the live aspect-ratio enums no longer declare
  `4:5` or `5:4`. Synchronize those enums exactly; do not silently convert ratios.
- Seedance 2.0, Fast, Mini and 2.5: the live schema still declares
  `first_frame_image` and `last_frame_image`, but omits their operation entries.
  Retain the previously supported frame operations using an explicit reviewed
  supplement first reviewed on 2026-09-12, not a merge of arbitrary stale
  metadata.
- Local Web source evidence:
  `upstream-new-api-src/relay/channel/task/seedance/adaptor.go`,
  `normalizeRequestForModel`, reads both frame fields and rejects mixed
  frame/general references. `adaptor_test.go`,
  `TestNormalizeRequestMapsFirstLastFramesToDocumentedPayload`, verifies that
  mapping and rejection. These were read in `PureTokenPlus_web`; that repository
  and production settings were not modified.

`scripts/reviewed-media-supplements.mjs` records local compatibility additions
separately from the public capture. It only fills absent operation/constraint
entries for the four exact models while both frame fields remain declared.
Explicit live definitions take precedence. Missing fields do not receive a
supplement. All other parameters come directly from the public schema.

No new real-host, authenticated catalog, paid task, or media delivery acceptance
is claimed by this refresh. Existing frame regression tests remain in place.
