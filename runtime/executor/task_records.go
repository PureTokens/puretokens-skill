package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const taskRecordFormat = "puretokens-task-v1"

// An explicit user artifact, never a hidden task database. Keep only the
// metadata needed to continue or deliver an existing task.
type taskRecord struct {
	Format                 string         `json:"format"`
	Kind                   string         `json:"kind"`
	TaskID                 string         `json:"task_id,omitempty"`
	Model                  string         `json:"model,omitempty"`
	OriginalOperation      string         `json:"original_operation,omitempty"`
	RequestedCount         int            `json:"requested_count,omitempty"`
	Parameters             map[string]any `json:"parameters,omitempty"`
	Status                 string         `json:"status,omitempty"`
	SubmissionOutcome      string         `json:"submission_outcome,omitempty"`
	ReconciliationRequired bool           `json:"reconciliation_required,omitempty"`
	RetryNotBefore         string         `json:"retry_not_before,omitempty"`
	OutputDir              string         `json:"output_dir,omitempty"`
	Downloaded             map[int]string `json:"downloaded,omitempty"`
	Delivered              []int          `json:"delivered,omitempty"`
}

type recordReceiptWriter struct {
	output io.Writer
	path   string
	record taskRecord
	err    error
}

func (writer *recordReceiptWriter) Write(data []byte) (int, error) {
	return writer.output.Write(data)
}

func (writer *recordReceiptWriter) writeReceipt(result receipt) {
	record := writer.record
	if result.TaskID != "" {
		record.TaskID = result.TaskID
	}
	if result.Model != "" {
		record.Model = result.Model
	}
	if result.OriginalOperation != "" {
		record.OriginalOperation = result.OriginalOperation
	}
	if result.RequestedCount > 0 {
		record.RequestedCount = result.RequestedCount
	}
	if len(result.Parameters) > 0 {
		record.Parameters = result.Parameters
	}
	if result.Status != "" {
		record.Status = result.Status
	}
	if result.SubmissionOutcome != "" {
		record.SubmissionOutcome = result.SubmissionOutcome
	}
	if record.TaskID == "" && result.FailurePhase == "validation" {
		record.SubmissionOutcome = "not_submitted"
	}
	record.ReconciliationRequired = result.ReconciliationRequired
	record.RetryNotBefore = result.RetryNotBefore
	if record.Downloaded == nil {
		record.Downloaded = make(map[int]string)
	}
	for index, path := range result.DownloadedPaths {
		if index < len(result.DownloadedIndexes) {
			record.Downloaded[result.DownloadedIndexes[index]] = path
		}
	}
	result.RecordPath = writer.path
	result.DeliveredIndexes = record.Delivered
	if result.TaskID != "" && len(result.DownloadedIndexes) == 0 {
		// Resume receipts expose only pending handoffs. The host can attach an
		// already downloaded file before issuing another content request.
		for index := range record.Downloaded {
			delivered := false
			for _, done := range record.Delivered {
				if done == index {
					delivered = true
					break
				}
			}
			path := record.Downloaded[index]
			format := recordedDownloadFormat(record.Kind, record.TaskID, index, path)
			if !delivered && format != "" && validMediaFile(path, format) {
				result.DownloadedIndexes = append(result.DownloadedIndexes, index)
			}
		}
		sort.Ints(result.DownloadedIndexes)
		for _, index := range result.DownloadedIndexes {
			result.DownloadedPaths = append(result.DownloadedPaths, record.Downloaded[index])
		}
		if len(result.DownloadedIndexes) > 0 {
			result.DeliveryStatus = "downloaded_awaiting_host_delivery"
		}
	}
	if err := saveTaskRecord(writer.path, record, false); err != nil {
		writer.err = err
		result.OK = false
		result.LocalErrorCode = "task_record_write_failed"
		result.ErrorMessage = "The task receipt could not be saved to its explicit record."
		result.NextAction = "Keep this receipt and task ID. Repair the record location before continuing; do not repeat the submission."
	} else {
		writer.record = record
	}
	writeJSON(writer.output, result)
}

func (record taskRecord) request() taskRequest {
	return taskRequest{
		Kind: record.Kind, Operation: "continue", TaskID: record.TaskID,
		Model: record.Model, OriginalOperation: record.OriginalOperation,
		RequestedCount: record.RequestedCount, Parameters: record.Parameters,
		TaskStatus: record.Status, RetryNotBefore: record.RetryNotBefore,
		ReconciliationRequired: record.ReconciliationRequired, OutputDir: record.OutputDir,
	}
}

func recordFromRequest(request taskRequest) taskRecord {
	result := taskReceipt(request, request.TaskID, request.TaskStatus)
	return taskRecord{
		Format: taskRecordFormat, Kind: result.Kind, TaskID: result.TaskID,
		Model: result.Model, OriginalOperation: result.OriginalOperation,
		RequestedCount: result.RequestedCount, Parameters: result.Parameters,
		Status: result.Status, RetryNotBefore: result.RetryNotBefore,
		ReconciliationRequired: result.ReconciliationRequired, OutputDir: request.OutputDir,
	}
}

func loadTaskRecord(path string) (taskRecord, error) {
	var record taskRecord
	if !filepath.IsAbs(path) {
		return record, errors.New("record requires an absolute path")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return record, errors.New("record must be an existing regular file")
	}
	data, err := readBoundedFile(path, maxResponseBytes)
	if err != nil {
		return record, errors.New("record unreadable")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return record, errors.New("record invalid")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return record, errors.New("record contains additional content")
	}
	if record.Format != taskRecordFormat || (record.Kind != "image" && record.Kind != "video") || (record.TaskID != "" && !validTaskID(record.TaskID)) || record.RequestedCount < 0 || record.RequestedCount > 6 {
		return record, errors.New("record has unsupported metadata")
	}
	if record.TaskID != "" {
		if err := validateTaskRequest(record.request()); err != nil {
			return record, errors.New("record has invalid task metadata")
		}
	}
	// Re-project any copied record before persisting or displaying it.
	safe := taskReceipt(record.request(), record.TaskID, record.Status)
	record.Model, record.OriginalOperation, record.Parameters = safe.Model, safe.OriginalOperation, safe.Parameters
	for index, path := range record.Downloaded {
		if index < 0 || index >= max(1, record.RequestedCount) || recordedDownloadFormat(record.Kind, record.TaskID, index, path) == "" {
			return record, errors.New("record has invalid download metadata")
		}
	}
	seen := map[int]bool{}
	for _, index := range record.Delivered {
		if _, ok := record.Downloaded[index]; !ok || seen[index] {
			return record, errors.New("record has invalid delivery metadata")
		}
		seen[index] = true
	}
	return record, nil
}

func recordedDownloadFormat(kind, taskID string, index int, path string) string {
	if !validTaskID(taskID) || !filepath.IsAbs(path) {
		return ""
	}
	stem := fmt.Sprintf("puretokens-%x", sha256.Sum256([]byte(contentPath(kind, taskID, index))))
	for _, format := range []string{"image/png", "image/jpeg", "image/webp", "image/gif", "image/avif", "video/mp4", "video/webm"} {
		if strings.HasPrefix(format, kind+"/") && filepath.Base(path) == stem+"."+extensionFor(format, kind) {
			return format
		}
	}
	return ""
}

func saveTaskRecord(path string, record taskRecord, create bool) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if create {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(data)
		syncErr := file.Sync()
		closeErr := file.Close()
		return errors.Join(writeErr, syncErr, closeErr)
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("record changed or removed")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".puretokens-record-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func lockTaskRecord(path string) (func(), error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("record requires an absolute path")
	}
	lock := path + ".lock"
	file, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("record locked or location unwritable")
	}
	if _, err := fmt.Fprintln(file, os.Getpid()); err != nil {
		file.Close()
		os.Remove(lock)
		return nil, err
	}
	if err := file.Close(); err != nil {
		os.Remove(lock)
		return nil, err
	}
	return func() { _ = os.Remove(lock) }, nil
}

func executeRecordedTask(command, path string, request taskRequest, index int, outputDir string, output io.Writer, svc service) error {
	unlock, err := lockTaskRecord(path)
	if err != nil {
		result := mergeFailure(taskReceipt(request, request.TaskID, request.TaskStatus), validationFailure("The task record is busy or its directory is unavailable. Check whether another command is using it; after confirming no command is running, remove only an abandoned adjacent .lock file."))
		result.LocalErrorCode = "task_record_unavailable"
		writeReceipt(output, result)
		return err
	}
	defer unlock()
	var record taskRecord
	if command == "submit" || command == "task" {
		if request.Operation == "continue" {
			writeReceipt(output, validationFailure("Use resume with the existing record instead of submit."))
			return errors.New("record submission cannot continue")
		}
		if err := validateTaskRequest(request); err != nil {
			writeReceipt(output, validationFailure("Correct the media request before creating its task record."))
			return err
		}
		record = recordFromRequest(request)
		record.SubmissionOutcome = "unknown"
		if err := saveTaskRecord(path, record, true); err != nil {
			writeReceipt(output, validationFailure("Choose a new writable task record path. Existing files are never overwritten by submit; no POST was sent."))
			return err
		}
	} else {
		record, err = loadTaskRecord(path)
		if err != nil {
			writeReceipt(output, validationFailure("The selected task record is unreadable or unsupported. Recover the original receipt; do not submit again."))
			return err
		}
		request = record.request()
		request.Index = index
		if outputDir != "" {
			request.OutputDir = outputDir
		}
		record.OutputDir = request.OutputDir
		if !validTaskID(request.TaskID) {
			result := mergeFailure(taskReceipt(request, "", record.Status), validationFailure("This record has no confirmed task ID. Its submission cannot be safely resumed."))
			result.SubmissionOutcome = record.SubmissionOutcome
			result.NextAction = "Check the original receipt or service task history. Do not automatically submit a replacement."
			writeReceipt(output, result)
			return errors.New("task record has no id")
		}
	}
	writer := &recordReceiptWriter{output: output, path: path, record: record}
	if command == "delivered" {
		if _, ok := record.Downloaded[index]; !ok || index < 0 {
			writeReceipt(output, mergeFailure(taskReceipt(request, request.TaskID, record.Status), validationFailure("Only a downloaded index can be marked delivered, after the host confirms attachment handoff.")))
			return errors.New("index not downloaded")
		}
		found := false
		for _, done := range record.Delivered {
			if done == index {
				found = true
			}
		}
		if !found {
			writer.record.Delivered = append(writer.record.Delivered, index)
			sort.Ints(writer.record.Delivered)
		}
		result := taskReceipt(request, request.TaskID, record.Status)
		result.DeliveryStatus = "partially_delivered"
		if record.RequestedCount > 0 && len(writer.record.Delivered) == record.RequestedCount || record.Kind == "video" && len(writer.record.Delivered) == 1 {
			result.DeliveryStatus = "delivered"
		}
		result.NextAction = "Delivery acknowledgement saved. Continue only the remaining indexes, if any."
		writer.writeReceipt(result)
		return writer.err
	}
	data, _ := json.Marshal(request)
	if command == "submit" || command == "task" {
		err = executeTask(bytes.NewReader(data), writer, svc)
	} else {
		if command == "resume" {
			command = "wait"
		}
		err = executeExistingTask(command, bytes.NewReader(data), writer, svc)
	}
	return errors.Join(err, writer.err)
}
