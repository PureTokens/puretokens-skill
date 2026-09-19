package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type taskIDFixtures struct {
	Accepted []string `json:"accepted"`
	Rejected []string `json:"rejected"`
}

func taskIDBoundaryFixtures(t *testing.T) taskIDFixtures {
	t.Helper()
	data, err := os.ReadFile("testdata/task-ids.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures taskIDFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	fixtures.Accepted = append(fixtures.Accepted, strings.Repeat("a", 256))
	fixtures.Rejected = append(fixtures.Rejected, strings.Repeat("a", 257))
	return fixtures
}

func TestPublicTaskIDBoundary(t *testing.T) {
	fixtures := taskIDBoundaryFixtures(t)
	for _, group := range []struct {
		ids      []string
		accepted bool
	}{{fixtures.Accepted, true}, {fixtures.Rejected, false}} {
		for _, id := range group.ids {
			t.Run(fmt.Sprintf("%t/%q", group.accepted, id), func(t *testing.T) {
				if validTaskID(id) != group.accepted {
					t.Fatalf("task ID acceptance must be %t", group.accepted)
				}
				request := taskRequest{Kind: "image", Operation: "continue", TaskID: id}
				if (validateTaskRequest(request) == nil) != group.accepted {
					t.Fatal("continuation disagrees with task ID validation")
				}
				for _, field := range []string{"id", "task_id"} {
					body, _ := json.Marshal(map[string]any{field: id, "status": "pending"})
					got, _, _, ok := taskIdentity(body)
					if ok != group.accepted || (ok && got != id) {
						t.Fatal("response identity was rejected or rewritten")
					}
				}
			})
		}
	}
}

func TestInvalidSubmissionIdentityNeverRetries(t *testing.T) {
	fixtures := taskIDBoundaryFixtures(t)
	for _, id := range fixtures.Rejected {
		t.Run(fmt.Sprintf("%q", id), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "pending"})
			}))
			defer server.Close()
			var output bytes.Buffer
			request := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2.5", Prompt: "fixture"}
			err := executePreparedTask(&output, fixtureService(server), request)
			got := decodeReceipt(t, &output)
			if err == nil || calls != 1 || got.TaskID != "" || got.SubmissionOutcome != "unknown" {
				t.Fatal("unsafe identity must stop after one POST without retaining an ID")
			}
		})
	}
}
