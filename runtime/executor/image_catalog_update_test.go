package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestModelIDBoundaryAndCatalogProjection(t *testing.T) {
	for _, id := range []string{"gpt-image-2", "gpt-image-2.5", "gpt-image-2(Sub)", strings.Repeat("a", 160)} {
		if !validModelID(id) {
			t.Fatalf("valid ID rejected: %s", id)
		}
	}
	for _, id := range []string{"", ".", "..", "../model", `..\model`, "model/other", "model()", "model(Sub)(Other)", "model(Sub", "$(cmd)", "model\n", strings.Repeat("a", 161)} {
		if validModelID(id) {
			t.Fatalf("unsafe ID accepted: %q", id)
		}
	}
	const body = `{"data":[{"id":"gpt-image-2(Sub)","capabilities":["image"],"input_schema":{}}]}`
	result, _, err := runModelQueryFixture(t, strings.NewReader(`{"model":"gpt-image-2(Sub)"}`), body, 200)
	if err != nil || len(modelQueryIDs(t, result)) != 1 {
		t.Fatalf("exact query lost parenthesized ID: %v", err)
	}
	request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2(Sub)", Prompt: "fixture"}
	if err := prepareProfileRequest(&request, profileService()); err != nil {
		t.Fatal(err)
	}
	record := recordFromRequest(request)
	if taskReceipt(request, "fixture-task", "pending").Model != request.Model || record.Model != request.Model {
		t.Fatal("model identity lost during continuation")
	}
}

func TestCurrentGPTImageParameters(t *testing.T) {
	for _, model := range []string{"gpt-image-2", "gpt-image-2.5", "gpt-image-2(Sub)"} {
		for _, format := range []string{"png", "jpeg", "webp"} {
			request := taskRequest{Kind: "image", Operation: "generate", Model: model, Prompt: "fixture",
				Parameters: map[string]any{"quality": "auto", "output_format": format, "response_format": "url"}}
			if err := prepareProfileRequest(&request, profileService()); err != nil {
				t.Fatal(err)
			}
			if recordFromRequest(request).Parameters["output_format"] != format || taskReceipt(request, "", "").Parameters["quality"] != "auto" {
				t.Fatal("safe output parameters lost")
			}
		}
		for _, parameters := range []map[string]any{
			{"image_size": "2K"}, {"image_size": "4K"}, {"quality": "xhigh"},
			{"quality": "max"}, {"output_format": "gif"}, {"response_format": "b64_json"}, {"n": 2},
		} {
			request := taskRequest{Kind: "image", Operation: "generate", Model: model, Prompt: "fixture", Parameters: parameters}
			if prepareProfileRequest(&request, profileService()) == nil {
				t.Fatalf("%s accepted unsupported parameters: %v", model, parameters)
			}
		}
	}
}

func TestImageSizeRatioMatrixSharedByQueryAndSubmission(t *testing.T) {
	var schema parameterSchema
	const raw = `{"properties":{"image_size":{"type":"string","enum":["1K","2K"],"default":"1K"},"aspect_ratio":{"type":"string","enum":["1:1","16:9"],"default":"1:1"}},"constraints":{"aspect_ratio_by_image_size":{"1K":["1:1"],"2K":["16:9"]}}}`
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		t.Fatal(err)
	}
	var inputSchema any
	json.Unmarshal([]byte(raw), &inputSchema)
	entry := map[string]any{"id": "fixture", "capabilities": []any{"image"}, "input_schema": inputSchema}
	for _, tc := range []struct {
		parameters map[string]any
		valid      bool
	}{
		{map[string]any{}, true},
		{map[string]any{"image_size": "1K", "aspect_ratio": "1:1"}, true},
		{map[string]any{"image_size": "2K", "aspect_ratio": "16:9"}, true},
		{map[string]any{"image_size": "1K", "aspect_ratio": "16:9"}, false},
		{map[string]any{"image_size": "2K"}, false},
		{map[string]any{"aspect_ratio": "16:9"}, false},
	} {
		query := modelQuery{Kind: "image", Parameters: tc.parameters}
		if modelQueryMatches(query, entry) != tc.valid {
			t.Fatalf("query mismatch: %v", tc.parameters)
		}
		profile := modelProfile{ID: "fixture", Capability: "image", Parameters: schema}
		svc, transport := catalogProfileService(t, profile)
		svc.profilesRoot = writeInstalledProfileFixture(t, profile)
		request := taskRequest{Kind: "image", Operation: "generate", Model: "fixture", Prompt: "fixture", Parameters: tc.parameters}
		if (prepareProfileRequest(&request, svc) == nil) != tc.valid || transport.calls != 0 {
			t.Fatalf("submission mismatch or unexpected discovery: %v", tc.parameters)
		}
	}
}
