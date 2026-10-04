package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type submissionIdentityCase struct {
	name, body, code string
}

func submissionIdentityCases() []submissionIdentityCase {
	return []submissionIdentityCase{
		{"invalid-json", `<html>private-response-marker</html>`, "task_response_unreadable"},
		{"truncated-json", `{"id":"private-response-marker"`, "task_response_unreadable"},
		{"trailing-json", `{"id":"private-response-marker"} {}`, "task_response_unreadable"},
		{"non-object", `["private-response-marker"]`, "task_response_unreadable"},
		{"null-object", `null`, "task_response_unreadable"},
		{"short-http-body", `{"id":"private-response-marker","status":"pending"}`, "task_response_unreadable"},
		{"missing-id", `{"status":"pending","message":"private-response-marker"}`, "task_id_missing"},
		{"empty-id", `{"task_id":"","id":null,"status":"pending"}`, "task_id_missing"},
		{"nested-id", `{"data":{"id":"private-response-marker"},"status":"pending"}`, "task_id_missing"},
		{"number-id", `{"id":123,"status":"pending"}`, "task_id_invalid"},
		{"number-primary-id", `{"task_id":123,"id":"safe-fallback","status":"pending"}`, "task_id_invalid"},
		{"object-id", `{"id":{"private-response-marker":true},"status":"pending"}`, "task_id_invalid"},
		{"unsafe-id", `{"id":"https://private-response-marker.example/task","status":"pending"}`, "task_id_invalid"},
		{"invalid-primary-id", `{"task_id":"private-response-marker/unsafe","id":"safe-fallback","status":"pending"}`, "task_id_invalid"},
		{"long-id", `{"id":"` + strings.Repeat("a", 257) + `","status":"pending"}`, "task_id_invalid"},
	}
}

// Run both submission entrances: an ordinary receipt and the explicit recovery
// record. A broken success response must never trigger polling or a second POST.
func TestSubmissionIdentityDiagnostics(t *testing.T) {
	for _, kind := range []string{"image", "video"} {
		for _, recorded := range []bool{false, true} {
			for _, scenario := range submissionIdentityCases() {
				name := kind + "/" + scenario.name
				if recorded {
					name += "/recorded"
				}
				t.Run(name, func(t *testing.T) {
					calls := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls++
						if r.Method != http.MethodPost {
							t.Error("diagnostic must not issue follow-up requests")
						}
						io.Copy(io.Discard, r.Body)
						w.Header().Set("X-Request-ID", "55a3d6e3-42dd-4e67-8b08-e5c6058b2f99")
						w.Header().Set("Retry-After", "7")
						if scenario.name == "short-http-body" {
							w.Header().Set("Content-Length", "4096")
						}
						io.WriteString(w, scenario.body)
					}))
					defer server.Close()
					request := taskRequest{Kind: kind, Operation: "generate", Model: "fixture-model",
						Prompt: "private-prompt-marker"}
					var output bytes.Buffer
					writer := &supportReceiptWriter{output: &output, command: "submit", host: "codex"}
					svc := fixtureService(server)
					svc.support = writer
					var err error
					if recorded {
						recordPath := filepath.Join(t.TempDir(), "task.json")
						err = executeRecordedTask("submit", recordPath, request, 0, "", writer, svc)
						record, loadErr := loadTaskRecord(recordPath)
						if loadErr != nil || record.TaskID != "" || record.SubmissionOutcome != "unknown" {
							t.Fatal("record must retain unknown acceptance without a fabricated task ID")
						}
					} else {
						err = executePreparedTask(writer, svc, request)
					}
					got := decodeReceipt(t, &output)
					if err == nil || calls != 1 || got.OK || got.TaskID != "" ||
						got.SubmissionOutcome != "unknown" || got.FailurePhase != "submission" ||
						got.HTTPStatus != 200 || got.NextStep != "await_user" ||
						got.RetryAfterSecs != 7 || got.RetryNotBefore == "" {
						t.Fatalf("invalid submission boundary: %+v (%d calls)", got, calls)
					}
					if got.LocalErrorCode != scenario.code || got.APIErrorCode != "" {
						t.Fatalf("want local %q, got local %q API %q", scenario.code, got.LocalErrorCode, got.APIErrorCode)
					}
					if strings.Contains(output.String(), "private-response-marker") ||
						strings.Contains(output.String(), "private-prompt-marker") {
						t.Fatal("diagnostic exposed response or prompt")
					}
					var envelope struct {
						Support map[string]any `json:"support"`
					}
					if json.Unmarshal(output.Bytes(), &envelope) != nil ||
						envelope.Support["local_error_code"] != scenario.code ||
						envelope.Support["executor_version"] != executorVersion ||
						envelope.Support["api_request_attempted"] != true ||
						envelope.Support["request_id"] != "55a3d6e3-42dd-4e67-8b08-e5c6058b2f99" {
						t.Fatal("support summary lost diagnosis or request correlation")
					}
				})
			}
		}
	}
}

func TestSubmissionHTTPFailureKeepsItsOwnCategory(t *testing.T) {
	for _, status := range []int{400, 429, 502} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			io.WriteString(w, `<html>private-response-marker</html>`)
		}))
		var output bytes.Buffer
		err := executePreparedTask(&output, fixtureService(server),
			taskRequest{Kind: "video", Operation: "generate", Model: "fixture-model", Prompt: "fixture"})
		server.Close()
		got := decodeReceipt(t, &output)
		want := "rejected"
		if status >= 500 {
			want = "unknown"
		}
		if err == nil || got.HTTPStatus != status || got.LocalErrorCode != "" || got.SubmissionOutcome != want {
			t.Fatal("non-success response was misclassified as an identity compatibility issue")
		}
	}
}

func TestSubmissionIdentityValidPrecedence(t *testing.T) {
	for _, body := range []string{
		`{"id":"video-seedance-2.0-abcdefghijklmnop","status":"queued"}`,
		`{"task_id":"video-seedance-2.0-abcdefghijklmnop","id":"other-id","status":"queued"}`,
		`{"task_id":"","id":"video-seedance-2.0-abcdefghijklmnop","status":"queued"}`,
	} {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			io.WriteString(w, body)
		}))
		var output bytes.Buffer
		err := executePreparedTask(&output, fixtureService(server),
			taskRequest{Kind: "video", Operation: "generate", Model: "seedance-2.0", Prompt: "fixture"})
		server.Close()
		got := decodeReceipt(t, &output)
		if err != nil || calls != 1 || !got.OK || got.TaskID != "video-seedance-2.0-abcdefghijklmnop" ||
			got.LocalErrorCode != "" || got.NextStep != "wait" || got.SubmissionOutcome != "accepted" {
			t.Fatal("diagnostic changed accepted task identity or performed follow-up work")
		}
	}
}

func TestExistingTaskIdentityDiagnosticsPreserveTask(t *testing.T) {
	for _, kind := range []string{"image", "video"} {
		for _, command := range []string{"status", "wait"} {
			for _, scenario := range submissionIdentityCases() {
				switch scenario.name {
				case "invalid-json", "short-http-body", "non-object", "missing-id", "unsafe-id":
				default:
					continue
				}
				t.Run(kind+"/"+command+"/"+scenario.name, func(t *testing.T) {
					calls := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls++
						if r.Method != http.MethodGet || r.URL.Path != statusPath(kind, "paid-task") {
							t.Error("recovery must only read the original task")
						}
						if scenario.name == "short-http-body" {
							w.Header().Set("Content-Length", "4096")
						}
						io.WriteString(w, scenario.body)
					}))
					defer server.Close()
					var output bytes.Buffer
					input := `{"kind":"` + kind + `","task_id":"paid-task","model":"fixture-model","original_operation":"generate","task_status":"pending","wait_windows_completed":1}`
					svc := fixtureService(server)
					svc.wait = func(context.Context, time.Duration) bool { return true }
					err := executeExistingTask(command, strings.NewReader(input), &output, svc)
					got := decodeReceipt(t, &output)
					if err == nil || calls != 1 || got.OK || got.TaskID != "paid-task" ||
						got.Model != "fixture-model" || got.OriginalOperation != "generate" ||
						got.WaitWindowsCompleted != 1 || got.FailurePhase != "status" ||
						got.HTTPStatus != 200 || got.NextStep != "await_user" ||
						got.SubmissionOutcome == "unknown" {
						t.Fatalf("status failure changed original task or continued automatically: %+v (%d calls)", got, calls)
					}
					if got.LocalErrorCode != scenario.code || got.APIErrorCode != "" ||
						!strings.Contains(got.NextAction, "Keep this task ID") {
						t.Fatalf("status diagnosis lost category or same-task guidance: %+v", got)
					}
					if strings.Contains(output.String(), "private-response-marker") {
						t.Fatal("status diagnostic exposed the response")
					}
				})
			}
		}
	}
}
