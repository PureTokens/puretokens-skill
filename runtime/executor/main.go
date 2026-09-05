// Pure Tokens API executor is the small, self-contained process used by the
// installable Skills. It intentionally has no dependency on Node, Python,
// MCP, a proxy, or a background service.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	apiOrigin              = "https://api.puretokensx.com"
	maxConfigBytes         = 1 << 20
	maxResponseBytes       = 2 << 20
	maxImageBytes    int64 = 25 << 20
	maxVideoBytes    int64 = 512 << 20
)

var executorVersion = "0.15.2"

type attachment struct {
	Field string `json:"field"`
	Path  string `json:"path"`
}

type taskRequest struct {
	Kind           string         `json:"kind"`
	Operation      string         `json:"operation"`
	Model          string         `json:"model"`
	Prompt         string         `json:"prompt"`
	TaskID         string         `json:"task_id"`
	Parameters     map[string]any `json:"parameters"`
	Attachments    []attachment   `json:"attachments"`
	RequestedCount int            `json:"requested_count"`
	OutputDir      string         `json:"output_dir"`
	Poll           *pollRequest   `json:"poll"`
}

type pollRequest struct {
	MaxStatusReads int `json:"max_status_reads"`
	DeadlineSecs   int `json:"deadline_seconds"`
}

type receipt struct {
	OK                     bool     `json:"ok"`
	Kind                   string   `json:"kind,omitempty"`
	Operation              string   `json:"operation,omitempty"`
	Model                  string   `json:"model,omitempty"`
	TaskID                 string   `json:"task_id,omitempty"`
	Status                 string   `json:"status,omitempty"`
	ReconciliationRequired bool     `json:"reconciliation_required,omitempty"`
	DeliveredPaths         []string `json:"delivered_paths,omitempty"`
	DeliveredCount         int      `json:"delivered_count,omitempty"`
	FailurePhase           string   `json:"failure_phase,omitempty"`
	HTTPStatus             int      `json:"http_status,omitempty"`
	APIErrorCode           string   `json:"api_error_code,omitempty"`
	ErrorMessage           string   `json:"error_message,omitempty"`
	NextAction             string   `json:"next_action,omitempty"`
	RetryAfterSecs         int      `json:"retry_after_seconds,omitempty"`
}

type initReceipt struct {
	OK                   bool     `json:"ok"`
	Command              string   `json:"command"`
	ExecutorVersion      string   `json:"executor_version"`
	ConfigurationStatus  string   `json:"configuration_status"`
	APIRequestExecuted   bool     `json:"api_request_executed"`
	APIIdentityConfirmed bool     `json:"api_identity_confirmed"`
	HTTPStatus           int      `json:"http_status,omitempty"`
	Message              string   `json:"message"`
	NextAction           string   `json:"next_action"`
	UsageExamples        []string `json:"usage_examples"`
}

type service struct {
	baseURL string
	client  *http.Client
	token   string
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		// Every expected failure is emitted by run as exactly one sanitized
		// receipt. Never append a second receipt that could make callers think a
		// request was repeated or had a different outcome.
		os.Exit(1)
	}
}

func run(args []string, input io.Reader, output io.Writer) error {
	if len(args) == 1 && args[0] == "--version" {
		_, err := fmt.Fprintln(output, executorVersion)
		return err
	}
	if len(args) < 1 {
		writeReceipt(output, validationFailure("Choose one of: init, connection, balance, models, task."))
		return errors.New("missing command")
	}

	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	host := flags.String("host", "", "")
	if err := flags.Parse(args[1:]); err != nil || *host == "" {
		writeReceipt(output, validationFailure("The current host must be identified before the API request can start."))
		return errors.New("invalid command")
	}

	token, err := credentialForHost(*host)
	if err != nil {
		if command == "init" {
			writeJSON(output, initCredentialFailure(err))
			return err
		}
		writeReceipt(output, receipt{OK: false, FailurePhase: "validation", APIErrorCode: "credential_unavailable", ErrorMessage: "The current host did not provide a verified Pure Tokens credential for this request.", NextAction: "Verify the Pure Tokens connection in the host, then start a new session."})
		return err
	}
	defer clearString(&token)
	svc := service{baseURL: apiOrigin, token: token, client: &http.Client{Timeout: 90 * time.Second}}

	switch command {
	case "init":
		return executeInit(output, svc)
	case "connection":
		return executeReadOnly(output, svc, "/v1", "connection")
	case "balance":
		return executeReadOnly(output, svc, "/api/product/desktop/account/balance", "balance")
	case "models":
		return executeReadOnly(output, svc, "/v1/media/models", "models")
	case "task":
		return executeTask(input, output, svc)
	default:
		writeReceipt(output, validationFailure("Choose one of: init, connection, balance, models, task."))
		return errors.New("unknown command")
	}
}

func executeInit(output io.Writer, svc service) error {
	result := initReceipt{Command: "init", ExecutorVersion: executorVersion, ConfigurationStatus: "unverified", UsageExamples: usageExamples()}
	body, status, _, _, _, err := svc.request(context.Background(), http.MethodGet, "/v1", nil, "")
	result.APIRequestExecuted = status > 0
	result.HTTPStatus = status
	if err != nil {
		if status > 0 {
			result.ConfigurationStatus = "api_response_unreadable"
			result.Message = "The fixed API responded, but its identity response could not be read safely."
			result.NextAction = "Check the network path to Pure Tokens, then run init again."
		} else {
			result.ConfigurationStatus = "api_network_unavailable"
			result.Message = "The fixed Pure Tokens API could not be reached from this host session."
			result.NextAction = "Check this host's network permission and connection, then run init again."
		}
		writeJSON(output, result)
		return errors.New("init API request failed")
	}
	if status < 200 || status >= 300 {
		result.ConfigurationStatus = "api_identity_rejected"
		result.Message = "The fixed Pure Tokens API rejected the identity check."
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			result.NextAction = "Verify the active Pure Tokens connection and key in the host, then run init again."
		} else {
			result.NextAction = "Check the current connection and network, then run init again."
		}
		writeJSON(output, result)
		return errors.New("init connection check failed")
	}
	var identity map[string]any
	if json.Unmarshal(body, &identity) != nil {
		result.ConfigurationStatus = "api_identity_unreadable"
		result.Message = "The fixed Pure Tokens API returned an unreadable identity response."
		result.NextAction = "Run init again in a new host session."
		writeJSON(output, result)
		return errors.New("init identity unreadable")
	}
	statusValue, _ := identity["status"].(string)
	name, _ := identity["name"].(string)
	basePath, _ := identity["base_url"].(string)
	if statusValue == "ok" && name == "Pure Tokens API" && basePath == "/v1" {
		result.OK = true
		result.ConfigurationStatus = "verified"
		result.APIIdentityConfirmed = true
		result.Message = "The current host can reach the fixed Pure Tokens API."
		result.NextAction = "Use puretokens-image, puretokens-video, puretokens-balance, puretokens-models, or puretokens-connection in a new host conversation."
	} else {
		result.ConfigurationStatus = "api_identity_unconfirmed"
		result.Message = "The fixed Pure Tokens API did not return a confirmable identity."
		result.NextAction = "Verify the Pure Tokens connection in the host, then run init again."
	}
	writeJSON(output, result)
	if !result.OK {
		return errors.New("init identity unconfirmed")
	}
	return nil
}

func initCredentialFailure(err error) initReceipt {
	status, message, nextAction := credentialFailureDetails(err)
	return initReceipt{
		Command:             "init",
		ExecutorVersion:     executorVersion,
		ConfigurationStatus: status,
		Message:             message,
		NextAction:          nextAction,
		UsageExamples:       usageExamples(),
	}
}

func usageExamples() []string {
	return []string{
		"Generate an image: create an image of a misty mountain lake at sunrise.",
		"Choose an image model: use grok-imagine-image to create a product illustration.",
		"Generate multiple variants: create 3 variations of the same image brief.",
		"Generate a video: create a 5-second cinematic shot of a paper boat on a river.",
		"Use a reference: use the attached image as the first frame or visual reference, and say which role you intend.",
		"Check current models: ask which Pure Tokens image/video models and operations are currently available.",
		"Check balance: ask for the current Pure Tokens balance.",
		"Update Skills: ask to update Pure Tokens Skills, then start a new host conversation.",
	}
}

func credentialForHost(host string) (string, error) {
	// A host adapter must only inspect its exact documented effective
	// connection files. It must not scan the home directory or fall back to a
	// random environment variable, because that could send an unrelated key to
	// the fixed Pure Tokens endpoint.
	switch host {
	case "codex":
		return credentialFromCodex()
	case "claude-code":
		return credentialFromClaudeCode()
	case "gemini-cli":
		return credentialFromGeminiCLI()
	case "workbuddy":
		return credentialFromWorkBuddy()
	case "grok-build":
		return credentialFromGrokBuild()
	case "opencode":
		return credentialFromOpenCode()
	default:
		return "", credentialFailure("host_credential_adapter_unavailable", "This host has no supported managed connection record for API requests.", "Use a supported host with a configured Pure Tokens connection, then run init again.")
	}
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := io.LimitReader(file, limit+1)
	data, err := io.ReadAll(reader)
	if err != nil || int64(len(data)) > limit {
		return nil, errors.New("configuration exceeds supported size")
	}
	return data, nil
}

func executeReadOnly(output io.Writer, svc service, path string, command string) error {
	body, status, retryAfter, apiCode, apiMessage, err := svc.request(context.Background(), http.MethodGet, path, nil, "")
	if err != nil || status < 200 || status >= 300 {
		writeReceipt(output, apiFailure("submission", status, retryAfter, apiCode, apiMessage, "The read-only API request did not complete. Verify the current Pure Tokens connection, then try again."))
		return errors.New("read-only request failed")
	}
	var payload any
	if json.Unmarshal(body, &payload) != nil {
		writeReceipt(output, receipt{OK: false, FailurePhase: "submission", HTTPStatus: status, ErrorMessage: "The API returned an unreadable response.", NextAction: "Try again in a new host session."})
		return errors.New("invalid json")
	}
	// The command result is retained as JSON, but no host configuration or
	// credential is added to it. The calling Skill selects its user-visible
	// subset from the API response.
	writeJSON(output, map[string]any{"ok": true, "command": command, "result": payload})
	return nil
}

func executeTask(input io.Reader, output io.Writer, svc service) error {
	request, err := decodeTaskRequest(input)
	if err != nil {
		writeReceipt(output, validationFailure("The media request is incomplete or invalid; no API request was sent."))
		return err
	}
	if err := validateTaskRequest(request); err != nil {
		writeReceipt(output, validationFailure("The requested model operation or attachment input is invalid; no API request was sent."))
		return err
	}
	if request.Operation == "continue" {
		result := receipt{OK: true, Kind: request.Kind, Operation: request.Operation, TaskID: request.TaskID, Status: "pending"}
		return pollAndDeliver(output, svc, request, result)
	}

	path, contentType, body, err := taskRequestBody(request)
	if err != nil {
		writeReceipt(output, validationFailure("The current attachment could not be prepared; no API request was sent."))
		return err
	}
	response, status, retryAfter, apiCode, apiMessage, err := svc.request(context.Background(), http.MethodPost, path, body, contentType)
	if err != nil || status < 200 || status >= 300 {
		writeReceipt(output, apiFailure("submission", status, retryAfter, apiCode, apiMessage, "The media API did not accept the request. Correct the request only after reviewing the returned error."))
		return errors.New("submission failed")
	}

	taskID, state, reconciliationRequired, ok := taskIdentity(response)
	if !ok {
		writeReceipt(output, receipt{OK: false, Kind: request.Kind, Operation: request.Operation, Model: request.Model, FailurePhase: "submission", HTTPStatus: status, ErrorMessage: "The API accepted a response but did not return a usable task ID. The submission outcome is unknown.", NextAction: "Do not resubmit automatically. Ask the user before creating a new billable request."})
		return errors.New("task id missing")
	}
	result := receipt{OK: true, Kind: request.Kind, Operation: request.Operation, Model: request.Model, TaskID: taskID, Status: state, ReconciliationRequired: reconciliationRequired}
	if reconciliationRequired {
		result.OK = false
		result.FailurePhase = "status"
		result.ErrorMessage = "The task requires server-side reconciliation before its result can be confirmed."
		result.NextAction = "Keep this task ID and ask the user before continuing the same task. Do not submit a replacement task."
		writeReceipt(output, result)
		return errors.New("task reconciliation required")
	}
	if terminalFailure(state) {
		result.OK = false
		result.FailurePhase = "submission"
		result.ErrorMessage = "The API returned a failed task state."
		result.NextAction = "Review the task error before creating a new request."
		writeReceipt(output, result)
		return errors.New("initial failed state")
	}
	if terminalSuccess(state) {
		return deliverTask(output, svc, request, result)
	}
	return pollAndDeliver(output, svc, request, result)
}

func decodeTaskRequest(input io.Reader) (taskRequest, error) {
	var request taskRequest
	decoder := json.NewDecoder(io.LimitReader(input, maxResponseBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return taskRequest{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return taskRequest{}, errors.New("task input must contain exactly one JSON object")
	}
	return request, nil
}

func validateTaskRequest(request taskRequest) error {
	if request.Kind != "image" && request.Kind != "video" {
		return errors.New("invalid kind")
	}
	if request.Operation != "generate" && request.Operation != "edit" && request.Operation != "continue" {
		return errors.New("invalid operation")
	}
	if request.RequestedCount < 0 || request.RequestedCount > 6 {
		return errors.New("invalid requested count")
	}
	if request.OutputDir != "" {
		info, err := os.Stat(request.OutputDir)
		if err != nil || !info.IsDir() {
			return errors.New("invalid output directory")
		}
	}
	if request.Operation == "continue" {
		if strings.TrimSpace(request.TaskID) == "" || strings.Contains(request.TaskID, "/") || strings.Contains(request.TaskID, "://") || len(request.Attachments) != 0 {
			return errors.New("invalid continuation")
		}
		return nil
	}
	if strings.TrimSpace(request.Model) == "" || strings.TrimSpace(request.Prompt) == "" {
		return errors.New("missing model or prompt")
	}
	if request.Operation == "edit" && len(request.Attachments) == 0 {
		return errors.New("missing edit attachment")
	}
	limit := attachmentLimit(request.Kind)
	var total int64
	for _, attachment := range request.Attachments {
		if attachment.Field == "" || !filepath.IsAbs(attachment.Path) {
			return errors.New("attachment must have a field and absolute path")
		}
		info, err := os.Stat(attachment.Path)
		if err != nil || info.IsDir() || info.Size() < 1 || info.Size() > limit {
			return errors.New("invalid attachment")
		}
		total += info.Size()
		if total > limit {
			return errors.New("attachments exceed the request size limit")
		}
	}
	return nil
}

func attachmentLimit(kind string) int64 {
	if kind == "image" {
		return maxImageBytes
	}
	return maxVideoBytes
}

func taskRequestBody(request taskRequest) (string, string, io.Reader, error) {
	path := endpointFor(request)
	if len(request.Attachments) == 0 {
		payload := map[string]any{"model": request.Model, "prompt": request.Prompt}
		for key, value := range request.Parameters {
			if key == "model" || key == "prompt" || key == "async" || key == "image" || key == "video" {
				return "", "", nil, errors.New("reserved parameter")
			}
			payload[key] = value
		}
		if request.Kind == "image" {
			payload["async"] = true
		}
		encoded, err := json.Marshal(payload)
		return path, "application/json", bytes.NewReader(encoded), err
	}

	reader, writerPipe := io.Pipe()
	writer := multipart.NewWriter(writerPipe)
	fields := map[string]string{"model": request.Model, "prompt": request.Prompt}
	if request.Kind == "image" {
		fields["async"] = "true"
	}
	for key, value := range request.Parameters {
		if key == "model" || key == "prompt" || key == "async" || key == "image" || key == "video" {
			return "", "", nil, errors.New("reserved parameter")
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return "", "", nil, err
		}
		fields[key] = strings.Trim(string(encoded), "\"")
	}
	contentType := writer.FormDataContentType()
	go func() {
		defer writerPipe.Close()
		for key, value := range fields {
			if err := writer.WriteField(key, value); err != nil {
				_ = writerPipe.CloseWithError(err)
				return
			}
		}
		limit := attachmentLimit(request.Kind)
		for _, attachment := range request.Attachments {
			file, err := os.Open(attachment.Path)
			if err != nil {
				_ = writerPipe.CloseWithError(err)
				return
			}
			part, err := writer.CreateFormFile(attachment.Field, filepath.Base(attachment.Path))
			if err == nil {
				_, err = io.Copy(part, io.LimitReader(file, limit+1))
			}
			closeErr := file.Close()
			if err != nil {
				_ = writerPipe.CloseWithError(err)
				return
			}
			if closeErr != nil {
				_ = writerPipe.CloseWithError(closeErr)
				return
			}
		}
		if err := writer.Close(); err != nil {
			_ = writerPipe.CloseWithError(err)
		}
	}()
	return path, contentType, reader, nil
}

func endpointFor(request taskRequest) string {
	if request.Kind == "image" {
		if request.Operation == "edit" {
			return "/v1/images/edits"
		}
		return "/v1/images/generations"
	}
	if request.Operation == "edit" {
		return "/v1/videos/edits"
	}
	return "/v1/videos"
}

func pollAndDeliver(output io.Writer, svc service, request taskRequest, result receipt) error {
	policy := pollingPolicy(request.Kind, request.Poll)
	deadline := time.Now().Add(time.Duration(policy.deadline) * time.Second)
	for read := 0; read < policy.maxReads; read++ {
		if read > 0 {
			delay := policy.delays[min(read-1, len(policy.delays)-1)]
			if time.Now().Add(time.Duration(delay) * time.Second).After(deadline) {
				break
			}
			time.Sleep(time.Duration(delay) * time.Second)
		}
		path := statusPath(request.Kind, result.TaskID)
		body, status, retryAfter, apiCode, apiMessage, err := svc.request(context.Background(), http.MethodGet, path, nil, "")
		if err != nil || status < 200 || status >= 300 {
			if status == http.StatusTooManyRequests && retryAfter > 0 && time.Now().Add(time.Duration(retryAfter)*time.Second).Before(deadline) {
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			}
			writeReceipt(output, apiFailure("status", status, retryAfter, apiCode, apiMessage, "The same task could not be read. Ask the user before continuing this task."))
			return errors.New("status read failed")
		}
		_, state, reconciliationRequired, ok := taskIdentity(body)
		if !ok {
			writeReceipt(output, receipt{OK: false, Kind: request.Kind, Operation: request.Operation, Model: request.Model, TaskID: result.TaskID, FailurePhase: "status", HTTPStatus: status, ErrorMessage: "The task status response was not usable.", NextAction: "Ask the user before continuing this same task."})
			return errors.New("unreadable status")
		}
		result.Status = state
		result.ReconciliationRequired = reconciliationRequired
		if reconciliationRequired {
			result.OK = false
			result.FailurePhase = "status"
			result.ErrorMessage = "The task requires server-side reconciliation before its result can be confirmed."
			result.NextAction = "Keep this task ID and ask the user before continuing the same task. Do not submit a replacement task."
			writeReceipt(output, result)
			return errors.New("task reconciliation required")
		}
		if terminalSuccess(state) {
			return deliverTask(output, svc, request, result)
		}
		if terminalFailure(state) {
			result.OK = false
			result.FailurePhase = "status"
			result.ErrorMessage = "The API reported that this task failed."
			result.NextAction = "Review the task error before creating a new request."
			writeReceipt(output, result)
			return errors.New("task failed")
		}
	}
	result.OK = false
	result.FailurePhase = "status"
	result.ErrorMessage = "The task is still pending after the automatic wait window."
	result.NextAction = "Ask the user whether to continue polling this same task; do not submit a replacement task."
	writeReceipt(output, result)
	return errors.New("polling deadline")
}

type pollPolicy struct {
	maxReads int
	deadline int
	delays   []int
}

func pollingPolicy(kind string, override *pollRequest) pollPolicy {
	policy := pollPolicy{maxReads: 6, deadline: 120, delays: []int{3, 6, 12, 24, 30}}
	if kind == "video" {
		policy = pollPolicy{maxReads: 7, deadline: 300, delays: []int{5, 10, 20, 40, 60, 60}}
	}
	if override != nil {
		if override.MaxStatusReads > 0 && override.MaxStatusReads <= policy.maxReads {
			policy.maxReads = override.MaxStatusReads
		}
		if override.DeadlineSecs > 0 && override.DeadlineSecs <= policy.deadline {
			policy.deadline = override.DeadlineSecs
		}
	}
	return policy
}

func deliverTask(output io.Writer, svc service, request taskRequest, result receipt) error {
	count := request.RequestedCount
	if count == 0 {
		count = 1
	}
	if request.Kind == "video" {
		count = 1
	}
	for index := 0; index < count; index++ {
		path := contentPath(request.Kind, result.TaskID, index)
		destination, status, retryAfter, apiCode, apiMessage, err := svc.download(context.Background(), path, request.Kind, request.OutputDir)
		if err != nil {
			writeReceipt(output, apiFailure("content", status, retryAfter, apiCode, apiMessage, "The task completed but its media bytes could not be delivered. Do not resubmit the task."))
			return err
		}
		result.DeliveredPaths = append(result.DeliveredPaths, destination)
	}
	result.DeliveredCount = len(result.DeliveredPaths)
	result.Status = "completed"
	writeReceipt(output, result)
	return nil
}

func statusPath(kind, id string) string {
	return "/v1/" + kind + "s/" + url.PathEscape(id)
}

func contentPath(kind, id string, index int) string {
	path := "/v1/" + kind + "s/" + url.PathEscape(id) + "/content"
	if kind == "image" {
		return path + "?index=" + strconv.Itoa(index)
	}
	return path
}

func taskIdentity(body []byte) (string, string, bool, bool) {
	var response map[string]any
	if json.Unmarshal(body, &response) != nil {
		return "", "", false, false
	}
	id, _ := response["task_id"].(string)
	if id == "" {
		id, _ = response["id"].(string)
	}
	state, _ := response["status"].(string)
	if id == "" {
		return "", "", false, false
	}
	reconciliationRequired, _ := response["reconciliation_required"].(bool)
	return id, strings.ToLower(state), reconciliationRequired, true
}

func terminalSuccess(state string) bool {
	return state == "completed" || state == "succeeded" || state == "success"
}

func terminalFailure(state string) bool {
	return state == "failed" || state == "cancelled" || state == "canceled" || state == "expired" || state == "error"
}

func (svc service) request(ctx context.Context, method, path string, body io.Reader, contentType string) ([]byte, int, int, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, svc.baseURL+path, body)
	if err != nil {
		return nil, 0, 0, "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+svc.token)
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response, err := svc.client.Do(req)
	if err != nil {
		return nil, 0, 0, "", "", err
	}
	defer response.Body.Close()
	bodyBytes, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if readErr != nil || len(bodyBytes) > maxResponseBytes {
		return nil, response.StatusCode, retryAfter(response), "", "", errors.New("response unreadable")
	}
	code, message := publicAPIError(bodyBytes)
	return bodyBytes, response.StatusCode, retryAfter(response), code, message, nil
}

func (svc service) download(ctx context.Context, path, kind, outputDir string) (string, int, int, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.baseURL+path, nil)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+svc.token)
	response, err := svc.client.Do(req)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
		code, message := publicAPIError(body)
		return "", response.StatusCode, retryAfter(response), code, message, errors.New("content request rejected")
	}
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if (kind == "image" && !strings.HasPrefix(contentType, "image/")) || (kind == "video" && !strings.HasPrefix(contentType, "video/")) || contentType == "image/svg+xml" {
		return "", response.StatusCode, retryAfter(response), "", "", errors.New("invalid media content type")
	}
	limit := maxImageBytes
	if kind == "video" {
		limit = maxVideoBytes
	}
	if response.ContentLength > limit {
		return "", response.StatusCode, retryAfter(response), "", "", errors.New("media delivery exceeds limit")
	}
	directory := outputDir
	if directory == "" {
		directory, err = os.MkdirTemp("", "puretokens-media-")
		if err != nil {
			return "", response.StatusCode, retryAfter(response), "", "", err
		}
	}
	extension := extensionFor(contentType, kind)
	file, err := os.CreateTemp(directory, "puretokens-*."+extension)
	if err != nil {
		return "", response.StatusCode, retryAfter(response), "", "", err
	}
	name := file.Name()
	writer := io.LimitReader(response.Body, limit+1)
	written, copyErr := io.Copy(file, writer)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written > limit {
		_ = os.Remove(name)
		return "", response.StatusCode, retryAfter(response), "", "", errors.New("media delivery exceeded limit")
	}
	return name, response.StatusCode, retryAfter(response), "", "", nil
}

func extensionFor(contentType, kind string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	default:
		return map[string]string{"image": "img", "video": "video"}[kind]
	}
}

func retryAfter(response *http.Response) int {
	value, err := strconv.Atoi(strings.TrimSpace(response.Header.Get("Retry-After")))
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func publicAPIError(body []byte) (string, string) {
	var document map[string]any
	if json.Unmarshal(body, &document) != nil {
		return "", ""
	}
	if errorValue, ok := document["error"].(map[string]any); ok {
		code, _ := errorValue["code"].(string)
		message, _ := errorValue["message"].(string)
		return safePublicString(code), safePublicString(message)
	}
	code, _ := document["code"].(string)
	message, _ := document["message"].(string)
	return safePublicString(code), safePublicString(message)
}

func safePublicString(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 || strings.Contains(value, "http://") || strings.Contains(value, "https://") || strings.Contains(value, "Bearer ") {
		return ""
	}
	return value
}

func apiFailure(phase string, status, retryAfterSecs int, code, message, next string) receipt {
	result := receipt{OK: false, FailurePhase: phase, HTTPStatus: status, APIErrorCode: code, ErrorMessage: message, NextAction: next, RetryAfterSecs: retryAfterSecs}
	if result.ErrorMessage == "" {
		result.ErrorMessage = "The Pure Tokens API request did not complete."
	}
	return result
}

func validationFailure(message string) receipt {
	return receipt{OK: false, FailurePhase: "validation", ErrorMessage: message, NextAction: "Correct the local request before trying again."}
}

func writeReceipt(output io.Writer, result receipt) {
	writeJSON(output, result)
}

func writeJSON(output io.Writer, value any) {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(true)
	_ = encoder.Encode(value)
}

func clearString(value *string) {
	// Go strings cannot be reliably zeroed. Dropping the reference immediately
	// after a request still prevents any intentional persistence or logging.
	*value = ""
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
