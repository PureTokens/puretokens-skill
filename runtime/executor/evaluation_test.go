package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const evaluationInputFixture = `{"model":"jev-latest","state":"private-state-marker","questions":{"department":{"type":"choice","instructions":"Choose a team","criteria":{"billing":null,"technical":"Technical problems"}},"urgency":{"type":"noul","instructions":"Is it urgent?"},"severity":{"type":"score","instructions":"Rate severity","criteria":["Low","Medium","High"]}}}`
const evaluationOutputFixture = `{"model":"jev-1.13.0","answers":{"department":{"type":"choice","choice":"billing","probabilities":{"billing":0.8,"technical":0.2},"confidence":0.5},"urgency":{"type":"noul","noul":0.9},"severity":{"type":"score","score":1.1,"legend":{"0":"Low","1":"Medium","2":"High"},"probabilities":{"0":0.1,"1":0.7,"2":0.2},"confidence":0.4}},"usage":{"input_tokens":200,"output_tokens":80},"debug":"private-response-marker"}`

type evaluationTransport func(*http.Request) (*http.Response, error)

func (f evaluationTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func evaluationCLI(t *testing.T, input string, transport evaluationTransport, options ...string) (map[string]any, error, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("CODEX_HOME", home)
	err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("model_provider='fixture'\n[model_providers.fixture]\nbase_url='https://api.puretokensx.com/v1'\nexperimental_bearer_token='synthetic-evaluation-token'\n"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(home, "request.json")
	if err = os.WriteFile(file, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
	var output bytes.Buffer
	err = run(append([]string{"evaluate", "--host", "codex", "--request", file}, options...), strings.NewReader(""), &output)
	var document map[string]any
	if json.Unmarshal(output.Bytes(), &document) != nil {
		t.Fatal("expected one JSON receipt")
	}
	return document, err, output.String()
}

func TestEvaluationPreservesStructuredNumbers(t *testing.T) {
	input := strings.Replace(evaluationInputFixture, `"private-state-marker"`, `{"ticket":9007199254740993,"amount":1.234567890123456789,"nested":[true,null,"你好"]}`, 1)
	calls := 0
	_, err, _ := evaluationCLI(t, input, func(r *http.Request) (*http.Response, error) {
		calls++
		body, _ := io.ReadAll(r.Body)
		for _, number := range []string{"9007199254740993", "1.234567890123456789"} {
			if !bytes.Contains(body, []byte(number)) {
				t.Errorf("structured input number changed: %s", number)
			}
		}
		return evaluationHTTP(200, evaluationOutputFixture), nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("structured evaluation failed: %v", err)
	}
}

func TestEvaluationAcceptsPowerShellUTF8BOM(t *testing.T) {
	input := append([]byte{0xef, 0xbb, 0xbf}, []byte(evaluationInputFixture)...)
	if _, err := decodeEvaluationRequest(input); err != nil {
		t.Fatal("Windows PowerShell UTF-8 request rejected")
	}
}

func TestEvaluationRejectsInvalidCLIWithoutSubmitting(t *testing.T) {
	for _, option := range []string{"--index=0", "--record=", "--output-dir=", "--host=", "--stream", "unexpected-positional"} {
		t.Run(option, func(t *testing.T) {
			calls := 0
			result, err, _ := evaluationCLI(t, evaluationInputFixture, func(*http.Request) (*http.Response, error) {
				calls++
				return evaluationHTTP(200, evaluationOutputFixture), nil
			}, option)
			if err == nil || calls != 0 || result["submission_outcome"] != "not_submitted" ||
				result["command"] != "evaluate" || result["local_error_code"] != "evaluation_request_invalid" {
				t.Fatalf("invalid CLI submitted evaluation or lost its receipt contract: %v calls=%d", result, calls)
			}
		})
	}
}

func TestEvaluationQuestionAndRequestLimits(t *testing.T) {
	for _, count := range []int{64, 65} {
		questions := map[string]any{}
		for i := 0; i < count; i++ {
			questions["q"+strconv.Itoa(i)] = map[string]any{"type": "noul", "instructions": "Is it urgent?"}
		}
		body, _ := json.Marshal(map[string]any{"model": "jev-preview", "state": []any{"你好"}, "questions": questions})
		_, err := decodeEvaluationRequest(body)
		if (err == nil) != (count == 64) {
			t.Fatalf("question bound %d: %v", count, err)
		}
	}
	for _, count := range []int{255, 256} {
		options := map[string]any{}
		for i := 0; i < count; i++ {
			options["选项"+strconv.Itoa(i)] = nil
		}
		body, _ := json.Marshal(map[string]any{"model": "jev-1.13.0", "state": "text", "questions": map[string]any{
			"q": map[string]any{"type": "choice", "instructions": map[string]any{"goal": "choose"}, "criteria": options},
		}})
		_, err := decodeEvaluationRequest(body)
		if (err == nil) != (count == 255) {
			t.Fatalf("option bound %d: %v", count, err)
		}
	}
	for _, count := range []int{10, 11} {
		levels := make([]any, count)
		for i := range levels {
			levels[i] = []any{"level", i}
		}
		body, _ := json.Marshal(map[string]any{"model": "jev-latest", "state": "text", "questions": map[string]any{
			"q": map[string]any{"type": "score", "instructions": "rate", "criteria": levels},
		}})
		_, err := decodeEvaluationRequest(body)
		if (err == nil) != (count == 10) {
			t.Fatalf("level bound %d: %v", count, err)
		}
	}
	for name, input := range map[string]string{
		"oversize":              strings.Replace(evaluationInputFixture, "private-state-marker", strings.Repeat("x", maxEvaluationBytes), 1),
		"depth":                 strings.Replace(evaluationInputFixture, `"private-state-marker"`, strings.Repeat("[", 33)+`"x"`+strings.Repeat("]", 33), 1),
		"object-type":           strings.Replace(evaluationInputFixture, `"type":"noul"`, `"type":{}`, 1),
		"invalid-noul-criteria": strings.Replace(evaluationInputFixture, `"type":"noul"`, `"criteria":{"maybe":"yes"},"type":"noul"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeEvaluationRequest([]byte(input)); err == nil {
				t.Fatal("invalid bounded request accepted")
			}
		})
	}
}

func TestEvaluationPinnedModelAndResponseBounds(t *testing.T) {
	request, err := decodeEvaluationRequest([]byte(strings.Replace(evaluationInputFixture, "jev-latest", "jev-1.13.0", 1)))
	if err != nil {
		t.Fatal(err)
	}
	for name, response := range map[string]string{
		"pinned-model-changed": strings.Replace(evaluationOutputFixture, "jev-1.13.0", "jev-1.14.0", 1),
		"oversize":             strings.Replace(evaluationOutputFixture, "private-response-marker", strings.Repeat("x", maxResponseBytes), 1),
		"object-type":          strings.Replace(evaluationOutputFixture, `"type":"noul"`, `"type":{}`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := projectEvaluationResult([]byte(response), request); ok {
				t.Fatal("invalid response accepted")
			}
		})
	}
}

func evaluationHTTP(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{
		"X-Request-Id": {"55a3d6e3-42dd-4e67-8b08-e5c6058b2f99"},
	}, Body: io.NopCloser(strings.NewReader(body))}
}

func TestEvaluationOneSynchronousRequestAndSafeResult(t *testing.T) {
	calls := 0
	result, err, raw := evaluationCLI(t, evaluationInputFixture, func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "POST" || r.URL.String() != "https://api.puretokensx.com/typesafe/v1/systemone" ||
			r.Header.Get("Authorization") != "Bearer synthetic-evaluation-token" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("incorrect native evaluation transport")
		}
		var body, expected map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		json.Unmarshal([]byte(evaluationInputFixture), &expected)
		a, _ := json.Marshal(body)
		b, _ := json.Marshal(expected)
		if !bytes.Equal(a, b) {
			t.Error("evaluation changed state/questions or added media/stream fields")
		}
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("synchronous evaluation needs a total deadline")
		}
		return evaluationHTTP(200, evaluationOutputFixture), nil
	})
	if err != nil || calls != 1 || result["ok"] != true || result["command"] != "evaluate" ||
		result["submission_outcome"] != "accepted" || result["next_step"] != "done" || result["task_id"] != nil {
		t.Fatalf("missing synchronous completion: %v (calls=%d, err=%v)", result, calls, err)
	}
	answers := jsonObject(jsonObject(result["result"])["answers"])
	if jsonObject(answers["urgency"])["noul"] != 0.9 || jsonObject(answers["severity"])["score"] != 1.1 {
		t.Fatal("typed results were lost")
	}
	for _, marker := range []string{"private-state-marker", "private-response-marker", "synthetic-evaluation-token"} {
		if strings.Contains(raw, marker) {
			t.Fatal("evaluation receipt leaked private content")
		}
	}
	support, _ := json.Marshal(result["support"])
	if strings.Contains(string(support), "billing") || jsonObject(result["support"])["command"] != "evaluate" {
		t.Fatal("support summary must contain command evidence, not evaluation inputs or results")
	}
}

func TestEvaluationRejectsInvalidInputBeforeHTTPRequest(t *testing.T) {
	for name, input := range map[string]string{
		"missing": `{}`, "null": `null`, "trailing": evaluationInputFixture + `{}`,
		"unknown-model":         strings.Replace(evaluationInputFixture, "jev-latest", "jev", 1),
		"stream":                strings.Replace(evaluationInputFixture, `"model":`, `"stream":true,"model":`, 1),
		"endpoint":              strings.Replace(evaluationInputFixture, `"model":`, `"url":"https://other.invalid","model":`, 1),
		"empty-questions":       `{"model":"jev-latest","state":"x","questions":{}}`,
		"scalar-state":          `{"model":"jev-latest","state":42,"questions":{"q":{"type":"noul","instructions":"x"}}}`,
		"invalid-question-type": strings.Replace(evaluationInputFixture, `"type":"noul"`, `"type":"chat"`, 1),
		"bool-criterion":        strings.Replace(evaluationInputFixture, `"billing":null`, `"billing":true`, 1),
		"one-score-level":       strings.Replace(evaluationInputFixture, `["Low","Medium","High"]`, `["Low"]`, 1),
		"missing-instructions":  strings.Replace(evaluationInputFixture, `"instructions":"Is it urgent?"`, `"criteria":{"false":"No"}`, 1),
		"unsafe-question-id":    strings.Replace(evaluationInputFixture, `"urgency":`, `"https://private.invalid":`, 1),
		"duplicate-key":         strings.Replace(evaluationInputFixture, `"model":`, `"model":"jev-preview","model":`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			result, err, _ := evaluationCLI(t, input, func(*http.Request) (*http.Response, error) {
				calls++
				return evaluationHTTP(200, evaluationOutputFixture), nil
			})
			if err == nil || calls != 0 || result["local_error_code"] != "evaluation_request_invalid" ||
				result["submission_outcome"] != "not_submitted" || result["next_step"] != "await_user" {
				t.Fatalf("invalid evaluation reached network or lost local category: %v calls=%d", result, calls)
			}
		})
	}
}

func TestEvaluationUnknownResponsesNeverRetryOrPoll(t *testing.T) {
	for name, body := range map[string]string{
		"invalid-json":       `<html>private-response-marker</html>`,
		"nested":             `{"data":` + evaluationOutputFixture + `}`,
		"missing":            strings.Replace(evaluationOutputFixture, `"urgency":`, `"wrong":`, 1),
		"wrong-type":         strings.Replace(evaluationOutputFixture, `"type":"noul"`, `"type":"choice"`, 1),
		"invalid-choice":     strings.Replace(evaluationOutputFixture, `"choice":"billing"`, `"choice":"unknown"`, 1),
		"wrong-argmax":       strings.Replace(evaluationOutputFixture, `"choice":"billing"`, `"choice":"technical"`, 1),
		"bad-probability":    strings.Replace(evaluationOutputFixture, `"noul":0.9`, `"noul":2`, 1),
		"null-probability":   strings.Replace(evaluationOutputFixture, `"noul":0.9`, `"noul":null`, 1),
		"invalid-sum":        strings.Replace(evaluationOutputFixture, `"billing":0.8`, `"billing":0.5`, 1),
		"invalid-score":      strings.Replace(evaluationOutputFixture, `"score":1.1`, `"score":5`, 1),
		"inconsistent-score": strings.Replace(evaluationOutputFixture, `"score":1.1`, `"score":0.1`, 1),
		"bad-confidence":     strings.Replace(evaluationOutputFixture, `"confidence":0.5`, `"confidence":null`, 1),
		"bad-legend":         strings.Replace(evaluationOutputFixture, `"0":"Low"`, `"9":"private-response-marker"`, 1),
		"unsafe-model":       strings.Replace(evaluationOutputFixture, "jev-1.13.0", "https://private.invalid", 1),
		"duplicate-answer":   strings.Replace(evaluationOutputFixture, `"noul":0.9`, `"noul":0.1,"noul":0.9`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			result, err, raw := evaluationCLI(t, evaluationInputFixture, func(*http.Request) (*http.Response, error) {
				calls++
				return evaluationHTTP(200, body), nil
			})
			if err == nil || calls != 1 || result["ok"] != false || result["result"] != nil ||
				result["submission_outcome"] != "unknown" || result["next_step"] != "await_user" ||
				result["local_error_code"] != "evaluation_response_invalid" || strings.Contains(raw, "private-response-marker") {
				t.Fatalf("invalid result accepted/leaked or evaluation repeated: %v calls=%d", result, calls)
			}
		})
	}
}

func TestEvaluationHTTPAndTransportFailuresStop(t *testing.T) {
	for _, status := range []int{0, 302, 401, 422, 429, 500, 529, 700} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			result, err, raw := evaluationCLI(t, evaluationInputFixture, func(*http.Request) (*http.Response, error) {
				calls++
				if status == 0 {
					return nil, errors.New("private-response-marker")
				}
				response := evaluationHTTP(status, `{"detail":"private-response-marker"}`)
				response.Header.Set("Location", "https://other.invalid")
				response.Header.Set("Retry-After", "5")
				return response, nil
			})
			want := "unknown"
			if status >= 400 && status < 500 {
				want = "rejected"
			}
			if err == nil || calls != 1 || result["submission_outcome"] != want || result["next_step"] != "await_user" ||
				strings.Contains(raw, "private-response-marker") || result["result"] != nil {
				t.Fatalf("unsafe failed evaluation: %v calls=%d", result, calls)
			}
			if status > 599 && result["http_status"] != nil {
				t.Fatal("nonstandard HTTP status violates receipt schema")
			}
		})
	}
}

func TestEvaluationDiscoveryUsesOneAuthenticatedGeneralCatalogRead(t *testing.T) {
	calls := 0
	server := evaluationTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || r.URL.String() != apiOrigin+"/v1/models" {
			t.Error("evaluation discovery used media catalog or submitted work")
		}
		return evaluationHTTP(200, `{"data":[{"id":"jev-latest","owned_by":"private-provider"},{"id":"jev-1.13.0"},{"id":"not-a-jev"},{"id":"jev-malicious"},{"id":"jev-latest"}]}`), nil
	})
	svc := service{baseURL: apiOrigin, client: &http.Client{Transport: server}, token: "synthetic-fixture"}
	var output bytes.Buffer
	err := executeModelQuery(&output, svc, strings.NewReader(`{"kind":"evaluation"}`))
	var document map[string]any
	json.Unmarshal(output.Bytes(), &document)
	result := jsonObject(document["result"])
	data, _ := result["data"].([]any)
	if err != nil || calls != 1 || len(data) != 2 || strings.Contains(output.String(), "private-provider") ||
		result["matching_scope"] != "reviewed_evaluation_models_listed_by_api" {
		t.Fatalf("evaluation discovery not supported safely: %s (%v)", output.String(), err)
	}
}
