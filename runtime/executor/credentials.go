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

type credentialResolutionError struct {
	status     string
	message    string
	nextAction string
}

func (err *credentialResolutionError) Error() string {
	return err.status
}

func credentialFailure(status, message, nextAction string) error {
	return &credentialResolutionError{status: status, message: message, nextAction: nextAction}
}

func credentialFailureDetails(err error) (string, string, string) {
	var resolution *credentialResolutionError
	if errors.As(err, &resolution) {
		return resolution.status, resolution.message, resolution.nextAction
	}
	return "active_connection_unavailable", "The active host connection record could not be read.", "Open the host connection settings, apply the Pure Tokens connection again, then run init again."
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
		return "", credentialFailure("active_connection_unavailable", "Codex has no active configured connection.", "Apply the Pure Tokens connection in Codex, then run init again.")
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
		return "", err
	}
	items := jsonArray(value)
	if items == nil {
		items = jsonArray(jsonObject(value)["models"])
	}
	if len(items) == 0 {
		return "", credentialFailure("active_connection_unavailable", "WorkBuddy does not have an active configured connection for this check.", "Select and apply the Pure Tokens connection in WorkBuddy, then run init again.")
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
			return "", credentialFailure("active_connection_credential_missing", "The configured Pure Tokens record has no usable credential.", "Apply the Pure Tokens connection in WorkBuddy again, then run init.")
		}
		return "", credentialFailure("active_connection_not_puretokens", "No supported Pure Tokens connection record was found.", "Apply the Pure Tokens connection in WorkBuddy, then run init.")
	}
	if len(keys) != 1 {
		return "", credentialFailure("active_connection_ambiguous", "WorkBuddy has no single unambiguous Pure Tokens credential for this check.", "Keep one active Pure Tokens connection in WorkBuddy, then run init again.")
	}
	for key := range keys {
		return key, nil
	}
	return "", credentialFailure("active_connection_credential_missing", "The active WorkBuddy Pure Tokens connection has no usable credential.", "Apply the Pure Tokens connection in WorkBuddy again, then run init again.")
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
		return "", credentialFailure("active_connection_unavailable", "Grok Build does not have an active configured connection for this check.", "Select and apply the Pure Tokens connection in Grok Build, then run init again.")
	}
	table := []string{"model", active}
	return matchingCredential(tomlValue(document, table, "base_url"), tomlValue(document, table, "api_key"), "/v1", "/v1/")
}

func credentialFromOpenCode() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return credentialFromOpenCodeFile(filepath.Join(home, ".config", "opencode", "opencode.json"))
}

func credentialFromOpenCodeFile(configPath string) (string, error) {
	document, err := readJSONObject(configPath)
	if err != nil {
		return "", err
	}
	model := jsonString(document["model"])
	providerID, _, found := strings.Cut(model, "/")
	if !found || providerID == "" {
		return "", credentialFailure("active_connection_unavailable", "OpenCode does not have an active configured connection for this check.", "Select and apply the Pure Tokens connection in OpenCode, then run init again.")
	}
	provider := jsonObject(jsonObject(document["provider"])[providerID])
	options := jsonObject(provider["options"])
	return matchingCredential(jsonString(options["baseURL"]), jsonString(options["apiKey"]), "/v1", "/v1/")
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
		return "", credentialFailure("active_connection_not_puretokens", "The active host connection does not target the Pure Tokens API.", "Select or apply the Pure Tokens connection in the host, then run init again.")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", credentialFailure("active_connection_credential_missing", "The active Pure Tokens connection has no usable credential.", "Apply the Pure Tokens connection in the host again, then run init again.")
	}
	return token, nil
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
