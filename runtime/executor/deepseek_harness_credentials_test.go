package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func harnessFixture(t *testing.T) (string, string) {
	t.Helper()
	data, err := os.ReadFile("../../test/fixtures/switch-deepseek-harness-connection.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Patch       string
		Credentials string
	}
	if json.Unmarshal(data, &fixture) != nil {
		t.Fatal("invalid writer fixture")
	}
	return fixture.Patch, fixture.Credentials
}

func TestDeepSeekHarnessConnectionBoundary(t *testing.T) {
	for _, variant := range []string{"writer", "renamed-provider", "foreign-endpoint", "missing-model", "duplicate-row", "duplicate-field", "alias", "extra-document", "disabled", "home-patch", "plugin-patch", "environment-override", "model-route", "provider-header", "second-key", "missing-ref", "old-store"} {
		t.Run(variant, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "profiles", "desktop"), 0700); err != nil {
				t.Fatal(err)
			}
			patch, creds := harnessFixture(t)
			getenv := func(string) string { return "" }
			switch variant {
			case "renamed-provider":
				patch = strings.ReplaceAll(patch, "puretokens", "renamed") // restore the fixed service origin
				patch = strings.ReplaceAll(patch, "api.renamedx.com", "api.puretokensx.com")
			case "foreign-endpoint":
				patch = strings.ReplaceAll(patch, "api.puretokensx.com", "example.invalid")
			case "missing-model":
				patch = strings.ReplaceAll(patch, "model: deepseek-v3", "model: missing")
			case "duplicate-row":
				patch += patch
			case "duplicate-field":
				patch = strings.Replace(patch, "api: openai-completions", "api: openai-completions\n        api: other", 1)
			case "alias":
				patch = strings.Replace(patch, "config:", "config: &config", 1)
			case "extra-document":
				patch += "---\n[]\n"
			case "disabled":
				patch = strings.Replace(patch, "- id: llm-pi-ai", "- id: llm-pi-ai\n  disabled: true", 1)
			case "home-patch":
				clientFixtureFile(t, root, "cordis.patch.yml", "- id: agent-default-model\n  config: {provider: another, model: other}\n")
			case "plugin-patch":
				patch += "- id: credentials-local\n  config: {path: other}\n"
			case "environment-override":
				getenv = func(string) string { return "synthetic-override" }
			case "model-route":
				patch = strings.Replace(patch, "- id: deepseek-v3", "- id: deepseek-v3\n            baseURL: https://example.invalid/v1", 1)
			case "provider-header":
				patch = strings.Replace(patch, "api: openai-completions", "headers: {}\n        api: openai-completions", 1)
			case "second-key":
				patch = strings.Replace(patch, "providers:", "providers:\n      other:\n        baseURL: https://api.puretokensx.com/v1\n        apiKeyEnv: SECOND_KEY", 1)
			case "missing-ref":
				creds = "version: 1\nrefs: {}\n"
			case "old-store":
				creds = strings.Replace(creds, "version: 1", "version: 0", 1)
			}
			file := "profiles/desktop/cordis.patch.yml"
			clientFixtureFile(t, root, file, patch)
			clientFixtureFile(t, root, ".credentials.yaml", creds)
			key, err := credentialFromDeepSeekHarnessRoot(root, getenv)
			ok := variant == "writer" || variant == "renamed-provider"
			if (err == nil) != ok || (ok && key != "pts-fixture-token") || (!ok && key != "") {
				t.Fatal("incorrect connection resolution")
			}
			after, _ := os.ReadFile(filepath.Join(root, file))
			if string(after) != patch {
				t.Fatal("adapter modified client configuration")
			}
			if err != nil {
				code, message, next := credentialFailureDetails(err)
				if strings.Contains(code+message+next, "pts-fixture-token") {
					t.Fatal("credential leaked")
				}
			}
		})
	}
}

func TestDeepSeekHarnessRootAndDispatcher(t *testing.T) {
	home := t.TempDir()
	for _, value := range []string{"relative", home + "/../other", home + "/./other", home + "\nother"} {
		if _, err := deepSeekHarnessRoot(home, func(string) string { return value }); err == nil {
			t.Fatal("unsupported override accepted")
		}
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("DSH_HOME", "")
	t.Setenv("PURETOKENS_SWITCH_DEEPSEEK_HARNESS_API_KEY", "")
	root := filepath.Join(home, ".dsh")
	locations := doctorHostLocations("deepseek-harness", home, func(name string) string {
		if name == "DSH_AGENTS_HOME" {
			return filepath.Join(home, "shared")
		}
		return ""
	})
	if len(locations) != 2 || locations[0].path != filepath.Join(root, "skills") || locations[1].path != filepath.Join(home, "shared", "skills") {
		t.Fatal("Harness discovery roots differ")
	}
	if err := os.MkdirAll(filepath.Join(root, "profiles", "desktop"), 0700); err != nil {
		t.Fatal(err)
	}
	patch, creds := harnessFixture(t)
	clientFixtureFile(t, root, "profiles/desktop/cordis.patch.yml", patch)
	clientFixtureFile(t, root, ".credentials.yaml", creds)
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	for _, status := range []int{200, 401, 403} {
		calls := 0
		http.DefaultTransport = evaluationTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.Method != "POST" || r.URL.String() != apiOrigin+"/v1/images/generations" || r.Header.Get("Authorization") != "Bearer pts-fixture-token" {
				t.Error("wrong dedicated request")
			}
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) != nil || body["model"] != "gpt-image-2.5-flare" {
				t.Error("exact model lost")
			}
			if status == 200 {
				return evaluationHTTP(status, `{"id":"image_fixture","status":"pending"}`), nil
			}
			return evaluationHTTP(status, `{"error":{"code":"permission_denied"}}`), nil
		})
		var out bytes.Buffer
		err := run([]string{"submit", "--host", "deepseek-harness"}, strings.NewReader(`{"kind":"image","operation":"generate","model":"gpt-image-2.5-flare","prompt":"synthetic"}`), &out)
		if calls != 1 || (err == nil) != (status == 200) || strings.Contains(out.String(), "pts-fixture-token") {
			t.Fatal("dispatcher retried, leaked or misclassified request")
		}
	}
	t.Setenv("PURETOKENS_SWITCH_DEEPSEEK_HARNESS_API_KEY", "synthetic-override")
	calls := 0
	http.DefaultTransport = evaluationTransport(func(*http.Request) (*http.Response, error) { calls++; return evaluationHTTP(500, `{}`), nil })
	var out bytes.Buffer
	if run([]string{"models", "--host", "deepseek-harness"}, strings.NewReader(""), &out) == nil || calls != 0 {
		t.Fatal("override did not stop before network")
	}
}
