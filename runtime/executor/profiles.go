package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type modelProfile struct {
	ID         string          `json:"id"`
	Capability string          `json:"capability"`
	Parameters parameterSchema `json:"parameterSchema"`
}
type parameterSchema struct {
	Properties  map[string]map[string]any `json:"properties"`
	Constraints map[string]map[string]any `json:"constraints"`
	Operations  map[string]mediaOperation `json:"operations"`
}
type mediaOperation struct {
	Request struct {
		Method      string `json:"method"`
		Path        string `json:"path"`
		ContentType string `json:"contentType"`
	} `json:"request"`
	Required []string `json:"requiredBodyFields"`
	Inputs   map[string]struct {
		Field      string   `json:"field"`
		Required   bool     `json:"required"`
		Min        int      `json:"minItems"`
		Max        int      `json:"maxItems"`
		Transports []string `json:"transports"`
	} `json:"inputs"`
}

func loadProfile(request taskRequest, svc service) (modelProfile, error) {
	var profile modelProfile
	if !regexp.MustCompile(`^[A-Za-z0-9_.-]{1,160}$`).MatchString(request.Model) || request.Model == "." || request.Model == ".." {
		return profile, errors.New("Choose an exact model ID from the model index.")
	}
	if svc.profilesRoot != "" {
		file := filepath.Join(svc.profilesRoot, "puretokens-"+request.Kind, "references", "profiles", request.Model+".json")
		if body, err := readBoundedFile(file, maxResponseBytes); err == nil && json.Unmarshal(body, &profile) == nil && profile.ID == request.Model && profile.Capability == request.Kind {
			return profile, nil
		}
	}
	// Unknown exact IDs can still submit core text requests without a preflight.
	if request.Operation == "generate" && len(request.Parameters) == 0 && len(request.Attachments) == 0 && request.MediaOperation == "" && request.RequestedCount <= 1 {
		return modelProfile{ID: request.Model, Capability: request.Kind}, nil
	}
	body, status, _, _, _, err := svc.request(context.Background(), http.MethodGet, "/v1/media/models", nil, "")
	if err == nil && status == 200 {
		var catalog struct {
			Data []struct {
				ID           string          `json:"id"`
				Capabilities []string        `json:"capabilities"`
				InputSchema  parameterSchema `json:"input_schema"`
			} `json:"data"`
		}
		if json.Unmarshal(body, &catalog) == nil {
			for _, entry := range catalog.Data {
				if entry.ID == request.Model && contains(entry.Capabilities, request.Kind) {
					return modelProfile{ID: entry.ID, Capability: request.Kind, Parameters: entry.InputSchema}, nil
				}
			}
		}
	}
	return profile, errors.New("The selected model's required parameters or operation could not be confirmed. Ask for current model details; no media task was created.")
}
func prepareProfileRequest(request *taskRequest, svc service) error {
	if request.Parameters == nil {
		request.Parameters = make(map[string]any)
	}
	profile, err := loadProfile(*request, svc)
	if err != nil {
		return err
	}
	if rule, exists := profile.Parameters.Properties["prompt"]; exists {
		if err = validateProperty("prompt", request.Prompt, rule); err != nil {
			return err
		}
	}
	opName := request.MediaOperation
	if request.Operation == "edit" && opName == "" {
		opName = request.Kind + "_edit"
	}
	if request.Operation == "edit" && opName != request.Kind+"_edit" {
		return errors.New("The edit request must select the model's declared edit operation.")
	}
	if len(request.Attachments) > 0 && opName == "" {
		return errors.New("Specify the attachment role as media_operation, such as image_to_video or reference_image_video.")
	}
	var operation mediaOperation
	if opName != "" {
		var exists bool
		operation, exists = profile.Parameters.Operations[opName]
		if !exists {
			return errors.New("The selected model does not declare this media operation. Choose a declared operation before submitting.")
		}
		if operation.Request.Method != "POST" || !allowedMediaPath(request.Kind, operation.Request.Path) {
			return errors.New("The model operation has no supported fixed API path.")
		}
		request.route = operation.Request.Path
		if operation.Request.ContentType == "multipart/form-data" && len(request.Attachments) == 0 {
			return errors.New("This operation requires a current local attachment. For a public URL use the declared generation reference field instead.")
		}
		if len(request.Attachments) > 0 && operation.Request.ContentType != "multipart/form-data" {
			return errors.New("Local attachments require a declared multipart operation.")
		}
		if operation.Request.ContentType != "multipart/form-data" && operation.Request.ContentType != "application/json" {
			return errors.New("The model operation uses an unsupported request representation.")
		}
	}
	counts := make(map[string]int)
	for _, file := range request.Attachments {
		matched := false
		for _, input := range operation.Inputs {
			if input.Field == file.Field && contains(input.Transports, "multipart_file") {
				matched = true
				counts[file.Field]++
				if input.Max > 0 && counts[file.Field] > input.Max {
					return fmt.Errorf("The attachment count exceeds the declared maximum of %d.", input.Max)
				}
			}
		}
		if !matched {
			return errors.New("An attachment field or transport is not declared for this operation.")
		}
	}
	for key, value := range request.Parameters {
		if counts[key] > 0 {
			return errors.New("Do not send the same media field as both a file and a JSON parameter.")
		}
		rule, exists := profile.Parameters.Properties[key]
		if !exists {
			input, ok := operation.Inputs[key]
			if !ok || operation.Request.ContentType != "application/json" {
				return errors.New("An optional field is not declared by this model; check its supported parameters before submitting.")
			}
			if !contains(input.Transports, "public_https_url") {
				return errors.New("This JSON attachment transport is not supported by the model.")
			}
			if err := validateReferences(value, []any{"public_https_url"}); err != nil {
				return err
			}
			rule = map[string]any{"type": "string[]"}
			if input.Max > 0 {
				rule["maxLength"] = float64(input.Max)
			}
		}
		if err = validateProperty(key, value, rule); err != nil {
			return err
		}
		if transports, exists := profile.Parameters.Constraints["reference_transport"][key]; exists {
			if len(request.Attachments) > 0 {
				if _, declared := operation.Inputs[key]; !declared {
					return errors.New("Mixing URL references and local attachments requires an explicitly declared combination operation.")
				}
			}
			if err = validateReferences(value, transports); err != nil {
				return err
			}
		}
		if _, unsupported := profile.Parameters.Constraints["unsupported_inputs"][key]; unsupported {
			return errors.New("The selected model explicitly excludes this input.")
		}
	}
	for _, input := range operation.Inputs {
		if operation.Request.ContentType == "multipart/form-data" && input.Required && counts[input.Field] < max(1, input.Min) {
			return errors.New("A required current attachment is missing for this operation.")
		}
	}
	for key, rule := range profile.Parameters.Properties {
		if rule["required"] == true && key != "prompt" && key != "model" {
			if _, exists := request.Parameters[key]; !exists && counts[key] == 0 {
				return errors.New("A required model parameter is missing.")
			}
		}
	}
	for _, key := range operation.Required {
		if key != "prompt" && key != "model" && counts[key] == 0 {
			if _, exists := request.Parameters[key]; !exists {
				return errors.New("A required operation input is missing.")
			}
		}
	}
	for key, others := range profile.Parameters.Constraints["requires_together"] {
		if _, exists := request.Parameters[key]; exists {
			for _, other := range array(others) {
				if _, ok := request.Parameters[fmt.Sprint(other)]; !ok {
					return errors.New("Width and height must be supplied together as declared by the model.")
				}
			}
		}
	}
	// Deterministically apply declared size precedence, without guessing values.
	order := array(profile.Parameters.Constraints["size_expression_precedence"]["order"])
	chosen := ""
	for _, item := range order {
		expr := fmt.Sprint(item)
		fields := strings.Split(expr, "+")
		present := true
		for _, field := range fields {
			if _, ok := request.Parameters[field]; !ok {
				present = false
			}
		}
		if present && chosen == "" {
			chosen = expr
			continue
		}
		if chosen != "" {
			for _, field := range fields {
				delete(request.Parameters, field)
			}
		}
	}
	mode := "text"
	if counts["image"] > 0 || request.Parameters["image"] != nil || counts["first_frame_image"] > 0 {
		mode = "image"
	}
	if opName == "reference_image_video" || opName == "reference_video" || opName == "reference_audio" {
		mode = "reference"
	}
	if allowed, exists := profile.Parameters.Constraints["resolution_by_mode"][mode]; exists && request.Parameters["resolution"] != nil && !hasValue(array(allowed), request.Parameters["resolution"]) {
		return errors.New("This resolution is unavailable for the selected reference mode. Choose a declared resolution.")
	}
	count := 1
	if n, exists := request.Parameters["n"]; exists {
		value, ok := number(n)
		if !ok || value != math.Trunc(value) || value < 1 || value > 6 {
			return errors.New("Image count must be an integer from 1 through 6.")
		}
		count = int(value)
	} else if request.RequestedCount > 1 {
		rule, exists := profile.Parameters.Properties["n"]
		if !exists {
			return errors.New("The selected model does not declare multiple outputs.")
		}
		if err = validateProperty("n", float64(request.RequestedCount), rule); err != nil {
			return err
		}
		count = request.RequestedCount
		request.Parameters["n"] = count
	}
	if request.RequestedCount != 0 && request.RequestedCount != count {
		return errors.New("requested_count and n disagree; correct the count before submitting.")
	}
	if request.Kind == "video" && count != 1 {
		return errors.New("One video task returns one video.")
	}
	request.RequestedCount = count
	return nil
}
func allowedMediaPath(kind, path string) bool {
	return (kind == "image" && (path == "/v1/images/generations" || path == "/v1/images/edits")) || (kind == "video" && (path == "/v1/videos" || path == "/v1/videos/edits"))
}
func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
func array(value any) []any {
	if values, ok := value.([]any); ok {
		return values
	}
	if values, ok := value.([]string); ok {
		result := make([]any, len(values))
		for i, v := range values {
			result[i] = v
		}
		return result
	}
	return nil
}
func number(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		v, e := n.Float64()
		return v, e == nil
	}
	return 0, false
}
func hasValue(values []any, value any) bool {
	for _, v := range values {
		if reflect.DeepEqual(v, value) {
			return true
		}
		a, ok := number(v)
		b, ok2 := number(value)
		if ok && ok2 && a == b {
			return true
		}
	}
	return false
}
func validateProperty(name string, value any, rule map[string]any) error {
	invalid := func() error {
		return fmt.Errorf("Parameter %s does not match the model's declared type, range or supported values.", name)
	}
	var length int
	switch rule["type"] {
	case "integer", "number":
		n, ok := number(value)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || (rule["type"] == "integer" && n != math.Trunc(n)) {
			return invalid()
		}
		for _, key := range []string{"min", "minimum"} {
			if min, ok := number(rule[key]); ok && n < min {
				return invalid()
			}
		}
		for _, key := range []string{"max", "maximum"} {
			if max, ok := number(rule[key]); ok && n > max {
				return invalid()
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return invalid()
		}
	case "string":
		text, ok := value.(string)
		if !ok {
			return invalid()
		}
		length = utf8.RuneCountInString(text)
	case "string[]":
		values := array(value)
		if values == nil {
			return invalid()
		}
		length = len(values)
		for _, v := range values {
			if _, ok := v.(string); !ok {
				return invalid()
			}
		}
	case "json":
		if _, err := json.Marshal(value); err != nil {
			return invalid()
		}
	default:
		return invalid()
	}
	if min, ok := number(rule["minLength"]); ok && length < int(min) {
		return invalid()
	}
	if max, ok := number(rule["maxLength"]); ok && length > int(max) {
		return invalid()
	}
	if options := array(rule["enum"]); len(options) > 0 && !hasValue(options, value) {
		return invalid()
	}
	return nil
}
func validateReferences(value, transport any) error {
	values := array(value)
	if text, ok := value.(string); ok {
		values = []any{text}
	} else if object, ok := value.(map[string]any); ok {
		values = []any{object}
	}
	allowsURL := false
	maxCount := 0
	for _, entry := range array(transport) {
		text := fmt.Sprint(entry)
		if text == "public_https_url" {
			allowsURL = true
		}
		if strings.HasPrefix(text, "max_") {
			maxCount, _ = strconv.Atoi(strings.TrimPrefix(text, "max_"))
		}
	}
	if !allowsURL || len(values) == 0 || (maxCount > 0 && len(values) > maxCount) {
		return errors.New("The model does not support this reference representation or count.")
	}
	for _, v := range values {
		text, ok := v.(string)
		if object, isObject := v.(map[string]any); isObject && len(object) == 1 {
			text, ok = object["url"].(string)
			if !ok {
				text, ok = object["image_url"].(string)
			}
		}
		if !ok {
			return errors.New("Reference inputs must contain only public HTTPS URLs.")
		}
		u, err := url.Parse(text)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || strings.EqualFold(u.Hostname(), "localhost") || strings.HasSuffix(strings.ToLower(u.Hostname()), ".local") {
			return errors.New("Reference URLs must be explicit public HTTPS URLs without embedded credentials.")
		}
		if ip := net.ParseIP(u.Hostname()); ip != nil && (ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsMulticast()) {
			return errors.New("Reference URLs must use publicly reachable hosts.")
		}
	}
	return nil
}
