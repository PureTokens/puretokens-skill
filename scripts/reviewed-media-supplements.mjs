// Local compatibility metadata, not an authenticated/live availability claim.
// Evidence and scope: references/media-parameter-review.md.
const seedanceModels = new Set(["seedance-2.0", "seedance-2.0-fast", "seedance-2.0-mini", "seedance-2.5"]);

export function supplementReviewedParameters(model) {
  const schema = structuredClone(model.parameterSchema);
  if (!seedanceModels.has(model.id) || !schema ||
      schema.properties?.first_frame_image?.type !== "string" ||
      schema.properties?.last_frame_image?.type !== "string") return schema;
  schema.operations ??= {};
  const sets = {
    first_frame_video: ["first_frame_image"],
    last_frame_video: ["last_frame_image"],
    first_last_frame_video: ["first_frame_image", "last_frame_image"]
  };
  for (const [name, fields] of Object.entries(sets)) {
    // A live definition wins; only fill the documented missing operation.
    schema.operations[name] ??= {
      request: { method: "POST", path: "/v1/videos", contentType: "multipart/form-data" },
      requiredBodyFields: ["model", "prompt", ...fields],
      inputs: Object.fromEntries(fields.map(field => [field, {
        field, required: true, minItems: 1, maxItems: 1, transports: ["multipart_file"]
      }]))
    };
  }
  schema.constraints ??= {};
  schema.constraints.reference_transport ??= {
    first_frame_image: ["public_https_url"], last_frame_image: ["public_https_url"]
  };
  schema.constraints.exclusive_reference_sets ??= {
    frames: ["first_frame_image", "last_frame_image"],
    references: ["image", "reference_images", "reference_videos", "reference_audios"]
  };
  return schema;
}
