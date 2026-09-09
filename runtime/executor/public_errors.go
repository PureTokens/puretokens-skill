package main

// Error text is an untrusted server field: it may echo prompts, media paths,
// credentials or upstream diagnostics. Only known categories have public text.
// Unknown codes/messages are omitted, leaving phase/HTTP status and local
// next_action to explain the failure. Never invent an API code from a message.
var publicErrorMessages = map[string]string{
	"invalid_request":          "The API rejected the request parameters or input.",
	"invalid_request_error":    "The API rejected the request parameters or input.",
	"invalid_parameter":        "The API rejected a request parameter.",
	"unsupported_parameter":    "The API does not support a requested parameter.",
	"unsupported_operation":    "The API does not support the requested operation.",
	"model_not_found":          "The requested model was not available for this request.",
	"model_not_allowed":        "The API did not authorize the requested model.",
	"content_policy_violation": "The API rejected the content under its content policy. Revise the request.",
	"content_moderation":       "The API rejected the content under its content policy. Revise the request.",
	"moderation_blocked":       "The API rejected the content under its content policy. Revise the request.",
	"auth_required":            "The API requires authentication for this request.",
	"invalid_api_key":          "The API rejected authentication for this request.",
	"authentication_error":     "The API rejected authentication for this request.",
	"permission_denied":        "The API did not authorize this request.",
	"query_denied":             "The API rejected this query.",
	"insufficient_quota":       "The API reported insufficient quota for this request.",
	"insufficient_balance":     "The API reported insufficient balance for this request.",
	"rate_limit":               "The API requested a slower request rate.",
	"rate_limited":             "The API requested a slower request rate.",
	"rate_limit_exceeded":      "The API requested a slower request rate.",
	"upstream_unavailable":     "The media service is temporarily unavailable.",
	"service_unavailable":      "The media service is temporarily unavailable.",
	"internal_error":           "The API could not complete the request.",
	"task_not_found":           "The API could not find this task.",
	"task_expired":             "The API reported that this task has expired.",
}

func publicErrorCategory(code, message string) (string, string) {
	if text, known := publicErrorMessages[code]; known {
		return code, text
	}
	// One existing service response has no code. Recognize its entire literal,
	// never a substring that could carry private request content.
	if message == "The generated images appear to be unsafe. Try modifying the prompt or seeds." {
		return "", "The API reported unsafe generated content. Revise the request."
	}
	return "", ""
}
