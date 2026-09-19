package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// This exercises command/HTTP/receipt ordering, not an LLM or real host UI.
// Simulated attachment handoff reads the output bytes before acknowledgement.
func TestOfflineMediaCommandTrajectories(t *testing.T) {
	for _, scenario := range []struct {
		name, kind, operation, model, id string
		count                            int
	}{
		{"two-images", "image", "generate", "seedream-5.0-pro", "pic-seedream-5.0-pro-abcdefghijklmnop", 2},
		{"local-image-edit", "image", "edit", "gpt-image-2", "pic-gpt-image-2-abcdefghijklmnop", 1},
		{"gpt-image-2.5", "image", "generate", "gpt-image-2.5", "pic-gpt-image-2.5-abcdefghijklmnop", 1},
		{"legacy-output-id", "image", "edit", "gpt-image-2", "pic-gpt-image-2-abcdefghijklmnop:image-1", 1},
		{"video", "video", "generate", "grok-imagine-video-1.5-preview", "video-grok-imagine-video-1.5-preview-abcdefghijklmnop", 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			request := taskRequest{Kind: scenario.kind, Operation: scenario.operation, Model: scenario.model, Prompt: "synthetic fixture", RequestedCount: scenario.count}
			if scenario.operation == "edit" {
				request.Attachments = []attachment{{Field: "image", Path: writeFixture(t, "current.png", string(fixturePNG(t)))}}
			}
			body, mediaType := fixturePNG(t), "image/png"
			if scenario.kind == "video" {
				body, mediaType = fixtureWebM(false, false), "video/webm"
			}
			var trace []string
			taskRoute := "/v1/" + scenario.kind + "s/" + scenario.id
			submitRoute := "/v1/videos"
			if scenario.kind == "image" {
				submitRoute = "/v1/images/generations"
				if scenario.operation == "edit" {
					submitRoute = "/v1/images/edits"
				}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, r.Method+" "+r.URL.RequestURI())
				switch {
				case r.Method == http.MethodPost && r.URL.Path == submitRoute:
					if scenario.operation == "edit" {
						if err := r.ParseMultipartForm(1 << 20); err != nil {
							t.Error(err)
						} else {
							defer r.MultipartForm.RemoveAll()
							if len(r.MultipartForm.File["image"]) != 1 {
								t.Error("edit lost its attachment")
							}
						}
					}
					fmt.Fprintf(w, `{"id":%q,"status":"pending"}`, scenario.id)
				case r.Method == http.MethodGet && r.URL.Path == taskRoute:
					fmt.Fprintf(w, `{"id":%q,"status":"completed"}`, scenario.id)
				case r.Method == http.MethodGet && r.URL.Path == taskRoute+"/content":
					w.Header().Set("Content-Type", mediaType)
					w.Write(body)
				default:
					t.Errorf("unexpected preflight, identity or replacement request: %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected", 500)
				}
			}))
			defer server.Close()
			svc := fixtureService(server)
			now := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)
			svc.now = func() time.Time { return now }
			svc.wait = func(_ context.Context, delay time.Duration) bool { now = now.Add(delay); return true }
			record, directory := filepath.Join(t.TempDir(), "task.json"), t.TempDir()
			invoke := func(command string, index int) receipt {
				t.Helper()
				trace = append(trace, "command:"+command)
				var output bytes.Buffer
				outDir := ""
				if command == "content" {
					outDir = directory
				}
				if err := executeRecordedTask(command, record, request, index, outDir, &output, svc); err != nil {
					t.Fatalf("%s failed: %v", command, err)
				}
				result := decodeReceipt(t, &output)
				if result.TaskID != scenario.id || result.OriginalOperation != scenario.operation {
					t.Fatal("lost task identity or original operation")
				}
				trace = append(trace, "receipt:"+result.NextStep)
				return result
			}
			if result := invoke("submit", 0); result.NextStep != "wait" {
				t.Fatal("submission did not return before waiting")
			}
			if result := invoke("resume", 0); result.NextStep != "content" {
				t.Fatal("completion not observed")
			}
			expected := []string{"command:submit", "POST " + submitRoute, "receipt:wait", "command:resume", "GET " + taskRoute, "receipt:content"}
			for index := 0; index < scenario.count; index++ {
				result := invoke("content", index)
				if len(result.DownloadedPaths) != 1 {
					t.Fatal("content must return one file")
				}
				// An interrupted handoff resumes locally, without a new GET.
				recovered := invoke("resume", index)
				if !reflect.DeepEqual(recovered.DownloadedPaths, result.DownloadedPaths) {
					t.Fatal("local handoff recovery changed the file")
				}
				got, err := os.ReadFile(result.DownloadedPaths[0])
				if err != nil || !bytes.Equal(got, body) {
					t.Fatal("simulated handoff did not receive the original bytes")
				}
				trace = append(trace, fmt.Sprintf("handoff:%d", index))
				ack := invoke("delivered", index)
				next := "content"
				if index == scenario.count-1 {
					next = "done"
				}
				if ack.NextStep != next {
					t.Fatalf("incorrect acknowledged progress: %+v", ack)
				}
				expected = append(expected, "command:content", "GET "+contentPath(scenario.kind, scenario.id, index), "receipt:deliver",
					"command:resume", "receipt:deliver", fmt.Sprintf("handoff:%d", index), "command:delivered", "receipt:"+next)
			}
			invoke("resume", 0)
			expected = append(expected, "command:resume", "receipt:done")
			if !reflect.DeepEqual(trace, expected) {
				t.Fatalf("command ordering differs:\n%s\nexpected:\n%s", strings.Join(trace, "\n"), strings.Join(expected, "\n"))
			}
		})
	}
}
