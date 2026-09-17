package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestModelQueryAndSubmissionShareReferenceConstraints(t *testing.T) {
	root := filepath.Join("..", "..", "skills")
	data, err := os.ReadFile(filepath.Join(root, "puretokens-video", "references", "profiles", "seedance-2.5.json"))
	if err != nil {
		t.Fatal(err)
	}
	var profile modelProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		t.Fatal(err)
	}
	var schema any
	raw, _ := json.Marshal(profile.Parameters)
	json.Unmarshal(raw, &schema)
	entry := map[string]any{"id": profile.ID, "capabilities": []any{"video"}, "input_schema": schema}
	for _, tc := range []struct {
		name       string
		parameters map[string]any
		valid      bool
	}{
		{"reference-only", map[string]any{}, true},
		{"first-frame-with-reference", map[string]any{"first_frame_image": "https://example.com/frame.png"}, false},
		{"last-frame-with-reference", map[string]any{"last_frame_image": "https://example.com/frame.png"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query := modelQuery{Kind: "video", Model: profile.ID, Operation: "reference_image_video", Parameters: tc.parameters}
			if got := modelQueryMatches(query, entry); got != tc.valid {
				t.Errorf("query match = %v, want %v", got, tc.valid)
			}
			request := taskRequest{Kind: "video", Operation: "generate", Model: profile.ID, Prompt: "fixture", MediaOperation: query.Operation,
				Parameters: tc.parameters, Attachments: []attachment{{Field: "reference_images", Path: "fixture.png"}}}
			if err := prepareProfileRequest(&request, service{profilesRoot: root}); (err == nil) != tc.valid {
				t.Errorf("submission validation disagrees: %v", err)
			}
		})
	}
}

const modelQueryFixture = `{"data":[
 {"id":"image-many","capabilities":["image"],"price":0.01,"provider":"private-upstream",
  "input_schema":{"properties":{"n":{"type":"integer","minimum":1,"maximum":6},"width":{"type":"integer","minimum":256},"height":{"type":"integer","minimum":256},"size":{"type":"string","enum":["1K","2K"]}},
   "constraints":{"requires_together":{"width":["height"],"height":["width"]}},
   "operations":{"image_edit":{"request":{"method":"POST","path":"/v1/images/edits","contentType":"multipart/form-data"},"inputs":{"image":{"field":"image","required":true,"maxItems":2,"transports":["multipart_file"]}}}},
   "lifecycle":{"poll":{"successStatuses":["completed"]}}}},
 {"id":"image-one","capabilities":["image"],"input_schema":{"properties":{"n":{"type":"integer","enum":[1]}}}},
 {"id":"video-audio","capabilities":["video"],"input_schema":{"properties":{"duration":{"type":"integer","enum":[5,10]},"generate_audio":{"type":"boolean"},"resolution":{"type":"string","enum":["720p","1080p"]}},
  "constraints":{"resolution_by_mode":{"text":["720p"],"image":["720p","1080p"]}},
  "operations":{"image_to_video":{"request":{"method":"POST","path":"/v1/videos","contentType":"multipart/form-data"},"inputs":{"image":{"field":"image","required":true,"transports":["multipart_file"]}}}}}},
 {"id":"video-silent","capabilities":["video"],"input_schema":{"properties":{"duration":{"type":"integer","enum":[5]}}}},
 {"id":"not-a-media-model","capabilities":["chat"]},
 {"id":"unsafe-model","capabilities":["image"],"input_schema":{"operations":{"image_edit":null}}}
]}`

func runModelQueryFixture(t *testing.T, input io.Reader, body string, status int) (map[string]any, string, error) {
	t.Helper()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		calls++
		if request.Method != http.MethodGet || request.URL.Path != "/v1/media/models" || request.URL.RawQuery != "" || request.ContentLength > 0 {
			t.Error("filters were submitted remotely or a non-catalog route was used")
		}
		w.WriteHeader(status)
		io.WriteString(w, body)
	}))
	defer server.Close()
	var output bytes.Buffer
	err := executeModelQuery(&output, service{baseURL: server.URL, client: server.Client(), token: "synthetic-model-fixture"}, input)
	if calls != 1 {
		t.Fatalf("expected exactly one catalog request, got %d", calls)
	}
	var result map[string]any
	if json.Unmarshal(output.Bytes(), &result) != nil {
		t.Fatalf("invalid model receipt: %s", output.String())
	}
	return result, output.String(), err
}

func modelQueryIDs(t *testing.T, result map[string]any) []string {
	t.Helper()
	payload, ok := result["result"].(map[string]any)
	if !ok {
		t.Fatal("missing models result")
	}
	ids := []string{}
	for _, entry := range payload["data"].([]any) {
		ids = append(ids, entry.(map[string]any)["id"].(string))
	}
	return ids
}

func TestModelQueryUnfilteredPreservesDeclaredSchemaOnly(t *testing.T) {
	result, output, err := runModelQueryFixture(t, nil, modelQueryFixture, 200)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(modelQueryIDs(t, result), []string{"image-many", "image-one", "video-audio", "video-silent", "unsafe-model"}) {
		t.Fatal("unfiltered catalog lost media entries")
	}
	if strings.Contains(output, "private-upstream") || strings.Contains(output, `"price"`) {
		t.Fatal("undeclared price or provider data escaped the public projection")
	}
	entry := result["result"].(map[string]any)["data"].([]any)[0].(map[string]any)
	schema := entry["input_schema"].(map[string]any)
	for _, key := range []string{"properties", "constraints", "operations", "lifecycle"} {
		if schema[key] == nil {
			t.Fatalf("public declaration %s was lost", key)
		}
	}
}

func TestModelQueryMatchesExplicitFieldsAndOperations(t *testing.T) {
	for _, test := range []struct {
		filter string
		want   []string
	}{
		{`{"kind":"image","parameters":{"n":3}}`, []string{"image-many"}},
		{`{"model":"image-one"}`, []string{"image-one"}},
		{`{"kind":"video","model":"image-one"}`, []string{}},
		{`{"kind":"image","operation":"image_edit"}`, []string{"image-many"}},
		{`{"kind":"video","parameters":{"duration":10,"generate_audio":true}}`, []string{"video-audio"}},
		{`{"kind":"video","parameters":{"resolution":"1080p"}}`, []string{}},
		{`{"kind":"video","operation":"image_to_video","parameters":{"resolution":"1080p"}}`, []string{"video-audio"}},
		{`{"kind":"image","parameters":{"width":1024}}`, []string{}},
		{`{"kind":"image","parameters":{"width":1024,"height":1024}}`, []string{"image-many"}},
		{`{"kind":"image","parameters":{"n":3.5}}`, []string{}},
		{`{"kind":"image","parameters":{"n":7}}`, []string{}},
		{`{"kind":"image","parameters":{"size":"4K"}}`, []string{}},
		{`{"parameters":{"quality":"best"}}`, []string{}},
		{`{"operation":"undeclared_operation"}`, []string{}},
	} {
		t.Run(test.filter, func(t *testing.T) {
			result, _, err := runModelQueryFixture(t, strings.NewReader(test.filter), modelQueryFixture, 200)
			if err != nil || !reflect.DeepEqual(modelQueryIDs(t, result), test.want) {
				t.Fatalf("unexpected matches: %+v; error: %v", result, err)
			}
		})
	}
}

func TestModelQueryRejectsInvalidFilterBeforeNetwork(t *testing.T) {
	for _, input := range []string{
		`{"kind":"audio"}`, `{"model":"https://invalid.example"}`,
		`{"model":"sk-secret-fixture"}`, `{"prompt":"unsupported-root-field"}`,
		`{"parameters":{"bad field":1}}`, `{} {}`, `[]`, `null`,
		`{"kind":`, `{"kind":"image","parameters":[]}`, `{"model":".."}`,
		`{"kind":null}`, `{"parameters":null}`,
	} {
		var output bytes.Buffer
		err := executeModelQuery(&output, service{}, strings.NewReader(input))
		if err == nil || !strings.Contains(output.String(), `"failure_phase":"validation"`) {
			t.Fatalf("invalid filter accepted: %q => %s", input, output.String())
		}
		if strings.Contains(output.String(), "sk-secret-fixture") || strings.Contains(output.String(), "unsupported-root-field") {
			t.Fatal("invalid user filter was echoed")
		}
	}
}

func TestModelQueryNeverEchoesParameterValues(t *testing.T) {
	result, output, err := runModelQueryFixture(t, strings.NewReader(`{"parameters":{"prompt":"sensitive-user-prompt"}}`), modelQueryFixture, 200)
	if err != nil || len(modelQueryIDs(t, result)) != 0 {
		t.Fatal("undeclared prompt filter should have no matches")
	}
	if strings.Contains(output, "sensitive-user-prompt") {
		t.Fatal("query echoed sensitive parameter value")
	}
}

func TestModelQueryMalformedAndRejectedCatalogsFailSafely(t *testing.T) {
	for _, test := range []struct {
		body   string
		status int
	}{
		{`{"data":"not-a-list"}`, 200},
		{`{"success":false,"data":[]}`, 200},
		{`<html>private internal stack trace</html>`, 200},
		{`{"error":{"code":"rate_limited","message":"private https://internal.example/secret"}}`, 429},
	} {
		result, output, err := runModelQueryFixture(t, nil, test.body, test.status)
		if err == nil || result["ok"] != false || strings.Contains(output, "internal.example") || strings.Contains(output, "<html>") {
			t.Fatalf("unsafe or falsely successful catalog failure: %s", output)
		}
	}
}
