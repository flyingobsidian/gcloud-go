package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// nmVFLFakeAPI records requests and answers every one with a done operation.
type nmVFLFakeAPI struct {
	requests []string
	bodies   []string
}

func (f *nmVFLFakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	f.bodies = append(f.bodies, string(body))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"name":"projects/p/locations/global/operations/op1","done":true}`)
}

func TestNetmgmtVFLConnectionLogging(t *testing.T) {
	for _, tc := range []struct {
		verb, choice, want, wantMask string
	}{
		{"create", "enabled", "CONNECTION_LOGGING_ENABLED", ""},
		{"create", "connection-logging-disabled", "CONNECTION_LOGGING_DISABLED", ""},
		{"update", "disabled", "CONNECTION_LOGGING_DISABLED", "updateMask=connection_logging"},
		{"update", "connection-logging-enabled", "CONNECTION_LOGGING_ENABLED", "updateMask=connection_logging"},
	} {
		t.Run(tc.verb+" "+tc.choice, func(t *testing.T) {
			api := &nmVFLFakeAPI{}
			inv, _, _ := newFakeAPIInvocation(t, api)
			if err := runInvocation(inv, "network-management", "vpc-flow-logs-configs", tc.verb, "c1",
				"--project", "p", "--connection-logging", tc.choice); err != nil {
				t.Fatal(err)
			}
			if len(api.bodies) != 1 {
				t.Fatalf("requests = %v, want one", api.requests)
			}
			var sent map[string]any
			if err := json.Unmarshal([]byte(api.bodies[0]), &sent); err != nil || sent["connectionLogging"] != tc.want {
				t.Errorf("body = %s, want connectionLogging %s", api.bodies[0], tc.want)
			}
			if tc.wantMask != "" && !strings.Contains(api.requests[0], tc.wantMask) {
				t.Errorf("request = %s, want %s", api.requests[0], tc.wantMask)
			}
		})
	}
}

func TestNetmgmtVFLConnectionLoggingInvalid(t *testing.T) {
	api := &nmVFLFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "network-management", "vpc-flow-logs-configs", "create", "c1",
		"--project", "p", "--connection-logging", "on")
	want := "argument --connection-logging: Invalid choice: 'on'.\n\nValid choices are " +
		"[connection-logging-disabled, connection-logging-enabled, disabled, enabled]."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
	if len(api.requests) != 0 {
		t.Errorf("requests = %v, want none", api.requests)
	}
}
