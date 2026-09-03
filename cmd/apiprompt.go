package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"google.golang.org/api/googleapi"
	su "google.golang.org/api/serviceusage/v1"
)

// --- gcloud "prompt to enable disabled API" flow (#1856) ---
//
// gcloud-python detects the "API has not been used ... before or it is
// disabled" error emitted by Service Usage on the first request against a
// project + API pair, prompts the caller, and — on confirmation — enables
// the API and reruns the failing command. gcloud-go plumbs the equivalent
// flow through main.go: after the command fails, MaybePromptEnableAPI is
// consulted; if it returns retry=true, main re-invokes Execute() with the
// same argv so the operation runs again against the now-enabled API.
//
// The prompt is suppressed when stdin is not a terminal or when --quiet is
// set, matching Python's behaviour. Failures during the enable request or
// operation-wait fall through to the original error so the caller isn't
// surprised by an interactive prompt turning a "PERMISSION_DENIED" into an
// enable-API RPC failure.

// apiNotEnabledFromMessage extracts the API service name from the classic
// Service Usage message:
//
//	Cloud Spanner API has not been used in project X before or it is
//	disabled. Enable it by visiting
//	https://console.developers.google.com/apis/api/spanner.googleapis.com/overview?project=X
//	then retry.
//
// The presence of the ".../apis/api/<service>/overview" URL fragment is used
// as the anchor rather than parsing english prose, so future wording
// tweaks that keep the URL intact still resolve.
var apiNotEnabledURLRE = regexp.MustCompile(`/apis/api/([a-z0-9\-.]+)/overview`)

// apiNotEnabledService returns the API service name (e.g. "spanner.googleapis.com")
// referenced by an "API has not been used" error, or "" if err is not that
// error. It walks the googleapi.Error details for a SERVICE_DISABLED
// ErrorInfo first (structured) and falls back to matching the Enable-URL
// substring on the human-readable message.
func apiNotEnabledService(err error) string {
	var gerr *googleapi.Error
	if !errors.As(err, &gerr) {
		return ""
	}
	if gerr.Code != 403 {
		return ""
	}
	// Structured detail: {"@type": ".../ErrorInfo", "reason":
	// "SERVICE_DISABLED", "metadata": {"service": "spanner.googleapis.com"}}.
	for _, d := range gerr.Details {
		m, ok := d.(map[string]any)
		if !ok {
			continue
		}
		if reason, _ := m["reason"].(string); reason != "SERVICE_DISABLED" {
			continue
		}
		md, ok := m["metadata"].(map[string]any)
		if !ok {
			continue
		}
		if service, _ := md["service"].(string); service != "" {
			return service
		}
	}
	// Fallback: parse the "Enable it by visiting …/apis/api/<service>/overview"
	// URL. Guarded on "not been used" so unrelated 403s don't leak through.
	if !strings.Contains(gerr.Message, "not been used in project") &&
		!strings.Contains(gerr.Message, "has not been used") {
		return ""
	}
	if m := apiNotEnabledURLRE.FindStringSubmatch(gerr.Message); len(m) == 2 {
		return m[1]
	}
	return ""
}

// promptYesNo asks the user for confirmation on stderr with the given text
// and default. It respects --quiet (returns dflt without prompting) and
// non-interactive stdin (also returns dflt). The prompt matches Python
// gcloud's `(y/N)?` style, so `(y/N)` when the default is No and `(Y/n)`
// when the default is Yes.
func promptYesNo(text string, dflt bool) bool {
	if flagQuiet {
		return dflt
	}
	if !IsInteractive() {
		return dflt
	}
	hint := "(y/N)?"
	if dflt {
		hint = "(Y/n)?"
	}
	fmt.Fprintf(os.Stderr, "%s %s ", text, hint)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return dflt
	}
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" {
		return dflt
	}
	return line == "y" || line == "yes"
}

// enableAPIAndWait enables service on project using Service Usage, then
// polls the returned operation until it completes or the context is
// cancelled. The initial 45s deadline reflects the Python client's
// behaviour of blocking until the enable propagates.
func enableAPIAndWait(ctx context.Context, project, service, account string) error {
	svc, err := gcp.ServiceUsageService(ctx, account)
	if err != nil {
		return fmt.Errorf("service usage client: %w", err)
	}
	op, err := svc.Services.Enable(serviceResourceName(project, service),
		&su.EnableServiceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling %s: %w", service, err)
	}
	deadline := time.Now().Add(45 * time.Second)
	backoff := time.Second
	for !op.Done {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for enable operation %s", op.Name)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 5*time.Second {
			backoff += time.Second
		}
		op, err = svc.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling enable operation: %w", err)
		}
	}
	if op.Error != nil {
		return fmt.Errorf("enable operation %s failed: %s", op.Name, op.Error.Message)
	}
	return nil
}

// MaybePromptEnableAPI inspects err for a Service Usage "API not enabled"
// signal, prompts the user (respecting --quiet and interactive tty), and
// enables the API on the current project. It returns retry=true when the
// caller should re-invoke the failing command; returned err is either nil
// (retry) or the original err (so the caller preserves existing output for
// the "no", non-interactive, or enable-failed paths).
func MaybePromptEnableAPI(origErr error) (retry bool, err error) {
	service := apiNotEnabledService(origErr)
	if service == "" {
		return false, origErr
	}
	project, perr := resolveProject()
	if perr != nil || project == "" {
		return false, origErr
	}
	msg := fmt.Sprintf("API [%s] not enabled on project [%s]. Would you like to enable and retry (this will take a few minutes)?",
		service, project)
	if !promptYesNo(msg, false) {
		return false, origErr
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if enErr := enableAPIAndWait(ctx, project, service, flagAccount); enErr != nil {
		fmt.Fprintf(os.Stderr, "Enable failed: %v\n", enErr)
		return false, origErr
	}
	fmt.Fprintf(os.Stderr, "Enabled [%s] on project [%s]; retrying command.\n", service, project)
	return true, nil
}
