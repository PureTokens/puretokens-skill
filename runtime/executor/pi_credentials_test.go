package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPiCredentialReferencesStopBeforeRequest(t *testing.T) {
	for _, value := range []string{"$SYNTHETIC_PI_KEY", "${SYNTHETIC_PI_KEY}", "prefix-$SYNTHETIC_PI_KEY", "$$literal", "$!literal", "!command"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			desktopFixtureFile(t, root, "models.json", `{"providers":{"arbitrary":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"`+value+`"}}}`)
			token, err := credentialFromPiFile(filepath.Join(root, "models.json"))
			if err == nil || token != "" {
				t.Fatal("unsupported Pi value resolution was accepted as a credential")
			}
		})
	}
}

func TestPiEffectiveAuthOverridesSavedInlineCredential(t *testing.T) {
	for _, tc := range []struct {
		name string
		auth string
		want string
	}{
		{"active", `{"arbitrary":{"type":"api_key","key":"synthetic-active"}}`, "synthetic-active"},
		{"unrelated", `{"unrelated":{"type":"api_key","key":"synthetic-unrelated"}}`, "synthetic-saved"},
		{"missing-key", `{"arbitrary":{"type":"api_key"}}`, ""},
		{"empty-key", `{"arbitrary":{"type":"api_key","key":""}}`, ""},
		{"oauth", `{"arbitrary":{"type":"oauth","access":"synthetic-access"}}`, ""},
		{"dynamic", `{"arbitrary":{"type":"api_key","key":"$SYNTHETIC_PI_KEY"}}`, ""},
		{"null-entry", `{"arbitrary":null}`, ""},
		{"malformed", `{"arbitrary":`, ""},
		{"duplicate", `{"arbitrary":{"type":"api_key","key":"one"},"arbitrary":{"type":"api_key","key":"two"}}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			desktopFixtureFile(t, root, "models.json", `{"providers":{"arbitrary":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"synthetic-saved"}}}`)
			desktopFixtureFile(t, root, "auth.json", tc.auth)
			token, err := credentialFromPiFile(filepath.Join(root, "models.json"))
			if tc.want == "" {
				if err == nil || token != "" {
					t.Fatal("unsupported active authentication fell back to the saved credential")
				}
			} else if err != nil || token != tc.want {
				t.Fatal("effective authentication precedence was not preserved")
			}
		})
	}
}

func TestPiBuiltInEnvironmentPrecedenceNeverFallsBack(t *testing.T) {
	// Complete raw-source set at Pi 71dca871, not a sample of model vendors.
	// env-api-keys.ts SHA-256: 6876b915ffe1342f8be69ef1f75721c8a265a332a1bbbf5afe36b6f80b332552.
	for _, id := range []string{
		"ant-ling", "qwen-token-plan", "qwen-token-plan-cn", "qwen-token-plan-individual",
		"openai", "azure-openai-responses", "nvidia", "deepseek", "google", "google-vertex",
		"groq", "cerebras", "xai", "radius", "openrouter", "vercel-ai-gateway",
		"zai", "zai-coding-cn", "mistral", "minimax", "minimax-cn",
		"moonshotai", "moonshotai-cn", "huggingface", "fireworks", "together", "baseten",
		"opencode", "opencode-go", "kimi-coding", "cloudflare-workers-ai", "cloudflare-ai-gateway",
		"xiaomi", "xiaomi-token-plan-cn", "xiaomi-token-plan-ams", "xiaomi-token-plan-sgp",
		"github-copilot", "anthropic", "amazon-bedrock",
	} {
		t.Run(id, func(t *testing.T) {
			root := t.TempDir()
			desktopFixtureFile(t, root, "models.json", `{"providers":{"`+id+`":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"synthetic-saved"}}}`)
			if token, err := credentialFromPiFile(filepath.Join(root, "models.json")); err == nil || token != "" {
				t.Fatal("saved fallback accepted despite unrepresented environment authentication")
			}
			desktopFixtureFile(t, root, "auth.json", `{"`+id+`":{"type":"api_key","key":"synthetic-active"}}`)
			if token, err := credentialFromPiFile(filepath.Join(root, "models.json")); err != nil || token != "synthetic-active" {
				t.Fatal("higher-priority inline authentication was not used")
			}
		})
	}
	t.Setenv("OPENAI_API_KEY", "synthetic-unrelated")
	for _, id := range []string{"puretokens", "arbitrary", "ollama", "inception", "alibaba"} {
		root := t.TempDir()
		desktopFixtureFile(t, root, "models.json", `{"providers":{"`+id+`":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"synthetic-saved"}}}`)
		if token, err := credentialFromPiFile(filepath.Join(root, "models.json")); err != nil || token != "synthetic-saved" {
			t.Fatal("unrelated environment authentication changed a custom provider")
		}
	}
}

func TestPiUnusableAuthStoreNeverFallsBack(t *testing.T) {
	root := t.TempDir()
	desktopFixtureFile(t, root, "models.json", `{"providers":{"arbitrary":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"synthetic-saved"}}}`)
	if err := os.Mkdir(filepath.Join(root, "auth.json"), 0700); err != nil {
		t.Fatal(err)
	}
	if token, err := credentialFromPiFile(filepath.Join(root, "models.json")); err == nil || token != "" {
		t.Fatal("unreadable authentication store fell back to a possibly stale credential")
	}
}

func TestPiSavedConnection(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`{"providers":{"arbitrary":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"synthetic-pi"}}}`, true},
		{`{"providers":{"puretokens":{"baseUrl":"https://example.invalid/v1","api":"openai-completions","apiKey":"synthetic-pi"}}}`, false},
		{`{"providers":{"a":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","apiKey":"!secret-command"}}}`, false},
		{`{"providers":{},"providers":{}}`, false},
		{`{"providers":{"a":{"baseUrl":"https://api.puretokensx.com/v1"},"b":{"baseUrl":"https://api.puretokensx.com/v1"}}}`, false},
	} {
		root := t.TempDir()
		desktopFixtureFile(t, root, "models.json", tc.body)
		token, err := credentialFromPiFile(filepath.Join(root, "models.json"))
		if tc.valid {
			if err != nil || token != "synthetic-pi" {
				t.Fatal("supported Pi record rejected")
			}
		} else if err == nil || token != "" {
			t.Fatal("unsafe Pi record accepted")
		}
	}
}

func TestPiRoot(t *testing.T) {
	home := t.TempDir()
	root, err := kimiQoderRoot("pi", home, func(string) string { return "" })
	if err != nil || root != filepath.Join(home, ".pi", "agent") {
		t.Fatal("incorrect Pi root")
	}
	for _, value := range []string{"relative", home + "/../other"} {
		if _, err := kimiQoderRoot("pi", home, func(string) string { return value }); err == nil {
			t.Fatal("unsafe override accepted")
		}
	}
}
