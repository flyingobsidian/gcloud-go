package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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
// flow through Main: after the command fails, invocation.offerEnableAPI is
// consulted; if it returns true, Main runs the same command again against
// the now-enabled API. Only the CLI does this: an embedder must not change
// project settings as a side effect.
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

// confirm asks the user for confirmation on w with the given text and
// default, reading the answer from r. When ask is false (--quiet, or stdin
// is not a terminal) it returns dflt without prompting. The prompt matches
// Python gcloud's `(y/N)?` style, so `(y/N)` when the default is No and
// `(Y/n)` when the default is Yes.
func confirm(r io.Reader, w io.Writer, ask bool, text string, dflt bool) bool {
	if !ask {
		return dflt
	}
	hint := "(y/N)?"
	if dflt {
		hint = "(Y/n)?"
	}
	fmt.Fprintf(w, "%s %s ", text, hint)
	line, err := bufio.NewReader(r).ReadString('\n')
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
func enableAPIAndWait(ctx context.Context, clients *gcp.Clients, project, service string) error {
	svc, err := clients.ServiceUsage(ctx)
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

// offerEnableAPI inspects err for a Service Usage "API not enabled" signal,
// asks the user (respecting --quiet and interactive tty) and enables the API
// on the current project. It returns true when the caller should run the
// failing command again; on false the caller reports err as usual (the
// "no", non-interactive and enable-failed paths).
func (inv *invocation) offerEnableAPI(ctx context.Context, err error) bool {
	service := apiNotEnabledService(err)
	if service == "" {
		return false
	}
	project, perr := inv.projectOrEmpty()
	if perr != nil || project == "" {
		return false
	}
	msg := fmt.Sprintf("API [%s] not enabled on project [%s]. Would you like to enable and retry (this will take a few minutes)?",
		service, project)
	if !inv.confirm(msg, false) {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if enErr := enableAPIAndWait(ctx, inv.clients(), project, service); enErr != nil {
		fmt.Fprintf(inv.stderr, "Enable failed: %v\n", enErr)
		return false
	}
	fmt.Fprintf(inv.stderr, "Enabled [%s] on project [%s]; retrying command.\n", service, project)
	return true
}
