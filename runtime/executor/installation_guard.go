package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const inventoryFile = ".puretokens-managed.json"
const inventoryFormat = "puretokens-managed-files-v1"

type installationInventory struct {
	Format string            `json:"format"`
	Name   string            `json:"name"`
	Files  map[string]string `json:"files"`
}

// Purely local, read-only installer support. The checksum-verified candidate
// binary checks ownership and emits an inventory; only sync writes installed
// files/markers. These commands never resolve credentials or make requests.
func runInstallationGuard(args []string, output io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	directory := flags.String("directory", "", "")
	name := flags.String("name", "", "")
	source := flags.String("source-directory", "", "")
	fail := func() error {
		fmt.Fprintln(output, "Installation ownership could not be verified, or files were added/modified. Existing files were preserved; review the directory before updating.")
		return errors.New("installation ownership conflict")
	}
	if flags.Parse(args[1:]) != nil || flags.NArg() != 0 || !validInstallationName(*name) {
		return fail()
	}
	actual, err := installationTree(*directory, *name)
	if err != nil {
		return fail()
	}
	if args[0] == "install-inventory" {
		return json.NewEncoder(output).Encode(actual)
	}
	if err := verifyInstallation(*directory, actual, *source); err != nil {
		return fail()
	}
	return nil
}

func validInstallationName(name string) bool {
	switch name {
	case ".puretokens-executor", "puretokens-balance", "puretokens-connection", "puretokens-image", "puretokens-video", "puretokens-models", "puretokens-update":
		return true
	}
	return false
}

func installationTree(directory, name string) (installationInventory, error) {
	result := installationInventory{Format: inventoryFormat, Name: name, Files: map[string]string{}}
	info, err := os.Lstat(directory)
	if !filepath.IsAbs(directory) || err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return result, errors.New("invalid installation directory")
	}
	var total int64
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == directory {
			return nil
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if len(result.Files) >= 2048 || strings.ContainsAny(relative, "\r\n") || entry.Type()&os.ModeSymlink != 0 {
			return errors.New("unsupported installation entry")
		}
		if entry.IsDir() {
			result.Files[relative+"/"] = "directory"
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("unsupported installation file")
		}
		if relative == inventoryFile {
			return nil
		}
		total += info.Size()
		if info.Size() > 64<<20 || total > 128<<20 {
			return errors.New("installation inventory too large")
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		opened, err := file.Stat()
		if err != nil || !os.SameFile(info, opened) {
			return errors.New("installation changed during inspection")
		}
		hash := sha256.New()
		if relative == "runtime.json" {
			data, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
			var object map[string]any
			if err != nil || len(data) > maxConfigBytes || json.Unmarshal(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}), &object) != nil || object == nil {
				return errors.New("invalid runtime manifest")
			}
			encoded, _ := json.Marshal(object)
			hash.Write(encoded)
		} else {
			n, err := io.Copy(hash, io.LimitReader(file, info.Size()+1))
			if err != nil || n != info.Size() {
				return errors.New("installation changed during inspection")
			}
		}
		result.Files[relative] = fmt.Sprintf("%x", hash.Sum(nil))
		return nil
	})
	return result, err
}

func verifyInstallation(directory string, actual installationInventory, source string) error {
	marker := filepath.Join(directory, inventoryFile)
	if _, err := os.Lstat(marker); err == nil {
		body, err := readBoundedFile(marker, maxConfigBytes)
		var recorded installationInventory
		if err != nil || json.Unmarshal(body, &recorded) != nil || recorded.Format != inventoryFormat || recorded.Name != actual.Name || !reflect.DeepEqual(actual.Files, recorded.Files) {
			return errors.New("managed files changed")
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	// Exact current source copies can be adopted without relying on names.
	if source != "" {
		expected, err := installationTree(source, actual.Name)
		if err == nil && reflect.DeepEqual(expected.Files, actual.Files) {
			return nil
		}
	}
	return errors.New("installation provenance unknown")
}
