package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Official Desktop only, independently of the community DSH Desktop layout.
// A saved default is not evidence of a session/model/CLI override. The Skill
// host binding restricts this adapter to the default local Desktop composition.
func credentialFromDeepSeekHarness() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	root, err := deepSeekHarnessRoot(home, os.Getenv)
	if err != nil {
		return "", err
	}
	return credentialFromDeepSeekHarnessRoot(root, os.Getenv)
}

var harnessCredentialRef = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func deepSeekHarnessRoot(home string, getenv func(string) string) (string, error) {
	root := getenv("DSH_HOME")
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(home, ".dsh")
	}
	if !clientAbsoluteRoot(root) {
		return "", desktopSelectionFailure()
	}
	for _, part := range strings.FieldsFunc(root, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == "." || part == ".." {
			return "", desktopSelectionFailure()
		}
	}
	return root, nil
}

func credentialFromDeepSeekHarnessRoot(root string, getenv func(string) string) (string, error) {
	// This higher-priority composition can change both the selected provider
	// and the credential service. Do not merge arbitrary plugin patches.
	homePatch, err := readHarnessPatch(filepath.Join(root, "cordis.patch.yml"))
	if err != nil && !os.IsNotExist(err) {
		return "", desktopSelectionFailure()
	}
	if len(homePatch) != 0 {
		return "", desktopSelectionFailure()
	}
	rows, err := readHarnessPatch(filepath.Join(root, "profiles", "desktop", "cordis.patch.yml"))
	if err != nil {
		return "", desktopSelectionFailure()
	}
	configs := make(map[string]map[string]any)
	for _, item := range rows {
		row := jsonObject(item)
		id := jsonString(row["id"])
		if id == "" {
			return "", desktopSelectionFailure()
		}
		if id != "llm-pi-ai" && id != "agent-default-model" {
			// Arbitrary plugin composition can replace credential/route services.
			return "", desktopSelectionFailure()
		}
		if configs[id] != nil || len(row) != 2 || jsonObject(row["config"]) == nil {
			return "", desktopSelectionFailure()
		}
		configs[id] = jsonObject(row["config"])
	}
	selected := configs["agent-default-model"]
	id, model := jsonString(selected["provider"]), jsonString(selected["model"])
	providers := jsonObject(configs["llm-pi-ai"]["providers"])
	provider := jsonObject(providers[id])
	if id == "" || model == "" || provider == nil {
		return "", desktopSelectionFailure()
	}
	endpoint := jsonString(provider["baseURL"])
	if !matchesPureTokensEndpoint(endpoint, "/v1", "/v1/") {
		return matchingCredential(endpoint, "", "/v1", "/v1/")
	}
	for otherID, item := range providers {
		if otherID != id && matchesPureTokensEndpoint(jsonString(jsonObject(item)["baseURL"]), "/v1", "/v1/") {
			return "", desktopSelectionFailure()
		}
	}
	ref := jsonString(provider["apiKeyEnv"])
	if jsonString(provider["api"]) != "openai-completions" || !harnessCredentialRef.MatchString(ref) ||
		getenv(ref) != "" ||
		hasClientOverride(provider, "apiKey", "headers", "auth", "authentication", "modelOverrides", "baseUrl", "base_url") {
		return "", desktopSelectionFailure()
	}
	found := false
	seen := make(map[string]bool)
	for _, item := range jsonArray(provider["models"]) {
		entry := jsonObject(item)
		modelID := jsonString(entry["id"])
		if modelID == "" || seen[modelID] ||
			hasClientOverride(entry, "baseURL", "baseUrl", "base_url", "api", "apiKey", "apiKeyEnv", "headers", "auth") {
			return "", desktopSelectionFailure()
		}
		seen[modelID] = true
		found = found || modelID == model
	}
	if !found {
		return "", desktopSelectionFailure()
	}
	// Only after endpoint and selection validation, read the exact declared
	// reference in the versioned store. Never use ambient/.env fallback.
	credentials, err := readDesktopYAML(filepath.Join(root, ".credentials.yaml"))
	if err != nil || credentials["version"] != 1 {
		return "", desktopSelectionFailure()
	}
	for key := range credentials {
		if key != "version" && key != "refs" && key != "records" {
			return "", desktopSelectionFailure()
		}
	}
	return matchingCredential(endpoint, jsonString(jsonObject(credentials["refs"])[ref]), "/v1", "/v1/")
}

func readHarnessPatch(path string) ([]any, error) {
	data, err := readBoundedFile(path, maxConfigBytes)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	err = decoder.Decode(&node)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil || !safeDesktopYAML(&node, 0) {
		return nil, errors.New("unsupported desktop composition")
	}
	var extra yaml.Node
	if decoder.Decode(&extra) != io.EOF {
		return nil, errors.New("unsupported desktop composition")
	}
	var rows []any
	if node.Decode(&rows) != nil {
		return nil, errors.New("unsupported desktop composition")
	}
	return rows, nil
}
