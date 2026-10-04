package main

import "testing"

// Exercise count/paired-dimension and multi-result recovery independently of
// live catalog membership. This profile is test-only and never distributed.
func syntheticMultiImageService(t *testing.T, svc service) service {
	t.Helper()
	profile := modelProfile{ID: "fixture-multi-image", Capability: "image", Parameters: parameterSchema{
		Properties: map[string]map[string]any{
			"n":          {"type": "integer", "min": 1, "max": 6},
			"width":      {"type": "integer", "min": 768, "max": 2048},
			"height":     {"type": "integer", "min": 768, "max": 2048},
			"size":       {"type": "string", "enum": []any{"2048x2048"}},
			"image_urls": {"type": "string[]", "maxLength": 10},
			"strength":   {"type": "string", "enum": []any{"LOW", "MID", "HIGH"}},
		},
		Constraints: map[string]map[string]any{
			"requires_together":   {"width": []any{"height"}, "height": []any{"width"}},
			"reference_transport": {"image_urls": []any{"public_https_url"}},
		},
	}}
	svc.profilesRoot = writeInstalledProfileFixture(t, profile)
	return svc
}
