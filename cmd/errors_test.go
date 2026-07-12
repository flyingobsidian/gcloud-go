package cmd

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
)

func TestHttpStatusName(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{400, "INVALID_ARGUMENT"},
		{403, "PERMISSION_DENIED"},
		{404, "NOT_FOUND"},
		{409, "ALREADY_EXISTS"},
		{500, "INTERNAL"},
		{999, ""},
	}
	for _, c := range cases {
		if got := httpStatusName(c.code); got != c.want {
			t.Errorf("httpStatusName(%d) = %q, want %q", c.code, got, c.want)
		}
	}
}

func TestCommandDottedPath(t *testing.T) {
	root := &cobra.Command{Use: "gcloud"}
	secrets := &cobra.Command{Use: "secrets"}
	del := &cobra.Command{Use: "delete"}
	root.AddCommand(secrets)
	secrets.AddCommand(del)

	if got := commandDottedPath(del); got != "gcloud.secrets.delete" {
		t.Errorf("commandDottedPath = %q, want gcloud.secrets.delete", got)
	}
	if got := commandDottedPath(nil); got != "gcloud" {
		t.Errorf("commandDottedPath(nil) = %q, want gcloud", got)
	}
}

func TestFormatCLIErrorGoogleapi(t *testing.T) {
	root := &cobra.Command{Use: "gcloud"}
	secrets := &cobra.Command{Use: "secrets"}
	del := &cobra.Command{Use: "delete"}
	root.AddCommand(secrets)
	secrets.AddCommand(del)

	// Emulate the wrap that runSecretsDelete applies before returning.
	apiErr := &googleapi.Error{
		Code:    404,
		Message: "Secret [projects/1234/secrets/test-secret-e1c17dbb] not found.",
	}
	wrapped := fmt.Errorf("deleting secret: %w", apiErr)

	got := FormatCLIError(del, wrapped)
	prefix := "ERROR: (gcloud.secrets.delete) NOT_FOUND: Secret [projects/1234/secrets/test-secret-e1c17dbb] not found."
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("FormatCLIError got:\n%s\nwant prefix:\n%s", got, prefix)
	}
	if strings.Contains(got, "deleting secret") {
		t.Errorf("outer wrap message leaked into output: %s", got)
	}
	if strings.Contains(got, "googleapi") {
		t.Errorf("raw googleapi label leaked into output: %s", got)
	}
	// No trailing suffix like ", notFound" in the surfaced message.
	if strings.Contains(got, "notFound") {
		t.Errorf("legacy trailing notFound tag leaked: %s", got)
	}
}

func TestFormatCLIErrorNonAPI(t *testing.T) {
	root := &cobra.Command{Use: "gcloud"}
	secrets := &cobra.Command{Use: "secrets"}
	create := &cobra.Command{Use: "create"}
	root.AddCommand(secrets)
	secrets.AddCommand(create)

	err := errors.New("no active account")
	got := FormatCLIError(create, err)
	want := "ERROR: (gcloud.secrets.create) no active account."
	if !strings.HasPrefix(got, want) {
		t.Errorf("FormatCLIError non-API got:\n%s\nwant prefix:\n%s", got, want)
	}
}

func TestFormatCLIErrorUnknownStatus(t *testing.T) {
	// An HTTP code with no gRPC mapping should still produce a clean message
	// (no dangling "STATUS:" prefix).
	apiErr := &googleapi.Error{Code: 418, Message: "I'm a teapot"}
	got := FormatCLIError(nil, apiErr)
	if !strings.HasPrefix(got, "ERROR: (gcloud) I'm a teapot.") {
		t.Errorf("unknown status output = %q", got)
	}
	if strings.Contains(got, ":  ") {
		t.Errorf("double space where STATUS: would have been: %q", got)
	}
}

func TestFormatCLIErrorAuthContextCredFile(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/creds.json")
	prevAccount := flagAccount
	flagAccount = "sa@example.iam.gserviceaccount.com"
	t.Cleanup(func() { flagAccount = prevAccount })

	got := FormatCLIError(nil, &googleapi.Error{Code: 404, Message: "nope"})
	if !strings.Contains(got, "using the credentials in /tmp/creds.json") {
		t.Errorf("expected credential_file_override context, got: %s", got)
	}
	if !strings.Contains(got, "sa@example.iam.gserviceaccount.com") {
		t.Errorf("expected active account in context, got: %s", got)
	}
	if !strings.Contains(got, "[auth/credential_file_override]") {
		t.Errorf("expected [auth/credential_file_override] property tag, got: %s", got)
	}
}

func TestFormatCLIErrorAuthContextAccount(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	prevAccount := flagAccount
	flagAccount = "user@example.com"
	t.Cleanup(func() { flagAccount = prevAccount })

	got := FormatCLIError(nil, &googleapi.Error{Code: 404, Message: "nope"})
	if !strings.Contains(got, "[core/account]") {
		t.Errorf("expected [core/account] property tag, got: %s", got)
	}
	if !strings.Contains(got, "user@example.com") {
		t.Errorf("expected active account, got: %s", got)
	}
}
