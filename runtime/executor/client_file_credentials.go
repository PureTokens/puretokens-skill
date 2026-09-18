package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/tailscale/hujson"
)

func clientFormatFailure() error {
	return credentialFailure("active_connection_format_unsupported", "The host's effective connection format or override could not be verified; no API request was sent.", "")
}

func clientSelectionFailure() error {
	return credentialFailure("active_connection_selection_unconfirmed", "The host's saved connection selection could not be confirmed; no API request was sent.", "")
}

func clientAbsoluteRoot(root string) bool {
	if !filepath.IsAbs(root) || strings.HasPrefix(root, `\\`) || strings.HasPrefix(root, "//") || strings.IndexFunc(root, unicode.IsControl) >= 0 {
		return false
	}
	for _, part := range strings.FieldsFunc(root, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return false
		}
	}
	return true
}

func newClientRoot(host, goos, home string, getenv func(string) string) (string, error) {
	if !clientAbsoluteRoot(home) {
		return "", clientFormatFailure()
	}
	var root string
	switch host {
	case "hermes":
		root = getenv("HERMES_HOME")
		if root == "" {
			root = filepath.Join(home, ".hermes")
			if goos == "windows" {
				local := getenv("LOCALAPPDATA")
				if !clientAbsoluteRoot(local) {
					return "", clientFormatFailure()
				}
				current := filepath.Join(local, "hermes")
				_, currentErr := os.Lstat(current)
				legacy, legacyErr := os.Lstat(root)
				if !os.IsNotExist(currentErr) || legacyErr != nil || !legacy.IsDir() {
					root = current
				}
			}
		}
	case "evox":
		root = getenv("EVOX_AGENT_DIR")
		other := getenv("EVOX_CODING_AGENT_DIR")
		if root != "" && other != "" && root != other {
			return "", clientFormatFailure()
		}
		if root == "" {
			root = other
		}
		if root == "" {
			root = filepath.Join(home, ".evox", "agent")
		}
	case "vscode":
		switch goos {
		case "darwin":
			root = filepath.Join(home, "Library", "Application Support", "Code", "User")
		case "windows":
			if !clientAbsoluteRoot(getenv("APPDATA")) {
				return "", clientFormatFailure()
			}
			root = filepath.Join(getenv("APPDATA"), "Code", "User")
		default:
			return "", clientFormatFailure()
		}
	case "octop":
		root = getenv("OCTOP_HOME")
		if root == "" {
			root = filepath.Join(home, ".octop")
		}
	default:
		return "", clientFormatFailure()
	}
	if !clientAbsoluteRoot(root) {
		return "", clientFormatFailure()
	}
	return root, nil
}

func checkClientFile(path string) error {
	if !clientAbsoluteRoot(path) {
		return clientFormatFailure()
	}
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if !clientPathSafe(current, info) || (current == path && !info.Mode().IsRegular()) || (current != path && !info.IsDir()) {
			return clientFormatFailure()
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}

func readClientBytes(path string) ([]byte, error) {
	if err := checkClientFile(path); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil || !singleLinkFile(path, before) {
		return nil, clientFormatFailure()
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, clientFormatFailure()
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) || opened.Size() > maxConfigBytes {
		return nil, clientFormatFailure()
	}
	data, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	after, statErr := os.Lstat(path)
	if err != nil || len(data) > maxConfigBytes || statErr != nil || checkClientFile(path) != nil ||
		!os.SameFile(opened, after) || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) ||
		!singleLinkFile(path, after) {
		clear(data)
		return nil, clientFormatFailure()
	}
	return data, nil
}

func readClientJSON(path string, jsonc bool) (any, error) {
	data, err := readClientBytes(path)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	return parseClientJSON(data, jsonc)
}

func parseClientJSON(data []byte, jsonc bool) (any, error) {
	var err error
	if !openCodeDepthBounded(data) {
		return nil, clientFormatFailure()
	}
	standard := bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	if jsonc {
		standard, err = hujson.Standardize(standard)
		if err != nil {
			return nil, clientFormatFailure()
		}
		defer clear(standard)
	}
	decoder := json.NewDecoder(bytes.NewReader(standard))
	decoder.UseNumber()
	value, err := decodeUniqueConfigValue(decoder, 0)
	if err != nil {
		return nil, clientFormatFailure()
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, clientFormatFailure()
	}
	return value, nil
}

func readClientYAML(path string) (map[string]any, error) {
	data, err := readClientBytes(path)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	document, err := parseDesktopYAML(data)
	if err != nil {
		return nil, clientFormatFailure()
	}
	return document, nil
}

func strictInlineCredential(endpoint, token string) (string, error) {
	if strings.ContainsAny(token, "$!") || strings.Contains(token, "{env:") || strings.Contains(token, "{file:") || strings.IndexFunc(token, unicode.IsControl) >= 0 {
		return "", clientFormatFailure()
	}
	return inlineHostCredential(endpoint, token)
}

func hasClientOverride(object map[string]any, fields ...string) bool {
	for _, field := range fields {
		if _, exists := object[field]; exists {
			return true
		}
	}
	return false
}

func credentialFromFileClient(host string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", clientSelectionFailure()
	}
	root, err := newClientRoot(host, runtime.GOOS, home, os.Getenv)
	if err != nil {
		return "", err
	}
	switch host {
	case "hermes":
		return credentialFromHermesRoot(root)
	case "evox":
		return credentialFromEvoXRoot(root)
	case "vscode":
		if os.Getenv("VSCODE_PORTABLE") != "" {
			return "", clientFormatFailure()
		}
		return credentialFromVSCodeFile(filepath.Join(root, "chatLanguageModels.json"))
	case "octop":
		return credentialFromOctopRoot(root, os.Getenv)
	}
	return "", clientFormatFailure()
}

func credentialFromHermesRoot(root string) (string, error) {
	document, err := readClientYAML(filepath.Join(root, "config.yaml"))
	if err != nil {
		return "", err
	}
	model := jsonObject(document["model"])
	id, custom := strings.CutPrefix(jsonString(model["provider"]), "custom:")
	provider := jsonObject(jsonObject(document["providers"])[id])
	if !custom || id == "" || jsonString(model["default"]) == "" || provider == nil || provider["enabled"] == false {
		return "", clientSelectionFailure()
	}
	endpoint := jsonString(provider["base_url"])
	if !matchesPureTokensEndpoint(endpoint, "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1", "/v1/")
	}
	if hasClientOverride(document, "provider", "base_url", "api_base", "custom_providers") ||
		hasClientOverride(model, "model", "base_url", "api_base", "api_key", "api_key_env", "key_env", "key_cmd", "api_mode", "transport") ||
		hasClientOverride(provider, "api", "url", "key_env", "api_key_env", "key_cmd", "extra_headers", "headers", "transport") ||
		jsonObject(document["model_aliases"])[jsonString(model["default"])] != nil {
		return "", clientFormatFailure()
	}
	switch jsonString(provider["api_mode"]) {
	case "chat_completions", "codex_responses", "anthropic_messages":
	default:
		return "", clientFormatFailure()
	}
	// Hermes pools outrank inline credentials. Pool selection/rotation is not
	// reproduced by this one-shot executor; never silently use a lower source.
	auth, err := readClientJSON(filepath.Join(root, "auth.json"), false)
	if err != nil && !os.IsNotExist(err) {
		return "", clientFormatFailure()
	}
	if err == nil {
		object := jsonObject(auth)
		if object == nil {
			return "", clientFormatFailure()
		}
		if pools, exists := object["credential_pool"]; exists {
			if jsonObject(pools) == nil || len(jsonObject(pools)) > 0 {
				return "", clientFormatFailure()
			}
		}
	}
	return strictInlineCredential(endpoint, jsonString(provider["api_key"]))
}

func credentialFromEvoXRoot(root string) (string, error) {
	for _, name := range []string{"models.json", "config.yaml", "auth.json"} {
		if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
			return "", clientFormatFailure()
		}
	}
	value, err := readClientJSON(filepath.Join(root, "settings.json"), false)
	if err != nil {
		return "", err
	}
	settings := jsonObject(value)
	id, model := jsonString(settings["defaultProvider"]), jsonString(settings["defaultModel"])
	preference := jsonObject(jsonObject(jsonObject(settings["config"])["modelPolicy"])["userPreference"])
	if id == "" || model == "" || jsonString(settings["defaultInstanceId"]) != "default" ||
		jsonString(preference["model"]) != id+"/"+model || jsonString(preference["instanceId"]) != "default" {
		return "", clientSelectionFailure()
	}
	provider := jsonObject(jsonObject(settings["providers"])[id])
	endpoint := jsonString(provider["baseUrl"])
	if !matchesPureTokensEndpoint(endpoint, "", "/", "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "", "/", "/v1", "/v1/")
	}
	if hasClientOverride(provider, "headers", "apiKey", "apiKeyEnv", "key_cmd") {
		return "", clientFormatFailure()
	}
	found := 0
	for _, entry := range jsonArray(provider["models"]) {
		entry := jsonObject(entry)
		if jsonString(entry["id"]) == model {
			if hasClientOverride(entry, "baseUrl", "headers", "apiKey", "apiKeyEnv", "key_cmd") ||
				(entry["api"] != nil && jsonString(entry["api"]) != jsonString(provider["api"])) {
				return "", clientFormatFailure()
			}
			found++
		}
	}
	if found != 1 {
		return "", clientSelectionFailure()
	}
	auth, err := readClientYAML(filepath.Join(root, "auth.yaml"))
	if err != nil {
		return "", err
	}
	entry := jsonObject(auth[id])
	if jsonString(entry["type"]) != "api_key" || jsonString(entry["base_url"]) != endpoint ||
		jsonString(entry["api"]) != jsonString(provider["api"]) {
		return "", clientFormatFailure()
	}
	switch jsonString(entry["api"]) {
	case "openai-completions", "openai-responses", "anthropic-messages":
	default:
		return "", clientFormatFailure()
	}
	return strictInlineCredential(endpoint, jsonString(entry["key"]))
}

func credentialFromVSCodeFile(path string) (string, error) {
	value, err := readClientJSON(path, true)
	if err != nil {
		return "", err
	}
	groups, ok := value.([]any)
	if !ok {
		return "", clientFormatFailure()
	}
	var selected map[string]any
	for _, raw := range groups {
		group := jsonObject(raw)
		for _, rawModel := range jsonArray(group["models"]) {
			model := jsonObject(rawModel)
			endpoint := jsonString(model["url"])
			if !matchesPureTokensEndpoint(endpoint, "/v1/chat/completions", "/v1/responses", "/v1/messages") {
				continue
			}
			if selected != nil {
				return "", credentialFailure("active_connection_ambiguous", "VS Code has multiple matching model connections without a verified effective selection; no API request was sent.", "")
			}
			if jsonString(group["vendor"]) != "customendpoint" || hasClientOverride(group, "apiKey", "requestHeaders") {
				return "", clientFormatFailure()
			}
			selected = model
		}
	}
	if selected == nil {
		return "", clientSelectionFailure()
	}
	headers := jsonObject(selected["requestHeaders"])
	if len(headers) != 1 {
		return "", clientFormatFailure()
	}
	token, ok := strings.CutPrefix(jsonString(headers["Authorization"]), "Bearer ")
	if !ok {
		return "", clientFormatFailure()
	}
	return strictInlineCredential(apiOrigin+"/v1", token)
}
