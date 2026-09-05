package main

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	balanceOrigin    = "https://console.puretokensx.com"
	balanceUsagePath = "/api/product/console/api-keys/usage"
	balanceUnitPath  = "/api/product/console/status"
)

// The published CC Switch usage route accepts the existing API key without a
// browser session. Unlimited keys return user wallet quotas; limited keys return
// their own allowance. Neither response includes subscription quotas.
// The legacy billing subscription endpoint has an unlimited-key placeholder,
// so it must never be used as a balance or as a fallback.
func executeBalance(output io.Writer, svc service) error {
	// A balance-only fixed console origin, never a configured credential URL.
	svc.baseURL = balanceOrigin
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	body, status, retry, code, message, err := svc.request(ctx, http.MethodGet, balanceUsagePath, nil, "")
	object, decodeErr := readAPIObject(body)
	if err != nil || status < 200 || status >= 300 || decodeErr != nil || object["code"] != true {
		return balanceFailure(output, status, retry, code, message, true)
	}
	data := jsonObject(object["data"])
	remaining, remainingOK := balanceQuota(data["total_available"])
	used, usedOK := balanceQuota(data["total_used"])
	total, totalOK := balanceQuota(data["total_granted"])
	unlimited, scopeOK := data["unlimited_quota"].(bool)
	if data["object"] != "token_usage" || !remainingOK || !usedOK || used < 0 || !totalOK || !scopeOK {
		return balanceFailure(output, status, 0, "", "The API did not return usable balance information.", true)
	}

	// The public status metadata supplies the USD quota divisor used by the
	// official CC Switch script. No credential is sent on this public request.
	body, status, retry, code, message, err = svc.requestWithAuthentication(ctx, http.MethodGet, balanceUnitPath, nil, "", false)
	object, decodeErr = readAPIObject(body)
	if err != nil || status < 200 || status >= 300 || decodeErr != nil || object["success"] != true {
		return balanceFailure(output, status, retry, code, message, false)
	}
	divisor, divisorOK := number(jsonObject(object["data"])["quota_per_unit"])
	if !divisorOK || !finiteBalanceNumber(divisor) || divisor <= 0 {
		return balanceFailure(output, status, 0, "", "The API did not confirm the balance display unit.", false)
	}
	remaining, used, total = remaining/divisor, used/divisor, total/divisor
	if !finiteBalanceNumber(remaining) || !finiteBalanceNumber(used) || !finiteBalanceNumber(total) {
		return balanceFailure(output, status, 0, "", "The API returned an unusable balance amount.", false)
	}
	scope := "key_allowance"
	if unlimited {
		scope = "account_wallet"
	}
	writeJSON(output, map[string]any{
		"ok": true, "command": "balance",
		"result": map[string]any{
			"remaining": remaining, "used": used, "total": total,
			"unit": "USD", "scope": scope, "includes_subscription_quota": false,
		},
	})
	return nil
}

func finiteBalanceNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func balanceQuota(value any) (float64, bool) {
	quota, ok := number(value)
	// The server publishes integer quotas. Refuse values that lose integer
	// precision in JSON; retain actual negative available quota (overdraft).
	return quota, ok && finiteBalanceNumber(quota) && math.Trunc(quota) == quota && math.Abs(quota) <= 9007199254740991
}

func balanceFailure(output io.Writer, status, retry int, code, message string, authenticatedRead bool) error {
	if message == "" || message == "ok" || message == "success" {
		message = "The current balance could not be confirmed."
	}
	next := "Check the Pure Tokens console balance or try again later; no amount was confirmed."
	if authenticatedRead && (status == http.StatusUnauthorized || status == http.StatusForbidden) {
		next = "Check that the active Pure Tokens connection is enabled and its API key is valid in your configuration tool."
	}
	writeReceipt(output, apiFailure("submission", status, retry, code, message, next))
	return errors.New("balance unavailable")
}
