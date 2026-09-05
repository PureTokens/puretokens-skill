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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestContinuationUnknownCountIsNotInvented(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		io.WriteString(w, `{"id":"paid-task","status":"completed","n":3}`)
	}))
	defer server.Close()
	var out bytes.Buffer
	err := executeExistingTask("status", strings.NewReader(`{"kind":"image","task_id":"paid-task"}`), &out, fixtureService(server))
	result := decodeReceipt(t, &out)
	if err != nil || result.RequestedCount != 0 || calls != 1 {
		t.Fatalf("invented metadata: %+v, %v", result, err)
	}
	request := taskRequest{Kind: "image", TaskID: "paid-task", TaskStatus: "completed", OutputDir: t.TempDir()}
	data, _ := json.Marshal(request)
	out.Reset()
	err = executeExistingTask("content", bytes.NewReader(data), &out, fixtureService(server))
	result = decodeReceipt(t, &out)
	if err == nil || result.LocalErrorCode != "original_count_unknown" || result.TaskID != "paid-task" || calls != 1 {
		t.Fatalf("unknown count downloaded: %+v", result)
	}
}

func TestContinuationPreservesMetadataOnLocalFailure(t *testing.T) {
	request := taskRequest{Kind: "image", TaskID: "paid-task", Model: "gpt-image-2", TaskStatus: "completed", RequestedCount: 3, OriginalOperation: "image_edit", OutputDir: filepath.Join(t.TempDir(), "deleted")}
	data, _ := json.Marshal(request)
	var out bytes.Buffer
	err := executeExistingTask("content", bytes.NewReader(data), &out, service{})
	result := decodeReceipt(t, &out)
	if err == nil || result.TaskID != request.TaskID || result.Model != request.Model || result.RequestedCount != 3 || result.Status != "completed" || result.OriginalOperation != "image_edit" || result.LocalErrorCode != "output_directory_unavailable" {
		t.Fatalf("context lost: %+v", result)
	}
	out.Reset()
	// An unsupported synthetic host fails before any real configuration lookup.
	err = run([]string{"status", "--host", "fixture-unsupported"}, bytes.NewReader(data), &out)
	result = decodeReceipt(t, &out)
	if err == nil || result.TaskID != request.TaskID || result.Model != request.Model || result.RequestedCount != 3 || result.OriginalOperation != "image_edit" {
		t.Fatalf("credential error lost context: %+v", result)
	}
}

func TestPollingRetryOverrideConsumedOnce(t *testing.T) {
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		if reads == 4 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			io.WriteString(w, `{"error":{"message":"Wait"}}`)
			return
		}
		io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
	}))
	defer server.Close()
	svc := fixtureService(server)
	now := time.Now().Truncate(time.Second)
	svc.now = func() time.Time { return now }
	delays := []time.Duration{}
	svc.wait = func(_ context.Context, d time.Duration) bool {
		delays = append(delays, d)
		now = now.Add(d)
		return true
	}
	var out bytes.Buffer
	request := taskRequest{Kind: "image", TaskID: "paid-task"}
	pollAndDeliver(&out, svc, request, taskReceipt(request, request.TaskID, "pending"))
	want := []time.Duration{3 * time.Second, 6 * time.Second, 12 * time.Second, 24 * time.Second, time.Second, 30 * time.Second}
	if reads != 6 || !reflect.DeepEqual(delays, want) {
		t.Fatalf("reads %d delays %v", reads, delays)
	}
}

func TestLongRetryRetainedAcrossCommands(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "3600")
		if r.Method == "POST" {
			io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
			return
		}
		w.WriteHeader(429)
		io.WriteString(w, `{"error":{"message":"Please wait"}}`)
	}))
	defer server.Close()
	svc := fixtureService(server)
	svc.wait = func(_ context.Context, d time.Duration) bool { return d <= time.Minute }
	var out bytes.Buffer
	executeTask(strings.NewReader(`{"kind":"image","operation":"generate","model":"gpt-image-2","prompt":"test"}`), &out, svc)
	submission := decodeReceipt(t, &out)
	if submission.RetryAfterSecs != 3600 || retryDelay(submission.RetryNotBefore, time.Now()) < 3599*time.Second {
		t.Fatal("submission retry omitted")
	}
	out.Reset()
	request := taskRequest{Kind: "image", TaskID: "paid-task", RetryNotBefore: submission.RetryNotBefore}
	data, _ := json.Marshal(request)
	executeExistingTask("status", bytes.NewReader(data), &out, svc)
	if calls != 1 || decodeReceipt(t, &out).LocalErrorCode != "retry_wait_required" {
		t.Fatal("read before retry time")
	}
	out.Reset()
	request.RetryNotBefore = ""
	pollAndDeliver(&out, svc, request, taskReceipt(request, request.TaskID, "pending"))
	result := decodeReceipt(t, &out)
	if calls != 2 || result.HTTPStatus != 429 || result.RetryAfterSecs != 3600 || result.RetryNotBefore == "" || result.APIErrorCode != "" || result.ErrorMessage != "Please wait" {
		t.Fatalf("long retry lost: %+v", result)
	}
}

func TestTaskRecordLifecycleAndExplicitDelivery(t *testing.T) {
	counts := map[string]int{}
	image := fixturePNG(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counts[r.Method]++
		if r.Method == "POST" {
			io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/content") {
			w.Header().Set("Content-Type", "image/png")
			w.Write(image)
			return
		}
		io.WriteString(w, `{"id":"paid-task","status":"completed"}`)
	}))
	defer server.Close()
	svc := fixtureService(server)
	svc.wait = func(context.Context, time.Duration) bool { return true }
	dir := t.TempDir()
	path := filepath.Join(dir, "task.json")
	request := taskRequest{Kind: "image", Operation: "generate", Model: "seedream-5.0-pro", Prompt: "private-prompt-marker", Parameters: map[string]any{"n": float64(3)}, OutputDir: dir}
	var out bytes.Buffer
	if err := executeRecordedTask("submit", path, request, 0, "", &out, svc); err != nil {
		t.Fatal(err, out.String())
	}
	data, _ := os.ReadFile(path)
	if bytes.Contains(data, []byte("private-prompt-marker")) || bytes.Contains(data, []byte(svc.token)) {
		t.Fatal("private data in record")
	}
	record, err := loadTaskRecord(path)
	if err != nil || record.RequestedCount != 3 || record.TaskID != "paid-task" {
		t.Fatalf("bad record %+v %v", record, err)
	}
	out.Reset()
	if err := executeRecordedTask("submit", path, request, 0, "", &out, svc); err == nil || counts["POST"] != 1 {
		t.Fatal("record allowed second POST")
	}
	out.Reset()
	if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, svc); err != nil || counts["POST"] != 1 {
		t.Fatal("resume resubmitted", err)
	}
	out.Reset()
	if err := executeRecordedTask("delivered", path, taskRequest{}, 1, "", &out, svc); err == nil {
		t.Fatal("undownloaded index marked delivered")
	}
	for index := 0; index < 3; index++ {
		out.Reset()
		if err := executeRecordedTask("content", path, taskRequest{}, index, dir, &out, svc); err != nil {
			t.Fatal(err, out.String())
		}
		result := decodeReceipt(t, &out)
		if result.DeliveryStatus != "downloaded_awaiting_host_delivery" {
			t.Fatal("download claimed delivered")
		}
		out.Reset()
		before := counts["GET"]
		if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, svc); err != nil {
			t.Fatal(err)
		}
		result = decodeReceipt(t, &out)
		if counts["GET"] != before || len(result.DownloadedIndexes) != 1 || result.DownloadedIndexes[0] != index || len(result.DownloadedPaths) != 1 {
			t.Fatalf("resume lost pending handoff %+v", result)
		}
		out.Reset()
		if err := executeRecordedTask("delivered", path, taskRequest{}, index, "", &out, svc); err != nil {
			t.Fatal(err)
		}
	}
	result := decodeReceipt(t, &out)
	if result.DeliveryStatus != "delivered" || len(result.DeliveredIndexes) != 3 || counts["POST"] != 1 {
		t.Fatalf("bad delivery %+v", result)
	}
	record, _ = loadTaskRecord(path)
	if len(record.Downloaded) != 3 || len(record.Delivered) != 3 {
		t.Fatal("record lost outputs")
	}
	if _, err := os.Stat(path + ".lock"); !os.IsNotExist(err) {
		t.Fatal("lock retained")
	}
}

func TestRecordDoesNotProjectArbitraryLocalFileAsDownloaded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "task.json")
	other := filepath.Join(dir, "unrelated.png")
	if err := os.WriteFile(other, fixturePNG(t), 0600); err != nil {
		t.Fatal(err)
	}
	record := taskRecord{Format: taskRecordFormat, Kind: "image", TaskID: "paid-task", RequestedCount: 1, Downloaded: map[int]string{0: other}}
	if err := saveTaskRecord(path, record, true); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTaskRecord(path); err == nil {
		t.Fatal("arbitrary file treated as downloaded task output")
	}
}

func TestUnknownRecordAndLockedRecordNeverSubmit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("network request not expected") }))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "unknown.json")
	record := taskRecord{Format: taskRecordFormat, Kind: "video", SubmissionOutcome: "unknown"}
	if err := saveTaskRecord(path, record, true); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, fixtureService(server)); err == nil {
		t.Fatal("unknown record resumed")
	}
	if !strings.Contains(out.String(), `"submission_outcome":"unknown"`) {
		t.Fatal("unknown outcome lost")
	}
	unlock, err := lockTaskRecord(path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	out.Reset()
	if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, fixtureService(server)); err == nil {
		t.Fatal("locked record modified")
	}
}

func TestNewRecordedSubmissionRejectsCopiedTaskIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("copied identity reached POST")
	}))
	defer server.Close()
	for _, field := range []string{"id", "state", "retry", "operation", "reconcile"} {
		request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "test"}
		switch field {
		case "id":
			request.TaskID = "unrelated-task"
		case "state":
			request.TaskStatus = "completed"
		case "retry":
			request.RetryNotBefore = time.Now().UTC().Format(time.RFC3339)
		case "operation":
			request.OriginalOperation = "image_edit"
		case "reconcile":
			request.ReconciliationRequired = true
		}
		path := filepath.Join(t.TempDir(), "task.json")
		var out bytes.Buffer
		if err := executeRecordedTask("submit", path, request, 0, "", &out, fixtureService(server)); err == nil {
			t.Fatalf("accepted copied %s", field)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("created record with copied %s", field)
		}
	}
}

func TestContentFailureIsActionableAndTimeoutSeparate(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	svc := fixtureService(server)
	svc.downloadTimeout = 20 * time.Millisecond
	request := taskRequest{Kind: "image", TaskID: "paid-task", TaskStatus: "completed", RequestedCount: 1, OutputDir: dir}
	var out bytes.Buffer
	deliverTask(&out, svc, request, taskReceipt(request, request.TaskID, "completed"))
	result := decodeReceipt(t, &out)
	if result.LocalErrorCode != "media_download_timeout" || result.TaskID != "paid-task" {
		t.Fatalf("missing local guidance %+v", result)
	}
	if (service{}).contentTimeout() != 10*time.Minute || svc.client.Timeout != 0 {
		t.Fatal("download timeout changed JSON client")
	}
}

func TestRecordWriteFailureKeepsAcceptedTaskReceipt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task.json")
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := os.Remove(path); err != nil {
			t.Error(err)
		}
		if err := os.Mkdir(path, 0700); err != nil {
			t.Error(err)
		}
		io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
	}))
	defer server.Close()
	request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "test"}
	var out bytes.Buffer
	err := executeRecordedTask("submit", path, request, 0, "", &out, fixtureService(server))
	result := decodeReceipt(t, &out)
	if err == nil || calls != 1 || result.TaskID != "paid-task" || result.SubmissionOutcome != "accepted" || result.LocalErrorCode != "task_record_write_failed" || strings.Count(out.String(), "\n") != 1 {
		t.Fatalf("lost accepted task %+v %v", result, err)
	}
}

func TestExplicitPreflightDoesNotSubmitOrReadOrdinaryCatalog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("ordinary preflight made network request") }))
	defer server.Close()
	var out bytes.Buffer
	request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "test"}
	if err := executePreflight(&out, fixtureService(server), request); err != nil {
		t.Fatal(err)
	}
	result := decodeReceipt(t, &out)
	if !result.OK || result.Operation != "preflight" || result.TaskID != "" || result.SubmissionOutcome != "not_submitted" {
		t.Fatalf("invalid preflight receipt %+v", result)
	}
}
