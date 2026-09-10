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
		return "", credentialFailure("active_connection_credential_missing", "The host connection has no supported inline credential.", "Apply the supported Pure Tokens connection in the host, then run init again.")
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
		return "", errors.New("active model unavailable")
	}
	provider := tomlValue(doc, []string{"models", model}, "provider")
	if provider == "" {
		return "", errors.New("active connection unavailable")
	}
	table := []string{"providers", provider}
	endpoint := tomlValue(doc, table, "base_url")
	if !matchesPureTokensEndpoint(endpoint, "", "/", "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1")
	}
	switch tomlValue(doc, table, "type") {
	case "openai", "openai_responses", "anthropic":
	default:
		return "", errors.New("unsupported authentication format")
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
		return "", errors.New("host configuration unavailable")
	}
	var selected map[string]any
	for _, raw := range providers {
		p := jsonObject(raw)
		if !matchesPureTokensEndpoint(jsonString(p["baseUrl"]), "", "/", "/v1", "/v1/") {
			continue
		}
		if selected != nil {
			return "", credentialFailure("active_connection_ambiguous", "The host has multiple matching Pure Tokens connections.", "Keep one matching connection for this Skill, then run init again.")
		}
		selected = p
	}
	if selected == nil {
		return "", credentialFailure("active_connection_not_puretokens", "The host has no matching Pure Tokens connection.", "Apply the Pure Tokens connection in the host, then run init again.")
	}
	if jsonString(selected["type"]) != "openai-compatible" || jsonString(selected["protocol"]) != "openai" {
		return "", errors.New("unsupported authentication format")
	}
	return inlineHostCredential(jsonString(selected["baseUrl"]), jsonString(selected["apiKey"]))
}
