package main

// Credential adapters read only the active connection records written by a
// supported host or by Pure Tokens Switch for that host. They never use the
// configured URL as the API target: requests use apiOrigin, or balanceOrigin
// for the documented balance-only console API endpoints.

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const pureTokensHost = "api.puretokensx.com"

const connectionDiagnosticNextAction = "Stop this automatic API flow. Keep the bound host; do not try another --host or invoke connection, init, doctor, models, or another Skill as a recovery probe. Preserve the original request and any existing task ID; never automatically resubmit. Only an explicit recheck request or verified new diagnostic evidence permits one relevant check on the same host. Changing host requires explicit user selection. Keep existing connections. Report the host version and this safe diagnostic code for adapter compatibility checks; do not share configuration contents or credentials. This result does not prove that the host was never configured. Do not reinstall, switch the chat model, or replace credentials solely because of this check."

type credentialResolutionError struct {
	status     string
	message    string
	nextAction string
}

func (err *credentialResolutionError) Error() string {
	return err.status
}

func credentialFailure(status, message, nextAction string) error {
	// Local selection/reading failures cannot establish that a user has never
	// configured the host, or justify replacing a working connection.
	if strings.HasPrefix(status, "active_connection_") || strings.HasPrefix(status, "workbuddy_") || status == "host_credential_adapter_unavailable" {
		nextAction = connectionDiagnosticNextAction
	}
	return &credentialResolutionError{status: status, message: message, nextAction: nextAction}
}

func credentialFailureDetails(err error) (string, string, string) {
	var resolution *credentialResolutionError
	if errors.As(err, &resolution) {
		return resolution.status, resolution.message, resolution.nextAction
	}
	if os.IsNotExist(err) {
		return "active_connection_record_missing", "The adapter's declared connection record was not found in this execution environment; no API request was sent.", connectionDiagnosticNextAction
	}
	return "active_connection_unavailable", "The adapter could not read or interpret the declared effective connection; no API request was sent.", connectionDiagnosticNextAction
}

func credentialFromCodex() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	directory, err := configuredDirectory(home, ".codex", "CODEX_HOME")
	if err != nil {
		return "", err
	}
	return credentialFromCodexFile(filepath.Join(directory, "config.toml"))
}

func credentialFromCodexFile(configPath string) (string, error) {
	document, err := readTomlConfig(configPath)
	if err != nil {
		return "", err
	}
	active := tomlValue(document, nil, "model_provider")
	if profile := tomlValue(document, nil, "profile"); profile != "" {
		if selected := tomlValue(document, []string{"profiles", profile}, "model_provider"); selected != "" {
			active = selected
		}
	}
	if active == "" {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Codex adapter could not resolve a connection from the declared selection.", "")
	}
	table := []string{"model_providers", active}
	endpoint := tomlValue(document, table, "base_url")
	// Check the effective provider BEFORE touching its declared credential source.
	if !matchesPureTokensEndpoint(endpoint, "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1", "/v1/")
	}
	token := tomlValue(document, table, "experimental_bearer_token")
	if token == "" {
		if envKey := tomlValue(document, table, "env_key"); envKey != "" {
			token = os.Getenv(envKey) // Only the exact variable selected by this provider.
		} else if tomlTable(document, table)["requires_openai_auth"] == true {
			auth, err := readJSONObject(filepath.Join(filepath.Dir(configPath), "auth.json"))
			if err != nil {
				return "", err
			}
			token = jsonString(auth["OPENAI_API_KEY"])
		}
	}
	return matchingCredential(endpoint, token, "/v1", "/v1/")
}

func credentialFromClaudeCode() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	directory, err := configuredDirectory(home, ".claude", "CLAUDE_CONFIG_DIR")
	if err != nil {
		return "", err
	}
	return credentialFromClaudeCodeFile(filepath.Join(directory, "settings.json"))
}

func credentialFromClaudeCodeFile(configPath string) (string, error) {
	document, err := readJSONObject(configPath)
	if err != nil {
		return "", err
	}
	env := jsonObject(document["env"])
	return matchingCredential(jsonString(env["ANTHROPIC_BASE_URL"]), jsonString(env["ANTHROPIC_AUTH_TOKEN"]), "", "/", "/v1", "/v1/")
}

func credentialFromGeminiCLI() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return credentialFromGeminiEnvFile(filepath.Join(home, ".gemini", ".env"))
}

func credentialFromGeminiEnvFile(configPath string) (string, error) {
	values, err := readEnvConfig(configPath)
	if err != nil {
		return "", err
	}
	return matchingCredential(values["GOOGLE_GEMINI_BASE_URL"], values["GEMINI_API_KEY"], "", "/", "/v1", "/v1/")
}

func credentialFromWorkBuddy() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	directory, err := configuredDirectory(home, ".workbuddy", "WORKBUDDY_CONFIG_DIR", "CODEBUDDY_CONFIG_DIR")
	if err != nil {
		return "", err
	}
	return credentialFromWorkBuddyFile(filepath.Join(directory, "models.json"))
}

func credentialFromWorkBuddyFile(configPath string) (string, error) {
	value, err := readJSONValue(configPath)
	if err != nil {
		status := "workbuddy_record_unreadable"
		if os.IsNotExist(err) {
			status = "workbuddy_record_missing"
		}
		return "", credentialFailure(status, "The supported WorkBuddy connection record could not be read; no API request was sent.", "Installation and connection readiness are separate. If Pure Tokens was already configured, report the WorkBuddy version and this diagnostic code without sharing configuration contents. The current chat model need not be changed.")
	}
	items, supported := value.([]any)
	if !supported {
		items, supported = jsonObject(value)["models"].([]any)
	}
	if !supported {
		return "", credentialFailure("workbuddy_record_format_unsupported", "The WorkBuddy connection record format is not supported.", "")
	}
	if len(items) == 0 {
		return "", credentialFailure("workbuddy_connection_not_found", "No connection entries were found in the supported WorkBuddy record.", "")
	}
	keys := make(map[string]struct{})
	matchedEndpoint := false
	for _, item := range items {
		record := jsonObject(item)
		if matchesPureTokensEndpoint(jsonString(record["url"]), "/v1/chat/completions") {
			matchedEndpoint = true
		}
		key, err := matchingCredential(jsonString(record["url"]), jsonString(record["apiKey"]), "/v1/chat/completions")
		if err == nil {
			keys[key] = struct{}{}
		}
	}
	if len(keys) == 0 {
		if matchedEndpoint {
			return "", credentialFailure("workbuddy_credential_missing", "The configured Pure Tokens record has no usable credential.", "")
		}
		return "", credentialFailure("workbuddy_connection_not_found", "The WorkBuddy adapter could not confirm a matching Pure Tokens connection in its supported record format.", "")
	}
	if len(keys) != 1 {
		return "", credentialFailure("workbuddy_connection_ambiguous", "WorkBuddy has no single unambiguous Pure Tokens credential for this check.", "")
	}
	for key := range keys {
		return key, nil
	}
	return "", credentialFailure("workbuddy_credential_missing", "The active WorkBuddy Pure Tokens connection has no usable credential.", "")
}

func credentialFromGrokBuild() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return credentialFromGrokBuildFile(filepath.Join(home, ".grok", "config.toml"))
}

func credentialFromGrokBuildFile(configPath string) (string, error) {
	document, err := readTomlConfig(configPath)
	if err != nil {
		return "", err
	}
	active := tomlValue(document, []string{"models"}, "default")
	if active == "" {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Grok Build adapter could not resolve a connection from the declared selection.", "")
	}
	table := []string{"model", active}
	return matchingCredential(tomlValue(document, table, "base_url"), tomlValue(document, table, "api_key"), "/v1", "/v1/")
}

func credentialFromOpenCode() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return credentialFromOpenCodeAt(home, cwd, os.Getenv)
}

func credentialFromOpenCodeFile(configPath string) (string, error) {
	document, err := readOpenCodeObject(configPath)
	if err != nil {
		return "", err
	}
	return credentialFromOpenCodeDocument(document, "")
}

// configuredDirectory accepts only the documented host configuration-directory
// override. It deliberately does not consult a generic environment key or
// search candidate folders.
func configuredDirectory(home string, defaultDirectory string, overrideNames ...string) (string, error) {
	for _, name := range overrideNames {
		value, exists := os.LookupEnv(name)
		if !exists || strings.TrimSpace(value) == "" {
			continue
		}
		path := filepath.Clean(strings.TrimSpace(value))
		if !filepath.IsAbs(path) {
			return "", errors.New("host configuration directory is invalid")
		}
		return path, nil
	}
	return filepath.Join(home, defaultDirectory), nil
}

func matchingCredential(endpoint, token string, allowedPaths ...string) (string, error) {
	if !matchesPureTokensEndpoint(endpoint, allowedPaths...) {
		return "", endpointRecognitionFailure(endpoint, allowedPaths...)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", credentialFailure("active_connection_credential_missing", "The active Pure Tokens connection has no usable credential.", "")
	}
	return token, nil
}

func endpointRecognitionFailure(endpoint string, allowedPaths ...string) error {
	message := "The adapter could not recognize the declared connection under the fixed Pure Tokens endpoint rules; no credential authentication or API request was performed."
	if strings.TrimSpace(endpoint) == "" {
		return credentialFailure("active_connection_endpoint_missing", "The adapter could not resolve an endpoint from the supported record fields; no API request was sent.", "")
	}
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err == nil && strings.EqualFold(parsed.Hostname(), pureTokensHost) && !matchesPureTokensEndpoint(endpoint, allowedPaths...) {
		return credentialFailure("active_connection_endpoint_unsupported", "The declared endpoint uses an unsupported form for this adapter; no credential authentication or API request was performed.", "")
	}
	// Keep the existing code for consumers, but do not turn an allowlist
	// mismatch into a claim about the user's configuration or key validity.
	return credentialFailure("active_connection_not_puretokens", message, "")
}

func matchesPureTokensEndpoint(value string, allowedPaths ...string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || !strings.EqualFold(parsed.Hostname(), pureTokensHost) || (parsed.Port() != "" && parsed.Port() != "443") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	for _, path := range allowedPaths {
		if parsed.EscapedPath() == path {
			return true
		}
	}
	return false
}

func readJSONValue(path string) (any, error) {
	bytes, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return nil, err
	}
	var value any
	if err := json.Unmarshal(bytes, &value); err != nil {
		return nil, errors.New("host configuration is unreadable")
	}
	return value, nil
}

func readJSONObject(path string) (map[string]any, error) {
	value, err := readJSONValue(path)
	if err != nil {
		return nil, err
	}
	document := jsonObject(value)
	if document == nil {
		return nil, errors.New("host configuration is unreadable")
	}
	return document, nil
}

func jsonObject(value any) map[string]any {
	document, _ := value.(map[string]any)
	return document
}

func jsonArray(value any) []any {
	items, _ := value.([]any)
	return items
}

func jsonString(value any) string {
	stringValue, _ := value.(string)
	return stringValue
}

func readEnvConfig(path string) (map[string]string, error) {
	bytes, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	for _, rawLine := range strings.Split(string(bytes), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		name, rawValue, found := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		if !found || name == "" {
			continue
		}
		value := strings.TrimSpace(rawValue)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			unquoted, err := strconv.Unquote(value)
			if err != nil {
				return nil, errors.New("host configuration is unreadable")
			}
			value = unquoted
		}
		values[name] = value
	}
	return values, nil
}

type tomlConfig map[string]any

func readTomlConfig(path string) (tomlConfig, error) {
	data, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return nil, err
	}
	return parseTomlConfig(string(data))
}

func parseTomlConfig(source string) (tomlConfig, error) {
	var document tomlConfig
	if err := toml.Unmarshal([]byte(source), &document); err != nil {
		return nil, errors.New("host configuration is unreadable")
	}
	return document, nil
}

func tomlTable(document tomlConfig, path []string) map[string]any {
	table := map[string]any(document)
	for _, name := range path {
		table, _ = table[name].(map[string]any)
	}
	return table
}

func tomlValue(document tomlConfig, table []string, key string) string {
	value, _ := tomlTable(document, table)[key].(string)
	return value
}
