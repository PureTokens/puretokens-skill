package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAttachmentChangesFailBeforeOrDuringTransmission(t *testing.T) {
	for _, change := range []string{"grow", "shrink", "replace", "same-size"} {
		t.Run(change, func(t *testing.T) {
			path := writeFixture(t, "reference.png", "original-data")
			req := taskRequest{Kind: "image", Operation: "edit", Model: "gpt-image-2", Prompt: "fixture", Attachments: []attachment{{Field: "image", Path: path}}}
			if err := validateTaskRequest(req); err != nil {
				t.Fatal(err)
			}
			mutate := func() {
				switch change {
				case "grow":
					os.Truncate(path, attachmentLimit("image")+100)
				case "shrink":
					os.Truncate(path, 1)
				case "replace":
					os.Remove(path)
					os.WriteFile(path, []byte("different"), 0600)
				case "same-size":
					os.WriteFile(path, []byte("modified-data"), 0600)
				}
			}
			// Hold and hash verified handles, then change input before streaming.
			files, err := openAttachments(req)
			if err != nil {
				t.Fatal(err)
			}
			defer closeAttachments(files)
			mutate()
			err = files[0].copyTo(io.Discard)
			// Replacement after opening retains the verified original handle and bytes.
			if change == "replace" {
				if err != nil {
					t.Fatal("original descriptor must remain valid", err)
				}
			} else if err == nil {
				t.Fatal("changed stream accepted")
			}
			if _, _, body, err := taskRequestBody(req); err == nil {
				body.(io.Closer).Close()
				t.Fatal("changed path accepted after validation")
			}
		})
	}
}
func TestUnprovenAndModifiedValidMediaCannotBeReused(t *testing.T) {
	root := t.TempDir()
	route := contentPath("image", "existing-task", 0)
	file := filepath.Join(root, fmt.Sprintf("puretokens-%x.png", sha256.Sum256([]byte(route))))
	original := fixturePNG(t)
	os.WriteFile(file, original, 0600)
	svc := service{baseURL: "invalid", downloadProofs: make(map[string]downloadProof)}
	if _, _, _, _, _, err := svc.download(context.Background(), route, "image", root); err == nil {
		t.Fatal("unproven file reused")
	}
	proof, err := fingerprintDownload(file, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	svc.downloadProofs[file] = proof
	if _, _, _, _, _, err := svc.download(context.Background(), route, "image", root); err != nil {
		t.Fatal("proven file rejected", err)
	}
	// Preserve a structurally valid PNG while changing its metadata bytes.
	changed := bytes.Clone(original)
	changed[24] ^= 1
	os.WriteFile(file, changed, 0600)
	if !validMediaFile(file, "image/png") {
		t.Fatal("fixture must pass existing structural validation")
	}
	if _, _, _, _, _, err := svc.download(context.Background(), route, "image", root); err == nil {
		t.Fatal("changed valid media reused")
	}
	actual, _ := os.ReadFile(file)
	if !bytes.Equal(actual, changed) {
		t.Fatal("user file overwritten")
	}
}

func TestAttachmentDigestCatchesSameSizeAndTimestampReplacement(t *testing.T) {
	path := writeFixture(t, "reference.png", "original-data")
	req := taskRequest{Kind: "image", Operation: "edit", Model: "gpt-image-2", Prompt: "fixture", Attachments: []attachment{{Field: "image", Path: path}}}
	if err := validateTaskRequest(req); err != nil {
		t.Fatal(err)
	}
	info := req.Attachments[0].snapshot
	os.WriteFile(path, []byte("modified-data"), 0600)
	os.Chtimes(path, info.ModTime(), info.ModTime())
	if _, _, body, err := taskRequestBody(req); err == nil {
		body.(io.Closer).Close()
		t.Fatal("same-size and same-timestamp content change escaped validation")
	}
}
func TestInvalidFreeTextNeverEntersRecordOrReceipt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task.json")
	req := taskRequest{Kind: "image", Operation: "generate", Model: "gpt-image-2", Prompt: "private prompt", Parameters: map[string]any{"size": "private draft about project alpha"}}
	var out bytes.Buffer
	if err := executeRecordedTask("submit", path, req, 0, "", &out, profileService()); err == nil {
		t.Fatal("invalid size accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("invalid request created a persistent record")
	}
	if bytes.Contains(out.Bytes(), []byte("private draft")) {
		t.Fatal("invalid parameter leaked in receipt")
	}
	if r := recordFromRequest(req); r.Parameters["size"] != nil {
		t.Fatal("unsafe free text projected from imported metadata")
	}
}

func TestLegacyDownloadRecordCanResumeButCannotAcknowledgeUnprovenFile(t *testing.T) {
	root := t.TempDir()
	id := "legacy-task"
	route := contentPath("image", id, 0)
	file := filepath.Join(root, fmt.Sprintf("puretokens-%x.png", sha256.Sum256([]byte(route))))
	os.WriteFile(file, fixturePNG(t), 0600)
	path := filepath.Join(root, "task.json")
	record := taskRecord{Format: taskRecordFormat, Kind: "image", TaskID: id, Status: "completed", RequestedCount: 1, Downloaded: map[int]string{0: file}}
	if err := saveTaskRecord(path, record, true); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := executeRecordedTask("resume", path, taskRequest{}, 0, "", &out, service{}); err != nil {
		t.Fatal("legacy task cannot resume", err)
	}
	result := decodeReceipt(t, &out)
	if len(result.DownloadedPaths) != 0 {
		t.Fatal("unproven legacy file offered for handoff")
	}
	out.Reset()
	if err := executeRecordedTask("delivered", path, taskRequest{}, 0, "", &out, service{}); err == nil {
		t.Fatal("unproven legacy file marked delivered")
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal("legacy user file removed")
	}
}
