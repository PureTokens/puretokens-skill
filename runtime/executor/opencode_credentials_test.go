package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func opencodeFixture(t *testing.T, root, name string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	desktopFixtureFile(t, root, name, string(body))
}

func opencodeTestEnvironment(t *testing.T) (opencodeEnvironment, map[string]string) {
	t.Helper()
	root := t.TempDir()
	home, cwd := filepath.Join(root, "home"), filepath.Join(root, "workspace", "nested")
	for _, dir := range []string{home, cwd} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	desktopFixtureFile(t, filepath.Dir(cwd), ".git", "")
	env := map[string]string{}
	return opencodeEnvironment{home: home, cwd: cwd, getenv: func(name string) string { return env[name] }}, env
}

func TestOpenCodeLayeredConfigurationPrecedence(t *testing.T) {
	e, env := opencodeTestEnvironment(t)
	config := filepath.Join(e.home, ".config", "opencode")
	env["OPENCODE_CONFIG"] = filepath.Join(e.home, "explicit.jsonc")
	env["OPENCODE_CONFIG_DIR"] = filepath.Join(e.home, "explicit-directory")
	e.managed = filepath.Join(e.home, "managed")
	opencodeFixture(t, e.home, "explicit.jsonc", map[string]any{})
	paths := []string{
		filepath.Join(config, "config.json"),
		filepath.Join(config, "opencode.json"),
		filepath.Join(config, "opencode.jsonc"),
		env["OPENCODE_CONFIG"],
		filepath.Join(filepath.Dir(e.cwd), "opencode.json"),
		filepath.Join(e.cwd, "opencode.jsonc"),
		filepath.Join(e.cwd, ".opencode", "opencode.json"),
		filepath.Join(filepath.Dir(e.cwd), ".opencode", "opencode.jsonc"),
		filepath.Join(e.home, ".opencode", "opencode.json"),
		filepath.Join(env["OPENCODE_CONFIG_DIR"], "opencode.jsonc"),
		filepath.Join(e.managed, "opencode.json"),
	}
	for i, path := range paths {
		want := "synthetic-layer-" + string(rune('a'+i))
		options := map[string]any{"apiKey": want}
		if i == 0 {
			options["baseURL"] = "https://api.puretokensx.com/v1"
		}
		opencodeFixture(t, filepath.Dir(path), filepath.Base(path), map[string]any{
			"provider": map[string]any{"p": map[string]any{"options": options}},
		})
		if token, err := credentialFromOpenCodeEnvironment(e); err != nil || token != want {
			t.Fatalf("layer %d did not take precedence", i)
		}
	}
	desktopFixtureFile(t, e.managed, "opencode.jsonc", `{"provider":`)
	if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
		t.Fatal("malformed higher-priority config fell back")
	}
}

func TestOpenCodeProjectToggleAndXDGRoots(t *testing.T) {
	e, env := opencodeTestEnvironment(t)
	env["XDG_CONFIG_HOME"] = filepath.Join(e.home, "xdg config")
	env["XDG_DATA_HOME"] = filepath.Join(e.home, "xdg data")
	opencodeFixture(t, filepath.Join(env["XDG_CONFIG_HOME"], "opencode"), "opencode.json", map[string]any{
		"provider": map[string]any{"p": map[string]any{"options": map[string]any{"baseURL": "https://api.puretokensx.com/v1"}}},
	})
	opencodeFixture(t, filepath.Join(env["XDG_DATA_HOME"], "opencode"), "auth.json", map[string]any{
		"p": map[string]any{"type": "api", "key": "synthetic-stored"},
	})
	if token, err := credentialFromOpenCodeEnvironment(e); err != nil || token != "synthetic-stored" {
		t.Fatal("explicit XDG records were not used")
	}
	opencodeFixture(t, e.cwd, "opencode.json", map[string]any{
		"provider": map[string]any{"p": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-project")},
	})
	if token, err := credentialFromOpenCodeEnvironment(e); err != nil || token != "synthetic-project" {
		t.Fatal("project override did not win")
	}
	env["OPENCODE_DISABLE_PROJECT_CONFIG"] = "true"
	if token, err := credentialFromOpenCodeEnvironment(e); err != nil || token != "synthetic-stored" {
		t.Fatal("disabled project config was still used")
	}
}

func TestOpenCodeNativeCredentialPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, kind, want string
		inline           any
		setInline        bool
	}{
		{"stored-api", "api", "synthetic-stored", nil, false},
		{"inline-wins", "api", "synthetic-inline", "synthetic-inline", true},
		{"empty-inline-stops", "api", "", "", true},
		{"null-inline-stops", "api", "", nil, true},
		{"external-inline-stops", "api", "", "{file:/private-config}", true},
		{"oauth-stops", "oauth", "", nil, false},
		{"oauth-inline-stops", "oauth", "", "synthetic-inline", true},
		{"remote-config-stops", "wellknown", "", "synthetic-inline", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, _ := opencodeTestEnvironment(t)
			options := map[string]any{"baseURL": "https://api.puretokensx.com/v1"}
			if tc.setInline {
				options["apiKey"] = tc.inline
			}
			opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
				"provider": map[string]any{"p": map[string]any{"options": options}},
			})
			opencodeFixture(t, filepath.Join(e.home, ".local", "share", "opencode"), "auth.json", map[string]any{
				"p":     map[string]any{"type": tc.kind, "key": "synthetic-stored"},
				"other": map[string]any{"type": "api", "key": "synthetic-unrelated"},
			})
			token, err := credentialFromOpenCodeEnvironment(e)
			if tc.want == "" {
				if err == nil || token != "" {
					t.Fatal("unsupported authentication fell back")
				}
				assertLocalConnectionDiagnostic(t, err)
			} else if err != nil || token != tc.want {
				t.Fatal("credential precedence was not preserved")
			}
		})
	}
}

func TestOpenCodeEndpointBeforeCredentialStore(t *testing.T) {
	e, _ := opencodeTestEnvironment(t)
	opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
		"model":    "p/model",
		"provider": map[string]any{"p": opencodeProvider("https://example.invalid/v1", "synthetic-secret")},
	})
	store := filepath.Join(e.home, ".local", "share", "opencode", "auth.json")
	if err := os.MkdirAll(store, 0700); err != nil {
		t.Fatal(err)
	}
	token, err := credentialFromOpenCodeEnvironment(e)
	status, _, _ := credentialFailureDetails(err)
	if token != "" || status != "active_connection_not_puretokens" {
		t.Fatal("credential store was consulted before rejecting the endpoint")
	}
}

func TestOpenCodeUnsupportedOverridesAndPlugins(t *testing.T) {
	for _, variable := range []string{
		"OPENCODE_CONFIG_CONTENT", "OPENCODE_AUTH_CONTENT", "OPENCODE_TEST_HOME",
		"OPENCODE_TEST_MANAGED_CONFIG_DIR", "OPENCODE_CONSOLE_TOKEN",
		"XDG_CONFIG_HOME", "XDG_DATA_HOME", "OPENCODE_CONFIG", "OPENCODE_CONFIG_DIR",
	} {
		t.Run(variable, func(t *testing.T) {
			e, env := opencodeTestEnvironment(t)
			env[variable] = "relative-or-unsupported"
			token, err := credentialFromOpenCodeEnvironment(e)
			if token != "" || err == nil {
				t.Fatal("unsupported configuration override was accepted")
			}
			assertLocalConnectionDiagnostic(t, err)
		})
	}
	for _, location := range []string{"plugins", "plugin"} {
		e, _ := opencodeTestEnvironment(t)
		if err := os.MkdirAll(filepath.Join(e.home, ".config", "opencode", location), 0700); err != nil {
			t.Fatal(err)
		}
		if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
			t.Fatal("unrepresented plugin hooks were ignored")
		}
	}
	e, _ := opencodeTestEnvironment(t)
	opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{"plugin": []any{"synthetic-plugin"}})
	opencodeFixture(t, e.cwd, "opencode.json", map[string]any{"plugin": []any{}})
	if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
		t.Fatal("higher empty plugin array hid lower-priority hooks")
	}
}

func TestOpenCodeProviderFilters(t *testing.T) {
	for _, tc := range []struct {
		name, field string
		value       any
		ok          bool
	}{
		{"allow-one", "enabled_providers", []any{"a"}, true},
		{"disable-other", "disabled_providers", []any{"b"}, true},
		{"disable-all", "disabled_providers", []any{"a", "b"}, false},
		{"enable-none", "enabled_providers", []any{}, false},
		{"bad-filter", "enabled_providers", "a", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			opencodeFixture(t, root, "opencode.json", map[string]any{
				"provider": map[string]any{
					"a": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-a"),
					"b": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-b"),
				},
				tc.field: tc.value,
			})
			token, err := credentialFromOpenCodeFile(filepath.Join(root, "opencode.json"))
			if tc.ok {
				if err != nil || token != "synthetic-a" {
					t.Fatal("provider filters were not applied")
				}
			} else if err == nil || token != "" {
				t.Fatal("invalid or empty selection was accepted")
			}
		})
	}
}

func TestOpenCodeJSONBoundaries(t *testing.T) {
	for _, body := range []string{
		`{}` + `{}`,
		strings.Repeat(`{"x":`, 66) + `{}` + strings.Repeat("}", 66),
		`{"x":"` + strings.Repeat("x", maxConfigBytes) + `"}`,
		`{"provider":{}/*`,
	} {
		root := t.TempDir()
		desktopFixtureFile(t, root, "record.jsonc", body)
		if _, err := readOpenCodeObject(filepath.Join(root, "record.jsonc")); err == nil {
			t.Fatal("unbounded or malformed configuration was accepted")
		}
	}
	root := t.TempDir()
	desktopFixtureFile(t, root, "auth.json", `{"p":{"type":"api","key":"synthetic-only"},}`)
	if _, err := readOpenCodeJSON(filepath.Join(root, "auth.json"), false); err == nil {
		t.Fatal("native JSON credential store accepted JSONC")
	}
}

func opencodeProvider(endpoint, token string) map[string]any {
	return map[string]any{"options": map[string]any{"baseURL": endpoint, "apiKey": token}}
}

func TestOpenCodeOptionalDefaultAndSelectionBoundary(t *testing.T) {
	for _, tc := range []struct {
		name   string
		model  any
		set    bool
		second bool
		want   string
	}{
		{"omitted-default", nil, false, false, ""},
		{"explicit-default", "primary/model", true, true, ""},
		{"ambiguous-without-default", nil, false, true, "active_connection_ambiguous"},
		{"empty-is-not-absent", "", true, false, "active_connection_selection_unconfirmed"},
		{"null-is-not-absent", nil, true, false, "active_connection_selection_unconfirmed"},
		{"unqualified-default", "model", true, false, "active_connection_selection_unconfirmed"},
		{"missing-model-id", "primary/", true, false, "active_connection_selection_unconfirmed"},
		{"unknown-selected-provider", "missing/model", true, false, "active_connection_selection_unconfirmed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			providers := map[string]any{"primary": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-only")}
			if tc.second {
				providers["secondary"] = opencodeProvider("https://api.puretokensx.com/v1", "synthetic-other")
			}
			doc := map[string]any{"provider": providers}
			if tc.set {
				doc["model"] = tc.model
			}
			opencodeFixture(t, root, "opencode.json", doc)
			token, err := credentialFromOpenCodeFile(filepath.Join(root, "opencode.json"))
			if tc.want == "" {
				if err != nil || token != "synthetic-only" {
					t.Fatal("supported saved connection was not resolved")
				}
			} else {
				status, _, _ := credentialFailureDetails(err)
				if token != "" || status != tc.want {
					t.Fatalf("unexpected diagnostic: %s", status)
				}
				assertLocalConnectionDiagnostic(t, err)
			}
		})
	}
}

func TestOpenCodeNoCrossSelectionFallback(t *testing.T) {
	root := t.TempDir()
	opencodeFixture(t, root, "opencode.json", map[string]any{
		"model": "selected/model",
		"provider": map[string]any{
			"selected": opencodeProvider("https://example.invalid/v1", "synthetic-secret"),
			"other":    opencodeProvider("https://api.puretokensx.com/v1", "synthetic-only"),
		},
	})
	token, err := credentialFromOpenCodeFile(filepath.Join(root, "opencode.json"))
	status, _, _ := credentialFailureDetails(err)
	if token != "" || status != "active_connection_not_puretokens" {
		t.Fatal("unselected provider was used as a fallback")
	}
	assertLocalConnectionDiagnostic(t, err)
}

func TestOpenCodeJSONCAndUnsupportedReferences(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
	}{
		{"comments-trailing-comma", `{
			// The default chat model is optional.
			"provider":{"p":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"synthetic-only",},},},
		}`, ""},
		{"duplicate-key", `{"provider":{},"provider":{}}`, "active_connection_format_unsupported"},
		{"escaped-duplicate", `{"provider":{},"pro\u0076ider":{}}`, "active_connection_format_unsupported"},
		{"malformed", `{"provider":`, "active_connection_format_unsupported"},
		{"external-file", `{"provider":{"p":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"{file:/private-config}"}}}}`, "active_connection_format_unsupported"},
		{"external-env", `{"provider":{"p":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"{env:SYNTHETIC}"}}}}`, "active_connection_format_unsupported"},
		{"auth-header", `{"provider":{"p":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"synthetic-only","headers":{"Authorization":"synthetic-secret"}}}}}`, "active_connection_format_unsupported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			desktopFixtureFile(t, root, "opencode.json", tc.body)
			token, err := credentialFromOpenCodeFile(filepath.Join(root, "opencode.json"))
			if tc.want == "" {
				if err != nil || token != "synthetic-only" {
					t.Fatal("valid JSONC was rejected")
				}
			} else {
				status, _, _ := credentialFailureDetails(err)
				if token != "" || status != tc.want {
					t.Fatalf("unexpected diagnostic: %s", status)
				}
				assertLocalConnectionDiagnostic(t, err)
			}
		})
	}
}

func TestOpenCodeInvalidOverridesDoNotRestoreInheritedValues(t *testing.T) {
	for _, override := range []string{
		`{"provider":null}`,
		`{"provider":{"p":null}}`,
		`{"provider":{"p":{"options":null}}}`,
		`{"provider":{"p":{"options":{"baseURL":null}}}}`,
		`{"provider":{"p":{"options":{"baseURL":"https://example.invalid/v1"}}}}`,
		`{"provider":{"p":{"options":{"apiKey":""}}}}`,
		`{"provider":{"p":{"options":{"apiKey":null}}}}`,
		`{"disabled_providers":["p"]}`,
		`{"enabled_providers":[]}`,
		`{"provider":{"p":{"models":{"test":{"headers":{"Authorization":"synthetic"}}}}}}`,
	} {
		t.Run(override, func(t *testing.T) {
			e, _ := opencodeTestEnvironment(t)
			opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
				"model":    "p/test",
				"provider": map[string]any{"p": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-inherited")},
			})
			desktopFixtureFile(t, e.cwd, "opencode.jsonc", override)
			if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
				t.Fatal("invalid higher-priority selection reused an inherited credential")
			}
		})
	}
}

func TestOpenCodeDynamicKeysStopBeforeMerge(t *testing.T) {
	for _, override := range []string{
		`{"provider":{"p":{"options":{"{env:FIELD}":"synthetic-new"}}}}`,
		`{"provider":{"p":{"options":{"{env:ENDPOINT_FIELD}":"https://example.invalid/v1"}}}}`,
		`{"{env:MODEL_FIELD}":"other/test"}`,
		`{/* {env:COMMENT_OVERRIDE} */}`,
		`{"provider":{"p":{"options":{"\u007benv:FIELD}":"synthetic-new"}}}}`,
	} {
		t.Run(override, func(t *testing.T) {
			e, _ := opencodeTestEnvironment(t)
			opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
				"provider": map[string]any{"p": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-inherited")},
			})
			desktopFixtureFile(t, e.cwd, "opencode.jsonc", override)
			token, err := credentialFromOpenCodeEnvironment(e)
			if status, _, _ := credentialFailureDetails(err); status != "active_connection_format_unsupported" || token != "" {
				t.Fatal("dynamic override retained an inherited credential")
			}
		})
	}
}

func TestOpenCodeUnrelatedAuthAndStoreFailures(t *testing.T) {
	for _, body := range []string{
		`{"other":{"type":"api","key":"synthetic-unrelated"}}`,
		`{"p":null}`,
		`{"p":{"type":"api","key":""}}`,
		`{"p":{"type":"api","key":"{env:UNRELATED}"}}`,
		`{"p":{"type":"api","key":"synthetic-a","key":"synthetic-b"}}`,
		`{"p":`,
	} {
		e, _ := opencodeTestEnvironment(t)
		opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
			"provider": map[string]any{"p": map[string]any{"options": map[string]any{"baseURL": "https://api.puretokensx.com/v1"}}},
		})
		desktopFixtureFile(t, filepath.Join(e.home, ".local", "share", "opencode"), "auth.json", body)
		if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
			t.Fatal("unusable selected auth store fell back or accepted invalid credentials")
		}
	}
}

func TestOpenCodeManagedPreferencesAndNonRegularRecords(t *testing.T) {
	e, _ := opencodeTestEnvironment(t)
	opencodeFixture(t, filepath.Join(e.home, ".config", "opencode"), "opencode.json", map[string]any{
		"provider": map[string]any{"p": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-only")},
	})
	e.preferences = []string{filepath.Join(e.home, "managed.plist")}
	desktopFixtureFile(t, e.home, "managed.plist", "not-read")
	if token, err := credentialFromOpenCodeEnvironment(e); err == nil || token != "" {
		t.Fatal("unrepresented managed preferences were ignored")
	}
	root := t.TempDir()
	desktopFixtureFile(t, root, "target.json", `{}`)
	link := filepath.Join(root, "link.json")
	if err := os.Symlink(filepath.Join(root, "target.json"), link); err == nil {
		if _, err := readOpenCodeObject(link); err == nil {
			t.Fatal("symlinked config record was accepted")
		}
	}
	if _, err := readOpenCodeObject(root); err == nil {
		t.Fatal("directory config record was accepted")
	}
}

func TestOpenCodeAmbiguityDoesNotCompareKeys(t *testing.T) {
	root := t.TempDir()
	opencodeFixture(t, root, "opencode.json", map[string]any{
		"provider": map[string]any{
			"a": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-same"),
			"b": opencodeProvider("https://api.puretokensx.com/v1", "synthetic-same"),
		},
	})
	_, err := credentialFromOpenCodeFile(filepath.Join(root, "opencode.json"))
	if status, _, _ := credentialFailureDetails(err); status != "active_connection_ambiguous" {
		t.Fatal("equal keys collapsed distinct connections")
	}
}

func TestOpenCodeTotalConfigurationBudget(t *testing.T) {
	e, env := opencodeTestEnvironment(t)
	env["OPENCODE_CONFIG"] = filepath.Join(e.home, "explicit.json")
	config := filepath.Join(e.home, ".config", "opencode")
	for _, path := range []string{
		filepath.Join(config, "config.json"), filepath.Join(config, "opencode.json"),
		filepath.Join(config, "opencode.jsonc"), env["OPENCODE_CONFIG"],
		filepath.Join(e.cwd, "opencode.json"),
	} {
		desktopFixtureFile(t, filepath.Dir(path), filepath.Base(path), `{"padding":"`+strings.Repeat("x", maxConfigBytes-20)+`"}`)
	}
	_, err := credentialFromOpenCodeEnvironment(e)
	if status, _, _ := credentialFailureDetails(err); status != "active_connection_format_unsupported" {
		t.Fatal("combined config budget was not enforced")
	}
}

func TestOpenCodeDepthGuard(t *testing.T) {
	for _, body := range []string{
		`{"text":"` + strings.Repeat("[", 100) + `"}`,
		`{/*` + strings.Repeat("[", 100) + `*/"x":true}`,
		"{//" + strings.Repeat("[", 100) + "\n\"x\":true}",
	} {
		if _, err := parseOpenCodeObject([]byte(body), true); err != nil {
			t.Fatal("string/comment delimiters counted as nesting")
		}
	}
	body := bytes.Repeat([]byte("["), maxConfigBytes/2)
	if _, err := parseOpenCodeObject(body, true); err == nil {
		t.Fatal("deep input reached the recursive parser")
	}
}

func FuzzOpenCodeJSONC(f *testing.F) {
	for _, seed := range []string{`{}`, `{"x":1,"x":2}`, `{"x":[true,null,],/*x*/}`, `{"provider":{"p":{"options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"synthetic-only"}}}}`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > maxConfigBytes {
			return
		}
		doc, err := parseOpenCodeObject([]byte(value), true)
		if err != nil {
			assertLocalConnectionDiagnostic(t, err)
			return
		}
		_, _ = credentialFromOpenCodeDocument(doc, "")
	})
}
