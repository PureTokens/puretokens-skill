// Pure Tokens API executor is the small, self-contained process used by the
// installable Skills. It intentionally has no dependency on Node, Python,
// MCP, a proxy, or a background service.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
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
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	apiOrigin              = "https://api.puretokensx.com"
	maxConfigBytes         = 1 << 20
	maxResponseBytes       = 2 << 20
	maxImageBytes    int64 = 25 << 20
	maxVideoBytes    int64 = 512 << 20
)

var executorVersion = "0.17.0"

type attachment struct {
	Field string `json:"field"`
	Path  string `json:"path"`
}

type taskRequest struct {
	MediaOperation         string `json:"media_operation,omitempty"`
	Index                  int    `json:"index,omitempty"`
	TaskStatus             string `json:"task_status,omitempty"`
	OriginalOperation      string `json:"original_operation,omitempty"`
	RetryNotBefore         string `json:"retry_not_before,omitempty"`
	ReconciliationRequired bool   `json:"reconciliation_required,omitempty"`
	route                  string

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
	RequestedCount    int            `json:"requested_count,omitempty"`
	Parameters        map[string]any `json:"parameters,omitempty"`
	SubmissionOutcome string         `json:"submission_outcome,omitempty"`
	DownloadedPaths   []string       `json:"downloaded_paths,omitempty"`
	DownloadedIndexes []int          `json:"downloaded_indexes,omitempty"`
	DeliveryStatus    string         `json:"delivery_status,omitempty"`
	LocalErrorCode    string         `json:"local_error_code,omitempty"`
	OriginalOperation string         `json:"original_operation,omitempty"`
	RetryNotBefore    string         `json:"retry_not_before,omitempty"`
	RecordPath        string         `json:"record_path,omitempty"`
	DeliveredIndexes  []int          `json:"delivered_indexes,omitempty"`

	OK                     bool   `json:"ok"`
	Kind                   string `json:"kind,omitempty"`
	Operation              string `json:"operation,omitempty"`
	Model                  string `json:"model,omitempty"`
	TaskID                 string `json:"task_id,omitempty"`
	Status                 string `json:"status,omitempty"`
	ReconciliationRequired bool   `json:"reconciliation_required,omitempty"`
	FailurePhase           string `json:"failure_phase,omitempty"`
	HTTPStatus             int    `json:"http_status,omitempty"`
	APIErrorCode           string `json:"api_error_code,omitempty"`
	ErrorMessage           string `json:"error_message,omitempty"`
	NextAction             string `json:"next_action,omitempty"`
	RetryAfterSecs         int    `json:"retry_after_seconds,omitempty"`
}

type initReceipt struct {
	OK                   bool     `json:"ok"`
	Command              string   `json:"command"`
	ExecutorVersion      string   `json:"executor_version"`
	ConfigurationStatus  string   `json:"configuration_status"`
	APIRequestExecuted   bool     `json:"api_request_executed"`
	CredentialVerified   bool     `json:"credential_verified"`
	APIIdentityConfirmed bool     `json:"api_identity_confirmed"`
	HTTPStatus           int      `json:"http_status,omitempty"`
	Message              string   `json:"message"`
	NextAction           string   `json:"next_action"`
	UsageExamples        []string `json:"usage_examples"`
}

type service struct {
	profilesRoot string

	baseURL         string
	client          *http.Client
	token           string
	downloadTimeout time.Duration
	now             func() time.Time
	wait            func(context.Context, time.Duration) bool
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
		writeReceipt(output, validationFailure("Choose one of: init, connection, balance, models, submit, status, wait, content."))
		return errors.New("missing command")
	}

	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	host := flags.String("host", "", "")
	requestFile := flags.String("request", "", "")
	recordFile := flags.String("record", "", "")
	index := flags.Int("index", 0, "")
	outputDir := flags.String("output-dir", "", "")
	if err := flags.Parse(args[1:]); err != nil || (*host == "" && command != "delivered") || len(flags.Args()) != 0 {
		writeReceipt(output, validationFailure("Identify the current host and use supported command options before starting the request."))
		return errors.New("invalid command")
	}
	isTask := command == "submit" || command == "task" || command == "status" || command == "wait" || command == "content" || command == "resume" || command == "delivered" || command == "preflight"
	if !isTask && command != "init" && command != "doctor" && command != "connection" && command != "balance" && command != "models" {
		writeReceipt(output, validationFailure("Choose init, doctor, connection, balance, models, preflight, submit, status, wait, content, resume or delivered."))
		return errors.New("unknown command")
	}
	if (*recordFile != "" && (!isTask || command == "preflight")) || ((*index != 0 || *outputDir != "") && (*recordFile == "" || (command != "content" && command != "delivered"))) || ((command == "resume" || command == "delivered") && *recordFile == "") || (*recordFile != "" && *requestFile != "" && command != "submit" && command != "task") || (!isTask && command != "models" && *requestFile != "") {
		writeReceipt(output, validationFailure("Use --request for media JSON, --record for same-task recovery, and --index/--output-dir only for recorded content. Never submit into an existing record."))
		return errors.New("incompatible command options")
	}
	if *requestFile != "" {
		data, err := readBoundedFile(*requestFile, maxResponseBytes)
		if err != nil || !json.Valid(data) {
			writeReceipt(output, validationFailure("The request file must contain one readable JSON object; no API request was sent."))
			return errors.New("invalid request file")
		}
		input = bytes.NewReader(data)
	}
	var request taskRequest
	if isTask {
		if *recordFile != "" && command != "submit" && command != "task" {
			record, err := loadTaskRecord(*recordFile)
			if err != nil {
				writeReceipt(output, validationFailure("The task record could not be read. Recover the original receipt; do not submit again."))
				return err
			}
			request = record.request()
		} else {
			var err error
			request, err = decodeTaskRequest(input)
			if err != nil {
				writeReceipt(output, validationFailure("A complete media request JSON is required; no API request was sent."))
				return err
			}
		}
		if command == "status" || command == "wait" || command == "content" || command == "resume" || command == "delivered" {
			request.Operation = "continue"
		}
	}
	svc := service{baseURL: apiOrigin, client: &http.Client{Timeout: 90 * time.Second, CheckRedirect: rejectRedirect}, profilesRoot: installedSkillsRoot()}
	if command == "delivered" {
		return executeRecordedTask(command, *recordFile, request, *index, *outputDir, output, svc)
	}
	token, err := credentialForHost(*host)
	if err != nil {
		if command == "init" {
			writeJSON(output, initCredentialFailure(err))
			return err
		}
		if command == "doctor" {
			return executeDoctorCredentialFailure(output, svc, *host, err)
		}
		status, message, next := credentialFailureDetails(err)
		result := receipt{}
		if isTask {
			if !validTaskID(request.TaskID) {
				request.TaskID = ""
			}
			result = taskReceipt(request, request.TaskID, request.TaskStatus)
		}
		writeReceipt(output, mergeFailure(result, receipt{FailurePhase: "validation", LocalErrorCode: status, ErrorMessage: message, NextAction: next}))
		return err
	}
	defer clearString(&token)
	svc.token = token
	if *recordFile != "" {
		return executeRecordedTask(command, *recordFile, request, *index, *outputDir, output, svc)
	}
	if isTask {
		data, _ := json.Marshal(request)
		input = bytes.NewReader(data)
	}
	switch command {
	case "init":
		return executeInit(output, svc)
	case "doctor":
		return executeDoctor(output, svc, *host)
	case "connection":
		return executeReadOnly(output, svc, "/v1", "connection")
	case "balance":
		return executeBalance(output, svc)
	case "models":
		if *requestFile == "" {
			input = nil
		}
		return executeModelQuery(output, svc, input)
	case "preflight":
		return executePreflight(output, svc, request)
	case "task", "submit":
		return executeTask(input, output, svc)
	case "status", "wait", "content":
		return executeExistingTask(command, input, output, svc)
	default:
		writeReceipt(output, validationFailure("This command is unavailable."))
		return errors.New("unavailable command")
	}
}

func executeInit(output io.Writer, svc service) error {
	result := initReceipt{Command: "init", ExecutorVersion: executorVersion, ConfigurationStatus: "unverified", UsageExamples: usageExamples()}
	body, status, _, _, _, err := svc.request(context.Background(), http.MethodGet, "/v1", nil, "")
	result.APIRequestExecuted = true
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
		result.APIIdentityConfirmed = true
		authBody, authStatus, _, _, _, authErr := svc.request(context.Background(), http.MethodGet, "/v1/media/models", nil, "")
		authObject, authDecodeErr := readAPIObject(authBody)
		_, catalogOK := authObject["data"].([]any)
		if authErr != nil || authStatus != http.StatusOK || authDecodeErr != nil || !catalogOK {
			result.ConfigurationStatus = "credential_unverified"
			result.HTTPStatus = authStatus
			result.Message = "The public API is reachable, but credential authentication was not confirmed."
			result.NextAction = "Check the selected Pure Tokens connection and its key permissions in the host, then run init again."
			writeJSON(output, result)
			return errors.New("credential authentication unconfirmed")
		}
		result.OK = true
		result.CredentialVerified = true
		result.ConfigurationStatus = "verified"
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
		"Generate multiple variants: use seedream-5.0-pro to create 3 variations of the same image brief.",
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
	return bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), nil
}

func executeReadOnly(output io.Writer, svc service, path string, command string) error {
	body, status, retryAfter, apiCode, apiMessage, err := svc.request(context.Background(), http.MethodGet, path, nil, "")
	if err != nil || status < 200 || status >= 300 {
		writeReceipt(output, apiFailure("submission", status, retryAfter, apiCode, apiMessage, "The read-only API request did not complete. Verify the current Pure Tokens connection, then try again."))
		return errors.New("read-only request failed")
	}
	payload, decodeErr := readAPIObject(body)
	if decodeErr != nil {
		writeReceipt(output, receipt{OK: false, FailurePhase: "submission", HTTPStatus: status, ErrorMessage: "The API returned an unreadable response.", NextAction: "Try again in a new host session."})
		return errors.New("invalid json")
	}
	// The command result is retained as JSON, but no host configuration or
	// credential is added to it. The calling Skill selects its user-visible
	// subset from the API response.
	var visible any = payload
	if command == "models" {
		if _, ok := payload["data"].([]any); !ok {
			writeReceipt(output, validationFailure("The API did not return a readable model catalog."))
			return errors.New("invalid catalog")
		}
		visible = projectCatalog(payload)
	}
	if command == "connection" {
		original := jsonObject(payload)
		basePath := ""
		if original["base_url"] == "/v1" {
			basePath = "/v1"
		}
		visible = map[string]any{"status": safePublicString(jsonString(original["status"])), "name": safePublicString(jsonString(original["name"])), "base_url": basePath, "credential_verified": false}
	}
	writeJSON(output, map[string]any{"ok": true, "command": command, "result": visible})
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
		result := taskReceipt(request, request.TaskID, continuationStatus(request.TaskStatus))
		return pollAndDeliver(output, svc, request, result)
	}

	if err := prepareProfileRequest(&request, svc); err != nil {
		writeReceipt(output, validationFailure(err.Error()))
		return err
	}
	path, contentType, body, err := taskRequestBody(request)
	if err != nil {
		writeReceipt(output, validationFailure("The current attachment could not be prepared; no API request was sent."))
		return err
	}
	if closer, ok := body.(io.Closer); ok {
		defer closer.Close()
	}
	response, status, retryAfter, apiCode, apiMessage, err := svc.request(context.Background(), http.MethodPost, path, body, contentType)
	if err != nil || status < 200 || status >= 300 {
		result := mergeFailure(taskReceipt(request, "", ""), apiFailure("submission", status, retryAfter, apiCode, apiMessage, "Review the returned error before changing this request."))
		result.SubmissionOutcome = "rejected"
		result = withRetry(result, retryAfter, svc.clock())
		if err != nil || status >= 500 || (status >= 300 && status < 400) {
			result.SubmissionOutcome = "unknown"
			result.NextAction = "Submission may have started. Do not automatically repeat the POST or infer billing; ask the user before creating another task."
		}
		writeReceipt(output, result)
		return errors.New("submission failed")
	}

	taskID, state, reconciliationRequired, ok := taskIdentity(response)
	if !ok {
		result := mergeFailure(taskReceipt(request, "", ""), receipt{FailurePhase: "submission", HTTPStatus: status, ErrorMessage: "The API accepted a response but did not return a usable task ID. The submission outcome is unknown.", NextAction: "Do not resubmit automatically. Ask the user before creating a new billable request."})
		result.SubmissionOutcome = "unknown"
		writeReceipt(output, withRetry(result, retryAfter, svc.clock()))
		return errors.New("task id missing")
	}
	result := taskReceipt(request, taskID, state)
	result = withRetry(result, retryAfter, svc.clock())
	result.SubmissionOutcome = "accepted"
	result.ReconciliationRequired = reconciliationRequired
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
		result.APIErrorCode, result.ErrorMessage = publicAPIError(response)
		if result.ErrorMessage == "" {
			result.ErrorMessage = "The API returned a failed task state."
		}
		result.NextAction = "Review the task error before creating a new request."
		writeReceipt(output, result)
		return errors.New("initial failed state")
	}
	if !terminalSuccess(state) && state != "pending" && state != "queued" && state != "processing" && state != "running" && state != "in_progress" {
		result.Status = "unknown"
		result.OK = false
		result.FailurePhase = "status"
		result.ErrorMessage = "The API returned an unsupported initial task state."
		result.NextAction = "Keep this task ID and query the same task later; do not resubmit."
		writeReceipt(output, result)
		return errors.New("unknown initial state")
	}
	// Return the accepted task immediately: no polling or downloads hide its ID.
	if terminalSuccess(state) {
		result.NextAction = "Download the completed task with content, one index at a time."
	} else {
		result.NextAction = "Keep this task ID. Use wait or status for this same task; never submit it again."
	}
	writeReceipt(output, result)
	return nil
}

func decodeTaskRequest(input io.Reader) (taskRequest, error) {
	var request taskRequest
	decoder := json.NewDecoder(io.LimitReader(input, maxResponseBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return taskRequest{}, err
	}
	// Do not wait for EOF on a terminal pipe after the complete JSON object.
	// --request is preferred; reject additional content already buffered.
	buffered, _ := io.ReadAll(decoder.Buffered())
	if len(bytes.TrimSpace(buffered)) != 0 {
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
	if request.Operation != "continue" && request.OutputDir != "" {
		info, err := os.Stat(request.OutputDir)
		if err != nil || !info.IsDir() {
			return errors.New("invalid output directory")
		}
	}
	if request.Operation == "continue" {
		if !validTaskID(request.TaskID) || len(request.Attachments) != 0 {
			return errors.New("invalid continuation")
		}
		if request.RetryNotBefore != "" {
			if _, err := time.Parse(time.RFC3339, request.RetryNotBefore); err != nil {
				return errors.New("invalid retry time")
			}
		}
		return nil
	}
	if request.TaskID != "" || request.TaskStatus != "" || request.OriginalOperation != "" || request.RetryNotBefore != "" || request.ReconciliationRequired || request.Index != 0 {
		return errors.New("new submissions cannot contain existing-task metadata")
	}
	if strings.TrimSpace(request.Model) == "" || strings.TrimSpace(request.Prompt) == "" {
		return errors.New("missing model or prompt")
	}

	limit := attachmentLimit(request.Kind)
	var total int64
	for _, attachment := range request.Attachments {
		if attachment.Field == "" || !filepath.IsAbs(attachment.Path) {
			return errors.New("attachment must have a field and absolute path")
		}
		info, err := os.Stat(attachment.Path)
		if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > limit {
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
			if key == "model" || key == "prompt" || key == "async" {
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
		if key == "model" || key == "prompt" || key == "async" {
			return "", "", nil, errors.New("reserved parameter")
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return "", "", nil, err
		}
		if text, ok := value.(string); ok {
			fields[key] = text
		} else {
			fields[key] = string(encoded)
		}
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
	if request.route != "" {
		return request.route
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.deadline)*time.Second)
	defer cancel()
	if request.ReconciliationRequired || terminalFailure(result.Status) || terminalSuccess(result.Status) {
		return finishKnownTask(output, result)
	}
	for read := 0; read < policy.maxReads; read++ {
		delay := time.Duration(policy.delays[min(read, len(policy.delays)-1)]) * time.Second
		if result.RetryNotBefore != "" {
			delay = retryDelay(result.RetryNotBefore, svc.clock())
		}
		if !svc.waitFor(ctx, delay) {
			break
		}
		result.RetryNotBefore = ""
		result.RetryAfterSecs = 0
		body, status, retry, code, message, err := svc.request(ctx, http.MethodGet, statusPath(request.Kind, result.TaskID), nil, "")
		if err != nil || status < 200 || status >= 300 {
			result = withRetry(mergeFailure(result, apiFailure("status", status, retry, code, message, "Keep this task ID. Continue this same task when the API is available; do not resubmit.")), retry, svc.clock())
			if status == 429 && retry > 0 && read+1 < policy.maxReads {
				continue
			}
			writeReceipt(output, result)
			return errors.New("status read failed")
		}
		result = clearTaskFailure(result)
		result, err = applyTaskStatus(result, body, status)
		result = withRetry(result, retry, svc.clock())
		if err != nil {
			writeReceipt(output, result)
			return err
		}
		if terminalSuccess(result.Status) {
			result.NextAction = "Use content for this task, one index at a time, then attach each downloaded file to the conversation."
			writeReceipt(output, result)
			return nil
		}
	}
	if result.HTTPStatus == http.StatusTooManyRequests || retryDelay(result.RetryNotBefore, svc.clock()) > 0 {
		result.OK = false
		result.FailurePhase = "status"
		result.NextAction = "Wait until retry_not_before, then continue this same task. The automatic wait window ended; do not submit again."
		if result.ErrorMessage == "" {
			result.ErrorMessage = "The API requested a wait longer than the remaining automatic wait window."
		}
		writeReceipt(output, result)
		return errors.New("retry exceeds wait window")
	}
	result.OK = false
	result.FailurePhase = "status"
	result.ErrorMessage = "The task is still pending after the automatic wait window."
	result.NextAction = "Keep this task ID and ask whether to continue waiting for it. Do not submit another task."
	writeReceipt(output, result)
	return errors.New("polling deadline")
}

func waitContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	if deadline, ok := ctx.Deadline(); ok && time.Now().Add(delay).After(deadline) {
		return false
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
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
	// One file per invocation allows the host to hand it off before the next read.
	if request.OutputDir == "" || !filepath.IsAbs(request.OutputDir) {
		writeReceipt(output, mergeFailure(result, validationFailure("Choose an existing absolute output directory for the media file.")))
		return errors.New("output directory required")
	}
	if info, err := os.Stat(request.OutputDir); err != nil || !info.IsDir() {
		result = mergeFailure(result, validationFailure("The output directory is missing or unavailable. Choose an existing output directory, then retrieve this same task."))
		result.LocalErrorCode = "output_directory_unavailable"
		writeReceipt(output, result)
		return errors.New("output directory unavailable")
	}
	if request.Kind == "image" && request.RequestedCount == 0 {
		result = mergeFailure(result, validationFailure("The original image count is unknown. Recover requested_count from the original submission receipt or task record before downloading."))
		result.LocalErrorCode = "original_count_unknown"
		writeReceipt(output, result)
		return errors.New("unknown image count")
	}
	if request.Index < 0 || request.Index >= max(1, request.RequestedCount) || (request.Kind == "video" && request.Index != 0) {
		writeReceipt(output, mergeFailure(result, validationFailure("Choose a valid zero-based content index.")))
		return errors.New("invalid index")
	}
	destination, status, retry, code, message, err := svc.download(context.Background(), contentPath(request.Kind, result.TaskID, request.Index), request.Kind, request.OutputDir)
	if err != nil {
		failure := contentFailure(err, status, retry, code, message)
		writeReceipt(output, withRetry(mergeFailure(result, failure), retry, svc.clock()))
		return err
	}
	result.DownloadedPaths = []string{destination}
	result.DownloadedIndexes = []int{request.Index}
	result.DeliveryStatus = "downloaded_awaiting_host_delivery"
	result.Status = "completed"
	result = withRetry(result, retry, svc.clock())
	result.NextAction = "Attach this local file to the user. Download the next missing index only after handing off this file. Keep outputs only as requested by the user."
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
	if !validTaskID(id) {
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
	return svc.requestWithAuthentication(ctx, method, path, body, contentType, true)
}

func (svc service) requestWithAuthentication(ctx context.Context, method, path string, body io.Reader, contentType string, authenticated bool) ([]byte, int, int, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, svc.baseURL+path, body)
	if err != nil {
		return nil, 0, 0, "", "", err
	}
	if authenticated && svc.token != "" {
		req.Header.Set("Authorization", "Bearer "+svc.token)
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	response, err := fixedClient(svc.client).Do(req)
	if err != nil {
		return nil, 0, 0, "", "", err
	}
	defer response.Body.Close()
	bodyBytes, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if readErr != nil || len(bodyBytes) > maxResponseBytes {
		return nil, response.StatusCode, retryAfter(response), "", "", errors.New("response unreadable")
	}
	bodyBytes = sanitizeResponseJSON(bodyBytes, svc.token)
	if bodyBytes == nil && response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil, response.StatusCode, retryAfter(response), "", "", errors.New("response unreadable")
	}
	code, message := publicAPIError(bodyBytes)
	return bodyBytes, response.StatusCode, retryAfter(response), code, message, nil
}

func (svc service) download(ctx context.Context, path, kind, outputDir string) (string, int, int, string, string, error) {
	if !filepath.IsAbs(outputDir) {
		return "", 0, 0, "", "", errors.New("absolute output directory required")
	}
	outputID := fmt.Sprintf("puretokens-%x", sha256.Sum256([]byte(path)))
	for _, format := range []string{"image/png", "image/jpeg", "image/webp", "image/gif", "image/avif", "video/mp4", "video/webm"} {
		candidate := filepath.Join(outputDir, outputID+"."+extensionFor(format, kind))
		if info, err := os.Lstat(candidate); err == nil {
			if info.Mode().IsRegular() && strings.HasPrefix(format, kind+"/") && validMediaFile(candidate, format) {
				return candidate, 200, 0, "", "", nil
			}
			return "", 0, 0, "", "", errors.New("existing output is not a validated task result; choose another output directory")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.baseURL+path, nil)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+svc.token)
	client := fixedClient(svc.client)
	client.Timeout = svc.contentTimeout()
	response, err := client.Do(req)
	if err != nil {
		return "", 0, 0, "", "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
		code, message := publicAPIError(sanitizeResponseJSON(body, svc.token))
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
		return "", response.StatusCode, 0, "", "", errors.New("explicit output directory required")
	}
	reader := bufio.NewReader(response.Body)
	prefix, _ := reader.Peek(512)
	if !validMediaPrefix(prefix, contentType) {
		return "", response.StatusCode, 0, "", "", errors.New("invalid media bytes")
	}
	extension := extensionFor(contentType, kind)
	file, err := os.CreateTemp(directory, ".puretokens-part-*."+extension)
	if err != nil {
		return "", response.StatusCode, retryAfter(response), "", "", err
	}
	name := file.Name()
	defer os.Remove(name)
	writer := io.LimitReader(reader, limit+1)
	written, copyErr := io.Copy(file, writer)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		return "", response.StatusCode, retryAfter(response), "", "", errors.Join(copyErr, closeErr)
	}
	if written == 0 || written > limit || !validMediaFile(name, contentType) {
		_ = os.Remove(name)
		return "", response.StatusCode, retryAfter(response), "", "", errors.New("media delivery exceeded limit")
	}
	destination := filepath.Join(directory, outputID+"."+extension)
	if err := os.Link(name, destination); err != nil {
		if os.IsExist(err) || copyDownloadExclusive(name, destination) != nil {
			return "", response.StatusCode, 0, "", "", errors.New("cannot finalize media file without overwriting existing output")
		}
	}
	return destination, response.StatusCode, retryAfter(response), "", "", nil
}

func extensionFor(contentType, kind string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/avif":
		return "avif"
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
	if err != nil {
		if at, parseErr := http.ParseTime(response.Header.Get("Retry-After")); parseErr == nil {
			value = int(time.Until(at).Seconds()) + 1
		}
	}
	if value <= 0 {
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
		return safePublicCode(code), safePublicString(message)
	}
	code, _ := document["code"].(string)
	message, _ := document["message"].(string)
	return safePublicCode(code), safePublicString(message)
}

var publicCodePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,100}$`)
var credentialPattern = regexp.MustCompile(`(?i)(sk-[a-z0-9_-]+|bearer\s+\S+|(?:api[_ -]?key|token|secret)\s*[:=]\s*\S+)`)

func safePublicString(value string) string {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	if len(value) > 500 || strings.Contains(lower, "http:") || strings.Contains(lower, "https:") || credentialPattern.MatchString(value) || strings.Contains(lower, "stack trace") {
		return ""
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return ""
		}
	}
	// Raw hostnames and paths do not belong in public API error messages.
	if regexp.MustCompile(`(?i)\b[a-z0-9-]+\.(com|net|org|internal|local|ai)\b`).MatchString(value) {
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
	if writer, ok := output.(*recordReceiptWriter); ok {
		writer.writeReceipt(result)
		return
	}
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
