package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskRequestBodyUsesFixedImageAsyncContract(t *testing.T) {
	path, contentType, body, err := taskRequestBody(taskRequest{
		Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "test", Parameters: map[string]any{"n": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/v1/images/generations" || contentType != "application/json" {
		t.Fatalf("unexpected route: %s %s", path, contentType)
	}
	var payload map[string]any
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["async"] != true || payload["model"] != "gpt-image-2" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestServiceDoesNotRetrySubmission(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"code":"upstream_unavailable","message":"try later"}}`))
	}))
	defer server.Close()
	svc := service{baseURL: server.URL, token: "redacted", client: server.Client()}
	_, status, _, code, _, err := svc.request(context.Background(), http.MethodPost, "/v1/images/generations", strings.NewReader(`{}`), "application/json")
	if err != nil || status != http.StatusBadGateway || code != "upstream_unavailable" || calls != 1 {
		t.Fatalf("status=%d code=%q calls=%d err=%v", status, code, calls, err)
	}
}

func TestPublicAPIErrorRemovesURLs(t *testing.T) {
	code, message := publicAPIError([]byte(`{"error":{"code":"invalid_request","message":"see https://internal.example"}}`))
	if code != "invalid_request" || message != "" {
		t.Fatalf("unexpected public error: %q %q", code, message)
	}
}

func TestMultipartTaskStreamsOnlyTheCurrentAttachment(t *testing.T) {
	directory := t.TempDir()
	attachmentPath := filepath.Join(directory, "reference.png")
	if err := os.WriteFile(attachmentPath, []byte("current-request-only"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, contentType, body, err := taskRequestBody(taskRequest{
		Kind: "image", Operation: "edit", Model: "gpt-image-2", Prompt: "edit", Attachments: []attachment{{Field: "image", Path: attachmentPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data;") || !bytes.Contains(encoded, []byte("current-request-only")) {
		t.Fatalf("multipart body did not contain the declared attachment")
	}
}

func TestContinuationOnlyReadsTheSuppliedTask(t *testing.T) {
	posts := 0
	reads := 0
	contents := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/images/task-1":
			reads++
			_, _ = w.Write([]byte(`{"task_id":"task-1","status":"completed"}`))
		case "/v1/images/task-1/content":
			contents++
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("png"))
		default:
			if r.Method == http.MethodPost {
				posts++
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	output := &bytes.Buffer{}
	err := executeTask(strings.NewReader(`{"kind":"image","operation":"continue","task_id":"task-1","requested_count":1,"output_dir":"`+strings.ReplaceAll(t.TempDir(), `\`, `\\`)+`","poll":{"max_status_reads":1,"deadline_seconds":1}}`), output, service{baseURL: server.URL, token: "not-output", client: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if posts != 0 || reads != 1 || contents != 1 {
		t.Fatalf("posts=%d reads=%d contents=%d", posts, reads, contents)
	}
	if strings.Contains(output.String(), "not-output") || !strings.Contains(output.String(), `"task_id":"task-1"`) {
		t.Fatalf("unsafe or incomplete receipt: %s", output.String())
	}
}

func TestReconciliationStopsBeforeContentDelivery(t *testing.T) {
	contentReads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/videos/task-1" {
			_, _ = w.Write([]byte(`{"task_id":"task-1","status":"completed","reconciliation_required":true}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/content") {
			contentReads++
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	output := &bytes.Buffer{}
	err := executeTask(strings.NewReader(`{"kind":"video","operation":"continue","task_id":"task-1","poll":{"max_status_reads":1,"deadline_seconds":1}}`), output, service{baseURL: server.URL, token: "not-output", client: server.Client()})
	if err == nil || contentReads != 0 || !strings.Contains(output.String(), `"reconciliation_required":true`) {
		t.Fatalf("err=%v contentReads=%d receipt=%s", err, contentReads, output.String())
	}
}

func TestDownloadRejectsNonMediaContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("not media"))
	}))
	defer server.Close()
	_, _, _, _, _, err := (service{baseURL: server.URL, token: "not-output", client: server.Client()}).download(context.Background(), "/v1/images/task-1/content?index=0", "image", t.TempDir())
	if err == nil {
		t.Fatal("expected non-media content to be rejected")
	}
}

func TestCredentialAdaptersUseActiveHostConnectionFiles(t *testing.T) {
	directory := t.TempDir()
	write := func(name, contents string) string {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	assertCredential := func(name string, resolve func(string) (string, error), contents string) {
		t.Helper()
		key, err := resolve(write(name, contents))
		if err != nil || key != "test-key" {
			t.Fatalf("%s key=%q err=%v", name, key, err)
		}
	}

	// Codex resolves the active provider table from config.toml. The provider
	// name is intentionally arbitrary so a CC Switch connection works too.
	assertCredential("codex.toml", credentialFromCodexFile, `model_provider = "cc-switch"
[model_providers.cc-switch]
base_url = "https://api.puretokensx.com/v1"
experimental_bearer_token = "test-key"
`)
	assertCredential("claude.json", credentialFromClaudeCodeFile, `{"env":{"ANTHROPIC_BASE_URL":"https://api.puretokensx.com","ANTHROPIC_AUTH_TOKEN":"test-key"}}`)
	assertCredential("gemini.env", credentialFromGeminiEnvFile, "GOOGLE_GEMINI_BASE_URL=https://api.puretokensx.com\nGEMINI_API_KEY=test-key\n")
	assertCredential("workbuddy.json", credentialFromWorkBuddyFile, `[{"url":"https://api.puretokensx.com/v1/chat/completions","apiKey":"test-key"}]`)
	assertCredential("grok.toml", credentialFromGrokBuildFile, `[models]
default = "chosen"
[model.chosen]
base_url = "https://api.puretokensx.com/v1"
api_key = "test-key"
`)
	assertCredential("opencode.json", credentialFromOpenCodeFile, `{"model":"cc-switch/gpt-5","provider":{"cc-switch":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"test-key"}}}}`)
}

func TestCredentialAdaptersRejectInactiveOrNonPureTokensConnection(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "config.toml")
	if err := os.WriteFile(config, []byte(`model_provider = "other"
[model_providers.puretokens]
base_url = "https://api.puretokensx.com/v1"
experimental_bearer_token = "unused-key"
[model_providers.other]
base_url = "https://other.example/v1"
experimental_bearer_token = "wrong-key"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := credentialFromCodexFile(config); err == nil {
		t.Fatal("expected inactive non-Pure-Tokens Codex connection to be rejected")
	}
	if matchesPureTokensEndpoint("https://api.puretokensx.com/v1/anything", "/v1") {
		t.Fatal("expected unsupported endpoint path to be rejected")
	}
}

func TestInitVerifiesIdentityWithoutReturningConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1" || r.Method != http.MethodGet {
			t.Fatal("unexpected init request")
		}
		_, _ = w.Write([]byte(`{"status":"ok","name":"Pure Tokens API","base_url":"/v1"}`))
	}))
	defer server.Close()

	output := &bytes.Buffer{}
	if err := executeInit(output, service{baseURL: server.URL, token: "secret-must-not-appear", client: server.Client()}); err != nil {
		t.Fatal(err)
	}
	var result initReceipt
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.ConfigurationStatus != "verified" || !result.APIRequestExecuted || !result.APIIdentityConfirmed || len(result.UsageExamples) == 0 {
		t.Fatalf("unexpected init receipt: %#v", result)
	}
	if strings.Contains(output.String(), "secret-must-not-appear") || strings.Contains(output.String(), apiOrigin) {
		t.Fatalf("init receipt exposed sensitive or internal connection data: %s", output.String())
	}
}
