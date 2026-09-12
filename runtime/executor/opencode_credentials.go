package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/tailscale/hujson"
)

func opencodeUnsupported() error {
	return credentialFailure("active_connection_format_unsupported", "The OpenCode adapter cannot verify this configuration or authentication form; no API request was sent.", "")
}

func opencodeUnconfirmed() error {
	return credentialFailure("active_connection_selection_unconfirmed", "The OpenCode adapter could not confirm a declared service connection; no API request was sent.", "")
}

func readOpenCodeObject(path string) (map[string]any, error) {
	return readOpenCodeJSON(path, true)
}

func readOpenCodeJSON(path string, jsonc bool) (map[string]any, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > maxConfigBytes {
		return nil, opencodeUnsupported()
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return nil, opencodeUnsupported()
	}
	data, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	defer clear(data)
	if err != nil || len(data) > maxConfigBytes {
		return nil, opencodeUnsupported()
	}
	return parseOpenCodeObject(data, jsonc)
}

func parseOpenCodeObject(data []byte, jsonc bool) (map[string]any, error) {
	if len(data) > maxConfigBytes || !openCodeDepthBounded(data) ||
		bytes.Contains(data, []byte("{env:")) || bytes.Contains(data, []byte("{file:")) || bytes.Contains(data, []byte("${")) {
		return nil, opencodeUnsupported()
	}
	standard := bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	var err error
	if jsonc {
		standard, err = hujson.Standardize(standard)
	}
	defer clear(standard)
	if err != nil {
		return nil, opencodeUnsupported()
	}
	decoder := json.NewDecoder(bytes.NewReader(standard))
	decoder.UseNumber()
	value, err := decodeUniqueConfigValue(decoder, 0)
	if err != nil || jsonObject(value) == nil || openCodeReferences(value) {
		return nil, opencodeUnsupported()
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, opencodeUnsupported()
	}
	return jsonObject(value), nil
}

func openCodeReferences(value any) bool {
	switch value := value.(type) {
	case string:
		return strings.Contains(value, "{env:") || strings.Contains(value, "{file:") || strings.Contains(value, "${")
	case map[string]any:
		for key, child := range value {
			if openCodeReferences(key) || openCodeReferences(child) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if openCodeReferences(child) {
				return true
			}
		}
	}
	return false
}

// Bound recursion before HuJSON builds its syntax tree. This is only a resource
// guard; HuJSON and the unique-key JSON decoder still validate the grammar.
func openCodeDepthBounded(data []byte) bool {
	depth := 0
	var state byte
	for i := 0; i < len(data); i++ {
		c := data[i]
		switch state {
		case '"':
			if c == '\\' {
				i++
			} else if c == '"' {
				state = 0
			}
		case '/':
			if c == '\n' || c == '\r' {
				state = 0
			}
		case '*':
			if c == '*' && i+1 < len(data) && data[i+1] == '/' {
				i++
				state = 0
			}
		default:
			switch c {
			case '"':
				state = '"'
			case '/':
				if i+1 < len(data) && (data[i+1] == '/' || data[i+1] == '*') {
					i++
					state = data[i]
				}
			case '{', '[':
				depth++
				if depth > 64 {
					return false
				}
			case '}', ']':
				depth--
				if depth < 0 {
					return false
				}
			}
		}
	}
	return true
}

func opencodeAbsolute(path string) bool {
	if !filepath.IsAbs(path) || strings.IndexFunc(path, unicode.IsControl) >= 0 {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

func opencodeXDG(home, variable, fallback string, getenv func(string) string) (string, error) {
	root := getenv(variable)
	if root == "" {
		root = filepath.Join(home, fallback)
	}
	if !opencodeAbsolute(root) {
		return "", opencodeUnsupported()
	}
	return filepath.Join(root, "opencode"), nil
}

// Only named OpenCode records are considered; never enumerate the home or
// consult another host, a session database, or recent-model history.
func credentialFromOpenCodeAt(home, cwd string, getenv func(string) string) (string, error) {
	environment := opencodeEnvironment{home: home, cwd: cwd, getenv: getenv}
	switch runtime.GOOS {
	case "darwin":
		environment.managed = "/Library/Application Support/opencode"
		identity, err := user.Current()
		if err != nil {
			return "", opencodeUnsupported()
		}
		environment.preferences = []string{
			filepath.Join("/Library/Managed Preferences", identity.Username, "ai.opencode.managed.plist"),
			"/Library/Managed Preferences/ai.opencode.managed.plist",
		}
	case "windows":
		base := getenv("ProgramData")
		if base == "" {
			base = `C:\ProgramData`
		}
		if !opencodeAbsolute(base) {
			return "", opencodeUnsupported()
		}
		environment.managed = filepath.Join(base, "opencode")
	default:
		environment.managed = "/etc/opencode"
	}
	return credentialFromOpenCodeEnvironment(environment)
}

type opencodeEnvironment struct {
	home, cwd, managed string
	preferences        []string
	getenv             func(string) string
}

func credentialFromOpenCodeEnvironment(environment opencodeEnvironment) (string, error) {
	home, cwd, getenv := environment.home, environment.cwd, environment.getenv
	if !opencodeAbsolute(home) || !opencodeAbsolute(cwd) {
		return "", opencodeUnsupported()
	}
	for _, name := range []string{"OPENCODE_CONFIG_CONTENT", "OPENCODE_AUTH_CONTENT", "OPENCODE_TEST_HOME", "OPENCODE_TEST_MANAGED_CONFIG_DIR", "OPENCODE_CONSOLE_TOKEN"} {
		if getenv(name) != "" {
			return "", opencodeUnsupported()
		}
	}
	for _, path := range environment.preferences {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return "", opencodeUnsupported()
		}
	}
	configRoot, err := opencodeXDG(home, "XDG_CONFIG_HOME", ".config", getenv)
	if err != nil {
		return "", err
	}
	dataRoot, err := opencodeXDG(home, "XDG_DATA_HOME", filepath.Join(".local", "share"), getenv)
	if err != nil {
		return "", err
	}
	for _, name := range []string{"OPENCODE_CONFIG", "OPENCODE_CONFIG_DIR"} {
		if value := getenv(name); value != "" && !opencodeAbsolute(value) {
			return "", opencodeUnsupported()
		}
	}
	// A legacy record requires the host's migration, not an invented merge here.
	if _, err := os.Lstat(filepath.Join(configRoot, "config")); !os.IsNotExist(err) {
		return "", opencodeUnsupported()
	}
	document := map[string]any{}
	totalBytes := int64(0)
	load := func(path string, required bool) error {
		info, err := os.Lstat(path)
		if os.IsNotExist(err) && !required {
			return nil
		}
		if err != nil {
			return err
		}
		totalBytes += info.Size()
		if totalBytes > 4*maxConfigBytes {
			return opencodeUnsupported()
		}
		next, err := readOpenCodeObject(path)
		if err != nil {
			return err
		}
		if raw, exists := next["plugin"]; exists {
			plugins, ok := raw.([]any)
			if !ok || len(plugins) != 0 {
				return opencodeUnsupported()
			}
		}
		mergeOpenCodeConfig(document, next)
		return nil
	}
	pair := func(dir string) error {
		for _, name := range []string{"opencode.json", "opencode.jsonc"} {
			if err := load(filepath.Join(dir, name), false); err != nil {
				return err
			}
		}
		return nil
	}
	if err := load(filepath.Join(configRoot, "config.json"), false); err != nil {
		return "", err
	}
	if err := pair(configRoot); err != nil {
		return "", err
	}
	if custom := getenv("OPENCODE_CONFIG"); custom != "" {
		if err := load(custom, true); err != nil {
			return "", err
		}
	}
	var directories []string
	disabled := getenv("OPENCODE_DISABLE_PROJECT_CONFIG")
	if disabled != "" && disabled != "false" && disabled != "0" && disabled != "true" && disabled != "1" {
		return "", opencodeUnsupported()
	}
	if disabled != "true" && disabled != "1" {
		for dir := filepath.Clean(cwd); ; dir = filepath.Dir(dir) {
			directories = append(directories, dir)
			if len(directories) > 64 {
				return "", opencodeUnsupported()
			}
			if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
				break
			} else if !os.IsNotExist(err) {
				return "", err
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
		// ConfigPaths.files reverses its child-first JSONC/JSON list.
		for i := len(directories) - 1; i >= 0; i-- {
			if err := pair(directories[i]); err != nil {
				return "", err
			}
		}
	}
	// ConfigPaths.directories retains child-first .opencode precedence.
	roots := []string{configRoot}
	for _, dir := range directories {
		roots = append(roots, filepath.Join(dir, ".opencode"))
	}
	roots = append(roots, filepath.Join(home, ".opencode"))
	if custom := getenv("OPENCODE_CONFIG_DIR"); custom != "" {
		roots = append(roots, filepath.Clean(custom))
	}
	seen := map[string]bool{}
	for _, dir := range roots {
		if seen[dir] {
			continue
		}
		seen[dir] = true
		for _, name := range []string{"plugin", "plugins"} {
			if _, err := os.Lstat(filepath.Join(dir, name)); !os.IsNotExist(err) {
				return "", opencodeUnsupported()
			}
		}
		if filepath.Base(dir) == ".opencode" || dir == filepath.Clean(getenv("OPENCODE_CONFIG_DIR")) {
			if err := pair(dir); err != nil {
				return "", err
			}
		}
	}
	if environment.managed != "" {
		if err := pair(environment.managed); err != nil {
			return "", err
		}
	}
	return credentialFromOpenCodeDocument(document, filepath.Join(dataRoot, "auth.json"))
}

func mergeOpenCodeConfig(dst, src map[string]any) {
	for key, value := range src {
		if object, previous := jsonObject(value), jsonObject(dst[key]); object != nil && previous != nil {
			mergeOpenCodeConfig(previous, object)
		} else {
			dst[key] = value
		}
	}
}

func opencodeStatic(value string) bool {
	return !strings.Contains(value, "{env:") && !strings.Contains(value, "{file:") &&
		!strings.Contains(value, "${") && strings.IndexFunc(value, unicode.IsControl) < 0
}

func opencodeAllowed(document map[string]any, id string) (bool, error) {
	allowed := true
	for _, field := range []string{"enabled_providers", "disabled_providers"} {
		raw, exists := document[field]
		if !exists {
			continue
		}
		values, ok := raw.([]any)
		if !ok {
			return false, opencodeUnsupported()
		}
		found := false
		for _, value := range values {
			entry, ok := value.(string)
			if !ok || entry == "" || !opencodeStatic(entry) {
				return false, opencodeUnsupported()
			}
			found = found || entry == id
		}
		if field == "enabled_providers" {
			allowed = allowed && found
		} else {
			allowed = allowed && !found
		}
	}
	return allowed, nil
}

func credentialFromOpenCodeDocument(document map[string]any, authPath string) (string, error) {
	providers := jsonObject(document["provider"])
	if len(providers) == 0 {
		return "", opencodeUnconfirmed()
	}
	// External plugins can alter authentication/transport. Do not assume that
	// a static record represents those hooks or expand credential references.
	if raw, exists := document["plugin"]; exists {
		plugins, ok := raw.([]any)
		if !ok || len(plugins) != 0 {
			return "", opencodeUnsupported()
		}
	}
	var id string
	var selected map[string]any
	if raw, exists := document["model"]; exists {
		model, ok := raw.(string)
		providerID, modelID, qualified := strings.Cut(model, "/")
		if !ok || !qualified || providerID == "" || modelID == "" || !opencodeStatic(model) {
			return "", opencodeUnconfirmed()
		}
		id, selected = providerID, jsonObject(providers[providerID])
		if selected == nil {
			return "", opencodeUnconfirmed()
		}
		allowed, err := opencodeAllowed(document, id)
		if err != nil {
			return "", err
		}
		if !allowed {
			return "", opencodeUnconfirmed()
		}
	} else {
		for candidateID, value := range providers {
			if candidateID == "" || !opencodeStatic(candidateID) {
				return "", opencodeUnsupported()
			}
			allowed, err := opencodeAllowed(document, candidateID)
			if err != nil {
				return "", err
			}
			if !allowed {
				continue
			}
			provider := jsonObject(value)
			options := jsonObject(provider["options"])
			endpoint := jsonString(options["baseURL"])
			if !opencodeStatic(endpoint) {
				return "", opencodeUnsupported()
			}
			if !matchesPureTokensEndpoint(endpoint, "/v1", "/v1/") {
				continue
			}
			if selected != nil {
				return "", credentialFailure("active_connection_ambiguous", "OpenCode has multiple matching saved service connections and no verified selection.", "")
			}
			id, selected = candidateID, provider
		}
		if selected == nil {
			return "", opencodeUnconfirmed()
		}
	}
	options := jsonObject(selected["options"])
	endpoint := jsonString(options["baseURL"])
	if !opencodeStatic(endpoint) {
		return "", opencodeUnsupported()
	}
	if !matchesPureTokensEndpoint(endpoint, "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1", "/v1/")
	}
	if raw, exists := options["headers"]; exists && (jsonObject(raw) == nil || len(jsonObject(raw)) != 0) {
		return "", opencodeUnsupported()
	}
	for _, raw := range jsonObject(selected["models"]) {
		if headers, exists := jsonObject(raw)["headers"]; exists && (jsonObject(headers) == nil || len(jsonObject(headers)) != 0) {
			return "", opencodeUnsupported()
		}
	}
	inline := func(raw any) (string, error) {
		token, ok := raw.(string)
		if !ok || !opencodeStatic(token) {
			return "", opencodeUnsupported()
		}
		return matchingCredential(endpoint, token, "/v1", "/v1/")
	}
	var auth map[string]any
	if authPath != "" {
		var err error
		auth, err = readOpenCodeJSON(authPath, false)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		for _, raw := range auth {
			if jsonString(jsonObject(raw)["type"]) == "wellknown" {
				return "", opencodeUnsupported()
			}
		}
	}
	entry := jsonObject(auth[id])
	if _, exists := auth[id]; exists && jsonString(entry["type"]) != "api" {
		return "", opencodeUnsupported()
	}
	// Inline provider options override the native API-key store, including an
	// explicitly empty value. Never try lower-priority credentials after failure.
	if token, exists := options["apiKey"]; exists {
		return inline(token)
	}
	if entry == nil {
		return "", credentialFailure("active_connection_credential_missing", "The OpenCode connection has no supported stored credential.", "")
	}
	return inline(entry["key"])
}
