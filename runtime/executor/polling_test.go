package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type pollingFixtureTransport func(*http.Request) (*http.Response, error)

func (transport pollingFixtureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type pollingReply struct {
	status    int
	state     string
	retry     int
	latency   time.Duration
	reconcile bool
}

var pollingFixtureTime = time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)

func runPollingTimeline(t *testing.T, kind string, override *pollRequest, reply func(int, time.Duration) pollingReply) (receipt, []time.Duration, error) {
	t.Helper()
	request := taskRequest{Kind: kind, TaskID: "paid-task", Poll: override}
	if kind == "image" {
		request.RetryNotBefore = pollingFixtureTime.Add(20 * time.Second).Format(time.RFC3339)
	}
	return runPollingRequestTimeline(t, request, reply)
}

// Exercise the actual wait command with a virtual clock and synchronous HTTP
// transport. No wall-clock sleeps, real credentials or external requests.
func runPollingRequestTimeline(t *testing.T, request taskRequest, reply func(int, time.Duration) pollingReply) (receipt, []time.Duration, error) {
	t.Helper()
	start := pollingFixtureTime
	now := start
	budget := 120 * time.Second
	if request.Kind == "video" {
		budget = 300 * time.Second
	}
	override := request.Poll
	if override != nil && override.DeadlineSecs > 0 && time.Duration(override.DeadlineSecs)*time.Second < budget {
		budget = time.Duration(override.DeadlineSecs) * time.Second
	}
	var reads []time.Duration
	statusURLPath := "/v1/" + request.Kind + "s/paid-task"
	transport := pollingFixtureTransport(func(request *http.Request) (*http.Response, error) {
		t.Helper()
		if request.Method != http.MethodGet || request.URL.Path != statusURLPath {
			t.Fatalf("wait made an unexpected request: %s %s", request.Method, request.URL.Path)
		}
		elapsed := now.Sub(start)
		if elapsed >= budget {
			t.Fatal("status read started after its deadline")
		}
		reads = append(reads, elapsed)
		response := reply(len(reads), elapsed)
		now = now.Add(response.latency)
		if response.status == 0 {
			response.status = http.StatusOK
		}
		if response.state == "" {
			response.state = "pending"
		}
		headers := make(http.Header)
		if response.retry > 0 {
			headers.Set("Retry-After", strconv.Itoa(response.retry))
		}
		body := fmt.Sprintf(`{"id":"paid-task","status":%q,"reconciliation_required":%t}`, response.state, response.reconcile)
		return &http.Response{StatusCode: response.status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	svc := service{
		baseURL: "https://fixture.invalid",
		client:  &http.Client{Transport: transport},
		now:     func() time.Time { return now },
		wait: func(ctx context.Context, delay time.Duration) bool {
			if ctx.Err() != nil || now.Add(delay).Sub(start) >= budget {
				return false
			}
			now = now.Add(delay)
			return true
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err = executeExistingTask("wait", bytes.NewReader(data), &out, svc)
	return decodeReceipt(t, &out), reads, err
}

func TestImagePollingFindsShortTasksWithoutLongBackoff(t *testing.T) {
	for _, test := range []struct {
		completedAt, detectedAt time.Duration
		reads                   int
	}{
		{40 * time.Second, 41 * time.Second, 8},
		{46 * time.Second, 47 * time.Second, 10},
		{59 * time.Second, 59 * time.Second, 14},
		{61 * time.Second, 62 * time.Second, 15},
	} {
		t.Run(test.completedAt.String(), func(t *testing.T) {
			result, reads, err := runPollingTimeline(t, "image", nil, func(_ int, elapsed time.Duration) pollingReply {
				if elapsed >= test.completedAt {
					return pollingReply{state: "completed"}
				}
				return pollingReply{}
			})
			if err != nil || result.Status != "completed" || result.TaskID != "paid-task" || len(reads) != test.reads || reads[len(reads)-1] != test.detectedAt {
				t.Fatalf("unexpected completion: %+v reads=%v err=%v", result, reads, err)
			}
		})
	}
}

func TestImagePollingDoesNotRestartInitialWaitAfterServerOrNetworkDelay(t *testing.T) {
	for _, test := range []struct {
		name   string
		reply  func(int, time.Duration) pollingReply
		expect []time.Duration
	}{
		{
			name: "rate limit crosses first minute",
			reply: func(read int, _ time.Duration) pollingReply {
				switch read {
				case 1:
					return pollingReply{status: 429, retry: 70}
				case 2:
					return pollingReply{}
				default:
					return pollingReply{state: "completed"}
				}
			},
			expect: []time.Duration{20 * time.Second, 90 * time.Second, 93 * time.Second},
		},
		{
			name: "slow response crosses first minute",
			reply: func(read int, _ time.Duration) pollingReply {
				if read == 1 {
					return pollingReply{latency: 58 * time.Second}
				}
				return pollingReply{state: "completed"}
			},
			expect: []time.Duration{20 * time.Second, 81 * time.Second},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, reads, err := runPollingTimeline(t, "image", nil, test.reply)
			if err != nil || !result.OK || result.RetryNotBefore != "" || !reflect.DeepEqual(reads, test.expect) {
				t.Fatalf("wait restarted or ignored: %+v reads=%v err=%v", result, reads, err)
			}
		})
	}
}

func TestImagePollingKeepsDeadlineAndReadBounds(t *testing.T) {
	for _, test := range []struct {
		name     string
		override *pollRequest
		retry    int
		count    int
		lastRead time.Duration
	}{
		{"default deadline", nil, 0, 34, 119 * time.Second},
		{"shorter window", &pollRequest{DeadlineSecs: 8}, 0, 0, 0},
		{"lower read limit", &pollRequest{MaxStatusReads: 3}, 0, 3, 26 * time.Second},
		{"server delays hit read limit first", nil, 1, 40, 59 * time.Second},
		{"oversized overrides cannot extend window", &pollRequest{MaxStatusReads: 100, DeadlineSecs: 300}, 0, 34, 119 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, reads, err := runPollingTimeline(t, "image", test.override, func(int, time.Duration) pollingReply {
				return pollingReply{retry: test.retry}
			})
			if err == nil || result.OK || result.TaskID != "paid-task" || result.Status != "pending" || len(reads) != test.count || (len(reads) > 0 && reads[len(reads)-1] != test.lastRead) {
				t.Fatalf("unbounded wait: %+v reads=%v err=%v", result, reads, err)
			}
		})
	}
}

func TestImagePollingStopsOnUnsafeContinuation(t *testing.T) {
	for _, test := range []struct {
		name  string
		reply pollingReply
	}{
		{"retry exceeds remaining budget", pollingReply{status: 429, retry: 118}},
		{"rate limit without retry", pollingReply{status: 429}},
		{"server failure", pollingReply{status: 503}},
		{"unknown status", pollingReply{state: "unrecognized"}},
		{"reconciliation", pollingReply{reconcile: true}},
		{"failed task", pollingReply{state: "failed"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, reads, err := runPollingTimeline(t, "image", nil, func(int, time.Duration) pollingReply { return test.reply })
			if err == nil || result.OK || result.TaskID != "paid-task" || len(reads) != 1 {
				t.Fatalf("unsafe continuation: %+v reads=%v err=%v", result, reads, err)
			}
			if test.reply.retry > 0 && result.RetryNotBefore == "" {
				t.Fatal("server retry time was lost")
			}
		})
	}
}

func TestVideoPollingRetainsExistingSchedule(t *testing.T) {
	result, reads, err := runPollingTimeline(t, "video", nil, func(int, time.Duration) pollingReply { return pollingReply{} })
	want := []time.Duration{5 * time.Second, 15 * time.Second, 35 * time.Second, 75 * time.Second, 135 * time.Second, 195 * time.Second, 255 * time.Second}
	if err == nil || result.OK || !reflect.DeepEqual(reads, want) {
		t.Fatalf("video schedule changed: %+v reads=%v err=%v", result, reads, err)
	}
}

func TestImageSubmissionSchedulesInitialWaitFromAcceptedResponse(t *testing.T) {
	for _, test := range []struct {
		name, kind, state, header string
		httpStatus, delay, retry  int
		wantError                 bool
	}{
		{name: "pending", kind: "image", state: "pending", delay: 20},
		{name: "queued", kind: "image", state: "queued", delay: 20},
		{name: "processing", kind: "image", state: "processing", delay: 20},
		{name: "invalid header", kind: "image", state: "pending", header: "invalid", delay: 20},
		{name: "server requests shorter wait", kind: "image", state: "pending", header: "5", delay: 5, retry: 5},
		{name: "server requests longer wait", kind: "image", state: "pending", header: "45", delay: 45, retry: 45},
		{name: "already complete", kind: "image", state: "completed"},
		{name: "failed task", kind: "image", state: "failed", wantError: true},
		{name: "unknown status", kind: "image", state: "unexpected", wantError: true},
		{name: "unknown submission", kind: "image", httpStatus: 503, wantError: true},
		{name: "video", kind: "video", state: "pending"},
	} {
		t.Run(test.name, func(t *testing.T) {
			now := pollingFixtureTime
			calls := 0
			svc := service{
				baseURL: "https://fixture.invalid",
				now:     func() time.Time { return now },
				client: &http.Client{Transport: pollingFixtureTransport(func(request *http.Request) (*http.Response, error) {
					calls++
					if request.Method != http.MethodPost {
						t.Fatal("submit performed a status or catalog read")
					}
					// Slow POST acceptance must not consume the post-acceptance
					// processing allowance, or depend on upstream timestamps.
					now = now.Add(11 * time.Second)
					status := test.httpStatus
					if status == 0 {
						status = http.StatusOK
					}
					headers := make(http.Header)
					headers.Set("Retry-After", test.header)
					body := fmt.Sprintf(`{"id":"paid-task","status":%q}`, test.state)
					return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
				})},
			}
			data := fmt.Sprintf(`{"kind":%q,"operation":"generate","model":"fixture-model","prompt":"fixture"}`, test.kind)
			var out bytes.Buffer
			err := executeTask(strings.NewReader(data), &out, svc)
			result := decodeReceipt(t, &out)
			wantTime := ""
			if test.delay > 0 {
				wantTime = now.Add(time.Duration(test.delay) * time.Second).Format(time.RFC3339)
			}
			if (err != nil) != test.wantError || calls != 1 || result.RetryNotBefore != wantTime || result.RetryAfterSecs != test.retry {
				t.Fatalf("initial wait incorrectly anchored or reported: %+v err=%v calls=%d", result, err, calls)
			}
		})
	}
}

func TestImageInitialWaitSurvivesHandoffAndTaskRecords(t *testing.T) {
	for _, recorded := range []bool{false, true} {
		for _, handoff := range []time.Duration{0, 15 * time.Second, 25 * time.Second} {
			t.Run(fmt.Sprintf("record=%t/handoff=%s", recorded, handoff), func(t *testing.T) {
				now := pollingFixtureTime
				posts, gets := 0, 0
				var waits []time.Duration
				svc := service{
					baseURL: "https://fixture.invalid",
					now:     func() time.Time { return now },
					wait: func(_ context.Context, delay time.Duration) bool {
						waits = append(waits, delay)
						now = now.Add(delay)
						return true
					},
					client: &http.Client{Transport: pollingFixtureTransport(func(request *http.Request) (*http.Response, error) {
						state := "pending"
						if request.Method == http.MethodPost {
							posts++
						} else if request.Method == http.MethodGet && request.URL.Path == "/v1/images/paid-task" {
							gets++
							state = "completed"
							if now.Before(pollingFixtureTime.Add(20 * time.Second)) {
								t.Fatal("queried before the recorded initial wait")
							}
						} else {
							t.Fatal("unexpected request")
						}
						return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"id":"paid-task","status":%q}`, state)))}, nil
					})},
				}
				request := taskRequest{Kind: "image", Operation: "generate", Model: "fixture-model", Prompt: "fixture"}
				recordPath := filepath.Join(t.TempDir(), "task.json")
				var out bytes.Buffer
				var err error
				if recorded {
					err = executeRecordedTask("submit", recordPath, request, 0, "", &out, svc)
				} else {
					data, _ := json.Marshal(request)
					err = executeTask(bytes.NewReader(data), &out, svc)
				}
				if err != nil {
					t.Fatal(err)
				}
				accepted := decodeReceipt(t, &out)
				if accepted.RetryNotBefore != "2026-09-06T00:00:20Z" || accepted.RetryAfterSecs != 0 {
					t.Fatalf("missing or misreported initial wait: %+v", accepted)
				}
				now = now.Add(handoff)
				out.Reset()
				if recorded {
					record, loadErr := loadTaskRecord(recordPath)
					if loadErr != nil || record.RetryNotBefore != accepted.RetryNotBefore || record.Format != "puretokens-task-v1" {
						t.Fatalf("record lost its scheduling boundary: %+v %v", record, loadErr)
					}
					err = executeRecordedTask("resume", recordPath, taskRequest{}, 0, "", &out, svc)
				} else {
					continuation := taskRequest{Kind: accepted.Kind, TaskID: accepted.TaskID, RetryNotBefore: accepted.RetryNotBefore}
					data, _ := json.Marshal(continuation)
					err = executeExistingTask("wait", bytes.NewReader(data), &out, svc)
				}
				result := decodeReceipt(t, &out)
				wantWait := max(time.Duration(0), 20*time.Second-handoff)
				if err != nil || !result.OK || result.TaskID != accepted.TaskID || result.Status != "completed" ||
					posts != 1 || gets != 1 || !reflect.DeepEqual(waits, []time.Duration{wantWait}) || result.RetryNotBefore != "" {
					t.Fatalf("initial wait restarted: %+v waits=%v posts=%d gets=%d err=%v", result, waits, posts, gets, err)
				}
			})
		}
	}
}

func TestImageOlderRecordsContinueWithoutInventingInitialWait(t *testing.T) {
	for _, timestamp := range []string{"", "2026-09-05T00:00:20Z"} {
		t.Run(timestamp, func(t *testing.T) {
			recordPath := filepath.Join(t.TempDir(), "old-task.json")
			// Both shapes predate this change; no new record version/field is
			// needed. A previous server delay must also be accepted unchanged.
			recordJSON := `{"format":"puretokens-task-v1","kind":"image","task_id":"paid-task","status":"pending"}`
			if timestamp != "" {
				recordJSON = strings.TrimSuffix(recordJSON, "}") + `,"retry_not_before":"` + timestamp + `"}`
			}
			if err := os.WriteFile(recordPath, []byte(recordJSON), 0600); err != nil {
				t.Fatal(err)
			}
			record, err := loadTaskRecord(recordPath)
			if err != nil {
				t.Fatal(err)
			}
			result, reads, err := runPollingRequestTimeline(t, record.request(), func(int, time.Duration) pollingReply { return pollingReply{state: "completed"} })
			if err != nil || !result.OK || !reflect.DeepEqual(reads, []time.Duration{0}) {
				t.Fatalf("old task delayed: %+v reads=%v err=%v", result, reads, err)
			}
		})
	}
	request := taskRequest{Kind: "image", TaskID: "paid-task"}
	result, reads, err := runPollingRequestTimeline(t, request, func(int, time.Duration) pollingReply { return pollingReply{} })
	if err == nil || result.OK || len(reads) != 40 || reads[0] != 0 || reads[len(reads)-1] != 117*time.Second {
		t.Fatalf("continuation window is unbounded: %+v reads=%v err=%v", result, reads, err)
	}
}
