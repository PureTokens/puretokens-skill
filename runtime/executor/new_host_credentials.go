package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func kimiQoderRoot(host, home string, getenv func(string) string) (string, error) {
	if host != "kimi-code" && host != "qoder" {
		return "", errors.New("unsupported host")
	}
	if !filepath.IsAbs(home) {
		return "", errors.New("host directory unavailable")
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

func credentialFromNewHost(host string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := kimiQoderRoot(host, home, os.Getenv)
	if err != nil {
		return "", err
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
