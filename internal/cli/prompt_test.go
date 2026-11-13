package cli

import (
	"context"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestInvocationConfirm(t *testing.T) {
	tests := []struct {
		name        string
		interactive bool
		quiet       bool
		input       string
		dflt        bool
		want        bool
		wantPrompt  string
	}{
		{"yes", true, false, "y\n", false, true, "Go? (y/N)? "},
		{"YES", true, false, "YES\n", false, true, "Go? (y/N)? "},
		{"no", true, false, "n\n", true, false, "Go? (Y/n)? "},
		{"empty takes default", true, false, "\n", true, true, "Go? (Y/n)? "},
		{"eof takes default", true, false, "", true, true, "Go? (Y/n)? "},
		{"quiet", true, true, "n\n", true, true, ""},
		{"not interactive", false, false, "y\n", false, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, stdout, stderr := newTestInvocation(t, nil, tt.input)
			inv.interactive = tt.interactive
			inv.flags.quiet = tt.quiet
			if got := inv.confirm("Go?", tt.dflt); got != tt.want {
				t.Errorf("confirm = %v, want %v", got, tt.want)
			}
			if got := stderr.String(); got != tt.wantPrompt {
				t.Errorf("prompt on stderr = %q, want %q", got, tt.wantPrompt)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want nothing", stdout.String())
			}
		})
	}
}

func TestInvocationEmit(t *testing.T) {
	inv, stdout, _ := newTestInvocation(t, nil, "")
	inv.flags.format = "value(name)"
	if err := inv.emit([]map[string]any{{"name": "a"}, {"name": "b"}}); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "a\nb\n" {
		t.Errorf("emit wrote %q, want %q", got, "a\nb\n")
	}

	stdout.Reset()
	if err := inv.printJSON([]string(nil)); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "[]\n" {
		t.Errorf("printJSON wrote %q, want %q", got, "[]\n")
	}
}

func TestOfferEnableAPIDeclines(t *testing.T) {
	disabled := &googleapi.Error{Code: 403, Details: []any{map[string]any{
		"@type":    "type.googleapis.com/google.rpc.ErrorInfo",
		"reason":   "SERVICE_DISABLED",
		"metadata": map[string]any{"service": "spanner.googleapis.com"},
	}}}
	tests := []struct {
		name        string
		err         error
		interactive bool
		answer      string
	}{
		{"other error", &googleapi.Error{Code: 404}, true, "y\n"},
		{"not interactive", disabled, false, "y\n"},
		{"user says no", disabled, true, "n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, _, _ := newTestInvocation(t, map[string]string{"CLOUDSDK_CORE_PROJECT": "my-project"}, tt.answer)
			inv.interactive = tt.interactive
			if inv.offerEnableAPI(context.Background(), tt.err) {
				t.Error("offerEnableAPI = true, want false")
			}
		})
	}
}
