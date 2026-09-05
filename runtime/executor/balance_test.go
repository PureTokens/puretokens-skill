package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const balanceUsageFixture = `{"code":true,"message":"ok","data":{"object":"token_usage","name":"private-key-name","total_available":5000000,"total_used":9836122,"total_granted":14836122,"unlimited_quota":true,"model_limits":{"private-model":true}}}`
const balanceUnitFixture = `{"success":true,"data":{"quota_per_unit":500000,"quota_display_type":"CNY","usd_exchange_rate":7}}`

type balanceFixtureTransport struct {
	t       *testing.T
	handler http.Handler
}

func (transport balanceFixtureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	if request.URL.Scheme != "https" || request.URL.Host != "console.puretokensx.com" || request.Method != "GET" {
		transport.t.Fatalf("balance used a non-fixed target or method: %s %s", request.Method, request.URL)
	}
	deadline, ok := request.Context().Deadline()
	if !ok || time.Until(deadline) > 30*time.Second {
		transport.t.Error("unbounded balance read")
	}
	if request.Header.Get("Cookie") != "" {
		transport.t.Error("unexpected login cookie")
	}
	switch request.URL.Path {
	case balanceUsagePath:
		if request.Header.Get("Authorization") != "Bearer synthetic-fixture-token" {
			transport.t.Error("missing private bearer authentication")
		}
	case balanceUnitPath:
		if request.Header.Get("Authorization") != "" {
			transport.t.Error("public metadata received a credential")
		}
	default:
		transport.t.Fatalf("unexpected balance endpoint: %s", request.URL.Path)
	}
	recorder := httptest.NewRecorder()
	transport.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}

func balanceFixtureService(t *testing.T, handler http.HandlerFunc) service {
	return service{
		baseURL: "https://unrelated-configured-origin.invalid", token: "synthetic-fixture-token",
		client: &http.Client{Transport: balanceFixtureTransport{t, handler}},
	}
}

func TestBalanceUsesCCSwitchAvailableQuotaAndScope(t *testing.T) {
	for _, fixture := range []struct {
		name, usage, unit, scope string
		remaining, used, total   float64
	}{
		{"unlimited key wallet", balanceUsageFixture, balanceUnitFixture, "account_wallet", 10, 19.672244, 29.672244},
		{"limited key allowance", strings.ReplaceAll(balanceUsageFixture, `"unlimited_quota":true`, `"unlimited_quota":false`), balanceUnitFixture, "key_allowance", 10, 19.672244, 29.672244},
		{"returned non-default divisor", balanceUsageFixture, `{"success":true,"data":{"quota_per_unit":1000000}}`, "account_wallet", 5, 9.836122, 14.836122},
		{"zero wallet", `{"code":true,"data":{"object":"token_usage","total_available":0,"total_used":0,"total_granted":0,"unlimited_quota":true}}`, balanceUnitFixture, "account_wallet", 0, 0, 0},
		{"overdraft is not clamped", `{"code":true,"data":{"object":"token_usage","total_available":-500000,"total_used":1000000,"total_granted":500000,"unlimited_quota":true}}`, balanceUnitFixture, "account_wallet", -1, 2, 1},
		{"large genuine quota is not a placeholder", `{"code":true,"data":{"object":"token_usage","total_available":100000000,"total_used":0,"total_granted":100000000,"unlimited_quota":false}}`, balanceUnitFixture, "key_allowance", 200, 0, 200},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			var paths []string
			svc := balanceFixtureService(t, func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				if r.URL.Path == balanceUsagePath {
					io.WriteString(w, fixture.usage)
				} else {
					io.WriteString(w, fixture.unit)
				}
			})
			var out bytes.Buffer
			if err := executeBalance(&out, svc); err != nil {
				t.Fatalf("%v: %s", err, out.String())
			}
			var result map[string]any
			json.Unmarshal(out.Bytes(), &result)
			want := map[string]any{"ok": true, "command": "balance", "result": map[string]any{
				"remaining": fixture.remaining, "used": fixture.used, "total": fixture.total,
				"scope": fixture.scope, "unit": "USD", "includes_subscription_quota": false,
			}}
			if !reflect.DeepEqual(result, want) {
				t.Fatalf("wrong balance projection: %s", out.String())
			}
			if !reflect.DeepEqual(paths, []string{balanceUsagePath, balanceUnitPath}) {
				t.Fatalf("extra reads: %v", paths)
			}
			for _, forbidden := range []string{"private-key-name", "private-model", "synthetic-fixture-token", "99999980", "100000000", "hard_limit_usd"} {
				if strings.Contains(out.String(), forbidden) {
					t.Errorf("private data or legacy placeholder leaked: %s", forbidden)
				}
			}
		})
	}
}

func TestBalanceRejectsMalformedUsageWithoutFallbackOrUnitRead(t *testing.T) {
	for _, body := range []string{
		`{"hard_limit_usd":100000000}`, `{"hard_limit_usd":16,"total_usage":600}`, `{}`,
		`{"error":{"message":"Denied"}}`, `{"code":false,"message":"Unavailable"}`, `<html>failure</html>`,
		strings.ReplaceAll(balanceUsageFixture, `"code":true`, `"code":false`),
		strings.ReplaceAll(balanceUsageFixture, `"code":true`, `"code":"true"`),
		strings.ReplaceAll(balanceUsageFixture, `"unlimited_quota":true`, `"unlimited_quota":"true"`),
		strings.ReplaceAll(balanceUsageFixture, `"unlimited_quota":true,`, ``),
		strings.ReplaceAll(balanceUsageFixture, `"total_available":5000000`, `"total_available":"5000000"`),
		strings.ReplaceAll(balanceUsageFixture, `"total_available":5000000,`, ``),
		strings.ReplaceAll(balanceUsageFixture, `"total_used":9836122`, `"total_used":-1`),
		strings.ReplaceAll(balanceUsageFixture, `"total_available":5000000`, `"total_available":1.5`),
		strings.ReplaceAll(balanceUsageFixture, `"total_available":5000000`, `"total_available":9007199254740992`),
		strings.ReplaceAll(balanceUsageFixture, `"total_granted":14836122`, `"total_granted":1e999`),
	} {
		calls := 0
		svc := balanceFixtureService(t, func(w http.ResponseWriter, r *http.Request) { calls++; io.WriteString(w, body) })
		var out bytes.Buffer
		if executeBalance(&out, svc) == nil || calls != 1 || strings.Contains(out.String(), `"remaining"`) {
			t.Fatalf("malformed usage interpreted or retried: calls=%d receipt=%s", calls, out.String())
		}
	}
}

func TestBalanceRequiresPublishedUnitMetadata(t *testing.T) {
	for _, body := range []string{
		`{}`, `{"success":false,"data":{"quota_per_unit":500000}}`, `{"success":true,"data":{}}`,
		`{"success":true,"data":{"quota_per_unit":0}}`, `{"success":true,"data":{"quota_per_unit":-1}}`,
		`{"success":true,"data":{"quota_per_unit":"500000"}}`, `{"success":true,"data":{"quota_per_unit":1e-320}}`,
		`{"success":true,"data":{"quota_per_unit":1e999}}`, `<html>Login</html>`,
	} {
		calls := 0
		svc := balanceFixtureService(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.URL.Path == balanceUsagePath {
				io.WriteString(w, balanceUsageFixture)
			} else {
				io.WriteString(w, body)
			}
		})
		var out bytes.Buffer
		if executeBalance(&out, svc) == nil || calls != 2 || strings.Contains(out.String(), `"remaining"`) {
			t.Fatalf("unconfirmed units guessed or retried: calls=%d receipt=%s", calls, out.String())
		}
	}
}

func TestBalanceFailuresDoNotRetryFollowRedirectsOrExposePrivateData(t *testing.T) {
	for _, phase := range []string{"usage", "unit"} {
		for _, status := range []int{301, 401, 403, 429, 503} {
			calls := 0
			svc := balanceFixtureService(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if phase == "unit" && r.URL.Path == balanceUsagePath {
					io.WriteString(w, balanceUsageFixture)
					return
				}
				w.Header().Set("Location", "https://untrusted.invalid/steal")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(status)
				// A public error may echo a secret, even through JSON escapes.
				io.WriteString(w, `{"error":{"code":"query_denied","message":"Rejected synthetic-fixture-\u0074oken"}}`)
			})
			var out bytes.Buffer
			if executeBalance(&out, svc) == nil {
				t.Fatal("HTTP failure accepted")
			}
			var got receipt
			json.Unmarshal(out.Bytes(), &got)
			wantCalls := 1
			if phase == "unit" {
				wantCalls = 2
			}
			if calls != wantCalls || got.OK || got.HTTPStatus != status || got.RetryAfterSecs != 60 || got.APIErrorCode != "query_denied" || got.NextAction == "" {
				t.Fatalf("wrong failure receipt: calls=%d %s", calls, out.String())
			}
			if status == 401 || status == 403 {
				if strings.Contains(got.NextAction, "configuration tool") != (phase == "usage") {
					t.Fatal("public metadata rejection was confused with API-key authentication failure")
				}
			}
			if strings.Contains(out.String(), "synthetic-fixture-token") || strings.Contains(out.String(), "untrusted") || strings.Contains(out.String(), `"remaining"`) {
				t.Fatal("failed balance exposed private data or an amount")
			}
		}
	}
}
