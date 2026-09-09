package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestZCodeConnectionSelection(t *testing.T) {
	good := `{"enabled":true,"kind":"openai-compatible","options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"synthetic-zcode-key"}}`
	for _, tc := range []struct {
		name, body string
		ok         bool
	}{
		{"unique", `{"provider":{"arbitrary":` + good + `}}`, true},
		{"legacy model ignored", `{"model":"other/model","provider":{"arbitrary":` + good + `,"other":{"enabled":true,"options":{"baseURL":"https://example.invalid"}}}}`, true},
		{"ambiguous", `{"provider":{"one":` + good + `,"two":` + good + `}}`, false},
		{"disabled", `{"provider":{"one":` + strings.Replace(good, `true`, `false`, 1) + `}}`, false},
		{"wrong endpoint", `{"provider":{"one":` + strings.Replace(good, "api.puretokensx.com", "example.invalid", 1) + `}}`, false},
		{"missing key", `{"provider":{"one":` + strings.Replace(good, "synthetic-zcode-key", "", 1) + `}}`, false},
		{"reference", `{"provider":{"one":` + strings.Replace(good, "synthetic-zcode-key", "${KEY}", 1) + `}}`, false},
		{"wrong kind", `{"provider":{"one":` + strings.Replace(good, "openai-compatible", "other", 1) + `}}`, false},
		{"duplicate", `{"provider":{"one":` + good + `,"one":` + good + `}}`, false},
		{"escaped duplicate", `{"provider":{},"prov\u0069der":{"one":` + good + `}}`, false},
		{"trailing", `{"provider":{"one":` + good + `}} {}`, false},
		{"deep", strings.Repeat("[", 70) + strings.Repeat("]", 70), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			desktopFixtureFile(t, root, "config.json", tc.body)
			token, err := credentialFromZCodeFile(filepath.Join(root, "config.json"))
			if tc.ok {
				if err != nil || token != "synthetic-zcode-key" {
					t.Fatal("valid connection rejected")
				}
			} else if err == nil || token != "" {
				t.Fatal("unsafe connection accepted")
			}
			if err != nil && strings.Contains(err.Error(), "synthetic-zcode-key") {
				t.Fatal("credential disclosure")
			}
		})
	}
}
func TestZCodeExplicitRoot(t *testing.T) {
	home := t.TempDir()
	base := filepath.Join(home, "space dir")
	for _, tc := range []struct{ base, want string }{{"", filepath.Join(home, ".zcode")}, {base, filepath.Join(base, ".zcode")}} {
		got, err := zcodeRoot(home, tc.base)
		if err != nil || got != tc.want {
			t.Fatal("incorrect data base selection")
		}
	}
	if _, err := zcodeRoot(home, "relative"); err == nil {
		t.Fatal("relative override accepted")
	}
}

func TestZCodeHostDispatchAndDoctor(t *testing.T) {
	base := t.TempDir()
	t.Setenv("ZCODE_DATA_BASE_DIR", base)
	desktopFixtureFile(t, base, ".zcode/v2/config.json", `{"provider":{"one":{"enabled":true,"kind":"openai-compatible","options":{"baseURL":"https://api.puretokensx.com/v1","apiKey":"synthetic-zcode-key"}}}}`)
	if token, err := credentialForHost("zcode"); err != nil || token != "synthetic-zcode-key" {
		t.Fatal("host dispatch failed")
	}
	locations := doctorHostLocations("zcode", t.TempDir(), func(key string) string {
		if key == "ZCODE_DATA_BASE_DIR" {
			return base
		}
		return ""
	})
	if len(locations) != 1 || locations[0].path != filepath.Join(base, ".zcode", "skills") {
		t.Fatal("doctor used an unrelated skill root")
	}
	missing := t.TempDir()
	t.Setenv("ZCODE_DATA_BASE_DIR", missing)
	desktopFixtureFile(t, missing, ".zcode/model-providers.json", `{"apiKey":"synthetic-legacy"}`)
	if token, err := credentialForHost("zcode"); err == nil || token != "" {
		t.Fatal("legacy fallback accepted")
	}
}
