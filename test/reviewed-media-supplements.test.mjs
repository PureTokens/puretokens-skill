import assert from "node:assert/strict";
import test from "node:test";
import { supplementReviewedParameters } from "../scripts/reviewed-media-supplements.mjs";

test("reviewed frame supplements preserve live authority and do not restore absent fields", () => {
  const model = { id: "seedance-2.0", parameterSchema: {
    properties: { first_frame_image: { type: "string" }, last_frame_image: { type: "string" } }
  } };
  const result = supplementReviewedParameters(model);
  assert.equal(Object.keys(result.operations).length, 3);
  assert.deepEqual(result.operations.first_last_frame_video.requiredBodyFields, ["model", "prompt", "first_frame_image", "last_frame_image"]);
  assert.equal(model.parameterSchema.operations, undefined);
  const live = structuredClone(model);
  live.parameterSchema.operations = { first_frame_video: { disabled: true } };
  assert.deepEqual(supplementReviewedParameters(live).operations.first_frame_video, { disabled: true });
  delete live.parameterSchema.properties.last_frame_image;
  assert.deepEqual(supplementReviewedParameters(live), live.parameterSchema);
  assert.equal(supplementReviewedParameters({ ...model, id: "unknown" }).operations, undefined);
  assert.equal(supplementReviewedParameters({ id: model.id }), undefined);
});
