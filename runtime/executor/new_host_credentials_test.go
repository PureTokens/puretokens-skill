package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewHostCredentialSelection(t *testing.T) {
	kimi := "default_model = 'custom-model'\n[models.custom-model]\nprovider = 'arbitrary-name'\n[providers.arbitrary-name]\ntype = 'openai'\nbase_url = 'https://api.puretokensx.com/v1'\napi_key = 'synthetic-new-host'\n"
	qoder := `{"providers":{"arbitrary-name":{"type":"openai-compatible","protocol":"openai","baseUrl":"https://api.puretokensx.com/v1","apiKey":"synthetic-new-host"}}}`
	for _, host := range []string{"kimi-code", "qoder"} {
		t.Run(host, func(t *testing.T) {
			good := kimi
			read := credentialFromKimiFile
			if host == "qoder" {
				good = qoder
				read = credentialFromQoderFile
			}
			cases := []struct {
				name, data string
				ok         bool
			}{
				{"valid arbitrary name", good, true},
				{"foreign endpoint", strings.ReplaceAll(good, "api.puretokensx.com", "example.invalid"), false},
				{"missing credential", strings.ReplaceAll(good, "synthetic-new-host", ""), false},
				{"environment reference", strings.ReplaceAll(good, "synthetic-new-host", "${SECRET}"), false},
				{"unsupported type", strings.ReplaceAll(good, "openai", "unsupported"), false},
				{"malformed", good + "[", false},
			}
			if host == "kimi-code" {
				cases = append(cases, struct {
					name, data string
					ok         bool
				}{"missing selected model", strings.Replace(good, "custom-model'", "missing'", 1), false})
				cases = append(cases, struct {
					name, data string
					ok         bool
				}{"duplicate TOML", good + "api_key = 'other'\n", false})
				for _, kind := range []string{"anthropic", "openai_responses"} {
					cases = append(cases, struct {
						name, data string
						ok         bool
					}{kind, strings.ReplaceAll(good, "'openai'", "'"+kind+"'"), true})
				}
			} else {
				entry := `{"type":"openai-compatible","protocol":"openai","baseUrl":"https://api.puretokensx.com/v1","apiKey":"synthetic-new-host"}`
				cases = append(cases, struct {
					name, data string
					ok         bool
				}{"ambiguous", `{"providers":{"a":` + entry + `,"b":` + entry + `}}`, false}, struct {
					name, data string
					ok         bool
				}{"duplicate JSON", `{"providers":{},"providers":{"a":` + entry + `}}`, false}, struct {
					name, data string
					ok         bool
				}{"deep", strings.Repeat("[", 70) + strings.Repeat("]", 70), false})
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					root := t.TempDir()
					desktopFixtureFile(t, root, "fixture", tc.data)
					token, err := read(filepath.Join(root, "fixture"))
					if tc.ok {
						if err != nil || token != "synthetic-new-host" {
							t.Fatal("valid connection rejected")
						}
					} else if err == nil || token != "" {
						t.Fatal("unsafe connection accepted")
					}
					if err != nil && strings.Contains(err.Error(), "synthetic-new-host") {
						t.Fatal("credential disclosure")
					}
				})
			}
		})
	}
}

func TestNewHostRootsAndDispatch(t *testing.T) {
	for _, host := range []string{"kimi-code", "qoder"} {
		t.Run(host, func(t *testing.T) {
			root := t.TempDir()
			variable := "KIMI_CODE_HOME"
			file := "config.toml"
			data := "default_model='m'\n[models.m]\nprovider='p'\n[providers.p]\ntype='openai'\nbase_url='https://api.puretokensx.com/v1'\napi_key='synthetic-new-host'\n"
			if host == "qoder" {
				variable = "QODER_CONFIG_DIR"
				file = "settings.json"
				data = `{"providers":{"p":{"type":"openai-compatible","protocol":"openai","baseUrl":"https://api.puretokensx.com/v1","apiKey":"synthetic-new-host"}}}`
			}
			t.Setenv(variable, root)
			desktopFixtureFile(t, root, file, data)
			if token, err := credentialForHost(host); err != nil || token != "synthetic-new-host" {
				t.Fatal("dispatch failed")
			}
			env := func(k string) string {
				if k == variable {
					return root
				}
				return ""
			}
			locations := doctorHostLocations(host, t.TempDir(), env)
			expectedLocations := 1
			if host == "kimi-code" {
				expectedLocations = 2
			}
			if len(locations) != expectedLocations || locations[0].path != filepath.Join(root, "skills") {
				t.Fatal("doctor root mismatch")
			}
			t.Setenv(variable, "relative")
			if _, err := credentialForHost(host); err == nil {
				t.Fatal("relative root accepted")
			}
		})
	}
	home := t.TempDir()
	for _, name := range []string{"..", "x/y", "x\\y"} {
		if _, err := kimiQoderRoot("qoder", home, func(k string) string {
			if k == "QODER_CONFIG_DIR_NAME" {
				return name
			}
			return ""
		}); err == nil {
			t.Fatal("invalid directory accepted")
		}
	}
}

func TestNewHostDirectoryPrecedenceAndNoFallback(t *testing.T) {
	home := t.TempDir()
	for _, host := range []string{"kimi-code", "qoder"} {
		root, err := kimiQoderRoot(host, home, func(string) string { return "" })
		if err != nil || root != filepath.Join(home, "."+host) {
			t.Fatal("incorrect default directory")
		}
	}
	base := filepath.Join(home, "parent with spaces")
	selected, err := kimiQoderRoot("qoder", home, func(k string) string {
		if k == "QODER_CLI_HOME" {
			return base
		}
		if k == "QODER_CONFIG_DIR_NAME" {
			return "alternate"
		}
		return ""
	})
	if err != nil || selected != filepath.Join(base, "alternate") {
		t.Fatal("incorrect parent override")
	}
	explicit := filepath.Join(home, "explicit")
	selected, err = kimiQoderRoot("qoder", home, func(k string) string {
		switch k {
		case "QODER_CONFIG_DIR":
			return explicit
		case "QODER_CLI_HOME":
			return "relative"
		case "QODER_CONFIG_DIR_NAME":
			return "../invalid"
		}
		return ""
	})
	if err != nil || selected != explicit {
		t.Fatal("explicit directory precedence lost")
	}
	t.Setenv("KIMI_CODE_HOME", filepath.Join(home, "absent-kimi"))
	t.Setenv("QODER_CONFIG_DIR", filepath.Join(home, "absent-qoder"))
	t.Setenv("OPENAI_API_KEY", "synthetic-must-not-use")
	for _, host := range []string{"kimi-code", "qoder"} {
		if token, err := credentialForHost(host); err == nil || token != "" {
			t.Fatal("missing file used fallback")
		}
	}
}

func TestUnknownHostDoesNotResolveQoderRoot(t *testing.T) {
	calls := 0
	_, err := kimiQoderRoot("unknown", t.TempDir(), func(string) string { calls++; return "" })
	if err == nil || calls != 0 {
		t.Fatal("unknown host reached configuration resolution")
	}
}
