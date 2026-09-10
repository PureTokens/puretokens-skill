package main

import "testing"

func TestImage25QualityAndEditBoundaries(t *testing.T) {
	for _, model := range []string{"gpt-image-2.5-flare", "gpt-image-2.5-sunburst"} {
		for _, quality := range []string{"low", "medium", "high", "xhigh", "max"} {
			r := taskRequest{Kind: "image", Operation: "generate", Model: model, Prompt: "fixture", Parameters: map[string]any{"quality": quality, "image_size": "2K", "aspect_ratio": "16:9"}}
			if err := prepareProfileRequest(&r, profileService()); err != nil {
				t.Fatal(err)
			}
			if recordFromRequest(r).Parameters["quality"] != quality {
				t.Fatal("quality lost from record")
			}
		}
		r := taskRequest{Kind: "image", Operation: "edit", Model: model, Prompt: "fixture"}
		if prepareProfileRequest(&r, profileService()) == nil {
			t.Fatal("edit without references accepted")
		}
		r.Operation = "generate"
		r.Parameters = map[string]any{"quality": "private free text"}
		if prepareProfileRequest(&r, profileService()) == nil {
			t.Fatal("invalid quality accepted")
		}
		if _, ok := taskReceipt(r, "", "").Parameters["quality"]; ok {
			t.Fatal("unsafe quality retained")
		}
	}
}
func TestSeedanceFrameReferences(t *testing.T) {
	for _, model := range []string{"seedance-2.0", "seedance-2.0-fast", "seedance-2.0-mini", "seedance-2.5"} {
		r := taskRequest{Kind: "video", Operation: "generate", Model: model, Prompt: "fixture", Parameters: map[string]any{"first_frame_image": "https://example.com/first.png", "last_frame_image": "https://example.com/last.png"}}
		if err := prepareProfileRequest(&r, profileService()); err != nil {
			t.Fatal(err)
		}
		if len(taskReceipt(r, "", "").Parameters) != 0 {
			t.Fatal("reference persisted")
		}
		r.Parameters["reference_videos"] = []any{"https://example.com/ref.mp4"}
		if prepareProfileRequest(&r, profileService()) == nil {
			t.Fatal("mixed frame/reference accepted")
		}
		r.Parameters = map[string]any{}
		r.MediaOperation = "first_last_frame_video"
		r.Attachments = []attachment{{Field: "first_frame_image", Path: "first.png"}, {Field: "last_frame_image", Path: "last.png"}}
		if err := prepareProfileRequest(&r, profileService()); err != nil {
			t.Fatal(err)
		}
		r.Attachments = r.Attachments[:1]
		if prepareProfileRequest(&r, profileService()) == nil {
			t.Fatal("missing last frame accepted")
		}
	}
}

func TestImage25ReferenceEdits(t *testing.T) {
	for _, model := range []string{"gpt-image-2.5-flare", "gpt-image-2.5-sunburst"} {
		for _, count := range []int{1, 6, 7} {
			r := taskRequest{Kind: "image", Operation: "edit", Model: model, Prompt: "fixture", Parameters: map[string]any{"quality": "high"}}
			for i := 0; i < count; i++ {
				r.Attachments = append(r.Attachments, attachment{Field: "image", Path: "fixture.png"})
			}
			err := prepareProfileRequest(&r, profileService())
			if (err == nil) != (count <= 6) {
				t.Fatalf("%s count %d: %v", model, count, err)
			}
		}
		for _, count := range []int{1, 6, 7} {
			refs := make([]any, count)
			for i := range refs {
				refs[i] = "https://example.com/ref.png"
			}
			r := taskRequest{Kind: "image", Operation: "generate", Model: model, Prompt: "fixture", Parameters: map[string]any{"image": refs}}
			err := prepareProfileRequest(&r, profileService())
			if (err == nil) != (count <= 6) {
				t.Fatalf("%s URL count %d: %v", model, count, err)
			}
			if _, exists := taskReceipt(r, "", "").Parameters["image"]; exists {
				t.Fatal("references persisted")
			}
		}
	}
}
func TestNanoBananaPortraitLandscapeRatios(t *testing.T) {
	for _, model := range []string{"nano-banana-2", "nano-banana-2-lite", "nano-banana-pro"} {
		for _, ratio := range []string{"4:5", "5:4"} {
			r := taskRequest{Kind: "image", Operation: "generate", Model: model, Prompt: "fixture", Parameters: map[string]any{"aspect_ratio": ratio, "image_size": "1K"}}
			if err := prepareProfileRequest(&r, profileService()); err != nil {
				t.Fatalf("%s %s: %v", model, ratio, err)
			}
		}
	}
	r := taskRequest{Kind: "image", Operation: "generate", Model: "nano-banana-2-lite", Prompt: "fixture", Parameters: map[string]any{"aspect_ratio": "4:5", "image_size": "2K"}}
	if prepareProfileRequest(&r, profileService()) == nil {
		t.Fatal("lite accepted unsupported 2K")
	}
}
