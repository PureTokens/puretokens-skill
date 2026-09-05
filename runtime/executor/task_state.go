package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func installedSkillsRoot() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(filepath.Dir(executable))
}
func rejectRedirect(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
func fixedClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	copy := *client
	copy.CheckRedirect = rejectRedirect
	return &copy
}
func validTaskID(id string) bool { return regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`).MatchString(id) }
func safePublicCode(code string) string {
	if !publicCodePattern.MatchString(code) {
		return ""
	}
	return safePublicString(code)
}
func taskReceipt(request taskRequest, id, status string) receipt {
	parameters := make(map[string]any)
	// Media URLs, prompts and file paths are deliberately excluded from receipts.
	for _, key := range []string{"n", "size", "image_size", "aspect_ratio", "width", "height", "duration", "resolution", "generate_audio", "strength"} {
		if value, exists := request.Parameters[key]; exists {
			if text, ok := value.(string); ok && safePublicString(text) == "" {
				continue
			}
			parameters[key] = value
		}
	}
	return receipt{OK: true, Kind: request.Kind, Operation: request.Operation, Model: request.Model, TaskID: id, Status: status, RequestedCount: max(1, request.RequestedCount), Parameters: parameters}
}
func mergeFailure(current, failure receipt) receipt {
	current.OK = false
	current.FailurePhase = failure.FailurePhase
	current.HTTPStatus = failure.HTTPStatus
	current.APIErrorCode = failure.APIErrorCode
	current.ErrorMessage = failure.ErrorMessage
	current.NextAction = failure.NextAction
	current.RetryAfterSecs = failure.RetryAfterSecs
	return current
}
func applyTaskStatus(result receipt, body []byte, httpStatus int) (receipt, error) {
	id, state, reconcile, ok := taskIdentity(body)
	if !ok || id != result.TaskID {
		return mergeFailure(result, receipt{FailurePhase: "status", HTTPStatus: httpStatus, ErrorMessage: "The status response did not identify this task.", NextAction: "Keep this task ID; do not create another task."}), errors.New("task identity mismatch")
	}
	result.Status = state
	result.ReconciliationRequired = reconcile
	if reconcile {
		result.OK = false
		result.FailurePhase = "status"
		result.ErrorMessage = "The task requires server-side reconciliation."
		result.NextAction = "Keep this task ID and check the same task later; do not infer a refund or submit a replacement."
		return result, errors.New("reconciliation required")
	}
	if terminalFailure(state) {
		code, message := publicAPIError(body)
		if message == "" {
			message = "The API reported that this task failed."
		}
		return mergeFailure(result, apiFailure("status", httpStatus, 0, code, message, "Review this task's error. Do not retry or change models automatically.")), errors.New("task failed")
	}
	if !terminalSuccess(state) && state != "pending" && state != "queued" && state != "processing" && state != "running" && state != "in_progress" {
		result.Status = "unknown"
		return mergeFailure(result, receipt{FailurePhase: "status", ErrorMessage: "The API returned an unsupported task state.", NextAction: "Keep this task ID and check its status later; do not resubmit."}), errors.New("unknown task state")
	}
	return result, nil
}
func executeExistingTask(command string, input io.Reader, output io.Writer, svc service) error {
	request, err := decodeTaskRequest(input)
	if err != nil {
		writeReceipt(output, validationFailure("A complete same-task request JSON is required."))
		return err
	}
	request.Operation = "continue"
	if err = validateTaskRequest(request); err != nil || !validTaskID(request.TaskID) {
		writeReceipt(output, validationFailure("Choose image or video and a valid existing task ID."))
		return errors.New("invalid existing task")
	}
	result := taskReceipt(request, request.TaskID, "pending")
	if command == "wait" {
		return pollAndDeliver(output, svc, request, result)
	}
	if command == "content" {
		// The caller must supply the completed state from the last same-task receipt.
		// The content endpoint itself is authoritative and rejects unfinished tasks.
		if !terminalSuccess(request.TaskStatus) {
			writeReceipt(output, mergeFailure(result, validationFailure("Read the same task status first; content requires task_status completed.")))
			return errors.New("completion not confirmed")
		}
		result.Status = "completed"
		return deliverTask(output, svc, request, result)
	}
	body, status, retry, code, message, err := svc.request(context.Background(), http.MethodGet, statusPath(request.Kind, result.TaskID), nil, "")
	if err != nil || status < 200 || status >= 300 {
		writeReceipt(output, mergeFailure(result, apiFailure("status", status, retry, code, message, "Keep this task ID; continue the same task later.")))
		return errors.New("status failed")
	}
	result, err = applyTaskStatus(result, body, status)
	result.RetryAfterSecs = retry
	if err == nil {
		result.NextAction = "Continue this same task if pending; use content if completed. Never submit it again."
	}
	writeReceipt(output, result)
	return err
}

func readAPIObject(body []byte) (map[string]any, error) {
	var result map[string]any
	if json.Unmarshal(body, &result) != nil || result == nil {
		return nil, errors.New("invalid response")
	}
	if result["error"] != nil || result["success"] == false {
		return nil, errors.New("API logical error")
	}
	return result, nil
}
func scrubSecret(body []byte, token string) []byte {
	if token == "" {
		return body
	}
	return []byte(strings.ReplaceAll(string(body), token, "[redacted credential]"))
}

func projectCatalog(value any) any {
	document := jsonObject(value)
	projected := []any{}
	for _, entry := range jsonArray(document["data"]) {
		model := jsonObject(entry)
		id := jsonString(model["id"])
		if !regexp.MustCompile(`^[A-Za-z0-9_.-]{1,160}$`).MatchString(id) {
			continue
		}
		result := map[string]any{"id": id}
		caps := []string{}
		for _, cap := range jsonArray(model["capabilities"]) {
			if cap == "image" || cap == "video" {
				caps = append(caps, cap.(string))
			}
		}
		result["capabilities"] = caps
		schema := jsonObject(model["input_schema"])
		selected := map[string]any{}
		for _, field := range []string{"properties", "constraints", "operations", "lifecycle", "outputSizes"} {
			if v, ok := schema[field]; ok {
				selected[field] = publicJSON(v, 0)
			}
		}
		result["input_schema"] = selected
		projected = append(projected, result)
	}
	return map[string]any{"data": projected}
}

// Catalog JSON is data, never instructions or a channel for raw error/config data.
func publicJSON(value any, depth int) any {
	if depth > 16 {
		return nil
	}
	switch v := value.(type) {
	case string:
		return safePublicString(v)
	case []any:
		result := make([]any, 0, len(v))
		for _, item := range v {
			result = append(result, publicJSON(item, depth+1))
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for key, item := range v {
			if !publicCodePattern.MatchString(key) {
				continue
			}
			switch strings.ToLower(key) {
			case "apikey", "api_key", "token", "secret", "authorization", "headers", "provider", "upstream":
				continue
			}
			result[key] = publicJSON(item, depth+1)
		}
		return result
	case float64, bool, nil:
		return v
	default:
		return nil
	}
}
