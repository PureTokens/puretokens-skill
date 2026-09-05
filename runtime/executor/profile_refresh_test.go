package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type profileCatalogTransport struct {
	t      *testing.T
	calls  int
	status int
	body   []byte
}

func (transport *profileCatalogTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	transport.calls++
	if request.Method != http.MethodGet || request.URL.Path != "/v1/media/models" {
		transport.t.Fatalf("profile validation attempted an unexpected request: %s %s", request.Method, request.URL.Path)
	}
	if request.Header.Get("Authorization") != "Bearer profile-fixture" {
		transport.t.Fatal("catalog request did not use the supplied service credential")
	}
	return &http.Response{
		StatusCode: transport.status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(transport.body)),
		Request:    request,
	}, nil
}

func readInstalledProfileFixture(t *testing.T, kind, id string) modelProfile {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("../../skills", "puretokens-"+kind, "references", "profiles", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var profile modelProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		t.Fatal(err)
	}
	return profile
}

func catalogProfileService(t *testing.T, profile modelProfile) (service, *profileCatalogTransport) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"data": []any{map[string]any{
		"id": profile.ID, "capabilities": []string{profile.Capability}, "input_schema": profile.Parameters,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	transport := &profileCatalogTransport{t: t, status: http.StatusOK, body: body}
	root, err := filepath.Abs("../../skills")
	if err != nil {
		t.Fatal(err)
	}
	return service{
		profilesRoot: root,
		baseURL:      "https://profile-fixture.invalid",
		token:        "profile-fixture",
		client:       &http.Client{Transport: transport},
	}, transport
}

func writeInstalledProfileFixture(t *testing.T, profile modelProfile) string {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "puretokens-"+profile.Capability, "references", "profiles")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, profile.ID+".json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestExistingProfileRefreshesMissingOptionalFieldOnce(t *testing.T) {
	profile := readInstalledProfileFixture(t, "image", "gpt-image-2")
	profile.Parameters.Properties["quality"] = map[string]any{"type": "string", "enum": []any{"high"}}
	svc, transport := catalogProfileService(t, profile)
	request := taskRequest{
		Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt",
		Parameters: map[string]any{"image_size": "2K", "quality": "high"},
	}
	if err := prepareProfileRequest(&request, svc); err != nil {
		t.Fatal(err)
	}
	if transport.calls != 1 || request.Parameters["quality"] != "high" || request.Parameters["image_size"] != "2K" {
		t.Fatalf("refresh count or user parameters changed: calls=%d parameters=%v", transport.calls, request.Parameters)
	}
	if request.Model != "gpt-image-2" || endpointFor(request) != "/v1/images/generations" {
		t.Fatal("refresh changed the requested model or core route")
	}
}

func TestExistingProfileRefreshesMissingOperationOnce(t *testing.T) {
	profile := readInstalledProfileFixture(t, "video", "seedance-2.0")
	edit := readInstalledProfileFixture(t, "video", "grok-imagine-video")
	profile.Parameters.Operations["video_edit"] = edit.Parameters.Operations["video_edit"]
	svc, transport := catalogProfileService(t, profile)
	request := taskRequest{
		Kind: "video", Operation: "edit", Model: profile.ID, Prompt: "fixture prompt",
		Attachments: []attachment{{Field: "video", Path: "/current/fixture.mp4"}},
	}
	if err := prepareProfileRequest(&request, svc); err != nil {
		t.Fatal(err)
	}
	if transport.calls != 1 || endpointFor(request) != "/v1/videos/edits" || request.Model != profile.ID {
		t.Fatal("missing operation did not resolve exactly once with its declared route")
	}
}

func TestExistingProfileRefreshesMissingReferenceTransportOnce(t *testing.T) {
	for _, mode := range []string{"json-operation", "reference-constraint", "native-transport", "native-field"} {
		t.Run(mode, func(t *testing.T) {
			local := readInstalledProfileFixture(t, "image", "gpt-image-2")
			live := readInstalledProfileFixture(t, "image", "gpt-image-2")
			request := taskRequest{Kind: "image", Operation: "generate", Model: local.ID, Prompt: "fixture prompt"}
			switch mode {
			case "json-operation":
				request.Operation = "edit"
				request.Parameters = map[string]any{"image": []any{"https://example.invalid/current.png"}}
				op := live.Parameters.Operations["image_edit"]
				op.Request.ContentType = "application/json"
				input := op.Inputs["image"]
				input.Transports = []string{"public_https_url"}
				op.Inputs["image"] = input
				live.Parameters.Operations["image_edit"] = op
			case "reference-constraint":
				delete(local.Parameters.Constraints, "reference_transport")
				request.Parameters = map[string]any{"image": []any{"https://example.invalid/current.png"}}
			case "native-transport", "native-field":
				request.Operation = "edit"
				request.Attachments = []attachment{{Field: "image", Path: "/current/fixture.png"}}
				op := local.Parameters.Operations["image_edit"]
				if mode == "native-transport" {
					input := op.Inputs["image"]
					input.Transports = nil
					op.Inputs["image"] = input
				} else {
					delete(op.Inputs, "image")
				}
				local.Parameters.Operations["image_edit"] = op
			}
			svc, transport := catalogProfileService(t, live)
			svc.profilesRoot = writeInstalledProfileFixture(t, local)
			if err := prepareProfileRequest(&request, svc); err != nil {
				t.Fatal(err)
			}
			if transport.calls != 1 {
				t.Fatalf("expected one catalog read, got %d", transport.calls)
			}
		})
	}
}

func TestProfileRefreshNeverRelaxesInvalidLocalValues(t *testing.T) {
	for _, parameters := range []map[string]any{
		{"image_size": "8K"},
		{"image_size": "8K", "quality": "high"},
		{"n": 2, "quality": "high"},
		{"image": []any{"https://localhost./current.png"}, "quality": "high"},
	} {
		profile := readInstalledProfileFixture(t, "image", "gpt-image-2")
		profile.Parameters.Properties["image_size"]["enum"] = []any{"1K", "2K", "4K", "8K"}
		profile.Parameters.Properties["n"]["enum"] = []any{1, 2}
		profile.Parameters.Properties["quality"] = map[string]any{"type": "string", "enum": []any{"high"}}
		svc, transport := catalogProfileService(t, profile)
		request := taskRequest{Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt", Parameters: parameters}
		original, _ := json.Marshal(request.Parameters)
		if prepareProfileRequest(&request, svc) == nil || transport.calls != 0 {
			t.Fatal("invalid local values caused a refresh or were accepted")
		}
		after, _ := json.Marshal(request.Parameters)
		if !bytes.Equal(original, after) {
			t.Fatal("invalid values were silently changed")
		}
	}
	profile := readInstalledProfileFixture(t, "image", "gpt-image-2")
	svc, transport := catalogProfileService(t, profile)
	request := taskRequest{
		Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt",
		RequestedCount: 2, Parameters: map[string]any{"quality": "high"},
	}
	if prepareProfileRequest(&request, svc) == nil || transport.calls != 0 {
		t.Fatal("invalid requested_count caused a refresh or was accepted")
	}
}

func TestProfileGapStopsAfterOneUnusableCatalogResponse(t *testing.T) {
	for _, mode := range []string{"unchanged", "wrong-model", "wrong-capability", "unavailable", "logical-error"} {
		t.Run(mode, func(t *testing.T) {
			profile := readInstalledProfileFixture(t, "image", "gpt-image-2")
			if mode == "wrong-model" {
				profile.ID = "different-model"
			}
			if mode == "wrong-capability" {
				profile.Capability = "video"
			}
			svc, transport := catalogProfileService(t, profile)
			if mode == "unavailable" {
				transport.status = http.StatusServiceUnavailable
			}
			if mode == "logical-error" {
				transport.body = []byte(`{"success":false,"data":[]}`)
			}
			request := taskRequest{
				Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "fixture prompt",
				Parameters: map[string]any{"quality": "high"},
			}
			if prepareProfileRequest(&request, svc) == nil || transport.calls != 1 {
				t.Fatalf("unusable catalog should stop after one read, got %d", transport.calls)
			}
		})
	}
}

func TestInstalledProfileFieldsAndNativeOperationsNeedNoRefresh(t *testing.T) {
	attachmentPath := writeFixture(t, "reference.png", string(fixturePNG(t)))
	for _, kind := range []string{"image", "video"} {
		files, err := filepath.Glob("../../skills/puretokens-" + kind + "/references/profiles/*.json")
		if err != nil || len(files) == 0 {
			t.Fatal("installed profile fixtures unavailable", err)
		}
		for _, file := range files {
			id := strings.TrimSuffix(filepath.Base(file), ".json")
			profile := readInstalledProfileFixture(t, kind, id)
			t.Run(id, func(t *testing.T) {
				svc, transport := catalogProfileService(t, profile)
				for name, rule := range profile.Parameters.Properties {
					if name == "prompt" || name == "model" {
						continue
					}
					for _, value := range profileFieldValues(t, name, rule) {
						request := taskRequest{
							Kind: kind, Operation: "generate", Model: id, Prompt: "fixture prompt",
							Parameters: map[string]any{name: value},
						}
						if name == "width" {
							request.Parameters["height"] = 1024
						}
						if name == "height" {
							request.Parameters["width"] = 1024
						}
						if err := prepareProfileRequest(&request, svc); err != nil {
							t.Fatalf("supported %s=%v failed: %v", name, value, err)
						}
						if _, _, _, err := taskRequestBody(request); err != nil {
							t.Fatal(err)
						}
					}
				}
				for name, op := range profile.Parameters.Operations {
					for _, upperBound := range []bool{false, true} {
						request := taskRequest{Kind: kind, Operation: "generate", Model: id, MediaOperation: name, Prompt: "fixture prompt"}
						if strings.HasSuffix(name, "_edit") {
							request.Operation = "edit"
						}
						for _, input := range op.Inputs {
							count := max(1, input.Min)
							if upperBound && input.Max > 0 {
								count = input.Max
							}
							for i := 0; i < count; i++ {
								request.Attachments = append(request.Attachments, attachment{Field: input.Field, Path: attachmentPath})
							}
						}
						if err := prepareProfileRequest(&request, svc); err != nil {
							t.Fatalf("supported %s failed: %v", name, err)
						}
						route, contentType, body, err := taskRequestBody(request)
						if err != nil {
							t.Fatal(err)
						}
						if route != op.Request.Path || !strings.HasPrefix(contentType, "multipart/form-data;") {
							t.Fatal("operation route or native transport changed")
						}
						_, err = io.Copy(io.Discard, body)
						closeErr := body.(io.Closer).Close()
						if err != nil || closeErr != nil {
							t.Fatal("native attachment serialization failed", err, closeErr)
						}
					}
				}
				if transport.calls != 0 {
					t.Fatalf("supported installed requests read the live catalog %d times", transport.calls)
				}
			})
		}
	}
}

func profileFieldValues(t *testing.T, name string, rule map[string]any) []any {
	t.Helper()
	if values := array(rule["enum"]); len(values) > 0 {
		return values
	}
	switch rule["type"] {
	case "integer", "number":
		minimum, ok := number(rule["min"])
		if !ok {
			minimum = 1
		}
		values := []any{minimum}
		if maximum, ok := number(rule["max"]); ok {
			values = append(values, maximum)
		}
		return values
	case "boolean":
		return []any{true, false}
	case "string":
		if mediaReferenceField(name) {
			return []any{"https://example.invalid/reference.png"}
		}
		return []any{"1024x1024"}
	case "string[]":
		values := []any{[]any{"https://example.invalid/reference.png"}}
		if maximum, ok := number(rule["maxLength"]); ok {
			references := make([]any, int(maximum))
			for i := range references {
				references[i] = fmt.Sprintf("https://example.invalid/reference-%d.png", i)
			}
			values = append(values, references)
		}
		return values
	case "json":
		return []any{
			"https://example.invalid/reference.png",
			[]any{"https://example.invalid/reference.png"},
			map[string]any{"url": "https://example.invalid/reference.png"},
			map[string]any{"image_url": "https://example.invalid/reference.png"},
		}
	default:
		t.Fatalf("missing fixture for installed property type %v", rule["type"])
		return nil
	}
}

func TestReferenceURLHostNormalizationAndStrictObjects(t *testing.T) {
	profile := readInstalledProfileFixture(t, "image", "grok-imagine-image")
	svc, transport := catalogProfileService(t, profile)
	for _, host := range []string{
		"localhost.", "LOCALHOST.", "camera.local.", "camera.local...", "sub.localhost.",
		"127.0.0.1.", "10.0.0.1.", "[::1]", "[fe80::1%25en0]", "[ff02::1%25en0]",
	} {
		for _, reference := range []any{
			"https://" + host + "/reference.png",
			[]any{"https://" + host + "/reference.png"},
			map[string]any{"url": "https://" + host + "/reference.png"},
			map[string]any{"image_url": "https://" + host + "/reference.png"},
		} {
			request := taskRequest{Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt", Parameters: map[string]any{"image": reference}}
			if prepareProfileRequest(&request, svc) == nil {
				t.Fatalf("non-public host accepted: %s", host)
			}
		}
	}
	for _, reference := range []any{
		map[string]any{"url": "https://example.invalid/reference.png", "extra": "not declared"},
		map[string]any{"url": "https://example.invalid/reference.png", "image_url": "https://example.invalid/other.png"},
		map[string]any{"url": map[string]any{"url": "https://example.invalid/reference.png"}},
		map[string]any{"URL": "https://example.invalid/reference.png"},
	} {
		request := taskRequest{Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt", Parameters: map[string]any{"image": reference}}
		if prepareProfileRequest(&request, svc) == nil {
			t.Fatal("undeclared URL-object shape accepted")
		}
	}
	for _, reference := range []any{
		"https://Example.invalid./reference.png?version=1",
		map[string]any{"url": "https://example.invalid./reference.png"},
		map[string]any{"image_url": "https://[2001:4860:4860::8888]/reference.png"},
	} {
		request := taskRequest{Kind: "image", Operation: "generate", Model: profile.ID, Prompt: "fixture prompt", Parameters: map[string]any{"image": reference}}
		if err := prepareProfileRequest(&request, svc); err != nil {
			t.Fatal("public reference rejected", err)
		}
		_, _, body, err := taskRequestBody(request)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.NewDecoder(body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payload["image"], reference) {
			t.Fatal("accepted public URL or object representation was rewritten")
		}
	}
	if transport.calls != 0 {
		t.Fatal("URL value validation incorrectly requested a catalog refresh")
	}
}
