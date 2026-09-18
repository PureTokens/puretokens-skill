package main

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clientFixtureRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func clientFixtureFile(t *testing.T, root, name, content string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

const hermesFixture = `model:
  provider: custom:fixture
  default: chat-model
providers:
  fixture:
    base_url: https://api.puretokensx.com/v1
    api_mode: chat_completions
    api_key: synthetic-client-fixture
`

func TestHermesInlineAndOverrideBoundaries(t *testing.T) {
	for _, test := range []struct {
		name   string
		config string
		auth   string
		ok     bool
	}{
		{"inline", hermesFixture, "", true},
		{"empty-pool", hermesFixture, `{"credential_pool":{}}`, true},
		{"pool-precedence", hermesFixture, `{"credential_pool":{"fixture":[{"access_token":"not-to-be-read"}]}}`, false},
		{"dynamic-key", strings.Replace(hermesFixture, "synthetic-client-fixture", "${SECRET}", 1), "", false},
		{"provider-env", hermesFixture + "    key_env: SECRET\n", "", false},
		{"model-override", strings.Replace(hermesFixture, "  default: chat-model", "  default: chat-model\n  api_key: other-fixture", 1), "", false},
		{"foreign-endpoint", strings.Replace(hermesFixture, "api.puretokensx.com", "example.invalid", 1), "malformed", false},
		{"duplicate-provider", hermesFixture + "  fixture: {}\n", "", false},
		{"disabled", hermesFixture + "    enabled: false\n", "", false},
		{"command", hermesFixture + "    key_cmd: forbidden-command\n", "", false},
		{"broken-auth", hermesFixture, "{", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := clientFixtureRoot(t)
			clientFixtureFile(t, root, "config.yaml", test.config)
			if test.auth != "" {
				clientFixtureFile(t, root, "auth.json", test.auth)
			}
			token, err := credentialFromHermesRoot(root)
			if (err == nil) != test.ok || (test.ok && token != "synthetic-client-fixture") || (!test.ok && token != "") {
				t.Fatal("unexpected credential resolution outcome")
			}
			if err != nil && strings.Contains(err.Error(), "synthetic") {
				t.Fatal("credential disclosed")
			}
		})
	}
}

const evoxSettingsFixture = `{"defaultProvider":"fixture","defaultModel":"chat-model","defaultInstanceId":"default","config":{"modelPolicy":{"userPreference":{"model":"fixture/chat-model","instanceId":"default"}}},"providers":{"fixture":{"baseUrl":"https://api.puretokensx.com/v1","api":"openai-completions","models":[{"id":"chat-model","api":"openai-completions"}]}}}`
const evoxAuthFixture = `fixture:
  type: api_key
  api: openai-completions
  base_url: https://api.puretokensx.com/v1
  key: synthetic-client-fixture
`

func TestEvoXEffectiveDefaultAndAuthentication(t *testing.T) {
	for _, test := range []struct {
		name     string
		settings string
		auth     string
		legacy   bool
		ok       bool
	}{
		{"inline", evoxSettingsFixture, evoxAuthFixture, false, true},
		{"selection-conflict", strings.Replace(evoxSettingsFixture, "fixture/chat-model", "other/chat-model", 1), evoxAuthFixture, false, false},
		{"model-api-override", strings.Replace(evoxSettingsFixture, `"id":"chat-model","api":"openai-completions"`, `"id":"chat-model","api":"openai-responses"`, 1), evoxAuthFixture, false, false},
		{"model-url-override", strings.Replace(evoxSettingsFixture, `"id":"chat-model"`, `"id":"chat-model","baseUrl":"https://example.invalid"`, 1), evoxAuthFixture, false, false},
		{"endpoint-conflict", evoxSettingsFixture, strings.Replace(evoxAuthFixture, "/v1", "/", 1), false, false},
		{"oauth", evoxSettingsFixture, strings.Replace(evoxAuthFixture, "api_key", "oauth", 1), false, false},
		{"dynamic", evoxSettingsFixture, strings.Replace(evoxAuthFixture, "synthetic-client-fixture", "'!command'", 1), false, false},
		{"legacy", evoxSettingsFixture, evoxAuthFixture, true, false},
		{"duplicate-json", `{"defaultModel":"a","defaultModel":"b"}`, evoxAuthFixture, false, false},
		{"duplicate-yaml", evoxSettingsFixture, evoxAuthFixture + "  key: other\n", false, false},
		{"missing-auth", evoxSettingsFixture, "", false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := clientFixtureRoot(t)
			clientFixtureFile(t, root, "settings.json", test.settings)
			if test.auth != "" {
				clientFixtureFile(t, root, "auth.yaml", test.auth)
			}
			if test.legacy {
				clientFixtureFile(t, root, "auth.json", `{}`)
			}
			token, err := credentialFromEvoXRoot(root)
			if (err == nil) != test.ok || (test.ok && token != "synthetic-client-fixture") || (!test.ok && token != "") {
				t.Fatal("unexpected credential resolution outcome")
			}
		})
	}
}

func TestVSCodeJSONCExactModelBoundary(t *testing.T) {
	const model = `{"id":"chat-model","url":"https://api.puretokensx.com/v1/responses","requestHeaders":{"Authorization":"Bearer synthetic-client-fixture"}}`
	for _, test := range []struct {
		name, body string
		ok         bool
	}{
		{"inline", `[// native JSONC
{"vendor":"customendpoint","models":[` + model + `,]},]`, true},
		{"multiple-even-with-same-key", `[{"vendor":"customendpoint","models":[` + model + `,` + model + `]}]`, false},
		{"foreign", `[{"vendor":"customendpoint","models":[` + strings.Replace(model, "api.puretokensx.com", "example.invalid", 1) + `]}]`, false},
		{"group-key", `[{"vendor":"customendpoint","apiKey":"unrepresented","models":[` + model + `]}]`, false},
		{"dynamic", `[{"vendor":"customendpoint","models":[` + strings.Replace(model, "synthetic-client-fixture", "${input:secret}", 1) + `]}]`, false},
		{"duplicate", `[{"vendor":"customendpoint","models":[],"models":[` + model + `]}]`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := clientFixtureFile(t, clientFixtureRoot(t), "chatLanguageModels.json", test.body)
			token, err := credentialFromVSCodeFile(path)
			if (err == nil) != test.ok || (test.ok && token != "synthetic-client-fixture") || (!test.ok && token != "") {
				t.Fatal("unexpected credential resolution outcome")
			}
		})
	}
}

func TestNewClientRootsAndFileGuards(t *testing.T) {
	home := clientFixtureRoot(t)
	env := map[string]string{"APPDATA": filepath.Join(home, "roaming"), "LOCALAPPDATA": filepath.Join(home, "local")}
	getenv := func(key string) string { return env[key] }
	for _, test := range []struct{ host, goos, want string }{
		{"hermes", "darwin", filepath.Join(home, ".hermes")},
		{"hermes", "windows", filepath.Join(home, "local", "hermes")},
		{"evox", "darwin", filepath.Join(home, ".evox", "agent")},
		{"evox", "windows", filepath.Join(home, ".evox", "agent")},
		{"vscode", "darwin", filepath.Join(home, "Library", "Application Support", "Code", "User")},
		{"vscode", "windows", filepath.Join(home, "roaming", "Code", "User")},
		{"octop", "windows", filepath.Join(home, ".octop")},
	} {
		got, err := newClientRoot(test.host, test.goos, home, getenv)
		if err != nil || got != test.want {
			t.Fatalf("%s/%s path mismatch", test.host, test.goos)
		}
	}
	for _, name := range []string{"HERMES_HOME", "EVOX_AGENT_DIR", "OCTOP_HOME"} {
		host := map[string]string{"HERMES_HOME": "hermes", "EVOX_AGENT_DIR": "evox", "OCTOP_HOME": "octop"}[name]
		for _, value := range []string{"relative", home + "/../other", "//server/share", home + "\n"} {
			env[name] = value
			if _, err := newClientRoot(host, "darwin", home, getenv); err == nil {
				t.Fatal("unsafe root accepted")
			}
		}
		delete(env, name)
	}
	env["EVOX_AGENT_DIR"], env["EVOX_CODING_AGENT_DIR"] = home, home+"/other"
	if _, err := newClientRoot("evox", "darwin", home, getenv); err == nil {
		t.Fatal("conflicting EvoX roots accepted")
	}
	target := clientFixtureFile(t, home, "target", "{}")
	link := filepath.Join(home, "link")
	if err := os.Symlink(target, link); err == nil {
		if _, err := readClientJSON(link, false); err == nil {
			t.Fatal("symlinked config accepted")
		}
	}
	hardlink := filepath.Join(home, "hardlink")
	if err := os.Link(target, hardlink); err == nil {
		if _, err := readClientJSON(target, false); err == nil {
			t.Fatal("hardlinked config accepted")
		}
	}
}

func octopFixture(t *testing.T, wal bool) (string, *sql.DB) {
	t.Helper()
	root := clientFixtureRoot(t)
	db, err := sql.Open("sqlite", filepath.Join(root, "octop.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if wal {
		if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA wal_autocheckpoint=0;"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE TABLE settings(key TEXT PRIMARY KEY, value TEXT);
CREATE TABLE providers(id INTEGER PRIMARY KEY,name TEXT,kind TEXT,base_url TEXT,api_key TEXT,enabled INTEGER,models_json TEXT);
INSERT INTO providers VALUES(1,'opaque-fixture','openai','https://api.puretokensx.com/v1','synthetic-client-fixture',1,'[{"id":"chat-model"}]');`); err != nil {
		t.Fatal(err)
	}
	return root, db
}

func TestOctopReadOnlySelectionAndWAL(t *testing.T) {
	for _, wal := range []bool{false, true} {
		root, db := octopFixture(t, wal)
		if _, err := db.Exec(`INSERT INTO settings VALUES('active_model','opaque-fixture/chat-model')`); err != nil {
			t.Fatal(err)
		}
		token, err := credentialFromOctopRoot(root, func(string) string { return "" })
		if err != nil || token != "synthetic-client-fixture" {
			t.Fatalf("WAL=%v selection failed: %v", wal, err)
		}
		var rows int
		if err := db.QueryRow("SELECT count(*) FROM providers").Scan(&rows); err != nil || rows != 1 {
			t.Fatal("provider contents changed")
		}
	}
}

func TestOctopAmbiguityOverridesAndNoCreation(t *testing.T) {
	root, db := octopFixture(t, false)
	for _, operation := range []string{
		`INSERT INTO providers SELECT 2,'other',kind,base_url,api_key,enabled,models_json FROM providers WHERE id=1`,
		`INSERT INTO settings VALUES('active_model','missing/chat-model')`,
	} {
		if _, err := db.Exec(operation); err != nil {
			t.Fatal(err)
		}
		if token, err := credentialFromOctopRoot(root, func(string) string { return "" }); err == nil || token != "" {
			t.Fatal("ambiguous or missing selection accepted")
		}
	}
	if _, err := credentialFromOctopRoot(root, func(key string) string {
		if key == "OCTOP_DATABASE_URL" {
			return "unsupported"
		}
		return ""
	}); err == nil {
		t.Fatal("database override accepted")
	}
	clientFixtureFile(t, root, "config.json", `{"database":{"driver":"postgresql"}}`)
	if _, err := credentialFromOctopRoot(root, func(string) string { return "" }); err == nil {
		t.Fatal("unsupported storage accepted")
	}
	missing := filepath.Join(clientFixtureRoot(t), "octop.db")
	if _, err := credentialFromOctopDatabase(missing); err == nil {
		t.Fatal("missing database accepted")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("database was created")
	}
}

func TestClientFixturesAreSanitizedAndImmutable(t *testing.T) {
	root := clientFixtureRoot(t)
	path := clientFixtureFile(t, root, "config.yaml", hermesFixture)
	before, _ := os.ReadFile(path)
	_, _ = credentialFromHermesRoot(root)
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("configuration modified")
	}
	payload, _ := json.Marshal(taskReceipt(taskRequest{Model: "gpt-image-2(Sub)"}, "fixture-task", "pending"))
	if !strings.Contains(string(payload), "gpt-image-2(Sub)") {
		t.Fatal("model identity not preserved")
	}
}

func TestOctopRejectsUnboundedOrAmbiguousDeclarations(t *testing.T) {
	for _, declarations := range []string{
		`[{"id":"chat-model","id":"other"}]`,
		`[{"id":"chat-model","enabled":null}]`,
		`[{"id":"chat-model","enabled":0}]`,
		`[{"id":"chat-model","enabled":""}]`,
		`[{"id":"chat-model","embedding":1}]`,
		`[{"id":"chat-model","enabled":false}]`,
		`[{"id":"chat-model","embedding":true}]`,
		`[{"id":"chat-model","task":"embedding"}]`,
		`[{"id":""}]`,
		strings.Repeat(" ", maxConfigBytes+1),
	} {
		root, db := octopFixture(t, false)
		if _, err := db.Exec("UPDATE providers SET models_json = ?", declarations); err != nil {
			t.Fatal(err)
		}
		if token, err := credentialFromOctopDatabase(filepath.Join(root, "octop.db")); err == nil || token != "" {
			t.Fatal("invalid model declarations accepted")
		}
	}
	root, db := octopFixture(t, false)
	if _, err := db.Exec("UPDATE providers SET api_key = ?", strings.Repeat("x", 16385)); err != nil {
		t.Fatal(err)
	}
	if _, err := credentialFromOctopDatabase(filepath.Join(root, "octop.db")); err == nil {
		t.Fatal("unbounded credential accepted")
	}
	if _, err := db.Exec("INSERT INTO settings VALUES('active_model', ?)", strings.Repeat("x", 1025)); err != nil {
		t.Fatal(err)
	}
	if _, err := credentialFromOctopDatabase(filepath.Join(root, "octop.db")); err == nil {
		t.Fatal("unbounded selection accepted")
	}
}
