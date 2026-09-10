package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkBuddyUnrecognizedRecordPreservesWorkingConnection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	if err := os.WriteFile(path, []byte(`[{"url":"https://example.invalid/v1/chat/completions","apiKey":"fixture"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := credentialFromWorkBuddyFile(path)
	if err == nil {
		t.Fatal("unrecognized endpoint accepted")
	}
	_, _, next := credentialFailureDetails(err)
	if !strings.Contains(next, "Keep existing connections") {
		t.Fatalf("misleading guidance: %+v", next)
	}
}

func TestWorkBuddySavedConnectionIndependentOfChatSelection(t *testing.T) {
	for _, selected := range []string{"builtin", "saved"} {
		path := filepath.Join(t.TempDir(), "models.json")
		body := `{"selectedModel":"` + selected + `","models":[{"id":"builtin"},{"id":"saved","url":"https://api.puretokensx.com/v1/chat/completions","apiKey":"fixture-key"}]}`
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := credentialFromWorkBuddyFile(path)
		if err != nil || got != "fixture-key" {
			t.Fatal("saved connection depended on chat selection")
		}
	}
}
func TestWorkBuddyRecordDiagnostics(t *testing.T) {
	for _, tt := range []struct{ body, status string }{
		{`{}`, "workbuddy_record_format_unsupported"},
		{`{"models":[]}`, "workbuddy_connection_not_found"},
		{`{"models":`, "workbuddy_record_unreadable"},
	} {
		path := filepath.Join(t.TempDir(), "models.json")
		if err := os.WriteFile(path, []byte(tt.body), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := credentialFromWorkBuddyFile(path)
		status, _, _ := credentialFailureDetails(err)
		if status != tt.status {
			t.Fatalf("got %s want %s", status, tt.status)
		}
	}
	_, err := credentialFromWorkBuddyFile(filepath.Join(t.TempDir(), "missing.json"))
	status, _, _ := credentialFailureDetails(err)
	if status != "workbuddy_record_missing" {
		t.Fatal(status)
	}
}
