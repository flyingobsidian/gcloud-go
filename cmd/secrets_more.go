package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/secrets"
	"github.com/spf13/cobra"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

// --- gcloud secrets IAM, replication, locations (#546) ---

var (
	flagSecretsIamMember       string
	flagSecretsIamRole         string
	flagSecretsIamFormat       string
	flagSecretsReplConfigFile  string
	flagSecretsReplUpdateMask  string
	flagSecretsLocDescribeFmt  string
	flagSecretsLocListFmt      string
	flagSecretsLocListLocation string
)

// --- IAM ---

var (
	secretsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SECRET_ID", Short: "Get the IAM policy for a secret",
		Args: cobra.ExactArgs(1), RunE: runSecretsGetIam,
	}
	secretsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SECRET_ID POLICY_FILE", Short: "Replace the IAM policy for a secret",
		Args: cobra.ExactArgs(2), RunE: runSecretsSetIam,
	}
	secretsAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding SECRET_ID", Short: "Add an IAM binding to a secret",
		Args: cobra.ExactArgs(1), RunE: runSecretsAddIam,
	}
	secretsRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding SECRET_ID", Short: "Remove an IAM binding from a secret",
		Args: cobra.ExactArgs(1), RunE: runSecretsRemoveIam,
	}
)

// --- replication ---

var secretsReplicationCmd = &cobra.Command{Use: "replication", Short: "Manage secret replication"}

var (
	secretsReplGetCmd = &cobra.Command{
		Use: "get SECRET_ID", Short: "Get the replication policy of a secret",
		Args: cobra.ExactArgs(1), RunE: runSecretsReplGet,
	}
	secretsReplSetCmd = &cobra.Command{
		Use: "set SECRET_ID", Short: "Replace the replication policy of a secret from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSecretsReplSet,
	}
	secretsReplUpdateCmd = &cobra.Command{
		Use: "update SECRET_ID", Short: "Patch the replication policy of a secret from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runSecretsReplUpdate,
	}
)

// --- locations ---

var secretsLocationsCmd = &cobra.Command{Use: "locations", Short: "Manage Secret Manager locations"}

var (
	secretsLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Secret Manager location",
		Args: cobra.ExactArgs(1), RunE: runSecretsLocDescribe,
	}
	secretsLocListCmd = &cobra.Command{
		Use: "list", Short: "List Secret Manager locations",
		Args: cobra.NoArgs, RunE: runSecretsLocList,
	}
)

func init() {
	// IAM
	for _, c := range []*cobra.Command{secretsGetIamCmd, secretsSetIamCmd, secretsAddIamCmd, secretsRemoveIamCmd} {
		c.Flags().StringVar(&flagSecretsLocation, "location", "",
			"Secret Manager location (for regional secrets)")
	}
	secretsGetIamCmd.Flags().StringVar(&flagSecretsIamFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{secretsAddIamCmd, secretsRemoveIamCmd} {
		c.Flags().StringVar(&flagSecretsIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagSecretsIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	secretsCmd.AddCommand(secretsGetIamCmd, secretsSetIamCmd, secretsAddIamCmd, secretsRemoveIamCmd)

	// replication
	for _, c := range []*cobra.Command{secretsReplGetCmd, secretsReplSetCmd, secretsReplUpdateCmd} {
		c.Flags().StringVar(&flagSecretsLocation, "location", "",
			"Secret Manager location (for regional secrets)")
	}
	for _, c := range []*cobra.Command{secretsReplSetCmd, secretsReplUpdateCmd} {
		c.Flags().StringVar(&flagSecretsReplConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Replication body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	secretsReplUpdateCmd.Flags().StringVar(&flagSecretsReplUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field, prefixed with 'replication.')")
	secretsReplicationCmd.AddCommand(secretsReplGetCmd, secretsReplSetCmd, secretsReplUpdateCmd)
	secretsCmd.AddCommand(secretsReplicationCmd)

	// locations
	secretsLocDescribeCmd.Flags().StringVar(&flagSecretsLocDescribeFmt, "format", "", "Output format")
	secretsLocListCmd.Flags().StringVar(&flagSecretsLocListFmt, "format", "", "Output format")
	secretsLocationsCmd.AddCommand(secretsLocDescribeCmd, secretsLocListCmd)
	secretsCmd.AddCommand(secretsLocationsCmd)
}

// --- IAM impl ---

func runSecretsGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Secrets.GetIamPolicy(secrets.SecretName(project, args[0], flagSecretsLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagSecretsIamFormat)
}

func runSecretsSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &secretmanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Secrets.SetIamPolicy(secrets.SecretName(project, args[0], flagSecretsLocation),
		&secretmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runSecretsAddIam(cmd *cobra.Command, args []string) error {
	return secretsModifyIam(args[0], func(p *secretmanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagSecretsIamRole {
				for _, m := range b.Members {
					if m == flagSecretsIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagSecretsIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &secretmanager.Binding{
			Role: flagSecretsIamRole, Members: []string{flagSecretsIamMember},
		})
	})
}

func runSecretsRemoveIam(cmd *cobra.Command, args []string) error {
	return secretsModifyIam(args[0], func(p *secretmanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagSecretsIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != flagSecretsIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func secretsModifyIam(secretID string, mutate func(*secretmanager.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := secrets.SecretName(project, secretID, flagSecretsLocation)
	policy, err := svc.Projects.Secrets.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Secrets.SetIamPolicy(resource, &secretmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- replication impl ---

func runSecretsReplGet(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	secret, err := svc.Projects.Secrets.Get(secrets.SecretName(project, args[0], flagSecretsLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting secret: %w", err)
	}
	return emitFormatted(secret.Replication, "")
}

func runSecretsReplSet(cmd *cobra.Command, args []string) error {
	return runSecretsReplPatch(args[0], "replication")
}

func runSecretsReplUpdate(cmd *cobra.Command, args []string) error {
	mask := flagSecretsReplUpdateMask
	return runSecretsReplPatch(args[0], mask)
}

func runSecretsReplPatch(secretID, mask string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	repl := &secretmanager.Replication{}
	if err := loadYAMLOrJSONInto(flagSecretsReplConfigFile, repl); err != nil {
		return err
	}
	if mask == "" {
		fields := nonEmptyJSONFields(repl)
		for i, f := range fields {
			fields[i] = "replication." + f
		}
		mask = joinMask(fields)
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &secretmanager.Secret{Replication: repl}
	got, err := svc.Projects.Secrets.Patch(secrets.SecretName(project, secretID, flagSecretsLocation), body).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating replication: %w", err)
	}
	return emitFormatted(got.Replication, "")
}

// --- locations impl ---

func runSecretsLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(got, flagSecretsLocDescribeFmt)
}

func runSecretsLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := secrets.NewService(ctx, flagAccount)
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
	if flagSecretsLocListFmt != "" {
		return emitFormatted(all, flagSecretsLocListFmt)
	}
	fmt.Printf("%-30s %s\n", "LOCATION_ID", "NAME")
	for _, l := range all {
		fmt.Printf("%-30s %s\n", l.LocationId, path.Base(l.Name))
	}
	return nil
}
