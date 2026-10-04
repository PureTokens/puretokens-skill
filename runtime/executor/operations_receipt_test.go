package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOperationsReceiptPrivacyAndFailureIsolation(t *testing.T) {
	t.Setenv("PTP_OPERATIONS_RECEIPTS", "")
	if operationsReceiptsEnabled() {
		t.Fatal("receipts must default off")
	}
	t.Setenv("PTP_OPERATIONS_RECEIPTS", "1")
	if !operationsReceiptsEnabled() {
		t.Fatal("explicit process opt-in not recognized")
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("receipt included authentication")
		}
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 6 {
			t.Error("receipt did not contain exactly six fixed fields")
		}
		if body["source"] != "skill" || body["event"] != "installed" {
			t.Error("receipt contains unexpected facts")
		}
		w.WriteHeader(503)
	}))
	defer server.Close()
	sendOperationsReceipt(server.Client(), server.URL, "installed")
	if calls != 1 {
		t.Fatal("failed receipts must not retry")
	}
}

func TestOperationsEventsDistinguishConnectionSubmissionAndLocalHelp(t *testing.T) {
	document := map[string]json.RawMessage{"ok": json.RawMessage("true"), "credential_verified": json.RawMessage("true")}
	for _, command := range []string{"doctor", "preflight", "models", "audio-verify", "status", "wait", "content", "delivered"} {
		if operationsEvent(command, true, document) != "" {
			t.Fatalf("%s must not be counted as a new execution", command)
		}
	}
	if operationsEvent("init", true, document) != "connection_verified" ||
		operationsEvent("submit", true, document) != "execution_succeeded" ||
		operationsEvent("submit", false, document) != "" {
		t.Fatal("completion evidence mixed up")
	}
}
