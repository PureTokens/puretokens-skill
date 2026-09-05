package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

type schemaExample struct {
	Name     string          `json:"name"`
	Schema   string          `json:"schema"`
	Document json.RawMessage `json:"document"`
}

// Node consumes actual JSON emitted by the executor, not hand-maintained
// replicas of its structs. Every request uses local fixtures and local mocks.
// The test also runs its behavioral checks when invoked directly by go test.
func TestContractExamples(t *testing.T) {
	var examples []schemaExample
	capture := func(name, schema string, data []byte) {
		t.Helper()
		if !json.Valid(data) {
			t.Fatalf("%s did not emit a JSON document", name)
		}
		examples = append(examples, schemaExample{name, schema, append(json.RawMessage(nil), data...)})
	}
	captureValue := func(name, schema string, value any) {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		capture(name, schema, data)
	}

	t.Run("documented media requests execute", func(t *testing.T) {
		pngPath := filepath.Join(t.TempDir(), "current.png")
		if err := os.WriteFile(pngPath, fixturePNG(t), 0600); err != nil {
			t.Fatal(err)
		}
		fences := regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n```")
		for _, kind := range []string{"image", "video"} {
			document, err := os.ReadFile(filepath.Join("../../skills", "puretokens-"+kind, "references/executor-usage.md"))
			if err != nil {
				t.Fatal(err)
			}
			blocks := fences.FindAllSubmatch(document, -1)
			if len(blocks) != 4 {
				t.Fatalf("%s guide: expected generation, attachment, status and content examples", kind)
			}
			for index, block := range blocks {
				name := fmt.Sprintf("docs-%s-%d", kind, index)
				capture(name, "executor-request.schema.json", block[1])
				request, err := decodeTaskRequest(bytes.NewReader(block[1]))
				if err != nil || request.Kind != kind {
					t.Fatalf("%s is not a kind-correct native request", name)
				}
				if request.TaskID != "" {
					request.Operation = "continue"
				}
				for index := range request.Attachments {
					request.Attachments[index].Path = pngPath
				}
				if request.OutputDir != "" {
					request.OutputDir = t.TempDir()
				}
				if err := validateTaskRequest(request); err != nil {
					t.Fatalf("%s fails native request validation: %v", name, err)
				}
				if request.Operation == "continue" {
					continue
				}
				if err := prepareProfileRequest(&request, profileService()); err != nil {
					t.Fatalf("%s is incompatible with its installed model profile: %v", name, err)
				}
				route := endpointFor(request)
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != http.MethodPost || r.URL.Path != route {
						t.Error("documented submission did an unexpected request")
					}
					if len(request.Attachments) > 0 {
						if err := r.ParseMultipartForm(1 << 20); err != nil {
							t.Error(err)
						} else {
							defer r.MultipartForm.RemoveAll()
							if len(r.MultipartForm.File[request.Attachments[0].Field]) != 1 {
								t.Error("documented native attachment was not sent")
							}
						}
					} else {
						io.Copy(io.Discard, r.Body)
					}
					io.WriteString(w, `{"id":"fixture-doc-task","status":"queued"}`)
				}))
				var output bytes.Buffer
				body, _ := json.Marshal(request)
				err = executeTask(bytes.NewReader(body), &output, fixtureService(server))
				server.Close()
				if err != nil || calls != 1 {
					t.Fatalf("%s did not submit once without status/content reads: %v (%d calls)", name, err, calls)
				}
				capture(name+"-receipt", "executor-receipt.schema.json", output.Bytes())
			}
		}
	})

	t.Run("recorded lifecycle and delivery", func(t *testing.T) {
		pngBytes := fixturePNG(t)
		var calls []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, r.Method+" "+r.URL.RequestURI())
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/v1/images/generations":
				io.WriteString(w, `{"id":"fixture-record-task","status":"queued"}`)
			case r.Method == http.MethodGet && r.URL.Path == "/v1/images/fixture-record-task":
				io.WriteString(w, `{"id":"fixture-record-task","status":"completed"}`)
			case r.Method == http.MethodGet && r.URL.Path == "/v1/images/fixture-record-task/content":
				w.Header().Set("Content-Type", "image/png")
				w.Write(pngBytes)
			default:
				t.Error("record lifecycle used an unexpected request")
				http.Error(w, "unexpected route", 500)
			}
		}))
		defer server.Close()
		recordPath := filepath.Join(t.TempDir(), "task.json")
		outputDir := t.TempDir()
		request := taskRequest{Kind: "image", Operation: "generate", Model: "seedream-5.0-pro",
			Prompt: "synthetic-private-prompt", Parameters: map[string]any{"n": 2.0, "size": "2048x2048", "image_urls": []any{"https://reference.example.test/private.png"}}}
		steps := []struct {
			command string
			index   int
		}{{"submit", 0}, {"resume", 0}, {"content", 0}, {"delivered", 0}, {"content", 1}, {"delivered", 1}}
		for index, step := range steps {
			var output bytes.Buffer
			dir := ""
			if step.command == "content" {
				dir = outputDir
			}
			if err := executeRecordedTask(step.command, recordPath, request, step.index, dir, &output, fixtureService(server)); err != nil {
				t.Fatalf("%s: %v", step.command, err)
			}
			name := fmt.Sprintf("record-%d-%s", index, step.command)
			capture(name+"-receipt", "executor-receipt.schema.json", output.Bytes())
			recordData, err := os.ReadFile(recordPath)
			if err != nil {
				t.Fatal(err)
			}
			capture(name+"-artifact", "task-record.schema.json", recordData)
			for _, secret := range []string{"synthetic-private-prompt", "reference.example.test", "synthetic-fixture-token", `"prompt"`} {
				if bytes.Contains(recordData, []byte(secret)) || bytes.Contains(output.Bytes(), []byte(secret)) {
					t.Fatalf("%s retained private request data", name)
				}
			}
			result := decodeReceipt(t, &output)
			if result.TaskID != "fixture-record-task" || result.OriginalOperation != "generate" || result.RequestedCount != 2 {
				t.Fatalf("%s lost original task context", name)
			}
			if step.command == "delivered" {
				expected := "partially_delivered"
				if step.index == 1 {
					expected = "delivered"
				}
				if result.DeliveryStatus != expected || len(result.DeliveredIndexes) != step.index+1 {
					t.Fatal("delivery acknowledgement lost progress")
				}
			}
		}
		if !reflect.DeepEqual(calls, []string{
			"POST /v1/images/generations", "GET /v1/images/fixture-record-task",
			"GET /v1/images/fixture-record-task/content?index=0", "GET /v1/images/fixture-record-task/content?index=1",
		}) {
			t.Fatalf("record continuation resubmitted, redownloaded or used extra requests: %v", calls)
		}
	})

	t.Run("submission failure and retry receipts", func(t *testing.T) {
		for _, status := range []int{400, 503} {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(status)
				io.WriteString(w, `{"error":{"code":"fixture_failure","message":"The request failed."}}`)
			}))
			var output bytes.Buffer
			err := executeTask(strings.NewReader(`{"kind":"image","operation":"generate","model":"gpt-image-2","prompt":"fixture"}`), &output, fixtureService(server))
			server.Close()
			if err == nil || calls != 1 {
				t.Fatal("failed submission was not one-shot")
			}
			capture(fmt.Sprintf("submission-%d", status), "executor-receipt.schema.json", output.Bytes())
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(429)
			io.WriteString(w, `{"error":{"code":"rate_limit","message":"Read this task later."}}`)
		}))
		defer server.Close()
		svc := fixtureService(server)
		svc.now = func() time.Time { return time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC) }
		var output bytes.Buffer
		err := executeExistingTask("status", strings.NewReader(`{"kind":"image","task_id":"fixture-retry","original_operation":"image_edit","model":"gpt-image-2","requested_count":2}`), &output, svc)
		if err == nil || decodeReceipt(t, &output).RetryNotBefore != "2026-09-05T00:00:10Z" {
			t.Fatal("status retry did not retain its absolute bound")
		}
		capture("status-retry", "executor-receipt.schema.json", output.Bytes())
	})

	t.Run("explicit validation", func(t *testing.T) {
		svc := profileService()
		svc.client = &http.Client{Transport: schemaRejectNetwork{t}}
		var output bytes.Buffer
		request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "fixture"}
		if err := executePreflight(&output, svc, request); err != nil {
			t.Fatal(err)
		}
		capture("validation-success", "executor-receipt.schema.json", output.Bytes())
		output.Reset()
		request.Parameters = map[string]any{"n": 7.0}
		if err := executePreflight(&output, svc, request); err == nil {
			t.Fatal("invalid count passed validation")
		}
		capture("validation-failure", "executor-receipt.schema.json", output.Bytes())
		output.Reset()
		request.Model = "seedream-5.0-pro"
		request.Parameters = map[string]any{"width": true, "height": "large", "strength": false}
		if err := executePreflight(&output, svc, request); err == nil {
			t.Fatal("invalid parameter types passed validation")
		}
		capture("validation-invalid-types", "executor-receipt.schema.json", output.Bytes())
	})

	t.Run("credential failure retains continuation context", func(t *testing.T) {
		var output bytes.Buffer
		err := run([]string{"status", "--host", "unsupported-fixture-host"},
			strings.NewReader(`{"kind":"image","task_id":"fixture-existing","original_operation":"image_edit","model":"gpt-image-2","requested_count":1}`), &output)
		if err == nil {
			t.Fatal("unsupported fixture host unexpectedly resolved credentials")
		}
		result := decodeReceipt(t, &output)
		if result.Operation != "continue" || result.TaskID != "fixture-existing" || result.OriginalOperation != "image_edit" {
			t.Fatal("credential failure lost continuation context")
		}
		capture("continuation-credential-failure", "executor-receipt.schema.json", output.Bytes())
	})

	t.Run("balance usage snapshot and published units", func(t *testing.T) {
		var calls []string
		svc := balanceFixtureService(t, func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, r.Method+" "+r.URL.Path)
			if r.URL.Path == balanceUsagePath {
				io.WriteString(w, balanceUsageFixture)
			} else {
				io.WriteString(w, balanceUnitFixture)
			}
		})
		var output bytes.Buffer
		if err := executeBalance(&output, svc); err != nil {
			t.Fatal(err)
		}
		capture("balance", "balance-receipt.schema.json", output.Bytes())
		var envelope struct{ Result json.RawMessage }
		if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		capture("balance-projection", "balance-snapshot.schema.json", envelope.Result)
		var result map[string]any
		json.Unmarshal(envelope.Result, &result)
		if result["total"] != 29.672244 || result["used"] != 19.672244 || result["remaining"] != float64(10) ||
			!reflect.DeepEqual(calls, []string{"GET " + balanceUsagePath, "GET " + balanceUnitPath}) {
			t.Fatal("balance did not project actual available quota and units correctly")
		}
	})

	t.Run("documented model filter", func(t *testing.T) {
		document, err := os.ReadFile("../../skills/puretokens-models/SKILL.md")
		if err != nil {
			t.Fatal(err)
		}
		match := regexp.MustCompile("`(\\{\"kind\":\"video\",\"operation\":\"image_to_video\",\"parameters\":\\{[^`]+\\}\\})`").FindSubmatch(document)
		if len(match) != 2 {
			t.Fatal("model Skill does not contain its executable filter example")
		}
		capture("models-filter", "model-query.schema.json", match[1])
		for _, filtered := range []bool{false, true} {
			var input io.Reader
			if filtered {
				input = bytes.NewReader(match[1])
			}
			result, output, err := runModelQueryFixture(t, input, modelQueryFixture, 200)
			if err != nil {
				t.Fatal(err)
			}
			if filtered && !reflect.DeepEqual(modelQueryIDs(t, result), []string{"video-audio"}) {
				t.Fatal("documented filter did not match declared capabilities")
			}
			capture(fmt.Sprintf("models-%t", filtered), "model-query-receipt.schema.json", []byte(output))
		}
	})

	t.Run("doctor and init", func(t *testing.T) {
		home := doctorIsolatedHome(t)
		root := filepath.Join(home, ".agents", "skills")
		doctorFixtureInstallation(t, root, executorVersion)
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.Method != http.MethodGet {
				t.Error("doctor created a task")
			}
			if r.URL.Path == "/v1" {
				io.WriteString(w, `{"status":"ok","name":"Pure Tokens API","base_url":"/v1"}`)
			} else if r.URL.Path == "/v1/media/models" {
				io.WriteString(w, `{"data":[]}`)
			} else {
				t.Error("doctor used an unexpected route")
			}
		}))
		defer server.Close()
		svc := service{baseURL: server.URL, client: server.Client(), token: "synthetic-doctor-fixture", profilesRoot: root}
		var output bytes.Buffer
		if err := executeDoctor(&output, svc, "codex"); err != nil {
			t.Fatal(err)
		}
		if calls != 2 {
			t.Fatal("doctor did not use exactly the two read-only checks")
		}
		capture("doctor-success", "doctor-receipt.schema.json", output.Bytes())
		var result doctorReceipt
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		captureValue("init-success", "init-receipt.schema.json", result.Connection)
		output.Reset()
		svc.token = ""
		if err := executeDoctorCredentialFailure(&output, svc, "codex", errors.New("synthetic-private-error")); err == nil || calls != 2 {
			t.Fatal("doctor credential failure used network")
		}
		capture("doctor-no-credential", "doctor-receipt.schema.json", output.Bytes())
		captureValue("init-no-credential", "init-receipt.schema.json", initCredentialFailure(errors.New("synthetic-private-error")))
	})

	if outputPath := os.Getenv("PURETOKENS_SCHEMA_EXAMPLES_OUT"); outputPath != "" {
		data, err := json.MarshalIndent(examples, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(outputPath, append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

type schemaRejectNetwork struct{ t *testing.T }

func (transport schemaRejectNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	transport.t.Error("installed-profile validation attempted network")
	return nil, errors.New("network forbidden in this fixture")
}
