package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/secrets"
	"github.com/spf13/cobra"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

// --- gcloud secrets IAM, replication, locations (#546) ---

var (
	flagSecretsLocListLocation string
)

// --- IAM ---

// --- replication ---

// --- managed rotation ---

// --- locations ---

func (g *secretsCmds) initSecretsMore() {
	// IAM
	for _, c := range []*cobra.Command{g.secretsGetIamCmd, g.secretsSetIamCmd, g.secretsAddIamCmd, g.secretsRemoveIamCmd} {
		c.Flags().StringVar(&g.flagSecretsLocation, "location", "",
			"Secret Manager location (for regional secrets)")
	}
	g.secretsGetIamCmd.Flags().StringVar(&g.flagSecretsIamFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{g.secretsAddIamCmd, g.secretsRemoveIamCmd} {
		c.Flags().StringVar(&g.flagSecretsIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagSecretsIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	g.secretsCmd.AddCommand(g.secretsGetIamCmd, g.secretsSetIamCmd, g.secretsAddIamCmd, g.secretsRemoveIamCmd)

	// replication
	for _, c := range []*cobra.Command{g.secretsReplGetCmd, g.secretsReplSetCmd, g.secretsReplUpdateCmd} {
		c.Flags().StringVar(&g.flagSecretsLocation, "location", "",
			"Secret Manager location (for regional secrets)")
	}
	for _, c := range []*cobra.Command{g.secretsReplSetCmd, g.secretsReplUpdateCmd} {
		c.Flags().StringVar(&g.flagSecretsReplConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Replication body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.secretsReplUpdateCmd.Flags().StringVar(&g.flagSecretsReplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field, prefixed with 'replication.')")
	g.secretsReplicationCmd.AddCommand(g.secretsReplGetCmd, g.secretsReplSetCmd, g.secretsReplUpdateCmd)
	g.secretsCmd.AddCommand(g.secretsReplicationCmd)

	// locations
	g.secretsLocDescribeCmd.Flags().StringVar(&g.flagSecretsLocDescribeFmt, "format", "", "Output format")
	g.secretsLocListCmd.Flags().StringVar(&g.flagSecretsLocListFmt, "format", "", "Output format")
	g.secretsLocationsCmd.AddCommand(g.secretsLocDescribeCmd, g.secretsLocListCmd)
	g.secretsCmd.AddCommand(g.secretsLocationsCmd)

	// managed rotation / rotate-secret
	for _, c := range []*cobra.Command{g.secretsEnableManagedRotationCmd, g.secretsRotateSecretCmd} {
		c.Flags().StringVar(&g.flagSecretsLocation, "location", "",
			"Secret Manager location (for regional secrets)")
	}
	g.secretsEnableManagedRotationCmd.Flags().StringVar(&g.flagSecretsRotInstanceID, "instance-id", "",
		"Cloud SQL instance ID (required for Cloud SQL Single User rotation)")
	g.secretsEnableManagedRotationCmd.Flags().StringVar(&g.flagSecretsRotUsername, "username", "",
		"Cloud SQL username (required for Cloud SQL Single User rotation)")
	g.secretsEnableManagedRotationCmd.Flags().StringVar(&g.flagSecretsRotPassword, "password", "",
		"Cloud SQL password (optional; the server generates one if omitted)")
	g.secretsCmd.AddCommand(g.secretsEnableManagedRotationCmd, g.secretsRotateSecretCmd)
}

// secretTypeEnum maps a kebab-case CLI value to the SecretType enum expected
// by the Secret Manager API. Empty input leaves the enum unspecified.
func secretTypeEnum(v string) (string, error) {
	switch strings.ToLower(strings.ReplaceAll(v, "_", "-")) {
	case "":
		return "", nil
	case "cloud-sql-single-user-credentials":
		return "CLOUD_SQL_SINGLE_USER_CREDENTIALS", nil
	case "other":
		return "OTHER", nil
	case "unspecified", "secret-type-unspecified":
		return "SECRET_TYPE_UNSPECIFIED", nil
	default:
		return "", fmt.Errorf("--secret-type must be one of cloud-sql-single-user-credentials, other, unspecified (got %q)", v)
	}
}

// --- IAM impl ---

func (g *secretsCmds) runSecretsGetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Secrets.GetIamPolicy(secrets.SecretName(project, args[0], g.flagSecretsLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, policy, g.flagSecretsIamFormat)
}

func (g *secretsCmds) runSecretsSetIam(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	policy := &secretmanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Secrets.SetIamPolicy(secrets.SecretName(project, args[0], g.flagSecretsLocation),
		&secretmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *secretsCmds) runSecretsAddIam(cmd *cobra.Command, args []string) error {
	return g.secretsModifyIam(args[0], func(p *secretmanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role == g.flagSecretsIamRole {
				for _, m := range b.Members {
					if m == g.flagSecretsIamMember {
						return
					}
				}
				b.Members = append(b.Members, g.flagSecretsIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &secretmanager.Binding{
			Role: g.flagSecretsIamRole, Members: []string{g.flagSecretsIamMember},
		})
	})
}

func (g *secretsCmds) runSecretsRemoveIam(cmd *cobra.Command, args []string) error {
	return g.secretsModifyIam(args[0], func(p *secretmanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role != g.flagSecretsIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != g.flagSecretsIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func (g *secretsCmds) secretsModifyIam(secretID string, mutate func(*secretmanager.Policy)) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	resource := secrets.SecretName(project, secretID, g.flagSecretsLocation)
	policy, err := svc.Projects.Secrets.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Secrets.SetIamPolicy(resource, &secretmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

// --- replication impl ---

func (g *secretsCmds) runSecretsReplGet(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	secret, err := svc.Projects.Secrets.Get(secrets.SecretName(project, args[0], g.flagSecretsLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting secret: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, secret.Replication, "")
}

func (g *secretsCmds) runSecretsReplSet(cmd *cobra.Command, args []string) error {
	return g.runSecretsReplPatch(args[0], "replication")
}

func (g *secretsCmds) runSecretsReplUpdate(cmd *cobra.Command, args []string) error {
	mask := g.flagSecretsReplUpdateMask
	return g.runSecretsReplPatch(args[0], mask)
}

func (g *secretsCmds) runSecretsReplPatch(secretID, mask string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	repl := &secretmanager.Replication{}
	if err := loadYAMLOrJSONInto(g.flagSecretsReplConfigFile, repl); err != nil {
		return err
	}
	if mask == "" {
		fields := nonEmptyJSONFields(repl)
		for i, f := range fields {
			fields[i] = "replication." + f
		}
		mask = joinMask(fields)
	}
	ctx := g.inv.context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	body := &secretmanager.Secret{Replication: repl}
	got, err := svc.Projects.Secrets.Patch(secrets.SecretName(project, secretID, g.flagSecretsLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating replication: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got.Replication, "")
}

// --- locations impl ---

func (g *secretsCmds) runSecretsLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagSecretsLocDescribeFmt)
}

func (g *secretsCmds) runSecretsLocList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	var all []*secretmanager.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagSecretsLocListFmt != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagSecretsLocListFmt)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-30s %s\n", "LOCATION_ID", "NAME")
	for _, l := range all {
		fmt.Fprintf(g.inv.stdout, "%-30s %s\n", l.LocationId, path.Base(l.Name))
	}
	return nil
}

// --- managed rotation impl ---

func (g *secretsCmds) runSecretsEnableManagedRotation(cmd *cobra.Command, args []string) error {
	if g.flagSecretsRotInstanceID == "" || g.flagSecretsRotUsername == "" {
		return fmt.Errorf("--instance-id and --username are required for Cloud SQL Single User managed rotation")
	}
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	req := &secretmanager.EnableManagedRotationRequest{
		CloudSqlSingleUserCredentials: &secretmanager.CloudSQLSingleUserCredentials{
			InstanceId: g.flagSecretsRotInstanceID,
			Username:   g.flagSecretsRotUsername,
			Password:   g.flagSecretsRotPassword,
		},
	}
	name := secrets.SecretName(project, args[0], g.flagSecretsLocation)
	got, err := svc.Projects.Secrets.EnableManagedRotation(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling managed rotation: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Enabled managed rotation for secret [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *secretsCmds) runSecretsRotateSecret(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().SecretManager(ctx)
	if err != nil {
		return err
	}
	name := secrets.SecretName(project, args[0], g.flagSecretsLocation)
	got, err := svc.Projects.Secrets.RotateSecret(name, &secretmanager.RotateSecretRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rotating secret: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Rotated secret [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, "")
}
