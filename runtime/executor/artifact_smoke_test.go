package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Runs the actual distributed native executable, not a freshly compiled test main.
func TestPackagedExecutorIdentityAndOfflinePreflight(t *testing.T) {
	platform := runtime.GOOS + "-" + runtime.GOARCH
	file := "bin/puretokens-api-" + platform
	if runtime.GOOS == "windows" {
		file += ".exe"
	}
	absolute, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		Version      string `json:"version"`
		SourceSHA256 string `json:"sourceSha256"`
		Artifacts    map[string]struct {
			SHA256 string `json:"sha256"`
		} `json:"artifacts"`
	}
	data, err := os.ReadFile("build-proof.json")
	if err != nil {
		t.Fatal("rebuild executor artifacts", err)
	}
	if err = json.Unmarshal(data, &proof); err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(absolute)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(binary)) != proof.Artifacts[platform].SHA256 {
		t.Fatal("packaged bytes differ from proof")
	}
	output, err := exec.Command(absolute, "--build-info").Output()
	if err != nil {
		t.Fatal("packaged executable failed", err)
	}
	var identity map[string]string
	if json.Unmarshal(output, &identity) != nil || identity["source_sha256"] != proof.SourceSHA256 || identity["version"] != proof.Version {
		t.Fatal("packaged identity differs")
	}
	root := t.TempDir()
	// Inline synthetic credential in an exact isolated Codex config. Preflight must
	// not touch the network; this deliberately unknown core model needs no catalog.
	desktopFixtureFile(t, root, "config.toml", "model_provider = 'fixture'\n[model_providers.fixture]\nbase_url = 'https://api.puretokensx.com/v1'\nexperimental_bearer_token = 'synthetic-artifact-token'\n")
	request := filepath.Join(root, "request.json")
	os.WriteFile(request, []byte(`{"kind":"image","operation":"generate","model":"artifact-fixture-model","prompt":"fixture"}`), 0600)
	t.Setenv("CODEX_HOME", root)
	command := exec.Command(absolute, "preflight", "--host", "codex", "--request", request)
	output, err = command.Output()
	if err != nil {
		t.Fatal("packaged preflight failed", err)
	}
	var result receipt
	if json.Unmarshal(output, &result) != nil || !result.OK || result.SubmissionOutcome != "not_submitted" || result.Operation != "preflight" {
		t.Fatal("packaged preflight receipt invalid")
	}
	if bytes.Contains(output, []byte("synthetic-artifact-token")) {
		t.Fatal("fixture credential leaked")
	}
}
