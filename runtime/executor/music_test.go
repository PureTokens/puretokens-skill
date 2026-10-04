package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func musicFixture() taskRequest {
	return taskRequest{Kind: "music", Operation: "generate", Model: "stepaudio-3-music-preview", Prompt: "Warm piano, no vocals", Parameters: map[string]any{"instrumental": true, "response_format": "wav"}}
}

func TestMusicTaskRecordSubmissionContinuationAndHandoff(t *testing.T) {
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/v1/audio/music/submit":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["caption"] != "Warm piano, no vocals" || body["instrumental"] != true || body["format"] != "wav" || len(body) != 4 {
				t.Errorf("wrong payload: %+v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"music_fixture","status":"pending"}`)
		case "/v1/audio/music/tasks/music_fixture":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"music_fixture","status":"succeeded"}`)
		case "/v1/audio/music/tasks/music_fixture/content":
			w.Header().Set("Content-Type", "audio/wav")
			w.Write(audioWAVFixture())
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	root := t.TempDir()
	recordPath := filepath.Join(root, "music.json")
	req := musicFixture()
	req.OutputDir = root
	svc := fixtureService(server)
	var out bytes.Buffer
	if err := executeRecordedTask("submit", recordPath, req, 0, "", &out, svc); err != nil {
		t.Fatalf("submit: %v %s", err, out.String())
	}
	if len(calls) != 1 {
		t.Fatal("submit polled or downloaded")
	}
	saved, err := loadTaskRecord(recordPath)
	if err != nil || saved.TaskID != "music_fixture" || saved.RequestedCount != 1 {
		t.Fatalf("bad record: %+v %v", saved, err)
	}
	data, _ := os.ReadFile(recordPath)
	if bytes.Contains(data, []byte("Warm piano")) || bytes.Contains(data, []byte("lyrics")) {
		t.Fatal("private text persisted")
	}
	out.Reset()
	if err := executeRecordedTask("status", recordPath, taskRequest{}, 0, "", &out, svc); err != nil {
		t.Fatal(err, out.String())
	}
	out.Reset()
	if err := executeRecordedTask("content", recordPath, taskRequest{}, 0, root, &out, svc); err != nil {
		t.Fatal(err, out.String())
	}
	saved, err = loadTaskRecord(recordPath)
	if err != nil || len(saved.Downloaded) != 1 || len(saved.Delivered) != 0 {
		t.Fatal("download mistaken for handoff")
	}
	out.Reset()
	if err := executeRecordedTask("resume", recordPath, taskRequest{}, 0, "", &out, svc); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || !strings.Contains(out.String(), `"next_step":"deliver"`) {
		t.Fatal("resume did not reuse original bytes")
	}
	out.Reset()
	if err := run([]string{"resume", "--host", "fixture-unsupported", "--record", recordPath}, strings.NewReader(""), &out); err != nil {
		t.Fatal("local music recovery tried credential resolution", err)
	}
	if len(calls) != 3 || !strings.Contains(out.String(), `"next_step":"deliver"`) {
		t.Fatal("local music recovery lost the downloaded file")
	}
	out.Reset()
	if err := executeRecordedTask("delivered", recordPath, taskRequest{}, 0, "", &out, svc); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"next_step":"done"`) || len(calls) != 3 {
		t.Fatal("handoff incomplete")
	}
	out.Reset()
	if err := executeRecordedTask("submit", recordPath, req, 0, "", &out, svc); err == nil || len(calls) != 3 {
		t.Fatal("existing record resubmitted")
	}
}

func TestMusicValidationPreventsPaidRewritesAndSynchronousFallback(t *testing.T) {
	for name, change := range map[string]func(*taskRequest){
		"model":         func(r *taskRequest) { r.Model = "other" },
		"edit":          func(r *taskRequest) { r.Operation = "edit" },
		"no-intent":     func(r *taskRequest) { delete(r.Parameters, "instrumental") },
		"contradiction": func(r *taskRequest) { r.Parameters["lyrics"] = "private lyrics" },
		"long-caption":  func(r *taskRequest) { r.Prompt = strings.Repeat("字", 1001) },
		"long-lyrics": func(r *taskRequest) {
			r.Parameters["instrumental"] = false
			r.Parameters["lyrics"] = strings.Repeat("字", 4001)
		},
		"group":    func(r *taskRequest) { r.Parameters["group"] = "unrequested" },
		"duration": func(r *taskRequest) { r.Parameters["duration"] = 60 },
		"batch":    func(r *taskRequest) { r.RequestedCount = 2 },
	} {
		t.Run(name, func(t *testing.T) {
			r := musicFixture()
			change(&r)
			if validateTaskRequest(r) == nil {
				t.Fatal("unsafe request accepted")
			}
		})
	}
	r := musicFixture()
	r.Parameters["instrumental"] = false
	r.Parameters["lyrics"] = strings.Repeat("字", 4000)
	if validateTaskRequest(r) != nil {
		t.Fatal("valid bounded lyrics rejected")
	}
	_, _, body, err := taskRequestBody(r)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	json.NewDecoder(body).Decode(&payload)
	if payload["lyrics"] != r.Parameters["lyrics"] || payload["instrumental"] != false {
		t.Fatal("lyrics/intention changed")
	}
	for _, input := range []string{
		`{"kind":"music","kind":"music"}`, `{"kind":"music","Kind":"music"}`, `{"kind":"music","model":null}`, `{"kind":"music","parameters":{"instrumental":true,"instrumental":false}}`,
	} {
		if _, err := decodeTaskRequest(strings.NewReader(input)); err == nil {
			t.Fatal("ambiguous music JSON accepted")
		}
	}
	if _, err := decodeAudioRequest([]byte(`{"operation":"music","model":"stepaudio-3-music-preview"}`)); err == nil {
		t.Fatal("music accepted in sync command")
	}
}

func TestMusicUnknownAndReconciliationKeepOriginalTask(t *testing.T) {
	for _, scenario := range []string{"missing-id", "server", "reconciliation", "expired", "bad-status"} {
		t.Run(scenario, func(t *testing.T) {
			posts, reads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "POST" {
					posts++
					if scenario == "server" {
						w.WriteHeader(503)
						io.WriteString(w, `{"error":"private"}`)
					} else if scenario == "missing-id" {
						io.WriteString(w, `{"status":"succeeded"}`)
					} else {
						io.WriteString(w, `{"id":"music_existing","status":"pending"}`)
					}
					return
				}
				reads++
				if strings.HasSuffix(r.URL.Path, "/content") {
					w.WriteHeader(410)
					io.WriteString(w, `{"error":{"code":"task_expired"}}`)
					return
				}
				if scenario == "reconciliation" {
					io.WriteString(w, `{"id":"music_existing","status":"unknown","reconciliation_required":true}`)
				} else if scenario == "bad-status" {
					io.WriteString(w, `{"id":"wrong_task","status":"succeeded"}`)
				} else {
					io.WriteString(w, `{"id":"music_existing","status":"succeeded"}`)
				}
			}))
			defer server.Close()
			root := t.TempDir()
			record := filepath.Join(root, "task.json")
			r := musicFixture()
			r.OutputDir = root
			svc := fixtureService(server)
			var out bytes.Buffer
			err := executeRecordedTask("submit", record, r, 0, "", &out, svc)
			if scenario == "server" || scenario == "missing-id" {
				if err == nil || posts != 1 || reads != 0 {
					t.Fatal("unknown submit did not stop")
				}
				saved, e := loadTaskRecord(record)
				if e != nil || saved.SubmissionOutcome != "unknown" || saved.TaskID != "" {
					t.Fatal("unknown not retained")
				}
				out.Reset()
				if executeRecordedTask("resume", record, taskRequest{}, 0, "", &out, svc) == nil || posts != 1 || reads != 0 {
					t.Fatal("unknown resumed or repeated")
				}
				out.Reset()
				if err := run([]string{"resume", "--host", "fixture-unsupported", "--record", record}, strings.NewReader(""), &out); err == nil {
					t.Fatal("unknown submission resumed through CLI")
				}
				if !strings.Contains(out.String(), `"submission_outcome":"unknown"`) || strings.Contains(out.String(), "credential") {
					t.Fatal("CLI lost recorded uncertainty or attempted credential discovery", out.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			out.Reset()
			err = executeRecordedTask("status", record, taskRequest{}, 0, "", &out, svc)
			if scenario == "reconciliation" {
				if err == nil || !strings.Contains(out.String(), `"reconciliation_required":true`) || !strings.Contains(out.String(), `"next_step":"await_user"`) {
					t.Fatal("reconciliation not stopped")
				}
				out.Reset()
				executeRecordedTask("resume", record, taskRequest{}, 0, "", &out, svc)
				if reads != 2 || posts != 1 {
					t.Fatal("reconciliation resume must read once")
				}
			} else if scenario == "bad-status" {
				if err == nil || !strings.Contains(out.String(), `"task_id":"music_existing"`) {
					t.Fatal("identity lost")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				out.Reset()
				err = executeRecordedTask("content", record, taskRequest{}, 0, root, &out, svc)
				if err == nil || !strings.Contains(out.String(), `"http_status":410`) || posts != 1 {
					t.Fatal("expired replaced task")
				}
			}
		})
	}
}

func TestMusicWaitBudgetDoesNotResetOnResume(t *testing.T) {
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"music_pending","status":"running"}`)
	}))
	defer server.Close()
	root := t.TempDir()
	record := filepath.Join(root, "task.json")
	r := musicFixture()
	r.Operation = "continue"
	r.Prompt = ""
	r.TaskID = "music_pending"
	r.TaskStatus = "running"
	r.RequestedCount = 1
	r.OriginalOperation = "generate"
	if err := saveTaskRecord(record, recordFromRequest(r), true); err != nil {
		t.Fatal(err)
	}
	svc := fixtureService(server)
	svc.wait = func(context.Context, time.Duration) bool { return true }
	for i := 1; i <= 2; i++ {
		var out bytes.Buffer
		if err := executeRecordedTask("wait", record, taskRequest{}, 0, "", &out, svc); err != nil {
			t.Fatal(err)
		}
		saved, _ := loadTaskRecord(record)
		if saved.WaitWindowsCompleted != i {
			t.Fatal("lost window count")
		}
	}
	if reads != 14 {
		t.Fatal("unbounded polling", reads)
	}
}

func TestMusicDownloadSizeAndStructuralValidation(t *testing.T) {
	for _, size := range []int{int(maxMusicBytes), int(maxMusicBytes) + 2, 364} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := make([]byte, size)
			copy(data, audioWAVFixture())
			binary.LittleEndian.PutUint32(data[4:], uint32(size-8))
			binary.LittleEndian.PutUint32(data[40:], uint32(size-44))
			if size == 364 {
				// A recognizable WAV prefix alone must not count as audio.
				data = data[:300]
			}
			svc := service{baseURL: apiOrigin, client: &http.Client{Transport: evaluationTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"audio/wav"}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
			})}}
			file, _, _, _, _, err := svc.download(context.Background(), "/v1/audio/music/tasks/music_fixture/content", "music", t.TempDir())
			if size == int(maxMusicBytes) {
				if err != nil {
					t.Fatal("32 MiB music rejected", err)
				}
				proof, err := fingerprintDownload(file, "audio/wav")
				if err != nil || !validDownloadProof(proof, "music") || proof.Bytes != maxMusicBytes {
					t.Fatal("valid music proof lost", err)
				}
			} else if err == nil || file != "" {
				t.Fatal("oversized or incomplete audio accepted")
			}
		})
	}
}

func TestMusicDownloadRejectsChangedFormatBeforeFileWrite(t *testing.T) {
	request := musicFixture() // requested WAV, returned MP3
	request.Operation, request.Prompt = "continue", ""
	request.TaskID, request.TaskStatus = "music_format", "succeeded"
	request.OutputDir = t.TempDir()
	svc := service{baseURL: apiOrigin, client: &http.Client{Transport: evaluationTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"audio/mpeg"}}, Body: io.NopCloser(bytes.NewReader(audioMP3Fixture()))}, nil
	})}}
	var out bytes.Buffer
	data, _ := json.Marshal(request)
	if err := executeExistingTask("content", bytes.NewReader(data), &out, svc); err == nil {
		t.Fatal("wrong requested audio format delivered")
	}
	files, err := os.ReadDir(request.OutputDir)
	if err != nil || len(files) != 0 {
		t.Fatal("wrong-format output saved")
	}
}
