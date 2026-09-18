package main

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func credentialFromOctopRoot(root string, getenv func(string) string) (string, error) {
	for _, key := range []string{"OCTOP_DATABASE_URL", "OCTOP_DATABASE_DRIVER", "OCTOP_DATABASE_SQLITE_PATH", "OCTOP_DATABASE_HOST", "OCTOP_DATABASE_PORT", "OCTOP_DATABASE_NAME", "OCTOP_DATABASE_USER", "OCTOP_DATABASE_PASSWORD"} {
		if getenv(key) != "" {
			return "", clientFormatFailure()
		}
	}
	config, err := readClientJSON(filepath.Join(root, "config.json"), false)
	if err != nil && !os.IsNotExist(err) {
		return "", clientFormatFailure()
	}
	if err == nil {
		document := jsonObject(config)
		if document == nil {
			return "", clientFormatFailure()
		}
		if value, exists := document["database"]; exists {
			database := jsonObject(value)
			if database == nil {
				return "", clientFormatFailure()
			}
			for _, key := range []string{"driver", "sqlite_path", "url"} {
				if value, exists := database[key]; exists {
					text, ok := value.(string)
					if !ok || (key == "driver" && text != "sqlite") ||
						(key == "sqlite_path" && text != "octop.db") || (key == "url" && text != "") {
						return "", clientFormatFailure()
					}
				}
			}
		}
	}
	return credentialFromOctopDatabase(filepath.Join(root, "octop.db"))
}

func checkOctopDatabasePath(path string) (os.FileInfo, error) {
	if err := checkClientFile(path); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil || !singleLinkFile(path, before) {
		return nil, clientFormatFailure()
	}
	for _, suffix := range []string{"-wal", "-shm", "-journal"} {
		companion := path + suffix
		info, err := os.Lstat(companion)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil || !info.Mode().IsRegular() || !singleLinkFile(companion, info) || checkClientFile(companion) != nil {
			return nil, clientFormatFailure()
		}
	}
	return before, nil
}

func credentialFromOctopDatabase(path string) (string, error) {
	before, err := checkOctopDatabasePath(path)
	if err != nil {
		return "", clientFormatFailure()
	}
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	if !strings.HasPrefix(uri.Path, "/") {
		uri.Path = "/" + uri.Path
	}
	query := url.Values{"mode": {"ro"}, "_pragma": {"query_only(1)", "trusted_schema(0)", "busy_timeout(250)"}}
	uri.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return "", clientFormatFailure()
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// A read transaction includes committed WAL frames. immutable=1 would
	// silently ignore a live host's latest connection selection.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", clientFormatFailure()
	}
	defer tx.Rollback()
	for _, table := range []string{"providers", "settings"} {
		var kind, definition string
		if tx.QueryRowContext(ctx, "SELECT type, CASE WHEN length(CAST(sql AS BLOB)) <= 16384 THEN sql END FROM sqlite_schema WHERE name = ?", table).Scan(&kind, &definition) != nil ||
			kind != "table" || !strings.HasPrefix(strings.ToUpper(definition), "CREATE TABLE") {
			return "", clientFormatFailure()
		}
	}
	var active string
	err = tx.QueryRowContext(ctx, "SELECT CASE WHEN length(CAST(value AS BLOB)) <= 1024 THEN value END FROM settings WHERE key = 'active_model'").Scan(&active)
	if err != nil && err != sql.ErrNoRows {
		return "", clientFormatFailure()
	}
	if len(active) > 1024 {
		return "", clientFormatFailure()
	}
	selectedName, model, hasSelection := strings.Cut(active, "/")
	if active != "" && (!hasSelection || selectedName == "" || model == "") {
		return "", clientSelectionFailure()
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,
CASE WHEN length(CAST(name AS BLOB)) <= 1024 THEN name END,
CASE WHEN length(CAST(kind AS BLOB)) <= 64 THEN kind END,
CASE WHEN length(CAST(base_url AS BLOB)) <= 4096 THEN base_url END, enabled,
CASE WHEN length(CAST(models_json AS BLOB)) <= ? THEN models_json END
FROM providers LIMIT 129`, maxConfigBytes)
	if err != nil {
		return "", clientFormatFailure()
	}
	var selectedID int64
	var endpoint string
	found, count := false, 0
	for rows.Next() {
		count++
		var id int64
		var name, kind string
		var base, models sql.NullString
		var enabled int
		if rows.Scan(&id, &name, &kind, &base, &enabled, &models) != nil || count > 128 || len(models.String) > maxConfigBytes {
			rows.Close()
			return "", clientFormatFailure()
		}
		if active != "" && name != selectedName {
			continue
		}
		if enabled != 1 || !matchesPureTokensEndpoint(base.String, "", "/", "/v1", "/v1/") {
			if active != "" {
				rows.Close()
				return "", clientSelectionFailure()
			}
			continue
		}
		if found {
			rows.Close()
			return "", credentialFailure("active_connection_ambiguous", "Octop has multiple matching saved connections; no API request was sent.", "")
		}
		value, parseErr := parseClientJSON([]byte(models.String), false)
		declarations, isArray := value.([]any)
		if kind != "openai" || parseErr != nil || !isArray || len(declarations) == 0 || len(declarations) > 160 {
			rows.Close()
			return "", clientFormatFailure()
		}
		eligible := false
		for _, raw := range declarations {
			declaration := jsonObject(raw)
			if declaration == nil || jsonString(declaration["id"]) == "" {
				rows.Close()
				return "", clientFormatFailure()
			}
			enabled, embedding := true, false
			for key, target := range map[string]*bool{"enabled": &enabled, "embedding": &embedding} {
				if value, exists := declaration[key]; exists {
					flag, ok := value.(bool)
					if !ok {
						rows.Close()
						return "", clientFormatFailure()
					}
					*target = flag
				}
			}
			if (model == "" || jsonString(declaration["id"]) == model) && enabled && !embedding &&
				!strings.EqualFold(jsonString(declaration["task"]), "embedding") {
				eligible = true
			}
		}
		if !eligible {
			rows.Close()
			return "", clientSelectionFailure()
		}
		selectedID, endpoint, found = id, base.String, true
	}
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil || !found {
		return "", clientSelectionFailure()
	}
	// Only the endpoint-verified row's credential is projected, never a dump.
	var token sql.NullString
	if tx.QueryRowContext(ctx, "SELECT CASE WHEN length(CAST(api_key AS BLOB)) <= 16384 THEN api_key END FROM providers WHERE id = ?", selectedID).Scan(&token) != nil || !token.Valid {
		return "", clientFormatFailure()
	}
	after, err := checkOctopDatabasePath(path)
	if err != nil || !os.SameFile(before, after) {
		return "", clientFormatFailure()
	}
	return strictInlineCredential(endpoint, token.String)
}
