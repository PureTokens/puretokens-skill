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

func TestPendingWindowsAreSuccessfulAndBoundAutomaticContinuation(t *testing.T) {
	for _, kind := range []string{"image", "video"} {
		request := taskRequest{Kind: kind, TaskID: "paid-task", Poll: &pollRequest{MaxStatusReads: 1}}
		for window := 1; window <= 3; window++ {
			result, reads, err := runPollingRequestTimeline(t, request, func(int, time.Duration) pollingReply { return pollingReply{} })
			wantNext := "await_user"
			if window == 1 {
				wantNext = "wait"
			}
			if err != nil || !result.OK || result.Status != "pending" || result.FailurePhase != "" || result.ErrorMessage != "" ||
				result.WaitOutcome != "window_ended" || result.WaitWindowsCompleted != min(window, 2) || result.NextStep != wantNext || len(reads) != 1 {
				t.Fatalf("window %d: %+v reads=%v err=%v", window, result, reads, err)
			}
			request.WaitWindowsCompleted = result.WaitWindowsCompleted
		}
	}
}

func TestServerWaitAndRealStatusFailureDoNotAutoContinue(t *testing.T) {
	for _, status := range []int{200, 429, 503} {
		result, _, err := runPollingRequestTimeline(t, taskRequest{Kind: "image", TaskID: "paid-task"}, func(int, time.Duration) pollingReply {
			return pollingReply{status: status, retry: 3600}
		})
		if result.NextStep != "await_user" || result.RetryNotBefore == "" {
			t.Fatalf("unsafe next step: %+v", result)
		}
		if status == 200 {
			if err != nil || !result.OK || result.WaitOutcome != "retry_deferred" || result.FailurePhase != "" {
				t.Fatalf("server wait reported failed: %+v %v", result, err)
			}
		} else if err == nil || result.OK || result.FailurePhase != "status" {
			t.Fatalf("real failure hidden: %+v %v", result, err)
		}
	}
}

func TestLastReadRetryStillUsesRemainingWindow(t *testing.T) {
	for _, delay := range []int{1, 3600} {
		result, reads, err := runPollingRequestTimeline(t, taskRequest{Kind: "image", TaskID: "paid-task", Poll: &pollRequest{MaxStatusReads: 1}}, func(int, time.Duration) pollingReply {
			return pollingReply{retry: delay}
		})
		want := "wait"
		if delay == 3600 {
			want = "await_user"
		}
		if err != nil || !result.OK || len(reads) != 1 || result.NextStep != want {
			t.Fatalf("last-read retry %d: %+v %v", delay, result, err)
		}
	}
}

func TestRecordedWindowsPreserveBudgetWithoutAnotherSubmission(t *testing.T) {
	posts, gets := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts++
		} else {
			gets++
		}
		io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
	}))
	defer server.Close()
	svc := fixtureService(server)
	svc.wait = func(context.Context, time.Duration) bool { return true }
	path := filepath.Join(t.TempDir(), "task.json")
	var out bytes.Buffer
	if err := executeRecordedTask("submit", path, taskRequest{Kind: "image", Operation: "generate", Model: "fixture-model", Prompt: "fixture"}, 0, "", &out, svc); err != nil {
		t.Fatal(err)
	}
	for window := 1; window <= 2; window++ {
		out.Reset()
		if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, svc); err != nil {
			t.Fatal(err)
		}
		result := decodeReceipt(t, &out)
		record, err := loadTaskRecord(path)
		if err != nil || record.request().WaitWindowsCompleted != window || result.WaitWindowsCompleted != window {
			t.Fatalf("window budget lost: %+v %v", result, err)
		}
	}
	if posts != 1 || gets != 80 {
		t.Fatalf("unexpected requests: posts=%d gets=%d", posts, gets)
	}
}

func TestInvalidDownloadedMediaGetsSpecificError(t *testing.T) {
	// Valid PNG prefix followed by a truncated file: exercise the complete
	// download branch, not just the error-to-receipt mapper.
	body := fixturePNG(t)
	body = body[:len(body)-12]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(body)
	}))
	defer server.Close()
	request := taskRequest{Kind: "image", TaskID: "paid-task", TaskStatus: "completed", RequestedCount: 1, OutputDir: t.TempDir()}
	data, _ := json.Marshal(request)
	var out bytes.Buffer
	err := executeExistingTask("content", bytes.NewReader(data), &out, fixtureService(server))
	result := decodeReceipt(t, &out)
	if err == nil || result.LocalErrorCode != "invalid_media_content" || result.TaskID != "paid-task" || len(result.DownloadedPaths) != 0 || strings.Contains(result.NextAction, "free disk") {
		t.Fatalf("wrong media failure guidance: %+v %v", result, err)
	}
}

func TestRecordedFlowReattachesWithoutRefetchOrResubmission(t *testing.T) {
	posts, reads, downloads := 0, 0, 0
	png := fixturePNG(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			posts++
			io.WriteString(w, `{"id":"paid-task","status":"pending"}`)
		case strings.HasSuffix(r.URL.Path, "/content"):
			downloads++
			w.Header().Set("Content-Type", "image/png")
			w.Write(png)
		default:
			reads++
			state := "pending"
			if reads > 40 {
				state = "completed"
			}
			io.WriteString(w, `{"id":"paid-task","status":"`+state+`"}`)
		}
	}))
	defer server.Close()
	svc := fixtureService(server)
	svc.wait = func(context.Context, time.Duration) bool { return true }
	dir := t.TempDir()
	recordPath := filepath.Join(dir, "task.json")
	var out bytes.Buffer
	invoke := func(command string, request taskRequest, wantNext string) receipt {
		t.Helper()
		out.Reset()
		if err := executeRecordedTask(command, recordPath, request, 0, dir, &out, svc); err != nil {
			t.Fatal(err)
		}
		result := decodeReceipt(t, &out)
		if result.NextStep != wantNext || result.TaskID != "paid-task" {
			t.Fatalf("%s: %+v", command, result)
		}
		return result
	}
	invoke("submit", taskRequest{Kind: "image", Operation: "generate", Model: "fixture-model", Prompt: "fixture"}, "wait")
	invoke("resume", taskRequest{}, "wait")
	invoke("resume", taskRequest{}, "content")
	downloaded := invoke("content", taskRequest{}, "deliver")
	// Simulate a host handoff failure: no acknowledgement is made. Resume
	// must expose the existing verified file, not read status/content again.
	out.Reset()
	// Unsupported synthetic host would fail credential resolution. A completed
	// recorded handoff must instead validate locally with no credential read.
	if err := run([]string{"resume", "--host", "fixture-unsupported", "--record", recordPath}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	recovered := decodeReceipt(t, &out)
	if len(recovered.DownloadedPaths) != 1 || recovered.DownloadedPaths[0] != downloaded.DownloadedPaths[0] {
		t.Fatal("existing attachment lost")
	}
	// This fixture simulates acknowledgement; it is not real-host acceptance.
	invoke("delivered", taskRequest{}, "done")
	if posts != 1 || reads != 41 || downloads != 1 {
		t.Fatalf("duplicated work: posts=%d status=%d content=%d", posts, reads, downloads)
	}
}

func TestLocalRecoveryRechecksRecordUnderLock(t *testing.T) {
	recordPath := filepath.Join(t.TempDir(), "task.json")
	if err := saveTaskRecord(recordPath, taskRecord{Format: taskRecordFormat, Kind: "image", TaskID: "paid-task", Status: "pending"}, true); err != nil {
		t.Fatal(err)
	}
	calls := 0
	svc := service{localRecoveryOnly: true, client: &http.Client{Transport: pollingFixtureTransport(func(*http.Request) (*http.Response, error) {
		calls++
		t.Fatal("local recovery performed a network request")
		return nil, nil
	})}}
	var out bytes.Buffer
	err := executeRecordedTask("resume", recordPath, taskRequest{}, 0, "", &out, svc)
	if err == nil || calls != 0 || decodeReceipt(t, &out).OK {
		t.Fatal("changed record escaped local recovery")
	}
}
