package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
)

// FormatCLIError renders err in the format the Python gcloud CLI uses for
// top-level command failures:
//
//	ERROR: (gcloud.<dotted.command.path>) <STATUS>: <message> <auth-context>
//
// When err (or any wrapped error in its chain) is a *googleapi.Error, the
// status is the canonical gRPC name for the HTTP code and the message is the
// server-reported detail; the outer fmt.Errorf wrappers are discarded. For
// any other error, we fall back to the error's own message.
//
// If cmd is nil, the parenthesised command path is "gcloud".
func FormatCLIError(cmd *cobra.Command, err error) string {
	var status, message string

	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		status = httpStatusName(apiErr.Code)
		message = strings.TrimSuffix(apiErr.Message, ".")
	} else {
		message = strings.TrimSuffix(err.Error(), ".")
	}

	var b strings.Builder
	b.WriteString("ERROR: (")
	b.WriteString(commandDottedPath(cmd))
	b.WriteByte(')')
	if status != "" {
		b.WriteByte(' ')
		b.WriteString(status)
		b.WriteByte(':')
	}
	b.WriteByte(' ')
	b.WriteString(message)
	b.WriteByte('.')

	if ctx := authContextMessage(); ctx != "" {
		b.WriteByte(' ')
		b.WriteString(ctx)
	}
	return b.String()
}

// commandDottedPath converts a cobra command chain like "gcloud secrets delete"
// into the dotted form used by gcloud diagnostics ("gcloud.secrets.delete").
// A nil command produces "gcloud".
func commandDottedPath(cmd *cobra.Command) string {
	if cmd == nil {
		return "gcloud"
	}
	return strings.ReplaceAll(cmd.CommandPath(), " ", ".")
}

// httpStatusName returns the canonical gRPC status name for the given HTTP
// status code (e.g. 404 → "NOT_FOUND"), matching the values that appear in
// gcloud's error output. Unknown codes yield an empty string, which suppresses
// the STATUS: prefix in the rendered message.
func httpStatusName(code int) string {
	switch code {
	case 400:
		return "INVALID_ARGUMENT"
	case 401:
		return "UNAUTHENTICATED"
	case 403:
		return "PERMISSION_DENIED"
	case 404:
		return "NOT_FOUND"
	case 409:
		return "ALREADY_EXISTS"
	case 412:
		return "FAILED_PRECONDITION"
	case 429:
		return "RESOURCE_EXHAUSTED"
	case 499:
		return "CANCELLED"
	case 500:
		return "INTERNAL"
	case 501:
		return "NOT_IMPLEMENTED"
	case 503:
		return "UNAVAILABLE"
	case 504:
		return "DEADLINE_EXCEEDED"
	}
	return ""
}

// authContextMessage returns the trailing "This command is authenticated as..."
// sentence gcloud appends to error output. Which sentence variant is used
// depends on how credentials were resolved, matching credentials/store.py's
// self.CredentialInfo(). Returns "" when no authentication context can be
// determined.
func authContextMessage() string {
	if f := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); f != "" {
		account := activeAccount()
		if account == "" {
			account = "the service account"
		}
		return fmt.Sprintf(
			"This command is authenticated as %s using the credentials in %s, specified by the [auth/credential_file_override] property.",
			account, f,
		)
	}
	if account := activeAccount(); account != "" {
		return fmt.Sprintf(
			"This command is authenticated as %s which is the active account specified by the [core/account] property.",
			account,
		)
	}
	return ""
}

// activeAccount returns the account the CLI would authenticate as: the value
// of --account when set, else the [core/account] property from the gcloud
// config, else "".
func activeAccount() string {
	if flagAccount != "" {
		return flagAccount
	}
	props, err := config.Load()
	if err != nil {
		return ""
	}
	return props.Core.Account
}
