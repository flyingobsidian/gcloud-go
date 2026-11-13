package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
)

// formatCLIError is FormatCLIError with the trailing authentication sentence
// (see authContextMessage) given as authContext; "" omits it.
func formatCLIError(cmd *cobra.Command, err error, authContext string) string {
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

	if authContext != "" {
		b.WriteByte(' ')
		b.WriteString(authContext)
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

// authContext returns the authentication sentence for credentials read from
// the file credFile (when not "") as account, or for the stored credentials
// of account.
func authContext(credFile, account string) string {
	if credFile != "" {
		if account == "" {
			account = "the service account"
		}
		return fmt.Sprintf(
			"This command is authenticated as %s using the credentials in %s, specified by the [auth/credential_file_override] property.",
			account, credFile,
		)
	}
	if account != "" {
		return fmt.Sprintf(
			"This command is authenticated as %s which is the active account specified by the [core/account] property.",
			account,
		)
	}
	return ""
}

// formatError is FormatCLIError for an error from a command run by inv: the
// authentication sentence describes inv's account and environment.
func (inv *invocation) formatError(cmd *cobra.Command, err error) string {
	account := inv.flags.account
	if account == "" {
		if props, perr := inv.properties(); perr == nil {
			account = props.Core.Account
		}
	}
	return formatCLIError(cmd, err, authContext(inv.env.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), account))
}
