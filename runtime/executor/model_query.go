package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
)

type modelQuery struct {
	Kind       string         `json:"kind,omitempty"`
	Model      string         `json:"model,omitempty"`
	Operation  string         `json:"operation,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

var modelQueryIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,160}$`)

// input is nil for an unfiltered catalog read; callers should not pass an open
// terminal stream when no request file was supplied. Filters stay local: the
// only network operation is one GET to the fixed live catalog.
func executeModelQuery(output io.Writer, svc service, input io.Reader) error {
	query, err := decodeModelQuery(input)
	if err != nil {
		writeReceipt(output, validationFailure("Use an optional model filter object with kind, exact model, declared operation and parameters. No API request was sent."))
		return err
	}
	body, status, retry, code, message, err := svc.request(context.Background(), http.MethodGet, "/v1/media/models", nil, "")
	if err != nil || status < 200 || status >= 300 {
		writeReceipt(output, apiFailure("submission", status, retry, code, message, "The live model catalog could not be read. No media task was submitted."))
		return errors.New("model catalog unavailable")
	}
	catalog, err := readAPIObject(body)
	entries, valid := catalog["data"].([]any)
	if err != nil || !valid {
		writeReceipt(output, apiFailure("submission", status, retry, code, message, "The API did not return a readable model catalog. No media task was submitted."))
		return errors.New("invalid model catalog")
	}
	matched := []any{}
	for _, value := range entries {
		entry, ok := value.(map[string]any)
		if ok && modelQueryMatches(query, entry) {
			matched = append(matched, entry)
		}
	}
	// Reuse the established public projection: no raw upstream fields, prices,
	// provider identities or arbitrary catalog metadata enter the receipt.
	projected := projectCatalog(map[string]any{"data": matched}).(map[string]any)
	parameterNames := make([]string, 0, len(query.Parameters))
	for name := range query.Parameters {
		parameterNames = append(parameterNames, name)
	}
	sort.Strings(parameterNames)
	projected["matched_count"] = len(matched)
	projected["filter"] = map[string]any{
		"kind": query.Kind, "model": query.Model, "operation": query.Operation,
		"parameter_names": parameterNames,
	}
	projected["matching_scope"] = "declared_schema_only"
	projected["note"] = "Matches reflect only supplied filters and returned declarations. Required prompts or media, execution authorization, price and output quality are not verified."
	writeJSON(output, map[string]any{"ok": true, "command": "models", "result": projected})
	return nil
}

func decodeModelQuery(input io.Reader) (modelQuery, error) {
	var query modelQuery
	if input == nil {
		return query, nil
	}
	body, err := io.ReadAll(io.LimitReader(input, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return query, errors.New("model filter unreadable")
	}
	body = bytes.TrimSpace(bytes.TrimPrefix(body, []byte{0xef, 0xbb, 0xbf}))
	if len(body) == 0 {
		return query, nil
	}
	if body[0] != '{' {
		return query, errors.New("model filter must be an object")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&query); err != nil {
		return query, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return query, errors.New("model filter must be one object")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil {
		return query, errors.New("model filter must be an object")
	}
	for _, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return query, errors.New("model filter fields must not be null")
		}
	}
	if query.Kind != "" && query.Kind != "image" && query.Kind != "video" {
		return query, errors.New("unknown media kind")
	}
	if query.Model != "" && (!modelQueryIDPattern.MatchString(query.Model) || query.Model == "." || query.Model == ".." || safePublicString(query.Model) != query.Model) {
		return query, errors.New("invalid exact model")
	}
	if query.Operation != "" && safePublicCode(query.Operation) != query.Operation {
		return query, errors.New("invalid operation")
	}
	for name := range query.Parameters {
		if name == "" || safePublicCode(name) != name {
			return query, errors.New("invalid parameter name")
		}
	}
	return query, nil
}

func modelQueryMatches(query modelQuery, entry map[string]any) bool {
	id, _ := entry["id"].(string)
	if !modelQueryIDPattern.MatchString(id) || id == "." || id == ".." || safePublicString(id) != id || (query.Model != "" && query.Model != id) {
		return false
	}
	capabilities := array(entry["capabilities"])
	if !hasValue(capabilities, "image") && !hasValue(capabilities, "video") {
		return false
	}
	if query.Kind != "" && !hasValue(capabilities, query.Kind) {
		return false
	}
	// An unfiltered read still exposes the full available public schema, even
	// when a model has no parameter schema. A filter never invents a declaration.
	if query.Operation == "" && len(query.Parameters) == 0 {
		return true
	}
	schemaValue, exists := entry["input_schema"]
	if !exists {
		return false
	}
	body, err := json.Marshal(schemaValue)
	var schema parameterSchema
	if err != nil || json.Unmarshal(body, &schema) != nil {
		return false
	}
	if query.Operation != "" {
		operation, exists := schema.Operations[query.Operation]
		if !exists || operation.Request.Method != http.MethodPost ||
			(operation.Request.ContentType != "application/json" && operation.Request.ContentType != "multipart/form-data") {
			return false
		}
		routeMatches := false
		for _, capability := range []string{"image", "video"} {
			if (query.Kind == "" || query.Kind == capability) && hasValue(capabilities, capability) && allowedMediaPath(capability, operation.Request.Path) {
				routeMatches = true
			}
		}
		if !routeMatches {
			return false
		}
	}
	for name, value := range query.Parameters {
		rule, declared := schema.Properties[name]
		if !declared || validateProperty(name, value, rule) != nil {
			return false
		}
		if _, unsupported := schema.Constraints["unsupported_inputs"][name]; unsupported {
			return false
		}
		if transports, declared := schema.Constraints["reference_transport"][name]; declared && validateReferences(value, transports) != nil {
			return false
		}
		if companions, declared := schema.Constraints["requires_together"][name]; declared {
			for _, companion := range array(companions) {
				key, ok := companion.(string)
				if !ok {
					return false
				}
				if _, provided := query.Parameters[key]; !provided {
					return false
				}
			}
		}
	}
	mode := "text"
	if query.Parameters["image"] != nil || query.Parameters["first_frame_image"] != nil || query.Operation == "image_to_video" {
		mode = "image"
	}
	if query.Operation == "reference_image_video" || query.Operation == "reference_video" || query.Operation == "reference_audio" {
		mode = "reference"
	}
	if allowed, declared := schema.Constraints["resolution_by_mode"][mode]; declared && query.Parameters["resolution"] != nil && !hasValue(array(allowed), query.Parameters["resolution"]) {
		return false
	}
	return true
}
