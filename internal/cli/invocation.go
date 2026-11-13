package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/mattn/go-isatty"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

// invocation is the state of one gcloud-go command run: its standard
// streams, environment, credentials and the values of the global flags.
// Commands read everything through it rather than through package globals,
// so several invocations can run in one process at the same time (#2070).
type invocation struct {
	stdin          io.Reader
	stdout, stderr io.Writer
	env            config.Env
	tokens         *auth.TokenProvider
	// interactive reports whether prompts may be shown on stdout and
	// answered on stdin.
	interactive bool

	flags globalFlags

	// ctx is the context the invocation runs under, for helpers that have
	// no command to take it from. Main and RunInProcess set it.
	ctx context.Context

	// clientOptions are applied to every Google API client; tests use them
	// to talk to fake servers.
	clientOptions []option.ClientOption

	// restTokenSource and restUserProject, when set, replace the
	// credentials and billing project of the invocation's REST clients;
	// tests use them.
	restTokenSource func(ctx context.Context, scopes ...string) (oauth2.TokenSource, error)
	restUserProject func() string

	propsOnce sync.Once
	props     *config.Properties
	propsErr  error
}

// globalFlags holds the values of the persistent flags on the root command.
type globalFlags struct {
	project           string
	zone              string
	account           string
	quiet             bool
	accessTokenFile   string
	billingProject    string
	configuration     string
	flagsFile         string
	filter            string
	flatten           []string
	format            string
	impersonateSA     string
	logHTTP           bool
	traceToken        string
	userOutputEnabled bool
	verbosity         string
}

// newOSInvocation returns an invocation using the process's standard
// streams and environment.
func newOSInvocation() *invocation {
	env := config.OSEnv()
	return &invocation{
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		env:         env,
		tokens:      auth.NewTokenProvider(env),
		interactive: isTerminal(os.Stdin),
	}
}

// isTerminal reports whether r is a terminal.
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// properties returns the active gcloud configuration, read at most once per
// invocation.
func (inv *invocation) properties() (*config.Properties, error) {
	inv.propsOnce.Do(func() {
		inv.props, inv.propsErr = inv.env.Load()
	})
	return inv.props, inv.propsErr
}

// resolve returns flag, else the environment variable envVar, else
// configVal.
func (inv *invocation) resolve(flag, envVar, configVal string) string {
	return inv.env.Resolve(flag, envVar, configVal)
}

// projectOrEmpty returns the project from --project, CLOUDSDK_CORE_PROJECT
// or the configuration, or "" if none is set.
func (inv *invocation) projectOrEmpty() (string, error) {
	props, err := inv.properties()
	if err != nil {
		return "", err
	}
	return inv.resolve(inv.flags.project, "CLOUDSDK_CORE_PROJECT", props.Core.Project), nil
}

// project returns the project the command applies to. When none is set it
// asks for one if the invocation is interactive, and fails otherwise.
func (inv *invocation) project() (string, error) {
	project, err := inv.projectOrEmpty()
	if err != nil || project != "" {
		return project, err
	}
	if !inv.interactive {
		return "", fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	project = inv.ask("Enter project: ")
	if project == "" {
		return "", fmt.Errorf("project is required")
	}
	return project, nil
}

// region returns the project and the Compute Engine region from
// CLOUDSDK_COMPUTE_REGION or the configuration, which may be empty. A missing
// project is an error.
func (inv *invocation) region() (string, string, error) {
	props, err := inv.properties()
	if err != nil {
		return "", "", err
	}
	project := inv.resolve(inv.flags.project, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	region := inv.resolve("", "CLOUDSDK_COMPUTE_REGION", props.Compute.Region)
	if project == "" {
		return "", "", fmt.Errorf("project is required")
	}
	return project, region, nil
}

// projectZone returns the project and Compute Engine zone the command
// applies to. A missing project is an error; a missing zone is asked for if
// the invocation is interactive.
func (inv *invocation) projectZone() (string, string, error) {
	props, err := inv.properties()
	if err != nil {
		return "", "", err
	}
	project := inv.resolve(inv.flags.project, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	zone := inv.resolve(inv.flags.zone, "CLOUDSDK_COMPUTE_ZONE", props.Compute.Zone)
	if project == "" {
		return "", "", fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	if zone == "" {
		if !inv.interactive {
			return "", "", fmt.Errorf("zone is required; set via --zone flag, CLOUDSDK_COMPUTE_ZONE env, or config")
		}
		zone = inv.ask("Enter zone (e.g. us-central1-a): ")
		if zone == "" {
			return "", "", fmt.Errorf("zone is required")
		}
	}
	return project, zone, nil
}

// ask writes prompt to stdout and returns the first word of the reply.
func (inv *invocation) ask(prompt string) string {
	fmt.Fprint(inv.stdout, prompt)
	var answer string
	fmt.Fscanln(inv.stdin, &answer)
	return answer
}

// httpDebugWriter returns where outgoing REST calls are logged as curl
// commands, or nil when that is off. --verbosity=debug turns it on, as does
// --log-http. Output goes to stderr so it never mixes into --format output.
func (inv *invocation) httpDebugWriter() io.Writer {
	if inv.flags.logHTTP || strings.EqualFold(inv.flags.verbosity, "debug") {
		return inv.stderr
	}
	return nil
}

// clients returns a factory for Google API clients authenticated as the
// invocation's account.
func (inv *invocation) clients() *gcp.Clients {
	return &gcp.Clients{
		Tokens:   inv.tokens,
		Account:  inv.flags.account,
		DebugOut: inv.httpDebugWriter(),
		Options:  inv.clientOptions,
	}
}

// confirm asks the user a yes/no question on stderr and returns the answer,
// or dflt without asking when --quiet is set or the invocation is not
// interactive.
func (inv *invocation) confirm(text string, dflt bool) bool {
	return confirm(inv.stdin, inv.stderr, inv.interactive && !inv.flags.quiet, text, dflt)
}

// context returns the context the invocation runs under.
func (inv *invocation) context() context.Context {
	if inv.ctx == nil {
		return context.Background()
	}
	return inv.ctx
}

// formatOutput is formatOutputTo writing to the invocation's stdout.
func (inv *invocation) formatOutput(v any, format string) error {
	return formatOutputTo(inv.stdout, v, format)
}

// promptContinue is promptContinue for the invocation: it uses --quiet and
// the invocation's streams.
func (inv *invocation) promptContinue(message, cancelString string) error {
	if inv.flags.quiet || !inv.interactive {
		return nil
	}
	return promptContinueIO(inv.stdin, inv.stderr, message, cancelString)
}

// emit writes v to stdout in the --format output format.
func (inv *invocation) emit(v any) error {
	return emitFormattedTo(inv.stdout, v, inv.flags.format)
}

// yaml writes v to stdout as YAML.
func (inv *invocation) yaml(v any) error {
	return emitFormattedTo(inv.stdout, v, "yaml")
}

// printJSON writes v to stdout as indented JSON.
func (inv *invocation) printJSON(v any) error {
	return emitJSON(inv.stdout, v)
}

// restClient returns a REST client for endpoint that uses the invocation's
// credentials, billing project and HTTP logging.
func (inv *invocation) restClient(endpoint string) *restClient {
	return inv.rest(newRESTClient(endpoint))
}

// rest returns a copy of c that uses the invocation's credentials, billing
// project and HTTP logging.
func (inv *invocation) rest(c *restClient) *restClient {
	return &restClient{endpoint: c.endpoint, inv: inv}
}

// billingProject is resolveBillingProject for the invocation: the project
// from --billing-project, billing/quota_project or the ADC file's
// quota_project_id, or "".
func (inv *invocation) billingProject() string {
	if inv.flags.billingProject != "" {
		return inv.flags.billingProject
	}
	if props, err := inv.properties(); err == nil && props != nil && props.Billing.QuotaProject != "" {
		return props.Billing.QuotaProject
	}
	return quotaProjectFromADC(inv.adcFilePath())
}

// quotaProject returns the billing / quota project for the invocation's
// REST calls.
func (inv *invocation) quotaProject() string {
	if inv.restUserProject != nil {
		return inv.restUserProject()
	}
	return inv.billingProject()
}

// adcFilePath is adcFilePath for the invocation's environment.
func (inv *invocation) adcFilePath() string {
	configDir := inv.env.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, _ := inv.env.HomeDir()
		configDir = filepath.Join(home, ".config", "gcloud")
	}
	return filepath.Join(configDir, "application_default_credentials.json")
}
