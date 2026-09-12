# Media Parameter Review

Reviewed 2026-09-12. Public source: the production console's
`/api/product/docs/model-catalog`. This is selection metadata, not authenticated
model access or paid generation proof.

- Nano Banana 2, Lite and Pro: the live aspect-ratio enums no longer declare
  `4:5` or `5:4`. Synchronize those enums exactly; do not silently convert ratios.
- Seedance 2.0, Fast, Mini and 2.5: the live schema still declares
  `first_frame_image` and `last_frame_image`, but omits their operation entries.
  Retain the previously supported frame operations using an explicit reviewed
  supplement, not a merge of arbitrary stale metadata.
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
