package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitResumeRefreshesReconciliationOnce(t *testing.T) {
	for _, state := range []string{"completed", "pending", "reconciling", "unavailable"} {
		t.Run(state, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "GET" || r.URL.Path != "/v1/images/existing-task" {
					t.Error("resume used a different task or submitted media")
				}
				if state == "unavailable" {
					w.WriteHeader(503)
					return
				}
				status := state
				if state == "reconciling" {
					status = "completed"
				}
				json.NewEncoder(w).Encode(map[string]any{"id": "existing-task", "status": status, "reconciliation_required": state == "reconciling"})
			}))
			defer server.Close()
			path := filepath.Join(t.TempDir(), "task.json")
			record := taskRecord{Format: taskRecordFormat, Kind: "image", TaskID: "existing-task", Model: "gpt-image-2", RequestedCount: 1, OriginalOperation: "generate", Status: "completed", ReconciliationRequired: true}
			if err := saveTaskRecord(path, record, true); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, fixtureService(server))
			result := decodeReceipt(t, &out)
			updated, loadErr := loadTaskRecord(path)
			if calls != 1 || loadErr != nil || result.TaskID != record.TaskID {
				t.Fatalf("wrong resume: calls=%d load=%v receipt=%+v", calls, loadErr, result)
			}
			wantReconcile := state == "reconciling" || state == "unavailable"
			if result.ReconciliationRequired != wantReconcile || updated.ReconciliationRequired != wantReconcile || (err != nil) != wantReconcile {
				t.Fatalf("incorrect reconciliation: %+v, %v", result, err)
			}
		})
	}
}

func fixtureGIF(t *testing.T) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	var out bytes.Buffer
	if err := gif.EncodeAll(&out, &gif.GIF{Image: []*image.Paletted{img, img}, Delay: []int{1, 1}}); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func ebmlFixture(id []byte, payload ...[]byte) []byte {
	body := bytes.Join(payload, nil)
	if len(body) >= 127 {
		panic("fixture requires a small EBML element")
	}
	return append(append(append([]byte{}, id...), byte(len(body))|0x80), body...)
}

func fixtureWebM(unknownSegment, unknownCluster bool) []byte {
	header := ebmlFixture([]byte{0x1a, 0x45, 0xdf, 0xa3}, ebmlFixture([]byte{0x42, 0x82}, []byte("webm")))
	tracks := ebmlFixture([]byte{0x16, 0x54, 0xae, 0x6b},
		ebmlFixture([]byte{0xae}, ebmlFixture([]byte{0xd7}, []byte{1}), ebmlFixture([]byte{0x83}, []byte{1}), ebmlFixture([]byte{0x86}, []byte("V_VP9"))))
	cluster := ebmlFixture([]byte{0x1f, 0x43, 0xb6, 0x75},
		ebmlFixture([]byte{0xe7}, []byte{0}), ebmlFixture([]byte{0xa3}, []byte{0x81, 0, 0, 0x80, 0x01, 0x02}))
	if unknownCluster {
		cluster[4] = 0xff
	}
	segment := ebmlFixture([]byte{0x18, 0x53, 0x80, 0x67}, tracks, cluster)
	if unknownSegment {
		segment[4] = 0xff
	}
	return append(header, segment...)
}

func TestMediaContainersRejectTruncationAndReuseOnlyCompleteFiles(t *testing.T) {
	for _, tc := range []struct {
		name, kind, mediaType string
		body                  []byte
	}{
		{"gif", "image", "image/gif", fixtureGIF(t)},
		{"webm", "video", "video/webm", fixtureWebM(false, false)},
		{"streamed-webm", "video", "video/webm", fixtureWebM(true, true)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Every prefix truncation of these fixtures must fail, including
			// the formerly accepted 13-byte GIF header.
			for cut := 0; cut < len(tc.body); cut++ {
				valid := false
				if tc.kind == "image" {
					valid = validGIF(bytes.NewReader(tc.body[:cut]))
				} else {
					valid = validWebM(bytes.NewReader(tc.body[:cut]), int64(cut))
				}
				if valid {
					t.Fatalf("accepted truncation at %d/%d", cut, len(tc.body))
				}
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", tc.mediaType)
				w.Write(tc.body)
			}))
			defer server.Close()
			dir := t.TempDir()
			svc := fixtureService(server)
			path := contentPath(tc.kind, "existing-task", 0)
			file, _, _, _, _, err := svc.download(context.Background(), path, tc.kind, dir)
			if err != nil {
				t.Fatal("complete container rejected", err)
			}
			if _, _, _, _, _, err = svc.download(context.Background(), path, tc.kind, dir); err != nil || calls != 1 {
				t.Fatal("validated result was not reused")
			}
			os.WriteFile(file, tc.body[:len(tc.body)-1], 0600)
			if _, _, _, _, _, err = svc.download(context.Background(), path, tc.kind, dir); err == nil || calls != 1 {
				t.Fatal("modified/truncated output reused or overwritten")
			}
		})
	}
}

func TestHeaderOnlyMediaIsNeverSaved(t *testing.T) {
	for _, tc := range []struct {
		kind, mediaType string
		body            []byte
	}{
		{"image", "image/gif", []byte{'G', 'I', 'F', '8', '9', 'a', 1, 0, 1, 0, 0, 0, 0}},
		{"video", "video/webm", append([]byte{0x1a, 0x45, 0xdf, 0xa3}, []byte("webm-no-body!!")...)},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", tc.mediaType)
			w.Write(tc.body)
		}))
		dir := t.TempDir()
		_, _, _, _, _, err := fixtureService(server).download(context.Background(), contentPath(tc.kind, "existing-task", 0), tc.kind, dir)
		server.Close()
		files, _ := os.ReadDir(dir)
		if err == nil || len(files) != 0 {
			t.Fatal("header-only media accepted or left an output")
		}
	}
}

func TestWebMEncoderFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/tiny.webm")
	if err != nil {
		t.Fatal(err)
	}
	if !validWebM(bytes.NewReader(data), int64(len(data))) {
		t.Fatal("actual encoder WebM output rejected")
	}
	for cut := 0; cut < len(data); cut++ {
		if validWebM(bytes.NewReader(data[:cut]), int64(cut)) {
			t.Fatalf("truncated sized encoder fixture accepted at %d", cut)
		}
	}
}

func TestPublicErrorsCannotEchoRequestOrUpstream(t *testing.T) {
	private := "synthetic unpublished campaign; /srv/private/jobs/42; worker.example.dev"
	for _, code := range []string{"invalid_request", "unrecognized_" + private, ""} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": code, "message": private}})
		}))
		var out bytes.Buffer
		executeReadOnly(&out, fixtureService(server), "/v1/media/models", "models")
		server.Close()
		if strings.Contains(out.String(), private) || strings.Contains(out.String(), "worker.example.dev") {
			t.Fatal("private server text entered receipt")
		}
		got := decodeReceipt(t, &out)
		if got.HTTPStatus != 400 || got.NextAction == "" || got.ErrorMessage == "" {
			t.Fatal("safe diagnostic context lost")
		}
	}
}

func FuzzPublicErrorCategories(f *testing.F) {
	f.Add("invalid_request", "private /srv/job https://example.test")
	f.Add("unknown_provider", "synthetic prompt")
	f.Fuzz(func(t *testing.T, code, message string) {
		gotCode, gotText := publicErrorCategory(code, message)
		if gotCode != "" && (gotCode != code || gotText != publicErrorMessages[gotCode]) {
			t.Fatal("invented code or uncontrolled error text")
		}
		if gotCode == "" && gotText != "" && gotText != "The API reported unsafe generated content. Revise the request." {
			t.Fatal("unknown server text escaped")
		}
	})
}

func FuzzMediaContainers(f *testing.F) {
	f.Add([]byte("GIF89a"))
	f.Add(fixtureWebM(false, false))
	f.Add(fixtureWebM(true, true))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		validGIF(bytes.NewReader(data))
		validWebM(bytes.NewReader(data), int64(len(data)))
	})
}

func FuzzTaskRequestParsing(f *testing.F) {
	f.Add(`{"kind":"image","operation":"generate","model":"gpt-image-2","prompt":"fixture"}`)
	f.Add(`{"kind":"video","operation":"continue","task_id":"known-task","retry_not_before":"invalid"}`)
	f.Fuzz(func(t *testing.T, data string) {
		if len(data) > maxResponseBytes {
			t.Skip()
		}
		request, err := decodeTaskRequest(strings.NewReader(data))
		if err != nil {
			return
		}
		// Validate only pure task identity/projection; fuzzing must not read
		// arbitrary attachment paths or configuration.
		if request.Operation == "continue" {
			validateTaskRequest(request)
		}
		result := taskReceipt(request, "", request.TaskStatus)
		if _, err := json.Marshal(result); err != nil && err != io.EOF {
			t.Fatal("unserializable receipt", err)
		}
	})
}
