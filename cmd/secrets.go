package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
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

// --- secrets versions list ---

var secretsVersionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret versions",
	Args:  cobra.NoArgs,
	RunE:  runSecretsVersionsList,
}

var (
	flagVersionsListFormat string
	flagVersionsListURI    bool
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
	flagSecretsURI    bool
)

// --- secrets describe ---

var secretsDescribeCmd = &cobra.Command{
	Use:   "describe SECRET_ID",
	Short: "Describe a secret's metadata",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsDescribe,
}

// --- secrets versions describe ---

var secretsVersionsDescribeCmd = &cobra.Command{
	Use:   "describe VERSION",
	Short: "Describe a secret version's metadata",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsDescribe,
}

// --- secrets versions disable ---

var secretsVersionsDisableCmd = &cobra.Command{
	Use:   "disable VERSION",
	Short: "Disable a secret version",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsDisable,
}

// --- secrets versions enable ---

var secretsVersionsEnableCmd = &cobra.Command{
	Use:   "enable VERSION",
	Short: "Enable a secret version",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsEnable,
}

// --- secrets versions destroy ---

var secretsVersionsDestroyCmd = &cobra.Command{
	Use:   "destroy VERSION",
	Short: "Destroy a secret version's data",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsVersionsDestroy,
}

// --- secrets update ---

var secretsUpdateCmd = &cobra.Command{
	Use:   "update SECRET_ID",
	Short: "Update a secret's properties",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsUpdate,
}

var (
	flagSecretUpdateLabels      map[string]string
	flagSecretRemoveLabels      []string
	flagSecretClearLabels       bool
	flagSecretUpdateExpireTime  string
	flagSecretUpdateTTL         string
	flagSecretRemoveExpiration  bool
	flagSecretNextRotationTime  string
	flagSecretRotationPeriod    string
)

// --- secrets delete ---

var secretsDeleteCmd = &cobra.Command{
	Use:   "delete SECRET_ID",
	Short: "Delete a secret and all its versions",
	Args:  cobra.ExactArgs(1),
	RunE:  runSecretsDelete,
}

var (
	flagSecretsLocation string
	flagDeleteEtag      string
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

	// secrets versions list
	secretsVersionsListCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsListCmd.MarkFlagRequired("secret")
	secretsVersionsListCmd.Flags().StringVar(&flagVersionsListFormat, "format", "", "Output format (e.g. json)")
	secretsVersionsListCmd.Flags().BoolVar(&flagVersionsListURI, "uri", false, "Print resource names")
	secretsVersionsListCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")

	// secrets versions add
	secretsVersionsAddCmd.Flags().StringVar(&flagVersionsAddDataFile, "data-file", "", "File with secret data, or - for stdin (required)")
	secretsVersionsAddCmd.MarkFlagRequired("data-file")

	// secrets list
	secretsListCmd.Flags().StringVar(&flagSecretsFilter, "filter", "", "Filter expression")
	secretsListCmd.Flags().StringVar(&flagSecretsFormat, "format", "", "Output format (e.g. json, 'get(name)')")
	secretsListCmd.Flags().BoolVar(&flagSecretsURI, "uri", false, "Print resource names")

	// secrets versions describe/disable/enable/destroy
	secretsVersionsDescribeCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsDescribeCmd.MarkFlagRequired("secret")
	secretsVersionsDisableCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsDisableCmd.MarkFlagRequired("secret")
	secretsVersionsEnableCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsEnableCmd.MarkFlagRequired("secret")
	secretsVersionsDestroyCmd.Flags().StringVar(&flagSecretName, "secret", "", "Secret name (required)")
	secretsVersionsDestroyCmd.MarkFlagRequired("secret")
	// --quiet is provided by the global persistent flag

	// secrets update
	secretsUpdateCmd.Flags().StringToStringVar(&flagSecretUpdateLabels, "update-labels", nil, "Labels to update (key=value)")
	secretsUpdateCmd.Flags().StringSliceVar(&flagSecretRemoveLabels, "remove-labels", nil, "Labels to remove")
	secretsUpdateCmd.Flags().BoolVar(&flagSecretClearLabels, "clear-labels", false, "Remove all labels")
	secretsUpdateCmd.Flags().StringVar(&flagSecretUpdateExpireTime, "expire-time", "", "Expiration time (RFC 3339)")
	secretsUpdateCmd.Flags().StringVar(&flagSecretUpdateTTL, "ttl", "", "Time-to-live duration")
	secretsUpdateCmd.Flags().BoolVar(&flagSecretRemoveExpiration, "remove-expiration", false, "Remove expiration")
	secretsUpdateCmd.Flags().StringVar(&flagSecretNextRotationTime, "next-rotation-time", "", "Next rotation time (RFC 3339)")
	secretsUpdateCmd.Flags().StringVar(&flagSecretRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d)")

	// secrets delete
	// --quiet is provided by the global persistent flag
	secretsDeleteCmd.Flags().StringVar(&flagDeleteEtag, "etag", "", "Etag for optimistic concurrency")

	// --location on all subcommands
	secretsVersionsAccessCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsAddCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsDescribeCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsDisableCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsEnableCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsVersionsDestroyCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsCreateCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsListCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsDescribeCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsDeleteCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	secretsUpdateCmd.Flags().StringVar(&flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")

	// Wire up command tree.
	secretsVersionsCmd.AddCommand(secretsVersionsAccessCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsListCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsAddCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsDescribeCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsDisableCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsEnableCmd)
	secretsVersionsCmd.AddCommand(secretsVersionsDestroyCmd)
	secretsCmd.AddCommand(secretsVersionsCmd)
	secretsCmd.AddCommand(secretsCreateCmd)
	secretsCmd.AddCommand(secretsListCmd)
	secretsCmd.AddCommand(secretsDescribeCmd)
	secretsCmd.AddCommand(secretsDeleteCmd)
	secretsCmd.AddCommand(secretsUpdateCmd)
	rootCmd.AddCommand(secretsCmd)
}

func resolveProject() (string, error) {
	props, err := config.Load()
	if err != nil {
		return "", err
	}
	project := config.Resolve(flagProject, "CLOUDSDK_CORE_PROJECT", props.Core.Project)
	if project == "" {
		if !IsInteractive() {
			return "", fmt.Errorf("project is required; set via --project flag, CLOUDSDK_CORE_PROJECT env, or config")
		}
		fmt.Print("Enter project: ")
		fmt.Scanln(&project)
		if project == "" {
			return "", fmt.Errorf("project is required")
		}
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

	if resp.Payload.DataCrc32c != 0 {
		crc := crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
		if int64(crc) != resp.Payload.DataCrc32c {
			return fmt.Errorf("secret data CRC32C mismatch: expected %d, got %d", resp.Payload.DataCrc32c, crc)
		}
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

func runSecretsVersionsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := secrets.SecretName(project, flagSecretName, flagSecretsLocation)
	var allVersions []*secretmanager.SecretVersion
	pageToken := ""
	for {
		call := svc.Projects.Secrets.Versions.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing secret versions: %w", err)
		}
		allVersions = append(allVersions, resp.Versions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagVersionsListURI {
		for _, v := range allVersions {
			fmt.Println(v.Name)
		}
		return nil
	}

	if flagVersionsListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allVersions)
	}

	fmt.Printf("%-60s %-12s %-25s\n", "NAME", "STATE", "CREATED")
	for _, v := range allVersions {
		fmt.Printf("%-60s %-12s %-25s\n", v.Name, v.State, v.CreateTime)
	}
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

	// Match Python gcloud: when --data-file is supplied, emit only the
	// version-created message; otherwise emit the secret-created message.
	// Both messages go to stderr via secrets_log.
	if flagDataFile != "" {
		data, err := readDataFile(flagDataFile)
		if err != nil {
			return err
		}
		return addVersion(ctx, svc, created.Name, data)
	}
	fmt.Fprintf(os.Stderr, "Created secret [%s].\n", secretID)
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

	var allSecrets []*secretmanager.Secret
	pageToken := ""
	for {
		call := svc.Projects.Secrets.List(parent).Context(ctx)
		if flagSecretsFilter != "" {
			call = call.Filter(flagSecretsFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing secrets: %w", err)
		}
		allSecrets = append(allSecrets, resp.Secrets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// Handle --uri.
	if flagSecretsURI {
		for _, secret := range allSecrets {
			fmt.Println(secret.Name)
		}
		return nil
	}

	// Handle --format.
	if flagSecretsFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allSecrets)
	}

	if strings.HasPrefix(flagSecretsFormat, "get(") && strings.HasSuffix(flagSecretsFormat, ")") {
		field := flagSecretsFormat[4 : len(flagSecretsFormat)-1]
		for _, s := range allSecrets {
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
	for _, s := range allSecrets {
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

func runSecretsVersionsDescribe(cmd *cobra.Command, args []string) error {
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
	v, err := svc.Projects.Secrets.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret version: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func runSecretsVersionsDisable(cmd *cobra.Command, args []string) error {
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
	req := &secretmanager.DisableSecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Disable(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling secret version: %w", err)
	}

	fmt.Printf("Disabled version [%s].\n", v.Name)
	return nil
}

func runSecretsVersionsEnable(cmd *cobra.Command, args []string) error {
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
	req := &secretmanager.EnableSecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Enable(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling secret version: %w", err)
	}

	fmt.Printf("Enabled version [%s].\n", v.Name)
	return nil
}

func runSecretsVersionsDestroy(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Printf("You are about to destroy version [%s] of secret [%s]. This action cannot be undone.\n", version, flagSecretName)
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

	name := secrets.VersionName(project, flagSecretName, version, flagSecretsLocation)
	req := &secretmanager.DestroySecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Destroy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("destroying secret version: %w", err)
	}

	fmt.Printf("Destroyed version [%s].\n", v.Name)
	return nil
}

func runSecretsUpdate(cmd *cobra.Command, args []string) error {
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

	// Get current secret to merge labels.
	current, err := svc.Projects.Secrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting secret: %w", err)
	}

	secret := &secretmanager.Secret{}
	var updateMask []string

	// Handle labels.
	if flagSecretClearLabels {
		secret.Labels = map[string]string{}
		updateMask = append(updateMask, "labels")
	} else if len(flagSecretUpdateLabels) > 0 || len(flagSecretRemoveLabels) > 0 {
		labels := make(map[string]string)
		for k, v := range current.Labels {
			labels[k] = v
		}
		for _, k := range flagSecretRemoveLabels {
			delete(labels, k)
		}
		for k, v := range flagSecretUpdateLabels {
			labels[k] = v
		}
		secret.Labels = labels
		updateMask = append(updateMask, "labels")
	}

	// Handle expiration.
	if flagSecretRemoveExpiration {
		secret.ExpireTime = ""
		secret.Ttl = ""
		updateMask = append(updateMask, "expire_time", "ttl")
	} else {
		if flagSecretUpdateExpireTime != "" {
			secret.ExpireTime = flagSecretUpdateExpireTime
			updateMask = append(updateMask, "expire_time")
		}
		if flagSecretUpdateTTL != "" {
			secret.Ttl = flagSecretUpdateTTL
			updateMask = append(updateMask, "ttl")
		}
	}

	// Handle rotation.
	if flagSecretNextRotationTime != "" || flagSecretRotationPeriod != "" {
		secret.Rotation = &secretmanager.Rotation{}
		if flagSecretNextRotationTime != "" {
			secret.Rotation.NextRotationTime = flagSecretNextRotationTime
		}
		if flagSecretRotationPeriod != "" {
			secret.Rotation.RotationPeriod = flagSecretRotationPeriod
		}
		updateMask = append(updateMask, "rotation")
	}

	if len(updateMask) == 0 {
		return fmt.Errorf("no update flags specified")
	}

	updated, err := svc.Projects.Secrets.Patch(name, secret).UpdateMask(strings.Join(updateMask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating secret: %w", err)
	}

	fmt.Printf("Updated secret [%s].\n", updated.Name)
	return nil
}

func runSecretsDelete(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := resolveProject()
	if err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "You are about to delete secret [%s]. This action cannot be undone.\n", secretID)
		fmt.Fprint(os.Stderr, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, flagSecretsLocation)
	call := svc.Projects.Secrets.Delete(name).Context(ctx)
	if flagDeleteEtag != "" {
		call = call.Etag(flagDeleteEtag)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Deleted secret [%s].\n", secretID)
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
	// Emit "Created version [N] of the secret [ID]." to stderr, matching
	// secrets_log.Versions().Created() in the Python reference.
	fmt.Fprintf(os.Stderr, "Created version [%s] of the secret [%s].\n",
		lastPathSegment(ver.Name), lastPathSegment(parent))
	return nil
}

// lastPathSegment returns the substring after the final "/" in name, or name
// itself when there is no "/".
func lastPathSegment(name string) string {
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return name
}
