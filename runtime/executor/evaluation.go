package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const evaluationPath = "/typesafe/v1/systemone"
const maxEvaluationBytes = 1 << 20
const evaluationTimeout = 90 * time.Second

var evaluationModels = map[string]bool{"jev-latest": true, "jev-1.13.0": true, "jev-preview": true}
var evaluationQuestionID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
var evaluationVersion = regexp.MustCompile(`^jev-[0-9]{1,6}\.[0-9]{1,6}\.[0-9]{1,6}$`)

// These objects never become task records or support-summary data.
type evaluationRequest struct {
	Model     string                    `json:"model"`
	State     any                       `json:"state"`
	Questions map[string]map[string]any `json:"questions"`
}

type evaluationReceipt struct {
	OK                bool           `json:"ok"`
	Command           string         `json:"command"`
	SubmissionOutcome string         `json:"submission_outcome"`
	NextStep          string         `json:"next_step"`
	Result            map[string]any `json:"result,omitempty"`
	FailurePhase      string         `json:"failure_phase,omitempty"`
	LocalErrorCode    string         `json:"local_error_code,omitempty"`
	HTTPStatus        int            `json:"http_status,omitempty"`
	APIErrorCode      string         `json:"api_error_code,omitempty"`
	ErrorMessage      string         `json:"error_message,omitempty"`
	NextAction        string         `json:"next_action,omitempty"`
	RetryAfterSecs    int            `json:"retry_after_seconds,omitempty"`
}

func evaluationFailure(outcome, phase, code, message string, status int) evaluationReceipt {
	if status < 100 || status > 599 {
		status = 0
	}
	return evaluationReceipt{Command: "evaluate", SubmissionOutcome: outcome, NextStep: "await_user",
		FailurePhase: phase, LocalErrorCode: code, HTTPStatus: status, ErrorMessage: message,
		NextAction: "Stop here. Do not retry automatically, poll a media task, change host or infer charges. Use the existing safe support summary for diagnosis."}
}

func runEvaluationCommand(output io.Writer, svc service, host, file string) error {
	var data []byte
	var err error
	if filepath.IsAbs(file) {
		data, err = readBoundedFile(file, maxEvaluationBytes)
	}
	request, decodeErr := decodeEvaluationRequest(data)
	if file == "" || !filepath.IsAbs(file) || err != nil || decodeErr != nil {
		writeJSON(output, evaluationFailure("not_submitted", "validation", "evaluation_request_invalid",
			"Use one absolute UTF-8 request file with a reviewed Jev model, text or structured state, and 1–64 valid typed questions. No API request was sent.", 0))
		return errors.New("invalid evaluation request")
	}
	token, err := credentialForHost(host)
	if err != nil {
		code, message, next := credentialFailureDetails(err)
		result := evaluationFailure("not_submitted", "validation", code, message, 0)
		result.NextAction = next
		writeJSON(output, result)
		return err
	}
	defer clearString(&token)
	svc.token = token
	return executeEvaluation(output, svc, request)
}

func executeEvaluation(output io.Writer, svc service, request evaluationRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err // The validated request contains JSON values only.
	}
	ctx, cancel := context.WithTimeout(context.Background(), evaluationTimeout)
	defer cancel()
	response, status, retry, apiCode, message, err := svc.request(ctx, http.MethodPost, evaluationPath, bytes.NewReader(body), "application/json")
	if err != nil || status < 200 || status >= 300 {
		outcome := "unknown"
		if err == nil && status >= 400 && status < 500 {
			outcome = "rejected"
		}
		local := "evaluation_request_failed"
		if err != nil && status >= 200 && status < 300 {
			local = "evaluation_response_invalid"
		}
		if message == "" {
			message = "The synchronous evaluation did not return a verified result."
		}
		result := evaluationFailure(outcome, "submission", local, message, status)
		result.APIErrorCode, result.RetryAfterSecs = apiCode, retry
		writeJSON(output, result)
		return errors.New("evaluation failed")
	}
	result, valid := projectEvaluationResult(response, request)
	if !valid {
		failure := evaluationFailure("unknown", "submission", "evaluation_response_invalid",
			"The evaluation response could not be verified against the requested questions. This does not prove that processing failed or that no charge occurred.", status)
		failure.RetryAfterSecs = retry
		writeJSON(output, failure)
		return errors.New("invalid evaluation response")
	}
	writeJSON(output, evaluationReceipt{OK: true, Command: "evaluate", SubmissionOutcome: "accepted", NextStep: "done", Result: result})
	return nil
}

// JSON duplicates otherwise silently overwrite a question or probability.
// Enforce the same boundary on requests and responses before projecting.
func strictEvaluationJSON(data []byte, limit int) (any, error) {
	if len(data) == 0 || len(data) > limit || !utf8.Valid(data) {
		return nil, errors.New("invalid evaluation JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	// Preserve identifiers, high-precision decimals and other structured state
	// exactly when re-encoding the request. Only result probabilities become
	// floating-point numbers for validation.
	decoder.UseNumber()
	var read func(int) (any, error)
	read = func(depth int) (any, error) {
		if depth > 32 {
			return nil, errors.New("evaluation JSON too deep")
		}
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		delim, container := token.(json.Delim)
		if !container {
			return token, nil
		}
		switch delim {
		case '{':
			object := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				key, ok := keyToken.(string)
				if err != nil || !ok {
					return nil, errors.New("invalid object key")
				}
				if _, exists := object[key]; exists {
					return nil, errors.New("duplicate object key")
				}
				value, err := read(depth + 1)
				if err != nil {
					return nil, err
				}
				object[key] = value
			}
			if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
				return nil, errors.New("invalid object")
			}
			return object, nil
		case '[':
			items := []any{}
			for decoder.More() {
				item, err := read(depth + 1)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
			}
			if end, err := decoder.Token(); err != nil || end != json.Delim(']') {
				return nil, errors.New("invalid array")
			}
			return items, nil
		default:
			return nil, errors.New("invalid JSON delimiter")
		}
	}
	value, err := read(0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("trailing evaluation JSON")
	}
	return value, nil
}

func evaluationEntry(value any) bool {
	switch value.(type) {
	case string, map[string]any, []any:
		return true
	default:
		return false
	}
}

func evaluationLabel(label string) bool {
	return strings.TrimSpace(label) != "" && utf8.RuneCountInString(label) <= 128 &&
		strings.IndexFunc(label, unicode.IsControl) < 0
}

func decodeEvaluationRequest(data []byte) (evaluationRequest, error) {
	fail := errors.New("invalid evaluation request")
	request := evaluationRequest{}
	if len(data) > maxEvaluationBytes {
		return request, fail
	}
	// Windows PowerShell 5.1 writes this prefix for UTF-8 files.
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	value, err := strictEvaluationJSON(data, maxEvaluationBytes)
	body, ok := value.(map[string]any)
	if err != nil || !ok || len(body) != 3 || !evaluationEntry(body["state"]) {
		return request, fail
	}
	model, _ := body["model"].(string)
	questions, ok := body["questions"].(map[string]any)
	if !evaluationModels[model] || !ok || len(questions) < 1 || len(questions) > 64 {
		return request, fail
	}
	request = evaluationRequest{Model: model, State: body["state"], Questions: map[string]map[string]any{}}
	for id, value := range questions {
		question, ok := value.(map[string]any)
		if !evaluationQuestionID.MatchString(id) || !ok || !evaluationEntry(question["instructions"]) {
			return request, fail
		}
		for key := range question {
			if key != "type" && key != "instructions" && key != "criteria" {
				return request, fail
			}
		}
		switch question["type"] {
		case "choice":
			options, ok := question["criteria"].(map[string]any)
			if !ok || len(options) < 1 || len(options) > 255 {
				return request, fail
			}
			for name, description := range options {
				if !evaluationLabel(name) || (description != nil && !evaluationEntry(description)) {
					return request, fail
				}
			}
		case "score":
			levels, ok := question["criteria"].([]any)
			if !ok || len(levels) < 2 || len(levels) > 10 {
				return request, fail
			}
			for _, level := range levels {
				if !evaluationEntry(level) {
					return request, fail
				}
			}
		case "noul":
			if criteria, exists := question["criteria"]; exists {
				descriptions, ok := criteria.(map[string]any)
				if !ok {
					return request, fail
				}
				for label, description := range descriptions {
					if (label != "true" && label != "false") || !evaluationEntry(description) {
						return request, fail
					}
				}
			}
		default:
			return request, fail
		}
		request.Questions[id] = question
	}
	return request, nil
}

func evaluationNumber(value any, maximum float64) (float64, bool) {
	raw, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	number, err := raw.Float64()
	return number, err == nil && !math.IsNaN(number) && !math.IsInf(number, 0) && number >= 0 && number <= maximum
}

func projectEvaluationResult(data []byte, request evaluationRequest) (map[string]any, bool) {
	value, err := strictEvaluationJSON(data, maxResponseBytes)
	body, ok := value.(map[string]any)
	if err != nil || !ok || body["error"] != nil || body["success"] == false {
		return nil, false
	}
	model, _ := body["model"].(string)
	if !evaluationModels[model] && !evaluationVersion.MatchString(model) {
		return nil, false
	}
	if evaluationVersion.MatchString(request.Model) && model != request.Model {
		return nil, false
	}
	answers, ok := body["answers"].(map[string]any)
	if !ok || len(answers) != len(request.Questions) {
		return nil, false
	}
	projected := map[string]any{}
	for id, question := range request.Questions {
		answer, ok := answers[id].(map[string]any)
		answerType, _ := answer["type"].(string)
		questionType, _ := question["type"].(string)
		if !ok || answerType != questionType {
			return nil, false
		}
		typed := map[string]any{"type": question["type"]}
		if question["type"] == "noul" {
			number, valid := evaluationNumber(answer["noul"], 1)
			if !valid {
				return nil, false
			}
			typed["noul"] = number
		} else {
			keys := map[string]bool{}
			levels, _ := question["criteria"].([]any)
			if question["type"] == "choice" {
				for key := range question["criteria"].(map[string]any) {
					keys[key] = true
				}
			} else {
				for index := range levels {
					keys[strconv.Itoa(index)] = true
				}
			}
			confidence, valid := evaluationNumber(answer["confidence"], 1)
			probabilities, ok := answer["probabilities"].(map[string]any)
			if !valid || !ok || len(probabilities) != len(keys) {
				return nil, false
			}
			sum, maximum, weighted := 0.0, 0.0, 0.0
			for key, value := range probabilities {
				probability, valid := evaluationNumber(value, 1)
				if !keys[key] || !valid {
					return nil, false
				}
				sum += probability
				maximum = math.Max(maximum, probability)
				index, _ := strconv.Atoi(key)
				weighted += float64(index) * probability
			}
			if math.Abs(sum-1) > 0.001 {
				return nil, false
			}
			typed["probabilities"], typed["confidence"] = probabilities, confidence
			if question["type"] == "choice" {
				choice, _ := answer["choice"].(string)
				probability, valid := evaluationNumber(probabilities[choice], 1)
				if !keys[choice] || !valid || probability < maximum-0.001 {
					return nil, false
				}
				typed["choice"] = choice
			} else {
				score, valid := evaluationNumber(answer["score"], float64(len(levels)-1))
				legend, ok := answer["legend"].(map[string]any)
				if !valid || math.Abs(score-weighted) > 0.001 || !ok || len(legend) != len(keys) {
					return nil, false
				}
				for key, value := range legend {
					if _, ok := value.(string); !keys[key] || !ok {
						return nil, false
					}
				}
				typed["score"], typed["level_count"] = score, len(levels)
			}
		}
		projected[id] = typed
	}
	result := map[string]any{"model": model, "answers": projected}
	usage := map[string]any{}
	for key, value := range jsonObject(body["usage"]) {
		if key == "input_tokens" || key == "output_tokens" {
			if number, valid := evaluationNumber(value, 9007199254740991); valid && math.Trunc(number) == number {
				usage[key] = number
			}
		}
	}
	if len(usage) > 0 {
		result["usage"] = usage
	}
	return result, true
}
