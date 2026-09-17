package main

import (
	"errors"
	"fmt"
)

// Query and submission use the same value and cross-field rules. Queries infer
// required operation inputs; submissions validate actual attachment counts.
func validateModelParameter(schema parameterSchema, operation mediaOperation, name string, value any, counts map[string]int, multipart bool) error {
	if counts[name] > 0 && multipart {
		return errors.New("Do not send the same media field as both a file and a JSON parameter.")
	}
	rule, exists := schema.Properties[name]
	if !exists {
		input, ok := operation.Inputs[name]
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
	if err := validateProperty(name, value, rule); err != nil {
		return err
	}
	if transports, exists := schema.Constraints["reference_transport"][name]; exists {
		if multipart {
			if _, declared := operation.Inputs[name]; !declared {
				return errors.New("Mixing URL references and local attachments requires an explicitly declared combination operation.")
			}
		}
		if err := validateReferences(value, transports); err != nil {
			return err
		}
	}
	if _, unsupported := schema.Constraints["unsupported_inputs"][name]; unsupported {
		return errors.New("The selected model explicitly excludes this input.")
	}
	return nil
}

func validateModelConstraints(schema parameterSchema, operation string, parameters map[string]any, counts map[string]int) error {
	present := func(key string) bool { return counts[key] > 0 || parameters[key] != nil }
	groups := 0
	for _, fields := range schema.Constraints["exclusive_reference_sets"] {
		for _, field := range array(fields) {
			if present(fmt.Sprint(field)) {
				groups++
				break
			}
		}
	}
	if groups > 1 {
		return errors.New("First/last frames cannot be combined with additional reference media.")
	}
	for key, companions := range schema.Constraints["requires_together"] {
		if present(key) {
			for _, companion := range array(companions) {
				if name, ok := companion.(string); !ok || !present(name) {
					return errors.New("Required companion parameters must be supplied together as declared by the model.")
				}
			}
		}
	}
	mode := "text"
	if present("image") || present("first_frame_image") || present("last_frame_image") || operation == "image_to_video" {
		mode = "image"
	}
	if operation == "reference_image_video" || operation == "reference_video" || operation == "reference_audio" {
		mode = "reference"
	}
	if allowed, exists := schema.Constraints["resolution_by_mode"][mode]; exists && parameters["resolution"] != nil && !hasValue(array(allowed), parameters["resolution"]) {
		return errors.New("This resolution is unavailable for the selected reference mode. Choose a declared resolution.")
	}
	return nil
}
