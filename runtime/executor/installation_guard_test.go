package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeInventoryFixture(t *testing.T, directory, name string) {
	t.Helper()
	var out bytes.Buffer
	if err := runInstallationGuard([]string{"install-inventory", "--directory", directory, "--name", name}, &out); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, inventoryFile), out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestInstallationGuardProtectsOwnedFilesAndUnmanagedNames(t *testing.T) {
	for _, change := range []string{"none", "extra", "modified", "deleted", "empty-directory", "marker", "symlink"} {
		t.Run(change, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("original"), 0600)
			os.WriteFile(filepath.Join(dir, "skill.json"), []byte(`{"name":"puretokens-image"}`), 0600)
			var out bytes.Buffer
			args := []string{"install-verify", "--directory", dir, "--name", "puretokens-image"}
			if runInstallationGuard(args, &out) == nil {
				t.Fatal("name alone established ownership")
			}
			writeInventoryFixture(t, dir, "puretokens-image")
			switch change {
			case "extra":
				os.WriteFile(filepath.Join(dir, "personal.txt"), []byte("keep"), 0600)
			case "modified":
				os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("local edit"), 0600)
			case "deleted":
				os.Remove(filepath.Join(dir, "SKILL.md"))
			case "empty-directory":
				os.Mkdir(filepath.Join(dir, "my-folder"), 0700)
			case "marker":
				os.WriteFile(filepath.Join(dir, inventoryFile), []byte(`{}`), 0600)
			case "symlink":
				if os.Symlink(filepath.Join(dir, "SKILL.md"), filepath.Join(dir, "link")) != nil {
					t.Skip("symlinks unavailable")
				}
			}
			out.Reset()
			err := runInstallationGuard(args, &out)
			if (err == nil) != (change == "none") {
				t.Fatalf("change %s: %v", change, err)
			}
		})
	}
}

func TestInstallationGuardAdoptsOnlyExactCurrentSource(t *testing.T) {
	source, target := t.TempDir(), t.TempDir()
	for _, dir := range []string{source, target} {
		os.Mkdir(filepath.Join(dir, "references"), 0700)
		os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("official"), 0600)
		os.WriteFile(filepath.Join(dir, "skill.json"), []byte(`{"name":"puretokens-image"}`), 0600)
	}
	var out bytes.Buffer
	args := []string{"install-verify", "--directory", target, "--name", "puretokens-image", "--source-directory", source}
	if err := runInstallationGuard(args, &out); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(target, "personal.txt"), []byte("keep"), 0600)
	if runInstallationGuard(args, &out) == nil {
		t.Fatal("source adoption ignored extra user file")
	}
}

func TestInstallationGuardAcceptsPowerShellRuntimeFormatting(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "runtime.json"), []byte(`{"schemaVersion":1,"name":"puretokens-api-executor","version":"0.17.1","platform":"windows-amd64"}`), 0600)
	writeInventoryFixture(t, dir, ".puretokens-executor")
	os.WriteFile(filepath.Join(dir, "runtime.json"), append([]byte{0xef, 0xbb, 0xbf}, []byte("{\r\n\"version\": \"0.17.1\", \"platform\":\"windows-amd64\", \"schemaVersion\":1,\"name\":\"puretokens-api-executor\"\r\n}")...), 0600)
	var out bytes.Buffer
	if err := runInstallationGuard([]string{"install-verify", "--directory", dir, "--name", ".puretokens-executor"}, &out); err != nil {
		t.Fatal(err)
	}
}

func TestInstallationHistoryHasExactFileInventories(t *testing.T) {
	var document struct {
		SchemaVersion int    `json:"schemaVersion"`
		SourceCommit  string `json:"sourceCommit"`
		Snapshots     []struct {
			Name     string            `json:"name"`
			Revision string            `json:"revision"`
			Files    map[string]string `json:"files"`
		} `json:"snapshots"`
	}
	if json.Unmarshal(installationHistory, &document) != nil || document.SchemaVersion != 1 || len(document.SourceCommit) != 40 || len(document.Snapshots) == 0 {
		t.Fatal("migration inventories unavailable")
	}
	for _, snapshot := range document.Snapshots {
		if !validInstallationName(snapshot.Name) || len(snapshot.Revision) != 40 || len(snapshot.Files) == 0 {
			t.Fatalf("invalid snapshot: %s", snapshot.Name)
		}
	}
}
