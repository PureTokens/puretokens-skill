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

// ZCode's data-base override is the parent of .zcode, not .zcode itself.
// This is shared by credential and diagnostic path selection; no discovery or
// read of ZCode settings, session databases or other clients is performed.
func zcodeRoot(home, dataBase string) (string, error) {
	base := strings.TrimSpace(dataBase)
	if base == "" {
		base = home
	}
	if !filepath.IsAbs(base) {
		return "", credentialFailure("active_connection_unavailable", "The ZCode data directory is unavailable.", "Run this Skill in ZCode's configured execution environment, then run init again.")
	}
	return filepath.Join(base, ".zcode"), nil
}

func credentialFromZCode() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := zcodeRoot(home, os.Getenv("ZCODE_DATA_BASE_DIR"))
	if err != nil {
		return "", err
	}
	return credentialFromZCodeFile(filepath.Join(root, "v2", "config.json"))
}

func credentialFromZCodeFile(path string) (string, error) {
	data, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return "", credentialFailure("active_connection_unavailable", "The ZCode connection record is unavailable.", "Open ZCode to finish any legacy migration and apply the Pure Tokens connection, then run init again.")
	}
	defer clear(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeZCodeValue(decoder, 0)
	if err != nil {
		return "", zcodeUnreadableConfig()
	}
	if _, err := decoder.Token(); err != io.EOF {
		return "", zcodeUnreadableConfig()
	}
	document := jsonObject(value)
	providers := jsonObject(document["provider"])
	if document == nil || providers == nil {
		return "", zcodeUnreadableConfig()
	}

	// v2/config.json describes available connections. ZCode delivers the
	// conversation's model separately at runtime; root model/model.main and
	// provider labels are not authoritative selection signals in this store.
	// Resolve only a unique enabled matching connection, as a configured
	// Pure Tokens API capability, never as proof of the chat model's identity.
	var selected map[string]any
	for _, entry := range providers {
		provider := jsonObject(entry)
		if provider["enabled"] != true {
			continue
		}
		options := jsonObject(provider["options"])
		if !matchesPureTokensEndpoint(jsonString(options["baseURL"]), "/v1", "/v1/") {
			continue
		}
		if selected != nil {
			return "", credentialFailure("active_connection_ambiguous", "ZCode has more than one enabled Pure Tokens connection for this request.", "Keep one enabled Pure Tokens connection in ZCode for this Skill, then run init again.")
		}
		selected = provider
	}
	if selected == nil {
		return "", credentialFailure("active_connection_not_puretokens", "ZCode has no enabled matching Pure Tokens connection for this request.", "Enable the Pure Tokens connection in ZCode, then run init again.")
	}
	if jsonString(selected["kind"]) != "openai-compatible" {
		return "", credentialFailure("active_connection_unavailable", "The enabled ZCode connection uses an unsupported authentication format.", "Apply the supported Pure Tokens connection in ZCode, then run init again.")
	}
	options := jsonObject(selected["options"])
	token := jsonString(options["apiKey"])
	// Only inline bearer credentials are supported. Do not interpret template
	// references, external stores, or arbitrary process environment variables.
	if strings.ContainsAny(token, "\r\n\x00") || strings.Contains(token, "${") || strings.Contains(token, "{env:") {
		return "", credentialFailure("active_connection_credential_missing", "The enabled Pure Tokens connection has no supported inline credential.", "Apply the Pure Tokens connection in ZCode, then run init again.")
	}
	return matchingCredential(jsonString(options["baseURL"]), token, "/v1", "/v1/")
}

func zcodeUnreadableConfig() error {
	return credentialFailure("active_connection_unavailable", "The ZCode connection record is unreadable or ambiguous.", "Repair the connection through ZCode's settings, then run init again.")
}

// Duplicate JSON keys would otherwise silently choose the last connection or
// credential. Reject them (including escaped aliases) before projecting fields.
func decodeZCodeValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, errors.New("configuration nesting limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, compound := token.(json.Delim)
	if !compound {
		return token, nil
	}
	switch delim {
	case '{':
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, errors.New("invalid object key")
			}
			if _, exists := object[key]; exists {
				return nil, errors.New("duplicate object key")
			}
			value, err := decodeZCodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return nil, errors.New("invalid object")
		}
		return object, nil
	case '[':
		array := []any{}
		for decoder.More() {
			value, err := decodeZCodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return nil, errors.New("invalid array")
		}
		return array, nil
	default:
		return nil, errors.New("unexpected delimiter")
	}
}
