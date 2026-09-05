package main

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
)

// The gateway's existing bearer-authenticated billing pair is independent of
// the Web account-session API. Its *_usd names do not declare actual currency
// or whether the server is configured to display account vs token statistics.
func executeBalance(output io.Writer, svc service) error {
	paths := []string{"/v1/dashboard/billing/subscription", "/v1/dashboard/billing/usage"}
	values := make([]map[string]any, 0, 2)
	for _, path := range paths {
		body, status, retry, code, message, err := svc.request(context.Background(), http.MethodGet, path, nil, "")
		object, decodeErr := readAPIObject(body)
		if err != nil || status < 200 || status >= 300 || decodeErr != nil {
			writeReceipt(output, apiFailure("submission", status, retry, code, message, "The authenticated billing read was not completed. Check the Pure Tokens console for account balance; do not estimate a value or infer key validity from this failure."))
			return errors.New("balance unavailable")
		}
		values = append(values, object)
	}
	limit, ok := number(values[0]["hard_limit_usd"])
	used, ok2 := number(values[1]["total_usage"])
	if !ok || !ok2 || math.IsNaN(limit) || math.IsInf(limit, 0) || math.IsNaN(used) || math.IsInf(used, 0) || limit < 0 || used < 0 {
		writeReceipt(output, validationFailure("The API did not return readable billing totals; the balance is unknown."))
		return errors.New("invalid billing response")
	}
	writeJSON(output, map[string]any{"ok": true, "command": "balance", "result": map[string]any{
		"reported_limit": limit, "reported_usage": used / 100, "reported_remaining": limit - used/100,
		"unit": "api_display_unit_unspecified", "scope": "api_account_or_key_scope_unspecified",
		"note": "Two sequential API snapshots; currency, account/key scope and unlimited-quota status are not declared. This is not a media price quote or authorization guarantee.",
	}})
	return nil
}
