package cmd

import (
	"context"
	"encoding/base64"
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

// --- secrets versions access ---

// --- secrets create ---

// --- secrets versions list ---

// --- secrets versions add ---

// --- secrets list ---

// --- secrets describe ---

// --- secrets versions describe ---

// --- secrets versions disable ---

// --- secrets versions enable ---

// --- secrets versions destroy ---

// --- secrets update ---

// --- secrets delete ---

func (g *secretsCmds) initSecrets() {
	// secrets versions access
	g.secretsVersionsAccessCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsAccessCmd.Flags().StringVar(&g.flagOutFile, "out-file", "", "Output file path")
	g.secretsVersionsAccessCmd.MarkFlagRequired("secret")

	// secrets create
	g.secretsCreateCmd.Flags().StringVar(&g.flagDataFile, "data-file", "", "File with secret data, or - for stdin")
	g.secretsCreateCmd.Flags().StringVar(&g.flagReplicationPolicy, "replication-policy", "automatic", "Replication policy (automatic or user-managed)")
	g.secretsCreateCmd.Flags().StringSliceVar(&g.flagLocations, "locations", nil, "Locations for user-managed replication")
	g.secretsCreateCmd.Flags().StringToStringVar(&g.flagSecretLabels, "labels", nil, "Labels as key=value pairs")
	g.secretsCreateCmd.Flags().StringVar(&g.flagExpireTime, "expire-time", "", "Expiration time (RFC 3339 format)")
	g.secretsCreateCmd.Flags().StringVar(&g.flagTTL, "ttl", "", "Time-to-live duration (e.g. 30d, 24h)")
	g.secretsCreateCmd.Flags().StringVar(&g.flagSecretType, "secret-type", "",
		"Type of the secret (e.g. cloud-sql-single-user-credentials, other); leave empty for the default UNSPECIFIED type")

	// secrets versions list
	g.secretsVersionsListCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsListCmd.MarkFlagRequired("secret")
	g.secretsVersionsListCmd.Flags().StringVar(&g.flagVersionsListFormat, "format", "", "Output format (e.g. json)")
	g.secretsVersionsListCmd.Flags().BoolVar(&g.flagVersionsListURI, "uri", false, "Print resource names")
	g.secretsVersionsListCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")

	// secrets versions add
	g.secretsVersionsAddCmd.Flags().StringVar(&g.flagVersionsAddDataFile, "data-file", "", "File with secret data, or - for stdin (required)")
	g.secretsVersionsAddCmd.MarkFlagRequired("data-file")

	// secrets list
	g.secretsListCmd.Flags().StringVar(&g.flagSecretsFilter, "filter", "", "Filter expression")
	g.secretsListCmd.Flags().StringVar(&g.flagSecretsFormat, "format", "", "Output format (e.g. json, 'get(name)')")
	g.secretsListCmd.Flags().BoolVar(&g.flagSecretsURI, "uri", false, "Print resource names")

	// secrets versions describe/disable/enable/destroy
	g.secretsVersionsDescribeCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsDescribeCmd.MarkFlagRequired("secret")
	g.secretsVersionsDisableCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsDisableCmd.MarkFlagRequired("secret")
	g.secretsVersionsEnableCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsEnableCmd.MarkFlagRequired("secret")
	g.secretsVersionsDestroyCmd.Flags().StringVar(&g.flagSecretName, "secret", "", "Secret name (required)")
	g.secretsVersionsDestroyCmd.MarkFlagRequired("secret")
	// --quiet is provided by the global persistent flag

	// secrets update
	g.secretsUpdateCmd.Flags().StringToStringVar(&g.flagSecretUpdateLabels, "update-labels", nil, "Labels to update (key=value)")
	g.secretsUpdateCmd.Flags().StringSliceVar(&g.flagSecretRemoveLabels, "remove-labels", nil, "Labels to remove")
	g.secretsUpdateCmd.Flags().BoolVar(&g.flagSecretClearLabels, "clear-labels", false, "Remove all labels")
	g.secretsUpdateCmd.Flags().StringVar(&g.flagSecretUpdateExpireTime, "expire-time", "", "Expiration time (RFC 3339)")
	g.secretsUpdateCmd.Flags().StringVar(&g.flagSecretUpdateTTL, "ttl", "", "Time-to-live duration")
	g.secretsUpdateCmd.Flags().BoolVar(&g.flagSecretRemoveExpiration, "remove-expiration", false, "Remove expiration")
	g.secretsUpdateCmd.Flags().StringVar(&g.flagSecretNextRotationTime, "next-rotation-time", "", "Next rotation time (RFC 3339)")
	g.secretsUpdateCmd.Flags().StringVar(&g.flagSecretRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d)")

	// secrets delete
	// --quiet is provided by the global persistent flag
	g.secretsDeleteCmd.Flags().StringVar(&g.flagDeleteEtag, "etag", "", "Etag for optimistic concurrency")

	// --location on all subcommands
	g.secretsVersionsAccessCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsVersionsAddCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsVersionsDescribeCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsVersionsDisableCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsVersionsEnableCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsVersionsDestroyCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsCreateCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsListCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsDescribeCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsDeleteCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")
	g.secretsUpdateCmd.Flags().StringVar(&g.flagSecretsLocation, "location", "", "Secret Manager location (for regional secrets)")

	// Wire up command tree.
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsAccessCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsListCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsAddCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsDescribeCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsDisableCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsEnableCmd)
	g.secretsVersionsCmd.AddCommand(g.secretsVersionsDestroyCmd)
	g.secretsCmd.AddCommand(g.secretsVersionsCmd)
	g.secretsCmd.AddCommand(g.secretsCreateCmd)
	g.secretsCmd.AddCommand(g.secretsListCmd)
	g.secretsCmd.AddCommand(g.secretsDescribeCmd)
	g.secretsCmd.AddCommand(g.secretsDeleteCmd)
	g.secretsCmd.AddCommand(g.secretsUpdateCmd)

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

func (g *secretsCmds) runSecretsVersionsAccess(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, g.flagSecretName, version, g.flagSecretsLocation)
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

	if g.flagOutFile != "" {
		if err := os.WriteFile(g.flagOutFile, data, 0600); err != nil {
			return fmt.Errorf("writing to %s: %w", g.flagOutFile, err)
		}
		return nil
	}

	g.inv.stdout.Write(data)
	return nil
}

func (g *secretsCmds) runSecretsVersionsList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	parent := secrets.SecretName(project, g.flagSecretName, g.flagSecretsLocation)
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

	if g.flagVersionsListURI {
		for _, v := range allVersions {
			fmt.Fprintln(g.inv.stdout, v.Name)
		}
		return nil
	}

	if g.flagVersionsListFormat == "json" {
		return g.inv.printJSON(allVersions)
	}

	printTableHeadingTo(g.inv.stdout, len(allVersions), "%-60s %-12s %-25s\n", "NAME", "STATE", "CREATED")
	for _, v := range allVersions {
		fmt.Fprintf(g.inv.stdout, "%-60s %-12s %-25s\n", v.Name, v.State, v.CreateTime)
	}
	return nil
}

func (g *secretsCmds) runSecretsCreate(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	// Create the secret with configured replication.
	secret := &secretmanager.Secret{}
	if g.flagReplicationPolicy == "user-managed" && len(g.flagLocations) > 0 {
		var replicas []*secretmanager.Replica
		for _, loc := range g.flagLocations {
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
	if len(g.flagSecretLabels) > 0 {
		secret.Labels = g.flagSecretLabels
	}
	if g.flagExpireTime != "" {
		secret.ExpireTime = g.flagExpireTime
	}
	if g.flagTTL != "" {
		secret.Ttl = g.flagTTL
	}
	if g.flagSecretType != "" {
		secretType, err := secretTypeEnum(g.flagSecretType)
		if err != nil {
			return err
		}
		secret.SecretType = secretType
	}

	parent := secrets.SecretParent(project, g.flagSecretsLocation)
	created, err := svc.Projects.Secrets.Create(parent, secret).SecretId(secretID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}

	// Match Python gcloud: when --data-file is supplied, emit only the
	// version-created message; otherwise emit the secret-created message.
	// Both messages go to stderr via secrets_log.
	if g.flagDataFile != "" {
		data, err := g.readDataFile(g.flagDataFile)
		if err != nil {
			return err
		}
		return g.addVersion(ctx, svc, created.Name, data)
	}
	fmt.Fprintf(g.inv.stderr, "Created secret [%s].\n", secretID)
	return nil
}

func (g *secretsCmds) runSecretsVersionsAdd(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	data, err := g.readDataFile(g.flagVersionsAddDataFile)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, g.flagSecretsLocation)
	return g.addVersion(ctx, svc, name, data)
}

func (g *secretsCmds) runSecretsList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	parent := secrets.SecretParent(project, g.flagSecretsLocation)

	var allSecrets []*secretmanager.Secret
	pageToken := ""
	for {
		call := svc.Projects.Secrets.List(parent).Context(ctx)
		if g.flagSecretsFilter != "" {
			call = call.Filter(g.flagSecretsFilter)
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
	if g.flagSecretsURI {
		for _, secret := range allSecrets {
			fmt.Fprintln(g.inv.stdout, secret.Name)
		}
		return nil
	}

	// Handle --format.
	if g.flagSecretsFormat == "json" {
		return g.inv.printJSON(allSecrets)
	}

	if strings.HasPrefix(g.flagSecretsFormat, "get(") && strings.HasSuffix(g.flagSecretsFormat, ")") {
		field := g.flagSecretsFormat[4 : len(g.flagSecretsFormat)-1]
		for _, s := range allSecrets {
			switch field {
			case "name":
				fmt.Fprintln(g.inv.stdout, s.Name)
			case "createTime":
				fmt.Fprintln(g.inv.stdout, s.CreateTime)
			default:
				fmt.Fprintln(g.inv.stdout, s.Name)
			}
		}
		return nil
	}

	// Default table format.
	printTableHeadingTo(g.inv.stdout, len(allSecrets), "%-60s %s\n", "NAME", "CREATED")
	for _, s := range allSecrets {
		fmt.Fprintf(g.inv.stdout, "%-60s %s\n", s.Name, s.CreateTime)
	}
	return nil
}

func (g *secretsCmds) runSecretsDescribe(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, g.flagSecretsLocation)
	secret, err := svc.Projects.Secrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret: %w", err)
	}

	return g.inv.printJSON(secret)
}

func (g *secretsCmds) runSecretsVersionsDescribe(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, g.flagSecretName, version, g.flagSecretsLocation)
	v, err := svc.Projects.Secrets.Versions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing secret version: %w", err)
	}

	return g.inv.printJSON(v)
}

func (g *secretsCmds) runSecretsVersionsDisable(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, g.flagSecretName, version, g.flagSecretsLocation)
	req := &secretmanager.DisableSecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Disable(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling secret version: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Disabled version [%s].\n", v.Name)
	return nil
}

func (g *secretsCmds) runSecretsVersionsEnable(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, g.flagSecretName, version, g.flagSecretsLocation)
	req := &secretmanager.EnableSecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Enable(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling secret version: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Enabled version [%s].\n", v.Name)
	return nil
}

func (g *secretsCmds) runSecretsVersionsDestroy(cmd *cobra.Command, args []string) error {
	version := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	if !g.inv.flags.quiet {
		fmt.Fprintf(g.inv.stdout, "You are about to destroy version [%s] of secret [%s]. This action cannot be undone.\n", version, g.flagSecretName)
		fmt.Fprint(g.inv.stdout, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(g.inv.stdout, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.VersionName(project, g.flagSecretName, version, g.flagSecretsLocation)
	req := &secretmanager.DestroySecretVersionRequest{}
	v, err := svc.Projects.Secrets.Versions.Destroy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("destroying secret version: %w", err)
	}

	fmt.Fprintf(g.inv.stdout, "Destroyed version [%s].\n", v.Name)
	return nil
}

func (g *secretsCmds) runSecretsUpdate(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, g.flagSecretsLocation)

	// Get current secret to merge labels.
	current, err := svc.Projects.Secrets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting secret: %w", err)
	}

	secret := &secretmanager.Secret{}
	var updateMask []string

	// Handle labels.
	if g.flagSecretClearLabels {
		secret.Labels = map[string]string{}
		updateMask = append(updateMask, "labels")
	} else if len(g.flagSecretUpdateLabels) > 0 || len(g.flagSecretRemoveLabels) > 0 {
		labels := make(map[string]string)
		for k, v := range current.Labels {
			labels[k] = v
		}
		for _, k := range g.flagSecretRemoveLabels {
			delete(labels, k)
		}
		for k, v := range g.flagSecretUpdateLabels {
			labels[k] = v
		}
		secret.Labels = labels
		updateMask = append(updateMask, "labels")
	}

	// Handle expiration.
	if g.flagSecretRemoveExpiration {
		secret.ExpireTime = ""
		secret.Ttl = ""
		updateMask = append(updateMask, "expire_time", "ttl")
	} else {
		if g.flagSecretUpdateExpireTime != "" {
			secret.ExpireTime = g.flagSecretUpdateExpireTime
			updateMask = append(updateMask, "expire_time")
		}
		if g.flagSecretUpdateTTL != "" {
			secret.Ttl = g.flagSecretUpdateTTL
			updateMask = append(updateMask, "ttl")
		}
	}

	// Handle rotation.
	if g.flagSecretNextRotationTime != "" || g.flagSecretRotationPeriod != "" {
		secret.Rotation = &secretmanager.Rotation{}
		if g.flagSecretNextRotationTime != "" {
			secret.Rotation.NextRotationTime = g.flagSecretNextRotationTime
		}
		if g.flagSecretRotationPeriod != "" {
			secret.Rotation.RotationPeriod = g.flagSecretRotationPeriod
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

	fmt.Fprintf(g.inv.stdout, "Updated secret [%s].\n", updated.Name)
	return nil
}

func (g *secretsCmds) runSecretsDelete(cmd *cobra.Command, args []string) error {
	secretID := args[0]
	project, err := g.inv.project()
	if err != nil {
		return err
	}

	if !g.inv.flags.quiet {
		fmt.Fprintf(g.inv.stderr, "You are about to delete secret [%s]. This action cannot be undone.\n", secretID)
		fmt.Fprint(g.inv.stderr, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(g.inv.stderr, "Aborted.")
			return nil
		}
	}

	ctx := cmd.Context()
	svc, err := secrets.NewService(ctx, g.inv.flags.account)
	if err != nil {
		return err
	}

	name := secrets.SecretName(project, secretID, g.flagSecretsLocation)
	call := svc.Projects.Secrets.Delete(name).Context(ctx)
	if g.flagDeleteEtag != "" {
		call = call.Etag(g.flagDeleteEtag)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}

	fmt.Fprintf(g.inv.stderr, "Deleted secret [%s].\n", secretID)
	return nil
}

// --- Helpers ---

func (g *secretsCmds) readDataFile(path string) ([]byte, error) {
	if path == "-" {
		data, err := io.ReadAll(g.inv.stdin)
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

func (g *secretsCmds) addVersion(ctx context.Context, svc *secretmanager.Service, parent string, data []byte) error {
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
	fmt.Fprintf(g.inv.stderr, "Created version [%s] of the secret [%s].\n",
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

// secretsCmds holds the commands and flag values of one invocation's "secrets"
// command group, so invocations don't share state (#2070).
type secretsCmds struct {
	inv *invocation

	secretsCmd                      *cobra.Command
	secretsVersionsCmd              *cobra.Command
	secretsVersionsAccessCmd        *cobra.Command
	flagSecretName                  string
	flagOutFile                     string
	secretsCreateCmd                *cobra.Command
	flagDataFile                    string
	flagReplicationPolicy           string
	flagLocations                   []string
	flagSecretLabels                map[string]string
	flagExpireTime                  string
	flagTTL                         string
	flagSecretType                  string
	secretsVersionsListCmd          *cobra.Command
	flagVersionsListFormat          string
	flagVersionsListURI             bool
	secretsVersionsAddCmd           *cobra.Command
	flagVersionsAddDataFile         string
	secretsListCmd                  *cobra.Command
	flagSecretsFilter               string
	flagSecretsFormat               string
	flagSecretsURI                  bool
	secretsDescribeCmd              *cobra.Command
	secretsVersionsDescribeCmd      *cobra.Command
	secretsVersionsDisableCmd       *cobra.Command
	secretsVersionsEnableCmd        *cobra.Command
	secretsVersionsDestroyCmd       *cobra.Command
	secretsUpdateCmd                *cobra.Command
	flagSecretUpdateLabels          map[string]string
	flagSecretRemoveLabels          []string
	flagSecretClearLabels           bool
	flagSecretUpdateExpireTime      string
	flagSecretUpdateTTL             string
	flagSecretRemoveExpiration      bool
	flagSecretNextRotationTime      string
	flagSecretRotationPeriod        string
	secretsDeleteCmd                *cobra.Command
	flagSecretsLocation             string
	flagDeleteEtag                  string
	flagSecretsIamMember            string
	flagSecretsIamRole              string
	flagSecretsIamFormat            string
	flagSecretsReplConfigFile       string
	flagSecretsReplUpdateMask       string
	flagSecretsLocDescribeFmt       string
	flagSecretsLocListFmt           string
	secretsGetIamCmd                *cobra.Command
	secretsSetIamCmd                *cobra.Command
	secretsAddIamCmd                *cobra.Command
	secretsRemoveIamCmd             *cobra.Command
	secretsReplicationCmd           *cobra.Command
	secretsReplGetCmd               *cobra.Command
	secretsReplSetCmd               *cobra.Command
	secretsReplUpdateCmd            *cobra.Command
	flagSecretsRotInstanceID        string
	flagSecretsRotUsername          string
	flagSecretsRotPassword          string
	secretsEnableManagedRotationCmd *cobra.Command
	secretsRotateSecretCmd          *cobra.Command
	secretsLocationsCmd             *cobra.Command
	secretsLocDescribeCmd           *cobra.Command
	secretsLocListCmd               *cobra.Command
}

// newSecretsCmd returns the "secrets" command group for inv.
func newSecretsCmd(inv *invocation) *cobra.Command {
	return newSecretsCmds(inv).secretsCmd
}

// newSecretsCmds builds the "secrets" command group for inv and returns its state.
func newSecretsCmds(inv *invocation) *secretsCmds {
	g := &secretsCmds{inv: inv}
	g.secretsCmd = &cobra.Command{
		Use:   "secrets",
		Short: "Manage Secret Manager secrets",
	}
	g.secretsVersionsCmd = &cobra.Command{
		Use:   "versions",
		Short: "Manage secret versions",
	}
	g.secretsVersionsAccessCmd = &cobra.Command{
		Use:   "access VERSION",
		Short: "Access a secret version's data",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsAccess,
	}
	g.secretsCreateCmd = &cobra.Command{
		Use:   "create SECRET_ID",
		Short: "Create a new secret",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsCreate,
	}
	g.secretsVersionsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List secret versions",
		Args:  cobra.NoArgs,
		RunE:  g.runSecretsVersionsList,
	}
	g.secretsVersionsAddCmd = &cobra.Command{
		Use:   "add SECRET_ID",
		Short: "Add a new version to a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsAdd,
	}
	g.secretsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Args:  cobra.NoArgs,
		RunE:  g.runSecretsList,
	}
	g.secretsDescribeCmd = &cobra.Command{
		Use:   "describe SECRET_ID",
		Short: "Describe a secret's metadata",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsDescribe,
	}
	g.secretsVersionsDescribeCmd = &cobra.Command{
		Use:   "describe VERSION",
		Short: "Describe a secret version's metadata",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsDescribe,
	}
	g.secretsVersionsDisableCmd = &cobra.Command{
		Use:   "disable VERSION",
		Short: "Disable a secret version",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsDisable,
	}
	g.secretsVersionsEnableCmd = &cobra.Command{
		Use:   "enable VERSION",
		Short: "Enable a secret version",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsEnable,
	}
	g.secretsVersionsDestroyCmd = &cobra.Command{
		Use:   "destroy VERSION",
		Short: "Destroy a secret version's data",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsVersionsDestroy,
	}
	g.secretsUpdateCmd = &cobra.Command{
		Use:   "update SECRET_ID",
		Short: "Update a secret's properties",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsUpdate,
	}
	g.secretsDeleteCmd = &cobra.Command{
		Use:   "delete SECRET_ID",
		Short: "Delete a secret and all its versions",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runSecretsDelete,
	}
	g.secretsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SECRET_ID", Short: "Get the IAM policy for a secret",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsGetIam,
	}
	g.secretsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SECRET_ID POLICY_FILE", Short: "Replace the IAM policy for a secret",
		Args: cobra.ExactArgs(2), RunE: g.runSecretsSetIam,
	}
	g.secretsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SECRET_ID", Short: "Add an IAM binding to a secret",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsAddIam,
	}
	g.secretsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SECRET_ID", Short: "Remove an IAM binding from a secret",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsRemoveIam,
	}
	g.secretsReplicationCmd = &cobra.Command{Use: "replication", Short: "Manage secret replication"}
	g.secretsReplGetCmd = &cobra.Command{
		Use: "get SECRET_ID", Short: "Get the replication policy of a secret",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsReplGet,
	}
	g.secretsReplSetCmd = &cobra.Command{
		Use: "set SECRET_ID", Short: "Replace the replication policy of a secret from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsReplSet,
	}
	g.secretsReplUpdateCmd = &cobra.Command{
		Use: "update SECRET_ID", Short: "Patch the replication policy of a secret from a --config-file",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsReplUpdate,
	}
	g.secretsEnableManagedRotationCmd = &cobra.Command{
		Use:   "enable-managed-rotation SECRET_ID",
		Short: "Enable managed rotation for a secret (Cloud SQL Single User credentials)",
		Args:  cobra.ExactArgs(1), RunE: g.runSecretsEnableManagedRotation,
	}
	g.secretsRotateSecretCmd = &cobra.Command{
		Use:   "rotate-secret SECRET_ID",
		Short: "Rotate a secret",
		Args:  cobra.ExactArgs(1), RunE: g.runSecretsRotateSecret,
	}
	g.secretsLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Secret Manager locations"}
	g.secretsLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Secret Manager location",
		Args: cobra.ExactArgs(1), RunE: g.runSecretsLocDescribe,
	}
	g.secretsLocListCmd = &cobra.Command{
		Use: "list", Short: "List Secret Manager locations",
		Args: cobra.NoArgs, RunE: g.runSecretsLocList,
	}
	g.initSecrets()
	g.initSecretsMore()
	return g
}
