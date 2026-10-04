package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

const operationsReceiptURL = "https://console.puretokensx.com/api/product/analytics/client-receipts"

// An explicit process preference, not a host-config read or a new credential
// path. Default off. No persistent ID, queue, retry or account attribution.
func operationsReceiptsEnabled() bool {
	return os.Getenv("PTP_OPERATIONS_RECEIPTS") == "1"
}

func operationsEvent(command string, attempted bool, document map[string]json.RawMessage) string {
	var ok, verified bool
	_ = json.Unmarshal(document["ok"], &ok)
	_ = json.Unmarshal(document["credential_verified"], &verified)
	if command == "init" && ok && verified {
		return "connection_verified"
	}
	if !attempted || (command != "submit" && command != "evaluate" && command != "audio") {
		return ""
	}
	if ok {
		return "execution_succeeded"
	}
	return "execution_failed"
}

func sendOperationsReceipt(client *http.Client, target, event string) {
	if event == "" {
		return
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	platform := runtime.GOOS
	if platform == "darwin" {
		platform = "macos"
	}
	if platform != "macos" && platform != "windows" && platform != "linux" {
		platform = "other"
	}
	body, _ := json.Marshal(map[string]any{
		"schemaVersion": 1,
		"id":            fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]),
		"source":        "skill", "event": event, "platform": platform,
		"occurredAt": time.Now().UTC().Format(time.RFC3339),
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err == nil {
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 2048))
	}
}

func dispatchOperationsReceipt(event string) {
	if !operationsReceiptsEnabled() {
		return
	}
	sendOperationsReceipt(&http.Client{Timeout: time.Second, CheckRedirect: rejectRedirect}, operationsReceiptURL, event)
}
