package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/secrets"
	"github.com/spf13/cobra"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage Secret Manager secrets",
}

var secretsVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage secret versions",
}

// --- secrets versions access ---

var secretsVersionsAccessCmd = &cobra.Command{
	Use:   "access VERSION",
	Short: "Access a secret version's data",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsAccess,
}

var (
	flagSecretName string
	flagOutFile    string
)

// --- secrets create ---

var secretsCreateCmd = &cobra.Command{
	Use:   "create SECRET_ID",
	Short: "Create a new secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsCreate,
}

var (
	flagDataFile          string
	flagReplicationPolicy string
	flagLocations         []string
	flagSecretLabels      map[string]string
	flagExpireTime        string
	flagTTL               string
)

// --- secrets versions add ---

var secretsVersionsAddCmd = &cobra.Command{
	Use:   "add SECRET_ID",
	Short: "Add a new version to a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsAdd,
}

var flagVersionsAddDataFile string

// --- secrets list ---

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Args:  cobra.NoArgs,
	RunE:  runSecretsList,
}

var (
	flagSecretsFilter string
	flagSecretsFormat string
)

// --- secrets describe ---

var secretsDescribeCmd = &cobra.Command{
	Use:   "describe SECRET_ID",
	Short: "Describe a secret's metadata",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsDescribe,
}

// --- secrets delete ---

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete SECRET_ID",
	Short: "Delete a secret and all its versions",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsDelete,
}

var (
	flagQuiet           bool
	flagSecretsLocation string
)

func init() {
	// secrets versions access
	secretsVersionsAccessCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsAccessCmd.Flags().StringVar(&flagOutFile, "out-file", "", "Output file path")
	secretsVersionsAccessCmd.MarkFlagRequired("secret")

	// secrets create
	secretsCreateCmd.Flags().StringVar(&flagDataFile, "data-file", "", "File with secret data, or - for stdin")
	secretsCreateCmd.Flags().StringVar(&flagReplicationPolicy, "replication-policy", "automatic", "Replication policy (automatic or user-managed)")
	secretsCreateCmd.Flags().StringSliceVar(&flagLocations, "locations", nil, "Locations for user-managed replication")
	secretsCreateCmd.Flags().StringToStringVar(&flagSecretLabels, "labels", nil, "Labels as key=value pairs")
	secretsCreateCmd.Flags().StringVar(&flagExpireTime, "expire-time", "", "Expiration time (RFC 3339 format)")
	secretsCreateCmd.Flags().StringVar(&flagTTL, "ttl", "", "Time-to-live duration (e.g. 30d, 24h)")

	// secrets versions add
	secretsVersionsAddCmd.Flags().StringVar(&flagVersionsAddDataFile, "data-file", "", "File with secret data, or - for stdin (required)")
	secretsVersionsAddCmd.MarkFlagRequired("data-file")

	// secrets list
	secretsListCmd.Flags().StringVar(&flagSecretsFilter, "filter", "", "Filter expression")
	secretsListCmd.Flags().StringVar(&flagSecretsFormat, "format", "", "Output format (e.g. json, 'get(name)')")

	// secrets delete
	secretsDeleteCmd.Flags().BoolVar(&flagQuiet, "quiet", false, "Suppress confirmation prompt")

	// --location on all subcommands
	secretsVersionsAccessCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsAddCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsCreateCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsListCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsDescribeCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsDeleteCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")

	// Wire up command tree.
	secretsVersionsCmd.AddCommand(secretsVersionsAccessCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsAddCmd)
	secretsCmd.AddCommand(secretsVersionsCmd)
	secretsCmd.AddCommand(secretsCreateCmd)
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsDescribeCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)
	rootCmd.AddCommand(secretsCmd)
}

func resolveProject() (string, error) {
	props, err := config.Load()
	if err != nil {
		return "", err
	}
	project := config.Resolve(flagProject, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	if project == "" {
		return "", fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
	}
	return project, nil
}

// --- Command implementations ---

func runSecretsVersionsAccess(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, flagSecretName, version, flagSecretsLocation)
	resp, err := svc.Projects.Secrets.Versions.Access(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("accessing secret version: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(resp.Payload.Data)
	if err != nil {
		return fmt.Errorf("decoding secret data: %w", err)
	}

	if flagOutFile != "" {
		if err := os.WriteFile(flagOutFile, data, 0600); err != nil {
			return fmt.Errorf("writing to %s: %w", flagOutFile, err)
		}
		return nil
	}

	os.Stdout.Write(data)
	return nil
}

func runSecretsCreate(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Create the secret with configured replication.
	secret := &secretmanager.Secret{}
	if flagReplicationPolicy == "user-managed" && len(flagLocations) > 0 {
		var replicas []*secretmanager.Replica
		for _, loc := range flagLocations {
			replicas = append(replicas, &secretmanager.Replica{Location: loc})
		}
		secret.Replication = &secretmanager.Replication{
			UserManaged: &secretmanager.UserManaged{Replicas: replicas},
		}
	} else {
		secret.Replication = &secretmanager.Replication{
			Automatic: &secretmanager.Automatic{},
		}
	}
	if len(flagSecretLabels) > 0 {
		secret.Labels = flagSecretLabels
	}
	if flagExpireTime != "" {
		secret.ExpireTime = flagExpireTime
	}
	if flagTTL != "" {
		secret.Ttl = flagTTL
	}

	parent := secrets.SecretParent(project, flagSecretsLocation)
	created, err := svc.Projects.Secrets.Create(parent, secret).SecretId(secretID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}

	fmt.Printf("Created secret [%s].\n", created.Name)

	// If --data-file provided, add initial version.
	if flagDataFile != "" {
		data, err := readDataFile(flagDataFile)
		if err != nil {
			return err
		}
		if err := addVersion(ctx, svc, created.Name, data); err != nil {
			return err
		}
	}

	return nil
}

func runSecretsVersionsAdd(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	data, err := readDataFile(flagVersionsAddDataFile)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, flagSecretsLocation)
	return addVersion(ctx, svc, name, data)
}

func runSecretsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := secrets.SecretParent(project, flagSecretsLocation)
	call := svc.Projects.Secrets.List(parent).Context(ctx)
	if flagSecretsFilter != "" {
		call = call.Filter(flagSecretsFilter)
	}

	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	// Handle --format.
	if flagSecretsFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Secrets)
	}

	if strings.HasPrefix(flagSecretsFormat, "get(") && strings.HasSuffix(flagSecretsFormat, ")") {
		field := flagSecretsFormat[4 : len(flagSecretsFormat)-1]
		for _, s := range resp.Secrets {
			switch field {
			case "name":
				fmt.Println(s.Name)
			case "createTime":
				fmt.Println(s.CreateTime)
			default:
				fmt.Println(s.Name)
			}
		}
		return nil
	}

	// Default table format.
	fmt.Printf("%-60s %s\n", "NAME", "CREATED")
	for _, s := range resp.Secrets {
		fmt.Printf("%-60s %s\n", s.Name, s.CreateTime)
	}
	return nil
}

func runSecretsDescribe(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, flagSecretsLocation)
	secret, err := svc.Projects.Secrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(secret)
}

func runSecretsDelete(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to delete secret [%s]. This action cannot be undone.\n", secretID)
		fmt.Print("Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, flagSecretsLocation)
	if _, err := svc.Projects.Secrets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}

	fmt.Printf("Deleted secret [%s].\n", secretID)
	return nil
}

// --- Helpers ---

func readDataFile(path string) ([]byte, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return data, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading data file %s: %w", path, err)
	}
	return data, nil
}

func addVersion(ctx context.Context, svc *secretmanager.Service, parent string, data []byte) error {
	req := &secretmanager.AddSecretVersionRequest{
		Payload: &secretmanager.SecretPayload{
			Data: base64.StdEncoding.EncodeToString(data),
		},
	}
	ver, err := svc.Projects.Secrets.AddVersion(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding secret version: %w", err)
	}
	fmt.Printf("Created version [%s].\n", ver.Name)
	return nil
}
