package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

const artifactRegistryExtensionVersion = "2.2.0"

var (
	flagArtPrintLocation   string
	flagArtPrintRepository string
	flagArtPrintJSONKey    string
	flagArtPrintNpmScope   string
)

var artifactsPrintSettingsCmd = &cobra.Command{
	Use:   "print-settings",
	Short: "Print snippets to add to native tools settings files",
	Long: `The snippets provide a credentials placeholder and configurations to allow
native tools to interact with Artifact Registry repositories.`,
}

var artifactsPrintSettingsGradleCmd = &cobra.Command{
	Use:   "gradle",
	Short: "Print a snippet to add a repository to the Gradle build.gradle file",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsPrintSettingsGradle,
}

var artifactsPrintSettingsMvnCmd = &cobra.Command{
	Use:   "mvn",
	Short: "Print a snippet to add a Maven repository to the pom.xml file",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsPrintSettingsMvn,
}

var artifactsPrintSettingsNpmCmd = &cobra.Command{
	Use:   "npm",
	Short: "Print credential settings to add to the .npmrc file",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsPrintSettingsNpm,
}

var artifactsPrintSettingsPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Print credential settings to add to the .pypirc and pip.conf files",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsPrintSettingsPython,
}

func init() {
	subs := []*cobra.Command{
		artifactsPrintSettingsGradleCmd,
		artifactsPrintSettingsMvnCmd,
		artifactsPrintSettingsNpmCmd,
		artifactsPrintSettingsPythonCmd,
	}
	for _, c := range subs {
		c.Flags().StringVar(&flagArtPrintLocation, "location", "", "Location of the repository")
		c.Flags().StringVar(&flagArtPrintRepository, "repository", "", "ID of the repository")
		c.Flags().StringVar(&flagArtPrintJSONKey, "json-key", "", "Path to service account JSON key")
	}
	artifactsPrintSettingsNpmCmd.Flags().StringVar(&flagArtPrintNpmScope, "scope", "",
		"Scope to associate with the Artifact Registry registry (must start with '@')")

	for _, c := range subs {
		artifactsPrintSettingsCmd.AddCommand(c)
	}
	artifactsCmd.AddCommand(artifactsPrintSettingsCmd)
}

// printSettingsInputs holds the resolved user inputs for a print-settings run.
type printSettingsInputs struct {
	project    string
	location   string
	repository string
	jsonKey    string
}

func resolvePrintSettingsInputs() (*printSettingsInputs, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, err
	}
	location := config.Resolve(flagArtPrintLocation, "CLOUDSDK_ARTIFACTS_LOCATION", "")
	if location == "" {
		return nil, fmt.Errorf("Failed to find attribute [location]. " +
			"The attribute can be set in the following ways:\n" +
			"- provide the argument [--location] on the command line\n" +
			"- set the environment variable [CLOUDSDK_ARTIFACTS_LOCATION]")
	}
	repository := config.Resolve(flagArtPrintRepository, "CLOUDSDK_ARTIFACTS_REPOSITORY", "")
	if repository == "" {
		return nil, fmt.Errorf("Failed to find attribute [repository]. " +
			"The attribute can be set in the following ways:\n" +
			"- provide the argument [--repository] on the command line\n" +
			"- set the environment variable [CLOUDSDK_ARTIFACTS_REPOSITORY]")
	}
	return &printSettingsInputs{
		project:    project,
		location:   location,
		repository: repository,
		jsonKey:    flagArtPrintJSONKey,
	}, nil
}

// fetchRepository loads the Repository resource and verifies its format matches
// the expected value (case-insensitive), matching gcloud-python's behaviour.
func fetchRepository(ctx context.Context, in *printSettingsInputs, expectFormat string) (*artifactregistry.Repository, error) {
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", in.project, in.location, in.repository)
	repo, err := svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching repository %s: %w", name, err)
	}
	if !strings.EqualFold(repo.Format, expectFormat) {
		return nil, fmt.Errorf("Invalid repository type %s. Valid type is %s.", repo.Format, expectFormat)
	}
	return repo, nil
}

// getServiceAccountCreds returns base64-encoded service account JSON credentials
// or "" when no service account credentials are available. When jsonKey is set,
// it is read from the given file; otherwise the active account credentials are
// consulted and only returned when they are of type service_account.
func getServiceAccountCreds(jsonKey string) (string, error) {
	if jsonKey != "" {
		content, err := os.ReadFile(jsonKey)
		if err != nil {
			return "", fmt.Errorf("reading JSON key file %s: %w", jsonKey, err)
		}
		var probe map[string]any
		if err := json.Unmarshal(content, &probe); err != nil {
			return "", fmt.Errorf("could not read JSON file %s: %w", jsonKey, err)
		}
		return base64.StdEncoding.EncodeToString(content), nil
	}

	account := flagAccount
	if account == "" {
		props, err := config.Load()
		if err == nil {
			account = props.Core.Account
		}
	}
	if account == "" {
		return "", nil
	}
	store, err := auth.NewStore()
	if err != nil {
		return "", nil
	}
	data, err := store.Load(account)
	if err != nil {
		return "", nil
	}
	var probe map[string]any
	if err := json.Unmarshal(data, &probe); err != nil {
		return "", nil
	}
	if probe["type"] != "service_account" {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// --- gradle ---

func runArtifactsPrintSettingsGradle(cmd *cobra.Command, args []string) error {
	in, err := resolvePrintSettingsInputs()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo, err := fetchRepository(ctx, in, "MAVEN")
	if err != nil {
		return err
	}
	saCreds, err := getServiceAccountCreds(in.jsonKey)
	if err != nil {
		return err
	}
	fmt.Print(renderGradleSnippet(in, repo.MavenConfig, saCreds))
	return nil
}

func renderGradleSnippet(in *printSettingsInputs, mavenCfg *artifactregistry.MavenRepositoryConfig, saCreds string) string {
	repoPath := in.project + "/" + in.repository
	tmpl := selectGradleTemplate(mavenCfg, saCreds)
	repl := map[string]string{
		"{location}":          in.location,
		"{repo_path}":         repoPath,
		"{extension_version}": artifactRegistryExtensionVersion,
	}
	if saCreds != "" {
		repl["{username}"] = "_json_key_base64"
		repl["{password}"] = saCreds
	}
	return applyReplacements(tmpl, repl)
}

func selectGradleTemplate(cfg *artifactregistry.MavenRepositoryConfig, saCreds string) string {
	policy := ""
	if cfg != nil {
		policy = cfg.VersionPolicy
	}
	switch {
	case policy == "SNAPSHOT" && saCreds != "":
		return gradleServiceAccountSnapshotTemplate
	case policy == "SNAPSHOT":
		return gradleNoServiceAccountSnapshotTemplate
	case policy == "RELEASE" && saCreds != "":
		return gradleServiceAccountReleaseTemplate
	case policy == "RELEASE":
		return gradleNoServiceAccountReleaseTemplate
	case saCreds != "":
		return gradleServiceAccountTemplate
	default:
		return gradleNoServiceAccountTemplate
	}
}

// --- mvn ---

func runArtifactsPrintSettingsMvn(cmd *cobra.Command, args []string) error {
	in, err := resolvePrintSettingsInputs()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo, err := fetchRepository(ctx, in, "MAVEN")
	if err != nil {
		return err
	}
	saCreds, err := getServiceAccountCreds(in.jsonKey)
	if err != nil {
		return err
	}
	fmt.Print(renderMavenSnippet(in, repo.MavenConfig, saCreds))
	return nil
}

func renderMavenSnippet(in *printSettingsInputs, mavenCfg *artifactregistry.MavenRepositoryConfig, saCreds string) string {
	repoPath := in.project + "/" + in.repository
	tmpl := selectMavenTemplate(mavenCfg, saCreds)
	scheme := "artifactregistry"
	repl := map[string]string{
		"{location}":          in.location,
		"{server_id}":         "artifact-registry",
		"{repo_path}":         repoPath,
		"{extension_version}": artifactRegistryExtensionVersion,
	}
	if saCreds != "" {
		scheme = "https"
		repl["{username}"] = "_json_key_base64"
		repl["{password}"] = saCreds
	}
	repl["{scheme}"] = scheme
	return applyReplacements(tmpl, repl)
}

func selectMavenTemplate(cfg *artifactregistry.MavenRepositoryConfig, saCreds string) string {
	policy := ""
	if cfg != nil {
		policy = cfg.VersionPolicy
	}
	switch {
	case policy == "SNAPSHOT" && saCreds != "":
		return mvnServiceAccountSnapshotTemplate
	case policy == "SNAPSHOT":
		return mvnNoServiceAccountSnapshotTemplate
	case policy == "RELEASE" && saCreds != "":
		return mvnServiceAccountReleaseTemplate
	case policy == "RELEASE":
		return mvnNoServiceAccountReleaseTemplate
	case saCreds != "":
		return mvnServiceAccountTemplate
	default:
		return mvnNoServiceAccountTemplate
	}
}

// --- npm ---

func runArtifactsPrintSettingsNpm(cmd *cobra.Command, args []string) error {
	in, err := resolvePrintSettingsInputs()
	if err != nil {
		return err
	}
	scope := flagArtPrintNpmScope
	if scope != "" && (!strings.HasPrefix(scope, "@") || len(scope) <= 1) {
		return fmt.Errorf(`Scope name must start with "@" and be longer than 1 character.`)
	}
	ctx := context.Background()
	if _, err := fetchRepository(ctx, in, "NPM"); err != nil {
		return err
	}
	saCreds, err := getServiceAccountCreds(in.jsonKey)
	if err != nil {
		return err
	}
	fmt.Print(renderNpmSnippet(in, scope, saCreds))
	return nil
}

func renderNpmSnippet(in *printSettingsInputs, scope, saCreds string) string {
	repoPath := in.project + "/" + in.repository
	registryPath := fmt.Sprintf("%s-npm.pkg.dev/%s/", in.location, repoPath)
	configuredRegistry := "registry"
	if scope != "" {
		configuredRegistry = scope + ":" + configuredRegistry
	}
	tmpl := npmNoServiceAccountTemplate
	repl := map[string]string{
		"{configured_registry}": configuredRegistry,
		"{registry_path}":       registryPath,
		"{repo_path}":           repoPath,
	}
	if saCreds != "" {
		tmpl = npmServiceAccountTemplate
		repl["{password}"] = saCreds
	}
	return applyReplacements(tmpl, repl)
}

// --- python ---

func runArtifactsPrintSettingsPython(cmd *cobra.Command, args []string) error {
	in, err := resolvePrintSettingsInputs()
	if err != nil {
		return err
	}
	ctx := context.Background()
	if _, err := fetchRepository(ctx, in, "PYTHON"); err != nil {
		return err
	}
	saCreds, err := getServiceAccountCreds(in.jsonKey)
	if err != nil {
		return err
	}
	fmt.Print(renderPythonSnippet(in, saCreds))
	return nil
}

func renderPythonSnippet(in *printSettingsInputs, saCreds string) string {
	repoPath := in.project + "/" + in.repository
	tmpl := pythonNoServiceAccountTemplate
	repl := map[string]string{
		"{location}":  in.location,
		"{repo_path}": repoPath,
		"{repo}":      in.repository,
	}
	if saCreds != "" {
		tmpl = pythonServiceAccountTemplate
		repl["{password}"] = saCreds
	}
	return applyReplacements(tmpl, repl)
}

// applyReplacements performs literal substitution of {key} tokens in tmpl
// using values from repl. Only tokens whose body matches [A-Za-z_][A-Za-z0-9_]*
// are considered; other "{...}" sequences (for example Groovy/Gradle blocks)
// are left untouched. Substitution is done in a single pass so that
// replacement values that themselves contain "{...}" (e.g. base64-encoded
// JSON keys) are not re-expanded.
func applyReplacements(tmpl string, repl map[string]string) string {
	var b strings.Builder
	b.Grow(len(tmpl))
	i := 0
	for i < len(tmpl) {
		if tmpl[i] != '{' {
			b.WriteByte(tmpl[i])
			i++
			continue
		}
		// Scan a candidate identifier body.
		j := i + 1
		for j < len(tmpl) && isTokenByte(tmpl[j]) {
			j++
		}
		if j > i+1 && j < len(tmpl) && tmpl[j] == '}' {
			token := tmpl[i : j+1]
			if val, ok := repl[token]; ok {
				b.WriteString(val)
			} else {
				b.WriteString(token)
			}
			i = j + 1
			continue
		}
		// Not a substitutable token — emit the `{` verbatim and keep scanning.
		b.WriteByte('{')
		i++
	}
	return b.String()
}

func isTokenByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z':
		return true
	case c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c == '_':
		return true
	}
	return false
}
