package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func kimiQoderRoot(host, home string, getenv func(string) string) (string, error) {
	if host != "kimi-code" && host != "qoder" && host != "pi" {
		return "", errors.New("unsupported host")
	}
	if !filepath.IsAbs(home) {
		return "", errors.New("host directory unavailable")
	}
	if host == "pi" {
		root := getenv("PI_CODING_AGENT_DIR")
		if root == "" {
			root = filepath.Join(home, ".pi", "agent")
		}
		if !filepath.IsAbs(root) {
			return "", errors.New("host directory unavailable")
		}
		for _, part := range strings.Split(filepath.ToSlash(root), "/") {
			if part == ".." {
				return "", errors.New("host directory unavailable")
			}
		}
		return root, nil
	}
	if host == "kimi-code" {
		root := getenv("KIMI_CODE_HOME")
		if root == "" {
			root = filepath.Join(home, ".kimi-code")
		}
		if !filepath.IsAbs(root) {
			return "", errors.New("host directory unavailable")
		}
		return root, nil
	}
	if root := getenv("QODER_CONFIG_DIR"); root != "" {
		if !filepath.IsAbs(root) {
			return "", errors.New("host directory unavailable")
		}
		return root, nil
	}
	name := getenv("QODER_CONFIG_DIR_NAME")
	if name == "" {
		name = ".qoder"
	}
	if name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return "", errors.New("host directory unavailable")
	}
	base := getenv("QODER_CLI_HOME")
	if base == "" {
		base = home
	}
	if !filepath.IsAbs(base) {
		return "", errors.New("host directory unavailable")
	}
	return filepath.Join(base, name), nil
}

func readPiObject(path string) (map[string]any, error) {
	data, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	decoder := json.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})))
	value, err := decodeUniqueConfigValue(decoder, 0)
	if err != nil || jsonObject(value) == nil {
		return nil, credentialFailure("active_connection_format_unsupported", "The Pi connection record could not be interpreted; no API request was sent.", "")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, credentialFailure("active_connection_format_unsupported", "The Pi connection record is unreadable.", "")
	}
	return jsonObject(value), nil
}

func credentialFromPiFile(path string) (string, error) {
	document, err := readPiObject(path)
	if err != nil {
		return "", err
	}
	providers := jsonObject(document["providers"])
	var p map[string]any
	var providerID string
	for id, raw := range providers {
		candidate := jsonObject(raw)
		if !matchesPureTokensEndpoint(jsonString(candidate["baseUrl"]), "/v1", "/v1/") {
			continue
		}
		if p != nil {
			return "", credentialFailure("active_connection_ambiguous", "The Pi adapter cannot safely select one matching saved connection.", "")
		}
		p = candidate
		providerID = id
	}
	if p == nil {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Pi adapter could not recognize a matching saved connection; no API request was sent.", "")
	}
	if jsonString(p["api"]) != "openai-completions" {
		return "", credentialFailure("active_connection_format_unsupported", "The Pi Pure Tokens provider uses an unsupported API format; no API request was sent.", "")
	}
	// The ID links this endpoint-verified entry to its higher-priority auth
	// record. It is never used as evidence of the service's identity.
	auth, err := readPiObject(filepath.Join(filepath.Dir(path), "auth.json"))
	if err != nil && !os.IsNotExist(err) {
		return "", credentialFailure("active_connection_format_unsupported", "The Pi effective authentication record could not be interpreted; no API request was sent.", "")
	}
	if value, exists := auth[providerID]; exists {
		entry := jsonObject(value)
		if jsonString(entry["type"]) != "api_key" {
			return "", credentialFailure("active_connection_format_unsupported", "The Pi effective authentication format is not supported; no fallback credential was used.", "")
		}
		return piInlineCredential(jsonString(p["baseUrl"]), jsonString(entry["key"]))
	}
	if piBuiltinEnvironmentAuth(providerID) {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Pi effective authentication may override the saved inline value; no fallback credential was used.", "")
	}
	return piInlineCredential(jsonString(p["baseUrl"]), jsonString(p["apiKey"]))
}

func piInlineCredential(endpoint, token string) (string, error) {
	if strings.Contains(token, "$") || strings.HasPrefix(strings.TrimSpace(token), "!") ||
		strings.IndexFunc(token, unicode.IsControl) >= 0 {
		return "", credentialFailure("active_connection_format_unsupported", "Pi dynamic or escaped credential values are not supported; no API request was sent.", "")
	}
	return inlineHostCredential(endpoint, token)
}

// Pi 71dca871: packages/ai/src/env-api-keys.ts. These IDs can resolve
// authentication before models.json. Refuse that fallback without reading env.
func piBuiltinEnvironmentAuth(id string) bool {
	switch id {
	case "github-copilot", "anthropic", "amazon-bedrock",
		"ant-ling", "qwen-token-plan", "qwen-token-plan-cn", "qwen-token-plan-individual",
		"openai", "google", "groq", "cerebras", "xai", "openrouter",
		"vercel-ai-gateway", "zai", "zai-coding-cn", "mistral", "minimax", "minimax-cn",
		"huggingface", "opencode", "opencode-go", "kimi-coding",
		"moonshotai", "moonshotai-cn", "nvidia", "deepseek", "radius", "baseten",
		"xiaomi", "together", "fireworks", "xiaomi-token-plan-cn",
		"xiaomi-token-plan-ams", "xiaomi-token-plan-sgp", "google-vertex",
		"azure-openai-responses", "cloudflare-workers-ai", "cloudflare-ai-gateway":
		return true
	}
	return false
}

func credentialFromNewHost(host string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := kimiQoderRoot(host, home, os.Getenv)
	if err != nil {
		return "", err
	}
	if host == "pi" {
		return credentialFromPiFile(filepath.Join(root, "models.json"))
	}
	if host == "kimi-code" {
		return credentialFromKimiFile(filepath.Join(root, "config.toml"))
	}
	return credentialFromQoderFile(filepath.Join(root, "settings.json"))
}

func inlineHostCredential(endpoint, token string) (string, error) {
	if strings.ContainsAny(token, "\r\n\x00") || strings.Contains(token, "${") || strings.Contains(token, "{env:") {
		return "", credentialFailure("active_connection_credential_missing", "The host connection has no supported inline credential.", "")
	}
	return matchingCredential(endpoint, token, "", "/", "/v1", "/v1/")
}

func credentialFromKimiFile(path string) (string, error) {
	doc, err := readTomlConfig(path)
	if err != nil {
		return "", err
	}
	model := tomlValue(doc, nil, "default_model")
	if model == "" {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Kimi Code adapter could not resolve the declared model selection.", "")
	}
	provider := tomlValue(doc, []string{"models", model}, "provider")
	if provider == "" {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Kimi Code adapter could not resolve the declared connection selection.", "")
	}
	table := []string{"providers", provider}
	endpoint := tomlValue(doc, table, "base_url")
	if !matchesPureTokensEndpoint(endpoint, "", "/", "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1")
	}
	switch tomlValue(doc, table, "type") {
	case "openai", "openai_responses", "anthropic":
	default:
		return "", credentialFailure("active_connection_format_unsupported", "The adapter cannot use the declared authentication format.", "")
	}
	return inlineHostCredential(endpoint, tomlValue(doc, table, "api_key"))
}

func credentialFromQoderFile(path string) (string, error) {
	data, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return "", err
	}
	defer clear(data)
	decoder := json.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})))
	decoder.UseNumber()
	value, err := decodeUniqueConfigValue(decoder, 0)
	if err != nil {
		return "", errors.New("host configuration unreadable")
	}
	if _, err = decoder.Token(); err != io.EOF {
		return "", errors.New("host configuration unreadable")
	}
	providers := jsonObject(jsonObject(value)["providers"])
	if providers == nil {
		return "", credentialFailure("active_connection_format_unsupported", "The Qoder adapter cannot interpret the saved connection record format.", "")
	}
	if len(providers) == 0 {
		return "", credentialFailure("active_connection_selection_unconfirmed", "The Qoder adapter found no connection entries in its supported saved record.", "")
	}
	var selected map[string]any
	var rejection error
	rejectionStatus := ""
	for _, raw := range providers {
		p := jsonObject(raw)
		if !matchesPureTokensEndpoint(jsonString(p["baseUrl"]), "", "/", "/v1", "/v1/") {
			candidate := endpointRecognitionFailure(jsonString(p["baseUrl"]), "", "/", "/v1", "/v1/")
			status, _, _ := credentialFailureDetails(candidate)
			if rejection == nil {
				rejection, rejectionStatus = candidate, status
			} else if status != rejectionStatus {
				rejection = credentialFailure("active_connection_selection_unconfirmed", "The Qoder adapter could not resolve a matching connection from the supported saved entries.", "")
				rejectionStatus = "active_connection_selection_unconfirmed"
			}
			continue
		}
		if selected != nil {
			return "", credentialFailure("active_connection_ambiguous", "The host has multiple matching Pure Tokens connections.", "")
		}
		selected = p
	}
	if selected == nil {
		return "", rejection
	}
	if jsonString(selected["type"]) != "openai-compatible" || jsonString(selected["protocol"]) != "openai" {
		return "", credentialFailure("active_connection_format_unsupported", "The Qoder adapter cannot use the declared authentication format.", "")
	}
	return inlineHostCredential(jsonString(selected["baseUrl"]), jsonString(selected["apiKey"]))
}
